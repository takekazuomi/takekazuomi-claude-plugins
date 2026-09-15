package workspace

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load は workspace.yaml を読み込む。
func Load(path string) (*Workspace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := yaml.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &ws, nil
}

// LoadOverlay は workspace.local.yaml を読み込む。
// ファイルが無ければ空の Overlay を返す（Overlay は任意）。
func LoadOverlay(path string) (*Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Overlay{}, nil
		}
		return nil, err
	}
	var ov Overlay
	if err := yaml.Unmarshal(data, &ov); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &ov, nil
}
