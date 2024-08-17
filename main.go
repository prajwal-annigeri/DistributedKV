package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
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

	var c config.Config
	if _, err := toml.DecodeFile(*configFile, &c); err != nil {
		log.Fatalf("toml.DecodeFile(%q): %v", *configFile, err)
	}
	log.Printf("%#v", &c)

	shardIndex := -1
	shardCount := len(c.Shards)
	addrs := make(map[int]string)

	for _, s := range c.Shards {		
		addrs[s.Idx] = s.Address

		if strings.EqualFold(s.Name, *shard) {
			shardIndex = s.Idx
		}
	}
	if shardIndex == -1 {
		log.Fatalf("Shard %q was not found", *shard)
	}

	log.Printf("Shard count: %d, current shard: %d", shardCount, shardIndex)

	db, DBCloseFunc, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("NewDatabse(%q): %v", *dbLocation, err)
	}
	defer DBCloseFunc()

	srv := web.NewServer(db, shardIndex, shardCount, addrs)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *webPort), nil))
}
