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

	db, closeFunc, err := db.NewDatabase(name, false)
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

func createTempDB(t *testing.T, readOnly bool) *db.Database {
	t.Helper()

	file, err := os.CreateTemp(os.TempDir(), "tempkvdb")
	if err != nil {
		t.Fatalf("Could not create temp file: %v\n", err)
	}
	name := file.Name()
	file.Close()
	t.Cleanup(func() {os.Remove(name)})

	db, closeFunc, err := db.NewDatabase(name, readOnly)
	if err != nil {
		t.Fatalf("Failed to create a new DB: %v\n", err)
	}
	t.Cleanup(func() {closeFunc()})

	return db
}
func TestGetSet(t *testing.T) {
	db := createTempDB(t, false)

	setKey(t, db, "test_key", "test_val")
	
	val, err := db.GetKey("test_key")
	if err != nil {
		t.Fatalf("Could not get key: %v", err)
	}

	if !bytes.Equal(val, []byte("test_val")) {
		t.Errorf("Wrong value. Expected: test_val, Got: %v", val)
	}
}

func TestDeleteReplicationKey(t *testing.T) {
	db := createTempDB(t, false)

	setKey(t, db, "test_key", "test_val")

	k, v, err := db.GetNextReplicaKey()
	if err != nil {
		t.Fatalf("GetNextReplicaKey error: %v\n", err)
	}

	if !bytes.Equal(k, []byte("test_key")) || !bytes.Equal(v, []byte("test_val")) {
		t.Errorf("GetNextReplicaKey error. Got %q: %q, wanted %q: %q\n", k, v, "test_key", "test_val")
	}

	if err := db.DeleteReplicaKey([]byte("test_key"), []byte("test_value")); err == nil {
		t.Fatalf(`DeleteReplicaKey("test_key", "test_val") should have thrown error, but it did not`)
	}

	if err := db.DeleteReplicaKey([]byte("test_key"), []byte("test_val")); err != nil {
		t.Fatalf(`DeleteReplicaKey("test_key", "test_val") got error: %q`, err)
	}

	k, v, err = db.GetNextReplicaKey()
	if err != nil {
		t.Fatalf("GetNextReplicaKey error: %v\n", err)
	}

	if k != nil || v != nil {
		t.Errorf("GetNextReplicaKey error. Got %q: %q, wanted nil: nil\n", k, v)
	}
}

func TestGetSetReadOnly(t *testing.T) {

	db := createTempDB(t, true)

	setKey(t, db, "test_key", "shouldn't_work")
	
	val, err := db.GetKey("test_key")
	if err != nil {
		t.Fatalf("Could not get key: %v", err)
	}

	if !bytes.Equal(val, []byte("test_val")) {
		t.Errorf("Wrong value. Expected: test_val, Got: %v", val)
	}
}