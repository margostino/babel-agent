package tools

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/pkg/common"
)

func walkFiles(root string, skipNamesMap map[string]struct{}) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, found := skipNamesMap[info.Name()]; found {
				return filepath.SkipDir
			}
		} else {
			if _, found := skipNamesMap[info.Name()]; !found {
				normalizedFileName := normalizeFileName(info.Name())
				if normalizedFileName != info.Name() {
					newPath := filepath.Join(filepath.Dir(path), normalizedFileName)
					if err := os.Rename(path, newPath); err != nil {
						log.Fatalf("Error renaming file: %v\n", err)
					} else {
						log.Printf("Renamed file: %s to %s\n", path, newPath)
					}
				}
			}
		}
		return nil
	})
}

func normalizeFileName(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	normalized := common.NewString(base).ToLower().
		ReplaceAll(" ", "_").
		ReplaceAll(".", "_").
		Value()
	return normalized
}

func listToMap(list []string) map[string]struct{} {
	m := make(map[string]struct{}, len(list))
	for _, v := range list {
		m[v] = struct{}{}
	}
	return m
}
func AssetsCleaner(config *config.Config) (bool, error) {
	log.Println("Running AssetsCleaner tool...")
	skipNames := []string{".git", "0-description"}
	skipNamesMap := listToMap(skipNames)

	if err := walkFiles(config.Repository.Path, skipNamesMap); err != nil {
		log.Printf("Error walking through the path: %v\n", err)
	}

	return true, nil

}
