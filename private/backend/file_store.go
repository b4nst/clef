package backend

import (
	"context"
	"encoding/binary"
	"os"
)

// FIleStore uses a file to store and retrieve values.
// FileStore is **not** recommended in a production environment.
// Please use it only for testing purposes.
type FileStore struct {
	fd *os.File
}

// NewFileStore creates a new FileStore baked by the file at filename.
// It will create the file if it doesn't exist.
func NewFileStore(filename string) (*FileStore, error) {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &FileStore{fd}, nil
}

// Close closes the filestore and its underlying file.
func (fs *FileStore) Close() error {
	return fs.fd.Close()
}

// Get implements the Store.Get method.
func (fs *FileStore) Get(ctx context.Context, k string) (string, error) {
	m, err := readBinaryMap(fs.fd)
	if err != nil {
		return "", err
	}
	v, ok := m[k]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

// Set implements the Store.Set method
func (fs *FileStore) Set(ctx context.Context, k, v string) error {
	m, err := readBinaryMap(fs.fd)
	if err != nil {
		return err
	}
	m[k] = v
	return writeBinaryMap(fs.fd, m)
}

func writeBinaryMap(file *os.File, data map[string]string) error {
	file.Seek(0, 0)
	for key, value := range data {
		keyLen := uint16(len(key))
		valueLen := uint16(len(value))

		// Write key length + key bytes
		if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
			return err
		}
		if _, err := file.Write([]byte(key)); err != nil {
			return err
		}

		// Write value length + value bytes
		if err := binary.Write(file, binary.LittleEndian, valueLen); err != nil {
			return err
		}
		if _, err := file.Write([]byte(value)); err != nil {
			return err
		}
	}
	return nil
}

func readBinaryMap(file *os.File) (map[string]string, error) {
	data := make(map[string]string)

	file.Seek(0, 0)
	for {
		var keyLen uint16
		if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
			break
		}
		key := make([]byte, keyLen)
		if _, err := file.Read(key); err != nil {
			return nil, err
		}

		var valueLen uint16
		if err := binary.Read(file, binary.LittleEndian, &valueLen); err != nil {
			return nil, err
		}
		value := make([]byte, valueLen)
		if _, err := file.Read(value); err != nil {
			return nil, err
		}

		data[string(key)] = string(value)
	}
	return data, nil
}
