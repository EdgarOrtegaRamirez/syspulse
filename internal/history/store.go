// Package history provides historical resource usage tracking.
package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Record holds a historical resource usage snapshot.
type Record struct {
	Timestamp string  `json:"timestamp"`
	CPU       float64 `json:"cpu_percent"`
	Memory    float64 `json:"memory_percent"`
	Disk      float64 `json:"disk_percent"`
}

// Store manages historical records.
type Store struct {
	path string
}

// New creates a new Store.
func New() *Store {
	cacheDir := filepath.Join(os.TempDir(), "syspulse")
	return &Store{
		path: filepath.Join(cacheDir, "history.json"),
	}
}

// Save adds a record to history.
func (s *Store) Save(record Record) error {
	os.MkdirAll(filepath.Dir(s.path), 0755)

	records := s.load()
	records = append(records, record)
	// Keep last 1000 records
	if len(records) > 1000 {
		records = records[len(records)-1000:]
	}

	data, err := json.Marshal(records)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Load returns all stored records.
func (s *Store) Load() ([]Record, error) {
	return s.load(), nil
}

func (s *Store) load() []Record {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil
	}
	return records
}

// Clear removes all history.
func (s *Store) Clear() error {
	return os.Remove(s.path)
}

// RecordCurrent records the current system state.
func (s *Store) RecordCurrent(cpu, mem, disk float64) error {
	return s.Save(Record{
		Timestamp: time.Now().Format(time.RFC3339),
		CPU:       cpu,
		Memory:    mem,
		Disk:      disk,
	})
}
