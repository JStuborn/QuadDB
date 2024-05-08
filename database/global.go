package database

import (
	"encoding/json"
	"time"
)

var (
	LastUsedDB      string
	LastUpdateTime  time.Time
	LastAddedRecord string
	LastReadRecord  json.RawMessage
)