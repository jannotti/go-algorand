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

package ledger

import (
	"context"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/agreement"
	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/crypto/merklesignature"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/txntest"
	ledgertesting "github.com/algorand/go-algorand/ledger/testing"
	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/algorand/go-algorand/util/execpool"
)

func TestBlockEvaluator(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genesisInitState, addrs, keys := ledgertesting.Genesis(10)

	l, err := OpenLedger(logging.TestingLog(t), t.Name(), true, genesisInitState, config.GetDefaultLocal())
	require.NoError(t, err)
	defer l.Close()

	genesisBlockHeader, err := l.BlockHdr(basics.Round(0))
	require.NoError(t, err)
	newBlock := bookkeeping.MakeBlock(genesisBlockHeader)
	eval, err := l.StartEvaluator(newBlock.BlockHeader, 0, 0, nil)
	require.NoError(t, err)

	genHash := l.GenesisHash()
	txn := transactions.Transaction{
		Type: protocol.PaymentTx,
		Header: transactions.Header{
			Sender:      addrs[0],
			Fee:         minFee,
			FirstValid:  newBlock.Round(),
			LastValid:   newBlock.Round(),
			GenesisHash: genHash,
		},
		PaymentTxnFields: transactions.PaymentTxnFields{
			Receiver: addrs[1],
			Amount:   basics.MicroAlgos{Raw: 100},
		},
	}

	// Correct signature should work
	st := txn.Sign(keys[0])
	err = eval.Transaction(st, transactions.ApplyData{})
	require.NoError(t, err)

	// Broken signature should fail
	stbad := st
	st.Sig[2] ^= 8
	txgroup := []transactions.SignedTxn{stbad}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)

	// Repeat should fail
	txgroup = []transactions.SignedTxn{st}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	err = eval.Transaction(st, transactions.ApplyData{})
	require.Error(t, err)

	// out of range should fail
	btxn := txn
	btxn.FirstValid++
	btxn.LastValid += 2
	st = btxn.Sign(keys[0])
	txgroup = []transactions.SignedTxn{st}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	err = eval.Transaction(st, transactions.ApplyData{})
	require.Error(t, err)

	// bogus group should fail
	btxn = txn
	btxn.Group[1] = 1
	st = btxn.Sign(keys[0])
	txgroup = []transactions.SignedTxn{st}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	err = eval.Transaction(st, transactions.ApplyData{})
	require.Error(t, err)

	// mixed fields should fail
	btxn = txn
	btxn.XferAsset = 3
	st = btxn.Sign(keys[0])
	txgroup = []transactions.SignedTxn{st}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	// We don't test eval.Transaction() here because it doesn't check txn.WellFormed(), instead relying on that to have already been checked by the transaction pool.
	// err = eval.Transaction(st, transactions.ApplyData{})
	// require.Error(t, err)

	selfTxn := transactions.Transaction{
		Type: protocol.PaymentTx,
		Header: transactions.Header{
			Sender:      addrs[2],
			Fee:         minFee,
			FirstValid:  newBlock.Round(),
			LastValid:   newBlock.Round(),
			GenesisHash: genHash,
		},
		PaymentTxnFields: transactions.PaymentTxnFields{
			Receiver: addrs[2],
			Amount:   basics.MicroAlgos{Raw: 100},
		},
	}
	stxn := selfTxn.Sign(keys[2])

	// TestTransactionGroup() and Transaction() should have the same outcome, but work slightly different code paths.
	txgroup = []transactions.SignedTxn{stxn}
	err = eval.TestTransactionGroup(txgroup)
	require.NoError(t, err)

	err = eval.Transaction(stxn, transactions.ApplyData{})
	require.NoError(t, err)

	t3 := txn
	t3.Amount.Raw++
	t4 := selfTxn
	t4.Amount.Raw++

	// a group without .Group should fail
	s3 := t3.Sign(keys[0])
	s4 := t4.Sign(keys[2])
	txgroup = []transactions.SignedTxn{s3, s4}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	txgroupad := transactions.WrapSignedTxnsWithAD(txgroup)
	err = eval.TransactionGroup(txgroupad)
	require.Error(t, err)

	// Test a group that should work
	var group transactions.TxGroup
	group.TxGroupHashes = []crypto.Digest{crypto.HashObj(t3), crypto.HashObj(t4)}
	t3.Group = crypto.HashObj(group)
	t4.Group = t3.Group
	s3 = t3.Sign(keys[0])
	s4 = t4.Sign(keys[2])
	txgroup = []transactions.SignedTxn{s3, s4}
	err = eval.TestTransactionGroup(txgroup)
	require.NoError(t, err)

	// disagreement on Group id should fail
	t4bad := t4
	t4bad.Group[3] ^= 3
	s4bad := t4bad.Sign(keys[2])
	txgroup = []transactions.SignedTxn{s3, s4bad}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)
	txgroupad = transactions.WrapSignedTxnsWithAD(txgroup)
	err = eval.TransactionGroup(txgroupad)
	require.Error(t, err)

	// missing part of the group should fail
	txgroup = []transactions.SignedTxn{s3}
	err = eval.TestTransactionGroup(txgroup)
	require.Error(t, err)

	validatedBlock, err := eval.GenerateBlock()
	require.NoError(t, err)

	accts := genesisInitState.Accounts
	bal0 := accts[addrs[0]]
	bal1 := accts[addrs[1]]
	bal2 := accts[addrs[2]]

	l.AddValidatedBlock(*validatedBlock, agreement.Certificate{})

	bal0new, _, _, err := l.LookupAccount(newBlock.Round(), addrs[0])
	require.NoError(t, err)
	bal1new, _, _, err := l.LookupAccount(newBlock.Round(), addrs[1])
	require.NoError(t, err)
	bal2new, _, _, err := l.LookupAccount(newBlock.Round(), addrs[2])
	require.NoError(t, err)

	require.Equal(t, bal0new.MicroAlgos.Raw, bal0.MicroAlgos.Raw-minFee.Raw-100)
	require.Equal(t, bal1new.MicroAlgos.Raw, bal1.MicroAlgos.Raw+100)
	require.Equal(t, bal2new.MicroAlgos.Raw, bal2.MicroAlgos.Raw-minFee.Raw)
}

