package obsidian

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type DiscoveredVault struct {
	SourceID  string `json:"source_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Open      bool   `json:"open"`
	Available bool   `json:"available"`
}

// DiscoverObsidianVaults 只读解析 Obsidian 官方注册表。
// 它不会创建、修改或格式化 obsidian.json。
func DiscoverObsidianVaults() ([]DiscoveredVault, error) {
	path, err := ObsidianConfigFile()
	if err != nil {
		return nil, err
	}
	return DiscoverObsidianVaultsFrom(path)
}

func DiscoverObsidianVaultsFrom(configPath string) ([]DiscoveredVault, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read Obsidian discovery config %s: %w", configPath, err)
	}

	var registry ObsidianVaultConfig
	if err := json.Unmarshal(content, &registry); err != nil {
		return nil, fmt.Errorf("parse Obsidian discovery config %s: %w", configPath, err)
	}
	if registry.Vaults == nil {
		return nil, fmt.Errorf("parse Obsidian discovery config %s: vaults must not be null", configPath)
	}

	result := make([]DiscoveredVault, 0, len(registry.Vaults))
	for sourceID, entry := range registry.Vaults {
		path := entry.Path
		if RunningInWSL() {
			path = adjustForWslMount(path)
		}
		path = filepath.Clean(path)
		info, statErr := os.Stat(path)
		result = append(result, DiscoveredVault{
			SourceID:  sourceID,
			Name:      vaultBaseName(path),
			Path:      path,
			Open:      entry.Open,
			Available: statErr == nil && info.IsDir(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Open != result[j].Open {
			return result[i].Open
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].SourceID < result[j].SourceID
	})
	return result, nil
}
