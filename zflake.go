// Package zflake implements distributed unique ID generator inspired
// by Twitter's Snowflake.
//
// A zflake ID is composed of
//   39 bits for time in units of 10 msec
//   16 bits for a sequence number
//    8 bits for a generator ID (GID)
//
// Above bit assigment dictate following zflake properties:
//
// - The lifetime of 174 years since the start of zflake epoch.
// - Can generate at most 2^16 IDs per 10ms for each generator ID.
// - Only 2^8 generators.
//
package zflake

import (
	"sync"
	"time"

	"github.com/rzajac/clock"
)

const (
	// BucketLen defines the length of zflake time bucket in nanoseconds.
	BucketLen = int64(10 * time.Millisecond)

	// BitLenTim number of bits assigned to time buckets.
	BitLenTim = 39

	// BitLenSeq number of bits assigned to sequence number.
	BitLenSeq = 16

	// BitLenGID number of bits assigned to generator ID (GID).
	BitLenGID = 63 - BitLenTim - BitLenSeq

	// DefaultEpoch represents default zflake epoch 2020-01-01T00:00:00Z as
	// nanoseconds since Unix epoch (1970-01-01T00:00:00Z).
	DefaultEpoch int64 = 1577836800000000000

	// DefaultGID represents default generator ID.
	DefaultGID byte = 0
)

// Bit masks.
const (
	maskTim = uint64((1<<(BitLenTim) - 1) << (BitLenSeq + BitLenGID))
	maskSeq = uint64((1<<BitLenSeq - 1) << BitLenGID)
	maskGID = uint64(1<<BitLenGID - 1)
	maskMSB = uint64(1 << (BitLenTim + BitLenSeq + BitLenGID))
)

// GID is Gen constructor option setting generator ID.
//
// If creating multiple generators it's up to the user to provide ID which is
// unique across generators/machines.
func GID(mid byte) func(*Gen) {
	return func(flake *Gen) {
		flake.gid = mid
	}
}

// Epoch is Gen constructor option setting zflake epoch.
func Epoch(epoch time.Time) func(*Gen) {
	return func(flake *Gen) {
		flake.epoch = uint64(DefaultEpoch / BucketLen)
		flake.epochns = epoch.UTC().UnixNano()
	}
}

// Clock is Flake constructor option injecting custom clock.
//
// This option is mostly used to test zflake behaviour.
func Clock(clk clock.Clock) func(*Gen) {
	return func(flake *Gen) {
		flake.now = clk
	}
}

// Gen represents distributed unique ID generator.
type Gen struct {
	epoch   uint64      // Number of zflake time buckets since Unix Epoch.
	epochns int64       // Generator epoch as nanoseconds.
	gid     byte        // Generator ID.
	bucket  uint64      // Current 10ms time bucket since epoch.
	seq     uint16      // Number of generated IDs in current time bucket.
	now     clock.Clock // Function returning current time.
	mx      *sync.Mutex // Guards generator.
}

// NewGen returns a new generator with default configuration.
// NewGen returns nil if epoch is set ahead of the current time.
func NewGen(opts ...func(*Gen)) *Gen {
	gen := &Gen{
		gid: DefaultGID,
		seq: uint16(1<<BitLenSeq - 1),
		now: clock.Now,
		mx:  &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(gen)
	}

	if gen.epoch == 0 {
		gen.epoch = uint64(DefaultEpoch / BucketLen)
		gen.epochns = DefaultEpoch
	} else if gen.epochns > gen.now().UTC().UnixNano() {
		return nil
	}

	return gen
}

// NextID generates a next unique ID.
// When 39 bit space for time buckets runs out NextID will panic.
func (gen *Gen) NextID() uint64 {
	gen.mx.Lock()
	defer gen.mx.Unlock()

	now := gen.bucketsSince(gen.now())
	if gen.bucket < now {
		gen.bucket = now
		gen.seq = 0
	} else {
		// Generating ID for the current bucket.
		gen.seq += 1
		if gen.seq == 0 {
			// We run out of IDs for current bucket.
			// Sleep till we are in the next one.
			gen.bucket++
			gen.sleep()
		}
	}

	if gen.bucket >= 1<<BitLenTim {
		panic("over the time limit")
	}

	return gen.bucket<<(BitLenSeq+BitLenGID) |
		uint64(gen.seq)<<BitLenGID |
		uint64(gen.gid)
}

// bucketsSince returns number of time buckets elapsed between the epoch and tim.
func (gen *Gen) bucketsSince(tim time.Time) uint64 {
	return uint64((tim.UTC().UnixNano() - gen.epochns) / BucketLen)
}

// sleep sleeps till the start of next bucket. It expects gen.bucket to point
// the bucket generator should wait (sleep) for.
func (gen *Gen) sleep() {
	next := gen.epochns + int64(gen.bucket)*BucketLen
	sleep := time.Duration(next - gen.now().UTC().UnixNano())
	time.Sleep(sleep)
}

// Decode decodes zflake ID.
func Decode(fid uint64) map[string]uint64 {
	msb := fid >> 63
	tim := fid >> (BitLenSeq + BitLenGID)
	seq := fid & maskSeq >> BitLenGID
	gid := fid & maskGID
	return map[string]uint64{
		"fid": fid,
		"msb": msb,
		"tim": tim,
		"seq": seq,
		"gid": gid,
	}
}
