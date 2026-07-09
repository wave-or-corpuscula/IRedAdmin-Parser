package apptesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAuthConfigs(t *testing.T) {
	configs, err := GetAuthConfigs()
	assert.NoError(t, err)

	assert.True(t, len(configs) > 0)
}
