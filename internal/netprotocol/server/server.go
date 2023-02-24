package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"tcppow/internal/netprotocol"

	"github.com/pkg/errors"
)

type Server struct {
}

func New() *Server {
	return &Server{}
}

func (s *Server) ListenAndServe(addr string, handler Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		// accept a new incoming connection
		conn, err := listener.Accept()
		if err != nil {
			return errors.WithMessage(err, "accept a new incoming connection")
		}

		// handle the connection in a separate goroutine
		go serveConnection(conn, handler)
	}
}

func writeStatus(conn net.Conn, status netprotocol.StatusCode) error {
	_, err := conn.Write([]byte(fmt.Sprintf("Status %d\n", status)))
	if err != nil {
		return err
	}
	return err
}

func serveConnection(conn net.Conn, handler Handler) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	connCtx := context.Background()

	for {
		// parse request method
		method, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection closed by client")
				return
			}
			writeStatus(conn, netprotocol.BAD_REQUEST)
			log.Println("err while reading method:", err)
			return
		}
		method = strings.TrimSpace(method)

		// parse request body
		body, err := reader.ReadBytes('\n')
		if err != nil {
			writeStatus(conn, netprotocol.BAD_REQUEST)
			log.Println("err while reading body:", err)
			return
		}
		body = bytes.TrimSpace(body)

		request := netprotocol.Request{
			Method:         method,
			Body:           body,
			ConnContext:    connCtx,
			RequestContext: context.Background(),
		}

		var response netprotocol.Response

		// serve the request
		err = handler.ServeTCPPOW(&response, &request)
		if err != nil {
			writeStatus(conn, netprotocol.INTERNAL_SERVER_ERROR)
			log.Println("err while serving request:", err)
			return
		}
		connCtx = request.ConnContext

		if err = writeStatus(conn, response.Status); err != nil {
			log.Println("err while writing status:", err)
			return
		}

		_, err = conn.Write(append(response.Body, '\n'))
		if err != nil {
			log.Println("err while writing response body:", err)
			return
		}
	}
}
