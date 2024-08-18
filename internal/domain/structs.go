package domain

import "github.com/ethereum/go-ethereum/common"

type KeyedLog struct {
	RootData   string `json:"rootData"`
	ParentHash string `json:"parentHash"`
	BlockTime  uint64 `json:"blockTime"`
}

// this is how I imagine we would use the root data
func (k *KeyedLog) GetRootDataAsByes() []byte {
	return common.Hex2Bytes(k.RootData)
}
