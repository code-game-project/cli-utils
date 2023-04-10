package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"

	"github.com/code-game-project/cli-utils/versions"
)

var moduleBinPath = filepath.Join(xdg.DataHome, "codegame", "modules")

func (m *Module) install(moduleVersion versions.Version) (string, error) {
	dirName := filepath.Join(moduleBinPath, m.Lang)
	err := os.MkdirAll(dirName, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create module binary directory: %w", err)
	}

	version, err := m.provider.FindExactVersion(m.providerVars, moduleVersion)
	if err != nil {
		return "", fmt.Errorf("failed to determine exact module version: %w", err)
	}

	var binPath string
	if p, ok := m.installedExecutables[version.String()]; ok {
		binPath = p
	} else {
		tempBinPath := filepath.Join(dirName, strings.ReplaceAll(moduleVersion.String(), ".", "-")+".temp")
		file, err := os.OpenFile(tempBinPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return "", fmt.Errorf("failed to create module binary file: %w", err)
		}
		defer func() {
			os.Remove(tempBinPath)
		}()

		err = m.provider.DownloadModuleBinary(file, m.providerVars, version)
		file.Close()
		if err != nil {
			return "", fmt.Errorf("failed to download module binary: %s", err)
		}

		binPath = filepath.Join(dirName, strings.ReplaceAll(version.String(), ".", "-"))
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}

		err = os.Rename(tempBinPath, binPath)
		if err != nil {
			return "", fmt.Errorf("failed to create module binary file: %w", err)
		}
	}

	return binPath, nil
}

func (m *Module) findCompatibleModuleVersion(projectType ProjectType, libraryVersion versions.Version) (versions.Version, error) {
	var versionMap map[string]string
	switch projectType {
	case ProjectType_CLIENT:
		versionMap = m.clientLibToModVersions
	case ProjectType_SERVER:
		versionMap = m.serverLibToModVersions
	}
	if versionMap == nil {
		return nil, ErrUnsupportedProjectType
	}

	v, err := versions.FindCompatibleInMap(libraryVersion, versionMap)
	if err != nil {
		return nil, fmt.Errorf("failed to find compatible module version: %w", err)
	}
	return v, nil
}

func (m *Module) findLatestModuleVersion(projectType ProjectType) (versions.Version, error) {
	var versionMap map[string]string
	switch projectType {
	case ProjectType_CLIENT:
		versionMap = m.clientLibToModVersions
	case ProjectType_SERVER:
		versionMap = m.serverLibToModVersions
	}
	if versionMap == nil {
		return nil, ErrUnsupportedProjectType
	}

	var latestLibVersion versions.Version
	for libVersion := range versionMap {
		libVer, err := versions.Parse(libVersion)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", libVersion, versions.ErrInvalidVersion)
		}
		if versions.Compare(latestLibVersion, libVer) == 1 {
			latestLibVersion = libVer
		}
	}

	if latestLibVersion == nil {
		return nil, ErrVersionNotFound
	}

	modVersionStr := versionMap[latestLibVersion.String()]
	modVersion, err := versions.Parse(modVersionStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", modVersionStr, err)
	}

	return modVersion, nil
}

func (m *Module) findLibraryVersionByCGVersion(projectType ProjectType, cgVersion versions.Version) (versions.Version, error) {
	var versionMap map[string]string
	switch projectType {
	case ProjectType_CLIENT:
		versionMap = m.clientCGToLibVersions
	case ProjectType_SERVER:
		versionMap = m.serverCGToLibVersions
	}
	if versionMap == nil {
		return nil, ErrUnsupportedProjectType
	}

	v, err := versions.FindCompatibleInMap(cgVersion, versionMap)
	if err != nil {
		return nil, fmt.Errorf("failed to find compatible library version: %w", err)
	}
	return v, nil
}

func installedBinaries(lang string) map[string]string { // version -> path
	entries, err := os.ReadDir(filepath.Join(moduleBinPath, lang))
	if err != nil {
		return make(map[string]string)
	}

	binaries := make(map[string]string, len(entries))
	for _, e := range entries {
		// TODO: validate file name (must be: major-minor-patch)
		if !e.IsDir() {
			binaries[strings.ReplaceAll(strings.TrimSuffix(e.Name(), ".exe"), "-", ".")] = filepath.Join(moduleBinPath, lang, e.Name())
		}
	}
	return binaries
}
