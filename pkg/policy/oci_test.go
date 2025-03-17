package policy

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

var multilayeredImage = "ghcr.io/nirmata/demo-image-compliance-policies:multilayered"

func Test_OCIFetcherMultiLayered(t *testing.T) {
	o, err := NewOCIPolicyFetcher(context.Background(), logr.Discard(), multilayeredImage, 0, nil, nil)
	assert.NoError(t, err)
	ivpols, err := o.Fetch()
	assert.NoError(t, err)
	assert.Equal(t, len(ivpols), 2)
	assert.Equal(t, ivpols[0].Name, "sample-2")
	assert.Equal(t, ivpols[1].Name, "sample-1")
}
