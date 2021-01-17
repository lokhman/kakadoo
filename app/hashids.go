package app

import (
	"github.com/speps/go-hashids"
)

type HashID struct {
	hashID *hashids.HashID
}

func NewHashID(prefix string) *HashID {
	hd := hashids.NewData()
	hd.MinLength = 8
	hd.Salt = prefix + SecretKey
	h, _ := hashids.NewWithData(hd)
	return &HashID{h}
}

func (h *HashID) Encode(id int) string {
	hash, _ := h.hashID.Encode([]int{id})
	return hash
}

func (h *HashID) Decode(hash string) int {
	ids, err := h.hashID.DecodeWithError(hash)
	if err != nil {
		return -1
	}
	return ids[0]
}

var GameHashID = NewHashID("game")
