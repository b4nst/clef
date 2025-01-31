package backend

import (
	"context"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteBinaryMap(t *testing.T) {
	t.Parallel()

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)

		assert.NoError(t, writeBinaryMap(f, map[string]string{}))
	})

	t.Run("non empty map", func(t *testing.T) {
		t.Parallel()
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)

		assert.NoError(t, writeBinaryMap(f, map[string]string{"key1": "value1"}))
		f.Seek(0, 0)
		content, err := io.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x04, 0x00, // Key length (4 bytes for "key1")
			0x6b, 0x65, 0x79, 0x31, // Key ("key1")
			0x06, 0x00, // Value length (6 bytes for "value1")
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
		}, content)
	})

	t.Run("nil file", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, writeBinaryMap(nil, map[string]string{"foo": "bar"}), "invalid argument")
	})
}

func TestReadBinaryMap(t *testing.T) {
	t.Parallel()

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)

		m, err := readBinaryMap(f)
		assert.NoError(t, err)
		assert.Empty(t, m)
	})

	t.Run("nil file", func(t *testing.T) {
		t.Parallel()
		m, err := readBinaryMap(nil)
		assert.NoError(t, err)
		assert.Empty(t, m)
	})

	t.Run("some", func(t *testing.T) {
		t.Parallel()
		content := []byte{
			0x04, 0x00, // Key length (4 bytes for "key1")
			0x6b, 0x65, 0x79, 0x31, // Key ("key1")
			0x06, 0x00, // Value length (6 bytes for "value1")
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
		}
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)
		_, err = f.Write(content)
		require.NoError(t, err)
		f.Seek(0, 0)

		m, err := readBinaryMap(f)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"key1": "value1"}, m)
	})
}

func TestNewFileStore(t *testing.T) {
	t.Parallel()

	t.Run("file not exist", func(t *testing.T) {
		t.Parallel()
		filename := path.Join(t.TempDir(), "testnewfilestore_filenotexist")
		require.NoFileExists(t, filename)

		fs, err := NewFileStore(filename)
		if assert.NoError(t, err) {
			defer fs.Close()
			assert.NotNil(t, fs)
			assert.FileExists(t, filename)
		}
	})
}

func TestFileStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		content := []byte{
			0x04, 0x00, // Key length (4 bytes for "key1")
			0x6b, 0x65, 0x79, 0x31, // Key ("key1")
			0x06, 0x00, // Value length (6 bytes for "value1")
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
		}
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)
		_, err = f.Write(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		fs, err := NewFileStore(f.Name())
		require.NoError(t, err)
		defer fs.Close()

		v, err := fs.Get(context.TODO(), "key1")
		if assert.NoError(t, err) {
			assert.Equal(t, "value1", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		content := []byte{
			0x04, 0x00, // Key length (4 bytes for "key1")
			0x6b, 0x65, 0x79, 0x31, // Key ("key1")
			0x06, 0x00, // Value length (6 bytes for "value1")
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
		}
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)
		_, err = f.Write(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		fs, err := NewFileStore(f.Name())
		require.NoError(t, err)
		defer fs.Close()

		v, err := fs.Get(context.TODO(), "nokey")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Empty(t, v)
	})

	t.Run("empty file", func(t *testing.T) {
		filename := path.Join(t.TempDir(), "testfilestore_get_emptyfile")
		require.NoFileExists(t, filename)
		fs, err := NewFileStore(filename)
		require.NoError(t, err)
		defer fs.Close()

		v, err := fs.Get(context.TODO(), "nokey")
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Empty(t, v)
	})
}

func TestFileStore_Set(t *testing.T) {
	t.Run("new file", func(t *testing.T) {
		filename := path.Join(t.TempDir(), "testfilestore_set_newfile")
		require.NoFileExists(t, filename)
		fs, err := NewFileStore(filename)
		require.NoError(t, err)
		defer fs.Close()

		if assert.NoError(t, fs.Set(context.TODO(), "key1", "value1")) {
			fs.fd.Seek(0, 0)
			content, err := io.ReadAll(fs.fd)
			require.NoError(t, err)
			assert.Equal(t, []byte{
				0x04, 0x00, // Key length (4 bytes for "key1")
				0x6b, 0x65, 0x79, 0x31, // Key ("key1")
				0x06, 0x00, // Value length (6 bytes for "value1")
				0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
			}, content)
		}
	})

	t.Run("existing file", func(t *testing.T) {
		content := []byte{
			0x04, 0x00, // Key length (4 bytes for "key1")
			0x6b, 0x65, 0x79, 0x31, // Key ("key1")
			0x06, 0x00, // Value length (6 bytes for "value1")
			0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
		}
		f, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)
		_, err = f.Write(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		fs, err := NewFileStore(f.Name())
		require.NoError(t, err)
		defer fs.Close()

		if assert.NoError(t, fs.Set(context.TODO(), "key2", "value2")) {
			fs.fd.Seek(0, 0)
			content, err := io.ReadAll(fs.fd)
			require.NoError(t, err)
			assert.Equal(t, []byte{
				0x04, 0x00, // Key length (4 bytes for "key1")
				0x6b, 0x65, 0x79, 0x31, // Key ("key1")
				0x06, 0x00, // Value length (6 bytes for "value1")
				0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, // Value ("value1")
				0x04, 0x00, // Key length (4 bytes for "key2")
				0x6b, 0x65, 0x79, 0x32, // Key ("key2")
				0x06, 0x00, // Value length (6 bytes for "value2")
				0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, // Value ("value2")
			}, content)
		}
	})
}
