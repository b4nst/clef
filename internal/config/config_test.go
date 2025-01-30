package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Parse(t *testing.T) {
	t.Parallel()

	t.Run("nominal config", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "configparse_nominal_store")
		require.NoError(t, err)

		conf := fmt.Sprintf(`
		 	default_store = "file"

		 	[stores.file]
		 	type = "filestore"
		 	[stores.file.config]
		 	path = "%s"
		 `, f.Name())
		c, err := Parse(conf)

		if assert.NoError(t, err) {
			assert.Equal(t, "file", c.DefaultStore)
			assert.Len(t, c.Stores, 1)
		}
	})

	t.Run("missing type", func(t *testing.T) {
		t.Parallel()
		conf := `
		 	default_store = "file"

		 	[stores.file]
		 	[stores.file.config]
		 	path = "foo"
		 `
		_, err := Parse(conf)

		assert.EqualError(t, err, "missing type for store file")
	})
}
