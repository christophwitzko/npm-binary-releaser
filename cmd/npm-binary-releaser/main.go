package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/christophwitzko/npm-binary-releaser/pkg/helper"
	"github.com/christophwitzko/npm-binary-releaser/pkg/templates"
	"github.com/spf13/cobra"
)

var VERSION string

type Config struct {
	BinDirPath        string
	BinName           string
	PackageNamePrefix string
	PackageVersion    string
	PackagesOutPath   string
	Publish           bool
}

var logger = log.New(os.Stderr, "[npm-binary-releaser]: ", 0)

func run(c Config) error {
	if c.PackageVersion == "" {
		return fmt.Errorf("package version is missing")
	}
	if c.BinName == "" {
		return fmt.Errorf("name is missing")
	}
	logger.Printf("creating release %s for %s", c.PackageVersion, c.BinName)
	logger.Printf("creating output directory: %s", c.PackagesOutPath)
	if err := helper.EnsureOutputDirectory(c.PackagesOutPath); err != nil {
		return err
	}

	logger.Printf("reading binary files from: %s", c.BinDirPath)
	files, err := os.ReadDir(c.BinDirPath)
	if err != nil {
		return err
	}

	foundFiles := make([]*helper.BinFile, 0, len(files))
	for _, file := range files {
		logger.Printf("checking file %s", file.Name())
		platform, arch := helper.ExtractOsAndArchFromFileName(file.Name())
		if platform == "" || arch == "" {
			logger.Printf("no os/arch found for %s", file.Name())
			continue
		}
		foundFiles = append(foundFiles, &helper.BinFile{
			Platform: platform,
			Arch:     arch,
			Path:     path.Join(c.BinDirPath, file.Name()),
			FileName: file.Name(),
		})
	}

	allPackageDirs := make([]string, 0, len(files)+1)
	optionalDependencies := make(map[string]string)
	for _, file := range foundFiles {
		packageName := fmt.Sprintf("%s-%s-%s", c.BinName, file.Platform, file.Arch)
		fullPackageName := fmt.Sprintf("%s%s", c.PackageNamePrefix, packageName)
		pkgDir := path.Join(c.PackagesOutPath, packageName)

		logger.Printf("[%s] creating package at %s", fullPackageName, pkgDir)
		if err := os.Mkdir(pkgDir, 0755); err != nil {
			return err
		}

		binFileName := packageName
		if file.Platform == "win32" {
			binFileName += ".exe"
		}

		logger.Printf("[%s] creating package.json", fullPackageName)
		pjsTemplate := templates.NewBinPackageJson(fullPackageName, c.PackageVersion, file.Platform, file.Arch, binFileName)
		pjsData, err := json.MarshalIndent(pjsTemplate, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(pkgDir, "package.json"), pjsData, 0644); err != nil {
			return err
		}

		logger.Printf("[%s] copying binary file to %s", fullPackageName, binFileName)
		if err := helper.CopyFile(file.Path, path.Join(pkgDir, binFileName)); err != nil {
			return err
		}
		optionalDependencies[fullPackageName] = c.PackageVersion
		allPackageDirs = append(allPackageDirs, pkgDir)
	}

	mainPackageDir := path.Join(c.PackagesOutPath, c.BinName)
	mainPackageName := fmt.Sprintf("%s%s", c.PackageNamePrefix, c.BinName)
	logger.Printf("[%s] creating main package at %s", mainPackageName, mainPackageDir)

	// create package folder
	if err := os.Mkdir(mainPackageDir, 0755); err != nil {
		return err
	}

	logger.Printf("[%s] creating package.json", mainPackageName)
	pjsTemplate := templates.NewMainPackageJson(mainPackageName, c.PackageVersion, c.BinName, optionalDependencies)
	pjsData, err := json.MarshalIndent(pjsTemplate, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(mainPackageDir, "package.json"), pjsData, 0644); err != nil {
		return err
	}

	// create run.js
	logger.Printf("[%s] creating run.js", mainPackageName)
	if err := os.WriteFile(path.Join(mainPackageDir, "run.js"), templates.RunJs, 0755); err != nil {
		return err
	}

	if !c.Publish {
		logger.Printf("skipping npm publish step")
		return nil
	}
	allPackageDirs = append(allPackageDirs, mainPackageDir)
	for _, pDir := range allPackageDirs {
		logger.Printf("running npm publish in %s", pDir)
		cmd := exec.Command("npm", "publish", pDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	logger.Println("done.")
	return nil
}

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

func main() {
	cmd := &cobra.Command{
		Use:     "npm-binary-releaser",
		Short:   "npm-binary-releaser - release binaries to npm",
		Run:     cliHandler,
		Version: VERSION,
	}

	cmd.Flags().StringP("input-path", "i", "./bin", "input path that contains the binary files")
	cmd.Flags().StringP("output-path", "o", "./generated-packages", "output directory")
	cmd.Flags().StringP("name", "n", "", "name of the binary and package (e.g my-cool-cli)")
	cmd.Flags().StringP("package-name-prefix", "p", "", "package name prefix for all created packages (e.g. @my-org/)")
	cmd.Flags().StringP("package-version", "r", "", "version of the created packages")
	cmd.Flags().Bool("publish", false, "run npm publish for all packages")
	cmd.Flags().SortFlags = true

	if err := cmd.Execute(); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}

func cliHandler(cmd *cobra.Command, args []string) {
	config := Config{
		BinDirPath:        mustGetString(cmd, "input-path"),
		BinName:           mustGetString(cmd, "name"),
		PackageNamePrefix: mustGetString(cmd, "package-name-prefix"),
		PackageVersion:    mustGetString(cmd, "package-version"),
		PackagesOutPath:   mustGetString(cmd, "output-path"),
		Publish:           mustGetBool(cmd, "publish"),
	}

	if err := run(config); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return
	}
}
