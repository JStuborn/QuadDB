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
	"time"
)

type Document struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type Database struct {
	filename string
	aesKey   []byte
}

func LoadDB(filename string, aesKey []byte) *Database {
	util.Info("Found Database")
	return &Database{
		filename: filename,
		aesKey:   aesKey,
	}
}

func (db *Database) LoadDocuments() (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		// If file does not exist, return an empty map
		if os.IsNotExist(err) {
			return make(map[string]json.RawMessage), nil
		}
		return nil, err
	}

	decryptedData, err := db.decrypt(data)
	if err != nil {
		return nil, err
	}

	// Decompress the data (only when needed)
	decompressedData, err := util.Decompress(decryptedData)
	if err != nil {
		return nil, err
	}

	var documents map[string]json.RawMessage
	err = json.Unmarshal(decompressedData, &documents)
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (db *Database) saveDocuments(documents map[string]json.RawMessage) error {
	data, err := json.Marshal(documents)
	if err != nil {
		return err
	}

	// Compress the data before encryption
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

func (db *Database) CreateDocument(key string, data json.RawMessage) error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; exists {
		return fmt.Errorf("document with key '%s' already exists", key)
	}

	documents[key] = data

	// Update last added record
	LastUsedDB = key

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ReadDocument(key string) (json.RawMessage, error) {
	documents, err := db.LoadDocuments()
	if err != nil {
		return nil, err
	}

	data, exists := documents[key]
	if !exists {
		return nil, fmt.Errorf("document with key '%s' not found", key)
	}

	// Update last used database and last update time
	LastUsedDB = db.filename
	LastUpdateTime = time.Now()

	return data, nil
}

func (db *Database) UpdateDocument(key string, data json.RawMessage) error {
	documents, err := db.LoadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	documents[key] = data

	// Update last update time
	LastUpdateTime = time.Now()

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	return nil
}

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

	// Unpad data
	ciphertext = db.unpadData(ciphertext)

	return ciphertext, nil
}

func (db *Database) unpadData(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil
	}
	return data[:(length - unpadding)]
}
