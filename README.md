# Go Mod Vendor Cloud Native Buildpack

The Go Mod Vendor CNB builds a Go application binary, using the [`go
mod vendor`](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more)
functionality provided by the Go Distribution CNB to package dependencies.

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configuration

The Go Mod Vendor buildpack does not support configurations via `buildpack.yml`.
