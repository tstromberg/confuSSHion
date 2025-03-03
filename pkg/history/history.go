package history

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/tstromberg/confuSSHion/pkg/personality"
)

// Entry represents a single command/response pair in the session
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
}

// Session represents a complete SSH session history
type Session struct {
	SID        string              `json:"sid"`
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	UserInfo   personality.UserInfo `json:"user_info"`
	NodeConfig personality.NodeConfig `json:"node_config"`
	Log        []Entry             `json:"log"`
}

// Store handles the persistence of session histories
type Store struct {
	db *badger.DB
}

// NewStore creates a new history store with the specified database path
func NewStore(dbPath string) (*Store, error) {
	opts := badger.DefaultOptions(dbPath)
	// Reduce logging noise from BadgerDB
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveSession persists a session to the database
func (s *Store) SaveSession(session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("session:" + session.SID)
		err := txn.Set(key, data)
		if err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		return nil
	})
}

// GetSession retrieves a session from the database
func (s *Store) GetSession(sid string) (*Session, error) {
	var session Session
	err := s.db.View(func(txn *badger.Txn) error {
		key := []byte("session:" + sid)
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &session)
		})
	})

	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions returns all session IDs in the database
func (s *Store) ListSessions() ([]string, error) {
	var sessions []string

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("session:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())
			sessions = append(sessions, key[len("session:"):])
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return sessions, nil
}
