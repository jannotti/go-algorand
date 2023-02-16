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
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/algorand/go-algorand/data/txntest"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/stretchr/testify/require"
)

// TestAppSharing confirms that as of v9, apps can be accessed across
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
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid App reference 500"))
	// In v9, the second can, because the first can.
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, nil)

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
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 8, nil,
		logic.NewExpect(0, "no account"))

	_, _, ledger := logic.MakeSampleEnv()
	ledger.NewAccount(appl0.Sender, 100_000)
	ledger.NewApp(appl0.Sender, 500, basics.AppParams{})
	ledger.NewLocals(appl0.Sender, 500) // opt in
	// Now txn0 passes, but txn1 has an error because it can't see app 500
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, ledger,
		logic.NewExpect(1, "invalid Local State access"))

	// But it's ok in appl2, because appl2 uses the same Sender, even though the
	// foreign-app is not repeated in appl2 because the holding being accessed
	// is the one from tx0.
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl2), 9, ledger)

	// Now, confirm that *setting* a local state in tx1 that was made available
	// in tx0 works.  The extra check here is that the change is recorded
	// properly in EvalDelta.
	putLocal := `
txn ApplicationArgs 0
byte "X"
int 74
app_local_put
int 1
`
	noop := `int 1`
	sources = []string{noop, putLocal}
	appl1.ApplicationArgs = [][]byte{appl0.Sender[:]} // tx1 will try to modify local state exposed in tx0
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, ledger,
		logic.NewExpect(1, "account "+appl0.Sender.String()+" is not opted into 888"))
	ledger.NewLocals(appl0.Sender, 888) // opt in
	ep := logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, ledger)
	require.Len(t, ep.TxnGroup, 2)
	ed := ep.TxnGroup[1].ApplyData.EvalDelta
	require.Len(t, ed.LocalDeltas, 1)
	require.Contains(t, ed.LocalDeltas, uint64(1)) // no tx.Accounts, 1 indicates first in SharedAccts
	sd := ed.LocalDeltas[1]
	require.Len(t, sd, 1)
	require.Contains(t, sd, "X")
	require.EqualValues(t, 74, sd["X"].Uint)
	require.Len(t, ed.SharedAccts, 1)
	require.Equal(t, ep.TxnGroup[0].Txn.Sender, ed.SharedAccts[0])
}

// TestBetterLocalErrors convirms that we get specific errors about the missing
// address or app when accessesing a Local State with only one available.
func TestBetterLocalErrors(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	joe := basics.Address{9, 9, 9}

	ep, tx, ledger := logic.MakeSampleEnv()
	ledger.NewAccount(joe, 5000000)
	ledger.NewApp(joe, 500, basics.AppParams{})
	ledger.NewLocals(joe, 500)

	getLocalEx := `
txn ApplicationArgs 0
txn ApplicationArgs 1; btoi
byte "some-key"
app_local_get_ex
pop; pop; int 1
`
	app := make([]byte, 8)
	binary.BigEndian.PutUint64(app, 500)

	tx.ApplicationArgs = [][]byte{joe[:], app}
	logic.TestApp(t, getLocalEx, ep, "invalid Local State")
	tx.Accounts = []basics.Address{joe}
	logic.TestApp(t, getLocalEx, ep, "invalid App reference 500")
	tx.ForeignApps = []basics.AppIndex{500}
	logic.TestApp(t, getLocalEx, ep)
	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 1)
	logic.TestApp(t, getLocalEx, ep)
	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 2) // beyond the txn.ForeignApps array
	logic.TestApp(t, getLocalEx, ep, "invalid App reference 2")

	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 1)
	tx.Accounts = []basics.Address{}
	logic.TestApp(t, getLocalEx, ep, "invalid Account reference "+joe.String())
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
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid Asset reference 400"))
	// In v9, the second can, because the first can.
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, nil)

	getBalance := `
int 0
int 400
asset_holding_get AssetBalance
pop; pop; int 1
`

	sources = []string{getBalance, getBalance}
	// In contrast, here there's no help from v9, because the second tx is
	// reading a holding for a different account.
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 8, nil,
		logic.NewExpect(1, "invalid Asset reference 400"))
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl1), 9, nil,
		logic.NewExpect(1, "invalid Holding access"))
	// But it's ok in appl2, because the same account is used, even though the
	// foreign-asset is not repeated in appl2.
	logic.TestApps(t, sources, txntest.Group(&appl0, &appl2), 9, nil)
}

