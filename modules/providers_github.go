package modules

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"time"

	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/request"
	"github.com/code-game-project/cli-utils/versions"
)

var ErrFileNotFound = errors.New("file not found")

type ProviderGithub struct{}

func (p *ProviderGithub) Name() string {
	return "github"
}

func (p *ProviderGithub) ValidateProviderVars(providerVars map[string]any) []string {
	var errs []string
	if _, ok := providerVars["owner"]; !ok {
		errs = append(errs, "missing 'owner' field")
	} else if _, ok := providerVars["owner"].(string); !ok {
		errs = append(errs, "value of 'owner' field must be a string")
	}
	if _, ok := providerVars["repository"]; !ok {
		errs = append(errs, "missing 'repository' field")
	} else if _, ok := providerVars["repository"].(string); !ok {
		errs = append(errs, "value of 'repository' field must be a string")
	}
	return errs
}

func (p *ProviderGithub) FindExactVersion(providerVars map[string]any, version versions.Version) (versions.Version, error) {
	tag, err := p.findTagByVersion(providerVars["owner"].(string), providerVars["repository"].(string), version)
	if err != nil {
		return nil, err
	}
	tagVersion, err := versions.Parse(tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag version: %w", err)
	}
	return tagVersion, nil
}

func (p *ProviderGithub) DownloadModuleBinary(target io.Writer, providerVars map[string]any, version versions.Version) error {
	downloadFileName := fmt.Sprintf("%s-%s-%s.tar.gz", providerVars["repository"], runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		downloadFileName = fmt.Sprintf("%s-%s-%s.zip", providerVars["repository"], runtime.GOOS, runtime.GOARCH)
	}

	feedback.InterceptProgress(request.FeedbackPkg, func(_ feedback.Package, _, _ string, current, total int64, unit cli.Unit) {
		feedback.Progress(FeedbackPkg, fmt.Sprintf("download %s", providerVars["repository"]), fmt.Sprintf("Downloading module %s", providerVars["repository"]), current, total, unit)
	})
	file, err := request.FetchFile(fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", providerVars["owner"], providerVars["repository"], fmt.Sprintf("v%s", version), downloadFileName), 0, true)
	defer feedback.UninterceptProgress(request.FeedbackPkg)
	if err != nil {
		return err
	}
	defer file.Close()

	fileName := providerVars["repository"].(string)

	if runtime.GOOS == "windows" {
		err = unzipFile(file, fileName+".exe", target)
	} else {
		err = untargzFile(file, fileName, target)
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *ProviderGithub) findTagByVersion(owner, repo string, version versions.Version) (string, error) {
	type response []struct {
		Name string `json:"name"`
	}
	res, err := request.FetchJSON[response](fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo), 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("find GitHub tag by version: %w", err)
	}
	for _, tag := range res {
		if strings.HasPrefix(tag.Name, "v"+version.String()) {
			return tag.Name, nil
		}
	}
	return "", ErrVersionNotFound
}

// untargzFile first decompresses source with gzip, then extracts the file with fileName into outputFileName.
func untargzFile(source io.Reader, fileName string, target io.Writer) error {
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
			_, err = io.Copy(target, tarReader)
			return err
		}
	}

	return ErrFileNotFound
}

// unzipFile first decompresses source with gzip, then extracts the file with fileName into outputFileName.
func unzipFile(source io.Reader, fileName string, target io.Writer) error {
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
			in, err := f.Open()
			if err != nil {
				return err
			}
			defer in.Close()

			_, err = io.Copy(target, in)
			return err
		}
	}

	return ErrFileNotFound
}
