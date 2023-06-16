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

package logic

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	bls12381fp "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	bls12381fr "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

type sError string

func (s sError) Error() string { return string(s) }

const (
	errNotOnCurve    = sError("point not on curve")
	errWrongSubgroup = sError("wrong subgroup")
	errEmptyInput    = sError("empty input")
)

// Input: Two byte slices at top of stack, each an uncompressed point
// Output: Single byte slice on top of stack which is the uncompressed sum of inputs
func opEcAdd(cx *EvalContext) error {
	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_add group %s", group)
	}

	last := len(cx.stack) - 1
	prev := last - 1
	a := cx.stack[prev].Bytes
	b := cx.stack[last].Bytes

	var res []byte
	var err error
	switch fs.field {
	case BN254g1:
		res, err = bn254G1Add(a, b)
	case BN254g2:
		res, err = bn254G2Add(a, b)
	case BLS12_381g1:
		res, err = bls12381G1Add(a, b)
	case BLS12_381g2:
		res, err = bls12381G2Add(a, b)
	default:
		err = fmt.Errorf("invalid ec_add group %s", group)
	}
	cx.stack[prev].Bytes = res
	cx.stack = cx.stack[:last]
	return err
}

// Input: ToS is a scalar, encoded as an unsigned big-endian, second to top is
// uncompressed bytes for g1 point
// Output: Single byte slice on top of stack which contains uncompressed bytes
// for product of scalar and point
func opEcScalarMul(cx *EvalContext) error {
	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_scalar_mul group %s", group)
	}

	last := len(cx.stack) - 1
	prev := last - 1
	aBytes := cx.stack[prev].Bytes
	kBytes := cx.stack[last].Bytes
	if len(kBytes) > scalarSize {
		return fmt.Errorf("ec_scalar_mul scalar len is %d, exceeds %d", len(kBytes), scalarSize)
	}
	k := new(big.Int).SetBytes(kBytes)

	var res []byte
	var err error
	switch fs.field {
	case BN254g1:
		res, err = bn254G1ScalarMul(aBytes, k)
	case BN254g2:
		res, err = bn254G2ScalarMul(aBytes, k)
	case BLS12_381g1:
		res, err = bls12381G1ScalarMul(aBytes, k)
	case BLS12_381g2:
		res, err = bls12381G2ScalarMul(aBytes, k)
	default:
		err = fmt.Errorf("invalid ec_scalar_mul group %s", group)
	}

	cx.stack = cx.stack[:last]
	cx.stack[prev].Bytes = res
	return err
}

// Input: Two byte slices, top is concatenated uncompressed bytes for k g2 points, and second to top is same for g1
// Output: Single uint at top representing bool for whether pairing of inputs was identity
func opEcPairingCheck(cx *EvalContext) error {
	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_pairing_check group %s", group)
	}

	last := len(cx.stack) - 1
	prev := last - 1
	g1Bytes := cx.stack[prev].Bytes
	g2Bytes := cx.stack[last].Bytes

	var err error
	ok = false
	switch fs.field {
	case BN254g2:
		g1Bytes, g2Bytes = g2Bytes, g1Bytes
		fallthrough
	case BN254g1:
		ok, err = bn254PairingCheck(g1Bytes, g2Bytes)
	case BLS12_381g2:
		g1Bytes, g2Bytes = g2Bytes, g1Bytes
		fallthrough
	case BLS12_381g1:
		ok, err = bls12381PairingCheck(g1Bytes, g2Bytes)
	default:
		err = fmt.Errorf("invalid ec_pairing_check group %s", group)
	}

	cx.stack = cx.stack[:last]
	cx.stack[prev] = boolToSV(ok)
	return err
}

// Input: Top of stack is slice of k scalars, second to top is slice of k group points as uncompressed bytes
// Output: Single byte slice that contains uncompressed bytes for point equivalent to p_1 * e_1 + p_2 * e_2 + ... + p_k^e_k, where p_i is i'th point from input and e_i is i'th scalar
func opEcMultiExp(cx *EvalContext) error {
	last := len(cx.stack) - 1
	prev := last - 1
	pointBytes := cx.stack[prev].Bytes
	scalarBytes := cx.stack[last].Bytes

	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_multiexp group %s", group)
	}

	var res []byte
	var err error
	switch fs.field {
	case BN254g1:
		res, err = bn254G1MultiExp(pointBytes, scalarBytes)
	case BN254g2:
		res, err = bn254G2MultiExp(pointBytes, scalarBytes)
	case BLS12_381g1:
		res, err = bls12381G1MultiExp(pointBytes, scalarBytes)
	case BLS12_381g2:
		res, err = bls12381G2MultiExp(pointBytes, scalarBytes)
	default:
		err = fmt.Errorf("invalid ec_multiexp group %s", group)
	}

	cx.stack = cx.stack[:last]
	cx.stack[prev].Bytes = res
	return err
}

