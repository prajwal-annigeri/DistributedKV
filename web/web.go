package web

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"

	"github.com/prajwal-annigeri/kv-store/config"
	"github.com/prajwal-annigeri/kv-store/db"
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

func (s *Server) getShard(key string) int {
	h := fnv.New64()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(s.shards.Count))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, targetShard int) {
	fmt.Printf("Redirecting from shard %d to shard %d", s.shards.CurIdx, targetShard)
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
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get("value")

	targetShard := s.getShard(key)

	if targetShard != s.shards.CurIdx {
		s.redirect(w, r, targetShard)
		return
	}

	err := s.db.SetKey(key, []byte(value))
	fmt.Fprintf(w, "Error = %v, shardIdx = %d", err, targetShard)
}
