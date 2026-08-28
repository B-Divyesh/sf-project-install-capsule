package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/B-Divyesh/sf-project-install-capsule/internal/capsule"
)

const helpText = `Project Install Capsule 0.1.0

Run an unfamiliar project with no host mounts and no general network access.

Usage:
  capsule init [options]       Write a reviewable capsule.json
  capsule inspect [options]    Show the capability diff; execute nothing
  capsule run [options]        Review, install, and run in a rootless capsule
  capsule teardown [options]   Remove a stale capsule and write a receipt
  capsule verify [options]     Probe home and network isolation live
  capsule version              Print the version

Common options:
  --config PATH                Config file (default capsule.json)
  --json                       Machine-readable output

Run options:
  --dry-run                    Print the exact engine invocation only

Exit codes: 0 success, 2 invalid input, 3 engine unavailable/unsafe,
4 runtime failure. There are no interactive prompts or telemetry.

Containers are not complete security boundaries. Use a disposable VM for
hostile code or high-value credentials.
`

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

type intList []int

func (s *intList) String() string { return fmt.Sprint([]int(*s)) }
func (s *intList) Set(v string) error {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("expected a port number, got %q", v)
	}
	*s = append(*s, n)
	return nil
}

func main() {
	code := execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func execute(ctx context.Context, args []string, stdout, stderr *os.File) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, helpText)
		return 0
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Fprintln(stdout, capsule.Version)
		return 0
	}
	var err error
	switch args[0] {
	case "init":
		err = initCommand(args[1:], stdout, stderr)
	case "inspect":
		err = inspectCommand(args[1:], stdout, stderr)
	case "run":
		err = runCommand(ctx, args[1:], stdout, stderr)
	case "teardown":
		err = teardownCommand(ctx, args[1:], stdout, stderr)
	case "verify":
		err = verifyCommand(ctx, args[1:], stdout, stderr)
	case "bridge":
		err = bridgeCommand(ctx, args[1:], stderr)
	case "self-test":
		err = selfTestCommand(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "capsule: unknown command %q\n\n", args[0])
		fmt.Fprint(stderr, helpText)
		return 2
	}
	if err == nil {
		return 0
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	fmt.Fprintf(stderr, "capsule: %v\n", err)
	lowerError := strings.ToLower(err.Error())
	if strings.Contains(lowerError, "engine") || strings.Contains(lowerError, "rootless") {
		return 3
	}
	var exitErr *runtimeError
	if errors.As(err, &exitErr) {
		return 4
	}
	return 2
}

