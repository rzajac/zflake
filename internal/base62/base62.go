package base62

import (
	"math"
	"unsafe"
)

// alpha represents base62 alphabet [A-Z][a-z][0-9].
const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Number of characters in the alphabet.
const base = 62

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
		res = base*res + uint64(alphabetMap[rune(str[i])])
	}
	return res, nil
}
