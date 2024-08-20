package replication

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/prajwal-annigeri/kv-store/db"
)

type NextKVPair struct {
	Key   string
	Value string
	Err   error
}

type client struct {
	db         *db.Database
	masterAddr string
}

func ClientLoop(db *db.Database, masterAddr string) {
	c := &client{
		db:         db,
		masterAddr: masterAddr,
	}
	for {
		present, err := c.loop()
		if err != nil {
			log.Printf("Loop error :%v\n", err)
			time.Sleep(time.Second)
			continue
		}

		if !present {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (c *client) loop() (present bool, err error) {
	resp, err := http.Get("http://" + c.masterAddr + "/next-replication-key")
	if err != nil {
		return false, err
	}

	var res NextKVPair
	if json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if res.Key == "" {
		return false, nil
	}

	if err := c.db.SetReplicaKey(res.Key, []byte(res.Value)); err != nil {
		return false, nil
	}

	if err := c.deleteFromReplicaBucket(res.Key, res.Value); err != nil {
		log.Printf("deleteFromReplicaBucket failed: %v\n", err)
	}

	return true, nil
}

func (c *client) deleteFromReplicaBucket(key, value string) error {
	u := url.Values{
		"key":   []string{key},
		"value": []string{value},
	}

	log.Printf("Deleting %q: %q from replication bucket of %q\n", key, value, c.masterAddr)

	resp, err := http.Get("http://" + c.masterAddr + "/delete-replication-key?" + u.Encode())
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !bytes.Equal(result, []byte("success")) {
		return errors.New(string(result))
	}

	return nil
}
