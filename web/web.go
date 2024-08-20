package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prajwal-annigeri/kv-store/config"
	"github.com/prajwal-annigeri/kv-store/db"
	"github.com/prajwal-annigeri/kv-store/replication"
)

type Server struct {
	db     *db.Database
	shards *config.Shards
}

func NewServer(db *db.Database, s *config.Shards) *Server {
	return &Server{
		db:     db,
		shards: s,
	}
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, targetShard int) {
	fmt.Printf("Redirecting from shard %d to shard %d\n", s.shards.CurIdx, targetShard)
	resp, err := http.Get("http://" + s.shards.Addrs[targetShard] + r.RequestURI)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error redirecting to shard")
		return
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	targetShard := s.shards.Index(key)
	value, err := s.db.GetKey(key)

	if targetShard != s.shards.CurIdx {
		s.redirect(w, r, targetShard)
		return
	}

	fmt.Fprintf(w, "Shard = %d addr = %s Current shard = %d Value = %q, error = %v", targetShard, s.shards.Addrs[targetShard], s.shards.CurIdx, value, err)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	targetShard := s.shards.Index(key)

	if targetShard != s.shards.CurIdx {
		s.redirect(w, r, targetShard)
		return
	}

	err := s.db.SetKey(key, []byte(value))
	if err != nil {
		fmt.Fprintf(w, "Error = %v, shardIdx = %d", err, targetShard)
	}
}

func (s *Server) DeleteExtraKeysHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Error: %v", s.db.DeleteExtraKeys(func(key string) bool {
		return s.shards.Index(key) != s.shards.CurIdx
	}))
}

func (s *Server) GetNextReplicaKey(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	k, v, err := s.db.GetNextReplicaKey()
	enc.Encode(&replication.NextKVPair{
		Key:   string(k),
		Value: string(v),
		Err:   err,
	})
}

func (s *Server) DeleteReplicaKey(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	key := r.Form.Get("key")
	value := r.Form.Get("value")

	err := s.db.DeleteReplicaKey([]byte(key), []byte(value))
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		fmt.Fprintf(w, "error: %v", err)
		return
	}

	fmt.Fprintf(w, "success")
}
