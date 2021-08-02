package merklekeystore

// Code generated by github.com/algorand/msgp DO NOT EDIT.

import (
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/msgp/msgp"
)

// The following msgp objects are implemented in this file:
// CommittablePublicKey
//           |-----> (*) MarshalMsg
//           |-----> (*) CanMarshalMsg
//           |-----> (*) UnmarshalMsg
//           |-----> (*) CanUnmarshalMsg
//           |-----> (*) Msgsize
//           |-----> (*) MsgIsZero
//
// Proof
//   |-----> MarshalMsg
//   |-----> CanMarshalMsg
//   |-----> (*) UnmarshalMsg
//   |-----> (*) CanUnmarshalMsg
//   |-----> Msgsize
//   |-----> MsgIsZero
//
// Signature
//     |-----> (*) MarshalMsg
//     |-----> (*) CanMarshalMsg
//     |-----> (*) UnmarshalMsg
//     |-----> (*) CanUnmarshalMsg
//     |-----> (*) Msgsize
//     |-----> (*) MsgIsZero
//
// Signer
//    |-----> (*) MarshalMsg
//    |-----> (*) CanMarshalMsg
//    |-----> (*) UnmarshalMsg
//    |-----> (*) CanUnmarshalMsg
//    |-----> (*) Msgsize
//    |-----> (*) MsgIsZero
//
// Verifier
//     |-----> (*) MarshalMsg
//     |-----> (*) CanMarshalMsg
//     |-----> (*) UnmarshalMsg
//     |-----> (*) CanUnmarshalMsg
//     |-----> (*) Msgsize
//     |-----> (*) MsgIsZero
//

// MarshalMsg implements msgp.Marshaler
func (z *CommittablePublicKey) MarshalMsg(b []byte) (o []byte) {
	o = msgp.Require(b, z.Msgsize())
	// omitempty: check for empty values
	zb0001Len := uint32(2)
	var zb0001Mask uint8 /* 3 bits */
	if (*z).VerifyingKey.MsgIsZero() {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if (*z).Round == 0 {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	// variable map header, size zb0001Len
	o = append(o, 0x80|uint8(zb0001Len))
	if zb0001Len != 0 {
		if (zb0001Mask & 0x2) == 0 { // if not empty
			// string "pk"
			o = append(o, 0xa2, 0x70, 0x6b)
			o = (*z).VerifyingKey.MarshalMsg(o)
		}
		if (zb0001Mask & 0x4) == 0 { // if not empty
			// string "rnd"
			o = append(o, 0xa3, 0x72, 0x6e, 0x64)
			o = msgp.AppendUint64(o, (*z).Round)
		}
	}
	return
}

func (_ *CommittablePublicKey) CanMarshalMsg(z interface{}) bool {
	_, ok := (z).(*CommittablePublicKey)
	return ok
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CommittablePublicKey) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 int
	var zb0002 bool
	zb0001, zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if _, ok := err.(msgp.TypeError); ok {
		zb0001, zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0001 > 0 {
			zb0001--
			bts, err = (*z).VerifyingKey.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "VerifyingKey")
				return
			}
		}
		if zb0001 > 0 {
			zb0001--
			(*z).Round, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "Round")
				return
			}
		}
		if zb0001 > 0 {
			err = msgp.ErrTooManyArrayFields(zb0001)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array")
				return
			}
		}
	} else {
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0002 {
			(*z) = CommittablePublicKey{}
		}
		for zb0001 > 0 {
			zb0001--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
			switch string(field) {
			case "pk":
				bts, err = (*z).VerifyingKey.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "VerifyingKey")
					return
				}
			case "rnd":
				(*z).Round, bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Round")
					return
				}
			default:
				err = msgp.ErrNoField(string(field))
				if err != nil {
					err = msgp.WrapError(err)
					return
				}
			}
		}
	}
	o = bts
	return
}

