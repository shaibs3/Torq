package lookup

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// DbType represents the supported database types
type DbType string

const (
	DbTypeCSV DbType = "csv"
	// Add more database types here as you implement them
	// DbTypeDatabase DbType = "database"
	// DbTypeMemory   DbType = "memory"
)

// String returns the string representation of the database type
func (dt DbType) String() string {
	return string(dt)
}

// IsValid checks if the database type is supported
func (dt DbType) IsValid() bool {
	switch dt {
	case DbTypeCSV:
		return true
	default:
		return false
	}
}

// DbConfig represents the configuration for a database provider
type DbConfig struct {
	DbType       DbType                 `json:"dbtype"`
	ExtraDetails map[string]interface{} `json:"extra_details"`
}

// GetDbProvider creates a database provider based on JSON configuration
func GetDbProvider(configJSON string, logger *zap.Logger) (DbProvider, error) {
	var config DbConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse database configuration JSON: %w", err)
	}

	logger.Info("creating database provider",
		zap.String("db_type", config.DbType.String()),
		zap.Any("extra_details", config.ExtraDetails))

	// Validate database type
	if !config.DbType.IsValid() {
		return nil, fmt.Errorf("unsupported database type: %s", config.DbType)
	}

	switch config.DbType {
	case DbTypeCSV:
		return createCSVProvider(config, logger)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.DbType)
	}
}

// createCSVProvider creates a CSV provider from configuration using existing NewCSVProvider
func createCSVProvider(config DbConfig, logger *zap.Logger) (DbProvider, error) {
	filePath, ok := config.ExtraDetails["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required for CSV provider")
	}

	logger.Info("creating CSV provider", zap.String("file_path", filePath))
	return NewCSVProvider(filePath, logger)
}
