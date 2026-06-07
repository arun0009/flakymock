package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MappingsDir returns the mappings directory under root.
func MappingsDir(root string) string {
	return filepath.Join(root, "mappings")
}

// LoadMappingsFromDir reads all scenario files from root/mappings/.
func LoadMappingsFromDir(root string) ([]Scenario, error) {
	dir := MappingsDir(root)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var loaded []Scenario
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		var scenarios []Scenario
		if ext == ".json" {
			if err := json.Unmarshal(data, &scenarios); err != nil {
				var single Scenario
				if err2 := json.Unmarshal(data, &single); err2 != nil {
					return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
				}
				scenarios = []Scenario{single}
			}
		} else {
			if err := yaml.Unmarshal(data, &scenarios); err != nil {
				return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
			}
		}
		loaded = append(loaded, scenarios...)
	}
	return loaded, nil
}

// PersistMappingsToDir writes all in-memory scenarios to root/mappings/ as JSON files.
func PersistMappingsToDir(root string) (int, error) {
	dir := MappingsDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}

	scenarios := ListScenarios()
	written := 0
	for _, summary := range scenarios {
		s, ok := GetScenarioByID(summary.ID)
		if !ok {
			continue
		}
		payload := ScenarioForPersist(s)
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return written, err
		}
		filename := safePersistFilename(summary.ID) + ".json"
		if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

// DeleteMappingFile removes a persisted scenario file by ID.
func DeleteMappingFile(root, id string) error {
	dir := MappingsDir(root)
	path := filepath.Join(dir, safePersistFilename(id)+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, ext := range []string{".yaml", ".yml"} {
		_ = os.Remove(filepath.Join(dir, safePersistFilename(id)+ext))
	}
	return nil
}

func safePersistFilename(id string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return replacer.Replace(id)
}
