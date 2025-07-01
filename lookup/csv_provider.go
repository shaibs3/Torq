package lookup

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

type CSVProvider struct {
	data map[string]record
	mu   sync.RWMutex
}

type record struct {
	city    string
	country string
}

func NewCSVProvider(path string) (*CSVProvider, error) {
	file, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	data := make(map[string]record)
	for _, row := range rows {
		if len(row) < 3 {
			continue // skip invalid rows
		}
		ip := row[0]
		city := row[1]
		country := row[2]
		data[ip] = record{city: city, country: country}
	}

	return &CSVProvider{data: data}, nil
}

func (p *CSVProvider) Lookup(ip string) (string, string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	rec, ok := p.data[ip]
	if !ok {
		return "", "", fmt.Errorf("IP not found")
	}
	return rec.city, rec.country, nil
}
