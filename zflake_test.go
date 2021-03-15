package zflake

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/rzajac/clock"
	"github.com/stretchr/testify/assert"
)

func Test_NewGen(t *testing.T) {
	// --- Given ---
	clk := clock.Deterministic(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Millisecond,
	)
	flk := NewGen(Clock(clk))

	// --- When ---
	fid0 := flk.NextFID()

	// --- Then ---
	assert.NotNil(t, flk)

	parts := DecodeFID(fid0)
	assert.Exactly(t, int64(0x1000000), parts["fid"])
	assert.Exactly(t, int64(0), parts["msb"])
	assert.Exactly(t, int64(1), parts["tim"])
	assert.Exactly(t, int64(0), parts["seq"])
	assert.Exactly(t, int64(0), parts["gid"])
}

func Test_NewGen_setGID(t *testing.T) {
	// --- Given ---
	flk := NewGen(GID(42))

	// --- When ---
	fid0 := flk.NextFID()

	// --- Then ---
	assert.NotNil(t, flk)

	parts := DecodeFID(fid0)
	assert.Exactly(t, int64(42), parts["gid"])
}

func Test_NewGen_epochInTheFuture(t *testing.T) {
	// --- When ---
	flk := NewGen(Epoch(time.Now().Add(time.Hour)))

	// --- Then ---
	assert.Nil(t, flk)
}

func Test_Gen_NextFID_outOfTime(t *testing.T) {
	// --- Given ---
	epoch := time.Unix(0, DefaultEpoch)

	// Last possible bucket.
	endBucket := maskTim >> (BitLenSeq + BitLenGID)
	// Last bucket start time.
	endTimeNS := epoch.UTC().UnixNano() + endBucket*BucketLen
	endTime := time.Unix(0, endTimeNS)

	clk := clock.Deterministic(endTime, 10*time.Millisecond)
	flk := NewGen(Clock(clk))

	// --- When ---
	// Able to generate one ID but the next clock tick goes over the limit.
	flk.NextFID()

	// --- Then ---
	assert.PanicsWithValue(t, "over the time limit", func() { flk.NextFID() })
}

func Test_Gen_parallel(t *testing.T) {
	// --- Given ---
	generators := 500
	idPerGen := 10000
	totalIDs := generators * idPerGen
	set := make(map[int64]struct{}, totalIDs)

	// --- When ---
	flk := NewGen()

	// --- Then ---
	fidC := make(chan int64, 1000)
	for i := 0; i < generators; i++ {
		go func() {
			for i := 0; i < idPerGen; i++ {
				fidC <- flk.NextFID()
			}
		}()
	}

	for i := 0; i < totalIDs; i++ {
		id := <-fidC
		if _, ok := set[id]; ok {
			t.Fatal("duplicated id")
		}
		set[id] = struct{}{}
	}
	assert.Exactly(t, idPerGen*generators, len(set))
}

func Test_Gen_NextSID_DecodeFID(t *testing.T) {
	// --- Given ---
	clk := clock.Deterministic(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Millisecond,
	)
	flk := NewGen(Clock(clk))

	// --- When ---
	sid0 := flk.NextSID()

	// --- Then ---
	assert.Exactly(t, "18OWG", sid0)

	fid0, err := DecodeSID(sid0)
	assert.NoError(t, err)
	assert.Exactly(t, int64(0x1000000), fid0)
}

func Test_EncodeFID(t *testing.T) {
	assert.Exactly(t, "18OWG", EncodeFID(0x1000000))
}

func Benchmark_zflake_fid(b *testing.B) {
	b.StopTimer()
	flk := NewGen()
	var id int64

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		id = flk.NextFID()
	}
	_ = id
}

func Benchmark_zflake_sid(b *testing.B) {
	b.StopTimer()
	flk := NewGen()
	var id string

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		id = flk.NextSID()
	}
	_ = id
}

// printInt64 prints binary representation of int64 number.
func printInt64(fid int64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(fid))
	ret := ""
	for i, b := range buf {
		ret += fmt.Sprintf("%08b ", b)
		if i == 3 {
			ret += "  "
		}
		if i == 1 || i == 5 {
			ret += " "
		}
	}
	fmt.Println(ret)
}

// printInt64 prints binary representation of int64 number.
func printInt64Hex(fid int64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(fid))
	ret := ""
	for i, b := range buf {
		ret += fmt.Sprintf("%02X ", b)
		if i == 3 {
			ret += "  "
		}
		if i == 1 || i == 5 {
			ret += " "
		}
	}
	fmt.Println(ret)
}

// printUint16 prints binary representation of uint16 number.
func printUint16(seq uint16) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, seq)
	ret := ""
	for _, b := range buf {
		ret += fmt.Sprintf("%08b ", b)
	}
	fmt.Println(ret)
}
