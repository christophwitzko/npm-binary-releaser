package templates

import (
	"encoding/json"
	"testing"

	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
)

func TestNewMainPackageJsonUsesNPMNormalizedFields(t *testing.T) {
	pkg := NewMainPackageJson(&config.Config{
		BinName:         "interloom",
		PackageVersion:  "1.0.0",
		Repository:      "https://github.com/interloom/cli",
		PublishRegistry: config.DefaultPublishRegistry,
	}, "interloom", map[string]string{"interloom-linux-x64": "1.0.0"}, false)

	if got := pkg.Bin["interloom"]; got != "run.js" {
		t.Fatalf("bin target = %q, want %q", got, "run.js")
	}
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	var generated struct {
		Repository RepositoryConfig `json:"repository"`
	}
	if err := json.Unmarshal(data, &generated); err != nil {
		t.Fatal(err)
	}
	if got := generated.Repository.URL; got != "git+https://github.com/interloom/cli.git" {
		t.Fatalf("repository.url = %q, want %q", got, "git+https://github.com/interloom/cli.git")
	}
}

func TestNormalizeRepositoryURL(t *testing.T) {
	tests := map[string]string{
		"github:interloom/cli":                     "git+https://github.com/interloom/cli.git",
		"https://github.com/interloom/cli":         "git+https://github.com/interloom/cli.git",
		"git+https://github.com/interloom/cli.git": "git+https://github.com/interloom/cli.git",
		"git@github.com:interloom/cli.git":         "git+ssh://git@github.com/interloom/cli.git",
		"git+ssh://git@github.com/interloom/cli":   "git+ssh://git@github.com/interloom/cli.git",
		"git://github.com/interloom/cli":           "git://github.com/interloom/cli.git",
		"":                                         "",
	}

	for input, want := range tests {
		if got := normalizeRepositoryURL(input); got != want {
			t.Fatalf("normalizeRepositoryURL(%q) = %q, want %q", input, got, want)
		}
	}
}
