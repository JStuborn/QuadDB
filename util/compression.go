// compression.go

package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// Compress data using gzip
func Compress(data json.RawMessage) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := gzip.NewWriter(buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return buf.Bytes(), nil
}

// Decompress data using gzip
func Decompress(compressedData []byte) (json.RawMessage, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(decompressedData), nil
}
