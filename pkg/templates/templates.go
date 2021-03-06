package templates

import (
	_ "embed"

	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
)

//go:embed run.js
var RunJs []byte

type PublishConfig struct {
	Registry string `json:"registry"`
	Access   string `json:"access"`
}

func NewPublishConfig(cfg *config.Config) PublishConfig {
	return PublishConfig{
		Registry: cfg.PublishRegistry,
		Access:   "public",
	}
}

type BinPackageJson struct {
	Name            string        `json:"name"`
	Version         string        `json:"version"`
	Description     string        `json:"description,omitempty"`
	License         string        `json:"license,omitempty"`
	Homepage        string        `json:"homepage,omitempty"`
	Repository      string        `json:"repository,omitempty"`
	OS              []string      `json:"os"`
	CPU             []string      `json:"cpu"`
	Main            string        `json:"main"`
	Files           []string      `json:"files"`
	PreferUnplugged bool          `json:"preferUnplugged"`
	PublishConfig   PublishConfig `json:"publishConfig"`
}

func NewBinPackageJson(cfg *config.Config, packageName, platform, arch, file string) BinPackageJson {
	return BinPackageJson{
		Name:            packageName,
		Version:         cfg.PackageVersion,
		Description:     cfg.Description,
		License:         cfg.License,
		Homepage:        cfg.Homepage,
		Repository:      cfg.Repository,
		OS:              []string{platform},
		CPU:             []string{arch},
		Main:            file,
		Files:           []string{file},
		PublishConfig:   NewPublishConfig(cfg),
		PreferUnplugged: true,
	}
}

type MainPackageJson struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Description          string            `json:"description,omitempty"`
	License              string            `json:"license,omitempty"`
	Homepage             string            `json:"homepage,omitempty"`
	Repository           string            `json:"repository,omitempty"`
	BinPkgPrefix         string            `json:"binPkgPrefix,omitempty"`
	Bin                  map[string]string `json:"bin"`
	Files                []string          `json:"files"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	PublishConfig        PublishConfig     `json:"publishConfig"`
}

func NewMainPackageJson(cfg *config.Config, packageName string, optDeps map[string]string) MainPackageJson {
	return MainPackageJson{
		Name:        packageName,
		Version:     cfg.PackageVersion,
		Description: cfg.Description,
		License:     cfg.License,
		Homepage:    cfg.Homepage,
		Repository:  cfg.Repository,
		Bin: map[string]string{
			cfg.BinName: "./run.js",
		},
		Files: []string{
			"run.js",
		},
		OptionalDependencies: optDeps,
		PublishConfig:        NewPublishConfig(cfg),
	}
}