// Input: Single byte slice on top of stack containing uncompressed bytes for g1 point
// Output: Single uint on stack top representing bool for whether the input was in the correct subgroup or not
func opEcSubgroupCheck(cx *EvalContext) error {
	last := len(cx.stack) - 1
	pointBytes := cx.stack[last].Bytes

	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_pairing_check group %s", group)
	}

	var err error
	ok = false
	switch fs.field {
	case BN254g1:
		ok, err = bn254G1SubgroupCheck(pointBytes)
	case BN254g2:
		ok, err = bn254G2SubgroupCheck(pointBytes)
	case BLS12_381g1:
		ok, err = bls12381G1SubgroupCheck(pointBytes)
	case BLS12_381g2:
		ok, err = bls12381G2SubgroupCheck(pointBytes)
	default:
		err = fmt.Errorf("invalid ec_pairing_check group %s", group)
	}

	cx.stack[last] = boolToSV(ok)
	return err
}

// Input: Single byte slice on top of stack representing single field element
// Output: Single byte slice on top of stack which contains uncompressed bytes
// for corresponding point (mapped to by input)
func opEcMapTo(cx *EvalContext) error {
	last := len(cx.stack) - 1
	fpBytes := cx.stack[last].Bytes

	group := EcGroup(cx.program[cx.pc+1])
	fs, ok := ecGroupSpecByField(group)
	if !ok { // no version check yet, both appeared at once
		return fmt.Errorf("invalid ec_pairing_check group %s", group)
	}

	var res []byte
	var err error
	switch fs.field {
	case BN254g1:
		res, err = bn254MapToG1(fpBytes)
	case BN254g2:
		res, err = bn254MapToG2(fpBytes)
	case BLS12_381g1:
		res, err = bls12381MapToG1(fpBytes)
	case BLS12_381g2:
		res, err = bls12381MapToG2(fpBytes)
	default:
		err = fmt.Errorf("invalid ec_pairing_check group %s", group)
	}
	cx.stack[last].Bytes = res
	return err
}

const (
	bls12381fpSize  = 48
	bls12381g1Size  = 2 * bls12381fpSize
	bls12381fp2Size = 2 * bls12381fpSize
	bls12381g2Size  = 2 * bls12381fp2Size

	bn254fpSize  = 32
	bn254g1Size  = 2 * bn254fpSize
	bn254fp2Size = 2 * bn254fpSize
	bn254g2Size  = 2 * bn254fp2Size

	scalarSize = 32
)

var bls12381Modulus = bls12381fp.Modulus()

func bytesToBLS12381Field(b []byte) (bls12381fp.Element, error) {
	var big big.Int
	big.SetBytes(b)
	if big.Cmp(bls12381Modulus) >= 0 {
		return bls12381fp.Element{}, fmt.Errorf("field element %s larger than modulus %s", &big, bls12381Modulus)
	}
	return *new(bls12381fp.Element).SetBigInt(&big), nil
}

func bytesToBLS12381G1(b []byte) (bls12381.G1Affine, error) {
	if len(b) != bls12381g1Size {
		return bls12381.G1Affine{}, fmt.Errorf("bad length %d. Expected %d", len(b), bls12381g1Size)
	}
	var point bls12381.G1Affine
	var err error
	point.X, err = bytesToBLS12381Field(b[:bls12381fpSize])
	if err != nil {
		return bls12381.G1Affine{}, err
	}
	point.Y, err = bytesToBLS12381Field(b[bls12381fpSize:bls12381g1Size])
	if err != nil {
		return bls12381.G1Affine{}, err
	}
	if !point.IsOnCurve() {
		return bls12381.G1Affine{}, errNotOnCurve
	}
	return point, nil
}

