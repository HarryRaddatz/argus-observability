package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"log/slog"
)

const bufferNameWidth = 16

// Spool is a FIFO of batches that survived the transport retries.
// Files are named with a zero-padded sequence so a restart keeps the order.
type Spool struct {
	dir      string
	maxBytes int64
	logger   *slog.Logger
}

type spoolItem struct {
	name     string
	Path     string          `json:"path"`
	BatchID  string          `json:"batch_id,omitempty"`
	Body     json.RawMessage `json:"body"`
	QueuedAt time.Time       `json:"queued_at"`
}

func OpenSpool(dir string, maxBytes int64, logger *slog.Logger) (*Spool, error) {
	if dir == "" {
		return nil, fmt.Errorf("buffer dir is empty")
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("buffer max bytes must be positive")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Spool{dir: dir, maxBytes: maxBytes, logger: logger}, nil
}

func (s *Spool) Put(path, batchID string, body []byte) error {
	if !json.Valid(body) {
		return fmt.Errorf("buffer body is not json")
	}
	item := spoolItem{Path: path, BatchID: batchID, Body: append(json.RawMessage(nil), body...), QueuedAt: time.Now().UTC()}
	raw, err := json.Marshal(item)
	if err != nil {
		return err
	}
	if int64(len(raw)) > s.maxBytes {
		return fmt.Errorf("buffer item is larger than the limit")
	}
	names, err := s.names()
	if err != nil {
		return err
	}
	if err := s.makeRoom(names, int64(len(raw))); err != nil {
		return err
	}
	if len(names) > 0 {
		if _, err := os.Stat(filepath.Join(s.dir, names[len(names)-1])); err != nil {
			names, err = s.names()
			if err != nil {
				return err
			}
		}
	}
	seq := int64(1)
	if len(names) > 0 {
		n, err := seqOf(names[len(names)-1])
		if err != nil {
			return err
		}
		seq = n + 1
	}
	final := filepath.Join(s.dir, fmt.Sprintf("%0*d.json", bufferNameWidth, seq))
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// Front returns the oldest readable item. A corrupt file is moved aside and skipped.
// A nil item means the spool is empty.
func (s *Spool) Front() (*spoolItem, error) {
	for {
		names, err := s.names()
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, nil
		}
		name := names[0]
		raw, err := os.ReadFile(filepath.Join(s.dir, name))
		if err != nil {
			return nil, err
		}
		var item spoolItem
		if err := json.Unmarshal(raw, &item); err != nil || item.Path == "" || !json.Valid(item.Body) {
			if renErr := os.Rename(filepath.Join(s.dir, name), filepath.Join(s.dir, name+".bad")); renErr != nil {
				return nil, renErr
			}
			continue
		}
		item.name = name
		return &item, nil
	}
}

func (s *Spool) Remove(name string) error {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		return fmt.Errorf("invalid buffer name")
	}
	err := os.Remove(filepath.Join(s.dir, name))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Spool) names() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !batchName(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func (s *Spool) makeRoom(names []string, incoming int64) error {
	used, err := s.used(names)
	if err != nil {
		return err
	}
	for used+incoming > s.maxBytes && len(names) > 0 {
		oldest := names[0]
		info, err := os.Stat(filepath.Join(s.dir, oldest))
		if err != nil {
			return err
		}
		if err := s.Remove(oldest); err != nil {
			return err
		}
		if s.logger != nil {
			s.logger.Warn("buffer full, dropping oldest batch", "file", oldest)
		}
		used -= info.Size()
		names = names[1:]
	}
	if used+incoming > s.maxBytes {
		return fmt.Errorf("buffer is full")
	}
	return nil
}

func (s *Spool) used(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		info, err := os.Stat(filepath.Join(s.dir, name))
		if err != nil {
			return 0, err
		}
		total += info.Size()
	}
	return total, nil
}

func batchName(name string) bool {
	if len(name) != bufferNameWidth+len(".json") || !strings.HasSuffix(name, ".json") {
		return false
	}
	for _, ch := range name[:bufferNameWidth] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func seqOf(name string) (int64, error) {
	var n int64
	for _, ch := range name[:bufferNameWidth] {
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
