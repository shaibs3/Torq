package lookup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDbType_IsValid(t *testing.T) {
	assert.True(t, DbTypeCSV.IsValid())
	assert.False(t, DbType("invalid").IsValid())
}

func TestDbType_String(t *testing.T) {
	assert.Equal(t, "csv", DbTypeCSV.String())
}

func TestGetDbProvider_CSV(t *testing.T) {
	logger := zap.NewNop()

	// Create a temporary CSV file for testing
	tempFile, err := os.CreateTemp("", "test-*.csv")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	// Write test data to the temporary file
	testData := "IP,CITY,COUNTRY\n1.2.3.4,New York,USA\n5.6.7.8,London,UK\n"
	_, err = tempFile.WriteString(testData)
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	// Test CSV provider configuration
	configJSON := `{
		"dbtype": "csv",
		"extra_details": {
			"file_path": "` + tempFile.Name() + `"
		}
	}`

	provider, err := GetDbProvider(configJSON, logger)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test that the provider actually works
	city, country, err := provider.Lookup("1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, "New York", city)
	assert.Equal(t, "USA", country)
}

func TestGetDbProvider_InvalidDbType(t *testing.T) {
	logger := zap.NewNop()

	// Test invalid database type
	configJSON := `{
		"dbtype": "invalid_type",
		"extra_details": {}
	}`

	provider, err := GetDbProvider(configJSON, logger)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestGetDbProvider_MissingRequiredField(t *testing.T) {
	logger := zap.NewNop()

	// Test missing required field for CSV provider
	configJSON := `{
		"dbtype": "csv",
		"extra_details": {}
	}`

	provider, err := GetDbProvider(configJSON, logger)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "file_path is required for CSV provider")
}

func TestGetDbProvider_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()

	// Test invalid JSON
	configJSON := `{ invalid json }`

	provider, err := GetDbProvider(configJSON, logger)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to parse database configuration JSON")
}
