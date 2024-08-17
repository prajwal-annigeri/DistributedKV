package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/prajwal-annigeri/kv-store/config"
	"github.com/prajwal-annigeri/kv-store/db"
	"github.com/prajwal-annigeri/kv-store/web"
)

var (
	dbLocation = flag.String("db-location", "", "Location for Bolt DB")
	webPort    = flag.String("web-port", ":8080", "Web Port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatalf("db-location is required")
	}
}

func main() {
	parseFlags()

	var c config.Config
	if _, err := toml.DecodeFile(*configFile, &c); err != nil {
		log.Fatalf("toml.DecodeFile(%q): %v", *configFile, err)
	}

	log.Printf("%#v", &c)

	db, DBCloseFunc, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("NewDatabse(%q): %v", *dbLocation, err)
	}
	defer DBCloseFunc()

	srv := web.NewServer(db)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)

	log.Fatal(http.ListenAndServe(*webPort, nil))
}