// TestMiningFees ensures that the proper portion of tx fees go to the proposer
func TestMiningFees(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	// Lots of balance checks that would be messed up by rewards
	genBalances, addrs, _ := ledgertesting.NewTestGenesis(ledgertesting.TurnOffRewards)
	miningBegins := 40
	ledgertesting.TestConsensusRange(t, miningBegins-1, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		proposer := basics.Address{0x01, 0x011}
		const eFee = 3_000_000
		dl.txn(
			&txntest.Txn{Type: "pay", Sender: addrs[1],
				Receiver: proposer, Amount: eFee + 50_000_000*1_000_000 + 1},
		)

		prp := lookup(dl.t, dl.generator, proposer)
		require.False(t, prp.IncentiveEligible)

		dl.txn(&txntest.Txn{
			Type:         "keyreg",
			Sender:       proposer,
			Fee:          eFee,
			VotePK:       crypto.OneTimeSignatureVerifier{0x01},
			SelectionPK:  crypto.VRFVerifier{0x02},
			StateProofPK: merklesignature.Commitment{0x03},
			VoteFirst:    1, VoteLast: 1000,
		})

		prp = lookup(dl.t, dl.generator, proposer)
		require.Equal(t, ver >= miningBegins, prp.IncentiveEligible)

		dl.fullBlock() // start with an empty block, so no mining fees are paid at start of next one

		presink := micros(dl.t, dl.generator, genBalances.FeeSink)
		preprop := micros(dl.t, dl.generator, proposer)
		t.Log(" presink", presink)
		t.Log(" preprop", preprop)
		dl.beginBlock()
		pay := txntest.Txn{
			Type:     "pay",
			Sender:   addrs[1],
			Receiver: addrs[2],
			Amount:   100000,
		}
		dl.txns(&pay, pay.Args("again"))
		vb := dl.endBlock(proposer)

		const bonus1 = 5_000_000 // the first bonus value, set in
		if ver >= miningBegins {
			require.True(t, dl.generator.GenesisProto().Mining().Enabled)    // version sanity check
			require.NotZero(t, dl.generator.GenesisProto().Mining().Percent) // version sanity check
			// new fields are in the header
			require.EqualValues(t, 2000, vb.Block().FeesCollected.Raw)
			require.EqualValues(t, bonus1, vb.Block().Bonus.Raw)
			require.EqualValues(t, bonus1+1_500, vb.Block().ProposerPayout.Raw)
		} else {
			require.False(t, dl.generator.GenesisProto().Mining().Enabled)
			require.Zero(t, dl.generator.GenesisProto().Mining().Percent) // version sanity check
			require.Zero(t, vb.Block().FeesCollected)
			require.Zero(t, vb.Block().Bonus)
			require.Zero(t, vb.Block().ProposerPayout)
		}

		postsink := micros(dl.t, dl.generator, genBalances.FeeSink)
		postprop := micros(dl.t, dl.generator, proposer)
		t.Log(" postsink", postsink)
		t.Log(" postprop", postprop)

		// At the end of the block, all fees are still in the sink.
		require.EqualValues(t, 2000, postsink-presink)
		require.EqualValues(t, 0, postprop-preprop)

		// Do the next block, which moves part of the fees + bonus to proposer
		dl.fullBlock()
		postsink = micros(dl.t, dl.generator, genBalances.FeeSink)
		postprop = micros(dl.t, dl.generator, proposer)
		t.Log(" postsink2", postsink)
		t.Log(" postprop2", postprop)

		if ver >= miningBegins {
			require.EqualValues(t, bonus1+1500, postprop-preprop) // based on 75% in config/consensus.go
			require.EqualValues(t, bonus1-500, presink-postsink)
		} else {
			// stayed in the feesink
			require.EqualValues(t, 0, postprop-preprop, "%v", proposer)
			require.EqualValues(t, 2000, postsink-presink)
		}
	})
}