func (_ *CommittablePublicKey) CanUnmarshalMsg(z interface{}) bool {
	_, ok := (z).(*CommittablePublicKey)
	return ok
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *CommittablePublicKey) Msgsize() (s int) {
	s = 1 + 3 + (*z).VerifyingKey.Msgsize() + 4 + msgp.Uint64Size
	return
}

// MsgIsZero returns whether this is a zero value
func (z *CommittablePublicKey) MsgIsZero() bool {
	return ((*z).VerifyingKey.MsgIsZero()) && ((*z).Round == 0)
}

// MarshalMsg implements msgp.Marshaler
func (z Proof) MarshalMsg(b []byte) (o []byte) {
	o = msgp.Require(b, z.Msgsize())
	if z == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendArrayHeader(o, uint32(len(z)))
	}
	for za0001 := range z {
		o = z[za0001].MarshalMsg(o)
	}
	return
}

func (_ Proof) CanMarshalMsg(z interface{}) bool {
	_, ok := (z).(Proof)
	if !ok {
		_, ok = (z).(*Proof)
	}
	return ok
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Proof) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 int
	var zb0003 bool
	zb0002, zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0003 {
		(*z) = nil
	} else if (*z) != nil && cap((*z)) >= zb0002 {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(Proof, zb0002)
	}
	for zb0001 := range *z {
		bts, err = (*z)[zb0001].UnmarshalMsg(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
	}
	o = bts
	return
}

func (_ *Proof) CanUnmarshalMsg(z interface{}) bool {
	_, ok := (z).(*Proof)
	return ok
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Proof) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for za0001 := range z {
		s += z[za0001].Msgsize()
	}
	return
}

// MsgIsZero returns whether this is a zero value
func (z Proof) MsgIsZero() bool {
	return len(z) == 0
}

// MarshalMsg implements msgp.Marshaler
func (z *Signature) MarshalMsg(b []byte) (o []byte) {
	o = msgp.Require(b, z.Msgsize())
	// omitempty: check for empty values
	zb0002Len := uint32(3)
	var zb0002Mask uint8 /* 4 bits */
	if (*z).ByteSignature.MsgIsZero() {
		zb0002Len--
		zb0002Mask |= 0x2
	}
	if len((*z).Proof) == 0 {
		zb0002Len--
		zb0002Mask |= 0x4
	}
	if (*z).VerifyingKey.MsgIsZero() {
		zb0002Len--
		zb0002Mask |= 0x8
	}
	// variable map header, size zb0002Len
	o = append(o, 0x80|uint8(zb0002Len))
	if zb0002Len != 0 {
		if (zb0002Mask & 0x2) == 0 { // if not empty
			// string "bsig"
			o = append(o, 0xa4, 0x62, 0x73, 0x69, 0x67)
			o = (*z).ByteSignature.MarshalMsg(o)
		}
		if (zb0002Mask & 0x4) == 0 { // if not empty
			// string "prf"
			o = append(o, 0xa3, 0x70, 0x72, 0x66)
			if (*z).Proof == nil {
				o = msgp.AppendNil(o)
			} else {
				o = msgp.AppendArrayHeader(o, uint32(len((*z).Proof)))
			}
			for zb0001 := range (*z).Proof {
				o = (*z).Proof[zb0001].MarshalMsg(o)
			}
		}
		if (zb0002Mask & 0x8) == 0 { // if not empty
			// string "vkey"
			o = append(o, 0xa4, 0x76, 0x6b, 0x65, 0x79)
			o = (*z).VerifyingKey.MarshalMsg(o)
		}
	}
	return
}

