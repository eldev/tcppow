package server

import "tcppow/internal/netprotocol"

type Handler interface {
	ServeTCPPOW(*netprotocol.Response, *netprotocol.Request) error
}

type HandlerFunc func(*netprotocol.Response, *netprotocol.Request) error

func (f HandlerFunc) ServeTCPPOW(resp *netprotocol.Response, req *netprotocol.Request) error {
	return f(resp, req)
}

type Mux struct {
	mux map[string]Handler
}

func NewMux() *Mux {
	return &Mux{
		mux: make(map[string]Handler),
	}
}

func (m *Mux) HandleFunc(pattern string, handler func(*netprotocol.Response, *netprotocol.Request) error) {
	m.mux[pattern] = HandlerFunc(handler)
}

func (m *Mux) ServeTCPPOW(resp *netprotocol.Response, req *netprotocol.Request) error {
	handler, ok := m.mux[req.Method]
	if !ok {
		resp.Status = netprotocol.METHOD_NOT_ALLOWED
		return nil
	}

	return handler.ServeTCPPOW(resp, req)
}
