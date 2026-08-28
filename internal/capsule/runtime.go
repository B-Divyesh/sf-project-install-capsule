package capsule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type RunOptions struct {
	ConfigPath string
	DryRun     bool
	JSON       bool
	Stdout     *os.File
	Stderr     *os.File
}

type RunPreview struct {
	Engine       string           `json:"engine"`
	Container    string           `json:"container"`
	Arguments    []string         `json:"arguments"`
	Capabilities CapabilityReview `json:"capabilities"`
}

type Receipt struct {
	Version      int              `json:"version"`
	Capsule      string           `json:"capsule"`
	Container    string           `json:"container"`
	Engine       string           `json:"engine"`
	StartedAt    time.Time        `json:"started_at"`
	FinishedAt   time.Time        `json:"finished_at"`
	Outcome      string           `json:"outcome"`
	ExitCode     int              `json:"exit_code"`
	Capabilities CapabilityReview `json:"capabilities"`
}

func Run(ctx context.Context, c Config, opts RunOptions) (string, error) {
	engine, err := FindEngine(ctx, !opts.DryRun)
	if err != nil {
		if opts.DryRun {
			engine = Engine{Path: "podman", Kind: "podman"}
		} else {
			return "", err
		}
	}
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate capsule binary: %w", err)
	}
	executable, _ = filepath.EvalSymlinks(executable)
	configID := ConfigID(opts.ConfigPath)
	containerName := "capsule-" + configID
	runtimeDir := filepath.Join(os.TempDir(), containerName)
	if !opts.DryRun {
		runtimeDir, err = os.MkdirTemp("", containerName+"-")
		if err != nil {
			return "", fmt.Errorf("create ephemeral bridge directory: %w", err)
		}
	}
	args := RuntimeArgs(c, containerName, runtimeDir, executable)
	preview := RunPreview{Engine: engine.Kind, Container: containerName, Arguments: args, Capabilities: Review(opts.ConfigPath, c)}
	if opts.DryRun {
		if opts.JSON {
			_ = json.NewEncoder(opts.Stdout).Encode(preview)
		} else {
			fmt.Fprint(opts.Stdout, preview.Capabilities.Text())
			fmt.Fprintf(opts.Stdout, "\nDRY RUN\n  %s %s\n", engine.Kind, shellJoin(args))
		}
		return "", nil
	}

	defer os.RemoveAll(runtimeDir)
	if err := os.Chmod(runtimeDir, 0o711); err != nil {
		return "", fmt.Errorf("set ephemeral bridge permissions: %w", err)
	}
	proxy, err := StartAllowProxy(filepath.Join(runtimeDir, "outbound.sock"), c.AllowHosts)
	if err != nil {
		return "", err
	}
	defer proxy.Close()
	publishers, err := StartPortPublishers(runtimeDir, c.Ports)
	if err != nil {
		return "", err
	}
	defer publishers.Close()

	if !opts.JSON {
		fmt.Fprint(opts.Stdout, preview.Capabilities.Text())
		fmt.Fprintln(opts.Stdout, "\nRUNNING  Press Ctrl+C to stop; teardown is automatic.")
	}
	started := time.Now().UTC()
	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	cmd := exec.CommandContext(runCtx, engine.Path, args...)
	cmd.Stdin, cmd.Stderr = os.Stdin, opts.Stderr
	cmd.Stdout = opts.Stdout
	if opts.JSON {
		// Keep the final machine-readable result on stdout; workload logs belong
		// on stderr in automation mode.
		cmd.Stdout = opts.Stderr
	}
	runErr := cmd.Run()
	_, _ = exec.CommandContext(context.Background(), engine.Path, "rm", "-f", containerName).CombinedOutput()
	exitCode := 0
	outcome := "completed"
	if runErr != nil {
		outcome = "failed"
		exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	receipt := Receipt{Version: 1, Capsule: configID, Container: containerName, Engine: engine.Kind, StartedAt: started, FinishedAt: time.Now().UTC(), Outcome: outcome, ExitCode: exitCode, Capabilities: preview.Capabilities}
	receiptPath, receiptErr := WriteReceipt(receipt)
	if receiptErr != nil {
		return "", receiptErr
	}
	if opts.JSON {
		_ = json.NewEncoder(opts.Stdout).Encode(map[string]any{"receipt": receiptPath, "outcome": outcome, "exit_code": exitCode})
	} else {
		fmt.Fprintf(opts.Stdout, "\nRECEIPT  %s\n", receiptPath)
	}
	if runErr != nil {
		return receiptPath, fmt.Errorf("capsule process exited with code %d", exitCode)
	}
	return receiptPath, nil
}

func RuntimeArgs(c Config, name, runtimeDir, executable string) []string {
	containerDir := "/run/capsule"
	args := baseRuntimeArgs(c.Image, name, runtimeDir, executable)
	args = append(args, "/capsule", "bridge", "--socket", containerDir+"/outbound.sock")
	for _, port := range c.Ports {
		args = append(args, "--forward", fmt.Sprintf("%d=%s/port-%d.sock", port, containerDir, port))
	}
	command := "set -eu; " + c.Install + "; exec /bin/sh -lc " + shellQuote(c.Run)
	args = append(args, "--", "/bin/sh", "-lc", command)
	return args
}

func baseRuntimeArgs(image, name, runtimeDir, executable string) []string {
	return []string{"run", "--rm", "--name", name, "--label", "in.sociobot.capsule=true", "--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--pids-limit=256", "--tmpfs", "/workspace:rw,nosuid,nodev,exec,size=1g", "--tmpfs", "/tmp:rw,nosuid,nodev,noexec,size=128m", "--mount", "type=bind,src=" + executable + ",dst=/capsule,ro", "--mount", "type=bind,src=" + runtimeDir + ",dst=/run/capsule", "--workdir", "/workspace", image}
}

func WriteReceipt(receipt Receipt) (string, error) {
	dir := filepath.Join(".capsule", "receipts")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create receipt directory: %w", err)
	}
	path := filepath.Join(dir, receipt.FinishedAt.Format("20060102T150405.000000000Z")+".json")
	b, _ := json.MarshalIndent(receipt, "", "  ")
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return "", fmt.Errorf("write teardown receipt: %w", err)
	}
	return path, nil
}

func Teardown(ctx context.Context, configPath string, c Config) (string, bool, error) {
	engine, err := FindEngine(ctx, true)
	if err != nil {
		return "", false, err
	}
	id := ConfigID(configPath)
	name := "capsule-" + id
	started := time.Now().UTC()
	out, removeErr := exec.CommandContext(ctx, engine.Path, "rm", "-f", name).CombinedOutput()
	removed := removeErr == nil
	outcome := "already absent"
	if removed {
		outcome = "removed"
	} else {
		lower := strings.ToLower(string(out))
		if !strings.Contains(lower, "no such") && !strings.Contains(lower, "not found") && !strings.Contains(lower, "does not exist") {
			return "", false, fmt.Errorf("remove %s: %s", name, compactOutput(out, removeErr))
		}
	}
	receipt := Receipt{Version: 1, Capsule: id, Container: name, Engine: engine.Kind, StartedAt: started, FinishedAt: time.Now().UTC(), Outcome: outcome, ExitCode: 0, Capabilities: Review(configPath, c)}
	path, err := WriteReceipt(receipt)
	return path, removed, err
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}
