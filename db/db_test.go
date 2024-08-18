package db_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/prajwal-annigeri/kv-store/db"
)

func setKey(t *testing.T, d *db.Database, key, value string) {
	t.Helper()

	if err := d.SetKey(key, []byte(value)); err != nil {
		t.Fatalf("SetKey(%q, %q) failed: %v", key, value, err)
	}
}

func getKey(t *testing.T, d *db.Database, key string) string {
	t.Helper()
	val, err := d.GetKey(key)
	if err != nil {
		t.Fatalf("GetKey(%q) failed: %v", key, err)
	}
	return string(val)
}

func TestDeleteExtraKeys(t *testing.T) {
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

	setKey(t, db, "test_key", "test_val")
	setKey(t, db, "test_key2", "test_val2")

	if val := getKey(t, db, "test_key"); val != "test_val" {
		t.Fatalf("Wrong value. Expected: test_val, Got: %v", val)
	}

	if err := db.DeleteExtraKeys(func (name string) bool {
		return name == "test_key2"
	}); err != nil {
		t.Fatalf("Error deleting extra keys: %v", err)
	}

	if val := getKey(t, db, "test_key"); val == "" {
		t.Fatalf(`Wrong value. Expected: "test_value", Got: "%v"`, val)
	}
	if val := getKey(t, db, "test_key2"); val != "" {
		t.Fatalf(`Wrong value. Expected: "", Got: "%v"`, val)
	}
}

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

	setKey(t, db, "test_key", "test_val")
	
	val, err := db.GetKey("test_key")
	if err != nil {
		t.Fatalf("Could not get key: %v", err)
	}

	if !bytes.Equal(val, []byte("test_val")) {
		t.Errorf("Wrong value. Expected: test_val, Got: %v", val)
	}
}