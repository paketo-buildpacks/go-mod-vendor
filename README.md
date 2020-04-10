# Go Mod Cloud Native Buildpack

## Integration

The Go Mod CNB provides go-mod as a dependency. Downstream
buildpacks can require the node dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Go Mod dependency is "go-mod". This value is
  # considered part of the public API for the buildpack and will not change
  # without a plan for deprecation.
  name = "go-mod"

  # Note: The version field is unsupported as there is no version for a set of
  # go-mod.

  # The Go Mod buildpack does not support any non-required metadata options.
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## Applications outside the root directory

If your application's main package is not in the root of the directory, then you'll need to specify this in the `buildpack.yml` file. Here's an example of how to do that:

```yaml
go:
  targets: ["./cmd/web"]
```