func bytesToBLS12381G1s(b []byte, checkSubgroup bool) ([]bls12381.G1Affine, error) {
	if len(b)%bls12381g1Size != 0 {
		return nil, fmt.Errorf("bad length %d. Expected %d multiple", len(b), bls12381g1Size)
	}
	if len(b) == 0 {
		return nil, errEmptyInput
	}
	points := make([]bls12381.G1Affine, len(b)/bls12381g1Size)
	for i := range points {
		var err error
		points[i], err = bytesToBLS12381G1(b[i*bls12381g1Size : (i+1)*bls12381g1Size])
		if err != nil {
			return nil, err
		}
		if checkSubgroup && !points[i].IsInSubGroup() {
			return nil, errWrongSubgroup
		}
	}
	return points, nil
}

func bytesToBLS12381G2(b []byte) (bls12381.G2Affine, error) {
	if len(b) != bls12381g2Size {
		return bls12381.G2Affine{}, fmt.Errorf("bad length %d. Expected %d", len(b), bls12381g2Size)
	}
	var err error
	var point bls12381.G2Affine
	point.X.A0, err = bytesToBLS12381Field(b[:bls12381fpSize])
	if err != nil {
		return bls12381.G2Affine{}, err
	}
	point.X.A1, err = bytesToBLS12381Field(b[bls12381fpSize : 2*bls12381fpSize])
	if err != nil {
		return bls12381.G2Affine{}, err
	}
	point.Y.A0, err = bytesToBLS12381Field(b[2*bls12381fpSize : 3*bls12381fpSize])
	if err != nil {
		return bls12381.G2Affine{}, err
	}
	point.Y.A1, err = bytesToBLS12381Field(b[3*bls12381fpSize : 4*bls12381fpSize])
	if err != nil {
		return bls12381.G2Affine{}, err
	}
	if !point.IsOnCurve() {
		return bls12381.G2Affine{}, errNotOnCurve
	}
	return point, nil
}

func bytesToBLS12381G2s(b []byte, checkSubgroup bool) ([]bls12381.G2Affine, error) {
	if len(b)%bls12381g2Size != 0 {
		return nil, fmt.Errorf("bad length %d. Expected %d multiple", len(b), bls12381g2Size)
	}
	if len(b) == 0 {
		return nil, errEmptyInput
	}
	points := make([]bls12381.G2Affine, len(b)/bls12381g2Size)
	for i := range points {
		var err error
		points[i], err = bytesToBLS12381G2(b[i*bls12381g2Size : (i+1)*bls12381g2Size])
		if err != nil {
			return nil, err
		}
		if checkSubgroup && !points[i].IsInSubGroup() {
			return nil, errWrongSubgroup
		}
	}
	return points, nil
}

func bls12381G1ToBytes(g1 *bls12381.G1Affine) []byte {
	retX := g1.X.Bytes()
	retY := g1.Y.Bytes()
	return append(retX[:], retY[:]...)
}

func bls12381G2ToBytes(g2 *bls12381.G2Affine) []byte {
	xFirst := g2.X.A0.Bytes()
	xSecond := g2.X.A1.Bytes()
	yFirst := g2.Y.A0.Bytes()
	ySecond := g2.Y.A1.Bytes()
	pointBytes := make([]byte, bls12381g2Size)
	copy(pointBytes, xFirst[:])
	copy(pointBytes[bls12381fpSize:], xSecond[:])
	copy(pointBytes[bls12381fp2Size:], yFirst[:])
	copy(pointBytes[bls12381fp2Size+bls12381fpSize:], ySecond[:])
	return pointBytes
}

func bls12381G1Add(aBytes, bBytes []byte) ([]byte, error) {
	a, err := bytesToBLS12381G1(aBytes)
	if err != nil {
		return nil, err
	}
	b, err := bytesToBLS12381G1(bBytes)
	if err != nil {
		return nil, err
	}
	return bls12381G1ToBytes(a.Add(&a, &b)), nil
}

func bls12381G2Add(aBytes, bBytes []byte) ([]byte, error) {
	a, err := bytesToBLS12381G2(aBytes)
	if err != nil {
		return nil, err
	}
	b, err := bytesToBLS12381G2(bBytes)
	if err != nil {
		return nil, err
	}

	return bls12381G2ToBytes(a.Add(&a, &b)), nil
}

func bls12381G1ScalarMul(aBytes []byte, k *big.Int) ([]byte, error) {
	a, err := bytesToBLS12381G1(aBytes)
	if err != nil {
		return nil, err
	}
	return bls12381G1ToBytes(a.ScalarMultiplication(&a, k)), nil
}