// TestBetterHoldingErrors convirms that we get specific errors about the missing
// address or asa when accessesing a holding with only one available.
func TestBetterHoldingErrors(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	joe := basics.Address{9, 9, 9}

	ep, tx, ledger := logic.MakeSampleEnv()
	ledger.NewAccount(joe, 5000000)
	ledger.NewAsset(joe, 200, basics.AssetParams{})
	// as creator, joe will also be opted in

	getHoldingBalance := `
txn ApplicationArgs 0
txn ApplicationArgs 1; btoi
asset_holding_get AssetBalance
pop; pop; int 1
`
	asa := make([]byte, 8)
	binary.BigEndian.PutUint64(asa, 200)

	tx.ApplicationArgs = [][]byte{joe[:], asa}
	logic.TestApp(t, getHoldingBalance, ep, "invalid Holding access "+joe.String())
	tx.Accounts = []basics.Address{joe}
	logic.TestApp(t, getHoldingBalance, ep, "invalid Asset reference 200")
	tx.ForeignAssets = []basics.AssetIndex{200}
	logic.TestApp(t, getHoldingBalance, ep)
	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 0)
	logic.TestApp(t, getHoldingBalance, ep)
	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 1) // beyond the txn.ForeignAssets array
	logic.TestApp(t, getHoldingBalance, ep, "invalid Asset reference 1")

	binary.BigEndian.PutUint64(tx.ApplicationArgs[1], 0)
	tx.Accounts = []basics.Address{}
	logic.TestApp(t, getHoldingBalance, ep, "invalid Account reference "+joe.String())
}

// TestNewAppAccount checks whether a newly created app can put its own address
// into `itxn_field Accounts`.
func TestNewAppAccount(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	ep, tx, ledger := logic.MakeSampleEnv()

	accept := logic.TestProg(t, "int 1", 6)
	alice := basics.Address{1, 1, 1, 1, 1}
	ledger.NewApp(alice, 4, basics.AppParams{
		ApprovalProgram: accept.Program,
	})
	callWithMyAccount := `
itxn_begin
	int appl;	                       itxn_field TypeEnum
    int 4;	                           itxn_field ApplicationID
    global CurrentApplicationAddress;  itxn_field Accounts
itxn_submit
int 1`
	tx.ForeignApps = []basics.AppIndex{4}
	ledger.NewAccount(appAddr(888), 50_000)
	logic.TestApp(t, callWithMyAccount, ep)
}

