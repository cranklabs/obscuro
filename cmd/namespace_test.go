package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janklabs/obscuro/internal/store"
)

func TestNamespaceSecretCommands(t *testing.T) {
	setup(t)
	initVault(t)
	cfgBefore, err := os.ReadFile(filepath.Join(".obscuro", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, ns := range []string{"default", "production", "staging"} {
		if _, _, err := execCmd(t, "set", "API_KEY", "-n", ns, "--password", testPassword, "--value", ns+"-value"); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := execCmd(t, "set", "DEFAULT_ONLY", "--value", "default-only", "--password", testPassword); err != nil {
		t.Fatal(err)
	}
	for _, ns := range []string{"default", "production", "staging"} {
		out, _, err := execCmd(t, "--namespace", ns, "get", "API_KEY", "--password", testPassword)
		if err != nil || out != ns+"-value" {
			t.Fatalf("get in %s = %q, %v", ns, out, err)
		}
	}
	// Omitting --namespace after a named operation still selects default.
	out, _, err := execCmd(t, "get", "API_KEY", "--password", testPassword)
	if err != nil || out != "default-value" {
		t.Fatalf("implicit default get = %q, %v", out, err)
	}
	out, _, err = execCmd(t, "list", "-n", "production", "--password", "wrong")
	if err != nil || out != "API_KEY\n" {
		t.Fatalf("names-only list = %q, %v", out, err)
	}
	if _, _, err := execCmd(t, "get", "DEFAULT_ONLY", "-n", "production", "--password", testPassword); err == nil {
		t.Fatal("get must not inherit default secrets")
	}
	if _, _, err := execCmd(t, "edit", "DEFAULT_ONLY", "-n", "production", "--password", testPassword); err == nil || !strings.Contains(err.Error(), "not found in namespace 'production'") {
		t.Fatalf("edit must not inherit default secrets: %v", err)
	}
	if _, _, err := execCmd(t, "set", "API_KEY", "-n", "production", "--password", testPassword, "--value", "updated"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execCmd(t, "rm", "API_KEY", "-n", "production", "--password", testPassword); err != nil {
		t.Fatal(err)
	}
	out, _, err = execCmd(t, "list", "-n", "production")
	if err != nil || out != "No secrets stored.\n" {
		t.Fatalf("list after remove = %q, %v", out, err)
	}
	for _, ns := range []string{"default", "staging"} {
		out, _, err := execCmd(t, "get", "API_KEY", "-n", ns, "--password", testPassword)
		if err != nil || out != ns+"-value" {
			t.Fatalf("production mutations affected %s: %q, %v", ns, out, err)
		}
	}
	cfgAfter, err := os.ReadFile(filepath.Join(".obscuro", "config.json"))
	if err != nil || string(cfgAfter) != string(cfgBefore) {
		t.Fatalf("namespace operations changed vault configuration: %v", err)
	}
}

func TestNamespaceInjectIsolation(t *testing.T) {
	setup(t)
	initVault(t)
	for _, args := range [][]string{
		{"set", "API_KEY", "--value", "default-key"},
		{"set", "DEFAULT_ONLY", "--value", "must-not-inject"},
		{"set", "API_KEY", "--value", "prod-key", "-n", "production"},
	} {
		if _, _, err := execCmd(t, append(args, "--password", testPassword)...); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { rootCmd.SetIn(nil) })
	rootCmd.SetIn(strings.NewReader("__API_KEY__ __DEFAULT_ONLY__"))
	out, stderr, err := execCmd(t, "inject", "-n", "production", "--password", testPassword)
	if err != nil || out != "prod-key __DEFAULT_ONLY__" || !strings.Contains(stderr, "unresolved placeholders: DEFAULT_ONLY") {
		t.Fatalf("namespaced inject: %q, %q, %v", out, stderr, err)
	}
	rootCmd.SetIn(strings.NewReader("__API_KEY__ __DEFAULT_ONLY__"))
	out, _, err = execCmd(t, "inject", "-n", "production", "--password", testPassword, "--strict")
	if err == nil || out != "" {
		t.Fatalf("strict inject must fail without output: %q, %v", out, err)
	}
	injectStrict = false
	rootCmd.SetIn(strings.NewReader("__API_KEY__"))
	out, _, err = execCmd(t, "inject", "-n", "missing", "--password", testPassword, "--strict")
	if err == nil || out != "" {
		t.Fatalf("absent namespace must not fall back: %q, %v", out, err)
	}
}

func TestNamespaceList(t *testing.T) {
	setup(t)
	if _, _, err := execCmd(t, "namespace", "list"); err == nil {
		t.Fatal("namespace list must require vault initialization")
	}
	initVault(t)
	out, _, err := execCmd(t, "namespace", "list")
	if err != nil || out != "default\n" {
		t.Fatalf("empty namespace list = %q, %v", out, err)
	}
	for _, ns := range []string{"staging", "production"} {
		// Invalid ciphertext is fine: namespace listing never decrypts values.
		if err := store.SaveNamespaceSecrets(ns, map[string]string{"KEY": "!!!"}); err != nil {
			t.Fatal(err)
		}
	}
	password = ""
	t.Setenv("OBSCURO_PASSWORD", "wrong")
	out, _, err = execCmd(t, "namespaces", "list")
	if err != nil || out != "default\nproduction\nstaging\n" {
		t.Fatalf("sorted namespace list = %q, %v", out, err)
	}
}

func TestInvalidNamespaceRejectedBeforeInputOrAuthentication(t *testing.T) {
	setup(t)
	withFakePassword(t) // Any attempted prompt fails this test.
	for _, args := range [][]string{
		{"set", "KEY", "--value", "value"}, {"get", "KEY"}, {"edit", "KEY"},
		{"remove", "KEY"}, {"list"}, {"inject"}, {"import", "missing.env"},
	} {
		for _, ns := range []string{"", "../escape", "Production"} {
			_, _, err := execCmd(t, append(args, "--namespace", ns)...)
			if err == nil || !strings.Contains(err.Error(), "invalid namespace") {
				t.Errorf("%v with namespace %q: %v", args, ns, err)
			}
		}
	}
}

func TestNamespaceWrongPasswordDoesNotWrite(t *testing.T) {
	setup(t)
	initVault(t)
	if _, _, err := execCmd(t, "set", "KEY", "-n", "production", "--password", "wrong", "--value", "value"); err == nil || !strings.Contains(err.Error(), "incorrect password") {
		t.Fatalf("expected password verification error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".obscuro", "namespaces")); !os.IsNotExist(err) {
		t.Fatalf("failed authentication created namespace: %v", err)
	}
}
