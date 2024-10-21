// Copyright (C) 2019-2024 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package heartbeat

import (
	"fmt"
	"testing"
	"time"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/data/account"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/committee"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/ledger/ledgercore"
	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/algorand/go-deadlock"
	"github.com/stretchr/testify/require"
)

type table map[basics.Address]ledgercore.AccountData

type mockedLedger struct {
	mu      deadlock.Mutex
	waiters map[basics.Round]chan struct{}
	history []table
	hdr     bookkeeping.BlockHeader

	participants map[basics.Address]*crypto.OneTimeSignatureSecrets
}

func newMockedLedger() mockedLedger {
	return mockedLedger{
		waiters: make(map[basics.Round]chan struct{}),
		history: []table{nil}, // some genesis accounts could go here
		hdr: bookkeeping.BlockHeader{
			UpgradeState: bookkeeping.UpgradeState{
				CurrentProtocol: protocol.ConsensusFuture,
			},
		},
	}
}

func (l *mockedLedger) LastRound() basics.Round {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastRound()
}
func (l *mockedLedger) lastRound() basics.Round {
	return basics.Round(len(l.history) - 1)
}

func (l *mockedLedger) WaitMem(r basics.Round) chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.waiters[r] == nil {
		l.waiters[r] = make(chan struct{})
	}

	// Return an already-closed channel if we already have the block.
	if r <= l.lastRound() {
		close(l.waiters[r])
		retChan := l.waiters[r]
		delete(l.waiters, r)
		return retChan
	}

	return l.waiters[r]
}

// BlockHdr allows the service access to consensus values
func (l *mockedLedger) BlockHdr(r basics.Round) (bookkeeping.BlockHeader, error) {
	if r > l.LastRound() {
		return bookkeeping.BlockHeader{}, fmt.Errorf("%d is beyond current block (%d)", r, l.LastRound())
	}
	// return the template hdr, with round
	hdr := l.hdr
	hdr.Round = r
	return hdr, nil
}

func (l *mockedLedger) addBlock(delta table) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Printf("addBlock %d\n", l.lastRound()+1)
	l.history = append(l.history, delta)

	for r, ch := range l.waiters {
		switch {
		case r < l.lastRound():
			fmt.Printf("%d < %d\n", r, l.lastRound())
			panic("why is there a waiter for an old block?")
		case r == l.lastRound():
			close(ch)
			delete(l.waiters, r)
		case r > l.lastRound():
			/* waiter keeps waiting */
		}
	}
	return nil
}

func (l *mockedLedger) LookupAccount(round basics.Round, addr basics.Address) (ledgercore.AccountData, basics.Round, basics.MicroAlgos, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if round > l.lastRound() {
		panic("mockedLedger.LookupAccount: future round")
	}

	for r := round; r <= round; r-- {
		if acct, ok := l.history[r][addr]; ok {
			more := basics.MicroAlgos{Raw: acct.MicroAlgos.Raw + 1}
			return acct, round, more, nil
		}
	}
	return ledgercore.AccountData{}, round, basics.MicroAlgos{}, nil
}

// waitFor confirms that the Service made it through the last block in the
// ledger and is waiting for the next. The Service is written such that it
// operates properly without this sort of wait, but for testing, we often want
// to wait so that we can confirm that the Service *didn't* do something.
func (l *mockedLedger) waitFor(s *Service, a *require.Assertions) {
	a.Eventually(func() bool { // delay and confirm that the service advances to wait for next block
		_, ok := l.waiters[l.LastRound()+1]
		return ok
	}, time.Second, 10*time.Millisecond)
}

