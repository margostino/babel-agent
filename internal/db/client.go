package db

import (
	"context"
	"fmt"
	"log"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func NewDBClient(openAiApiKey string, port int) *weaviate.Client {
	cfg := weaviate.Config{
		Host:   fmt.Sprintf("%s:%d", "localhost", port),
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": openAiApiKey,
		},
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	_, err = client.Misc().ReadyChecker().Do(context.Background())
	if err != nil {
		log.Println("Weaviate is NOT ready yet")
	} else {
		// log.Printf("Weaviate is ready: %t\n", ready)
	}
	return client
}