// TestIncentiveEligible checks that keyreg with extra fee turns on the incentive eligible flag
func TestIncentiveEligible(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	miningBegins := 40
	ledgertesting.TestConsensusRange(t, miningBegins-1, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		tooSmall := basics.Address{0x01, 0x011}
		smallest := basics.Address{0x01, 0x022}

		// They begin ineligible
		for _, addr := range []basics.Address{tooSmall, smallest} {
			acct, _, _, err := dl.generator.LookupLatest(addr)
			require.NoError(t, err)
			require.False(t, acct.IncentiveEligible)
		}

		// Fund everyone
		dl.txns(&txntest.Txn{Type: "pay", Sender: addrs[1], Receiver: tooSmall, Amount: 10_000_000},
			&txntest.Txn{Type: "pay", Sender: addrs[1], Receiver: smallest, Amount: 10_000_000},
		)

		// Keyreg (but offline) with various fees. No effect on incentive eligible
		dl.txns(&txntest.Txn{Type: "keyreg", Sender: tooSmall, Fee: 2_000_000 - 1},
			&txntest.Txn{Type: "keyreg", Sender: smallest, Fee: 2_000_000},
		)

		for _, addr := range []basics.Address{tooSmall, smallest} {
			acct, _, _, err := dl.generator.LookupLatest(addr)
			require.NoError(t, err)
			require.False(t, acct.IncentiveEligible)
		}

		// Keyreg to get online with various fees. Sufficient fee gets `smallest` eligible
		keyreg := txntest.Txn{
			Type:         "keyreg",
			VotePK:       crypto.OneTimeSignatureVerifier{0x01},
			SelectionPK:  crypto.VRFVerifier{0x02},
			StateProofPK: merklesignature.Commitment{0x03},
			VoteFirst:    1, VoteLast: 1000,
		}
		tooSmallKR := keyreg
		tooSmallKR.Sender = tooSmall
		tooSmallKR.Fee = 2_000_000 - 1

		smallKR := keyreg
		smallKR.Sender = smallest
		smallKR.Fee = 2_000_000
		dl.txns(&tooSmallKR, &smallKR)
		a, _, _, err := dl.generator.LookupLatest(tooSmall)
		require.NoError(t, err)
		require.False(t, a.IncentiveEligible)
		a, _, _, err = dl.generator.LookupLatest(smallest)
		require.NoError(t, err)
		require.Equal(t, a.IncentiveEligible, ver >= miningBegins)
	})
}

// TestAbsentTracking checks that LastProposed and LastHeartbeat are updated
// properly.
func TestAbsentTracking(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis(func(cfg *ledgertesting.GenesisCfg) {
		cfg.OnlineCount = 2 // So we know proposer should propose every 2 rounds, on average
	})
	checkingBegins := 40
	ledgertesting.TestConsensusRange(t, checkingBegins-1, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		// have addrs[1] go online, which makes it eligible for suspension
		dl.txn(&txntest.Txn{
			Type:        "keyreg",
			Sender:      addrs[1],
			VotePK:      [32]byte{1},
			SelectionPK: [32]byte{1},
		})

		totals, err := dl.generator.Totals(1)
		require.NoError(t, err)
		require.NotZero(t, totals.Online.Money.Raw)

		// as configured above, only the first two accounts should be online
		require.True(t, lookup(t, dl.generator, addrs[0]).Status == basics.Online)
		require.True(t, lookup(t, dl.generator, addrs[1]).Status == basics.Online)
		require.False(t, lookup(t, dl.generator, addrs[2]).Status == basics.Online)
		// genesis accounts don't begin IncentiveEligible, even if online
		require.False(t, lookup(t, dl.generator, addrs[0]).IncentiveEligible)
		require.False(t, lookup(t, dl.generator, addrs[1]).IncentiveEligible)
		require.False(t, lookup(t, dl.generator, addrs[2]).IncentiveEligible)

		dl.fullBlock()

		// although it's not even online, we'll use addrs[7] as the proposer
		proposer := addrs[7]
		dl.beginBlock()
		dl.txns(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[1],
			Receiver: addrs[2],
			Amount:   100_000,
		})
		dl.endBlock(proposer)

		// no changes until the next block
		prp := lookup(t, dl.validator, proposer)
		require.Zero(t, prp.LastProposed)
		require.Zero(t, prp.LastHeartbeat)
		require.False(t, prp.IncentiveEligible)

		// addr[1] is spent to an offline account, so Online totals decrease
		newtotals, err := dl.generator.Totals(dl.generator.Latest())
		require.NoError(t, err)
		// payment and fee left the online account
		require.Equal(t, totals.Online.Money.Raw-100_000-1000, newtotals.Online.Money.Raw)
		totals = newtotals

		dl.fullBlock()

		prp = lookup(t, dl.validator, proposer)
		require.False(t, prp.IncentiveEligible)

		if ver >= checkingBegins {
			// version sanity check
			require.NotZero(t, prp.LastProposed)
			require.Zero(t, prp.LastHeartbeat) // genesis participants have never hb
		} else {
			require.Zero(t, prp.LastProposed)
			require.Zero(t, prp.LastHeartbeat)
		}

		// addrs[2] was already offline
		dl.txns(&txntest.Txn{Type: "keyreg", Sender: addrs[2]}) // OFFLINE keyreg
		regger := lookup(t, dl.validator, addrs[2])

		// total were unchanged by an offline keyreg from an offline account
		newtotals, err = dl.generator.Totals(dl.generator.Latest())
		require.NoError(t, err)
		require.Equal(t, totals.Online.Money.Raw, newtotals.Online.Money.Raw)

		// an an offline keyreg transaction records no activity
		require.Zero(t, regger.LastProposed)
		require.Zero(t, regger.LastHeartbeat)

		// ONLINE keyreg without extra fee
		dl.txns(&txntest.Txn{
			Type:        "keyreg",
			Sender:      addrs[2],
			VotePK:      [32]byte{1},
			SelectionPK: [32]byte{1},
		})
		// online totals have grown
		newtotals, err = dl.generator.Totals(dl.generator.Latest())
		require.NoError(t, err)
		require.Greater(t, newtotals.Online.Money.Raw, totals.Online.Money.Raw)

		regger = lookup(t, dl.validator, addrs[2])
		require.Zero(t, regger.LastProposed)
		require.True(t, regger.Status == basics.Online)

		if ver >= checkingBegins {
			require.NotZero(t, regger.LastHeartbeat) // online keyreg caused update
		} else {
			require.Zero(t, regger.LastHeartbeat)
		}
		require.False(t, regger.IncentiveEligible)

		// ONLINE keyreg with extra fee
		dl.txns(&txntest.Txn{
			Type:        "keyreg",
			Fee:         2_000_000,
			Sender:      addrs[2],
			VotePK:      [32]byte{1},
			SelectionPK: [32]byte{1},
		})

		regger = lookup(t, dl.validator, addrs[2])
		require.Equal(t, ver >= checkingBegins, regger.IncentiveEligible)

		for i := 0; i < 5; i++ {
			dl.fullBlock()
			require.True(t, lookup(t, dl.generator, addrs[0]).Status == basics.Online)
			require.True(t, lookup(t, dl.generator, addrs[1]).Status == basics.Online)
			require.True(t, lookup(t, dl.generator, addrs[2]).Status == basics.Online)
		}

		// all are still online after a few blocks
		require.True(t, lookup(t, dl.generator, addrs[0]).Status == basics.Online)
		require.True(t, lookup(t, dl.generator, addrs[1]).Status == basics.Online)
		require.True(t, lookup(t, dl.generator, addrs[2]).Status == basics.Online)

		for i := 0; i < 30; i++ {
			dl.fullBlock()
		}

		// addrs 0-2 all have about 1/3 of stake, so seemingly (see next block
		// of checks) become eligible for suspension after 30 rounds. We're at
		// about 35. But, since blocks are empty, nobody's suspendible account
		// is noticed.
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[0]).Status)
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[1]).Status)
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[2]).Status)
		require.Equal(t, ver >= checkingBegins, lookup(t, dl.generator, addrs[2]).IncentiveEligible)

		// when 2 pays 0, they both get noticed but addr[0] is not considered absent
		vb := dl.fullBlock(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[2],
			Receiver: addrs[0],
			Amount:   0,
		})
		if ver >= checkingBegins {
			require.Equal(t, vb.Block().AbsentParticipationAccounts, []basics.Address{addrs[2]})
		}
		// addr[0] has never proposed or heartbeat so it is not considered absent
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[0]).Status)
		// addr[1] still hasn't been "noticed"
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[1]).Status)
		require.Equal(t, ver >= checkingBegins, lookup(t, dl.generator, addrs[2]).Status == basics.Offline)
		require.False(t, lookup(t, dl.generator, addrs[2]).IncentiveEligible)

		// now, when 2 pays 1, 1 gets suspended (unlike 0, we had 1 keyreg early on, so LastHeartbeat>0)
		vb = dl.fullBlock(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[2],
			Receiver: addrs[1],
			Amount:   0,
		})
		if ver >= checkingBegins {
			require.Equal(t, vb.Block().AbsentParticipationAccounts, []basics.Address{addrs[1]})
		}
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[0]).Status)
		require.Equal(t, ver >= checkingBegins, lookup(t, dl.generator, addrs[1]).Status == basics.Offline)
		require.False(t, lookup(t, dl.generator, addrs[1]).IncentiveEligible)
		require.Equal(t, ver >= checkingBegins, lookup(t, dl.generator, addrs[2]).Status == basics.Offline)
		require.False(t, lookup(t, dl.generator, addrs[2]).IncentiveEligible)

		// now, addrs[2] proposes, so it gets back online, but not immediately,
		// because processing happens after the proposing block
		dl.proposer = addrs[2]
		dl.fullBlock()
		require.Equal(t, ver >= checkingBegins, lookup(t, dl.generator, addrs[2]).Status == basics.Offline)
		dl.fullBlock()
		require.Equal(t, basics.Online, lookup(t, dl.generator, addrs[2]).Status)
	})
}