func bls12381G2ScalarMul(aBytes []byte, k *big.Int) ([]byte, error) {
	a, err := bytesToBLS12381G2(aBytes)
	if err != nil {
		return nil, err
	}
	return bls12381G2ToBytes(a.ScalarMultiplication(&a, k)), nil
}

func bls12381PairingCheck(g1Bytes, g2Bytes []byte) (bool, error) {
	g1, err := bytesToBLS12381G1s(g1Bytes, true)
	if err != nil {
		return false, err
	}
	g2, err := bytesToBLS12381G2s(g2Bytes, true)
	if err != nil {
		return false, err
	}
	ok, err := bls12381.PairingCheck(g1, g2)
	if err != nil {
		return false, err
	}
	return ok, nil
}

var eccMontgomery = ecc.MultiExpConfig{ScalarsMont: true}

const bls12381G1MultiExpThreshold = 2 // determined by BenchmarkFindMultiExpCutoff

func bls12381G1MultiExp(pointBytes, scalarBytes []byte) ([]byte, error) {
	points, err := bytesToBLS12381G1s(pointBytes, false)
	if err != nil {
		return nil, err
	}
	if len(scalarBytes) != scalarSize*len(points) {
		return nil, fmt.Errorf("bad scalars length %d. Expected %d", len(scalarBytes), scalarSize*len(points))
	}
	if len(points) <= bls12381G1MultiExpThreshold {
		return bls12381G1MultiExpSmall(points, scalarBytes)
	}
	return bls12381G1MultiExpLarge(points, scalarBytes)
}

func bls12381G1MultiExpLarge(points []bls12381.G1Affine, scalarBytes []byte) ([]byte, error) {
	scalars := make([]bls12381fr.Element, len(points))
	for i := range scalars {
		scalars[i].SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
	}
	res, err := new(bls12381.G1Affine).MultiExp(points, scalars, eccMontgomery)
	if err != nil {
		return nil, err
	}
	return bls12381G1ToBytes(res), nil
}

func bls12381G1MultiExpSmall(points []bls12381.G1Affine, scalarBytes []byte) ([]byte, error) {
	// There must be at least one point. Start with it, rather than the identity.
	k := new(big.Int).SetBytes(scalarBytes[:scalarSize])
	var sum bls12381.G1Jac
	sum.ScalarMultiplicationAffine(&points[0], k)
	for i := range points {
		if i == 0 {
			continue
		}
		k.SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
		var prod bls12381.G1Jac
		prod.ScalarMultiplicationAffine(&points[i], k)
		sum.AddAssign(&prod)
	}
	var res bls12381.G1Affine
	res.FromJacobian(&sum)
	return bls12381G1ToBytes(&res), nil
}

const bls12381G2MultiExpThreshold = 2 // determined by BenchmarkFindMultiExpCutoff

func bls12381G2MultiExp(pointBytes, scalarBytes []byte) ([]byte, error) {
	points, err := bytesToBLS12381G2s(pointBytes, false)
	if err != nil {
		return nil, err
	}
	if len(scalarBytes) != scalarSize*len(points) {
		return nil, fmt.Errorf("bad scalars length %d. Expected %d", len(scalarBytes), scalarSize*len(points))
	}
	if len(points) <= bls12381G2MultiExpThreshold {
		return bls12381G2MultiExpSmall(points, scalarBytes)
	}
	return bls12381G2MultiExpLarge(points, scalarBytes)
}

func bls12381G2MultiExpLarge(points []bls12381.G2Affine, scalarBytes []byte) ([]byte, error) {
	scalars := make([]bls12381fr.Element, len(points))
	for i := range scalars {
		scalars[i].SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
	}
	res, err := new(bls12381.G2Affine).MultiExp(points, scalars, eccMontgomery)
	if err != nil {
		return nil, err
	}
	return bls12381G2ToBytes(res), nil
}

func bls12381G2MultiExpSmall(points []bls12381.G2Affine, scalarBytes []byte) ([]byte, error) {
	// There must be at least one point. Start with it, rather than the identity.
	k := new(big.Int).SetBytes(scalarBytes[:scalarSize])
	var sum bls12381.G2Jac
	sum.FromAffine(&points[0])
	sum.ScalarMultiplication(&sum, k)
	for i := range points {
		if i == 0 {
			continue
		}
		k.SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
		var prod bls12381.G2Jac
		prod.FromAffine(&points[i])
		prod.ScalarMultiplication(&prod, k)
		sum.AddAssign(&prod)
	}
	var res bls12381.G2Affine
	res.FromJacobian(&sum)
	return bls12381G2ToBytes(&res), nil
}

