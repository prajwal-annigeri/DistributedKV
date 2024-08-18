package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prajwal-annigeri/kv-store/config"
	"github.com/prajwal-annigeri/kv-store/db"
	"github.com/prajwal-annigeri/kv-store/web"
)

var (
	dbLocation = flag.String("db-location", "", "Location for Bolt DB")
	webPort    = flag.String("web-port", "8080", "Web Port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "", "Name of the shard for the data")
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

	log.Printf("Shard count: %d, current shard: %d", shards.Count, shards.CurIdx)

	db, DBCloseFunc, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("NewDatabse(%q): %v", *dbLocation, err)
	}
	defer DBCloseFunc()
	srv := web.NewServer(db, shards)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)
	http.HandleFunc("/purge", srv.DeleteExtraKeysHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *webPort), nil))
}
