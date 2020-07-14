# Go Mod Vendor Cloud Native Buildpack

The Go Mod Vendor CNB builds a Go application binary, using the [`go
mod vendor`](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more)
functionality provided by the Go Distribution CNB to package dependencies.

## Integration

The Go Mod Vendor CNB provides go-mod as a dependency. Downstream
buildpacks can require the go-mod dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Go Mod Vendor buildpack dependency is "go-mod". This value is
  # considered part of the public API for the buildpack and will not change
  # without a plan for deprecation.
  name = "go-mod"

  # Note: The version field is unsupported as there is no version for a set of
  # go-mod.

  # The Go Mod Vendor buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Go Mod Vendor buildpack
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Go Mod Vendor
    # during its build process, this flag should be set to true.
    build = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configuration

```yaml
go:
  # this allows you to override the location of the main package of the app
  targets: ["./cmd/web"]

  # this allows you to set Go ldflags for compilation
  ldflags:
    main.version: v1.2.3
    main.sha: 1234567
```
