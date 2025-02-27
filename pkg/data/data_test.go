package data

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrds(t *testing.T) {
	data, err := Crds()
	assert.NoError(t, err)
	{
		file, err := fs.Stat(data, "policies.kyverno.io_imageverificationpolicies.yaml")
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.False(t, file.IsDir())
	}
}
