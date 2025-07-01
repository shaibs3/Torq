package lookup

import (
	"fmt"
	"os"
)

func NewProvider(backend string) (LookupProvider, error) {
	path := os.Getenv("IP_DB_PATH")
	// Todo remove this env variable

	switch backend {
	case "csv":
		return NewCSVProvider(path)
	default:
		return nil, fmt.Errorf("unsupported IP_DB_PROVIDER: %s", backend)
	}
}
