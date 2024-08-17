package web

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"

	"github.com/prajwal-annigeri/kv-store/db"
)

type Server struct {
	db         *db.Database
	shardIdx   int
	shardCount int
	addrs map[int]string
}

func NewServer(db *db.Database, shardIdx int, shardCount int, addrs map[int]string) *Server {
	return &Server{
		db:         db,
		shardIdx:   shardIdx,
		shardCount: shardCount,
		addrs: addrs,
	}
}

func (s *Server) getShard(key string) int {
	h := fnv.New64()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(s.shardCount))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, targetShard int) {
	fmt.Printf("Redirecting from shard %d to shard %d", s.shardIdx, targetShard)
	resp, err := http.Get("http://" + s.addrs[targetShard] + r.RequestURI)
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

	targetShard := s.getShard(key)
	value, err := s.db.GetKey(key)

	if targetShard != s.shardIdx {
		s.redirect(w, r, targetShard)
		return
	}

	fmt.Fprintf(w, "Shard = %d addr = %s Current shard = %d Value = %q, error = %v", targetShard, s.addrs[targetShard], s.shardIdx,value, err)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get("value")
	
	targetShard := s.getShard(key)

	if targetShard != s.shardIdx {
		s.redirect(w, r, targetShard)
		return
	}

	err := s.db.SetKey(key, []byte(value))
	fmt.Fprintf(w, "Error = %v, shardIdx = %d", err, targetShard)
}
