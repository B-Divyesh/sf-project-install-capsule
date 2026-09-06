package capsule

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

// The demo is embedded so an installed capsule binary always carries the same
// reviewable sample that is documented in examples/.

//go:embed demo_assets/README.md
var demoReadme []byte

// DemoResult names every file created by capsule demo. A fresh directory is
// required so trying the sample never overwrites a visitor's project.
type DemoResult struct {
	Directory    string           `json:"directory"`
	Config       string           `json:"config"`
	SampleInput  string           `json:"sample_input"`
	Review       CapabilityReview `json:"review"`
	ContainerRun string           `json:"container_run"`
}

func DemoConfig() Config {
	return Config{
		Version: 1,
		Image:   "docker.io/library/alpine:3.20",
		Install: "apk add --no-cache git python3 && git clone https://github.com/octocat/Hello-World.git .",
		Run:     "python3 -m http.server 3000 --bind 127.0.0.1",
		AllowHosts: []string{
			"codeload.github.com",
			"dl-cdn.alpinelinux.org",
			"github.com",
		},
		Ports: []int{3000},
	}
}

// CreateDemo makes a self-contained, temporary sample declaration and a
// bundled project note. It deliberately does not invoke a container engine;
// users can inspect the exact declaration before choosing to run it.
func CreateDemo(requestedDirectory string) (DemoResult, error) {
	directory := requestedDirectory
	var err error
	if directory == "" {
		directory, err = os.MkdirTemp("", "capsule-demo-")
		if err != nil {
			return DemoResult{}, fmt.Errorf("create demo directory: %w", err)
		}
	} else {
		if err := os.Mkdir(directory, 0o700); err != nil {
			if os.IsExist(err) {
				return DemoResult{}, fmt.Errorf("demo directory %s already exists; choose a new empty path", directory)
			}
			return DemoResult{}, fmt.Errorf("create demo directory: %w", err)
		}
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return DemoResult{}, fmt.Errorf("protect demo directory: %w", err)
	}

	sampleDir := filepath.Join(directory, "sample-project")
	if err := os.Mkdir(sampleDir, 0o700); err != nil {
		return DemoResult{}, fmt.Errorf("create sample input directory: %w", err)
	}
	sampleInput := filepath.Join(sampleDir, "README.md")
	if err := os.WriteFile(sampleInput, demoReadme, 0o600); err != nil {
		return DemoResult{}, fmt.Errorf("write bundled sample input: %w", err)
	}
	configPath := filepath.Join(directory, DefaultConfigPath)
	config := DemoConfig()
	if err := WriteConfig(configPath, config, false); err != nil {
		return DemoResult{}, err
	}
	review := Review(configPath, config)
	return DemoResult{
		Directory:    directory,
		Config:       configPath,
		SampleInput:  sampleInput,
		Review:       review,
		ContainerRun: "capsule run --config " + configPath,
	}, nil
}
