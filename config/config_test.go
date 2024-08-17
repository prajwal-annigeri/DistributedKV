package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/prajwal-annigeri/kv-store/config"
)

func createConfig(t *testing.T, content string) config.Config {
	t.Helper()

	f, err := os.CreateTemp(os.TempDir(), "config.toml")
	if err != nil {
		t.Fatalf("Couldn't create temp file: %v", err)
	}

	defer f.Close()

	tempFileName := f.Name()
	defer os.Remove(tempFileName)

	_, err = f.WriteString(content)
	if err != nil {
		t.Fatalf("Could not write the config contents")
	}

	c, err := config.ParseFile(tempFileName)
	if err != nil {
		t.Fatalf("Could not parse config file: %v", err)
	}
	return c
}

func TestConfigParse(t *testing.T) {
	res := createConfig(t, `[[shards]]
		name = "shard1"
		idx = 0
		address = "localhost:8080"`,
	)

	expected := config.Config{
		Shards: []config.Shard{
			{
				Name:    "shard1",
				Idx:     0,
				Address: "localhost:8080",
			},
		},
	}

	if !reflect.DeepEqual(expected, res) {
		t.Errorf("config mismatch. got: %#v, want: %#v", res, expected)
	}
}

func TestParseShards(t *testing.T) {
	c := createConfig(t, `
	[[shards]]
		name = "shard1"
		idx = 0
		address = "localhost:8080"
	[[shards]]
		name = "shard2"
		idx = 1
		address = "localhost:8081"`,
	)

	res, err := config.ParseShards(c.Shards, "shard2")
	if err != nil {
		t.Fatalf("Could not parse shards %#v: %v", c.Shards, err)
	}

	expected := &config.Shards{
		Count: 2,
		CurIdx: 1,
		Addrs: map[int]string{
			0: "localhost:8080",
			1: "localhost:8081",
		},
	}

	if !reflect.DeepEqual(expected, res) {
		t.Errorf("Shards config does not match: res: %#v, expected: %#v", res, expected)
	}
}