func (_ *Signature) CanMarshalMsg(z interface{}) bool {
	_, ok := (z).(*Signature)
	return ok
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Signature) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0002 int
	var zb0003 bool
	zb0002, zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if _, ok := err.(msgp.TypeError); ok {
		zb0002, zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0002 > 0 {
			zb0002--
			bts, err = (*z).ByteSignature.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "ByteSignature")
				return
			}
		}
		if zb0002 > 0 {
			zb0002--
			var zb0004 int
			var zb0005 bool
			zb0004, zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "Proof")
				return
			}
			if zb0005 {
				(*z).Proof = nil
			} else if (*z).Proof != nil && cap((*z).Proof) >= zb0004 {
				(*z).Proof = ((*z).Proof)[:zb0004]
			} else {
				(*z).Proof = make(Proof, zb0004)
			}
			for zb0001 := range (*z).Proof {
				bts, err = (*z).Proof[zb0001].UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "struct-from-array", "Proof", zb0001)
					return
				}
			}
		}
		if zb0002 > 0 {
			zb0002--
			bts, err = (*z).VerifyingKey.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "VerifyingKey")
				return
			}
		}
		if zb0002 > 0 {
			err = msgp.ErrTooManyArrayFields(zb0002)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array")
				return
			}
		}
	} else {
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0003 {
			(*z) = Signature{}
		}
		for zb0002 > 0 {
			zb0002--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
			switch string(field) {
			case "bsig":
				bts, err = (*z).ByteSignature.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "ByteSignature")
					return
				}
			case "prf":
				var zb0006 int
				var zb0007 bool
				zb0006, zb0007, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Proof")
					return
				}
				if zb0007 {
					(*z).Proof = nil
				} else if (*z).Proof != nil && cap((*z).Proof) >= zb0006 {
					(*z).Proof = ((*z).Proof)[:zb0006]
				} else {
					(*z).Proof = make(Proof, zb0006)
				}
				for zb0001 := range (*z).Proof {
					bts, err = (*z).Proof[zb0001].UnmarshalMsg(bts)
					if err != nil {
						err = msgp.WrapError(err, "Proof", zb0001)
						return
					}
				}
			case "vkey":
				bts, err = (*z).VerifyingKey.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "VerifyingKey")
					return
				}
			default:
				err = msgp.ErrNoField(string(field))
				if err != nil {
					err = msgp.WrapError(err)
					return
				}
			}
		}
	}
	o = bts
	return
}

func (_ *Signature) CanUnmarshalMsg(z interface{}) bool {
	_, ok := (z).(*Signature)
	return ok
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Signature) Msgsize() (s int) {
	s = 1 + 5 + (*z).ByteSignature.Msgsize() + 4 + msgp.ArrayHeaderSize
	for zb0001 := range (*z).Proof {
		s += (*z).Proof[zb0001].Msgsize()
	}
	s += 5 + (*z).VerifyingKey.Msgsize()
	return
}

// MsgIsZero returns whether this is a zero value
func (z *Signature) MsgIsZero() bool {
	return ((*z).ByteSignature.MsgIsZero()) && (len((*z).Proof) == 0) && ((*z).VerifyingKey.MsgIsZero())
}

// MarshalMsg implements msgp.Marshaler
func (z *Signer) MarshalMsg(b []byte) (o []byte) {
	o = msgp.Require(b, z.Msgsize())
	// omitempty: check for empty values
	zb0002Len := uint32(5)
	var zb0002Mask uint8 /* 7 bits */
	if (*z).ArrayBase == 0 {
		zb0002Len--
		zb0002Mask |= 0x2
	}
	if (*z).Interval == 0 {
		zb0002Len--
		zb0002Mask |= 0x4
	}
	if (*z).FirstValid == 0 {
		zb0002Len--
		zb0002Mask |= 0x10
	}
	if len((*z).SignatureAlgorithms) == 0 {
		zb0002Len--
		zb0002Mask |= 0x20
	}
	if (*z).Tree.MsgIsZero() {
		zb0002Len--
		zb0002Mask |= 0x40
	}
	// variable map header, size zb0002Len
	o = append(o, 0x80|uint8(zb0002Len))
	if zb0002Len != 0 {
		if (zb0002Mask & 0x2) == 0 { // if not empty
			// string "az"
			o = append(o, 0xa2, 0x61, 0x7a)
			o = msgp.AppendUint64(o, (*z).ArrayBase)
		}
		if (zb0002Mask & 0x4) == 0 { // if not empty
			// string "iv"
			o = append(o, 0xa2, 0x69, 0x76)
			o = msgp.AppendUint64(o, (*z).Interval)
		}
		if (zb0002Mask & 0x10) == 0 { // if not empty
			// string "rnd"
			o = append(o, 0xa3, 0x72, 0x6e, 0x64)
			o = msgp.AppendUint64(o, (*z).FirstValid)
		}
		if (zb0002Mask & 0x20) == 0 { // if not empty
			// string "sks"
			o = append(o, 0xa3, 0x73, 0x6b, 0x73)
			if (*z).SignatureAlgorithms == nil {
				o = msgp.AppendNil(o)
			} else {
				o = msgp.AppendArrayHeader(o, uint32(len((*z).SignatureAlgorithms)))
			}
			for zb0001 := range (*z).SignatureAlgorithms {
				o = (*z).SignatureAlgorithms[zb0001].MarshalMsg(o)
			}
		}
		if (zb0002Mask & 0x40) == 0 { // if not empty
			// string "tree"
			o = append(o, 0xa4, 0x74, 0x72, 0x65, 0x65)
			o = (*z).Tree.MarshalMsg(o)
		}
	}
	return
}

