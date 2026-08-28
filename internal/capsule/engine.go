package capsule

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Engine struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func FindEngine(ctx context.Context, checkRootless bool) (Engine, error) {
	if selected := os.Getenv("CAPSULE_ENGINE"); selected != "" {
		path, err := exec.LookPath(selected)
		if err != nil {
			return Engine{}, fmt.Errorf("CAPSULE_ENGINE %q was not found", selected)
		}
		engine := Engine{Path: path, Kind: engineKind(selected)}
		if engine.Kind == "" {
			return Engine{}, errors.New("CAPSULE_ENGINE must name podman or docker")
		}
		if checkRootless {
			if err := engine.RequireRootless(ctx); err != nil {
				return Engine{}, err
			}
		}
		return engine, nil
	}
	for _, candidate := range []string{"podman", "docker"} {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		engine := Engine{Path: path, Kind: candidate}
		if !checkRootless || engine.RequireRootless(ctx) == nil {
			return engine, nil
		}
	}
	return Engine{}, errors.New("no rootless container engine found; install rootless Podman or Docker")
}

func engineKind(value string) string {
	base := value
	if i := strings.LastIndexAny(base, `/\\`); i >= 0 {
		base = base[i+1:]
	}
	base = strings.ToLower(base)
	if strings.Contains(base, "podman") {
		return "podman"
	}
	if strings.Contains(base, "docker") {
		return "docker"
	}
	return ""
}

func (e Engine) RequireRootless(ctx context.Context) error {
	var args []string
	if e.Kind == "podman" {
		args = []string{"info", "--format", "{{.Host.Security.Rootless}}"}
	} else {
		args = []string{"info", "--format", "{{json .SecurityOptions}}"}
	}
	out, err := exec.CommandContext(ctx, e.Path, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot inspect %s engine: %s", e.Kind, compactOutput(out, err))
	}
	answer := strings.ToLower(strings.TrimSpace(string(out)))
	rootless := answer == "true" || strings.Contains(answer, "rootless")
	if !rootless {
		return fmt.Errorf("%s engine is not rootless; refusing to weaken the capsule boundary", e.Kind)
	}
	return nil
}

func compactOutput(out []byte, fallback error) string {
	s := strings.TrimSpace(string(out))
	if s == "" {
		return fallback.Error()
	}
	if len(s) > 300 {
		s = s[:300] + "…"
	}
	return s
}
