package types

import (
	"encoding/json"
	"sync"
)

// Document represents a document in the database.
type Document struct {
	Key  string          `json:"key"`
	Data json.RawMessage `json:"data"`
}

// Database represents a document-based database or else it shits itself
// it also just shits itself whenever it wants
type Database struct {
	documents map[string]json.RawMessage
	mu        sync.RWMutex
	filename  string // Unencrypted filename to store the database
	aesKey    []byte // AES encryption key"
}