func (_ *Signer) CanMarshalMsg(z interface{}) bool {
	_, ok := (z).(*Signer)
	return ok
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Signer) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0002 int
	var zb0003 bool
	zb0002, zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if _, ok := err.(msgp.TypeError); ok {
		zb0002, zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0002 > 0 {
			zb0002--
			var zb0004 int
			var zb0005 bool
			zb0004, zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "SignatureAlgorithms")
				return
			}
			if zb0005 {
				(*z).SignatureAlgorithms = nil
			} else if (*z).SignatureAlgorithms != nil && cap((*z).SignatureAlgorithms) >= zb0004 {
				(*z).SignatureAlgorithms = ((*z).SignatureAlgorithms)[:zb0004]
			} else {
				(*z).SignatureAlgorithms = make([]crypto.SignatureAlgorithm, zb0004)
			}
			for zb0001 := range (*z).SignatureAlgorithms {
				bts, err = (*z).SignatureAlgorithms[zb0001].UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "struct-from-array", "SignatureAlgorithms", zb0001)
					return
				}
			}
		}
		if zb0002 > 0 {
			zb0002--
			(*z).FirstValid, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "FirstValid")
				return
			}
		}
		if zb0002 > 0 {
			zb0002--
			(*z).ArrayBase, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "ArrayBase")
				return
			}
		}
		if zb0002 > 0 {
			zb0002--
			(*z).Interval, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "Interval")
				return
			}
		}
		if zb0002 > 0 {
			zb0002--
			bts, err = (*z).Tree.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "Tree")
				return
			}
		}
		if zb0002 > 0 {
			err = msgp.ErrTooManyArrayFields(zb0002)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array")
				return
			}
		}
	} else {
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0003 {
			(*z) = Signer{}
		}
		for zb0002 > 0 {
			zb0002--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
			switch string(field) {
			case "sks":
				var zb0006 int
				var zb0007 bool
				zb0006, zb0007, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "SignatureAlgorithms")
					return
				}
				if zb0007 {
					(*z).SignatureAlgorithms = nil
				} else if (*z).SignatureAlgorithms != nil && cap((*z).SignatureAlgorithms) >= zb0006 {
					(*z).SignatureAlgorithms = ((*z).SignatureAlgorithms)[:zb0006]
				} else {
					(*z).SignatureAlgorithms = make([]crypto.SignatureAlgorithm, zb0006)
				}
				for zb0001 := range (*z).SignatureAlgorithms {
					bts, err = (*z).SignatureAlgorithms[zb0001].UnmarshalMsg(bts)
					if err != nil {
						err = msgp.WrapError(err, "SignatureAlgorithms", zb0001)
						return
					}
				}
			case "rnd":
				(*z).FirstValid, bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "FirstValid")
					return
				}
			case "az":
				(*z).ArrayBase, bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "ArrayBase")
					return
				}
			case "iv":
				(*z).Interval, bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Interval")
					return
				}
			case "tree":
				bts, err = (*z).Tree.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "Tree")
					return
				}
			default:
				err = msgp.ErrNoField(string(field))
				if err != nil {
					err = msgp.WrapError(err)
					return
				}
			}
		}
	}
	o = bts
	return
}

