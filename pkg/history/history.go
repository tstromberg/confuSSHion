package history

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
)

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
func (s *Store) SaveSession(sess *SessionContext) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("session:" + sess.SID)
		err := txn.Set(key, data)
		if err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		return nil
	})
}

// GetSession retrieves a session from the database
func (s *Store) GetSession(sid string) (*SessionContext, error) {
	var session SessionContext
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

// Entry represents a single command/response pair in the session
type Entry struct {
	T    time.Time `json:"timestamp"`
	Kind string    `json:"kind`
	In   string    `json:"in"`
	Out  string    `json:"out"`
}

// SessionContext represents a complete SSH session history
type SessionContext struct {
	SID             string    `json:"sid"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	RemoteAddr      string    `json:"remote_addr"`
	User            string    `json:"user"`
	AuthUser        string    `json:"auth_user"`
	PublicKey       string    `json:"public_key"`
	OS              string    `json:"os"`
	Arch            string    `json:"arch"`
	Hostname        string    `json:"hostname"`
	RoleDescription string    `json:"role_description"`
	LoginCommand    []string  `json:"login_command"`
	CurrentCommand  string    `json:"current_command"`
	Environ         []string  `json:"environ"`
	History         []Entry   `json:"log"`
}
