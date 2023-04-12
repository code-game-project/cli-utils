package cgfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type CodeGameFileData struct {
	Game        string         `json:"game"`
	GameVersion string         `json:"game_version,omitempty"`
	Type        string         `json:"type"`
	Lang        string         `json:"lang"`
	LangConfig  map[string]any `json:"lang_config,omitempty"`
	URL         string         `json:"url,omitempty"`
}

func Load(projectRoot string) (*CodeGameFileData, error) {
	file, err := os.Open(filepath.Join(projectRoot, ".codegame.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &CodeGameFileData{
		LangConfig: make(map[string]any),
	}
	err = json.NewDecoder(file).Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *CodeGameFileData) Write(dir string) error {
	os.MkdirAll(dir, 0o755)

	file, err := os.Create(filepath.Join(dir, ".codegame.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
