package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/margostino/babel-agent/internal/common"
	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/openai"
	"github.com/margostino/babel-agent/internal/utils"
)

func getRelativePath(absolutePath string) (string, error) {
	const pattern = ".babel/db"
	absolutePath = filepath.Clean(absolutePath)
	index := strings.Index(absolutePath, pattern)
	if index == -1 {
		return "", fmt.Errorf("pattern %s not found in path %s", pattern, absolutePath)
	}
	relativePath := absolutePath[index+len(pattern)+1:] // +1 to remove the trailing separator
	return relativePath, nil
}

func writePrettyJSONToFile(jsonString, filePath string) {
	// Unmarshal the JSON string into a map
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	common.Check(err, "Failed to unmarshal JSON")

	// Marshal the map into a pretty-formatted JSON string
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	common.Check(err, "Failed to marshal JSON")

	dir := filepath.Dir(filePath)
	err = os.MkdirAll(dir, 0755)
	common.Check(err, "Failed to create directory")

	err = os.WriteFile(fmt.Sprintf("%s.json", filePath), prettyJSON, 0644)
	common.Check(err, "Failed to write JSON to file")
}

func walkAndEnrichMetadata(root string, skipNamesMap map[string]struct{}, openAiAPIKey string) error {
	metadataPath := filepath.Join(root, "metadata")
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
				relativePath, err := filepath.Rel(root, path)
				common.Check(err, "Failed to get relative path")
				metadataFilePath := filepath.Join(metadataPath, relativePath)
				metadataDir := filepath.Dir(metadataFilePath)

				// Ensure the metadata directory exists
				if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
					if err := os.MkdirAll(metadataDir, os.ModePerm); err != nil {
						return err
					}
				}

				content, err := os.ReadFile(path)
				common.Check(err, "Failed to read file content")

				relativeFilePath, err := getRelativePath(path)
				common.Check(err, "Failed to get relative path")

				metadata, err := openai.GetChatCompletionForMetadata(openAiAPIKey, relativeFilePath, string(content))
				common.Check(err, "Failed to get metadata")

				if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
					log.Printf("Created metadata for %s\n", path)
				} else {
					log.Printf("Updated Metadata for %s\n", path)
				}

				writePrettyJSONToFile(metadata, metadataFilePath)
			}
		}
		return nil
	})
}

func EnrichMetadata(config *config.Config, relativeFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(fmt.Sprintf("Running MetadataEnrichment tool for file: %s", relativeFilePath))

	//skipNames := []string{".git", "metadata", "0-description", "0-babel", "metadata_index"}
	//skipNamesMap := utils.ListToMap(skipNames)
	openAiAPIKey := config.OpenAi.ApiKey

	root := config.Repository.Path
	metadataPath := filepath.Join(root, "metadata")
	absoluteFilePath := filepath.Join(root, relativeFilePath)
	metadataFilePath := filepath.Join(metadataPath, relativeFilePath)

	content, err := os.ReadFile(absoluteFilePath)
	common.Check(err, "Failed to read file content")

	metadata, err := openai.GetChatCompletionForMetadata(openAiAPIKey, relativeFilePath, string(content))
	common.Check(err, "Failed to get metadata")

	writePrettyJSONToFile(metadata, metadataFilePath)
}

func EnrichMetadataInBulk(config *config.Config) {
	log.Println(fmt.Sprintf("Running MetadataEnrichment tool in bulk..."))
	skipNames := []string{".git", "metadata", "0-description", "0-babel", "metadata_index"}
	skipNamesMap := utils.ListToMap(skipNames)
	openAiAPIKey := config.OpenAi.ApiKey

	if err := walkAndEnrichMetadata(config.Repository.Path, skipNamesMap, openAiAPIKey); err != nil {
		log.Printf("Error walking through the path: %v\n", err)
	}
}
