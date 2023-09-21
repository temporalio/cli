# Develop

This doc is for contributors to Temporal CLI (hopefully that's you!)

[comment]: <> (**Note:** All contributors also need to fill out the [Temporal Contributor License Agreement]&#40;https://gist.github.com/samarabbas/7dcd41eb1d847e12263cc961ccfdb197&#41; before we can merge in any of your changes.)

## Prerequisites

### Build prerequisites

-   [Go Lang](https://golang.org/) (minimum version required is 1.19):
    -   Install on macOS with `brew install go`.
    -   Install on Ubuntu with `sudo apt install golang`.

## Check out the code

Temporal CLI uses go modules, there is no dependency on `$GOPATH` variable. Clone the repo into the preferred location:

```bash
git clone https://github.com/temporalio/cli.git
```

## Makefile

There is a very simple `Makefile` available for the convenience of UNIXy
developers which covers most of the points below. The default target will do a
clean build (`make clean && make build`), and `make test` will run tests.

## Build

Build the `temporal` binary:

```bash
go build ./cmd/temporal
```

## Generate docs

```bash
go run ./cmd/docgen
```

Docs are generated as Markdown files in `./docs`.

## Run tests

Run all tests:

```bash
go test ./...
```

## Run Temporal CLI locally

By default the server runs in in-memory mode:

```bash
go run ./cmd/temporal server start-dev
```

Pass `--db-filename` to persist the state in an SQLite DB

Now you can create default namespace:

```bash
temporal namespace register default
```

and run samples from [Go](https://github.com/temporalio/samples-go) and [Java](https://github.com/temporalio/samples-java) samples repos.

When you are done, press `Ctrl+C` to stop the server.

## License headers

This project is Open Source Software, and requires a header at the beginning of
all source files. To verify that all files contain the header execute:

```bash
go run ./cmd/copyright
```

## Third party code

The license, origin, and copyright of all third party code is tracked in `LICENSE-3rdparty.csv`.
To verify that this file is up to date execute:

```bash
go run ./cmd/licensecheck
```

## Release process

If you're a Temporal engineer / code owner, here's how to do a release:

1. Do some manual testing to make sure things look good--start a dev-server, run
   some workflows, etc.
2. Make sure CI is passing on `main`.
3. Create a tag of the form `vX.Y.Z` off of `main` and push it to GitHub.
4. GoReleaser will automatically build and publish the release artifacts.
5. The `temporal.download` links will automatically update once the artifacts
   are available.
6. Open a PR against Homebrew to modify [the temporal formula] to fetch the
   source tarball from the latest release on GitHub.  Don't forget to update
   the checksum of the source tarball. ([Example PR])
   - Homebrew should take care of rebottling the formula and updating the bottle
     checksums; you don't need to do this yourself.
7. Re-generate the docs, and follow [the docs team's process] to update [the
   CLI documentation].

For reference, here are all of the places where releases get published:
https://docs.temporal.io/cli/#installation

[the temporal formula]: https://github.com/Homebrew/homebrew-core/blob/fc3afa61f03205d6d88f5443e805a7f782171aa2/Formula/t/temporal.rb
[Example PR]: https://github.com/Homebrew/homebrew-core/commit/23f09be7fe7e2b00eee53a12db9a5a15c1c206ff
[the docs team's process]: https://github.com/temporalio/documentation/blob/main/README.md#how-to-make-changes-to-this-repository
[the CLI documentation]: https://github.com/temporalio/documentation/tree/main/docs-src/cli
