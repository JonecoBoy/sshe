package dotfiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// DeployDotfiles backs up current ~/.config folder and copies new configuration files.
func DeployDotfiles(sourceDir, targetConfigDir string, dryRun bool, logFunc func(string)) error {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		logFunc(fmt.Sprintf("Warning: Dotfiles source directory '%s' does not exist. Skipping deployment.", sourceDir))
		return nil
	}

	// 1. Create backup folder with timestamp if target directory exists
	timestamp := time.Now().Format("20060102_150405")
	backupDir := fmt.Sprintf("%s_backup_%s", targetConfigDir, timestamp)

	if !dryRun {
		logFunc(fmt.Sprintf("Creating backup of ~/.config at: %s", backupDir))
		if err := copyDir(targetConfigDir, backupDir); err != nil {
			logFunc(fmt.Sprintf("Warning creating backup (might be first run): %v", err))
		}
	} else {
		logFunc(fmt.Sprintf("[DRY-RUN] Backup of '%s' would be saved to: '%s'", targetConfigDir, backupDir))
	}

	// 2. Copy new dotfiles to target directory
	logFunc(fmt.Sprintf("Copying dotfiles from '%s' to '%s'", sourceDir, targetConfigDir))
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(targetConfigDir, relPath)

		if info.IsDir() {
			if !dryRun {
				return os.MkdirAll(destPath, info.Mode())
			}
			return nil
		}

		if dryRun {
			logFunc(fmt.Sprintf("[DRY-RUN] Would copy: %s -> %s", path, destPath))
			return nil
		}

		return copyFile(path, destPath)
	})

	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