func bls12381MapToG1(fpBytes []byte) ([]byte, error) {
	fp, err := bytesToBLS12381Field(fpBytes)
	if err != nil {
		return nil, err
	}
	point := bls12381.MapToG1(fp)
	return bls12381G1ToBytes(&point), nil
}

func bls12381MapToG2(fpBytes []byte) ([]byte, error) {
	if len(fpBytes) != bls12381fp2Size {
		return nil, fmt.Errorf("bad encoded element length: %d", len(fpBytes))
	}
	g2 := bls12381.G2Affine{}
	var err error
	g2.X.A0, err = bytesToBLS12381Field(fpBytes[0:bls12381fpSize])
	if err != nil {
		return nil, err
	}
	g2.X.A1, err = bytesToBLS12381Field(fpBytes[bls12381fpSize:])
	if err != nil {
		return nil, err
	}
	point := bls12381.MapToG2(g2.X)
	return bls12381G2ToBytes(&point), nil
}

func bls12381G1SubgroupCheck(pointBytes []byte) (bool, error) {
	point, err := bytesToBLS12381G1(pointBytes)
	if err != nil {
		return false, err
	}
	return point.IsInSubGroup(), nil
}

func bls12381G2SubgroupCheck(pointBytes []byte) (bool, error) {
	point, err := bytesToBLS12381G2(pointBytes)
	if err != nil {
		return false, err
	}
	return point.IsInSubGroup(), nil
}

var bn254Modulus = bn254fp.Modulus()

func bytesToBN254Field(b []byte) (bn254fp.Element, error) {
	var big big.Int
	big.SetBytes(b)
	if big.Cmp(bn254Modulus) >= 0 {
		return bn254fp.Element{}, fmt.Errorf("field element %s larger than modulus %s", &big, bn254Modulus)
	}
	return *new(bn254fp.Element).SetBigInt(&big), nil
}

func bytesToBN254G1(b []byte) (bn254.G1Affine, error) {
	if len(b) != bn254g1Size {
		return bn254.G1Affine{}, fmt.Errorf("bad length %d. Expected %d", len(b), bn254g1Size)
	}
	var point bn254.G1Affine
	var err error
	point.X, err = bytesToBN254Field(b[:bn254fpSize])
	if err != nil {
		return bn254.G1Affine{}, err
	}
	point.Y, err = bytesToBN254Field(b[bn254fpSize:bn254g1Size])
	if err != nil {
		return bn254.G1Affine{}, err
	}
	if !point.IsOnCurve() {
		return bn254.G1Affine{}, errNotOnCurve
	}
	return point, nil
}

func bytesToBN254G1s(b []byte, checkSubgroup bool) ([]bn254.G1Affine, error) {
	if len(b)%bn254g1Size != 0 {
		return nil, fmt.Errorf("bad length %d. Expected %d multiple", len(b), bn254g1Size)
	}
	if len(b) == 0 {
		return nil, errEmptyInput
	}
	points := make([]bn254.G1Affine, len(b)/bn254g1Size)
	for i := range points {
		var err error
		points[i], err = bytesToBN254G1(b[i*bn254g1Size : (i+1)*bn254g1Size])
		if err != nil {
			return nil, err
		}
		if checkSubgroup && !points[i].IsInSubGroup() {
			return nil, errWrongSubgroup
		}
	}
	return points, nil
}

func bytesToBN254G2(b []byte) (bn254.G2Affine, error) {
	if len(b) != bn254g2Size {
		return bn254.G2Affine{}, fmt.Errorf("bad length %d. Expected %d", len(b), bn254g2Size)
	}
	var err error
	var point bn254.G2Affine
	point.X.A0, err = bytesToBN254Field(b[:bn254fpSize])
	if err != nil {
		return bn254.G2Affine{}, err
	}
	point.X.A1, err = bytesToBN254Field(b[bn254fpSize : 2*bn254fpSize])
	if err != nil {
		return bn254.G2Affine{}, err
	}
	point.Y.A0, err = bytesToBN254Field(b[2*bn254fpSize : 3*bn254fpSize])
	if err != nil {
		return bn254.G2Affine{}, err
	}
	point.Y.A1, err = bytesToBN254Field(b[3*bn254fpSize : 4*bn254fpSize])
	if err != nil {
		return bn254.G2Affine{}, err
	}
	if !point.IsOnCurve() {
		return bn254.G2Affine{}, errNotOnCurve
	}
	return point, nil
}

