package templates

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"

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

type RepositoryConfig struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func NewRepositoryConfig(repository string) *RepositoryConfig {
	url := normalizeRepositoryURL(repository)
	if url == "" {
		return nil
	}
	return &RepositoryConfig{
		Type: "git",
		URL:  url,
	}
}

var (
	hostedRepositoryShortcutPattern = regexp.MustCompile(`^(github|gitlab|bitbucket):([^/\s]+/.+)$`)
	hostedRepositoryHTTPSPattern    = regexp.MustCompile(`^(?:git\+)?https?://(github\.com|gitlab\.com|bitbucket\.org)/(.+)$`)
	hostedRepositoryGitPattern      = regexp.MustCompile(`^git://(github\.com|gitlab\.com|bitbucket\.org)/(.+)$`)
	sshRepositoryPattern            = regexp.MustCompile(`^git@([^:]+):(.+)$`)
	sshURLRepositoryPattern         = regexp.MustCompile(`^(?:git\+)?ssh://git@([^/]+)/(.+)$`)
)

func normalizeRepositoryURL(repository string) string {
	repository = strings.TrimSpace(repository)
	if repository == "" {
		return ""
	}

	if matches := hostedRepositoryShortcutPattern.FindStringSubmatch(repository); matches != nil {
		return "git+https://" + shortcutHost(matches[1]) + "/" + withGitSuffix(matches[2])
	}
	if matches := hostedRepositoryHTTPSPattern.FindStringSubmatch(repository); matches != nil {
		return "git+https://" + matches[1] + "/" + withGitSuffix(matches[2])
	}
	if matches := hostedRepositoryGitPattern.FindStringSubmatch(repository); matches != nil {
		return "git://" + matches[1] + "/" + withGitSuffix(matches[2])
	}
	if matches := sshRepositoryPattern.FindStringSubmatch(repository); matches != nil {
		return "git+ssh://git@" + matches[1] + "/" + withGitSuffix(matches[2])
	}
	if matches := sshURLRepositoryPattern.FindStringSubmatch(repository); matches != nil {
		return "git+ssh://git@" + matches[1] + "/" + withGitSuffix(matches[2])
	}

	return repository
}

func shortcutHost(shortcut string) string {
	switch shortcut {
	case "bitbucket":
		return "bitbucket.org"
	case "gitlab":
		return "gitlab.com"
	default:
		return "github.com"
	}
}

func withGitSuffix(repositoryPath string) string {
	repositoryPath = strings.TrimRight(repositoryPath, "/")
	if strings.HasSuffix(repositoryPath, ".git") || strings.Contains(repositoryPath, ".git#") {
		return repositoryPath
	}
	if path, ref, found := strings.Cut(repositoryPath, "#"); found {
		return path + ".git#" + ref
	}
	return repositoryPath + ".git"
}

type BinPackageJson struct {
	Name            string        `json:"name"`
	Version         string        `json:"version"`
	Description     string        `json:"description,omitempty"`
	License         string        `json:"license,omitempty"`
	Homepage        string        `json:"homepage,omitempty"`
	Repository      string        `json:"-"`
	OS              []string      `json:"os"`
	CPU             []string      `json:"cpu"`
	Main            string        `json:"main"`
	Files           []string      `json:"files"`
	PreferUnplugged bool          `json:"preferUnplugged"`
	PublishConfig   PublishConfig `json:"publishConfig"`
}

func (pkg BinPackageJson) MarshalJSON() ([]byte, error) {
	type Alias BinPackageJson
	return json.Marshal(&struct {
		Alias
		Repository *RepositoryConfig `json:"repository,omitempty"`
	}{
		Alias:      Alias(pkg),
		Repository: NewRepositoryConfig(pkg.Repository),
	})
}

func packageFiles(files []string, includeReadme bool) []string {
	if includeReadme {
		files = append(files, "README.md")
	}
	return files
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
	Repository           string            `json:"-"`
	BinPkgPrefix         string            `json:"binPkgPrefix,omitempty"`
	Bin                  map[string]string `json:"bin"`
	Files                []string          `json:"files"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	PublishConfig        PublishConfig     `json:"publishConfig"`
}

func (pkg MainPackageJson) MarshalJSON() ([]byte, error) {
	type Alias MainPackageJson
	return json.Marshal(&struct {
		Alias
		Repository *RepositoryConfig `json:"repository,omitempty"`
	}{
		Alias:      Alias(pkg),
		Repository: NewRepositoryConfig(pkg.Repository),
	})
}

func NewMainPackageJson(cfg *config.Config, packageName string, optDeps map[string]string, includeReadme bool) MainPackageJson {
	return MainPackageJson{
		Name:        packageName,
		Version:     cfg.PackageVersion,
		Description: cfg.Description,
		License:     cfg.License,
		Homepage:    cfg.Homepage,
		Repository:  cfg.Repository,
		Bin: map[string]string{
			cfg.BinName: "run.js",
		},
		Files:                packageFiles([]string{"run.js"}, includeReadme),
		OptionalDependencies: optDeps,
		PublishConfig:        NewPublishConfig(cfg),
	}
}
