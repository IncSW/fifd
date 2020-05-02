[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/IncSW/go-51degrees?style=flat-square)](https://goreportcard.com/report/github.com/IncSW/go-51degrees)

Device Detection based on 51Degrees Trie. Golang version of [51Degrees/Device-Detection](https://github.com/51Degrees/Device-Detection) without cgo.

## Installation

`go get github.com/IncSW/go-51degrees`

```go
import fiftyonedegrees "github.com/IncSW/go-51degrees"
```

## Quick Start

```go
reader, err := fiftyonedegrees.NewReaderFromFile("path/to/51DegreesV3.4.trie")
if err != nil {
	panic(err)
}
device := reader.MatchDevice("UserAgent")
fmt.Println(device.GetValue("BrowserName"), device.GetValue("BrowserVersion"))
```

## Performance

### Go 1.14.2, Debian 9.1, i7-7700

fixed ua: `Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:60.0) Gecko/20100101 Firefox/60.0`

```
go/fixed-8               1291940    927 ns/op     0 B/op    0 allocs/op
cgo/fixed-8               533149   2214 ns/op   224 B/op   19 allocs/op

go/fixed-parallel-8      6384787    189 ns/op     0 B/op    0 allocs/op
cgo/fixed-parallel-8     2139506    554 ns/op   224 B/op   19 allocs/op

go/range-8                422196   2720 ns/op     0 B/op    0 allocs/op
cgo/range-8               404926   2964 ns/op   235 B/op   18 allocs/op

go/range-parallel-8      2738305    430 ns/op     0 B/op    0 allocs/op
cgo/range-parallel-8     1810068    657 ns/op   235 B/op   18 allocs/op
```

## License

[MIT License](LICENSE).
