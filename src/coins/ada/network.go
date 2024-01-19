package ada

type Network struct {
	NetworkId     uint32
	ProtocolMagic uint64
}

var MainNet = Network{
	NetworkId:     0b0001,
	ProtocolMagic: 764824073,
}

var TestNet = Network{
	NetworkId:     0b0000,
	ProtocolMagic: 1097911063,
}
