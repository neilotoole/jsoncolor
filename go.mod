module github.com/neilotoole/jsoncolor

go 1.16

require (
	github.com/fatih/color v1.16.0
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
)

require (
	// Only used for benchmark/comparision.
	github.com/nwidger/jsoncolor v0.3.2

	// DO NOT UPGRADE: This functionality is only used in benchmark/tests, and
	// we're trying to stay synced with the version of segmentio/encoding
	// that we forked from. Although, maybe we should upgrade to the latest
	// go 1.16 compatible version.
	github.com/segmentio/encoding v0.3.6
	github.com/stretchr/testify v1.8.4
)