func (l *mockedLedger) Keys(rnd basics.Round) []account.ParticipationRecordForRound {
	var ret []account.ParticipationRecordForRound
	for addr, secrets := range l.participants {
		if rnd > l.LastRound() { // Usually we're looking for key material for a future round
			rnd = l.LastRound()
		}
		acct, _, _, err := l.LookupAccount(rnd, addr)
		if err != nil {
			panic(err.Error())
		}

		ret = append(ret, account.ParticipationRecordForRound{
			ParticipationRecord: account.ParticipationRecord{
				ParticipationID: [32]byte{},
				Account:         addr,
				Voting:          secrets,
				FirstValid:      acct.VoteFirstValid,
				LastValid:       acct.VoteLastValid,
				KeyDilution:     acct.VoteKeyDilution,
			},
		})
	}
	return ret
}

func (l *mockedLedger) addParticipant(addr basics.Address, otss *crypto.OneTimeSignatureSecrets) {
	if l.participants == nil {
		l.participants = make(map[basics.Address]*crypto.OneTimeSignatureSecrets)
	}
	l.participants[addr] = otss
}

type txnSink [][]transactions.SignedTxn

func (ts *txnSink) BroadcastInternalSignedTxGroup(group []transactions.SignedTxn) error {
	fmt.Printf("sinking %+v\n", group[0].Txn.Header)
	*ts = append(*ts, group)
	return nil
}

func TestStartStop(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	a := require.New(t)
	sink := txnSink{}
	ledger := newMockedLedger()
	s := NewService(&ledger, &ledger, &sink, logging.TestingLog(t))
	a.NotNil(s)
	a.NoError(ledger.addBlock(nil))
	s.Start()
	a.NoError(ledger.addBlock(nil))
	s.Stop()
}

func makeBlock(r basics.Round) bookkeeping.Block {
	return bookkeeping.Block{
		BlockHeader: bookkeeping.BlockHeader{Round: r},
		Payset:      []transactions.SignedTxnInBlock{},
	}
}

func TestHeartBeatOnlyWhenChallenged(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	a := require.New(t)
	sink := txnSink{}
	ledger := newMockedLedger()
	s := NewService(&ledger, &ledger, &sink, logging.TestingLog(t))
	s.Start()

	joe := basics.Address{0xcc}  // 0xcc will matter when we set the challenge
	mary := basics.Address{0xaa} // 0xaa will matter when we set the challenge
	ledger.addParticipant(joe, nil)
	ledger.addParticipant(mary, nil)

	acct := ledgercore.AccountData{}

	a.NoError(ledger.addBlock(table{joe: acct}))
	ledger.waitFor(s, a)
	a.Empty(sink)

	// now they are online, but not challenged, so no heartbeat
	acct.Status = basics.Online
	acct.VoteKeyDilution = 100
	otss := crypto.GenerateOneTimeSignatureSecrets(
		basics.OneTimeIDForRound(ledger.LastRound(), acct.VoteKeyDilution).Batch,
		5)
	acct.VoteID = otss.OneTimeSignatureVerifier
	ledger.addParticipant(joe, otss)
	ledger.addParticipant(mary, otss)

	a.NoError(ledger.addBlock(table{joe: acct, mary: acct}))
	a.Empty(sink)

	// now we have to make it seem like joe has been challenged. We obtain the
	// payout rules to find the first challenge round, skip forward to it, then
	// go forward half a grace period. Only then should the service heartbeat
	hdr, err := ledger.BlockHdr(ledger.LastRound())
	ledger.hdr.Seed = committee.Seed{0xc8} // share 5 bits with 0xcc
	a.NoError(err)
	rules := config.Consensus[hdr.CurrentProtocol].Payouts
	for ledger.LastRound() < basics.Round(rules.ChallengeInterval+rules.ChallengeGracePeriod/2) {
		a.NoError(ledger.addBlock(table{}))
		ledger.waitFor(s, a)
		a.Empty(sink)
	}

	a.NoError(ledger.addBlock(table{joe: acct}))
	ledger.waitFor(s, a)
	a.Len(sink, 1) // only one heartbeat (for joe)
	a.Len(sink[0], 1)
	a.Equal(sink[0][0].Txn.Type, protocol.HeartbeatTx)
	a.Equal(sink[0][0].Txn.HbAddress, joe)

	s.Stop()
}
