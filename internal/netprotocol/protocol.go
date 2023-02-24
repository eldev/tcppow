package netprotocol

import "context"

type Request struct {
	Method         string
	Body           []byte
	ConnContext    context.Context
	RequestContext context.Context
}

type StatusCode int

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	METHOD_NOT_ALLOWED    StatusCode = 405
	INTERNAL_SERVER_ERROR StatusCode = 500
)

type Response struct {
	Status StatusCode
	Body   []byte
}
