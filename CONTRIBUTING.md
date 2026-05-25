# Contributing to jsoncolor

Thanks for your interest in improving `jsoncolor`! This document covers how to
get set up, what's expected of a change, and how releases work.

## Background

`jsoncolor` is a drop-in replacement for the standard library's `encoding/json`
that emits colorized JSON. It is layered onto a fork of
[`segmentio/encoding`](https://github.com/segmentio/encoding); much of the code
(and many of the tests) is inherited from that project. See
[`SEGMENTIO_README.md`](SEGMENTIO_README.md) for the upstream documentation.

## Prerequisites

- Go **1.25** or newer (see the `go` directive in [`go.mod`](go.mod)).
- [`golangci-lint`](https://golangci-lint.run) v2 (CI runs v2.12.2).

## Development

```shell
go build ./...
go test ./...
golangci-lint run ./...
```

Run the encoding benchmarks with:

```shell
go test -bench=BenchmarkEncode -benchtime=5s
```

## Submitting a change

1. Fork the repository and create a topic branch.
2. Make your change with accompanying tests. New behavior should be covered;
   bug fixes should include a regression test.
3. Ensure `go build ./...`, `go test ./...`, and `golangci-lint run ./...` all
   pass locally.
4. Add a bullet to the `## CHANGELOG` section of [`README.md`](README.md) under
   the current (unreleased) version heading, referencing the issue or PR number.
5. Open a pull request against `master`. CI must pass, and the branch requires
   one approving review before it can be merged.

Please keep pull requests focused; unrelated refactors are much easier to review
as separate PRs.

## Reporting bugs and requesting features

Use the GitHub [issue tracker](https://github.com/neilotoole/jsoncolor/issues).
For security vulnerabilities, follow [`SECURITY.md`](SECURITY.md) instead of
opening a public issue.

## Releases

Maintainers cut a release by pushing an annotated `vX.Y.Z` tag to `origin`. The
`## CHANGELOG` section in [`README.md`](README.md) is the human-readable
changelog. This project uses tags as the release mechanism and does not publish
GitHub Releases.
