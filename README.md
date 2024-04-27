# solo-machine-go

TODO: Write more about this project (once I know what it is)

## Potential roadmap

- [ ] Basic solo machine with built-in relayer and one-way ICS20 support
  - [x] Single setup command that creates signer keys, clients, connections and an ICS20 channel
  - [ ] Command to send ICS20 packets
  - [ ] Status command to print out current state of chains and their clients, connections and channels
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
