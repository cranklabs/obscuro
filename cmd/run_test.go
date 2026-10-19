package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/janklabs/obscuro/internal/crypto"
	"github.com/janklabs/obscuro/internal/store"
)

// The test binary doubles as a portable child command, without requiring a shell.
func TestRunHelperProcess(t *testing.T) {
	separator := slices.Index(os.Args, "--")
	if separator < 0 || len(os.Args) <= separator+2 || os.Args[separator+1] != "obscuro-run-test-helper" {
		return
	}
	args := os.Args[separator+2:]
	switch args[0] {
	case "env":
		values := make(map[string]string)
		for _, name := range args[1:] {
			if value, ok := os.LookupEnv(name); ok {
				values[name] = value
			}
		}
		_ = json.NewEncoder(os.Stdout).Encode(values)
	case "args":
		_ = json.NewEncoder(os.Stdout).Encode(args[1:])
	case "io":
		_, _ = io.Copy(os.Stdout, os.Stdin)
		fmt.Fprintln(os.Stderr, "child stderr")
	case "cwd":
		dir, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, dir)
	case "marker":
		if err := os.WriteFile(args[1], []byte("started"), 0o600); err != nil {
			os.Exit(1)
		}
	case "exit":
		code, _ := strconv.Atoi(args[1])
		os.Exit(code)
	case "signal":
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM)
		fmt.Fprintln(os.Stdout, os.Getpid())
		<-signals
		os.Exit(23)
	case "terminate":
		process, _ := os.FindProcess(os.Getpid())
		_ = process.Signal(syscall.SIGTERM)
		select {}
	case "cli":
		rootCmd.SetArgs(args[1:])
		Execute()
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

func runHelperArgs(t *testing.T, args ...string) []string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return append([]string{executable, "-test.run=^TestRunHelperProcess$", "--", "obscuro-run-test-helper"}, args...)
}

func saveRunSecrets(t *testing.T, ns string, values map[string]string) {
	t.Helper()
	cfg, err := store.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	salt, err := cfg.DecodeSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey(testPassword, salt)
	secrets := make(map[string]string, len(values))
	for name, value := range values {
		encrypted, err := crypto.Encrypt(key, []byte(value))
		if err != nil {
			t.Fatal(err)
		}
		secrets[name] = encrypted
	}
	if err := store.SaveNamespaceSecrets(ns, secrets); err != nil {
		t.Fatal(err)
	}
}