// TestAbsenteeChallenges ensures that online accounts that don't (do) respond
// to challenges end up off (on) line.
func TestAbsenteeChallenges(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis(func(cfg *ledgertesting.GenesisCfg) {
		cfg.OnlineCount = 5 // Make online stake big, so these accounts won't be expected to propose
	})
	checkingBegins := 40

	ledgertesting.TestConsensusRange(t, checkingBegins-1, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		// We'll generate a challenge for accounts that start with 0xaa.
		propguy := basics.Address{0xaa, 0xaa, 0xaa} // Will propose during the challenge window
		regguy := basics.Address{0xaa, 0xbb, 0xbb}  // Will re-reg during the challenge window
		badguy := basics.Address{0xaa, 0x11, 0x11}  // Will ignore the challenge

		// Fund them all and have them go online. That makes them eligible to be challenged
		for i, guy := range []basics.Address{propguy, regguy, badguy} {
			dl.txns(&txntest.Txn{
				Type:     "pay",
				Sender:   addrs[0],
				Receiver: guy,
				Amount:   10_000_000,
			}, &txntest.Txn{
				Type:        "keyreg",
				Fee:         5_000_000, // enough to be incentive eligible
				Sender:      guy,
				VotePK:      [32]byte{byte(i + 1)},
				SelectionPK: [32]byte{byte(i + 1)},
			})
			acct := lookup(t, dl.generator, guy)
			require.Equal(t, basics.Online, acct.Status)
			require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible, guy)
		}

		for vb := dl.fullBlock(); vb.Block().Round() < 999; vb = dl.fullBlock() {
			// we just advancing to one before the challenge round
		}
		// All still online, same eligibility
		for _, guy := range []basics.Address{propguy, regguy, badguy} {
			acct := lookup(t, dl.generator, guy)
			require.Equal(t, basics.Online, acct.Status)
			require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible, guy)
		}
		// make the BlockSeed start with 0xa in the challenge round
		dl.beginBlock()
		dl.endBlock(basics.Address{0xaa}) // This becomes the seed, which is used for the challenge

		for vb := dl.fullBlock(); vb.Block().Round() < 1200; vb = dl.fullBlock() {
			// advance through first grace period
		}
		dl.beginBlock()
		dl.endBlock(propguy) // propose, which is a fine (though less likely) way to respond

		// All still online, unchanged eligibility
		for _, guy := range []basics.Address{propguy, regguy, badguy} {
			acct := lookup(t, dl.generator, guy)
			require.Equal(t, basics.Online, acct.Status)
			require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible, guy)
		}

		for vb := dl.fullBlock(); vb.Block().Round() < 1220; vb = dl.fullBlock() {
			// advance into knockoff period. but no transactions means
			// unresponsive accounts go unnoticed.
		}
		// All still online, same eligibility
		for _, guy := range []basics.Address{propguy, regguy, badguy} {
			acct := lookup(t, dl.generator, guy)
			require.Equal(t, basics.Online, acct.Status)
			require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible, guy)
		}

		// badguy never responded, he gets knocked off when paid
		vb := dl.fullBlock(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[0],
			Receiver: badguy,
		})
		if ver >= checkingBegins {
			require.Equal(t, vb.Block().AbsentParticipationAccounts, []basics.Address{badguy})
		}
		acct := lookup(t, dl.generator, badguy)
		require.Equal(t, ver >= checkingBegins, basics.Offline == acct.Status) // if checking, badguy fails
		require.False(t, acct.IncentiveEligible)

		// propguy proposed during the grace period, he stays on even when paid
		dl.txns(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[0],
			Receiver: propguy,
		})
		acct = lookup(t, dl.generator, propguy)
		require.Equal(t, basics.Online, acct.Status)
		require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible)

		// regguy keyregs before he's caught, which is a heartbeat, he stays on as well
		dl.txns(&txntest.Txn{
			Type:        "keyreg", // Does not pay extra fee, since he's still eligible
			Sender:      regguy,
			VotePK:      [32]byte{1},
			SelectionPK: [32]byte{1},
		})
		acct = lookup(t, dl.generator, regguy)
		require.Equal(t, basics.Online, acct.Status)
		require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible)
		dl.txns(&txntest.Txn{
			Type:     "pay",
			Sender:   addrs[0],
			Receiver: regguy,
		})
		acct = lookup(t, dl.generator, regguy)
		require.Equal(t, basics.Online, acct.Status)
		require.Equal(t, ver >= checkingBegins, acct.IncentiveEligible)
	})
}

