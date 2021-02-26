package base62

import (
	"errors"
	"math"
	"unsafe"
)

// alpha represents base62 alphabet [0-9][A-Z][a-z].
const alpha = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Number of characters in the alphabet.
const base = 62

// ErrInvalidSID an error decoding string representation of the zflake.
var ErrInvalidSID = errors.New("invalid zflake string representation")

// alphabetMap maps base62 alphabet rune to its index.
var alphabetMap = make(map[rune]int, base)

func init() {
	for i := 0; i < len(alpha); i++ {
		alphabetMap[rune(alpha[i])] = i
	}
}

// Encode encodes uint64 number as base62 string.
func Encode(id uint64) string {
	if id == 0 {
		return "0"
	}
	n := int(math.Log(float64(id))/math.Log(base) + 1)
	buf := make([]byte, n)
	r := id % base
	q := id / base
	buf[n-1] = alpha[r]

	for i := n - 2; i >= 0; i-- {
		r = q % base
		q = q / base
		buf[i] = alpha[r]
	}
	return *(*string)(unsafe.Pointer(&buf))
}

// Decode decodes base62 string to uint64.
func Decode(str string) (uint64, error) {
	res := uint64(0)
	for i := 0; i < len(str); i++ {
		r := rune(str[i])
		v, ok := alphabetMap[r]
		if !ok {
			return 0, ErrInvalidSID
		}
		res = base*res + uint64(v)
	}
	return res, nil
}
