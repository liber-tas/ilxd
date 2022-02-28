// Copyright (c) 2022 The illium developers
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/project-illium/ilxd/models"
	"github.com/project-illium/ilxd/params/hash"
)

// InclusionProof is a merkle inclusion proof which proves that
// a given element is in the set with the given accumulator root.
type InclusionProof struct {
	ID          models.ID
	Accumulator [][]byte
	Hashes      [][]byte
	Flags       uint64
	last        []byte
	index       uint64
}

// Accumulator is a hashed-based cryptographic data structure similar to a
// markle tree. Like a merkle tree, the added elements (leafs) are hashed
// together to form branches an ultimately a root hash.
//
// This accumulator, however, uses a special type of tree designed for fast
// appends. Another name for the accumulator is a Merkle Mountain Range.
// Instead of a single tree, the data structure consists of multiple trees,
// each one smaller than the previous.
//
// When a new data element is added that would make the tree unbalanced, it
// is used to start a new tree. Consider the following balanced tree:
//                            peak0
//                       /            \
//                    h12              h34
//                  /     \          /     \
//                h1       h2      h3       h4
//
// When h5 is added it will start a new tree. The follow depicts the state
// after h5 and h6 are added:
//                            peak0
//                       /            \
//                    h12              h34             peak1
//                  /     \          /     \         /      \
//                h1       h2      h3       h4      h5      h6
// Notice we have two peaks. And if h7 were added that would form a third peak.
// The "root" of the entire data structure is the hash of all the peaks,
// hash(peak0 || peak1) in this case. Whenever two peaks reach the same height,
// the two trees are merged back together into a single tree. Thus, as elements
// are added, the number of peaks initially fans out before ultimately collapsing
// back together.
//
// To add new elements to the tree and calculate the new root we only need to
// store the peaks, nothing else.
type Accumulator struct {
	acc       [][]byte
	nElements uint64
	proofs    map[models.ID]*InclusionProof
	lookupMap map[models.ID]*InclusionProof
}

// NewAccumulator returns a new Accumulator.
func NewAccumulator() *Accumulator {
	return &Accumulator{
		acc:       make([][]byte, 1),
		proofs:    make(map[models.ID]*InclusionProof),
		lookupMap: make(map[models.ID]*InclusionProof),
		nElements: 0,
	}
}

// Insert inserts a data element into the accumulator. The element is prepended
// with data index and hashed. This will change the accumulator root. It is not
// possible to go backwards and undo this operation so make sure you mean to do
// it.
//
// If you wish to keep track of an InclusionProof for this data element set
// 'protect' true. This must be done at the time of adding as it's not possible
// to go back and protect previous items after the accumulator has been mutated.
func (a *Accumulator) Insert(data []byte, protect bool) {
	a.nElements++
	n := hashWithIndex(data, a.nElements-1)

	// If one of our protected hashes is at acc[0] then it was an
	// odd number leaf and the very next leaf must be part of its
	// inclusion proof.
	proof, ok := a.proofs[models.NewID(a.acc[0])]
	if ok {
		c := make([]byte, len(n))
		copy(c, n)
		proof.Hashes = append(proof.Hashes, c)
		proof.last = hashMerkleBranches(a.acc[0], n)
		proof.Flags = 1
	}

	if protect {
		ip := &InclusionProof{
			ID:    models.NewID(data),
			index: a.nElements - 1,
		}
		a.proofs[models.NewID(n)] = ip
		a.lookupMap[models.NewID(data)] = ip
		// If acc[0] is not nil then this means the new leaf is
		// and even number and the previous leaf is part of its
		// inclusion proof.
		if a.acc[0] != nil {
			c1 := make([]byte, len(a.acc[0]))
			c2 := make([]byte, len(n))
			copy(c1, a.acc[0])
			copy(c2, n)
			ip.Hashes = append(ip.Hashes, c1)
			ip.last = hashMerkleBranches(a.acc[0], c2)
		}
	}

	h := 0
	r := a.acc[h]
	for r != nil {
		n = hashMerkleBranches(r, n)

		// Iterate over all proofs and update them before we prune
		// branches off the tree.
		for _, proof := range a.proofs {
			h2 := h + 1
			l := len(proof.Hashes)
			if l > 0 && h2 >= l && h2 <= a.len() {
				if !bytes.Equal(proof.last, n) { // Right
					c := make([]byte, len(n))
					copy(c, n)

					proof.Hashes = append(proof.Hashes, c)
					proof.last = hashMerkleBranches(proof.last, n)

					f := uint64(1) << uint64(len(proof.Hashes)-1)
					proof.Flags |= f
				} else { // Left
					if a.acc[h+1] == nil {
						continue
					}
					c := make([]byte, len(a.acc[h+1]))
					copy(c, a.acc[h+1])

					proof.Hashes = append(proof.Hashes, c)
					proof.last = hashMerkleBranches(a.acc[h+1], proof.last)
				}
			}
		}

		a.maybeResizeAndSet(h, nil)
		h++
		r = a.acc[h]
	}
	a.maybeResizeAndSet(h, n)
}

// Root returns the root hash of the accumulator. This is not cached
// and a new hash is calculated each time this method is called.
func (a *Accumulator) Root() models.ID {
	return models.NewID(catAndHash(a.acc))
}

// GetProof returns an inclusion proof, if it exists, for the provided hash.
//
// This is NOT safe for concurrent access.
func (a *Accumulator) GetProof(data []byte) (*InclusionProof, error) {
	proof, ok := a.lookupMap[models.NewID(data)]
	if !ok {
		return nil, errors.New("not found")
	}
	acc := make([][]byte, 0, len(a.acc))
	for _, peak := range a.acc {
		if peak != nil {
			peakCopy := make([]byte, len(peak))
			copy(peakCopy, peak)
			acc = append(acc, peakCopy)
		}
	}
	newProof := &InclusionProof{ID: models.NewID(data), Accumulator: acc, Flags: proof.Flags, Hashes: make([][]byte, len(proof.Hashes))}
	copy(newProof.Hashes, proof.Hashes)
	return newProof, nil
}

// DropProof ceases tracking of the inclusion proof for the given
// element and deletes all tree branches related to the proof.
//
// This is NOT safe for concurrent access.
func (a *Accumulator) DropProof(data []byte) {
	proof, ok := a.lookupMap[models.NewID(data)]
	if !ok {
		return
	}

	n := hashWithIndex(data, proof.index)

	delete(a.lookupMap, models.NewID(data))
	delete(a.proofs, models.NewID(n))
}

// The Insert method often checks the value of the accumulator element
// at index len(acc) which would cause an index out of range panic. So
// This function not only adds the data to the accumulator, but increases
// the capacity if necessary to avoid a panic.
func (a *Accumulator) maybeResizeAndSet(pos int, h []byte) {
	if pos+2 > len(a.acc) {
		a.acc = append(a.acc, nil, nil)
	}
	a.acc[pos] = h
}

func (a *Accumulator) len() int {
	l := 0
	for _, d := range a.acc {
		if d != nil {
			l++
		}
	}
	return l
}

func nElementsToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func hashWithIndex(data []byte, index uint64) []byte {
	d := make([]byte, len(data)+8)
	copy(d[:8], nElementsToBytes(index))
	copy(d[8:], data)
	return hash.HashFunc(d)
}

func catAndHash(data [][]byte) []byte {
	combined := make([]byte, 0, hash.HashSize*len(data))
	for _, peak := range data {
		peakCopy := make([]byte, len(peak))
		copy(peakCopy, peak)
		combined = append(combined, peakCopy...)
	}
	return hash.HashFunc(combined)
}