// TestHoldingGet tests some of the corner cases for the asset_holding_get
// opcode: the asset doesn't exist, the account doesn't exist, account not opted
// in, vs it has none of the asset. This is tested here, even though it should
// be well tested in 'logic' package, because we want to make sure that errors
// come out of the real ledger in the way that the logic package expects (it
// uses a mock ledger for testing).
func TestHoldingGet(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	// 24 is first version with apps
	ledgertesting.TestConsensusRange(t, 24, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		makegold := txntest.Txn{
			Type:   protocol.AssetConfigTx,
			Sender: addrs[0],
			AssetParams: basics.AssetParams{
				Total:     10,
				UnitName:  "gold",
				AssetName: "oz",
			},
		}

		// written without assert or swap, so we can use teal v2 and test back to consensus v24
		source := `
#pragma version 2
txn ApplicationID
bnz main
int 1; return
main:
 txn NumAccounts				// Sender, or Accounts[n]
 txn ApplicationArgs 0; btoi
 asset_holding_get AssetBalance
 txn ApplicationArgs 1; btoi; ==; bz bad
 txn ApplicationArgs 2; btoi; ==; return
bad: err
`

		// Advance the ledger so that there's ambiguity of asset index or foreign array index
		for i := 0; i < 10; i++ {
			dl.fullBlock(&txntest.Txn{Type: "pay", Sender: addrs[2], Receiver: addrs[2]})
		}

		create := txntest.Txn{
			Type:            protocol.ApplicationCallTx,
			Sender:          addrs[0],
			ApprovalProgram: source,
		}

		vb := dl.fullBlock(&create) // create the app
		checker := basics.AppIndex(vb.Block().TxnCounter)
		gold := basics.AssetIndex(checker + 2) // doesn't exist yet
		goldBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(goldBytes, uint64(gold))

		check := txntest.Txn{
			Type:            protocol.ApplicationCallTx,
			Sender:          addrs[0],
			ApplicationID:   checker,
			ApplicationArgs: [][]byte{goldBytes, {0}, {0}}, // exist=0 value=0
		}

		dl.fullBlock(&check)
		vb = dl.fullBlock(&makegold) // Works, despite asset not existing
		require.EqualValues(t, gold, vb.Block().TxnCounter)

		// confirm hardcoded "gold" is correct
		b, ok := holding(t, dl.generator, addrs[0], gold)
		require.True(t, ok)
		require.EqualValues(t, 10, b)

		// The asset exists now. asset_holding_get gives 1,10 for the creator
		// (who is auto-opted in)
		check.ApplicationArgs = [][]byte{goldBytes, {1}, {10}} // exist=1 value=10
		dl.fullBlock(&check)

		// but still gives 0,0 for un opted-in addrs[1], because it means
		// "exists" in the given account, i.e. opted in
		check.Sender = addrs[1]
		check.ApplicationArgs = [][]byte{goldBytes, {0}, {0}}
		dl.fullBlock(&check)

		// opt-in addr[1]
		dl.fullBlock(&txntest.Txn{Type: "axfer", XferAsset: gold, Sender: addrs[1], AssetReceiver: addrs[1]})
		check.ApplicationArgs = [][]byte{goldBytes, {1}, {0}}
		dl.fullBlock(&check)

		// non-existent account, with existing asset, cleanly reports exists=0, value=0
		check.Accounts = []basics.Address{{0x01, 0x02}}
		check.ApplicationArgs = [][]byte{goldBytes, {0}, {0}}
		dl.fullBlock(&check)
	})
}

