# solo-machine-go

TODO: Write more about this project (once I know what it is)

## Potential roadmap

- [ ] Basic solo machine with built-in relayer and ICS20 support
  - [x] Single setup command that creates signer keys, clients, connections and an ICS20 channel
  - [ ] Status command to print out current state of chains and their clients, connections and channels
  - [ ] Command to send ICS20 packets
- [x] Multi-chain support
- [ ] More IBC application support
- [ ] Separate out the signer (i.e. make it more configurable)
- [ ] Multi-sig support

Maybe:
- [ ] External relayer support? (undecided if this even makes sense)
- [ ] External state machine support? (a separate solo machine state from all the plumbing - i.e. build your own state machine/whatever)