// TestOtherTxSharing tests resource sharing across other kinds of transactions besides appl.
func TestOtherTxSharing(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	_, _, ledger := logic.MakeSampleEnv()

	senderAcct := basics.Address{1, 2, 3, 4, 5, 6, 1}
	ledger.NewAccount(senderAcct, 2001)
	senderBalance := "txn ApplicationArgs 0; balance; int 2001; =="

	receiverAcct := basics.Address{1, 2, 3, 4, 5, 6, 2}
	ledger.NewAccount(receiverAcct, 2002)
	receiverBalance := "txn ApplicationArgs 0; balance; int 2002; =="

	otherAcct := basics.Address{1, 2, 3, 4, 5, 6, 3}
	ledger.NewAccount(otherAcct, 2003)
	otherBalance := "txn ApplicationArgs 0; balance; int 2003; =="

	appl := txntest.Txn{
		Type:            protocol.ApplicationCallTx,
		Sender:          basics.Address{5, 5, 5, 5}, // different from all other accounts used
		ApplicationArgs: [][]byte{senderAcct[:]},
	}

	keyreg := txntest.Txn{
		Type:   protocol.KeyRegistrationTx,
		Sender: senderAcct,
	}
	pay := txntest.Txn{
		Type:     protocol.PaymentTx,
		Sender:   senderAcct,
		Receiver: receiverAcct,
	}
	acfg := txntest.Txn{
		Type:   protocol.AssetConfigTx,
		Sender: senderAcct,
		AssetParams: basics.AssetParams{
			Manager:  otherAcct, // other is here to show they _don't_ become available
			Reserve:  otherAcct,
			Freeze:   otherAcct,
			Clawback: otherAcct,
		},
	}
	axfer := txntest.Txn{
		Type:          protocol.AssetTransferTx,
		XferAsset:     100, // must be < 256, later code assumes it fits in a byte
		Sender:        senderAcct,
		AssetReceiver: receiverAcct,
		AssetSender:   otherAcct,
	}
	afrz := txntest.Txn{
		Type:          protocol.AssetFreezeTx,
		FreezeAsset:   200, // must be < 256, later code assumes it fits in a byte
		Sender:        senderAcct,
		FreezeAccount: otherAcct,
	}

	sources := []string{"", senderBalance}
	rsources := []string{senderBalance, ""}
	for _, send := range []txntest.Txn{keyreg, pay, acfg, axfer, afrz} {
		logic.TestApps(t, sources, txntest.Group(&send, &appl), 9, ledger)
		logic.TestApps(t, rsources, txntest.Group(&appl, &send), 9, ledger)

		logic.TestApps(t, sources, txntest.Group(&send, &appl), 8, ledger,
			logic.NewExpect(1, "invalid Account reference"))
		logic.TestApps(t, rsources, txntest.Group(&appl, &send), 8, ledger,
			logic.NewExpect(0, "invalid Account reference"))
	}

	holdingAccess := `
	txn ApplicationArgs 0
	txn ApplicationArgs 1; btoi
	asset_holding_get AssetBalance
	pop; pop; int 1
`
	sources = []string{"", holdingAccess}
	rsources = []string{holdingAccess, ""}

	t.Run("keyreg", func(t *testing.T) {
		appl.ApplicationArgs = [][]byte{senderAcct[:], {200}}
		logic.TestApps(t, sources, txntest.Group(&keyreg, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Asset reference 200"))
		withRef := appl
		withRef.ForeignAssets = []basics.AssetIndex{200}
		logic.TestApps(t, sources, txntest.Group(&keyreg, &withRef), 9, ledger,
			logic.NewExpect(1, "invalid Holding access "+senderAcct.String()))
	})
	t.Run("pay", func(t *testing.T) {
		// The receiver is available for algo balance reading
		appl.ApplicationArgs = [][]byte{receiverAcct[:]}
		logic.TestApps(t, []string{"", receiverBalance}, txntest.Group(&pay, &appl), 9, ledger)

		// The other account is not (it's not even in the pay txn)
		appl.ApplicationArgs = [][]byte{otherAcct[:]}
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&pay, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Account reference "+otherAcct.String()))

		// The other account becomes accessible because used in CloseRemainderTo
		withClose := pay
		withClose.CloseRemainderTo = otherAcct
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&withClose, &appl), 9, ledger)
	})

	t.Run("acfg", func(t *testing.T) {
		// The other account is not available even though it's all the extra addresses
		appl.ApplicationArgs = [][]byte{otherAcct[:]}
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&acfg, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Account reference "+otherAcct.String()))
	})

	t.Run("axfer", func(t *testing.T) {
		// The receiver is also available for algo balance reading
		appl.ApplicationArgs = [][]byte{receiverAcct[:]}
		logic.TestApps(t, []string{"", receiverBalance}, txntest.Group(&axfer, &appl), 9, ledger)

		// as is the "other" (AssetSender)
		appl.ApplicationArgs = [][]byte{otherAcct[:]}
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&axfer, &appl), 9, ledger)

		// receiver holding is available
		appl.ApplicationArgs = [][]byte{receiverAcct[:], {byte(axfer.XferAsset)}}
		logic.TestApps(t, []string{"", holdingAccess}, txntest.Group(&axfer, &appl), 9, ledger)

		// The other account becomes accessible because used in CloseRemainderTo
		// (for asa and algo)
		withClose := axfer
		withClose.AssetCloseTo = otherAcct
		appl.ApplicationArgs = [][]byte{otherAcct[:], {byte(axfer.XferAsset)}}
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&withClose, &appl), 9, ledger)
		logic.TestApps(t, []string{"", holdingAccess}, txntest.Group(&withClose, &appl), 9, ledger)
	})

	t.Run("afrz", func(t *testing.T) {
		// The other account is available (for algo and asset)
		appl.ApplicationArgs = [][]byte{otherAcct[:], {byte(afrz.FreezeAsset)}}
		logic.TestApps(t, []string{"", otherBalance}, txntest.Group(&afrz, &appl), 9, ledger)
		logic.TestApps(t, []string{"", holdingAccess}, txntest.Group(&afrz, &appl), 9, ledger)
	})
}

