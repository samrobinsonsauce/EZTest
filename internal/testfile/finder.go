package testfile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TestFile struct {
	Path         string
	AbsolutePath string
}

func FindTestFiles(rootDir string) ([]TestFile, error) {
	testDir := filepath.Join(rootDir, "test")

	info, err := os.Stat(testDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("no 'test/' directory found in %s\nAre you in an Elixir/Phoenix project?", rootDir)
	}
	if err != nil {
		return nil, fmt.Errorf("error accessing test directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("'test' exists but is not a directory")
	}

	var testFiles []TestFile

	err = filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == "support" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, "_test.exs") {
			return nil
		}

		if d.Name() == "test_helper.exs" {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			relPath = path
		}

		testFiles = append(testFiles, TestFile{
			Path:         relPath,
			AbsolutePath: path,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning test directory: %w", err)
	}

	if len(testFiles) == 0 {
		return nil, fmt.Errorf("no test files (*_test.exs) found in %s", testDir)
	}

	sort.Slice(testFiles, func(i, j int) bool {
		return testFiles[i].Path < testFiles[j].Path
	})

	return testFiles, nil
}
