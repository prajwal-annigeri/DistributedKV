package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prajwal-annigeri/kv-store/config"
	"github.com/prajwal-annigeri/kv-store/db"
	"github.com/prajwal-annigeri/kv-store/replication"
	"github.com/prajwal-annigeri/kv-store/web"
)

var (
	dbLocation = flag.String("db-location", "", "Location for Bolt DB")
	addr       = flag.String("addr", "127.0.0.1:8080", "Address")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "", "Name of the shard for the data")
	replica    = flag.Bool("replica", false, "Run as replica")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatalf("db-location is required")
	}

	if *shard == "" {
		log.Fatalf("shard is required")
	}
}

func main() {
	parseFlags()

	c, err := config.ParseFile(*configFile)
	if err != nil {
		log.Fatalf("ParseFile(%s): %v", *configFile, err)
	}

	shards, err := config.ParseShards(c.Shards, *shard)
	if err != nil {
		log.Fatalf("ParseShards(): %v", err)
	}

	log.Printf("Shard count: %d, current shard: %d, replica: %t\n", shards.Count, shards.CurIdx, *replica)

	db, DBCloseFunc, err := db.NewDatabase(*dbLocation, *replica)
	if err != nil {
		log.Fatalf("NewDatabase(%q): %v", *dbLocation, err)
	}
	defer DBCloseFunc()

	if *replica {
		masterAddr, ok := shards.Addrs[shards.CurIdx]
		if !ok {
			log.Fatalf("Could not find address for leader shard of %d", shards.CurIdx)
		}

		go replication.ClientLoop(db, masterAddr)
	}

	srv := web.NewServer(db, shards)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)
	http.HandleFunc("/purge", srv.DeleteExtraKeysHandler)
	http.HandleFunc("/next-replication-key", srv.GetNextReplicaKey)
	http.HandleFunc("/delete-replication-key", srv.DeleteReplicaKey)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