// TestLocalGetEx tests some of the corner cases for the app_local_get_ex
// opcode: the app doesn't exist, the account doesn't exist, account not opted
// in, local key doesn't exists. This is tested here, even though it should be
// well tested in 'logic' package, because we want to make sure that errors come
// out of the real ledger in the way that the logic package expects (it uses a
// mock ledger for testing).
func TestLocalGetEx(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	// 24 is first version with apps
	ledgertesting.TestConsensusRange(t, 24, 0, func(t *testing.T, ver int, cv protocol.ConsensusVersion, cfg config.Local) {
		dl := NewDoubleLedger(t, genBalances, cv, cfg)
		defer dl.Close()

		makeapp := txntest.Txn{
			Type:   "appl",
			Sender: addrs[0],
			LocalStateSchema: basics.StateSchema{
				NumUint: 1,
			},
			GlobalStateSchema: basics.StateSchema{
				NumByteSlice: 3,
			},
		}

		// written without assert or swap, so we can use teal v2 and test back to consensus v24
		source := `
#pragma version 2
txn ApplicationID
bnz main
int 1; return
main:
 txn NumAccounts				// Sender, or Accounts[n]
 txn ApplicationArgs 0; btoi
 byte "KEY"
 app_local_get_ex
 txn ApplicationArgs 1; btoi; ==; bz bad
 txn ApplicationArgs 2; btoi; ==; return
bad: err
`

		// Advance the ledger so that there's no ambiguity of app ID or foreign array slot
		for i := 0; i < 10; i++ {
			dl.fullBlock(&txntest.Txn{Type: "pay", Sender: addrs[2], Receiver: addrs[2]})
		}

		create := txntest.Txn{
			Type:            protocol.ApplicationCallTx,
			Sender:          addrs[0],
			ApprovalProgram: source,
		}

		vb := dl.fullBlock(&create) // create the checker app
		// Since we are testing back to v24, we can't get appID from EvalDelta
		checker := basics.AppIndex(vb.Block().TxnCounter)
		state := checker + 1 // doesn't exist yet
		stateBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(stateBytes, uint64(state))
		check := txntest.Txn{
			Type:            protocol.ApplicationCallTx,
			Sender:          addrs[0],
			ApplicationID:   checker,
			ApplicationArgs: [][]byte{stateBytes, {0}, {0}}, // exist=0 value=0
		}

		// unlike assets, you can't even do `app_local_get_ex` for an address
		// that has not been opted into the app.  For local state, the existence
		// bit is only used to distinguish "key existence". The local state
		// bundle MUST exist or the program fails.
		dl.txn(&check, "cannot fetch key")

		// so we make the app and try again
		dl.fullBlock(&makeapp)
		// confirm hardcoded "state" index is correct
		g, ok := globals(t, dl.generator, addrs[0], state)
		require.True(t, ok)
		require.EqualValues(t, 3, g.GlobalStateSchema.NumByteSlice)

		// still no good, because creating an app does not opt in the creator
		dl.txn(&check, "cannot fetch key")

		// opt-in addr[0]
		dl.fullBlock(&txntest.Txn{Type: "appl", ApplicationID: state, Sender: addrs[0], OnCompletion: transactions.OptInOC})
		check.ApplicationArgs = [][]byte{stateBytes, {0}, {0}}
		dl.fullBlock(&check)
	})
}

