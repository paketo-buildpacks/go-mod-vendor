# Go Mod Cloud Native Buildpack
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
