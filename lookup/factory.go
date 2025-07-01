package lookup

import (
	"fmt"
	"os"
)

func NewProvider() (LookupProvider, error) {
	backend := os.Getenv("IP_DB_PROVIDER")
	path := os.Getenv("IP_DB_PATH")

	switch backend {
	case "csv":
		return NewCSVProvider(path)
	default:
		return nil, fmt.Errorf("unsupported IP_DB_PROVIDER: %s", backend)
	}
}
