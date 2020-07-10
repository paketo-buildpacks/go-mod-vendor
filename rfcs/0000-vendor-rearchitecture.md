# Go Mod Vendor Rearchitecture 

## Proposal

The main functionality provided by the go-mod-vendor buildpack will be to
invoke the `go mod vendor` command. More information about go modules can be
found [here](https://golang.org/ref/mod).

`go mod vendor` first looks for the `go.mod` file in the app root directory,
which would be there as a result of running `go mod init` before the buildpack
gets run.

The `go.mod` file will contain the app dependencies, and gets updated by
different go commands such as `go get`, `go mod tidy` or in this case `go mod
vendor`.

The [official documentation](https://golang.org/ref/mod#tmp_25) explains how go modules deal with vendoring:

> The go mod vendor command constructs a directory named vendor in the main
> module's root directory that contains copies of all packages needed to
> support builds and tests of packages in the main module.

> Once vendoring is enabled, packages are loaded from the vendor directory
> instead of accessing the network or the module cache.

## Renaming

The buildpack will be renamed to go-mod-vendor from go-mod. Renaming the
buildpack reflects the single function of the buildpack which is running `go
mod vendor`.

The original go-mod buildpack contains logic that deals with binary building,
which will be removed. Instead, this logic has become a responsibility of the
[go-build](https://github.com/paketo-buildpacks/go-build) buildpack.

## Implementation

Detection will pass if a `go.mod` file is present in the app's source code.

On detection, the buildpack will perform the following, it will:

- Run `go mod vendor` from the app's root directory which downloads dependency source code to a `vendor` directory.

- Move `vendor` directory to a separate `dependencies` layer

- Create a symlink in the app directory to the `dependencies` layer. The symlink is
  respected by the go build command that would run in the go-build buildpack,
  and provides the added benefit of keeping the vendor directory separate from
  the app code. (why is this a benefit?)

- Set the `cache` metadata flag to `true` for the vendored dependencies layer.

## GOPATH 

Since the `go mod vendor` command does not use the GOPATH to locate
dependencies, this iteration of the buildpack will relinquish the
responsibility of setting the GOPATH. This responsibility now lies with the
`go-build` buildpack.

## Motivation

Using go modules provides a simpler way to manage app dependencies (without the
need of the GOPATH), and allows for caching of dependencies for faster
subsequent builds. A big effort of the CNB project is to move towards
buildpacks that are responsible for a single function makes the whole buildpack
family structure more flexible, lightweight, and easier to maintain. With the
proposed changes, the sole responsibility of this buildpack becomes managing
application dependencies with go modules. As mentioned, the binary building
logic gets offloaded to the go-build buildpack.

## Source Material (Optional)

[Golang Documentation on Vendoring](https://golang.org/ref/mod#tmp_25)

