package base62

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Encode_Decode(t *testing.T) {
	tt := []struct {
		fid uint64
		exp string
	}{
		{59061089258255360, "4MV1b01dcO"},
		{89569285645, "1ZlfarV"},
	}

	for _, tc := range tt {
		t.Run(tc.exp, func(t *testing.T) {
			// --- When ---
			sid := Encode(tc.fid)
			fid, err := Decode(sid)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, tc.exp, sid)
			assert.Exactly(t, tc.fid, fid)
		})
	}
}

func Test_Encode_Zero(t *testing.T) {
	// --- When ---
	sid := Encode(0)

	// --- Then ---
	assert.Exactly(t, "0", sid)

	fid, err := Decode("0")
	assert.NoError(t, err)
	assert.Exactly(t, uint64(0), fid)
}

func BenchmarkBase62Encode(b *testing.B) {
	id := uint64(59061089258255360)
	b.ReportAllocs()
	var sid string
	for i := 0; i < b.N; i++ {
		sid = Encode(id)
	}
	_ = sid
}

func BenchmarkBase62Decode(b *testing.B) {
	var err error
	var dec uint64
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dec, err = Decode("04MV1b01dO")
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = dec
}
