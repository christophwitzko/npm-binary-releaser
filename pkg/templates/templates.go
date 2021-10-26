package templates

import _ "embed"

//go:embed run.js
var RunJs []byte

type PublishConfig struct {
	Registry string `json:"registry"`
	Access   string `json:"access"`
}

func NewDefaultPublishConfig() PublishConfig {
	return PublishConfig{
		Registry: "https://registry.npmjs.org/",
		Access:   "public",
	}
}

type BinPackageJson struct {
	Name          string        `json:"name"`
	Version       string        `json:"version"`
	OS            []string      `json:"os"`
	CPU           []string      `json:"cpu"`
	Main          string        `json:"main"`
	Files         []string      `json:"files"`
	PublishConfig PublishConfig `json:"publishConfig"`
}

func NewBinPackageJson(packageName, version, platform, arch, file string) BinPackageJson {
	return BinPackageJson{
		Name:          packageName,
		Version:       version,
		OS:            []string{platform},
		CPU:           []string{arch},
		Main:          file,
		Files:         []string{file},
		PublishConfig: NewDefaultPublishConfig(),
	}
}

type MainPackageJson struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Bin                  map[string]string `json:"bin"`
	Files                []string          `json:"files"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	PublishConfig        PublishConfig     `json:"publishConfig"`
}

func NewMainPackageJson(packageName, version, binName string, optDeps map[string]string) MainPackageJson {
	return MainPackageJson{
		Name:    packageName,
		Version: version,
		Bin: map[string]string{
			binName: "./run.js",
		},
		Files: []string{
			"run.js",
		},
		OptionalDependencies: optDeps,
		PublishConfig:        NewDefaultPublishConfig(),
	}
}
