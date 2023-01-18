// Copyright (C) 2019-2023 Algorand, Inc.
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

package logic_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/algorand/go-algorand/data/txntest"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
)

func TestNewAppEvalParams(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	params := []config.ConsensusParams{
		{Application: true, MaxAppProgramCost: 700},
		config.Consensus[protocol.ConsensusV29],
		config.Consensus[protocol.ConsensusFuture],
	}

	// Create some sample transactions. The main reason this a blackbox test
	// (_test package) is to have access to txntest.
	payment := txntest.Txn{
		Type:     protocol.PaymentTx,
		Sender:   basics.Address{1, 2, 3, 4},
		Receiver: basics.Address{4, 3, 2, 1},
		Amount:   100,
	}.SignedTxnWithAD()

	appcall1 := txntest.Txn{
		Type:          protocol.ApplicationCallTx,
		Sender:        basics.Address{1, 2, 3, 4},
		ApplicationID: basics.AppIndex(1),
	}.SignedTxnWithAD()

	appcall2 := appcall1
	appcall2.Txn.ApplicationID = basics.AppIndex(2)

	type evalTestCase struct {
		group       []transactions.SignedTxnWithAD
		numAppCalls int
	}

	// Create some groups with these transactions
	cases := []evalTestCase{
		{[]transactions.SignedTxnWithAD{payment}, 0},
		{[]transactions.SignedTxnWithAD{appcall1}, 1},
		{[]transactions.SignedTxnWithAD{payment, payment}, 0},
		{[]transactions.SignedTxnWithAD{appcall1, payment}, 1},
		{[]transactions.SignedTxnWithAD{payment, appcall1}, 1},
		{[]transactions.SignedTxnWithAD{appcall1, appcall2}, 2},
		{[]transactions.SignedTxnWithAD{appcall1, appcall2, appcall1}, 3},
		{[]transactions.SignedTxnWithAD{payment, appcall1, payment}, 1},
		{[]transactions.SignedTxnWithAD{appcall1, payment, appcall2}, 2},
	}

	for i, param := range params {
		param := param
		for j, testCase := range cases {
			i, j, param, testCase := i, j, param, testCase
			t.Run(fmt.Sprintf("i=%d,j=%d", i, j), func(t *testing.T) {
				t.Parallel()
				ep := logic.NewEvalParams(testCase.group, &param, nil)
				require.NotNil(t, ep)
				require.Equal(t, ep.TxnGroup, testCase.group)
				require.Equal(t, *ep.Proto, param)
				if reflect.DeepEqual(param, config.Consensus[protocol.ConsensusV29]) || testCase.numAppCalls == 0 {
					require.Nil(t, ep.PooledApplicationBudget)
				} else if reflect.DeepEqual(param, config.Consensus[protocol.ConsensusFuture]) {
					require.Equal(t, *ep.PooledApplicationBudget, param.MaxAppProgramCost*testCase.numAppCalls)
				}
			})
		}
	}
}

// TestAppSharing confirms that as of v9, assets can be accessed across
// groups, but that before then, they could not.
func TestAppSharing(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	// Create some sample transactions. The main reason this a blackbox test
	// (_test package) is to have access to txntest.
	appl0 := txntest.Txn{
		Type:        protocol.ApplicationCallTx,
		Sender:      basics.Address{1, 2, 3, 4},
		ForeignApps: []basics.AppIndex{500},
	}

	appl1 := txntest.Txn{
		Type:   protocol.ApplicationCallTx,
		Sender: basics.Address{4, 3, 2, 1},
	}

	appl2 := txntest.Txn{
		Type:   protocol.ApplicationCallTx,
		Sender: basics.Address{1, 2, 3, 4},
	}

	getSchema := `
int 500
app_params_get AppGlobalNumByteSlice
!; assert; pop; int 1
`
	sources := []string{getSchema, getSchema}
	// In v8, the first tx can read app params of 500, because it's in its
	// foreign array, but the second can't
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid App reference 500"))
	// In v9, the second can, because the first can.
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 9, nil)

	getLocalEx := `
int 0							// Sender
int 500
byte "some-key"
app_local_get_ex
pop; pop; int 1
`

	sources = []string{getLocalEx, getLocalEx}
	// In contrast, here there's no help from v9, because the second tx is
	// reading the locals for a different account.

	// app_local_get* requires the address and the app exist, else the program fails
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 8, nil,
		logic.NewExpect(0, "no account"))

	_, _, ledger := logic.MakeSampleEnv()
	ledger.NewAccount(appl0.Sender, 100_000)
	ledger.NewApp(appl0.Sender, 500, basics.AppParams{})
	ledger.NewLocals(appl0.Sender, 500) // opt in
	// Now txn0 passes, but txn1 has an error because it can't see app 500
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 9, ledger,
		logic.NewExpect(1, "invalid Local State access"))

	// But it's ok in appl2, because appl2 uses the same Sender, even though the
	// foreign-app is not repeated in appl2 so the holding being accessed is is
	// the one from tx0.
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl2), 9, ledger)
}

// TestAssetSharing confirms that as of v9, assets can be accessed across
// groups, but that before then, they could not.
func TestAssetSharing(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	// Create some sample transactions. The main reason this a blackbox test
	// (_test package) is to have access to txntest.
	appl0 := txntest.Txn{
		Type:          protocol.ApplicationCallTx,
		Sender:        basics.Address{1, 2, 3, 4},
		ForeignAssets: []basics.AssetIndex{400},
	}

	appl1 := txntest.Txn{
		Type:   protocol.ApplicationCallTx,
		Sender: basics.Address{4, 3, 2, 1},
	}

	appl2 := txntest.Txn{
		Type:   protocol.ApplicationCallTx,
		Sender: basics.Address{1, 2, 3, 4},
	}

	getTotal := `
int 400
asset_params_get AssetTotal
pop; pop; int 1
`
	sources := []string{getTotal, getTotal}
	// In v8, the first tx can read asset 400, because it's in its foreign arry,
	// but the second can't
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid Asset reference 400"))
	// In v9, the second can, because the first can.
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 9, nil)

	getBalance := `
int 0
int 400
asset_holding_get AssetBalance
pop; pop; int 1
`

	sources = []string{getBalance, getBalance}
	// In contrast, here there's no help from v9, because the second tx is
	// reading a holding for a different account.
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid Asset reference 400"))
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl1), 9, nil,
		logic.NewExpect(1, "invalid Holding access"))
	// But it's ok in appl2, because the same account is used, even though the
	// foreign-asset is not repeated in appl2.
	logic.TestApps(t, sources, txntest.SignedTxns(&appl0, &appl2), 9, nil)
}
