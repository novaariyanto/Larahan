package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts a zip archive into destDir.
// When flattenSingleRoot is true and the archive has a single top-level folder,
// that folder's contents are moved up into destDir.
func ExtractZip(zipPath, destDir string, flattenSingleRoot bool) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		if err := extractFile(f, destAbs); err != nil {
			return err
		}
	}

	if flattenSingleRoot {
		if err := flattenIfSingleRoot(destAbs); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(f *zip.File, destAbs string) error {
	name := filepath.Clean(f.Name)
	if name == "." || name == "" {
		return nil
	}
	// Prevent ZipSlip
	target := filepath.Join(destAbs, name)
	if !strings.HasPrefix(target, destAbs+string(os.PathSeparator)) && target != destAbs {
		return fmt.Errorf("illegal zip path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func flattenIfSingleRoot(destAbs string) error {
	entries, err := os.ReadDir(destAbs)
	if err != nil {
		return err
	}
	dirs := make([]os.DirEntry, 0)
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}
	if len(entries) != 1 || len(dirs) != 1 {
		return nil
	}

	rootName := dirs[0].Name()
	rootPath := filepath.Join(destAbs, rootName)
	children, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}
	for _, child := range children {
		from := filepath.Join(rootPath, child.Name())
		to := filepath.Join(destAbs, child.Name())
		if err := os.Rename(from, to); err != nil {
			return err
		}
	}
	return os.Remove(rootPath)
}
