package challengeresponse

import (
	"context"
	"log"
	"tcppow/internal/netprotocol"
	netprotocolsrv "tcppow/internal/netprotocol/server"
	"tcppow/internal/pow/functions"

	"github.com/pkg/errors"
)

// in order to avoid context key collisions
type connRequestAccessEnabledKeyType string

const (
	connRequestAccessEnabledKey connRequestAccessEnabledKeyType = "access_enabled"
)

type ChallengeResponseHandler struct {
	next  netprotocolsrv.Handler
	powFn functions.ServerFunction
}

func New(next netprotocolsrv.Handler,
	powFn functions.ServerFunction) *ChallengeResponseHandler {
	return &ChallengeResponseHandler{
		next:  next,
		powFn: powFn,
	}
}

func (h *ChallengeResponseHandler) requestChallenge(resp *netprotocol.Response,
	req *netprotocol.Request) error {

	//log.Printf("request-challenge %+v\n", req)
	log.Println("request-challenge")

	return h.powFn.RequestChallenge(resp, req)
}

func (h *ChallengeResponseHandler) verifyChallengeResponse(resp *netprotocol.Response,
	req *netprotocol.Request) error {

	//log.Printf("verify-challenge %+v\n", req)
	log.Println("verify-challenge")

	err := h.powFn.VerifyChallengeResponse(resp, req)
	if err != nil {
		return errors.WithMessage(err, "verifying challenge response")
	}

	log.Println("client challenge's response verified successfully")

	// if PoW function verifies challenge response successfully
	// the access enabled flag should be added to the connection's context.
	// After that, the client is able to request a wisdom from the server.
	req.ConnContext = context.WithValue(req.ConnContext, connRequestAccessEnabledKey, true)

	return nil
}

func (h *ChallengeResponseHandler) ServeTCPPOW(resp *netprotocol.Response, req *netprotocol.Request) error {
	// if the user has an access to our service,
	// then we should handle next handler (most likely it will be the service's actual muxer)
	if v := req.ConnContext.Value(connRequestAccessEnabledKey); v != nil {
		accessEnabled, ok := v.(bool)
		if !ok {
			return errors.New("accessEnabled type assertion failed")
		}
		if accessEnabled {
			return h.next.ServeTCPPOW(resp, req)
		}
	}

	// otherwise, we should handle methods of the challenge-response protocol
	switch req.Method {
	case "request-challenge":
		return h.requestChallenge(resp, req)
	case "verify-challenge":
		return h.verifyChallengeResponse(resp, req)
	default:
		resp.Status = netprotocol.METHOD_NOT_ALLOWED
		return errors.New("invalid method")
	}
}
