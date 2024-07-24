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
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
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

func writePrettyJSONToFile(metadataContent string, filePath string, metadataPath string, relativeFilePath string) map[string]interface{} {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(metadataContent), &data)
	common.Check(err, "Failed to unmarshal JSON")

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	common.Check(err, "Failed to marshal JSON")

	dir := filepath.Dir(filePath)
	err = os.MkdirAll(dir, 0755)
	common.Check(err, "Failed to create directory")

	err = os.WriteFile(fmt.Sprintf("%s.json", filePath), prettyJSON, 0644)
	common.Check(err, "Failed to write JSON to file")

	indexFilePath := filepath.Join(metadataPath, "index.json")
	newIndexEntry := map[string]interface{}{
		"highlights": data["highlights"],
		"summary":    data["summary"],
	}
	updateIndexFile(indexFilePath, relativeFilePath, newIndexEntry)
	return data
}

func updateIndexFile(indexFilePath, relativeFilePath string, newIndexEntry map[string]interface{}) {
	indexFileContent, err := os.ReadFile(indexFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to read index file: %v\n", err)
		return
	}

	indexData := make(map[string]map[string]interface{})
	if len(indexFileContent) > 0 {
		err = json.Unmarshal(indexFileContent, &indexData)
		if err != nil {
			log.Printf("Failed to unmarshal index file: %v\n", err)
			return
		}
	}

	if newIndexEntry != nil {
		indexData[relativeFilePath] = newIndexEntry
	} else {
		delete(indexData, relativeFilePath)
	}

	indexJSON, err := json.MarshalIndent(indexData, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal index data to JSON: %v\n", err)
		return
	}

	err = os.WriteFile(indexFilePath, indexJSON, 0644)
	if err != nil {
		log.Printf("Failed to write JSON to index file %s: %v\n", indexFilePath, err)
	}
}

func DeleteMetadata(dbClient *weaviate.Client, id string, config *config.Config, relativeFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	// log.Println(fmt.Sprintf("Running MetadataDeletion tool for file: %s", relativeFilePath))

	root := config.Repository.Path
	metadataPath := filepath.Join(root, "metadata")
	indexFilePath := filepath.Join(metadataPath, "index.json")

	metadataFilePath := fmt.Sprintf("%s.json", filepath.Join(metadataPath, relativeFilePath))

	if _, err := os.Stat(metadataFilePath); !os.IsNotExist(err) {
		err := os.Remove(metadataFilePath)
		common.Check(err, "Failed to remove metadata file")
		log.Printf("Deleted metadata for %s\n", relativeFilePath)
		updateIndexFile(indexFilePath, relativeFilePath, nil)

		DeleteObject(dbClient, id)
	}
}

func EnrichMetadata(dbClient *weaviate.Client, id *string, config *config.Config, relativeFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	// log.Println(fmt.Sprintf("Running MetadataEnrichment tool for file: %s", relativeFilePath))
	openAiAPIKey := config.OpenAi.ApiKey
	root := config.Repository.Path
	absoluteFilePath := filepath.Join(root, relativeFilePath)

	skipNames := []string{".git", "metadata", "0-description", "0-babel", "metadata_index"}
	skipNamesMap := utils.ListToMap(skipNames)

	info, err := os.Stat(absoluteFilePath)
	if os.IsNotExist(err) {
		log.Fatalf("path does not exist: %v", err)
	} else if err != nil {
		log.Fatalf("failed to get file info: %v", err)
	}
	if info.IsDir() {
		log.Printf("skipping directory: %s\n", absoluteFilePath)
		return
	}

	if _, found := skipNamesMap[info.Name()]; !found {
		metadataPath := filepath.Join(root, "metadata")
		metadataFilePath := filepath.Join(metadataPath, relativeFilePath)
		content, err := os.ReadFile(absoluteFilePath)
		common.Check(err, "Failed to read file content")

		metadataContent, err := openai.GetChatCompletionForMetadata(openAiAPIKey, relativeFilePath, string(content))
		common.Check(err, "Failed to get metadata")

		fileContent, err := writePrettyJSONToFile(metadataContent, metadataFilePath, metadataPath, relativeFilePath), nil

		if id == nil {
			CreateObject(dbClient, fileContent)
		} else {
			UpdateObject(dbClient, *id, fileContent)
		}
	}

}
