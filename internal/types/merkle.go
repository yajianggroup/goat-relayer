package types

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"slices"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

var sha256Pool = &sync.Pool{
	New: func() any {
		return sha256.New()
	},
}

func DoubleSHA256Sum(data []byte) []byte {
	h := sha256Pool.Get().(hash.Hash)
	defer sha256Pool.Put(h)

	h.Reset()
	_, _ = h.Write(data)

	buf := make([]byte, 0, 32)
	first := h.Sum(buf)

	h.Reset()
	_, _ = h.Write(first)
	return h.Sum(buf)
}

func ComputeParentNode(left, right *chainhash.Hash) *chainhash.Hash {
	combined := slices.Concat(left[:], right[:])

	parent := new(chainhash.Hash)
	copy(parent[:], DoubleSHA256Sum(combined))
	return parent
}

func cloneChainHash(h *chainhash.Hash) *chainhash.Hash {
	res := new(chainhash.Hash)
	copy(res[:], h[:])
	return res
}

func ComputeMerkleRoot(txhs []*chainhash.Hash) *chainhash.Hash {
	if len(txhs) == 0 {
		return nil
	}

	if len(txhs) == 1 {
		return cloneChainHash(txhs[0])
	}

	if len(txhs)%2 != 0 {
		padding := new(chainhash.Hash)
		_ = padding.SetBytes(txhs[len(txhs)-1][:])
		txhs = append(txhs, padding)
	}

	parents := make([]*chainhash.Hash, 0, len(txhs)/2)
	for i := 0; i < len(txhs); i += 2 {
		parents = append(parents, ComputeParentNode(txhs[i], txhs[i+1]))
	}

	return ComputeMerkleRoot(parents)
}

func ComputeMerkleRootAndProof(txhs []*chainhash.Hash, txIndex int, proof *[]*chainhash.Hash) *chainhash.Hash {
	if len(txhs) == 0 {
		return nil
	}

	if len(txhs) == 1 {
		return cloneChainHash(txhs[0])
	}

	if len(txhs)%2 != 0 {
		txhs = append(txhs, cloneChainHash(txhs[len(txhs)-1]))
	}

	var newIndex int
	parents := make([]*chainhash.Hash, 0, len(txhs)/2)
	for i := 0; i < len(txhs); i += 2 {
		parents = append(parents, ComputeParentNode(txhs[i], txhs[i+1]))

		if i == txIndex || i+1 == txIndex {
			if i == txIndex {
				*proof = append(*proof, cloneChainHash(txhs[i+1]))
			} else {
				*proof = append(*proof, cloneChainHash(txhs[i]))
			}
			newIndex = len(parents) - 1
		}
	}

	return ComputeMerkleRootAndProof(parents, newIndex, proof)
}

func VerifyProof(txid, root *chainhash.Hash, txIndex int, path []*chainhash.Hash) bool {
	if txid != nil && txid.IsEqual(root) && txIndex == 0 && len(path) == 0 {
		return true
	}

	current := cloneChainHash(txid)
	for i := 0; i < len(path); i++ {
		if txIndex&1 == 0 {
			current = ComputeParentNode(current, path[i])
		} else {
			current = ComputeParentNode(path[i], current)
		}
		txIndex >>= 1
	}

	return current.IsEqual(root)
}

// proof = txid || intermediateNodes || merkleRoot
func VerifyRawProof(proof []byte, index int) bool {
	if len(proof)%32 != 0 {
		return false
	}

	if len(proof) == 64 {
		return bytes.Equal(proof[:32], proof[32:])
	}

	current := proof[:32]
	for i := 1; i < (len(proof)/32)-1; i++ {
		start := i * 32
		end := start + 32
		next := proof[start:end]
		if index&1 == 0 {
			current = DoubleSHA256Sum(slices.Concat(current[:], next))
		} else {
			current = DoubleSHA256Sum(slices.Concat(next, current[:]))
		}
		index >>= 1
	}

	return bytes.Equal(current, proof[len(proof)-32:])
}
