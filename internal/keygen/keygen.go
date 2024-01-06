package keygen

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

func GenerateKey() string {
	return base58.Encode(uuid.New().NodeID())
}
