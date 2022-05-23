package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func mustGetString(cmd *cobra.Command, name string) string {
	res, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(err)
	}
	return res
}

func mustGetBool(cmd *cobra.Command, name string) bool {
	res, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(err)
	}
	return res
}

type Config struct {
	InputBinDirPath        string
	BinName                string
	PackageName            string
	PackageNamePrefix      string
	PackageVersion         string
	OutputDirPath          string
	Homepage               string
	License                string
	PublishRegistry        string
	Description            string
	Repository             string
	Publish                bool
	NoPrefixForMainPackage bool
}

func getRepositoryFromEnv() (string, string) {
	serverUrl := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	if serverUrl == "" || repo == "" {
		return "", ""
	}
	return fmt.Sprintf("github:%s", repo), fmt.Sprintf("%s/%s", serverUrl, repo)
}

func InitConfig(cmd *cobra.Command) {
	repo, homepage := getRepositoryFromEnv()
	cmd.Flags().StringP("input-path", "i", "./bin", "input path that contains the binary files")
	cmd.Flags().StringP("output-path", "o", "./generated-packages", "output directory")
	cmd.Flags().StringP("name", "n", "", "name of the binary (e.g my-cool-cli)")
	cmd.Flags().StringP("package-name-prefix", "p", "", "package name prefix for all created packages (e.g. @my-org/)")
	cmd.Flags().StringP("package-version", "r", "", "version of the created packages")
	cmd.Flags().String("package-name", "", "package name [defaults to 'name'] (e.g. my-cool-cli)")
	cmd.Flags().String("license", "", "package SPDX license (e.g. MIT)")
	cmd.Flags().String("homepage", homepage, "package homepage")
	cmd.Flags().String("description", "", "package description")
	cmd.Flags().String("repository", repo, "package repository")
	cmd.Flags().String("publish-registry", "https://registry.npmjs.org/", "npm registry endpoint")
	cmd.Flags().Bool("publish", false, "run npm publish for all packages")
	cmd.Flags().Bool("no-prefix-for-main-package", false, "ignore the configured package name prefix for the main package")
	cmd.Flags().SortFlags = true
}

func NewConfig(cmd *cobra.Command) *Config {
	return &Config{
		InputBinDirPath:        mustGetString(cmd, "input-path"),
		BinName:                mustGetString(cmd, "name"),
		PackageName:            mustGetString(cmd, "package-name"),
		PackageNamePrefix:      mustGetString(cmd, "package-name-prefix"),
		PackageVersion:         mustGetString(cmd, "package-version"),
		OutputDirPath:          mustGetString(cmd, "output-path"),
		Homepage:               mustGetString(cmd, "homepage"),
		License:                mustGetString(cmd, "license"),
		PublishRegistry:        mustGetString(cmd, "publish-registry"),
		Description:            mustGetString(cmd, "description"),
		Repository:             mustGetString(cmd, "repository"),
		Publish:                mustGetBool(cmd, "publish"),
		NoPrefixForMainPackage: mustGetBool(cmd, "no-prefix-for-main-package"),
	}
}
