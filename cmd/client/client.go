package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"tcppow/internal/netprotocol"
	"tcppow/internal/pow/functions"
	merkletreeFn "tcppow/internal/pow/functions/merkletree"
	"time"

	"tcppow/internal/config"

	"github.com/pkg/errors"
)

const (
	envClientCfg = "CLIENT_CONFIG_PATH"
)

type merkleTreeCfg struct {
	HashAlgo string `yaml:"hash_algo"`
}

type clientCfg struct {
	ServerAddress string        `yaml:"server_address"`
	MerkleTreeCfg merkleTreeCfg `yaml:"merkletree"`
}

func initConfig() (*clientCfg, error) {
	cfg, err := config.InitYamlCfg[clientCfg](envClientCfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

type getWisdomResponse struct {
	WordOfWisdom string `json:"word_of_wisdom"`
}

func checkResponseStatusCode(reader *bufio.Reader) error {
	var status int

	statusStr, err := reader.ReadString('\n')
	if err != nil {
		return errors.WithMessage(err, "reading status from challenge response body")
	}

	_, err = fmt.Sscanf(statusStr, "Status %d\n", &status)
	if err != nil {
		return errors.WithMessage(err, "scanning raw string status")
	}
	if netprotocol.StatusCode(status) != netprotocol.OK {
		return errors.WithMessage(err, "verifying challenge verification status")
	}

	return nil
}

func getWisdom(serverAddr string, powFn functions.ClientFunction) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return errors.WithMessage(err, "tcp dialing")
	}
	defer conn.Close()

	log.Println("connected to", serverAddr)

	reader := bufio.NewReader(conn)

	// 1. request challenge
	requestChallengePayload := []byte("request-challenge\n\n") // with empty body => double '\n'
	_, err = conn.Write(requestChallengePayload)
	if err != nil {
		return errors.WithMessage(err, "requesting challenge")
	}

	// 2. solve the challenge
	if err = checkResponseStatusCode(reader); err != nil {
		return err
	}

	challengeBody, err := reader.ReadString('\n')
	if err != nil {
		return errors.WithMessage(err, "reading challenge from response body")
	}

	challengeRawResp, err := powFn.SolveChallenge([]byte(challengeBody))
	if err != nil {
		return errors.WithMessage(err, "solving challenge")
	}

	// 3. send the challenge response to the server
	var verifyChallengePayload []byte
	verifyChallengePayload = append(verifyChallengePayload,
		[]byte("verify-challenge\n")...) // netprotocol method
	verifyChallengePayload = append(verifyChallengePayload, challengeRawResp...) // netprotocol body
	verifyChallengePayload = append(verifyChallengePayload, '\n')                // netprotocol needs '\n' after body

	_, err = conn.Write(verifyChallengePayload)
	if err != nil {
		return errors.WithMessage(err, "writing challenge")
	}

	// 4. check server's verification status
	if err = checkResponseStatusCode(reader); err != nil {
		return err
	}
	_, err = reader.ReadString('\n')
	if err != nil {
		return errors.WithMessage(err, "reading body of verify-challenge response")
	}

	// 5. actual request to the server main logic
	_, err = conn.Write([]byte("get-wisdom\n\n")) // with empty body => double '\n'
	if err != nil {
		return errors.WithMessage(err, "writing get-wisdom method")
	}

	// 6. check response status and print a wisdom
	if err = checkResponseStatusCode(reader); err != nil {
		return err
	}

	rawGetWisdomResp, err := reader.ReadString('\n')
	if err != nil {
		return errors.WithMessage(err, "reading status of get-wisdom request")
	}

	var getWisdomResp getWisdomResponse
	err = json.Unmarshal([]byte(rawGetWisdomResp), &getWisdomResp)
	if err != nil {
		return errors.WithMessage(err, "unmarshaling get-wisdom response")
	}

	log.Println(getWisdomResp.WordOfWisdom)
	return nil
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}

	merkleTreeClientWrapper := merkletreeFn.NewMTClientWrapper(cfg.MerkleTreeCfg.HashAlgo)

	// the client connects to the server every 5 seconds
	// and requests wisdom quote
	ticksCh := time.Tick(5 * time.Second)
	for {
		<-ticksCh

		err = getWisdom(cfg.ServerAddress, merkleTreeClientWrapper)
		if err != nil {
			log.Fatal(err)
		}
	}
}