func bytesToBN254G2s(b []byte, checkSubgroup bool) ([]bn254.G2Affine, error) {
	if len(b)%bn254g2Size != 0 {
		return nil, fmt.Errorf("bad length %d. Expected %d multiple", len(b), bn254g2Size)
	}
	if len(b) == 0 {
		return nil, errEmptyInput
	}
	points := make([]bn254.G2Affine, len(b)/bn254g2Size)
	for i := range points {
		var err error
		points[i], err = bytesToBN254G2(b[i*bn254g2Size : (i+1)*bn254g2Size])
		if err != nil {
			return nil, err
		}
		if checkSubgroup && !points[i].IsInSubGroup() {
			return nil, errWrongSubgroup
		}
	}
	return points, nil
}

func bn254G1ToBytes(g1 *bn254.G1Affine) []byte {
	retX := g1.X.Bytes()
	retY := g1.Y.Bytes()
	return append(retX[:], retY[:]...)
}

func bn254G2ToBytes(g2 *bn254.G2Affine) []byte {
	xFirst := g2.X.A0.Bytes()
	xSecond := g2.X.A1.Bytes()
	yFirst := g2.Y.A0.Bytes()
	ySecond := g2.Y.A1.Bytes()
	pointBytes := make([]byte, bn254g2Size)
	copy(pointBytes, xFirst[:])
	copy(pointBytes[bn254fpSize:], xSecond[:])
	copy(pointBytes[bn254fp2Size:], yFirst[:])
	copy(pointBytes[bn254fp2Size+bn254fpSize:], ySecond[:])
	return pointBytes
}

func bn254G1Add(aBytes, bBytes []byte) ([]byte, error) {
	a, err := bytesToBN254G1(aBytes)
	if err != nil {
		return nil, err
	}
	b, err := bytesToBN254G1(bBytes)
	if err != nil {
		return nil, err
	}
	return bn254G1ToBytes(a.Add(&a, &b)), nil
}

func bn254G2Add(aBytes, bBytes []byte) ([]byte, error) {
	a, err := bytesToBN254G2(aBytes)
	if err != nil {
		return nil, err
	}
	b, err := bytesToBN254G2(bBytes)
	if err != nil {
		return nil, err
	}
	return bn254G2ToBytes(a.Add(&a, &b)), nil
}

func bn254G1ScalarMul(aBytes []byte, k *big.Int) ([]byte, error) {
	a, err := bytesToBN254G1(aBytes)
	if err != nil {
		return nil, err
	}
	return bn254G1ToBytes(a.ScalarMultiplication(&a, k)), nil
}

func bn254G2ScalarMul(aBytes []byte, k *big.Int) ([]byte, error) {
	a, err := bytesToBN254G2(aBytes)
	if err != nil {
		return nil, err
	}
	return bn254G2ToBytes(a.ScalarMultiplication(&a, k)), nil
}

func bn254PairingCheck(g1Bytes, g2Bytes []byte) (bool, error) {
	g1, err := bytesToBN254G1s(g1Bytes, true)
	if err != nil {
		return false, err
	}
	g2, err := bytesToBN254G2s(g2Bytes, true)
	if err != nil {
		return false, err
	}
	ok, err := bn254.PairingCheck(g1, g2)
	if err != nil {
		return false, err
	}
	return ok, nil
}

const bn254G1MultiExpThreshold = 2 // determined by BenchmarkFindMultiExpCutoff

func bn254G1MultiExp(pointBytes, scalarBytes []byte) ([]byte, error) {
	points, err := bytesToBN254G1s(pointBytes, false)
	if err != nil {
		return nil, err
	}
	if len(scalarBytes) != scalarSize*len(points) {
		return nil, fmt.Errorf("bad scalars length %d. Expected %d", len(scalarBytes), scalarSize*len(points))
	}
	if len(points) <= bn254G1MultiExpThreshold {
		return bn254G1MultiExpSmall(points, scalarBytes)
	}
	return bn254G1MultiExpLarge(points, scalarBytes)
}

