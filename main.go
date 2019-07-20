package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	DefaultHTTPAddr = ":11000"
	DefaultRaftAddr = ":12000"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

var (
	httpAddr string
	raftAddr string
	joinAddr string
	nodeID   string
)

func init() {
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No raft storage directory specified\n")
		os.Exit(1)
	}

	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	os.MkdirAll(raftDir, 0700)

	s := newStore()

	s.RaftDir = raftDir
	s.RaftBind = raftAddr

	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("%v", err)
	}

	service := newService(httpAddr, *s)
	if err := service.Start(); err != nil {
		log.Fatal("failed to start HTTP service: %s", err.Error())
	}

	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("started successfully ...")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("exiting ...")
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr),
		"application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
