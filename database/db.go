package database

import (
	"CyberDefenseEd/QuadDB/util"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type Document struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type Database struct {
	filename   string
	aesKey     []byte
	fieldIndex map[string]map[string][]string
	indexLock  sync.RWMutex
}

// LoadDB initializes a new Database instance
func LoadDB(filename string, aesKey []byte) *Database {
	db := &Database{
		filename:   filename,
		aesKey:     aesKey,
		fieldIndex: make(map[string]map[string][]string),
	}

	// Load existing documents and build indexes
	db.buildIndex()

	return db
}

// buildIndex builds the index for all documents based on their fields
func (db *Database) buildIndex() error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	db.indexLock.Lock()
	defer db.indexLock.Unlock()

	for key, rawMessage := range documents {
		var docMap map[string]interface{}
		err := json.Unmarshal(rawMessage, &docMap)
		if err != nil {
			return err
		}

		for fieldPath := range docMap {
			fieldValue, found := traverseNestedFields(strings.Split(fieldPath, "."), docMap)
			if found {
				lowerFieldPath := strings.ToLower(fieldPath)
				lowerFieldValue := strings.ToLower(fieldValue)

				if db.fieldIndex[lowerFieldPath] == nil {
					db.fieldIndex[lowerFieldPath] = make(map[string][]string)
				}
				db.fieldIndex[lowerFieldPath][lowerFieldValue] = append(db.fieldIndex[lowerFieldPath][lowerFieldValue], key)
			}
		}
	}

	return nil
}

// LoadDocuments reads and decrypts the database file, returning the documents as a map
func (db *Database) LoadDocuments() (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]json.RawMessage), nil
		}
		return nil, err
	}

	decryptedData, err := db.decrypt(data)
	if err != nil {
		return nil, err
	}

	decompressedData, err := util.Decompress(decryptedData)
	if err != nil {
		return nil, err
	}

	var documents map[string]json.RawMessage
	err = msgpack.Unmarshal(decompressedData, &documents)
	if err != nil {
		return nil, err
	}

	return documents, nil
}

// LoadDocumentsPaginated reads and decrypts the database file and returns a paginated subset of documents
func (db *Database) LoadDocumentsPaginated(offset, limit int) (map[string]json.RawMessage, error) {
	documents, err := db.LoadDocuments()
	if err != nil {
		return nil, err
	}

	paginatedDocuments := make(map[string]json.RawMessage)
	keys := make([]string, 0, len(documents))
	for key := range documents {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	if offset >= len(keys) {
		// If the offset is beyond the number of keys, return an empty result else panic err
		return paginatedDocuments, nil
	}

	end := offset + limit
	if end > len(keys) {
		end = len(keys)
	}

	for _, key := range keys[offset:end] {
		paginatedDocuments[key] = documents[key]
	}

	return paginatedDocuments, nil
}

// saveDocuments compresses, encrypts, and writes the documents map to the database file
func (db *Database) saveDocuments(documents map[string]json.RawMessage) error {
	data, err := msgpack.Marshal(documents)
	if err != nil {
		return err
	}

	compressedData, err := util.Compress(data)
	if err != nil {
		return err
	}

	encryptedData, err := db.encrypt(compressedData)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.filename, encryptedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// CreateDocument adds a new document with a unique key; generates a UUID if the key is empty
func (db *Database) CreateDocument(key string, data json.RawMessage) error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	if key == "" {
		key = uuid.New().String()
	}

	if _, exists := documents[key]; exists {
		return fmt.Errorf("document with key '%s' already exists", key)
	}

	documents[key] = data
	LastUsedDB = key

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	// Update the index
	db.indexLock.Lock()
	defer db.indexLock.Unlock()
	var docMap map[string]interface{}
	err = json.Unmarshal(data, &docMap)
	if err != nil {
		return err
	}
	for fieldPath := range docMap {
		fieldValue, found := traverseNestedFields(strings.Split(fieldPath, "."), docMap)
		if found {
			if db.fieldIndex[fieldPath] == nil {
				db.fieldIndex[fieldPath] = make(map[string][]string)
			}
			db.fieldIndex[fieldPath][fieldValue] = append(db.fieldIndex[fieldPath][fieldValue], key)
		}
	}

	return nil
}

// ReadDocument retrieves a document by key
func (db *Database) ReadDocument(key string) (json.RawMessage, error) {
	documents, err := db.LoadDocuments()
	if err != nil {
		return nil, err
	}

	data, exists := documents[key]
	if !exists {
		return nil, fmt.Errorf("document with key '%s' not found", key)
	}

	LastUsedDB = db.filename
	LastUpdateTime = time.Now()

	return data, nil
}

// UpdateDocument modifies an existing document by key
func (db *Database) UpdateDocument(key string, data json.RawMessage) error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	documents[key] = data
	LastUpdateTime = time.Now()

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	// Update the index
	db.buildIndex()

	return nil
}