// TestSharedInnerTxns checks how inner txns access resources.
func TestSharedInnerTxns(t *testing.T) {
	_, _, ledger := logic.MakeSampleEnv()

	const asa1 = 201
	const asa2 = 202

	senderAcct := basics.Address{1, 2, 3, 4, 5, 6, 1}
	ledger.NewAccount(senderAcct, 2001)
	ledger.NewHolding(senderAcct, asa1, 1, false)

	receiverAcct := basics.Address{1, 2, 3, 4, 5, 6, 2}
	ledger.NewAccount(receiverAcct, 2002)
	ledger.NewHolding(receiverAcct, asa1, 1, false)

	otherAcct := basics.Address{1, 2, 3, 4, 5, 6, 3}
	ledger.NewAccount(otherAcct, 2003)
	ledger.NewHolding(otherAcct, asa1, 1, false)

	unusedAcct := basics.Address{1, 2, 3, 4, 5, 6, 4}

	payToArg := `
itxn_begin
  int pay;               itxn_field TypeEnum
  int 100;               itxn_field Amount
  txn ApplicationArgs 0; itxn_field Receiver
itxn_submit
int 1
`
	axferToArgs := `
itxn_begin
  int axfer;                   itxn_field TypeEnum
  int 2;                       itxn_field AssetAmount
  txn ApplicationArgs 0;       itxn_field AssetReceiver
  txn ApplicationArgs 1; btoi; itxn_field XferAsset
itxn_submit
int 1
`

	acfgArg := `
itxn_begin
  int acfg;                    itxn_field TypeEnum
  txn ApplicationArgs 0; btoi; itxn_field ConfigAsset
itxn_submit
int 1
`

	appl := txntest.Txn{
		Type:          protocol.ApplicationCallTx,
		ApplicationID: 1234,
		Sender:        basics.Address{5, 5, 5, 5}, // different from all other accounts used
	}
	appAcct := appl.ApplicationID.Address()
	// App will do a lot of txns. Start well funded.
	ledger.NewAccount(appAcct, 1_000_000)
	// And needs some ASAs for inner axfer testing
	ledger.NewHolding(appAcct, asa1, 1_000_000, false)

	t.Run("keyreg", func(t *testing.T) {
		keyreg := txntest.Txn{
			Type:   protocol.KeyRegistrationTx,
			Sender: senderAcct,
		}

		// appl has no foreign ref to senderAcct, but can still inner pay it
		appl.ApplicationArgs = [][]byte{senderAcct[:]}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&keyreg, &appl), 9, ledger)
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&keyreg, &appl), 8, ledger,
			logic.NewExpect(1, "invalid Account reference "+senderAcct.String()))

		// confirm you can't just pay _anybody_. receiverAcct is not in use at all.
		appl.ApplicationArgs = [][]byte{receiverAcct[:]}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&keyreg, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Account reference "+receiverAcct.String()))
	})

	t.Run("pay", func(t *testing.T) {
		pay := txntest.Txn{
			Type:     protocol.PaymentTx,
			Sender:   senderAcct,
			Receiver: receiverAcct,
		}

		// appl has no foreign ref to senderAcct or receiverAcct, but can still inner pay them
		appl.ApplicationArgs = [][]byte{senderAcct[:]}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&pay, &appl), 9, ledger)
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&pay, &appl), 8, ledger,
			logic.NewExpect(1, "invalid Account reference "+senderAcct.String()))

		appl.ApplicationArgs = [][]byte{receiverAcct[:]}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&pay, &appl), 9, ledger)
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&pay, &appl), 8, ledger,
			logic.NewExpect(1, "invalid Account reference "+receiverAcct.String()))

		// confirm you can't just pay _anybody_. otherAcct is not in use at all.
		appl.ApplicationArgs = [][]byte{otherAcct[:]}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&pay, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Account reference "+otherAcct.String()))
	})

	t.Run("axfer", func(t *testing.T) {
		axfer := txntest.Txn{
			Type:          protocol.AssetTransferTx,
			XferAsset:     asa1,
			Sender:        senderAcct,
			AssetReceiver: receiverAcct,
			AssetSender:   otherAcct,
		}

		// appl can pay the axfer sender
		appl.ApplicationArgs = [][]byte{senderAcct[:], {asa1}}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&axfer, &appl), 9, ledger)
		// but can't axfer to sender, because appAcct doesn't have holding access for the asa
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Holding access"))
		// and to the receiver
		appl.ApplicationArgs = [][]byte{receiverAcct[:], {asa1}}
		logic.TestApps(t, []string{payToArg}, txntest.Group(&appl, &axfer), 9, ledger)
		logic.TestApps(t, []string{axferToArgs}, txntest.Group(&appl, &axfer), 9, ledger,
			logic.NewExpect(0, "invalid Holding access"))
		// and to the clawback
		appl.ApplicationArgs = [][]byte{otherAcct[:], {asa1}}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&axfer, &appl), 9, ledger)
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Holding access"))

		// Those axfers become possible by adding the asa to the appl's ForeignAssets
		appl.ForeignAssets = []basics.AssetIndex{asa1}
		appl.ApplicationArgs = [][]byte{senderAcct[:], {asa1}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger)
		appl.ApplicationArgs = [][]byte{receiverAcct[:], {asa1}}
		logic.TestApps(t, []string{axferToArgs}, txntest.Group(&appl, &axfer), 9, ledger)
		appl.ApplicationArgs = [][]byte{otherAcct[:], {asa1}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger)

		// but can't axfer a different asset
		appl.ApplicationArgs = [][]byte{senderAcct[:], {asa2}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger,
			logic.NewExpect(1, fmt.Sprintf("invalid Asset reference %d", asa2)))
		// or correct asset to an unknown address
		appl.ApplicationArgs = [][]byte{unusedAcct[:], {asa1}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&axfer, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Account reference"))

		// appl can acfg the asset from tx0 (which requires asset available, not holding)
		appl.ApplicationArgs = [][]byte{{asa1}}
		logic.TestApps(t, []string{"", acfgArg}, txntest.Group(&axfer, &appl), 9, ledger)
		appl.ApplicationArgs = [][]byte{{asa2}} // but not asa2
		logic.TestApps(t, []string{"", acfgArg}, txntest.Group(&axfer, &appl), 9, ledger,
			logic.NewExpect(1, fmt.Sprintf("invalid Asset reference %d", asa2)))

		// Now, confirm that access to account from a pay in one tx, and asa
		// from another don't allow inner axfer in the third (because there's no
		// access to that payer's holding.)
		payAcct := basics.Address{3, 2, 3, 2, 3, 2}
		pay := txntest.Txn{
			Type:     protocol.PaymentTx,
			Sender:   payAcct,
			Receiver: payAcct,
		}
		// the asset is acfg-able
		appl.ApplicationArgs = [][]byte{{asa1}}
		logic.TestApps(t, []string{"", "", acfgArg}, txntest.Group(&pay, &axfer, &appl), 9, ledger)
		logic.TestApps(t, []string{"", "", acfgArg}, txntest.Group(&axfer, &pay, &appl), 9, ledger)
		// payAcct (the pay sender) is payable
		appl.ApplicationArgs = [][]byte{payAcct[:]}
		logic.TestApps(t, []string{"", "", payToArg}, txntest.Group(&axfer, &pay, &appl), 9, ledger)
		// but the cross-product is not available, so no axfer (opting in first, to prevent that error)
		ledger.NewHolding(payAcct, asa1, 1, false)
		appl.ApplicationArgs = [][]byte{payAcct[:], {asa1}}
		logic.TestApps(t, []string{"", "", axferToArgs}, txntest.Group(&axfer, &pay, &appl), 9, ledger,
			logic.NewExpect(2, "invalid Holding access "+payAcct.String()))
	})

	t.Run("afrz", func(t *testing.T) {
		appl.ForeignAssets = []basics.AssetIndex{} // reset after previous test
		afrz := txntest.Txn{
			Type:          protocol.AssetFreezeTx,
			FreezeAsset:   asa1,
			Sender:        senderAcct,
			FreezeAccount: otherAcct,
		}

		// appl can pay to the sender & freeze account
		appl.ApplicationArgs = [][]byte{senderAcct[:], {asa1}}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&afrz, &appl), 9, ledger)
		appl.ApplicationArgs = [][]byte{otherAcct[:], {asa1}}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&afrz, &appl), 9, ledger)

		// can't axfer to the afrz sender because appAcct holding is not available from afrz
		appl.ApplicationArgs = [][]byte{senderAcct[:], {asa1}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&afrz, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Holding access "+appAcct.String()))
		appl.ForeignAssets = []basics.AssetIndex{asa1}
		// _still_ can't axfer to sender because afrz sender's holding does NOT
		// become available (not note that complaint is now about that account)
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&afrz, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Holding access "+senderAcct.String()))

		// and not to the receiver which isn't in afrz
		appl.ApplicationArgs = [][]byte{receiverAcct[:], {asa1}}
		logic.TestApps(t, []string{payToArg}, txntest.Group(&appl, &afrz), 9, ledger,
			logic.NewExpect(0, "invalid Account reference "+receiverAcct.String()))
		logic.TestApps(t, []string{axferToArgs}, txntest.Group(&appl, &afrz), 9, ledger,
			logic.NewExpect(0, "invalid Account reference "+receiverAcct.String()))

		// otherAcct is the afrz target, it's holding and account are available
		appl.ApplicationArgs = [][]byte{otherAcct[:], {asa1}}
		logic.TestApps(t, []string{"", payToArg}, txntest.Group(&afrz, &appl), 9, ledger)
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&afrz, &appl), 9, ledger)

		// but still can't axfer a different asset
		appl.ApplicationArgs = [][]byte{otherAcct[:], {asa2}}
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&afrz, &appl), 9, ledger,
			logic.NewExpect(1, fmt.Sprintf("invalid Asset reference %d", asa2)))
		appl.ForeignAssets = []basics.AssetIndex{asa2}
		// once added to appl's foreign array, the appl still lacks access to other's holding
		logic.TestApps(t, []string{"", axferToArgs}, txntest.Group(&afrz, &appl), 9, ledger,
			logic.NewExpect(1, "invalid Holding access "+otherAcct.String()))

		// appl can acfg the asset from tx0 (which requires asset available, not holding)
		appl.ForeignAssets = []basics.AssetIndex{}
		appl.ApplicationArgs = [][]byte{{asa1}}
		logic.TestApps(t, []string{"", acfgArg}, txntest.Group(&afrz, &appl), 9, ledger)
		appl.ApplicationArgs = [][]byte{{asa2}} // but not asa2
		logic.TestApps(t, []string{"", acfgArg}, txntest.Group(&afrz, &appl), 9, ledger,
			logic.NewExpect(1, fmt.Sprintf("invalid Asset reference %d", asa2)))

	})

}
