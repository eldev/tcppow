package merkletree

import (
	"crypto/sha256"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type TestContent struct {
	x string
}

func (t TestContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (t TestContent) Equals(other LeafContent) (bool, error) {
	otherTC, ok := other.(TestContent)
	if !ok {
		return false, errors.New("value is not of type TestContent")
	}
	return t.x == otherTC.x, nil
}

func TestFindK(t *testing.T) {
	testCases := []struct {
		n    int
		want int
	}{
		{2, 1}, {1, 0}, {0, 0},
		{4, 2}, {5, 3}, {6, 3},
		{4, 2}, {5, 3}, {6, 3},
		{7, 3}, {8, 3}, {9, 4},
		{13, 4}, {15, 4}, {16, 4},
		{17, 5}, {31, 5},
		{127, 7}, {128, 7},
	}
	for idx, test := range testCases {
		k := findK(test.n)
		require.Equalf(t, test.want, k, "test case %d failed", idx)
	}
}

func TestBuildTreeFailed(t *testing.T) {
	data := []LeafContent{
		TestContent{x: "qwer"},
		TestContent{x: "asdfghjkl"},
		TestContent{x: "zxcvb"},
		TestContent{x: "123456"},
	}
	var niltree *MerkleTree = nil

	tree, err := New(data, HashStrategyType("invalidHashStrategy"))
	require.EqualError(t, err, "invalid hash strategy")
	require.Equal(t, niltree, tree)
}

func TestVerifyContentFailed(t *testing.T) {
	data := []LeafContent{
		TestContent{x: "qwer"},
		TestContent{x: "asdfghjkl"},
	}
	var niltree *MerkleTree = nil

	tree, err := New(data, SHA_256)
	require.EqualError(t, err, "count of leafs should be >= 4 for reliability")
	require.Equal(t, niltree, tree)
}

func TestVerifyContentSuccessfully(t *testing.T) {
	testCases := []struct {
		data          []LeafContent
		strategy      HashStrategyType
		verifyContent TestContent
		expected      bool
	}{
		{
			data: []LeafContent{
				TestContent{x: "qwer"},
				TestContent{x: "asdfghjkl"},
				TestContent{x: "zxcvb"},
				TestContent{x: "123456"},
			},
			strategy:      SHA_256,
			verifyContent: TestContent{x: "zxcvb"},
			expected:      true,
		},
		{
			data: []LeafContent{
				TestContent{x: "AMRdmug$xFH"},
				TestContent{x: "7zM1KGxpp.&bbtrgo3vfav"},
				TestContent{x: "jU0GaY8PG6"},
				TestContent{x: "GqLbSL9QOHvre34"},
				TestContent{x: "X0QKylDFg"},
				TestContent{x: "mp84kbXxbL3v!fD3"},
				TestContent{x: "l2sqq*3?"},
			},
			strategy:      SHA_256,
			verifyContent: TestContent{x: "zxcvb"},
			expected:      false,
		},
		{
			data: []LeafContent{
				TestContent{x: "S1R5tNXMOg"},
				TestContent{x: "3d8egNwJVZ"},
				TestContent{x: "AwEs6Hl8jC"},
				TestContent{x: "3Y0EUd4uZT"},
				TestContent{x: "swmSmJSG0M"},
				TestContent{x: "GIZuIYaXYg"},
				TestContent{x: "jkKDMWG7Ha"},
			},
			strategy:      SHA_512,
			verifyContent: TestContent{x: "jkKDMWG7Ha"},
			expected:      true,
		},
		{
			data: []LeafContent{
				TestContent{x: "S1R5tNXMOg"},
				TestContent{x: "3d8egNwJVZ"},
				TestContent{x: "AwEs6Hl8jC"},
				TestContent{x: "3Y0EUd4uZT"},
				TestContent{x: "swmSmJSG0M"},
				TestContent{x: "GIZuIYaXYg"},
				TestContent{x: "jkKDMWG7Ha"},
				TestContent{x: "8JIoZa8J9j"},
			},
			strategy:      SHA_256,
			verifyContent: TestContent{x: "3d8egNwJVZ"},
			expected:      true,
		},
		{
			data: []LeafContent{
				TestContent{x: "S1R5tNXMOg"},
				TestContent{x: "3d8egNwJVZ"},
				TestContent{x: "AwEs6Hl8jC"},
				TestContent{x: "3Y0EUd4uZT"},
				TestContent{x: "swmSmJSG0M"},
				TestContent{x: "GIZuIYaXYg"},
				TestContent{x: "jkKDMWG7Ha"},
				TestContent{x: "8JIoZa8J9j"},
			},
			strategy:      SHA_256,
			verifyContent: TestContent{x: "3d8egNwJV"},
			expected:      false,
		},
	}

	for tcId, tc := range testCases {
		tree, err := New(tc.data, tc.strategy)
		require.NoErrorf(t, err, "test case %d failed", tcId)

		out, err := tree.VerifyContent(tc.verifyContent)
		require.NoErrorf(t, err, "test case %d failed", tcId)
		require.Equal(t, tc.expected, out)
	}
}
