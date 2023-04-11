package components

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"

	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/versions"
)

var (
	componentBinPath   = filepath.Join(xdg.DataHome, "codegame", "components")
	ErrVersionNotFound = errors.New("version not found")
)

func findLatestCompatibleVersionSupportedByComponent(componentName string, version versions.Version) (component, supported versions.Version, err error) {
	versionMap, err := request.FetchJSON[map[string]versions.Version](fmt.Sprintf("https://raw.githubusercontent.com/code-game-project/%s/main/versions.json", componentName), 24*time.Hour)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch version map: %w", err)
	}
	for sup, comp := range versionMap {
		s, err := versions.Parse(sup)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid version map: %w", err)
		}
		if versions.Compare(supported, s) == 1 && s.IsCompatible(version) {
			supported = s
			component = comp
		}
	}
	if supported == nil {
		return nil, nil, ErrVersionNotFound
	}
	return component, supported, nil
}

func findLatestCompatibleVersionSupportedByComponentInOverrides(componentName string, version versions.Version) (supported versions.Version, binPath string, err error) {
	overrides := loadOverrides(componentName)
	for sup, path := range overrides {
		v, err := versions.Parse(sup)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load %s overrides: %w", componentName, err)
		}
		if versions.Compare(supported, v) == 1 && v.IsCompatible(version) {
			supported = v
			binPath = path
			if versions.Compare(version, v) == 0 {
				return v, binPath, nil
			}
		}
	}
	if supported == nil {
		return nil, "", ErrVersionNotFound
	}
	return supported, binPath, nil
}

func install(componentName string, version versions.Version) (string, error) {
	dirName := filepath.Join(componentBinPath, componentName)
	err := os.MkdirAll(dirName, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create component binary directory: %w", err)
	}

	tag, err := findGitHubTagByVersion(componentName, version)
	if err != nil {
		return "", fmt.Errorf("failed to find compatible tag: %w", err)
	}

	binPath := filepath.Join(dirName, strings.ReplaceAll(strings.TrimPrefix(tag, "v"), ".", "-"))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if _, err = os.Stat(binPath); err == nil {
		return binPath, nil
	}

	downloadFileName := fmt.Sprintf("%s-%s-%s.tar.gz", componentName, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		downloadFileName = fmt.Sprintf("%s-%s-%s.zip", componentName, runtime.GOOS, runtime.GOARCH)
	}

	file, err := request.FetchFile(fmt.Sprintf("https://github.com/code-game-project/%s/releases/download/%s/%s", componentName, tag, downloadFileName), 0)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", componentName, err)
	}
	defer file.Close()

	if runtime.GOOS == "windows" {
		err = unzipFile(file, componentName+".exe", binPath)
	} else {
		err = untargzFile(file, componentName, binPath)
	}
	if err != nil {
		return "", fmt.Errorf("failed to uncompress %s: %w", componentName, err)
	}
	return binPath, nil
}

func findGitHubTagByVersion(componentName string, version versions.Version) (string, error) {
	type response []struct {
		Name string `json:"name"`
	}
	res, err := request.FetchJSON[response](fmt.Sprintf("https://api.github.com/repos/code-game-project/%s/tags", componentName), 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to find GitHub tag by version: %w", err)
	}
	for _, tag := range res {
		if strings.HasPrefix(tag.Name, "v"+version.String()) {
			return tag.Name, nil
		}
	}
	return "", ErrVersionNotFound
}

// untargzFile first decompresses source with gzip, then extracts the file with fileName into outputFileName.
func untargzFile(source io.Reader, fileName, outputFileName string) error {
	archive, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer archive.Close()

	tarReader := tar.NewReader(archive)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()
		if !info.IsDir() && info.Name() == fileName {
			file, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, tarReader)
			return err
		}
	}

	return errors.New("file not found")
}

// unzipFile first decompresses source with gzip, then extracts the file with fileName into outputFileName.
func unzipFile(source io.Reader, fileName, outputFileName string) error {
	data, err := ioutil.ReadAll(source)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		if !f.FileInfo().IsDir() && f.FileInfo().Name() == fileName {
			out, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			in, err := f.Open()
			if err != nil {
				return err
			}
			defer in.Close()

			_, err = io.Copy(out, in)
			return err
		}
	}

	return errors.New("file not found")
}
