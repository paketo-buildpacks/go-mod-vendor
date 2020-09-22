# Go Mod Vendor Cloud Native Buildpack

The Go Mod Vendor CNB executes the [`go mod
vendor`](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more)
command in the app's working directory to make vendored copy of dependencies.

## Integration

The Go Mod Vendor CNB does not provide any dependencies. In order to
execute the `go mod vendor` command, the buildpack requires the `go`
dependency that can be provided by a buildpack like the [Go Distribution
CNB](https://github.com/paketo-buildpacks/go-dist).

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can
supply another value as the first argument to package.sh.

## `buildpack.yml` Configuration

The Go Mod Vendor buildpack does not support configurations via `buildpack.yml`.

## Go Version

This buildpack will request the latest minor version of the `major.minor`
version it finds in the `go.mod` file from the `go-dist` buildpack.
