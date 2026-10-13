package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DefaultNamespace holds legacy secrets stored in .obscuro/secrets.json.
const DefaultNamespace = "default"

const namespacesDir = "namespaces"

var namespaceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// ValidateNamespace checks that a name is safe and portable as a filename.
func ValidateNamespace(name string) error {
	if !namespaceNameRe.MatchString(name) {
		return fmt.Errorf("invalid namespace %q: use 1-63 lowercase letters, digits, hyphens or underscores, starting with a letter or digit", name)
	}
	// Windows device names remain reserved even with a .json extension.
	switch name {
	case "con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9":
		return fmt.Errorf("invalid namespace %q: reserved filename on Windows", name)
	}
	return nil
}

func namespaceSecretsPath(namespace string) (string, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}
	if namespace == DefaultNamespace {
		return secretsPath()
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, namespacesDir, namespace+".json"), nil
}

// LoadNamespaceSecrets reads only the selected namespace. An absent namespace
// is empty; secrets are never inherited from the default or another namespace.
func LoadNamespaceSecrets(namespace string) (map[string]string, error) {
	p, err := namespaceSecretsPath(namespace)
	if err != nil {
		return nil, err
	}
	return loadSecrets(p)
}

// SaveNamespaceSecrets writes a namespace atomically, creating named namespaces
// on first write. The default keeps the legacy secrets.json format and path.
func SaveNamespaceSecrets(namespace string, secrets map[string]string) error {
	p, err := namespaceSecretsPath(namespace)
	if err != nil {
		return err
	}
	if namespace != DefaultNamespace {
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			return fmt.Errorf("creating namespaces directory: %w", err)
		}
	}
	if secrets == nil {
		secrets = map[string]string{}
	}
	return writeJSON(p, secrets)
}

// DeleteNamespaceSecret removes a key from only the selected namespace.
func DeleteNamespaceSecret(namespace, name string) error {
	secrets, err := LoadNamespaceSecrets(namespace)
	if err != nil {
		return err
	}
	if _, ok := secrets[name]; !ok {
		return fmt.Errorf("secret '%s' not found in namespace '%s'", name, namespace)
	}
	delete(secrets, name)
	return SaveNamespaceSecrets(namespace, secrets)
}

// ListNamespaces returns sorted namespace names, including the default. Only
// persisted named namespaces appear; reading an absent namespace creates none.
func ListNamespaces() ([]string, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	names := []string{DefaultNamespace}
	entries, err := os.ReadDir(filepath.Join(dir, namespacesDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return names, nil
		}
		return nil, fmt.Errorf("reading namespaces: %w", err)
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if name != DefaultNamespace && ValidateNamespace(name) == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}
