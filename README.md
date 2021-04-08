# zflake

[![GoDoc](https://godoc.org/github.com/rzajac/zflake?status.svg)](http://godoc.org/github.com/rzajac/zflake)
[![Go Report Card](https://goreportcard.com/badge/github.com/rzajac/zflake)](https://goreportcard.com/report/github.com/rzajac/zflake)

Package `zflake` is a distributed unique ID generator inspired by
[Twitter's Snowflake](https://blog.twitter.com/2010/announcing-snowflake).

The `zflake` was created mainly to be able to generate great number of unique
int64 IDs in bursts. In fact the number of distributed generators
was less important than the ability to create a lot of IDs in a short time.

The `zflake` bit assignment in int64 is as follows:

    38 bits for time in units of 10 msec
    14 bits for a sequence number
    11 bits for a generator ID (GID)

`zflake` properties:

- The lifetime of 87 years since the start of `zflake` epoch.
- Can generate at most 2^14 IDs per 10ms for each generator ID.
- 2^11 generators.
- Ability to generate Base62 string representations of int64 IDs. 

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
fid := gen.NextFID() // Generate unique int64 ID.
sid := gen.NextSID() // Generate unique Base62 encoded ID.
```

By default `zflake` uses `2020-01-01T00:00:00Z` as an epoch.

When 39 bit space for time buckets (10ms) runs out `NextFID` and `NextSID` 
will **panic**.

## Decode `zflake` ID

```
parts := zflake.DecodeFID(134362890512629802)
fmt.Println(parts) // map[fid:134362890512629802 gid:42 msb:0 seq:0 tim:4004326180]
```

## Benchmark

```
Benchmark_zflake
Benchmark_zflake_fid-12     1973349      608 ns/op      0 B/op      0 allocs/op
Benchmark_zflake_sid-12     1967682      610 ns/op     16 B/op      1 allocs/op
```

## License

BSD-2-Clause,
see [LICENSE](https://github.com/rzajac/zflake/blob/master/LICENSE) for details.