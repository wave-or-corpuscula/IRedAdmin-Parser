package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePasswords(t *testing.T) {
	N := 1000
	for range N {
		password, err := GeneratePassword()
		assert.NoError(t, err)
		assert.Len(t, password, 10)
	}
}
