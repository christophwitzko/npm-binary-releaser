package main

import (
	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func SetFlags(cmd *cobra.Command) {
	envInfo := config.GetRepositoryAndHomepageFromEnv()
	cmd.PersistentFlags().StringP("input-path", "i", "", "input path that contains the binary files [uses ./bin or ./dist as default]")
	cmd.PersistentFlags().StringP("output-path", "o", config.DefaultOutputDirPath, "output directory")
	cmd.PersistentFlags().StringP("name", "n", envInfo.PackageName, "name of the binary (e.g my-cool-cli)")
	cmd.PersistentFlags().StringP("package-name-prefix", "p", "", "package name prefix for all created packages (e.g. @my-org/)")
	cmd.PersistentFlags().StringP("package-version", "r", "", "version of the created packages")
	cmd.PersistentFlags().String("package-name", "", "package name [defaults to the name of the binary] (e.g. my-cool-cli)")
	cmd.PersistentFlags().String("license", "", "package SPDX license (e.g. MIT)")
	cmd.PersistentFlags().String("homepage", envInfo.Homepage, "package homepage")
	cmd.PersistentFlags().String("description", "", "package description")
	cmd.PersistentFlags().String("repository", envInfo.Repository, "package repository")
	cmd.PersistentFlags().String("publish-registry", config.DefaultPublishRegistry, "npm registry endpoint")
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

func NewConfig(cmd *cobra.Command) *config.Config {
	packageVersion, err := cmd.Flags().GetString("package-version")
	must(err)
	c := &config.Config{
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
