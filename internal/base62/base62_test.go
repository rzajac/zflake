package base62

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Encode_Decode(t *testing.T) {
	tt := []struct {
		sid string
		fid int64
	}{
		{"4MV1b01dcO", 59061089258255360},
		{"1ZlfarV", 89569285645},
		{"pt0", 0x30b1e},
		{"pt1", 0x30b1f},
		{"pt2", 0x30b20},
		{"18OWH", 0x1000001},
	}

	for _, tc := range tt {
		t.Run(tc.sid, func(t *testing.T) {
			// --- When ---
			fid, err := Decode(tc.sid)
			sid := Encode(tc.fid)

			// --- Then ---
			assert.NoError(t, err)
			assert.Exactly(t, tc.fid, fid)
			assert.Exactly(t, tc.sid, sid)
		})
	}
}

func Test_Decode_checkAlphabet(t *testing.T) {
	// --- When ---
	fid, err := Decode("%%%")

	// --- Then ---
	assert.ErrorIs(t, err, ErrInvalidSID)
	assert.Exactly(t, int64(0), fid)
}

func Test_Encode_Zero(t *testing.T) {
	// --- When ---
	sid := Encode(0)

	// --- Then ---
	assert.Exactly(t, "0", sid)

	fid, err := Decode("0")
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), fid)
}

func BenchmarkBase62Encode(b *testing.B) {
	id := int64(59061089258255360)
	b.ReportAllocs()
	var sid string
	for i := 0; i < b.N; i++ {
		sid = Encode(id)
	}
	_ = sid
}

func BenchmarkBase62Decode(b *testing.B) {
	var err error
	var dec int64
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dec, err = Decode("04MV1b01dO")
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = dec
}
