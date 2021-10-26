package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
	"github.com/christophwitzko/npm-binary-releaser/pkg/helper"
	"github.com/christophwitzko/npm-binary-releaser/pkg/templates"
	"github.com/spf13/cobra"
)

var VERSION string

var logger = log.New(os.Stderr, "[npm-binary-releaser]: ", 0)

func main() {
	cmd := &cobra.Command{
		Use:     "npm-binary-releaser",
		Short:   "npm-binary-releaser - release binaries to npm",
		Run:     cliHandler,
		Version: VERSION,
	}

	config.InitConfig(cmd)

	if err := cmd.Execute(); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}

func cliHandler(cmd *cobra.Command, args []string) {
	if err := run(config.NewConfig(cmd)); err != nil {
		logger.Println(err)
		os.Exit(1)
		return
	}
}

func run(c *config.Config) error {
	if c.PackageVersion == "" {
		return fmt.Errorf("package version is missing")
	}
	if c.BinName == "" {
		return fmt.Errorf("name is missing")
	}
	if c.PackageName == "" {
		c.PackageName = c.BinName
	}
	logger.Printf("creating release %s for %s (%s)", c.PackageVersion, c.PackageName, c.BinName)
	logger.Printf("creating output directory: %s", c.OutputDirPath)
	if err := helper.EnsureOutputDirectory(c.OutputDirPath); err != nil {
		return err
	}

	logger.Printf("reading binary files from: %s", c.InputBinDirPath)
	files, err := os.ReadDir(c.InputBinDirPath)
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
			Path:     path.Join(c.InputBinDirPath, file.Name()),
			FileName: file.Name(),
		})
	}

	if len(foundFiles) == 0 {
		return fmt.Errorf("no binary files found at %s", c.InputBinDirPath)
	}

	allPackageDirs := make([]string, 0, len(files)+1)
	optionalDependencies := make(map[string]string)
	for _, file := range foundFiles {
		packageName := fmt.Sprintf("%s-%s-%s", c.PackageName, file.Platform, file.Arch)
		fullPackageName := fmt.Sprintf("%s%s", c.PackageNamePrefix, packageName)
		pkgDir := path.Join(c.OutputDirPath, packageName)

		logger.Printf("[%s] creating package at %s", fullPackageName, pkgDir)
		if err := os.Mkdir(pkgDir, 0755); err != nil {
			return err
		}

		binFileName := packageName
		if file.Platform == "win32" {
			binFileName += ".exe"
		}

		logger.Printf("[%s] creating package.json", fullPackageName)
		pjsTemplate := templates.NewBinPackageJson(c, fullPackageName, file.Platform, file.Arch, binFileName)
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

	mainPackageDir := path.Join(c.OutputDirPath, c.PackageName)
	mainPackageName := fmt.Sprintf("%s%s", c.PackageNamePrefix, c.PackageName)
	if c.NoPrefixForMainPackage && c.PackageNamePrefix != "" {
		mainPackageName = c.PackageName
	}
	logger.Printf("[%s] creating main package at %s", mainPackageName, mainPackageDir)

	// create package folder
	if err := os.Mkdir(mainPackageDir, 0755); err != nil {
		return err
	}

	logger.Printf("[%s] creating package.json", mainPackageName)
	pjsTemplate := templates.NewMainPackageJson(c, mainPackageName, optionalDependencies)
	if c.NoPrefixForMainPackage && c.PackageNamePrefix != "" {
		pjsTemplate.BinPkgPrefix = c.PackageNamePrefix
	}
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