func bn254G1MultiExpLarge(points []bn254.G1Affine, scalarBytes []byte) ([]byte, error) {
	scalars := make([]bn254fr.Element, len(points))
	for i := range scalars {
		scalars[i].SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
	}
	res, err := new(bn254.G1Affine).MultiExp(points, scalars, eccMontgomery)
	if err != nil {
		return nil, err
	}
	return bn254G1ToBytes(res), nil
}

func bn254G1MultiExpSmall(points []bn254.G1Affine, scalarBytes []byte) ([]byte, error) {
	// There must be at least one point. Start with it, rather than the identity.
	k := new(big.Int).SetBytes(scalarBytes[:scalarSize])
	var sum bn254.G1Jac
	sum.ScalarMultiplicationAffine(&points[0], k)
	for i := range points {
		if i == 0 {
			continue
		}
		k.SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
		var prod bn254.G1Jac
		prod.ScalarMultiplicationAffine(&points[i], k)
		sum.AddAssign(&prod)
	}
	var res bn254.G1Affine
	res.FromJacobian(&sum)
	return bn254G1ToBytes(&res), nil
}

const bn254G2MultiExpThreshold = 2 // determined by BenchmarkFindMultiExpCutoff

func bn254G2MultiExp(pointBytes, scalarBytes []byte) ([]byte, error) {
	points, err := bytesToBN254G2s(pointBytes, false)
	if err != nil {
		return nil, err
	}
	if len(scalarBytes) != scalarSize*len(points) {
		return nil, fmt.Errorf("bad scalars length %d. Expected %d", len(scalarBytes), scalarSize*len(points))
	}
	if len(points) <= bn254G2MultiExpThreshold {
		return bn254G2MultiExpSmall(points, scalarBytes)
	}
	return bn254G2MultiExpLarge(points, scalarBytes)
}

func bn254G2MultiExpLarge(points []bn254.G2Affine, scalarBytes []byte) ([]byte, error) {
	scalars := make([]bn254fr.Element, len(points))
	for i := range scalars {
		scalars[i].SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
	}
	res, err := new(bn254.G2Affine).MultiExp(points, scalars, eccMontgomery)
	if err != nil {
		return nil, err
	}
	return bn254G2ToBytes(res), nil
}

func bn254G2MultiExpSmall(points []bn254.G2Affine, scalarBytes []byte) ([]byte, error) {
	// There must be at least one point. Start with it, rather than the identity.
	k := new(big.Int).SetBytes(scalarBytes[:scalarSize])
	var sum bn254.G2Jac
	sum.FromAffine(&points[0])
	sum.ScalarMultiplication(&sum, k)
	for i := range points {
		if i == 0 {
			continue
		}
		k.SetBytes(scalarBytes[i*scalarSize : (i+1)*scalarSize])
		var prod bn254.G2Jac
		prod.FromAffine(&points[i])
		prod.ScalarMultiplication(&prod, k)
		sum.AddAssign(&prod)
	}
	var res bn254.G2Affine
	res.FromJacobian(&sum)
	return bn254G2ToBytes(&res), nil
}

func bn254MapToG1(fpBytes []byte) ([]byte, error) {
	fp, err := bytesToBN254Field(fpBytes)
	if err != nil {
		return nil, err
	}
	point := bn254.MapToG1(fp)
	return bn254G1ToBytes(&point), nil
}

func bn254MapToG2(fpBytes []byte) ([]byte, error) {
	if len(fpBytes) != bn254fp2Size {
		return nil, fmt.Errorf("bad encoded element length: %d", len(fpBytes))
	}
	fp2 := bn254.G2Affine{}.X // no way to declare an fptower.E2
	var err error
	fp2.A0, err = bytesToBN254Field(fpBytes[0:bn254fpSize])
	if err != nil {
		return nil, err
	}
	fp2.A1, err = bytesToBN254Field(fpBytes[bn254fpSize:])
	if err != nil {
		return nil, err
	}
	point := bn254.MapToG2(fp2)
	return bn254G2ToBytes(&point), nil
}

func bn254G1SubgroupCheck(pointBytes []byte) (bool, error) {
	point, err := bytesToBN254G1(pointBytes)
	if err != nil {
		return false, err
	}
	return point.IsInSubGroup(), nil
}

func bn254G2SubgroupCheck(pointBytes []byte) (bool, error) {
	point, err := bytesToBN254G2(pointBytes)
	if err != nil {
		return false, err
	}
	return point.IsInSubGroup(), nil
}
