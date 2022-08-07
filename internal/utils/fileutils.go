package utils

import (
	"os"
	"path/filepath"
)

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		info.Mode()
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			if filepath.IsAbs(root) {
				if filepath.Dir(path) == root {
					matches = append(matches, path)
				} else {
					return nil
				}
			} else {
				matches = append(matches, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
