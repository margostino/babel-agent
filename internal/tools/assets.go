package tools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/utils"

	"github.com/margostino/babel-agent/internal/common"
)

func walkAndNormalizeFiles(root string, skipNamesMap map[string]struct{}) error {
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

func CleanAssets(config *config.Config, relativeFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(fmt.Sprintf("Running AssetsCleaner tool for file: %s", relativeFilePath))

	root := config.Repository.Path
	oldPath := filepath.Join(root, relativeFilePath)
	pathParts := common.NewString(relativeFilePath).Split("/").Values()
	filename := pathParts[len(pathParts)-1]
	normalizedFileName := normalizeFileName(filename)
	if normalizedFileName != filename {
		newPath := filepath.Join(root, relativeFilePath, normalizedFileName)
		if err := os.Rename(oldPath, newPath); err != nil {
			log.Fatalf("Error renaming file: %v\n", err)
		} else {
			log.Printf("Renamed file: %s to %s\n", oldPath, newPath)
		}
	}
}

func CleanAssetsInBulk(config *config.Config) {
	log.Println("Running AssetsCleaner tool...")
	skipNames := []string{".git", "0-description", "0-babel", "metadata_index"}
	skipNamesMap := utils.ListToMap(skipNames)

	if err := walkAndNormalizeFiles(config.Repository.Path, skipNamesMap); err != nil {
		log.Printf("Error walking through the path: %v\n", err)
	}
}
