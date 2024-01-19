package types

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
	"time"
)

type TipSetKey []cid.Cid

type Version struct {
	Version    string
	APIVersion uint32
	BlockDelay uint64
}

type BeaconEntry struct {
	Round uint64
	Data  []byte
}

type IpldObject struct {
	Cid cid.Cid
	Obj interface{}
}

type HeadChange struct {
	Type string
	Val  *TipSet
}

type ObjStat struct {
	Size  uint64
	Links uint64
}

type MessageSendSpec struct {
	MaxFee abi.TokenAmount `json:"MaxFee"`
}

type BlockMessages struct {
	BlsMessages   []*Message       `json:"BlsMessages"`
	SecpkMessages []*SignedMessage `json:"SecpkMessages"`
	Cids          []cid.Cid        `json:"Cids"`
}

type Actor struct {
	Code    cid.Cid         `json:"Code"`
	Head    cid.Cid         `json:"Head"`
	Nonce   uint64          `json:"Nonce"`
	Balance decimal.Decimal `json:"Balance"`
}

type Ticket struct {
	VRFProof []byte
}

type TipSet struct {
	Cids   []cid.Cid
	Blocks []*BlockHeader
	Height int64
}

type MessageReceipt struct {
	ExitCode int64
	Return   []byte
	GasUsed  int64
}

type Loc struct {
	File     string
	Line     int
	Function string
}

type GasTrace struct {
	Name string

	Location          []Loc `json:"loc"`
	TotalGas          int64 `json:"tg"`
	ComputeGas        int64 `json:"cg"`
	StorageGas        int64 `json:"sg"`
	TotalVirtualGas   int64 `json:"vtg"`
	VirtualComputeGas int64 `json:"vcg"`
	VirtualStorageGas int64 `json:"vsg"`

	TimeTaken time.Duration `json:"tt"`
	Extra     interface{}   `json:"ex,omitempty"`
}

type ExecutionTrace struct {
	Msg        *Message
	MsgRct     *MessageReceipt
	Error      string
	Duration   time.Duration
	GasCharges []*GasTrace

	Subcalls []ExecutionTrace
}

type InvocResult struct {
	Msg            *Message
	MsgRct         *MessageReceipt
	ExecutionTrace ExecutionTrace
	Error          string
	Duration       time.Duration
}

type MsgLookup struct {
	Message   cid.Cid // Can be different than requested, in case it was replaced, but only gas values changed
	Receipt   MessageReceipt
	ReturnDec interface{}
	TipSet    TipSetKey
	Height    int64
}