func commandFlags(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func initCommand(args []string, stdout, stderr io.Writer) error {
	fs := commandFlags("init", stderr)
	configPath := fs.String("config", capsule.DefaultConfigPath, "config file")
	image := fs.String("image", "docker.io/library/node:22-bookworm", "OCI image")
	install := fs.String("install", "", "install command")
	run := fs.String("run", "", "run command")
	force := fs.Bool("force", false, "replace an existing config")
	jsonOut := fs.Bool("json", false, "machine-readable output")
	var hosts stringList
	var ports intList
	fs.Var(&hosts, "allow-host", "approved HTTP/HTTPS hostname (repeatable)")
	fs.Var(&ports, "port", "approved loopback port (repeatable)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: capsule init --install COMMAND --run COMMAND [--allow-host HOST] [--port PORT]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	c := capsule.Config{Version: 1, Image: *image, Install: *install, Run: *run, AllowHosts: hosts, Ports: ports}
	if c.AllowHosts == nil {
		c.AllowHosts = []string{}
	}
	if c.Ports == nil {
		c.Ports = []int{}
	}
	if err := capsule.WriteConfig(*configPath, c, *force); err != nil {
		return err
	}
	if *jsonOut {
		return json.NewEncoder(stdout).Encode(map[string]any{"config": *configPath, "created": true})
	}
	fmt.Fprintf(stdout, "Created %s\n\nNext: capsule inspect --config %s\n", *configPath, *configPath)
	return nil
}

func inspectCommand(args []string, stdout, stderr io.Writer) error {
	fs := commandFlags("inspect", stderr)
	configPath := fs.String("config", capsule.DefaultConfigPath, "config file")
	jsonOut := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	c, err := capsule.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	review := capsule.Review(*configPath, c)
	if *jsonOut {
		return json.NewEncoder(stdout).Encode(review)
	}
	fmt.Fprint(stdout, review.Text())
	return nil
}

type runtimeError struct{ error }

func runCommand(ctx context.Context, args []string, stdout, stderr *os.File) error {
	fs := commandFlags("run", stderr)
	configPath := fs.String("config", capsule.DefaultConfigPath, "config file")
	jsonOut := fs.Bool("json", false, "machine-readable output")
	dryRun := fs.Bool("dry-run", false, "print engine command without running")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	c, err := capsule.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	_, err = capsule.Run(ctx, c, capsule.RunOptions{ConfigPath: *configPath, DryRun: *dryRun, JSON: *jsonOut, Stdout: stdout, Stderr: stderr})
	if err != nil && !strings.Contains(err.Error(), "engine") && !strings.Contains(err.Error(), "rootless") {
		return &runtimeError{err}
	}
	return err
}

func teardownCommand(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := commandFlags("teardown", stderr)
	configPath := fs.String("config", capsule.DefaultConfigPath, "config file")
	jsonOut := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	c, err := capsule.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	path, removed, err := capsule.Teardown(ctx, *configPath, c)
	if err != nil {
		return err
	}
	if *jsonOut {
		return json.NewEncoder(stdout).Encode(map[string]any{"removed": removed, "receipt": path})
	}
	state := "already absent"
	if removed {
		state = "removed"
	}
	fmt.Fprintf(stdout, "Capsule %s. Receipt: %s\n", state, path)
	return nil
}

func verifyCommand(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := commandFlags("verify", stderr)
	configPath := fs.String("config", capsule.DefaultConfigPath, "config file")
	jsonOut := fs.Bool("json", false, "machine-readable output")
	static := fs.Bool("static", false, "check generated arguments without starting an engine")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	c, err := capsule.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	if !*static {
		result, err := capsule.VerifyIsolation(ctx, c, *configPath)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "engine") || strings.Contains(strings.ToLower(err.Error()), "rootless") {
				return err
			}
			return &runtimeError{err}
		}
		if *jsonOut {
			return json.NewEncoder(stdout).Encode(result)
		}
		fmt.Fprintln(stdout, "PASS  rootless engine")
		fmt.Fprintln(stdout, "PASS  seeded home secret is unreadable")
		fmt.Fprintln(stdout, "PASS  direct network is denied")
		fmt.Fprintln(stdout, "PASS  undeclared host is denied")
		return nil
	}
	name := "capsule-" + capsule.ConfigID(*configPath)
	generated := capsule.RuntimeArgs(c, name, "/tmp/capsule-verify", "/path/to/capsule")
	joined := strings.Join(generated, "\x00")
	homeMounted := false
	if home := os.Getenv("HOME"); home != "" {
		for _, argument := range generated {
			if strings.HasPrefix(argument, "type=bind,src="+home+",") {
				homeMounted = true
			}
		}
	}
	checks := map[string]bool{
		"network denied":        strings.Contains(joined, "--network=none"),
		"host home not mounted": !homeMounted,
		"read-only root":        strings.Contains(joined, "--read-only"),
		"capabilities dropped":  strings.Contains(joined, "--cap-drop=ALL"),
		"no new privileges":     strings.Contains(joined, "--security-opt=no-new-privileges"),
		"empty tmpfs workspace": strings.Contains(joined, "/workspace:rw,nosuid,nodev,exec,size=1g"),
	}
	ok := true
	for _, passed := range checks {
		ok = ok && passed
	}
	if *jsonOut {
		return json.NewEncoder(stdout).Encode(map[string]any{"ok": ok, "checks": checks})
	}
	for label, passed := range checks {
		mark := "PASS"
		if !passed {
			mark = "FAIL"
		}
		fmt.Fprintf(stdout, "%-4s  %s\n", mark, label)
	}
	if !ok {
		return errors.New("one or more isolation checks failed")
	}
	return nil
}

func selfTestCommand(args []string, stdout, stderr io.Writer) error {
	fs := commandFlags("self-test", stderr)
	hostSecret := fs.String("host-secret", "", "host sentinel path")
	denyHost := fs.String("deny-host", "", "undeclared hostname")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *hostSecret == "" || *denyHost == "" {
		return errors.New("self-test requires --host-secret and --deny-host")
	}
	result, err := capsule.SelfTest(*hostSecret, *denyHost)
	_ = json.NewEncoder(stdout).Encode(result)
	return err
}

func bridgeCommand(ctx context.Context, args []string, stderr io.Writer) error {
	fs := commandFlags("bridge", stderr)
	socket := fs.String("socket", "", "host proxy Unix socket")
	var forwards stringList
	fs.Var(&forwards, "forward", "port=unix-socket")
	if err := fs.Parse(args); err != nil {
		return err
	}
	command := fs.Args()
	if len(command) > 0 && command[0] == "--" {
		command = command[1:]
	}
	return capsule.Bridge(ctx, *socket, forwards, command)
}
