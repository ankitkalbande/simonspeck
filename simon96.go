// Copyright 2013 Samuel Isaacson. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package simonspeck

const (
	roundsSimon96_96  = 52
	roundsSimon96_144 = 54
)

// Use NewSimon96 below to expand a Simon96 key. Simon96Cipher
// implements the cipher.Block interface.
type Simon96Cipher struct {
	k      []uint64
	rounds int
}

// NewSimon64 creates and returns a new Simon64Cipher. It accepts
// either a 96-bit key (for Simon64/96) or a 128-bit key (for
// Simon64/128). See the documentation on Simon32 or the test suite
// for our endianness convention.
func NewSimon96(key []byte) *Simon96Cipher {
	cipher := new(Simon96Cipher)
	var keyWords int
	var z uint64

	switch len(key) {
	case 12:
		keyWords = 2
		z = zSeq2
		cipher.rounds = roundsSimon96_96
	case 18:
		keyWords = 3
		z = zSeq3
		cipher.rounds = roundsSimon96_144
	default:
		panic("NewSimon96() requires either a 96- or 144-bit key")
	}
	cipher.k = make([]uint64, cipher.rounds)
	for i := 0; i < keyWords; i++ {
		cipher.k[i] = littleEndianBytesToUInt48(key[6*i : 6*i+6])
	}
	for i := keyWords; i < cipher.rounds; i++ {
		tmp := rightRotate48(cipher.k[i-1], 3)
		tmp ^= rightRotate48(tmp, 1)
		lfsrBit := (z >> uint((i-keyWords)%62)) & 1
		cipher.k[i] = ^cipher.k[i-keyWords] ^ tmp ^ uint64(lfsrBit) ^ 3
		cipher.k[i] &= bitMask48
	}
	return cipher
}

// Simon96 has a 96-bit block length.
func (cipher *Simon96Cipher) BlockSize() int {
	return 12
}

func simonScramble48(x uint64) uint64 {
	return (leftRotate48(x, 1) & leftRotate48(x, 8)) ^ leftRotate48(x, 2)
}

// Encrypt encrypts the first block in src into dst.
// Dst and src may point at the same memory. See crypto/cipher.
func (cipher *Simon96Cipher) Encrypt(dst, src []byte) {
	y := littleEndianBytesToUInt48(src[0:6])
	x := littleEndianBytesToUInt48(src[6:12])
	for i := 0; i < cipher.rounds; i++ {
		x, y = y^simonScramble48(x)^cipher.k[i], x
	}
	storeLittleEndianUInt48(dst[0:6], y)
	storeLittleEndianUInt48(dst[6:12], x)
}

// Decrypt decrypts the first block in src into dst.
// Dst and src may point at the same memory. See crypto/cipher.
func (cipher *Simon96Cipher) Decrypt(dst, src []byte) {
	y := littleEndianBytesToUInt48(src[0:6])
	x := littleEndianBytesToUInt48(src[6:12])
	for i := cipher.rounds - 1; i >= 0; i-- {
		x, y = y, x^simonScramble48(y)^cipher.k[i]
	}
	storeLittleEndianUInt48(dst[0:6], y)
	storeLittleEndianUInt48(dst[6:12], x)
}
