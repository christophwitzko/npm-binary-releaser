package releaser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
	"github.com/christophwitzko/npm-binary-releaser/pkg/helper"
	"github.com/christophwitzko/npm-binary-releaser/pkg/templates"
)

func Run(c *config.Config, logger Logger) error {
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
		fPath := path.Join(c.InputBinDirPath, file.Name())
		if file.IsDir() {
			execPath, err := helper.FindFirstExecutableFileInDir(fPath)
			if err != nil {
				logger.Printf("could not find bin file in dir %s %v", fPath, err)
				continue
			}
			fPath = execPath
		}
		foundFiles = append(foundFiles, &helper.BinFile{
			Platform: platform,
			Arch:     arch,
			Path:     fPath,
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

	if os.Getenv("NPM_CONFIG_USERCONFIG") == "" {
		if _, err = os.Stat(".npmrc"); os.IsNotExist(err) {
			registryName := strings.TrimPrefix(c.PublishRegistry, "https://")
			logger.Printf("creating .npmrc for %s", registryName)
			npmRcData := fmt.Sprintf("//%s:_authToken=${NPM_TOKEN}\n", registryName)
			if err := os.WriteFile(".npmrc", []byte(npmRcData), 0644); err != nil {
				return err
			}
		}
	}

	allPackageDirs = append(allPackageDirs, mainPackageDir)
	for _, pDir := range allPackageDirs {
		publishDir, err := filepath.Abs(pDir)
		if err != nil {
			return err
		}
		logger.Printf("running npm publish in %s", publishDir)
		cmd := exec.Command("npm", "publish", publishDir)
		cmd.Stdout = prefixedWriter(logger, "publish")
		cmd.Stderr = prefixedWriter(logger, "publish")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	logger.Println("done.")
	return nil
}