func TestRunNamespaceEnvironment(t *testing.T) {
	setup(t)
	initVault(t)
	t.Setenv("OBSCURO_RUN_SHARED", "parent-value")
	t.Setenv("OBSCURO_RUN_INHERITED", "keep-me")
	// Ensure these names are absent even if the developer has set them locally.
	for _, name := range []string{"OBSCURO_RUN_DEFAULT_ONLY", "OBSCURO_RUN_ADDED", "OBSCURO_RUN_EMPTY"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	saveRunSecrets(t, "default", map[string]string{
		"OBSCURO_RUN_SHARED":       "default-value",
		"OBSCURO_RUN_DEFAULT_ONLY": "default-only",
	})
	values := map[string]string{
		"OBSCURO_RUN_SHARED": "spaces 'quotes' \"double quotes\" $literal=equals\nsecond line",
		"OBSCURO_RUN_ADDED":  "namespace-value",
		"OBSCURO_RUN_EMPTY":  "",
	}
	saveRunSecrets(t, "my-app", values)
	keys := []string{"OBSCURO_RUN_SHARED", "OBSCURO_RUN_DEFAULT_ONLY", "OBSCURO_RUN_ADDED", "OBSCURO_RUN_EMPTY", "OBSCURO_RUN_INHERITED"}
	child := runHelperArgs(t, append([]string{"env"}, keys...)...)
	args := append([]string{"run", "-n", "my-app", "--password", testPassword, "--"}, child...)
	out, stderr, err := execCmd(t, args...)
	if err != nil || stderr != "" {
		t.Fatalf("run failed: %v, stderr: %q", err, stderr)
	}
	var got map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	values["OBSCURO_RUN_INHERITED"] = "keep-me"
	if !reflect.DeepEqual(got, values) {
		t.Fatalf("child environment = %q, want %q", got, values)
	}
	if os.Getenv("OBSCURO_RUN_SHARED") != "parent-value" {
		t.Fatal("run changed the parent environment")
	}
	if _, ok := os.LookupEnv("OBSCURO_RUN_ADDED"); ok {
		t.Fatal("run added a secret to the parent environment")
	}
	// A subsequent invocation without -n must still select default.
	out, _, err = execCmd(t, append([]string{"run", "--password", testPassword, "--"}, child...)...)
	if err != nil {
		t.Fatal(err)
	}
	got = nil
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["OBSCURO_RUN_SHARED"] != "default-value" || got["OBSCURO_RUN_DEFAULT_ONLY"] != "default-only" {
		t.Fatalf("default namespace not selected: %q", got)
	}
	if _, ok := got["OBSCURO_RUN_ADDED"]; ok {
		t.Fatal("default namespace inherited secrets from my-app")
	}
}

func TestRunEmptyOrAbsentNamespace(t *testing.T) {
	setup(t)
	initVault(t)
	t.Setenv("OBSCURO_RUN_SHARED", "inherited")
	child := runHelperArgs(t, "env", "OBSCURO_RUN_SHARED")
	for _, ns := range []string{"default", "absent"} {
		out, _, err := execCmd(t, append([]string{"run", "-n", ns, "--password", testPassword, "--"}, child...)...)
		if err != nil || strings.TrimSpace(out) != `{"OBSCURO_RUN_SHARED":"inherited"}` {
			t.Fatalf("empty namespace %s: output %q, error %v", ns, out, err)
		}
		if ns == "default" {
			saveRunSecrets(t, "default", map[string]string{"OBSCURO_RUN_SHARED": "must-not-fall-back"})
		}
	}
	if _, err := os.Stat(filepath.Join(".obscuro", "namespaces")); !os.IsNotExist(err) {
		t.Fatalf("reading absent namespace created files: %v", err)
	}
}

func TestRunPassesArgumentsLiterally(t *testing.T) {
	setup(t)
	initVault(t)
	want := []string{"compose", "up", "-d", "-n", "child-namespace", "--password", "child-password", "--help", "--", "two words", "$API_KEY", "*", ""}
	child := runHelperArgs(t, append([]string{"args"}, want...)...)
	// Also accept simple commands without a separator; child flags still belong
	// to the child once its executable has been encountered.
	for _, prefix := range [][]string{{"run", "--password", testPassword, "--"}, {"run", "--password", testPassword}} {
		out, _, err := execCmd(t, append(prefix, child...)...)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("child args = %q, want %q", got, want)
		}
	}
}

func TestRunStreamsAndWorkingDirectory(t *testing.T) {
	setup(t)
	initVault(t)
	if err := os.Mkdir("app", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("app"); err != nil {
		t.Fatal(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	store.ResetRoot()
	rootCmd.SetIn(strings.NewReader("input passed through\n"))
	t.Cleanup(func() { rootCmd.SetIn(nil) })
	out, stderr, err := execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "io")...)...)
	if err != nil || out != "input passed through\n" || stderr != "child stderr\n" {
		t.Fatalf("stdio = %q, %q, %v", out, stderr, err)
	}
	out, _, err = execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "cwd")...)...)
	if err != nil || out != dir {
		t.Fatalf("child cwd = %q, %v; want %q", out, err, dir)
	}
}

func TestRunExitStatus(t *testing.T) {
	setup(t)
	initVault(t)
	for _, code := range []int{0, 7, 42, 127} {
		_, stderr, err := execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "exit", strconv.Itoa(code))...)...)
		if code == 0 {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var exit *exitErr
		if !errors.As(err, &exit) || exit.code != code {
			t.Fatalf("exit status = %v, want %d", err, code)
		}
		if strings.Contains(stderr, "Usage:") {
			t.Fatalf("child failure printed Obscuro usage: %q", stderr)
		}
	}
	// Verify Execute() returns the child's status to the OS, not just from RunE.
	child := runHelperArgs(t, "exit", "42")
	args := runHelperArgs(t, append([]string{"cli", "run", "--password", testPassword, "--"}, child...)...)
	process := exec.Command(args[0], args[1:]...)
	var exit *exec.ExitError
	if err := process.Run(); !errors.As(err, &exit) || exit.ExitCode() != 42 {
		t.Fatalf("CLI exit status = %v, want 42", err)
	}
}

