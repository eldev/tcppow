package functions

import "tcppow/internal/netprotocol"

type ServerFunction interface {
	RequestChallenge(resp *netprotocol.Response, req *netprotocol.Request) error
	VerifyChallengeResponse(resp *netprotocol.Response, req *netprotocol.Request) error
}

type ClientFunction interface {
	SolveChallenge(rawChallengePayload []byte) ([]byte, error)
}
