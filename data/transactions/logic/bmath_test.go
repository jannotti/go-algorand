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

package logic

import (
	"testing"

	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/stretchr/testify/require"
)

const bmathCompiled = "800301234549a0a049a0a149a0a249a0a349a0a4a0af49a0a5a0af49a0a6a0af49a0a7a0af49a0a8a0af49a0a9a0af49a0aa49a0ab49a0ac49a0ada0ae"

const bmathNonsense = `
 pushbytes 0x012345
 dup
 b+
 dup
 b-
 dup
 b/
 dup
 b*
 dup
 b<
 bzero
 dup
 b>
 bzero
 dup
 b<=
 bzero
 dup
 b>=
 bzero
 dup
 b==
 bzero
 dup
 b!=
 bzero
 dup
 b%
 dup
 b|
 dup
 b&
 dup
 b^
 b~
`

func TestDeprecation(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()
	var txn transactions.SignedTxn
	txn.Lsig.Logic = []byte{byte(multiVersion), 0x80, 0x01, 0x01, 0x49, 0xa2}
	ep := defaultEvalParamsWithVersion(multiVersion, txn)
	_, err := EvalSignature(0, ep)
	require.ErrorContains(t, err, "deprecated opcode")
}

func TestMultiBytes(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()
	sources := []string{"pushbytes 0x01; pushbytes 0x02; b/"}
	for _, source := range sources {
		ops, err := AssembleStringWithVersion(source, AssemblerMaxVersion)
		if len(ops.Errors) > 0 || err != nil || ops == nil || ops.Program == nil {
			ops, err := assembleWithTrace(source, AssemblerMaxVersion)
			require.NoError(t, err)
			t.Log(ops.Trace)
		}
		require.Empty(t, ops.Errors)
		require.Equal(t, []byte{AssemblerMaxVersion, 0x80, 0x01, 0x01, 0x80, 0x01, 0x02, 0xa0, 0xa2}, ops.Program)
		testProg(t, source, AssemblerMaxVersion)
	}
}
