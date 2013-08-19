package simon

const (
	roundsSimon48_72 = 36
	roundsSimon48_96 = 36
)

func leftRotate24(n uint32, shift uint) uint32 {
	return ((n << shift) & 0x00ffffff) | (n >> (24 - shift))
}

// Use NewSimon48 below to expand a Simon48 key. Simon48Cipher
// implements the cipher.Block interface.
type Simon48Cipher struct {
	k      []uint32
	rounds int
}

func littleEndianBytesToUInt24(b []byte) uint32 {
	return uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16)
}

func storeLittleEndianUInt24(dst []byte, n uint32) {
	dst[0] = byte(n)
	dst[1] = byte(n >> 8)
	dst[2] = byte(n >> 16)
}

// NewSimon48 creates and returns a new Simon48Cipher. It accepts
// either a 96-bit key (for Simon48/96) or a 128-bit key (for
// Simon48/128). See the documentation on Simon32 or the test suite
// for our endianness convention.
func NewSimon48(key []byte) *Simon48Cipher {
	cipher := new(Simon48Cipher)
	var keyWords int
	var z uint64

	switch len(key) {
	case 9:
		keyWords = 3
		z = zSeq0
		cipher.rounds = roundsSimon48_72
	case 12:
		keyWords = 4
		z = zSeq1
		cipher.rounds = roundsSimon48_96
	default:
		panic("NewSimon48() requires either a 72- or 96-bit key")
	}
	cipher.k = make([]uint32, cipher.rounds)
	for i := 0; i < keyWords; i++ {
		cipher.k[i] = littleEndianBytesToUInt24(key[3*i : 3*i+3])
	}
	for i := keyWords; i < cipher.rounds; i++ {
		tmp := leftRotate24(cipher.k[i-1], 21)
		if keyWords == 4 {
			tmp ^= cipher.k[i-3]
		}
		tmp ^= leftRotate24(tmp, 23)
		lfsrBit := (z >> uint((i-keyWords)%62)) & 1
		cipher.k[i] = ^cipher.k[i-keyWords] ^ tmp ^ uint32(lfsrBit) ^ 3
		cipher.k[i] &= 0x00ffffff
	}
	return cipher
}

// Simon48 has a 48-bit block length.
func (cipher *Simon48Cipher) BlockSize() int {
	return 6
}

func simonScramble24(x uint32) uint32 {
	return (leftRotate24(x, 1) & leftRotate24(x, 8)) ^ leftRotate24(x, 2)
}

// Encrypt encrypts the first block in src into dst.
// Dst and src may point at the same memory. See crypto/cipher.
func (cipher *Simon48Cipher) Encrypt(dst, src []byte) {
	y := littleEndianBytesToUInt24(src[0:3])
	x := littleEndianBytesToUInt24(src[3:6])
	for i := 0; i < cipher.rounds; i += 2 {
		y ^= simonScramble24(x) ^ cipher.k[i]
		x ^= simonScramble24(y) ^ cipher.k[i+1]
	}
	storeLittleEndianUInt24(dst[0:3], y)
	storeLittleEndianUInt24(dst[3:6], x)
}

// Decrypt decrypts the first block in src into dst.
// Dst and src may point at the same memory. See crypto/cipher.
func (cipher *Simon48Cipher) Decrypt(dst, src []byte) {
	y := littleEndianBytesToUInt24(src[0:3])
	x := littleEndianBytesToUInt24(src[3:6])
	for i := cipher.rounds - 1; i >= 0; i -= 2 {
		x ^= simonScramble24(y) ^ cipher.k[i]
		y ^= simonScramble24(x) ^ cipher.k[i-1]
	}
	storeLittleEndianUInt24(dst[0:3], y)
	storeLittleEndianUInt24(dst[3:6], x)
}