// Copyright (C) 2019-2022 Algorand, Inc.
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

package v2

import (
	"fmt"
	"strings"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/verify"
	"github.com/algorand/go-algorand/ledger"
)

// ==============================
// > Simulation Ledger
// ==============================

type LedgerForSimulator interface {
	ledger.LedgerForDebugger
}

type apiSimulatorLedgerConnector struct {
	LedgerForAPI
	hdr bookkeeping.BlockHeader
}

// Latest is part of the LedgerForSimulator interface.
// We override this to use the set hdr to prevent racing with the network
func (l apiSimulatorLedgerConnector) Latest() basics.Round {
	return l.hdr.Round
}

// BlockHdr is part of the LedgerForSimulator interface.
// We override this to use the set hdr to prevent racing with the network
func (l apiSimulatorLedgerConnector) BlockHdr(round basics.Round) (bookkeeping.BlockHeader, error) {
	if round != l.Latest() {
		err := fmt.Errorf(
			"BlockHdr() evaluator called this function for the wrong round %d, "+
				"latest round is %d",
			round, l.Latest())
		return bookkeeping.BlockHeader{}, err
	}

	return l.LedgerForAPI.BlockHdr(round)
}

// GenesisHash is part of LedgerForSimulator interface.
func (l apiSimulatorLedgerConnector) GenesisHash() crypto.Digest {
	return l.hdr.GenesisHash
}

// GenesisProto is part of LedgerForSimulator interface.
func (l apiSimulatorLedgerConnector) GenesisProto() config.ConsensusParams {
	return config.Consensus[l.hdr.CurrentProtocol]
}

// GetCreatorForRound is part of LedgerForSimulator interface.
func (l apiSimulatorLedgerConnector) GetCreatorForRound(round basics.Round, cidx basics.CreatableIndex, ctype basics.CreatableType) (creator basics.Address, ok bool, err error) {
	if round != l.Latest() {
		err = fmt.Errorf(
			"GetCreatorForRound() evaluator called this function for the wrong round %d, "+
				"latest round is %d",
			round, l.Latest())
		return
	}

	return l.GetCreator(cidx, ctype)
}

func makeLedgerForSimulatorFromLedgerForAPI(ledgerForAPI LedgerForAPI, hdr bookkeeping.BlockHeader) LedgerForSimulator {
	return &apiSimulatorLedgerConnector{ledgerForAPI, hdr}
}

// ==============================
// > Simulator Errors
// ==============================

type SimulatorError struct {
	error
}

// An invalid transaction group was submitted to the simulator.
type InvalidTxGroupError struct {
	SimulatorError
}

// A scoped simulator error is a simulator error that has 2 errors, one for internal use and one for
// displaying publicly. THe external error is useful for API routes/etc.
type ScopedSimulatorError struct {
	SimulatorError        // the original error for internal use
	External       string // the external error for public use
}

// ==============================
// > Simulator Utility Methods
// ==============================

func isAppBudgetError(err error) bool {
	appBudgetErrorFragments := []string{
		"approval program too long",
		"clear state program too long",
		"app programs too long",
	}

	for _, fragment := range appBudgetErrorFragments {
		if strings.Contains(err.Error(), fragment) {
			return true
		}
	}

	return false
}

// ==============================
// > Simulator
// ==============================

type Simulator struct {
	ledger LedgerForSimulator
}

func MakeSimulator(ledger LedgerForSimulator) *Simulator {
	return &Simulator{
		ledger: ledger,
	}
}

func MakeSimulatorFromAPILedger(ledgerForAPI LedgerForAPI, hdr bookkeeping.BlockHeader) *Simulator {
	ledger := makeLedgerForSimulatorFromLedgerForAPI(ledgerForAPI, hdr)
	return MakeSimulator(ledger)
}

// checkWellFormed checks that the transaction is well-formed. A failure message is returned if the transaction is not well-formed.
func (s Simulator) checkWellFormed(txgroup []transactions.SignedTxn) (failureMessage string, err error) {
	hdr, err := s.ledger.BlockHdr(s.ledger.Latest())
	if err != nil {
		return "", ScopedSimulatorError{SimulatorError{fmt.Errorf("Please contact us, this shouldn't happen. Current block error: %v.", err)}, "current block error"}
	}

	_, err = verify.TxnGroupBatchVerify(txgroup, hdr, nil, nil)
	if err != nil {
		// catch app budget errors, let them through, but mark that the txgroup would fail
		isBudgetError := isAppBudgetError(err)
		if isBudgetError {
			failureMessage = err.Error()
			return
		}

		// catch verifier is nil error. This is an expected error if nothing else goes wrong.
		if strings.Contains(err.Error(), "verifier is nil") {
			return "", nil
		}

		// otherwise the transaction group was invalid in some way and we should return an error
		return "", InvalidTxGroupError{SimulatorError{err}}
	}

	return "", nil
}

// Simulate a transaction group using the simulator. Will error if the transaction group is not well-formed or an
// unexpected error occurs. Otherwise, evaluation failure messages are returned.
func (s Simulator) SimulateSignedTxGroup(txgroup []transactions.SignedTxn) (result generated.SimulationResult, err error) {
	// check that the transaction is well-formed. Signatures are checked after evaluation
	failureMessage, err := s.checkWellFormed(txgroup)
	if err != nil {
		return
	}
	result.FailureMessage = &failureMessage

	_, _, evalErr := ledger.EvalForDebugger(s.ledger, txgroup)
	if evalErr != nil {
		errStr := evalErr.Error()
		result.FailureMessage = &errStr
	}

	return
}