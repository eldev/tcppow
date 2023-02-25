package merkletree

import (
	"context"
	"tcppow/internal/netprotocol"
	"tcppow/pkg/merkletree"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChallengeRequestAndVerification(t *testing.T) {
	testCases := []struct {
		dataBlockCount int
		dataBlockSize  int
		srvStrategy    merkletree.HashStrategyType
		clientStrategy merkletree.HashStrategyType
	}{
		{
			dataBlockCount: 8,
			dataBlockSize:  12,
			srvStrategy:    merkletree.SHA_256,
			clientStrategy: merkletree.SHA_256,
		},
		{
			dataBlockCount: 32,
			dataBlockSize:  16,
			srvStrategy:    merkletree.SHA_512,
			clientStrategy: merkletree.SHA_512,
		},
		{
			dataBlockCount: 128,
			dataBlockSize:  64,
			srvStrategy:    merkletree.SHA_512,
			clientStrategy: merkletree.SHA_512,
		},
		{
			dataBlockCount: 2048,
			dataBlockSize:  256,
			srvStrategy:    merkletree.SHA_512,
			clientStrategy: merkletree.SHA_512,
		},
	}

	for _, tc := range testCases {
		srvWrapper := NewMTServerWrapper(tc.dataBlockCount,
			tc.dataBlockSize, string(tc.srvStrategy))
		clientWrapper := NewMTClientWrapper(string(tc.clientStrategy))

		var srvResponse netprotocol.Response
		srvRequest := netprotocol.Request{
			ConnContext:    context.Background(),
			RequestContext: context.Background(),
		}

		err := srvWrapper.RequestChallenge(&srvResponse, &srvRequest)
		require.NoError(t, err)
		require.Equal(t, netprotocol.OK, srvResponse.Status)
		require.NotNil(t, srvResponse.Body)

		treePayload, err := clientWrapper.SolveChallenge(srvResponse.Body)
		require.NoError(t, err)

		srvRequest.Body = treePayload
		err = srvWrapper.VerifyChallengeResponse(&srvResponse, &srvRequest)
		require.NoError(t, err)
	}
}

func TestHashStrategyNonCompliance(t *testing.T) {
	testCases := []struct {
		dataBlockCount int
		dataBlockSize  int
		srvStrategy    merkletree.HashStrategyType
		clientStrategy merkletree.HashStrategyType
	}{
		{
			dataBlockCount: 32,
			dataBlockSize:  20,
			srvStrategy:    merkletree.SHA_256,
			clientStrategy: merkletree.SHA_512,
		},
		{
			dataBlockCount: 8,
			dataBlockSize:  16,
			srvStrategy:    merkletree.SHA_512,
			clientStrategy: merkletree.SHA_256,
		},
	}

	for _, tc := range testCases {
		srvWrapper := NewMTServerWrapper(tc.dataBlockCount,
			tc.dataBlockSize, string(tc.srvStrategy))
		clientWrapper := NewMTClientWrapper(string(tc.clientStrategy))

		var srvResponse netprotocol.Response
		srvRequest := netprotocol.Request{
			ConnContext:    context.Background(),
			RequestContext: context.Background(),
		}

		err := srvWrapper.RequestChallenge(&srvResponse, &srvRequest)
		require.NoError(t, err)
		require.Equal(t, netprotocol.OK, srvResponse.Status)
		require.NotNil(t, srvResponse.Body)

		treePayload, err := clientWrapper.SolveChallenge(srvResponse.Body)
		require.NoError(t, err)

		srvRequest.Body = treePayload
		err = srvWrapper.VerifyChallengeResponse(&srvResponse, &srvRequest)
		require.ErrorContains(t, err, "tree content not found in the tree")
	}
}