// DeleteDocument removes a document by key
func (db *Database) DeleteDocument(key string) error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	delete(documents, key)

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	// Update the index
	db.buildIndex()

	return nil
}

// CountDocuments returns the number of documents in the database
func (db *Database) CountDocuments() (int, error) {
	documents, err := db.LoadDocuments()
	if err != nil {
		return 0, err
	}
	return len(documents), nil
}

// encrypt encrypts data using AES in CBC mode
func (db *Database) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(db.aesKey)
	if err != nil {
		return nil, err
	}

	data = util.PadData(data, block.BlockSize())

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// decrypt decrypts data using AES in CBC mode
func (db *Database) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(db.aesKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	ciphertext = db.unpadData(ciphertext)

	return ciphertext, nil
}

// unpadData removes padding from the decrypted data
func (db *Database) unpadData(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil
	}
	return data[:(length - unpadding)]
}

// WHAT THE FUCK IS A KOLOMITORRR 🦅🦅
// FetchDocumentsByFieldValues returns documents matching specified field-value pairs
func (db *Database) FetchDocumentsByFieldValues(fieldValues map[string]string) (map[string]json.RawMessage, error) {
	db.indexLock.RLock()
	defer db.indexLock.RUnlock()

	matchingKeys := make(map[string]int) // Map to count matching fields

	for fieldPath, value := range fieldValues {
		lowerFieldPath := strings.ToLower(fieldPath)
		lowerValue := strings.ToLower(value)

		keys, exists := db.fieldIndex[lowerFieldPath][lowerValue]
		if !exists {
			// If any field does not match, no need to proceed further
			return nil, nil
		}

		for _, key := range keys {
			matchingKeys[key]++
		}
	}

	// Filter out keys that matched all field-value pairs
	expectedMatches := len(fieldValues)
	documents, err := db.LoadDocuments()
	if err != nil {
		return nil, err
	}

	matchingDocuments := make(map[string]json.RawMessage)
	for key, matchCount := range matchingKeys {
		if matchCount == expectedMatches {
			if data, exists := documents[key]; exists {
				matchingDocuments[key] = data
			}
		}
	}

	return matchingDocuments, nil
}

// traverseNestedFields retrieves a nested field value from a document map (returns string and is case-insensitive)
func traverseNestedFields(fields []string, docMap map[string]interface{}) (string, bool) {
	fieldValue := interface{}(docMap)

	for _, field := range fields {
		switch v := fieldValue.(type) {
		case map[string]interface{}:
			var found bool
			fieldValue, found = v[field]
			if !found {
				return "", false
			}
		case []interface{}:
			var sliceValues []string
			for _, elem := range v {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					if val, ok := elemMap[field]; ok {
						sliceValues = append(sliceValues, fmt.Sprintf("%v", val))
					}
				}
			}
			if len(sliceValues) > 0 {
				return strings.Join(sliceValues, ","), true
			}
			return "", false
		default:
			return "", false
		}
	}

	return fmt.Sprintf("%v", fieldValue), true
}
