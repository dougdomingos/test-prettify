package coverage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dougdomingos/test-prettify/internal/modutil"
)

// SrcResolver maps source file import paths recorded in a coverage profile to
// their location on the local filesystem.
type SrcResolver struct {

	// modulePath is the import path declared in the project's go.mod, used as
	// a prefix to strip from profile paths (e.g.
	// "example.com/mod/internal/foo/a.go" -> "/internal/foo/a.go").
	modulePath string

	// rootDir is the directory that contains go.mod, used as the base onto which
	// stripped paths are joined.
	rootDir string
}

// NewSrcResolver builds a SrcResolver rooted at dir. It looks for a go.mod
// file within the directory (or its ancestors) to learn the module import
// path. Paths that do not share that prefix are left unresolved.
func NewSrcResolver(dir string) (*SrcResolver, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	modulePath, modDir, err := modutil.FindModule(abs)
	if err != nil {
		return nil, fmt.Errorf("go.mod not found: %w: run from the module root or pass an explicit --coverage-root", err)
	}

	return &SrcResolver{modulePath: modulePath, rootDir: modDir}, nil
}

// Resolve maps a profile file path to its absolute local path. It returns
// false when the path does not belong to this module or the file does not
// exist on disk.
func (r *SrcResolver) Resolve(profilePath string) (string, bool) {
	rel := r.GetRelPath(profilePath)
	if rel == profilePath || rel == "" {
		return "", false
	}

	abs := filepath.Join(r.rootDir, filepath.FromSlash(rel))
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return "", false
	}

	return abs, true
}

// GetRelPath returns the module-relative display path for a profile file path.
// When the path does not belong to this module it is returned unchanged.
func (r *SrcResolver) GetRelPath(profilePath string) string {
	if r.modulePath == "" || !strings.HasPrefix(profilePath, r.modulePath) {
		return profilePath
	}

	rest := profilePath[len(r.modulePath):]
	if rest == "" || rest[0] != '/' {
		return profilePath
	}

	return strings.Trim(rest, "/")
}
