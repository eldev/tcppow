package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"tcppow/internal/netprotocol"
	"testing"

	"github.com/stretchr/testify/require"
)

// testHandler is an auxiliary handler for test purposes.
type testHandler struct {
	callback func(resp *netprotocol.Response, req *netprotocol.Request)
	retErr   error
}

func (th *testHandler) ServeTCPPOW(resp *netprotocol.Response, req *netprotocol.Request) error {
	th.callback(resp, req)
	return th.retErr
}

func TestServeConnectionMethodAndBodyParsing(t *testing.T) {
	testCases := []struct {
		method string
		body   []byte
		retErr error
	}{
		{
			method: "require-challenge",
			body:   []byte{0x1, 0x2, 0x3, 0x4},
			retErr: nil,
		},
		{
			method: "get-wisdom",
			body:   nil,
			retErr: nil,
		},
		{
			method: "some-method",
			body:   []byte("someMethodBody"),
			retErr: nil,
		},
	}

	for tcId, tc := range testCases {
		method := tc.method
		body := tc.body
		callback := func(resp *netprotocol.Response, req *netprotocol.Request) {
			require.Equalf(t, method, req.Method, "test case %d failed", tcId)
			require.Equal(t, body, req.Body, "test case %d failed", tcId)
		}
		handler := &testHandler{
			callback: callback,
			retErr:   tc.retErr,
		}
		r, w := net.Pipe()

		go func() {
			serveConnection(r, handler)
		}()

		toWrite := []byte(method)
		toWrite = append(toWrite, '\n')
		toWrite = append(toWrite, body...)
		w.Write(toWrite)
		w.Close()
	}
}

func TestServeConnectionServerResponse(t *testing.T) {
	testCases := []struct {
		respStatus netprotocol.StatusCode
		respBody   []byte
		retErr     error
	}{
		{
			respStatus: netprotocol.OK,
			respBody:   []byte{0x1, 0x2, 0x3},
			retErr:     nil,
		},
		{
			respStatus: netprotocol.INTERNAL_SERVER_ERROR,
			respBody:   []byte{},
			retErr:     errors.New("some-internal-error"),
		},
	}

	for tcId, tc := range testCases {
		callback := func(resp *netprotocol.Response, req *netprotocol.Request) {
			resp.Status = tc.respStatus
			resp.Body = tc.respBody
		}
		handler := &testHandler{
			callback: callback,
			retErr:   tc.retErr,
		}
		r, w := net.Pipe()

		go func() {
			serveConnection(r, handler)
		}()

		toWrite := []byte("method")
		toWrite = append(toWrite, '\n')
		toWrite = append(toWrite, []byte{0x10, 0x20, 0x30, '\n'}...)
		w.Write(toWrite)

		reader := bufio.NewReader(w)

		// check server response status
		statusStr, err := reader.ReadString('\n')
		require.NoErrorf(t, err, "test case %d failed", tcId)
		require.Equalf(t, fmt.Sprintf("Status %d\n", tc.respStatus), statusStr,
			"test case %d failed", tcId)

		if tc.retErr != nil {
			// if handler returns some error
			// then the server should close the connection
			// so here try to read another bytes
			// and check if EOF error occurs
			_, err = reader.ReadBytes('\n')
			require.ErrorIs(t, err, io.EOF)
			w.Close()
			continue
		}

		// check server response body
		body, err := reader.ReadBytes('\n')
		require.NoErrorf(t, err, "test case %d failed", tcId)
		require.Equalf(t, byte('\n'), body[len(body)-1],
			"test case %d failed", tcId) // the last byte should be \n
		require.Equalf(t, tc.respBody, body[:len(body)-1],
			"test case %d failed", tcId)

		w.Close()
	}
}
