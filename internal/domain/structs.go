package domain

import "github.com/ethereum/go-ethereum/common"

type KeyedLog struct {
	RootData   string
	ParentHash string
	BlockTime  uint64
}

// this is how I imagine we would use the root data
func (k *KeyedLog) GetRootDataAsByes() []byte {
	return common.Hex2Bytes(k.RootData)
}
