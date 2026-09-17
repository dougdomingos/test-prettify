package modutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ReadModulePath extracts the module path from the "module" directive of a
// go.mod file at the given path.
func ReadModulePath(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(string(raw), "\n") {
		line = strings.TrimSpace(line)

		if after, ok := strings.CutPrefix(line, "module "); ok {
			module := strings.TrimSpace(after)
			module = strings.Trim(module, `"`)
			if module == "" || strings.HasPrefix(module, "//") {
				return "", errors.New("go.mod has an empty module directive")
			}
			return module, nil
		}
	}

	return "", errors.New("go.mod is missing a module directive")
}

// ProjectName walks up from the current directory looking for a go.mod file,
// reads the module path, and returns its last segment as the project name.
func ProjectName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		modPath := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(modPath); err == nil && !info.IsDir() {
			module, err := ReadModulePath(modPath)
			if err != nil {
				return "", err
			}

			if i := strings.LastIndex(module, "/"); i >= 0 {
				return module[i+1:], nil
			}
			return module, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
