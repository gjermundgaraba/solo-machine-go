# Solo Machine

A simple IBC solo machine written in Go. For now this is just a proof of concept for learning purposes.
I would, however, love to make this into a production-ready tool if there are real-world use cases for it.  
Get in touch if you think you might have one of those, and we can talk about how to make that happen.

## Current state

The Solo Machine is a simple command line tool that can do the following:
* Initialize a full IBC setup (`solo-machine init`)
  * Including creating a signer key, clients, connections and an ICS20 channel to a tendermint chain
* Send ICS20 packets (`solo-machine transfer`)
* Print out the current state of the chains and their clients, connections and channels (`solo-machine status`)
* Update light clients (`solo-machine update`)
* Relay all its own packets from solo-machine to chain (only)
* Supports multiple chains

Storage and configuration:
* Chain connection configuration is stored in a `config.toml` file under `~/.solo-machine`
  * It will be initialized the first time you just run an empty `solo-machine init` command
* Solo-machine storage (light clients and keys and stuff) is saved in a database file under `~/.solo-machine`

Current limitations:
* The solo machine itself has no state machine or storage outside storing keys, client, connections and channels.
  * That means that currently it does not know anything about its own balances or accounts or anything like that
* Tests and documentation is lacking right now, mostly because it has just been a learning-based project so far.
  * I do hope to change that :)

## Installation

```shell
$ make install
```

## Potential roadmap

- [x] Basic solo machine with built-in relayer and one-way ICS20 support
  - [x] Single setup command that creates signer keys, clients, connections and an ICS20 channel
  - [x] Command to send ICS20 packets
  - [x] Status command to print out current state of chains and their clients, connections and channels
- [x] Multi-chain support
- [ ] Receive ICS20 packets (?)
- [ ] Unit testing
- [ ] Integration testing with interchaintest
- [ ] More IBC application support
- [ ] Separate out the signer (i.e. make it more configurable)
- [ ] Multi-sig support

Maybe:
- [ ] External relayer support? (undecided if this even makes sense)
- [ ] External state machine support? (a separate solo machine state from all the plumbing - i.e. build your own state machine/whatever)

Some other cleanups that should be done:
- [ ] Make the separation of concerns clearer between solomachine and storage
