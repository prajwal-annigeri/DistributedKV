package config

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Shard struct {
	Name    string
	Idx     int
	Address string
}

type Config struct {
	Shards []Shard
}

func ParseFile(configFile string) (Config, error) {
	var c Config
	if _, err := toml.DecodeFile(configFile, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}
type Shards struct {
	Count int
	CurIdx int
	Addrs map[int]string
}

func ParseShards(shards []Shard, curShardName string) (*Shards, error) {
	shardCount := len(shards)
	shardIdx := -1
	addrs := make(map[int]string)

	for _, s := range shards {
		if _, ok := addrs[s.Idx]; ok {
			return nil, fmt.Errorf("duplicate shard index: %d", s.Idx)
		}

		addrs[s.Idx] = s.Address
		if strings.EqualFold(s.Name, curShardName) {
			shardIdx = s.Idx
		}
	}

	for i := 0; i < shardCount; i++ {
		if _, ok := addrs[i]; !ok {
			return nil, fmt.Errorf("shard index %d is not found", i)
		}
	}

	if shardIdx < 0 {
		return nil, fmt.Errorf("shard %s not found", curShardName)
	}

	return &Shards{
		Count: shardCount,
		CurIdx: shardIdx,
		Addrs: addrs,
	}, nil
}

func (s *Shards) Index(key string) int {
	h := fnv.New64()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(s.Count))
}