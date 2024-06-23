package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/margostino/babel-agent/internal/common"
)

var BASE_URL = "https://api.openai.com/v1"
var CHAT_COMPLETION_PATH = "/chat/completions"
var MODEL = "gpt-4o"
var PROMPT_FILE_PATH = "../../prompts/metadata_enricher.yml"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormatType string

const (
	JSON_OBJECT ResponseFormatType = "json_object"
)

type ResponseFormat struct {
	Type ResponseFormatType `json:"type"`
}

type RequestBody struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	ResponseFormat ResponseFormat `json:"response_format"`
}

type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

type ApiResponse struct {
	Choices []Choice `json:"choices"`
}

func getPrompt() (string, error) {
	content, err := os.ReadFile(PROMPT_FILE_PATH)
	common.Check(err, "Failed to read metadata file")

	var data map[string]interface{}
	err = yaml.Unmarshal(content, &data)
	common.Check(err, "Failed to unmarshal metadata file")

	prompt, ok := data["prompt"].(string)

	if !ok {
		return "", fmt.Errorf("Prompt not found in metadata file")
	}

	return prompt, nil
}

func GetChatCompletionForMetadata(apiKey string, path string, input string) (string, error) {
	apiURL := BASE_URL + CHAT_COMPLETION_PATH
	systemPrompt, err := getPrompt()
	common.Check(err, "Failed to get system prompt")

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("File Path: %s", path),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("File content: %s", input),
		},
	}

	requestBody := RequestBody{
		Model:    MODEL,
		Messages: messages,
		ResponseFormat: ResponseFormat{
			Type: JSON_OBJECT,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	common.Check(err, "Failed to marshal OpenAI request body")

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	common.Check(err, "Failed to create OpenAI request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	common.Check(err, "Failed to send OpenAI request")
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return "", fmt.Errorf("OpenAI request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	common.Check(err, "Failed to read OpenAI response body")

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	common.Check(err, "Failed to unmarshal OpenAI response body")

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("No choices found in OpenAI response")
	}

	return apiResponse.Choices[0].Message.Content, nil
}
