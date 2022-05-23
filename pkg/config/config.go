package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func GetRepositoryAndHomepageFromEnv() (string, string) {
	serverUrl := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	if serverUrl == "" || repo == "" {
		return "", ""
	}
	return fmt.Sprintf("github:%s", repo), fmt.Sprintf("%s/%s", serverUrl, repo)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func SetFlags(cmd *cobra.Command) {
	repo, homepage := GetRepositoryAndHomepageFromEnv()
	_, packageName, _ := strings.Cut(repo, "/")
	cmd.PersistentFlags().StringP("input-path", "i", "", "input path that contains the binary files [uses ./bin or ./dist as default]")
	cmd.PersistentFlags().StringP("output-path", "o", "./generated-packages", "output directory")
	cmd.PersistentFlags().StringP("name", "n", packageName, "name of the binary (e.g my-cool-cli)")
	cmd.PersistentFlags().StringP("package-name-prefix", "p", "", "package name prefix for all created packages (e.g. @my-org/)")
	cmd.PersistentFlags().StringP("package-version", "r", "", "version of the created packages")
	cmd.PersistentFlags().String("package-name", "", "package name [defaults to the name of the binary] (e.g. my-cool-cli)")
	cmd.PersistentFlags().String("license", "", "package SPDX license (e.g. MIT)")
	cmd.PersistentFlags().String("homepage", homepage, "package homepage")
	cmd.PersistentFlags().String("description", "", "package description")
	cmd.PersistentFlags().String("repository", repo, "package repository")
	cmd.PersistentFlags().String("publish-registry", "https://registry.npmjs.org/", "npm registry endpoint")
	cmd.PersistentFlags().Bool("publish", false, "run npm publish for all packages")
	cmd.PersistentFlags().Bool("no-prefix-for-main-package", false, "ignore the configured package name prefix for the main package")
	cmd.PersistentFlags().SortFlags = true

	must(viper.BindPFlag("inputPath", cmd.PersistentFlags().Lookup("input-path")))
	must(viper.BindPFlag("outputPath", cmd.PersistentFlags().Lookup("output-path")))
	must(viper.BindPFlag("name", cmd.PersistentFlags().Lookup("name")))
	must(viper.BindPFlag("packageNamePrefix", cmd.PersistentFlags().Lookup("package-name-prefix")))
	must(viper.BindPFlag("packageName", cmd.PersistentFlags().Lookup("package-name")))
	must(viper.BindPFlag("license", cmd.PersistentFlags().Lookup("license")))
	must(viper.BindPFlag("homepage", cmd.PersistentFlags().Lookup("homepage")))
	must(viper.BindPFlag("description", cmd.PersistentFlags().Lookup("description")))
	must(viper.BindPFlag("repository", cmd.PersistentFlags().Lookup("repository")))
	must(viper.BindPFlag("publishRegistry", cmd.PersistentFlags().Lookup("publish-registry")))
	must(viper.BindPFlag("publish", cmd.PersistentFlags().Lookup("publish")))
	must(viper.BindPFlag("noPrefixForMainPackage", cmd.PersistentFlags().Lookup("no-prefix-for-main-package")))
}

func NewConfig(cmd *cobra.Command) *Config {
	packageVersion, err := cmd.Flags().GetString("package-version")
	must(err)
	c := &Config{
		InputBinDirPath:        viper.GetString("inputPath"),
		TryDefaultInputPaths:   !viper.IsSet("inputPath"),
		OutputDirPath:          viper.GetString("outputPath"),
		BinName:                viper.GetString("name"),
		PackageNamePrefix:      viper.GetString("packageNamePrefix"),
		PackageVersion:         packageVersion,
		PackageName:            viper.GetString("packageName"),
		License:                viper.GetString("license"),
		Homepage:               viper.GetString("homepage"),
		Description:            viper.GetString("description"),
		Repository:             viper.GetString("repository"),
		PublishRegistry:        viper.GetString("publishRegistry"),
		Publish:                viper.GetBool("publish"),
		NoPrefixForMainPackage: viper.GetBool("noPrefixForMainPackage"),
	}
	if c.PackageName == "" {
		c.PackageName = c.BinName
	}
	return c
}

func InitConfig() error {
	viper.AddConfigPath(".")
	viper.SetConfigName(".npm-binary-releaser.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return nil
}
