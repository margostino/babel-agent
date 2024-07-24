package db

import (
	"context"
	"log"

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
	ready, err := client.Misc().ReadyChecker().Do(context.Background())
	if err != nil {
		log.Println("Weaviate is NOT ready yet")
	} else {
		log.Printf("Weaviate is ready: %t\n", ready)
	}
	return client
}
