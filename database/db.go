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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type Document struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type Database struct {
	filename string
	aesKey   []byte
}

// LoadDB initializes a new Database instance
func LoadDB(filename string, aesKey []byte) *Database {
	return &Database{
		filename: filename,
		aesKey:   aesKey,
	}
}

// validatePath ensures the provided userPath is within the basePath directory
func validatePath(basePath, userPath string) (string, error) {
	cleanedPath := filepath.Clean(userPath)
	fullPath := filepath.Join(basePath, cleanedPath)

	if !strings.HasPrefix(fullPath, filepath.Clean(basePath)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path")
	}

	return fullPath, nil
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

// FetchDocumentsByFieldValue returns documents matching a specified field value
func (db *Database) FetchDocumentsByFieldValue(fieldPath string, value string) (map[string]json.RawMessage, error) {
	documents, err := db.LoadDocuments()
	if err != nil {
		return nil, err
	}

	matchingDocuments := make(map[string]json.RawMessage)

	for key, rawMessage := range documents {
		var docMap map[string]interface{}
		err := json.Unmarshal(rawMessage, &docMap)
		if err != nil {
			return nil, err
		}

		fieldValue, found := traverseNestedFields(strings.Split(fieldPath, "."), docMap)
		if found && fieldValue == value {
			matchingDocuments[key] = rawMessage
		}
	}

	return matchingDocuments, nil
}

// traverseNestedFields retrieves a nested field value from a document map
func traverseNestedFields(fields []string, docMap map[string]interface{}) (string, bool) {
	var fieldValue interface{} = docMap
	var found bool

	for _, field := range fields {
		if nestedMap, ok := fieldValue.(map[string]interface{}); ok {
			fieldValue, found = nestedMap[field]
			if !found {
				return "", false
			}
		} else if nestedSlice, ok := fieldValue.([]interface{}); ok {
			var sliceValues []string
			for _, elem := range nestedSlice {
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
		} else {
			return "", false
		}
	}

	return fmt.Sprintf("%v", fieldValue), true
}
