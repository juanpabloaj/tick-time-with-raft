package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct {
	addr  string
	store Store
}

func newService(addr string, store Store) *Service {
	return &Service{
		addr:  addr,
		store: store,
	}
}

func (s *Service) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/join", s.handleJoin)

	go func() {
		log.Printf("listening on %s ...", s.addr)
		http.ListenAndServe(s.addr, router)
	}()
	return nil
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
