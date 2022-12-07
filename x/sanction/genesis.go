package sanction

func NewGenesisState(addrs []string) *GenesisState {
	return &GenesisState{
		SanctionedAddresses: addrs,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil)
}

func (g GenesisState) Validate() error {
	// TODO[1046]: Implement GenesisState.Validate
	panic("not implemented")
}
