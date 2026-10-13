package store

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestValidateNamespace(t *testing.T) {
	for _, name := range []string{DefaultNamespace, "production", "staging-2", "team_a", "123", strings.Repeat("a", 63)} {
		if err := ValidateNamespace(name); err != nil {
			t.Errorf("valid namespace %q: %v", name, err)
		}
	}
	for _, name := range []string{"", ".", "..", "../outside", "/absolute", `a\b`, "prod/dev", "Production", "a b", "a.json", "-prod", "_prod", "con", "nul", "com1", "lpt9", strings.Repeat("a", 64)} {
		if err := ValidateNamespace(name); err == nil {
			t.Errorf("invalid namespace %q accepted", name)
		}
	}
}

func TestNamespaceSecretsIsolationAndLegacyCompatibility(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	if err := SaveSecrets(map[string]string{"KEY": "legacy", "DEFAULT_ONLY": "default"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Dir()
	legacyBefore, err := os.ReadFile(filepath.Join(dir, SecretsFile))
	if err != nil {
		t.Fatal(err)
	}
	for namespace, value := range map[string]string{"production": "prod", "staging": "stage"} {
		if err := SaveNamespaceSecrets(namespace, map[string]string{"KEY": value}); err != nil {
			t.Fatal(err)
		}
	}
	for namespace, value := range map[string]string{DefaultNamespace: "legacy", "production": "prod", "staging": "stage"} {
		secrets, err := LoadNamespaceSecrets(namespace)
		if err != nil {
			t.Fatal(err)
		}
		if secrets["KEY"] != value {
			t.Errorf("%s KEY = %q, want %q", namespace, secrets["KEY"], value)
		}
		if namespace != DefaultNamespace && len(secrets) != 1 {
			t.Errorf("%s inherited secrets: %v", namespace, secrets)
		}
	}
	if err := DeleteNamespaceSecret("production", "KEY"); err != nil {
		t.Fatal(err)
	}
	prod, err := LoadNamespaceSecrets("production")
	if err != nil || len(prod) != 0 {
		t.Fatalf("production after delete: %v, %v", prod, err)
	}
	if err := DeleteNamespaceSecret("production", "DEFAULT_ONLY"); err == nil {
		t.Fatal("delete must not fall back to the default namespace")
	}
	legacyAfter, err := os.ReadFile(filepath.Join(dir, SecretsFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(legacyAfter) != string(legacyBefore) {
		t.Fatal("named namespace operations changed legacy secrets.json")
	}
	if err := SaveNamespaceSecrets(DefaultNamespace, map[string]string{"NEW": "value"}); err != nil {
		t.Fatal(err)
	}
	legacy, err := LoadSecrets()
	if err != nil || legacy["NEW"] != "value" {
		t.Fatalf("default namespace not readable through legacy API: %v, %v", legacy, err)
	}
}

func TestMissingNamespaceReadsDoNotCreateFiles(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	secrets, err := LoadNamespaceSecrets("missing")
	if err != nil || len(secrets) != 0 || secrets == nil {
		t.Fatalf("missing namespace: %v, %v", secrets, err)
	}
	if err := DeleteNamespaceSecret("missing", "KEY"); err == nil {
		t.Fatal("expected error for missing key")
	}
	dir, _ := Dir()
	if _, err := os.Stat(filepath.Join(dir, namespacesDir)); !os.IsNotExist(err) {
		t.Fatalf("reading absent namespace created a directory: %v", err)
	}
	names, err := ListNamespaces()
	if err != nil || !reflect.DeepEqual(names, []string{DefaultNamespace}) {
		t.Fatalf("namespaces = %v, %v", names, err)
	}
}

func TestNamespaceFilesPermissionsAndSortedKeys(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	if err := SaveNamespaceSecrets("production", map[string]string{"Z": "last", "A": "first"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Dir()
	p := filepath.Join(dir, namespacesDir, "production.json")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(data), `"A"`) > strings.Index(string(data), `"Z"`) {
		t.Fatalf("keys not sorted: %s", data)
	}
	if runtime.GOOS != "windows" {
		for path, mode := range map[string]os.FileMode{filepath.Dir(p): 0o700, p: 0o600} {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != mode {
				t.Errorf("%s permissions = %o, want %o", path, info.Mode().Perm(), mode)
			}
		}
	}
}

func TestListNamespacesSortedAndFiltersUnrelatedFiles(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	for _, namespace := range []string{"staging", "production", "alpha"} {
		if err := SaveNamespaceSecrets(namespace, nil); err != nil {
			t.Fatal(err)
		}
	}
	dir, _ := Dir()
	for _, name := range []string{"production.json.tmp", "README", "Invalid.json", "default.json"} {
		if err := os.WriteFile(filepath.Join(dir, namespacesDir, name), []byte("ignored"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, namespacesDir, "directory.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	names, err := ListNamespaces()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"alpha", "default", "production", "staging"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("ListNamespaces = %v, want %v", names, want)
	}
}

func TestNamespaceStoreRejectsInvalidNames(t *testing.T) {
	setup(t)
	for _, name := range []string{"", "../outside", `..\outside`, "Production"} {
		if _, err := LoadNamespaceSecrets(name); err == nil {
			t.Errorf("LoadNamespaceSecrets(%q) succeeded", name)
		}
		if err := SaveNamespaceSecrets(name, map[string]string{"KEY": "value"}); err == nil {
			t.Errorf("SaveNamespaceSecrets(%q) succeeded", name)
		}
		if err := DeleteNamespaceSecret(name, "KEY"); err == nil {
			t.Errorf("DeleteNamespaceSecret(%q) succeeded", name)
		}
	}
}

func TestNamespaceCorruptJSONAndNullMap(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	if err := SaveNamespaceSecrets("production", nil); err != nil {
		t.Fatal(err)
	}
	p, _ := namespaceSecretsPath("production")
	if err := os.WriteFile(p, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadNamespaceSecrets("production"); err == nil || !strings.Contains(err.Error(), "parsing secrets") {
		t.Fatalf("expected parse error, got %v", err)
	}
	if err := os.WriteFile(p, []byte("null"), 0o600); err != nil {
		t.Fatal(err)
	}
	secrets, err := LoadNamespaceSecrets("production")
	if err != nil || secrets == nil {
		t.Fatalf("null JSON should load as a writable empty map: %v, %v", secrets, err)
	}
}

func TestNamespacesResolveFromSubdirectory(t *testing.T) {
	setup(t)
	if err := Init([]byte("0123456789abcdef"), "token"); err != nil {
		t.Fatal(err)
	}
	dir, _ := Dir()
	sub := filepath.Join(filepath.Dir(dir), "deep", "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	ResetRoot()
	if err := SaveNamespaceSecrets("production", map[string]string{"KEY": "value"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, namespacesDir, "production.json")); err != nil {
		t.Fatalf("namespace not stored at repo root: %v", err)
	}
}
