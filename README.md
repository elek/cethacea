# cethacea

Cethacea is a command line Ethereum client and Swiss-army toolset.

## Status

> All things are in motion like streams

This toolset is created for experiments and subject to incompatible change at any time. You are warned.

If you are not ready to check the source code for more information, probably it's not for you, sorry. As of now this is not product just some code here.

## Configuration

Most of the subcommands requires (1) a chain API endpoint (2) a public/private key pair. Some subcommands also require
a (3) contract address.

**Chain/API endpoint** can be configured with either `--chain` or `CETH_CHAIN` environment variable or a `chain` YAML
key in the `.ceth.yaml` of the current directory.

It can be either:

* An RPC URL of the chain
* An alias

If it's an alias, the corresponding configuration should be in the file `~/.config/cethacea/chains.yaml`.

You can add new entries to this config file with:

```
ceth chain add alias https://....
```

**Private key / address** can be configured with `--account` CLI argument or with the `CETH_ACCOUNT` environment
variable or with a `account` key in the `.ceth.yaml` file of the current directory.

* A private key in hex format
* A file which contains the private key in hex format
* An alias

If it's an alias, the key should be configured in the `.accounts.yaml` file of the local directory.

You can generate new key to this file with

```
ceth account generate
```

And check available keys with

```
ceth account list
```

If no account is defined an ephemeral will be generated (different for each operations).

**Contract address** can be configured with `--contract` CLI argument or with the `CETH_CONTRACT` environment variable
or with a `contract` key in the `.ceth.yaml` file of the current directory.

It can be either:

* A hex address of the contract
* Alias

Aliases are resolved from the `.contract.yaml` file.

New alias can be recorded with:

```
ceth account add NAME 0xADD0E00
```

Contract address is required only for special subcommands.
