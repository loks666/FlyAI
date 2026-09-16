package webui

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Assets exposes the static Next.js export generated in frontend/dist.
// Files stay on disk so a frontend rebuild does not require rebuilding Go.
type Assets struct {
	root string
}

// Load resolves and validates the frontend distribution directory.
func Load() (*Assets, error) {
	candidates := make([]string, 0, 5)
	if configured := strings.TrimSpace(os.Getenv("WEB_DIST_PATH")); configured != "" {
		candidates = append(candidates, configured)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "webui", "frontend", "dist"),
			filepath.Join(cwd, "dist"),
		)
	}
	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDir, "webui", "frontend", "dist"),
			filepath.Join(executableDir, "dist"),
		)
	}

	checked := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		absolute = filepath.Clean(absolute)
		key := strings.ToLower(absolute)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		checked = append(checked, absolute)
		if isFile(filepath.Join(absolute, "index.html")) && isDirectory(filepath.Join(absolute, "_next")) {
			return &Assets{root: absolute}, nil
		}
	}

	return nil, fmt.Errorf("未找到前端 dist（需要 index.html 和 _next），已检查: %s", strings.Join(checked, ", "))
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Root returns the validated absolute distribution directory.
func (assets *Assets) Root() string {
	return assets.root
}

// Read returns one generated page or public asset.
func (assets *Assets) Read(name string) ([]byte, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("非法前端资源路径: %q", name)
	}
	return os.ReadFile(filepath.Join(assets.root, clean))
}

// NextAssets returns a filesystem rooted at Next.js' generated asset folder.
func (assets *Assets) NextAssets() http.FileSystem {
	return http.Dir(filepath.Join(assets.root, "_next"))
}