func TestRekeying(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	// Bring up a ledger
	genesisInitState, addrs, keys := ledgertesting.Genesis(10)

	l, err := OpenLedger(logging.TestingLog(t), t.Name(), true, genesisInitState, config.GetDefaultLocal())
	require.NoError(t, err)
	defer l.Close()

	// Make a new block
	nextRound := l.Latest() + basics.Round(1)
	genHash := l.GenesisHash()

	// Test plan
	// Syntax: [A -> B][C, D] means transaction from A that rekeys to B with authaddr C and actual sig from D
	makeTxn := func(sender, rekeyto, authaddr basics.Address, signer *crypto.SignatureSecrets, uniq uint8) transactions.SignedTxn {
		txn := transactions.Transaction{
			Type: protocol.PaymentTx,
			Header: transactions.Header{
				Sender:      sender,
				Fee:         minFee,
				FirstValid:  nextRound,
				LastValid:   nextRound,
				GenesisHash: genHash,
				RekeyTo:     rekeyto,
				Note:        []byte{uniq},
			},
			PaymentTxnFields: transactions.PaymentTxnFields{
				Receiver: sender,
			},
		}
		sig := signer.Sign(txn)
		return transactions.SignedTxn{Txn: txn, Sig: sig, AuthAddr: authaddr}
	}

	tryBlock := func(stxns []transactions.SignedTxn) error {
		// We'll make a block using the evaluator.
		// When generating a block, the evaluator doesn't check transaction sigs -- it assumes the transaction pool already did that.
		// So the ValidatedBlock that comes out isn't necessarily actually a valid block. We'll call Validate ourselves.
		genesisHdr, err := l.BlockHdr(basics.Round(0))
		require.NoError(t, err)
		newBlock := bookkeeping.MakeBlock(genesisHdr)
		eval, err := l.StartEvaluator(newBlock.BlockHeader, 0, 0, nil)
		require.NoError(t, err)

		for _, stxn := range stxns {
			err = eval.Transaction(stxn, transactions.ApplyData{})
			if err != nil {
				return err
			}
		}
		validatedBlock, err := eval.GenerateBlock()
		if err != nil {
			return err
		}

		backlogPool := execpool.MakeBacklog(nil, 0, execpool.LowPriority, nil)
		defer backlogPool.Shutdown()
		_, err = l.Validate(context.Background(), validatedBlock.Block(), backlogPool)
		return err
	}

	// Preamble transactions, which all of the blocks in this test will start with
	// [A -> 0][0,A] (normal transaction)
	// [A -> B][0,A] (rekey)
	txn0 := makeTxn(addrs[0], basics.Address{}, basics.Address{}, keys[0], 0) // Normal transaction
	txn1 := makeTxn(addrs[0], addrs[1], basics.Address{}, keys[0], 1)         // Rekey transaction

	// Test 1: Do only good things
	// (preamble)
	// [A -> 0][B,B] (normal transaction using new key)
	// [A -> A][B,B] (rekey back to A, transaction still signed by B)
	// [A -> 0][0,A] (normal transaction again)
	test1txns := []transactions.SignedTxn{
		txn0, txn1, // (preamble)
		makeTxn(addrs[0], basics.Address{}, addrs[1], keys[1], 2),         // [A -> 0][B,B]
		makeTxn(addrs[0], addrs[0], addrs[1], keys[1], 3),                 // [A -> A][B,B]
		makeTxn(addrs[0], basics.Address{}, basics.Address{}, keys[0], 4), // [A -> 0][0,A]
	}
	err = tryBlock(test1txns)
	require.NoError(t, err)

	// Test 2: Use old key after rekeying
	// (preamble)
	// [A -> A][0,A] (rekey back to A, but signed by A instead of B)
	test2txns := []transactions.SignedTxn{
		txn0, txn1, // (preamble)
		makeTxn(addrs[0], addrs[0], basics.Address{}, keys[0], 2), // [A -> A][0,A]
	}
	err = tryBlock(test2txns)
	require.Error(t, err)

	// TODO: More tests
}

func testEvalAppPoolingGroup(t *testing.T, schema basics.StateSchema, approvalProgram string, consensusVersion protocol.ConsensusVersion) error {
	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	cfg := config.GetDefaultLocal()
	l := newSimpleLedgerWithConsensusVersion(t, genBalances, consensusVersion, cfg)
	defer l.Close()

	eval := nextBlock(t, l)

	appcall1 := txntest.Txn{
		Sender:            addrs[0],
		Type:              protocol.ApplicationCallTx,
		GlobalStateSchema: schema,
		ApprovalProgram:   approvalProgram,
	}

	appcall2 := txntest.Txn{
		Sender:        addrs[0],
		Type:          protocol.ApplicationCallTx,
		ApplicationID: basics.AppIndex(1),
	}

	appcall3 := txntest.Txn{
		Sender:        addrs[1],
		Type:          protocol.ApplicationCallTx,
		ApplicationID: basics.AppIndex(1),
	}

	return txgroup(t, l, eval, &appcall1, &appcall2, &appcall3)
}

// TestEvalAppPooledBudgetWithTxnGroup ensures 3 app call txns can successfully pool
// budgets in a group txn and return an error if the budget is exceeded
func TestEvalAppPooledBudgetWithTxnGroup(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	source := func(n int, m int) string {
		return "#pragma version 4\nbyte 0x1337BEEF\n" + strings.Repeat("keccak256\n", n) +
			strings.Repeat("substring 0 4\n", m) + "pop\nint 1\n"
	}

	params := []protocol.ConsensusVersion{
		protocol.ConsensusV29,
		protocol.ConsensusFuture,
	}

	cases := []struct {
		prog                 string
		isSuccessV29         bool
		isSuccessVFuture     bool
		expectedErrorV29     string
		expectedErrorVFuture string
	}{
		{source(5, 47), true, true,
			"",
			""},
		{source(5, 48), false, true,
			"pc=157 dynamic cost budget exceeded, executing pushint",
			""},
		{source(16, 17), false, true,
			"pc= 12 dynamic cost budget exceeded, executing keccak256",
			""},
		{source(16, 18), false, false,
			"pc= 12 dynamic cost budget exceeded, executing keccak256",
			"pc= 78 dynamic cost budget exceeded, executing pushint"},
	}

	for i, param := range params {
		for j, testCase := range cases {
			t.Run(fmt.Sprintf("i=%d,j=%d", i, j), func(t *testing.T) {
				err := testEvalAppPoolingGroup(t, basics.StateSchema{NumByteSlice: 3}, testCase.prog, param)
				if !testCase.isSuccessV29 && reflect.DeepEqual(param, protocol.ConsensusV29) {
					require.Error(t, err)
					require.Contains(t, err.Error(), testCase.expectedErrorV29)
				} else if !testCase.isSuccessVFuture && reflect.DeepEqual(param, protocol.ConsensusFuture) {
					require.Error(t, err)
					require.Contains(t, err.Error(), testCase.expectedErrorVFuture)
				}
			})
		}
	}
}

