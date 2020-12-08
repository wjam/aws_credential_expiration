package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsFile_preferEnvVar(t *testing.T) {
	existing, ok := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE")
	defer func() {
		if ok {
			_ = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", existing)
		} else {
			_ = os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		}
	}()

	expected := "/env/var/directory"
	_ = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", expected)
	actual, err := credentialsFile()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestCredentialsFile_fallBackToHomeDir(t *testing.T) {
	dir := t.TempDir()
	existing := osUserHomeDir
	defer func() {
		osUserHomeDir = existing
	}()
	osUserHomeDir = func() (string, error) {
		return dir, nil
	}

	actual, err := credentialsFile()
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s/.aws/credentials", dir), actual)
}
