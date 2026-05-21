module github.com/liguangsheng/goost/lru/benchmark

go 1.25.10

require (
	github.com/bluele/gcache v0.0.2
	github.com/dgraph-io/ristretto/v2 v2.4.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/liguangsheng/goost v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/liguangsheng/goost => ../..
