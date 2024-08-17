package db_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/prajwal-annigeri/kv-store/db"
)

func TestGetSet(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "dbtest")
	if err != nil {
		t.Fatalf("Could not creater temp file: %v", err)
	}
	name := f.Name()
	f.Close()

	defer os.Remove(name)
	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Database creation failed: %v", err)
	}
	defer closeFunc()

	if err = db.SetKey("test_key", []byte("test_val")); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	
	val, err := db.GetKey("test_key")
	if err != nil {
		t.Fatalf("Could not get key: %v", err)
	}

	if !bytes.Equal(val, []byte("test_val")) {
		t.Errorf("Wrong value. Expected: test_val, Got: %v", val)
	}
}