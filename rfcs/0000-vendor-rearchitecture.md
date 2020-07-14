# Go Mod Vendor Rearchitecture 

## Proposal

The main functionality provided by the `go-mod-vendor` buildpack will be to
provide vendored dependencies using the `go mod vendor` command. More
information about go modules can be found [here](https://golang.org/ref/mod).

`go mod vendor` first looks for the `go.mod` file in the app root directory,
which would be there as a result of running `go mod init` before the buildpack
is run. This would be the app developer's responsibility if they want
to use go modules.

The `go.mod` file lives at the root of the app directory and  contains the app
dependencies, and gets updated by various go commands such as `go get`, `go
mod tidy` or in this case `go mod vendor`.

The [official documentation](https://golang.org/ref/mod#tmp_25) explains how go
modules deal with vendoring:

> The `go mod vendor` command constructs a directory named `vendor` in the main
> module's root directory that contains copies of all packages needed to
> support builds and tests of packages in the main module.

> Once vendoring is enabled, packages are loaded from the `vendor` directory
> instead of accessing the network or the module cache.

## Integration

The proposed buildpack requires `go` and provides none.

The former `go-mod` buildpack used to provide `go-mod`.

## Renaming

The buildpack will be renamed to `go-mod-vendor` from `go-mod`. This will more
clearly illustrate the specific function of the buildpack.

The original `go-mod` buildpack contains logic that deals with binary building.
Binary building logic will be removed from this buildpack and instead becomes a
responsibility of the [`go-build`](https://github.com/paketo-buildpacks/go-build) buildpack.

## Implementation

Detection will pass if a `go.mod` file is present in the app's source code.

On detection, the buildpack will perform the following steps in the Build
phase:

- Retrieve `dependencies` layer if it exists and create a new layer if it does not.

  - If the `dependencies` layer exists, compare the checksums of the current
    app directory with that of the previous build. This is possible because the
    checksum ignores the presence of a symlink.

    - If the checksums are the same, then we do not need to update any of the
      vendored dependencies, and we can reused the cached `vendor` directory
      from the `dependencies` layer.

    - If the checksums are not the same, then we need to update our vendored
      dependencies. To do this, we follow the steps below:

- If the `dependencies` layer does not exist OR we have the case above (checksums are different) then:

- Run `go mod vendor` from the app's root directory which downloads dependency
  source code to a `vendor` directory.

- Move the populated `vendor` directory to a separate `dependencies` layer

- Create a symlink in the app directory to the `dependencies` layer. The
  symlink is respected by the `go build` command that would run in the
  `go-build` buildpack, eventually allowing for reuse of the `vendor` directory between builds when able.

- Set the `cache` metadata flag to `true` for the vendored dependencies layer
  to be used in subsequent builds.

## GOPATH 

A benefit of using go modules is that it does not use the `GOPATH` to locate
dependencies. Because of this, this iteration of the buildpack will relinquish
the responsibility of setting the `GOPATH` variable. This responsibility now
lies with the `go-build` buildpack.

## Motivation

Using go modules provides a simpler way to manage app dependencies (without the
need of the `GOPATH`), and allows for caching of dependencies for faster
subsequent builds. A big effort of the CNB project is to move towards
buildpacks that are responsible for a single function making the whole buildpack
family structure more flexible, lightweight, and easier to maintain. With the
proposed changes, the sole responsibility of this buildpack becomes managing
application dependencies with go modules. As mentioned, the binary building
logic gets offloaded to the `go-build` buildpack.

## Source Material (Optional)

[Golang Documentation on Vendoring](https://golang.org/ref/mod#tmp_25)
