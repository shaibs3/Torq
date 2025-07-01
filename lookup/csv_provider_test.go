package lookup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test-*.csv")
	require.NoError(t, err)
	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())
	return tmpfile.Name()
}

func TestNewCSVProvider_ValidFile(t *testing.T) {
	csvContent := "1.2.3.4,New York,USA\n5.6.7.8,London,UK\n"
	path := createTempCSV(t, csvContent)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove temp file: %v", err)
		}
	}(path)

	provider, err := NewCSVProvider(path)
	require.NoError(t, err)

	city, country, err := provider.Lookup("1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, "New York", city)
	assert.Equal(t, "USA", country)
}

func TestNewCSVProvider_InvalidFile(t *testing.T) {
	_, err := NewCSVProvider("nonexistent.csv")
	assert.Error(t, err)
}

func TestCSVProvider_Lookup_NotFound(t *testing.T) {
	csvContent := "1.2.3.4,New York,USA\n"
	path := createTempCSV(t, csvContent)
	defer os.Remove(path) //nolint:errcheck

	provider, err := NewCSVProvider(path)
	require.NoError(t, err)

	_, _, err = provider.Lookup("8.8.8.8")
	assert.Error(t, err)
}
