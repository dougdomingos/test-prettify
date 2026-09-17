package modutil

import (
	"errors"
	"os"
	"path"
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

// FindModule walks up from dir looking for a go.mod file and returns the
// module import path declared in it together with the directory that holds
// the file.
func FindModule(dir string) (string, string, error) {
	for {
		info, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil && !info.IsDir() {
			modulePath, err := ReadModulePath(filepath.Join(dir, "go.mod"))
			if err != nil {
				return "", "", err
			}
			return modulePath, dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

// ProjectName walks up from the current directory looking for a go.mod file,
// reads the module path, and returns its last segment as the project name.
func ProjectName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	module, _, err := FindModule(dir)
	if err != nil {
		return "", err
	}

	return path.Base(module), nil
}
