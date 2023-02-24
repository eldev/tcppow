// This package contains an implementation of Merkle Tree.
// It is inspired by wonderful project https://github.com/cbergoon/merkletree.
// But this project can not be serialized/deserialized by the standard gob package
// (because gob doesn't work fine with recursive data structure).
//
// So the current implementation is simpler,
// because it uses an ideal binary tree structure and represents it as array
// in order to get rid of recursive tree data structure.
// (many heap implementations uses a such approach).
// So it can be serialized/deserialized easily by standard gob package.

package merkletree

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/pkg/errors"
)

type LeafContent interface {
	CalculateHash() ([]byte, error)
	Equals(other LeafContent) (bool, error)
}

type HashStrategyType string

const (
	SHA_256 HashStrategyType = "sha256"
	SHA_512 HashStrategyType = "sha512"
)

// Simple implementation
type MerkleTree struct {
	K                int // number of leafs would be 2^K, i.e. it is an ideal binary tree.
	Leafs            []LeafContent
	Nodes            []Node
	HashStrategyName HashStrategyType
	HashStrategy     func() hash.Hash
}

type Node struct {
	Hash []byte
}

func concatenateHashes(h1, h2 []byte) []byte {
	// here might be more specific logic (with sorting, mixing and etc.)
	return append(h1, h2...)
}

func (t *MerkleTree) calculateLeafNode(leafIdx int) ([]byte, error) {
	leaf := t.Leafs[leafIdx]

	hash, err := leaf.CalculateHash()
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *MerkleTree) calculateNonLeafNode(leftNodeIdx, rightNodeIdx int) ([]byte, error) {
	leftHash := t.Nodes[leftNodeIdx].Hash
	rightHash := t.Nodes[rightNodeIdx].Hash
	hasher := t.HashStrategy()
	concatenatedHash := concatenateHashes(leftHash, rightHash)
	_, err := hasher.Write(concatenatedHash)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func (t *MerkleTree) buildTree(curNodeIdx int) error {
	var err error

	// the last layer (leafs layer) of the ideal binary tree
	// always has index range from (N-1), ... 2*N-2, where N is a number of leafs.
	N := len(t.Leafs)

	if curNodeIdx >= N-1 {
		// the current node is a leaf
		nodeIdxInLeafs := curNodeIdx - (N - 1)
		t.Nodes[curNodeIdx].Hash, err = t.calculateLeafNode(nodeIdxInLeafs)
		if err != nil {
			return err
		}
		return nil
	}

	// if not a leaf

	// use these known formulas for left and right children indexes
	// for ideal binary tree (root index equals to 0).
	leftNodeIdx := 2*curNodeIdx + 1
	rightNodeIdx := 2*curNodeIdx + 2

	t.buildTree(leftNodeIdx)
	t.buildTree(rightNodeIdx)

	t.Nodes[curNodeIdx].Hash, err = t.calculateNonLeafNode(leftNodeIdx, rightNodeIdx)
	if err != nil {
		return err
	}

	return nil
}

func findK(leafNr int) int {
	K := 0
	val := 1
	for (val << K) < leafNr {
		K++
	}
	return K
}

func New(leafs []LeafContent, hashStrategyName HashStrategyType) (*MerkleTree, error) {
	K := findK(len(leafs))

	if K < 2 {
		return nil, errors.New("count of leafs should be >= 4 for reliability")
	}

	leafsNumber := 1 << K

	// total number of nodes can be calculated as a sum of a geometric progression,
	// because we have an ideal binary tree.
	// a1 -- start value of the progression would be 1 (because the root layer has only one node)
	// progressionRatio, of course, would be 2.
	a1 := 1
	progressionRatio := 2
	// NOTE: in fact, the formula for a geometric progression sum contains exponentiation,
	// but for progressionRatio=2 exponentiation is equivalent to a shift by K bits.
	totalNumberOfNodes := a1 * ((progressionRatio << K) - 1) / (progressionRatio - 1)

	tree, err := NewEmptyTree(hashStrategyName)
	if err != nil {
		return nil, err
	}

	tree.K = K
	tree.Leafs = make([]LeafContent, 0, leafsNumber)
	tree.Nodes = make([]Node, totalNumberOfNodes)

	// fill leafs content
	for i := 0; i < leafsNumber; i++ {
		if i < len(leafs) {
			tree.Leafs = append(tree.Leafs, leafs[i])
		} else {
			// fill by last element's duplicates to get ideal binary tree structure
			duplicate := leafs[len(leafs)-1]
			tree.Leafs = append(tree.Leafs, duplicate)
		}
	}

	// calculate all hashes
	err = tree.buildTree(0) // start from root node
	if err != nil {
		return nil, errors.WithMessage(err, "building tree")
	}

	return tree, nil
}

func NewEmptyTree(hashStrategyName HashStrategyType) (*MerkleTree, error) {
	var hashStrategy func() hash.Hash
	switch hashStrategyName {
	case SHA_256:
		hashStrategy = sha256.New
	case SHA_512:
		hashStrategy = sha512.New
	default:
		return nil, errors.New("invalid hash strategy")
	}

	tree := &MerkleTree{
		HashStrategyName: hashStrategyName,
		HashStrategy:     hashStrategy,
	}
	return tree, nil
}

func (t *MerkleTree) VerifyContent(content LeafContent) (bool, error) {
	// the function should find the content among the tree's leafs
	// and verify hashes from this leaf to root
	for leafIdx, leaf := range t.Leafs {
		ok, err := leaf.Equals(content)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		// the last layer (leafs layer) of the ideal binary tree
		// always has index range from (N-1), ... 2*N-2, where N is a number of leafs.
		N := len(t.Leafs)
		nodeIdx := leafIdx + N - 1

		leafHash, err := t.calculateLeafNode(leafIdx)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(leafHash, t.Nodes[nodeIdx].Hash) {
			return false, nil
		}

		// verify all hashes from the found leaf to root of the tree
		curNodeIdx := (nodeIdx - 1) / 2 // the index of parent node
		for curNodeIdx >= 0 {
			leftNodeIdx := 2*curNodeIdx + 1
			rightNodeIdx := 2*curNodeIdx + 2
			nodeHash, err := t.calculateNonLeafNode(leftNodeIdx, rightNodeIdx)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(nodeHash, t.Nodes[curNodeIdx].Hash) {
				return false, nil
			}

			if curNodeIdx == 0 {
				break
			}
			curNodeIdx = (curNodeIdx - 1) / 2
		}
		return true, nil

	}
	return false, nil
}
