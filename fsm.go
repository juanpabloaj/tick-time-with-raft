package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

type Store struct {
	RaftDir  string
	RaftBind string
	raft     *raft.Raft
}

func (s *Store) Open(enableSingle bool, localID string) error {

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	log.Printf("Open [%v] [%v]", config, config.LocalID)

	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore

	logStore = raft.NewInmemStore()
	stableStore = raft.NewInmemStore()

	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

func (s *Store) Join(nodeID, addr string) error {
	log.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	log.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

func newStore() *Store {
	return &Store{}
}

type fsm Store

func (f *fsm) Apply(l *raft.Log) interface{} {

	log.Printf("apply [%v] [%v]", l, l.Data)

	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {

	log.Printf("snapshot")

	return &fsmSnapshot{}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {

	log.Printf("restore [%v]", rc)

	return nil
}

type fsmSnapshot struct {
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b := []byte("hello from persist")

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
