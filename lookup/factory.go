package lookup

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func GetDbProvider(backend string, logger *zap.Logger) (DbProvider, error) {
	path := os.Getenv("IP_DB_PATH")
	// Todo remove this env variable

	switch backend {
	case "csv":
		return NewCSVProvider(path, logger)
	default:
		return nil, fmt.Errorf("unsupported IP_DB_PROVIDER: %s", backend)
	}
}
