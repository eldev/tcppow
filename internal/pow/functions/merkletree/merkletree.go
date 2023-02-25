// The package contains server and client wrappers of pkg/merkletree.
// The client wrapper builds full Merkle tree for data blocks received from the server
// and sends this tree to the server serialized by the standard gob package.
// On other hand, the server wrapper deserialize this tree
// and verify that one of generated data block is actually contained in the tree.
// So the client does more work, its time complexity grows linearly
// with the data blocks count. And the server calculates hashes only from one leaf
// to the tree's root, i.e. its time complexity is O(log N), where N is the tree size.
package mekrletree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"math/rand"
	"tcppow/internal/netprotocol"
	"tcppow/pkg/merkletree"

	"github.com/pkg/errors"
)

func init() {
	// you can set seed, just uncomment

	//rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// in order to avoid context key collisions
type connRequestChallengeKeyType string

const (
	connRequestChallengeKey connRequestChallengeKeyType = "challenge"
)

// TreeContent implements the Content interface provided by pkg/merkletree.
type TreeContent struct {
	X string `json:"data"`
}

// CalculateHash calculates a hash value of a TreeContent
func (t TreeContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.X)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Equals checks for equality of two Contents
func (t TreeContent) Equals(other merkletree.LeafContent) (bool, error) {
	otherTC, ok := other.(TreeContent)
	if !ok {
		return false, errors.New("value is not of type TreeContent")
	}
	return t.X == otherTC.X, nil
}

// MTFunctionWrapper is a wrapper of pkg/merkletree implementation.
// It generates a challenge for clients, store some challenge's info
// in the connection's context in order to verify the client's response later.
type MTServerWrapper struct {
	dataBlockCount int
	dataBlockSize  int
	treeHashAlgo   string
}

func NewMTServerWrapper(dataBlockCount int, dataBlockSize int, treeHashAlgo string) *MTServerWrapper {
	return &MTServerWrapper{
		dataBlockCount: dataBlockCount,
		dataBlockSize:  dataBlockSize,
		treeHashAlgo:   treeHashAlgo,
	}
}

func (f *MTServerWrapper) RequestChallenge(resp *netprotocol.Response,
	req *netprotocol.Request) error {

	// generate a challenge for the current client
	// i.e. random data blocks for Merkle tree
	data := make([]TreeContent, f.dataBlockCount)
	for i := 0; i < len(data); i++ {
		data[i] = TreeContent{
			X: randString(f.dataBlockSize),
		}
	}

	// challenge data might be stored in the connection context
	req.ConnContext = context.WithValue(req.ConnContext, connRequestChallengeKey, data)

	// prepare challenge response
	challenge := mtChallenge{
		Data: data,
	}
	challengeJsonBytes, err := json.Marshal(challenge)
	if err != nil {
		return errors.WithMessage(err, "json marshaling of merkle tree")
	}

	resp.Status = netprotocol.OK
	resp.Body = challengeJsonBytes
	return nil
}

func (f *MTServerWrapper) VerifyChallengeResponse(resp *netprotocol.Response,
	req *netprotocol.Request) error {
	gob.Register(TreeContent{})

	// Choose randomly a data block among data blocks
	// stored in the connection context previously.
	// This data block will participate in the verification.
	var contentToVerify TreeContent
	if v := req.ConnContext.Value(connRequestChallengeKey); v != nil {
		challengdeData, ok := v.([]TreeContent)
		if !ok {
			return errors.New("challenge data type assertion failed")
		}
		randIdxToVerify := rand.Int() % len(challengdeData)
		contentToVerify = challengdeData[randIdxToVerify]
	}

	var clientResp mtResponse
	err := json.Unmarshal(req.Body, &clientResp)
	if err != nil {
		return errors.WithMessage(err, "unmarshalling client response")
	}

	// deserialize a full Merkle tree built by client
	r := bytes.NewReader(clientResp.MTreeBytes)
	dec := gob.NewDecoder(r)

	newTree, err := merkletree.NewEmptyTree(merkletree.HashStrategyType(f.treeHashAlgo))
	if err != nil {
		return errors.WithMessage(err, "creating new empty merkle tree")
	}
	err = dec.Decode(&newTree)
	if err != nil {
		return errors.WithMessage(err, "gob decoding client response tree")
	}

	// after tree deserialization, verify the random data block
	// If the tree contains this random data block, then the server
	// can be sure that the client solves the challenge correctly.
	found, err := newTree.VerifyContent(contentToVerify)
	if err != nil {
		return errors.WithMessage(err, "merkle tree verifying")
	}
	if !found {
		return errors.New("tree content not found in the tree")
	}

	resp.Status = netprotocol.OK
	resp.Body = nil
	return nil
}

type MTClientWrapper struct {
	treeHashAlgo string
}

func NewMTClientWrapper(treeHashAlgo string) *MTClientWrapper {
	return &MTClientWrapper{
		treeHashAlgo: treeHashAlgo,
	}
}

type mtChallenge struct {
	Data []TreeContent `json:"data"`
}

type mtResponse struct {
	MTreeBytes []byte `json:"tree"`
}

func (f *MTClientWrapper) SolveChallenge(rawChallengePayload []byte) ([]byte, error) {
	gob.Register(TreeContent{})

	var mtChallengePayload mtChallenge
	err := json.Unmarshal(rawChallengePayload, &mtChallengePayload)
	if err != nil {
		return nil, err
	}
	var data []merkletree.LeafContent
	for _, cont := range mtChallengePayload.Data {
		data = append(data, cont)
	}

	// build Merkle tree
	tree, err := merkletree.New(data, merkletree.HashStrategyType(f.treeHashAlgo))
	if err != nil {
		return nil, errors.WithMessage(err, "building merkle tree")
	}

	var network bytes.Buffer
	enc := gob.NewEncoder(&network)

	// serialize the tree by gob package
	err = enc.Encode(tree)
	if err != nil {
		return nil, errors.WithMessage(err, "gob encoding")
	}
	gobBytesTree := network.Bytes()

	// prepare response
	mtResp := mtResponse{
		MTreeBytes: gobBytesTree,
	}
	mtRespJsonBytes, err := json.Marshal(mtResp)
	if err != nil {
		return nil, errors.WithMessage(err, "json marshaling of merkle tree")
	}

	return mtRespJsonBytes, nil
}