func TestMinBalanceChanges(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	cfg := config.GetDefaultLocal()
	l := newSimpleLedgerWithConsensusVersion(t, genBalances, protocol.ConsensusCurrentVersion, cfg)
	defer l.Close()

	createTxn := txntest.Txn{
		Type:   "acfg",
		Sender: addrs[0],
		AssetParams: basics.AssetParams{
			Total:    3,
			Manager:  addrs[1],
			Reserve:  addrs[2],
			Freeze:   addrs[3],
			Clawback: addrs[4],
		},
	}

	const expectedID basics.AssetIndex = 1001
	optInTxn := txntest.Txn{
		Type:          "axfer",
		Sender:        addrs[5],
		XferAsset:     expectedID,
		AssetReceiver: addrs[5],
	}

	ad0init, _, _, err := l.LookupLatest(addrs[0])
	require.NoError(t, err)
	ad5init, _, _, err := l.LookupLatest(addrs[5])
	require.NoError(t, err)

	eval := nextBlock(t, l)
	txns(t, l, eval, &createTxn, &optInTxn)
	endBlock(t, l, eval)

	ad0new, _, _, err := l.LookupLatest(addrs[0])
	require.NoError(t, err)
	ad5new, _, _, err := l.LookupLatest(addrs[5])
	require.NoError(t, err)

	proto := l.GenesisProto()
	// Check balance and min balance requirement changes
	require.Equal(t, ad0init.MicroAlgos.Raw, ad0new.MicroAlgos.Raw+1000)                   // fee
	require.Equal(t, ad0init.MinBalance(&proto).Raw, ad0new.MinBalance(&proto).Raw-100000) // create
	require.Equal(t, ad5init.MicroAlgos.Raw, ad5new.MicroAlgos.Raw+1000)                   // fee
	require.Equal(t, ad5init.MinBalance(&proto).Raw, ad5new.MinBalance(&proto).Raw-100000) // optin

	optOutTxn := txntest.Txn{
		Type:          "axfer",
		Sender:        addrs[5],
		XferAsset:     expectedID,
		AssetReceiver: addrs[0],
		AssetCloseTo:  addrs[0],
	}

	closeTxn := txntest.Txn{
		Type:        "acfg",
		Sender:      addrs[1], // The manager, not the creator
		ConfigAsset: expectedID,
	}

	eval = nextBlock(t, l)
	txns(t, l, eval, &optOutTxn, &closeTxn)
	endBlock(t, l, eval)

	ad0final, _, _, err := l.LookupLatest(addrs[0])
	require.NoError(t, err)
	ad5final, _, _, err := l.LookupLatest(addrs[5])
	require.NoError(t, err)
	// Check we got our balance "back"
	require.Equal(t, ad0final.MinBalance(&proto), ad0init.MinBalance(&proto))
	require.Equal(t, ad5final.MinBalance(&proto), ad5init.MinBalance(&proto))
}

// TestAppInsMinBalance checks that accounts with MaxAppsOptedIn are accepted by block evaluator
// and do not cause any MaximumMinimumBalance problems
func TestAppInsMinBalance(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	genBalances, addrs, _ := ledgertesting.NewTestGenesis()
	cfg := config.GetDefaultLocal()
	l := newSimpleLedgerWithConsensusVersion(t, genBalances, protocol.ConsensusV30, cfg)
	defer l.Close()

	const appID basics.AppIndex = 1

	maxAppsOptedIn := config.Consensus[protocol.ConsensusV30].MaxAppsOptedIn
	require.Greater(t, maxAppsOptedIn, 0)
	maxAppsCreated := config.Consensus[protocol.ConsensusV30].MaxAppsCreated
	require.Greater(t, maxAppsCreated, 0)
	maxLocalSchemaEntries := config.Consensus[protocol.ConsensusV30].MaxLocalSchemaEntries
	require.Greater(t, maxLocalSchemaEntries, uint64(0))

	txnsCreate := make([]*txntest.Txn, 0, maxAppsOptedIn)
	txnsOptIn := make([]*txntest.Txn, 0, maxAppsOptedIn)
	appsCreated := make(map[basics.Address]int, len(addrs)-1)

	acctIdx := 0
	for i := 0; i < maxAppsOptedIn; i++ {
		creator := addrs[acctIdx]
		createTxn := txntest.Txn{
			Type:             protocol.ApplicationCallTx,
			Sender:           creator,
			ApprovalProgram:  "int 1",
			LocalStateSchema: basics.StateSchema{NumByteSlice: maxLocalSchemaEntries},
			Note:             ledgertesting.RandomNote(),
		}
		txnsCreate = append(txnsCreate, &createTxn)
		count := appsCreated[creator]
		count++
		appsCreated[creator] = count
		if count == maxAppsCreated {
			acctIdx++
		}

		optInTxn := txntest.Txn{
			Type:          protocol.ApplicationCallTx,
			Sender:        addrs[9],
			ApplicationID: appID + basics.AppIndex(i),
			OnCompletion:  transactions.OptInOC,
		}
		txnsOptIn = append(txnsOptIn, &optInTxn)
	}

	eval := nextBlock(t, l)
	txns1 := append(txnsCreate, txnsOptIn...)
	txns(t, l, eval, txns1...)
	vb := endBlock(t, l, eval)
	mods := vb.Delta()
	appAppResources := mods.Accts.GetAllAppResources()
	appParamsCount := 0
	appLocalStatesCount := 0
	for _, ap := range appAppResources {
		if ap.Params.Params != nil {
			appParamsCount++
		}
		if ap.State.LocalState != nil {
			appLocalStatesCount++
		}
	}
	require.Equal(t, appLocalStatesCount, 50)
	require.Equal(t, appParamsCount, 50)
}
