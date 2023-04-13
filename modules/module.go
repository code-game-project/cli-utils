package modules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-game-project/cli-utils/config"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/request"
)

const FeedbackPkg = feedback.Package("modules")

var (
	ErrUnsupportedProjectType     = errors.New("unsupported project type")
	ErrUnsupportedCodeGameVersion = errors.New("unsupported codegame version")
)

var modules = make(map[string]*Module)

type Module struct {
	Lang                   string
	DisplayName            string
	clientCGToLibVersions  map[string]string // CodeGame version -> library version
	serverCGToLibVersions  map[string]string // CodeGame version -> library version
	clientLibToModVersions map[string]string // client library version -> module version
	serverLibToModVersions map[string]string // server library version -> module version
	installedExecutables   map[string]string // module version -> executable path

	provider     provider
	providerVars map[string]any
}

var rawModules map[string]rawModule

type rawModule struct {
	DisplayName               string          `json:"display_name"`
	Source                    map[string]any  `json:"source"`
	LibraryToModuleVersions   json.RawMessage `json:"library_to_module_versions"`
	CodeGameToLibraryVersions json.RawMessage `json:"codegame_to_library_versions"`
}

func LoadModule(lang string) (*Module, error) {
	if rawModules == nil {
		err := loadModules()
		if err != nil {
			return nil, err
		}
	}
	if m, ok := modules[lang]; ok {
		return m, nil
	}
	m, ok := rawModules[lang]
	if !ok {
		return nil, fmt.Errorf("no module available for lang '%s'", lang)
	}

	if m.Source == nil {
		return nil, errors.New("missing 'source' field")
	}

	module := &Module{
		Lang:                   lang,
		DisplayName:            m.DisplayName,
		providerVars:           make(map[string]any),
		clientLibToModVersions: make(map[string]string),
		clientCGToLibVersions:  make(map[string]string),
		serverLibToModVersions: make(map[string]string),
		serverCGToLibVersions:  make(map[string]string),
		installedExecutables:   make(map[string]string),
	}

	providerNameAny, ok := m.Source["provider"]
	if !ok {
		return nil, errors.New("missing 'source.provider' field")
	}
	providerName, ok := providerNameAny.(string)
	if !ok {
		return nil, errors.New("value of 'source.provider' field must be a string")
	}

	prov, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("unknown module provider: %s", providerName)
	}
	module.provider = prov

	for n, v := range m.Source {
		if n != "provider" {
			module.providerVars[n] = v
		}
	}

	errs := prov.ValidateProviderVars(module.providerVars)
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid module source: %s", strings.Join(errs, ", "))
	}

	err := module.loadVersions(m.LibraryToModuleVersions, m.CodeGameToLibraryVersions)
	if err != nil {
		return nil, err
	}

	err = module.loadInstalledVersions()
	if err != nil {
		return nil, fmt.Errorf("load installed versions: %w", err)
	}

	modules[lang] = module

	return module, nil
}

func (m *Module) loadVersions(libraryToModuleVersions, codegameToLibraryVersions json.RawMessage) error {
	var err error

	if m.provider.Name() != "local" {
		m.clientLibToModVersions, m.serverLibToModVersions, err = loadVersionMap(libraryToModuleVersions)
		if err != nil {
			return fmt.Errorf("load library version compatibility list: %w", err)
		}
	}

	m.clientCGToLibVersions, m.serverCGToLibVersions, err = loadVersionMap(codegameToLibraryVersions)
	if err != nil {
		return fmt.Errorf("load codegame version compatibility list: %w", err)
	}

	return nil
}

func loadVersionMap(jsonData json.RawMessage) (client, server map[string]string, err error) {
	type versionsObj struct {
		Client json.RawMessage `json:"client"`
		Server json.RawMessage `json:"server"`
	}

	versions, err := loadJSONObjectInlineOrLocalOrRemote[versionsObj](jsonData)
	if err != nil {
		return nil, nil, err
	}

	if versions.Client != nil {
		client, err = loadJSONObjectInlineOrLocalOrRemote[map[string]string](versions.Client)
		if err != nil {
			return nil, nil, err
		}
	}

	if versions.Server != nil {
		server, err = loadJSONObjectInlineOrLocalOrRemote[map[string]string](versions.Server)
		if err != nil {
			return nil, nil, err
		}
	}

	return client, server, nil
}

func loadJSONObjectInlineOrLocalOrRemote[T any](jsonData json.RawMessage) (T, error) {
	file := io.NopCloser(bytes.NewBuffer(jsonData))

	var object T

	var uri string
	err := json.Unmarshal(jsonData, &uri)
	if err == nil {
		if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
			file, err = request.FetchFile(uri, 24*time.Hour, false)
			if err != nil {
				return object, err
			}
		} else {
			file, err = os.Open(uri)
			if err != nil {
				return object, err
			}
		}
	}

	err = json.NewDecoder(file).Decode(&object)
	if err != nil {
		return object, err
	}
	return object, nil
}

func (m *Module) loadInstalledVersions() error {
	var err error
	if m.provider.Name() == "local" {
		err = m.loadLocalModules()
	} else {
		m.installedExecutables = installedBinaries(m.Lang)
	}
	return err
}

func loadModules() error {
	file, err := os.Open(filepath.Join(config.ConfigDir(), "lang_modules.json"))
	if err != nil {
		return fmt.Errorf("open language modules config file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&rawModules)
	if err != nil {
		return fmt.Errorf("decode language modules config file: %w", err)
	}
	return nil
}

type AvailableLanguage struct {
	DisplayName    string
	SupportsClient bool
	SupportsServer bool
}

var availableLanguages map[string]AvailableLanguage

func AvailableLanguages() map[string]AvailableLanguage { // name -> display name
	if rawModules == nil {
		err := loadModules()
		if err != nil {
			feedback.Error(FeedbackPkg, "Failed to load available languages: %s", err)
		}
	}
	if availableLanguages != nil {
		return availableLanguages
	}
	availableLanguages = make(map[string]AvailableLanguage, len(rawModules))

	for n, m := range rawModules {
		client, server, err := loadVersionMap(m.LibraryToModuleVersions)
		if err != nil {
			feedback.Error(FeedbackPkg, "Failed to load supported project types of %s module: %s", n, err)
			continue
		}
		availableLanguages[n] = AvailableLanguage{
			DisplayName:    m.DisplayName,
			SupportsClient: len(client) > 0,
			SupportsServer: len(server) > 0,
		}
	}
	return availableLanguages
}
