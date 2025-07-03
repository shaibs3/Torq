package lookup

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ProviderFactory defines the interface for creating database providers
type ProviderFactory interface {
	CreateProvider(configJSON string) (DbProvider, error)
}

// Factory implements ProviderFactory for creating database providers
type Factory struct {
	logger *zap.Logger
}

// NewFactory creates a new factory instance
func NewFactory(logger *zap.Logger) *Factory {
	return &Factory{
		logger: logger.Named("factory"),
	}
}

func (f *Factory) CreateProvider(configJSON string) (DbProvider, error) {
	var config DbProviderConfig
	f.logger.Info("parsing configuration", zap.String("configJSON", configJSON))

	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse database configuration JSON: %w", err)
	}

	f.logger.Info("creating database provider",
		zap.String("db_type", config.DbType.String()),
		zap.Any("extra_details", config.ExtraDetails))

	// Validate database type
	if !config.DbType.IsValid() {
		return nil, fmt.Errorf("unsupported database type: %s", config.DbType)
	}

	switch config.DbType {
	case DbTypeCSV:
		return NewCSVProvider(config, f.logger)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.DbType)
	}
}
