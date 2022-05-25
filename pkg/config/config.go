package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BinName                string `yaml:"name"`
	InputBinDirPath        string `yaml:"inputPath,omitempty"`
	TryDefaultInputPaths   bool   `yaml:"-"`
	PackageName            string `yaml:"packageName"`
	Description            string `yaml:"description"`
	License                string `yaml:"license"`
	Homepage               string `yaml:"homepage"`
	Repository             string `yaml:"repository"`
	PackageNamePrefix      string `yaml:"packageNamePrefix"`
	NoPrefixForMainPackage bool   `yaml:"noPrefixForMainPackage"`
	PackageVersion         string `yaml:"-"`
	OutputDirPath          string `yaml:"outputPath"`
	PublishRegistry        string `yaml:"publishRegistry"`
	Publish                bool   `yaml:"publish"`
}

var defaultInputDirPaths = []string{"./bin", "./dist"}

func (c *Config) Validate() error {
	if c.PackageName == "" {
		c.PackageName = c.BinName
	}
	if c.PackageVersion == "" {
		return fmt.Errorf("package version is missing")
	}
	if c.BinName == "" {
		return fmt.Errorf("name is missing")
	}
	if c.TryDefaultInputPaths {
		c.InputBinDirPath = ""
		for _, dirPath := range defaultInputDirPaths {
			_, err := os.Stat(dirPath)
			if err == nil {
				c.InputBinDirPath = dirPath
				break
			}
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
	} else {
		if _, err := os.Stat(c.InputBinDirPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.InputBinDirPath = ""
			} else {
				return err
			}
		}
	}
	if c.InputBinDirPath == "" {
		return fmt.Errorf("input path is missing or does not exist")
	}
	return nil
}

type EnvInfo struct {
	Repository  string
	Homepage    string
	PackageName string
}

func GetRepositoryAndHomepageFromEnv() EnvInfo {
	serverUrl := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	if serverUrl == "" || repo == "" {
		return EnvInfo{}
	}
	_, packageName, _ := strings.Cut(repo, "/")
	return EnvInfo{
		Repository:  fmt.Sprintf("github:%s", repo),
		Homepage:    fmt.Sprintf("%s/%s", serverUrl, repo),
		PackageName: packageName,
	}
}
