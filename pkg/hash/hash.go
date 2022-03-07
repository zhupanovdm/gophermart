package hash

import (
	"crypto"
	"encoding/hex"
)

type StringFunc func(string) string

func StringWith(hash crypto.Hash) StringFunc {
	h := hash.New()
	return func(s string) string {
		return hex.EncodeToString(h.Sum([]byte(s)))
	}
}
