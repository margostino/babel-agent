package tools

import (
	"log"
	"os"
	"path/filepath"
	"strings"

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
	normalized := common.NewString(base).
		TrimSpace().
		ToLower().
		ReplaceAll(" ", "_").
		ReplaceAll(".", "_").
		Value()
	return normalized
}

func CleanAssets(config *config.Config, relativeFilePath string) string {
	// log.Println(fmt.Sprintf("Running AssetsCleaner tool for file: %s", relativeFilePath))

	skipNames := []string{".git", "0-description", "0-babel", "metadata_index", "metadata"}
	skipNamesMap := utils.ListToMap(skipNames)

	root := config.Repository.Path
	oldPath := filepath.Join(root, relativeFilePath)

	info, err := os.Stat(oldPath)
	if os.IsNotExist(err) {
		log.Fatalf("path does not exist: %v", err)
	} else if err != nil {
		log.Fatalf("failed to get file info: %v", err)
	}
	if info.IsDir() {
		return relativeFilePath
	}

	if _, found := skipNamesMap[info.Name()]; !found {
		pathParts := common.NewString(relativeFilePath).Split("/").Values()
		filename := pathParts[len(pathParts)-1]
		normalizedFileName := normalizeFileName(filename)
		if normalizedFileName != filename {
			normalizedFilePath := common.NewString(relativeFilePath).ReplaceAll(filename, normalizedFileName).Value()
			newPath := common.NewString(oldPath).ReplaceAll(filename, normalizedFileName).Value()
			//newPath := filepath.Join(root, normalizedFileName)
			if err := os.Rename(oldPath, newPath); err != nil {
				common.Check(err, "Failed to rename file")
			} else {
				log.Printf("Renamed file: %s to %s\n", oldPath, newPath)
			}
			return normalizedFilePath
		}
	}

	return relativeFilePath
}
