package helper

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

type BinFile struct {
	Platform string
	Arch     string
	Path     string
	FileName string
}

func CopyFile(from, to string) error {
	info, err := os.Stat(from)
	if err != nil {
		return err
	}

	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE, info.Mode())
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	return err
}

func toNodePlatform(platform string) string {
	switch platform {
	case "solaris":
		return "sunos"
	case "windows":
		return "win32"
	}
	return platform
}

func toNodeArch(arch string) string {
	switch arch {
	case "386":
		return "ia32"
	case "amd64":
		return "x64"
	}
	return arch
}

var osArchRegexp = regexp.MustCompile("(?i)(android|darwin|dragonfly|freebsd|linux|nacl|netbsd|openbsd|plan9|solaris|windows)(_|-)(i?386|amd64p32|amd64|arm64|arm|mips64le|mips64|mipsle|mips|ppc64le|ppc64|s390x|x86_64)")

func ExtractOsAndArchFromFileName(fileName string) (string, string) {
	osArch := osArchRegexp.FindAllStringSubmatch(fileName, -1)
	if len(osArch) < 1 || len(osArch[0]) < 4 {
		return "", ""
	}
	return toNodePlatform(strings.ToLower(osArch[0][1])), toNodeArch(strings.ToLower(osArch[0][3]))
}

func EnsureOutputDirectory(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		// path exists => remove
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}
	return nil
}

func FindFirstExecutableFileInDir(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		return path.Join(dir, file.Name()), nil
	}
	return "", fmt.Errorf("no executable file was found")
}
