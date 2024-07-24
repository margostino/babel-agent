package db

import (
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func NewDBClient(openAiApiKey string) *weaviate.Client {
	cfg := weaviate.Config{
		Host:   "localhost:8080",
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": openAiApiKey,
		},
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}
