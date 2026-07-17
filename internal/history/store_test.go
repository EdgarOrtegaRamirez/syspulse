package history

import (
	"path/filepath"
	"testing"
)

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := &Store{path: filepath.Join(dir, "history.json")}

	record := Record{
		Timestamp: "2026-01-01T00:00:00Z",
		CPU:       50.0,
		Memory:    60.0,
		Disk:      30.0,
	}

	if err := store.Save(record); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].CPU != 50.0 {
		t.Errorf("expected CPU 50.0, got %f", records[0].CPU)
	}
}

func TestStore_RecordCurrent(t *testing.T) {
	dir := t.TempDir()
	store := &Store{path: filepath.Join(dir, "history.json")}

	if err := store.RecordCurrent(50.0, 60.0, 30.0); err != nil {
		t.Fatalf("RecordCurrent failed: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Memory != 60.0 {
		t.Errorf("expected Memory 60.0, got %f", records[0].Memory)
	}
}

func TestStore_Clear(t *testing.T) {
	dir := t.TempDir()
	store := &Store{path: filepath.Join(dir, "history.json")}

	store.Save(Record{Timestamp: "2026-01-01T00:00:00Z", CPU: 50.0})
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("Load after clear failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after clear, got %d", len(records))
	}
}

func TestStore_LoadNonExistent(t *testing.T) {
	store := &Store{path: "/nonexistent/path/history.json"}
	records, err := store.Load()
	if err != nil {
		t.Fatalf("Load from nonexistent path should not error, got: %v", err)
	}
	if records != nil {
		t.Errorf("expected nil for nonexistent path, got %d records", len(records))
	}
}