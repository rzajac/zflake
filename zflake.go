// Package zflake implements distributed unique ID generator inspired
// by Twitter's Snowflake.
//
// A zflake ID is composed of
//
//	 1 bit (most significant) reserved
//	38 bits for time in units of 10 msec
//	13 bits for a sequence number
//	12 bits for a generator ID (GID)
//
// Above bit assigment dictate following zflake properties:
//
// - The lifetime of ~87 years since the start of `zflake` epoch.
// - Can generate at most 2^13 (8192) IDs per 10ms for each generator ID.
// - 2^12 (4096) generators.
// - Ability to generate Base62 string representations of int64 IDs.
package zflake

import (
	"sync"
	"time"

	"github.com/rzajac/zflake/internal/base62"
)

const (
	// BucketLen defines the length of zflake time bucket in nanoseconds.
	BucketLen = int64(10 * time.Millisecond)

	// BitLenTim number of bits assigned to time buckets.
	BitLenTim = 38

	// BitLenSeq number of bits assigned to sequence number.
	BitLenSeq = 13

	// BitLenGID number of bits assigned to generator ID (GID).
	BitLenGID = 63 - BitLenTim - BitLenSeq

	// DefaultEpoch represents default zflake epoch 2020-01-01T00:00:00Z as
	// nanoseconds since Unix epoch (1970-01-01T00:00:00Z).
	DefaultEpoch int64 = 1577836800000000000

	// DefaultGID represents default generator ID.
	DefaultGID uint16 = 0
)

// Bit masks.
const (
	maskTim = int64((1<<(BitLenTim) - 1) << (BitLenSeq + BitLenGID))
	maskSeq = int64((1<<BitLenSeq - 1) << BitLenGID)
	maskGID = int64(1<<BitLenGID - 1)
	seqMax  = uint16(1<<BitLenSeq - 1)
	gidMax  = uint16(1<<BitLenGID - 1)
)

// GID is Gen constructor option setting generator ID. Will panic if gid is
// greater than 4095.
//
// It is caller responsibility to provide ID which is unique across
// generators / machines.
func GID(gid uint16) func(*Gen) {
	if gid > gidMax {
		panic("zflake GID out of bounds")
	}
	return func(flake *Gen) {
		flake.gid = gid
	}
}

// Epoch is Gen constructor option setting zflake epoch.
func Epoch(epoch time.Time) func(*Gen) {
	return func(flake *Gen) {
		flake.epoch = DefaultEpoch / BucketLen
		flake.epochns = epoch.UTC().UnixNano()
	}
}

// Clock is Flake constructor option injecting custom clock.
//
// This option is mostly used to test zflake behaviour.
func Clock(clk func() time.Time) func(*Gen) {
	return func(flake *Gen) {
		flake.now = clk
	}
}

// Gen represents distributed unique ID generator.
type Gen struct {
	epoch   int64            // Number of zflake time buckets since Unix Epoch.
	epochns int64            // Generator epoch as nanoseconds.
	gid     uint16           // Generator ID.
	bucket  int64            // Current 10ms time bucket since epoch.
	seq     uint16           // Number of generated IDs in current time bucket.
	now     func() time.Time // Function returning current time.
	mx      *sync.Mutex      // Guards generator.
}

// NewGen returns a new generator with default configuration.
// NewGen returns nil if epoch is set ahead of the current time.
func NewGen(opts ...func(*Gen)) *Gen {
	gen := &Gen{
		gid: DefaultGID,
		seq: uint16(1<<BitLenSeq - 1),
		now: time.Now,
		mx:  &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(gen)
	}

	if gen.epoch == 0 {
		gen.epoch = DefaultEpoch / BucketLen
		gen.epochns = DefaultEpoch
	} else if gen.epochns > gen.now().UTC().UnixNano() {
		return nil
	}

	return gen
}

// NextFID generates a next unique ID.
// When 39 bit space for time buckets runs out NextFID will panic.
func (gen *Gen) NextFID() int64 {
	gen.mx.Lock()
	defer gen.mx.Unlock()

	now := gen.bucketsSince(gen.now())
	if gen.bucket < now {
		gen.bucket = now
		gen.seq = 0
	} else {
		// Generating ID for the current bucket.
		gen.seq += 1
		if gen.seq > seqMax {
			// We run out of IDs for current bucket.
			// Sleep till we are in the next one.
			gen.bucket++
			gen.seq = 0
			gen.sleep()
		}
	}

	if gen.bucket >= 1<<BitLenTim {
		panic("zflake over the time limit")
	}

	return gen.bucket<<(BitLenSeq+BitLenGID) |
		int64(gen.seq)<<BitLenGID |
		int64(gen.gid)
}

// NextSID returns zflake id encoded using Base62 [0-9][A-Z][a-z].
func (gen *Gen) NextSID() string {
	return base62.Encode(gen.NextFID())
}

// bucketsSince returns number of time buckets elapsed between the epoch and tim.
func (gen *Gen) bucketsSince(tim time.Time) int64 {
	return (tim.UTC().UnixNano() - gen.epochns) / BucketLen
}

// sleep sleeps till the start of next bucket. It expects gen.bucket to point
// the bucket generator should wait (sleep) for.
func (gen *Gen) sleep() {
	next := gen.epochns + gen.bucket*BucketLen
	sleep := time.Duration(next - gen.now().UTC().UnixNano())
	time.Sleep(sleep)
}

// DecodeFID decodes zflake ID.
func DecodeFID(fid int64) map[string]int64 {
	msb := fid >> 63
	tim := fid >> (BitLenSeq + BitLenGID)
	seq := fid & maskSeq >> BitLenGID
	gid := fid & maskGID
	return map[string]int64{
		"fid": fid,
		"msb": msb,
		"tim": tim,
		"seq": seq,
		"gid": gid,
	}
}

// EncodeFID returns Base62 string representation of the zflake ID.
func EncodeFID(fid int64) string {
	return base62.Encode(fid)
}

// DecodeSID decodes Base62 representation of the zflake ID back to int64.
func DecodeSID(sid string) (int64, error) {
	return base62.Decode(sid)
}
