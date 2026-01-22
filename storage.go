package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Storage handles all data persistence and business logic
type Storage struct {
	mu sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{}
}

// SaveConfig saves the fan curve configuration to disk
func (s *Storage) Save(name string, config any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(name, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the configuration from disk
func (s *Storage) LoadConfig() (*FanCurveConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no saved configuration found")
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FanCurveConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
