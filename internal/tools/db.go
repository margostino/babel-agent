package tools

import (
	"context"
	"errors"
	"log"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

func GetObject(dbClient *weaviate.Client, config *config.Config, relativeFilePath string) (*string, error) {
	where := filters.Where().
		WithPath([]string{"path"}).
		WithOperator(filters.Equal).
		WithValueText(relativeFilePath)

	query := dbClient.GraphQL().Get().WithClassName("Babel").
		WithLimit(1).
		WithWhere(where).
		WithFields(
			graphql.Field{Name: "path"},
			graphql.Field{Name: "_additional{id}"},
		)

	// Execute the query
	response, err := query.Do(context.Background())
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}

	// Process the response
	if len(response.Errors) > 0 {
		log.Fatalf("query returned errors: %v", response.Errors)
	}

	results := response.Data["Get"].(map[string]interface{})["Babel"].([]interface{})
	if len(results) > 0 {
		result := results[0].(map[string]interface{})
		additional := result["_additional"].(map[string]interface{})
		id := additional["id"].(string)
		return &id, nil
	}

	return nil, errors.New("No results found")
}

func DeleteObject(dbClient *weaviate.Client, id string) {
	err := dbClient.Data().Deleter().
		WithClassName("Babel").
		WithID(id).
		Do(context.Background())

	if err != nil {
		log.Printf("failed to delete object: %v", err)
	}
}

func UpdateObject(dbClient *weaviate.Client, id string, metadata map[string]interface{}) {
	err := dbClient.Data().Updater().
		WithMerge().
		WithID(id).
		WithClassName("Babel").
		WithProperties(metadata).
		Do(context.Background())

	if err != nil {
		log.Printf("failed to update object: %v", err)
	}
}

func CreateObject(dbClient *weaviate.Client, metadata map[string]interface{}) {

	w, err := dbClient.Data().Creator().
		WithClassName("Babel").
		WithProperties(metadata).
		Do(context.Background())

	if err != nil {
		log.Printf("failed to create object: %v", err)
	} else {
		log.Printf("created object: %v", w.Object.ID)
	}
}