func (_ *Signer) CanUnmarshalMsg(z interface{}) bool {
	_, ok := (z).(*Signer)
	return ok
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Signer) Msgsize() (s int) {
	s = 1 + 4 + msgp.ArrayHeaderSize
	for zb0001 := range (*z).SignatureAlgorithms {
		s += (*z).SignatureAlgorithms[zb0001].Msgsize()
	}
	s += 4 + msgp.Uint64Size + 3 + msgp.Uint64Size + 3 + msgp.Uint64Size + 5 + (*z).Tree.Msgsize()
	return
}

// MsgIsZero returns whether this is a zero value
func (z *Signer) MsgIsZero() bool {
	return (len((*z).SignatureAlgorithms) == 0) && ((*z).FirstValid == 0) && ((*z).ArrayBase == 0) && ((*z).Interval == 0) && ((*z).Tree.MsgIsZero())
}

// MarshalMsg implements msgp.Marshaler
func (z *Verifier) MarshalMsg(b []byte) (o []byte) {
	o = msgp.Require(b, z.Msgsize())
	// omitempty: check for empty values
	zb0001Len := uint32(2)
	var zb0001Mask uint8 /* 3 bits */
	if (*z).Root.MsgIsZero() {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if (*z).HasValidRoot == false {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	// variable map header, size zb0001Len
	o = append(o, 0x80|uint8(zb0001Len))
	if zb0001Len != 0 {
		if (zb0001Mask & 0x2) == 0 { // if not empty
			// string "r"
			o = append(o, 0xa1, 0x72)
			o = (*z).Root.MarshalMsg(o)
		}
		if (zb0001Mask & 0x4) == 0 { // if not empty
			// string "vr"
			o = append(o, 0xa2, 0x76, 0x72)
			o = msgp.AppendBool(o, (*z).HasValidRoot)
		}
	}
	return
}

func (_ *Verifier) CanMarshalMsg(z interface{}) bool {
	_, ok := (z).(*Verifier)
	return ok
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Verifier) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 int
	var zb0002 bool
	zb0001, zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if _, ok := err.(msgp.TypeError); ok {
		zb0001, zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0001 > 0 {
			zb0001--
			bts, err = (*z).Root.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "Root")
				return
			}
		}
		if zb0001 > 0 {
			zb0001--
			(*z).HasValidRoot, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array", "HasValidRoot")
				return
			}
		}
		if zb0001 > 0 {
			err = msgp.ErrTooManyArrayFields(zb0001)
			if err != nil {
				err = msgp.WrapError(err, "struct-from-array")
				return
			}
		}
	} else {
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		if zb0002 {
			(*z) = Verifier{}
		}
		for zb0001 > 0 {
			zb0001--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
			switch string(field) {
			case "r":
				bts, err = (*z).Root.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "Root")
					return
				}
			case "vr":
				(*z).HasValidRoot, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "HasValidRoot")
					return
				}
			default:
				err = msgp.ErrNoField(string(field))
				if err != nil {
					err = msgp.WrapError(err)
					return
				}
			}
		}
	}
	o = bts
	return
}

func (_ *Verifier) CanUnmarshalMsg(z interface{}) bool {
	_, ok := (z).(*Verifier)
	return ok
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Verifier) Msgsize() (s int) {
	s = 1 + 2 + (*z).Root.Msgsize() + 3 + msgp.BoolSize
	return
}

// MsgIsZero returns whether this is a zero value
func (z *Verifier) MsgIsZero() bool {
	return ((*z).Root.MsgIsZero()) && ((*z).HasValidRoot == false)
}
