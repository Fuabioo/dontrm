# dontrm

Don't remove your system 🤡

## Installation

TODO download

Set up an alias for user shell usage:

```sh
alias rm="dontrm"
```

Or set up a function if you want this to also apply to
CLI's or scripts:

```sh
rm() {
    dontrm $@
}
```

## Build from source

```sh
go install
```

Or

```sh
go build && sudo mv dontrm /usr/bin/dontrm
```

## Usage

Executing the following should be safe:
(don't test it on your system though)

```sh
sudo dontrm -fr /*
```

```sh
dontrm ./path/or/file/to/remove
```

You can also use a `DRY_RUN` environment variable
to prevent any changes from happening.

```sh
DRY_RUN=1 dontrm ./path/or/file/to/remove
```

## TODO

- [x] Set up goreleaser
- [ ] Implement installation script
- [ ] Configurable rm path
- [ ] Mount a virtualized environment to test the more dangerous commands