func TestRunRejectsMissingCommand(t *testing.T) {
	setup(t)
	withFakePassword(t)
	for _, args := range [][]string{{"run"}, {"run", "--"}} {
		if _, _, err := execCmd(t, args...); err == nil || !strings.Contains(err.Error(), "requires at least 1 arg") {
			t.Fatalf("expected missing-command error, got %v", err)
		}
	}
}

func TestRunAuthenticationFailureDoesNotStartCommand(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "started")
	child := runHelperArgs(t, "marker", marker)
	for _, initialized := range []bool{false, true} {
		want := "not initialized"
		if initialized {
			initVault(t)
			want = "incorrect password"
		}
		out, _, err := execCmd(t, append([]string{"run", "--password", "wrong-password", "--"}, child...)...)
		if err == nil || !strings.Contains(err.Error(), want) || out != "" {
			t.Fatalf("authentication failure = %q, %v; want %q", out, err, want)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("child started before authentication: %v", err)
		}
	}
}

func TestRunCorruptSecretDoesNotStartCommand(t *testing.T) {
	setup(t)
	initVault(t)
	const value = "must-not-be-logged"
	saveRunSecrets(t, "default", map[string]string{"A_VALID": value})
	secrets, err := store.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	secrets["Z_CORRUPTED"] = "!!!"
	if err := store.SaveSecrets(secrets); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "started")
	out, stderr, err := execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "marker", marker)...)...)
	if err == nil || !strings.Contains(err.Error(), "Z_CORRUPTED") || out != "" {
		t.Fatalf("corrupt-secret error = %q, %v", out, err)
	}
	if strings.Contains(stderr, value) || strings.Contains(err.Error(), value) {
		t.Fatal("run logged a decrypted value")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("child started despite a corrupt secret: %v", err)
	}
}

func TestRunRejectsInvalidEnvironmentEntries(t *testing.T) {
	for _, tc := range []struct{ name, key, value string }{
		{"empty name", "", "value"},
		{"equals in name", "BAD=KEY", "value"},
		{"NUL in name", "BAD\x00KEY", "value"},
		{"NUL in value", "KEY", "private\x00value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)
			initVault(t)
			saveRunSecrets(t, "default", map[string]string{tc.key: tc.value})
			marker := filepath.Join(t.TempDir(), "started")
			out, stderr, err := execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "marker", marker)...)...)
			if err == nil || !strings.Contains(err.Error(), "cannot be used as an environment variable") || out != "" {
				t.Fatalf("invalid environment entry = %q, %v", out, err)
			}
			if strings.Contains(stderr, tc.value) || strings.Contains(err.Error(), tc.value) {
				t.Fatal("error exposed the secret value")
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("child started despite invalid environment: %v", err)
			}
		})
	}
}

func TestRunMissingExecutable(t *testing.T) {
	setup(t)
	initVault(t)
	_, _, err := execCmd(t, "run", "--password", testPassword, "--", filepath.Join(t.TempDir(), "missing"))
	var exit *exitErr
	if err == nil || errors.As(err, &exit) || !strings.Contains(err.Error(), "running command") {
		t.Fatalf("expected process-start error, got %v", err)
	}
}

func TestRunForwardsTermination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support sending SIGTERM with os.Process.Signal")
	}
	setup(t)
	initVault(t)
	child := runHelperArgs(t, "signal")
	args := runHelperArgs(t, append([]string{"cli", "run", "--password", testPassword, "--"}, child...)...)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, args[0], args[1:]...)
	stdout, err := process.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = process.Process.Kill() })
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatalf("waiting for child readiness: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		t.Fatal(err)
	}
	childProcess, err := os.FindProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	finished := false
	t.Cleanup(func() {
		if !finished {
			_ = childProcess.Kill()
		}
		_ = childProcess.Release()
	})
	if err := process.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	var exit *exec.ExitError
	if err := process.Wait(); !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("child did not receive forwarded termination: %v", err)
	}
	finished = true
}

func TestRunSignalExitStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support sending SIGTERM with os.Process.Signal")
	}
	setup(t)
	initVault(t)
	_, _, err := execCmd(t, append([]string{"run", "--password", testPassword, "--"}, runHelperArgs(t, "terminate")...)...)
	var exit *exitErr
	if !errors.As(err, &exit) || exit.code != 128+int(syscall.SIGTERM) {
		t.Fatalf("signal exit status = %v, want %d", err, 128+int(syscall.SIGTERM))
	}
}
