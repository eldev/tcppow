package main

import (
	"log"
	"tcppow/internal/config"
	tcppowsrv "tcppow/internal/netprotocol/server"
	merkletreeFn "tcppow/internal/pow/functions/merkletree"
	"tcppow/internal/pow/protocol/challengeresponse"
	"tcppow/internal/wordofwisdom/domain"
	"tcppow/internal/wordofwisdom/handlers/getwisdom"
	"tcppow/internal/wordofwisdom/repository/inmemory"
)

const (
	envServerCfg = "SERVER_CONFIG_PATH"
)

type merkleTreeCfg struct {
	DataBlockCount int    `yaml:"data_block_count"`
	DataBlockSize  int    `yaml:"data_block_size"`
	HashAlgo       string `yaml:"hash_algo"`
}

type serverCfg struct {
	Address       string        `yaml:"address"`
	MerkleTreeCfg merkleTreeCfg `yaml:"merkletree"`
}

func initConfig() (*serverCfg, error) {
	cfg, err := config.InitYamlCfg[serverCfg](envServerCfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}

	// init main logic's objects of the service
	mux := tcppowsrv.NewMux()
	wisdomRepo := inmemory.New()
	model := domain.New(wisdomRepo)
	getWisdomHandler := getwisdom.New(model)
	mux.HandleFunc("get-wisdom", getWisdomHandler.HandlerFunc())

	// init challenge-response PoW wiht Merkle tree function.
	// This might be considered as a kind of middleware.
	// After challenge-response verification the client has an access
	// to the mux object of the main logic.
	merkletreeFunc := merkletreeFn.NewMTServerWrapper(cfg.MerkleTreeCfg.DataBlockCount,
		cfg.MerkleTreeCfg.DataBlockSize, cfg.MerkleTreeCfg.HashAlgo)

	challengeresponseMw := challengeresponse.New(mux, merkletreeFunc)

	// init TCP server
	server := tcppowsrv.New()
	if err := server.ListenAndServe(cfg.Address, challengeresponseMw); err != nil {
		log.Fatal(err)
	}
}
