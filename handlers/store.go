package handlers

import (
	"encoding/json"
	"fmt"
	"os"
)

// getFileName checks if a test file is requested, otherwise defaults to production.
func GetFileName() string {
	name := os.Getenv("DB_FILE")
	if name == "" {
		return "todos.json" // Default for production
	}
	return name
}

// SaveToJSON handles Atomic Writes to prevent data corruption.
func SaveToJSON() error {
	Mu.RLock()
	defer Mu.RUnlock()

	data, err := json.MarshalIndent(Todos, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal data: %v", err)
	}

	// USE THE FUNCTION HERE
	filename := GetFileName()
	tempFile := filename + ".tmp"

	err = os.WriteFile(tempFile, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write to temp file: %v", err)
	}

	// USE THE FUNCTION HERE
	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("could not rename temp file: %v", err)
	}

	return nil
}

// LoadFromJSON populates the memory slice from the disk on startup.
func LoadFromJSON() error {
	Mu.Lock()
	defer Mu.Unlock()

	// USE THE FUNCTION HERE
	filename := GetFileName()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	fileData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(fileData, &Todos); err != nil {
		return err
	}

	// Set NextID to be higher than the highest existing ID
	for _, t := range Todos {
		if t.ID >= NextID {
			NextID = t.ID + 1
		}
	}

	return nil
}
