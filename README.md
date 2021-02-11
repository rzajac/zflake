# zflake

[![GoDoc](https://godoc.org/github.com/rzajac/zflake?status.svg)](http://godoc.org/github.com/rzajac/zflake)
[![Go Report Card](https://goreportcard.com/badge/github.com/rzajac/zflake)](https://goreportcard.com/report/github.com/rzajac/zflake)

Package `zflake` is a distributed unique ID generator inspired by
[Twitter's Snowflake](https://blog.twitter.com/2010/announcing-snowflake).

The `zflake` was created mainly to be able to generate great number of unique
uint64 IDs in bursts. In fact the number of distributed generators
was less important than the ability to create a lot of IDs in a short time.

The `zflake` bits assignment in uint64 is as follows:

    39 bits for time in units of 10 msec
    16 bits for a sequence number
     8 bits for a generator ID (GID)

Above bit assigment dictate following `zflake` properties:

- The lifetime of 174 years since the start of `zflake` epoch.
- Can generate at most 2^16 IDs per 10ms for each generator ID.
- Only 2^8 generators.

## Installation

```
go get github.com/rzajac/zflake
```

## Usage

To create `zflake` generator with default configuration (GID 0, epoch starting 
at 2020-01-01T00:00:00Z) call constructor function without options:

```
func NewGen() *Gen
```

You can customize `zflake` using constructor option functions:

```
func GID(gid byte)          // Set generator ID.
func Epoch(epoch time.Time) // Set epoch. 
```

Example:

```
gen := zflake.NewGen(zflake.GID(42), zflake.Epoch(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)))
fid := gen.NextID()
```

By default `zflake` uses `2020-01-01T00:00:00Z` as an epoch.

When 39 bit space for time buckets (10ms) runs out `NextID` will **panic**.

## Decode `zflake` ID

```
parts := zflake.Decode(59061089258255360)
fmt.Println(parts) // map[fid:59061089258255360 gid:0 msb:0 seq:18740 tim:3520315245]
```

## Benchmark

```
Benchmark_zflake
Benchmark_zflake-12    	 7974894      155 ns/op      0 B/op      0 allocs/op
```

## License

BSD-2-Clause,
see [LICENSE](https://github.com/rzajac/zflake/blob/master/LICENSE) for details.