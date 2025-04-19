package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

func replaceTableNames(query string) string {
	tripsS3Table := fmt.Sprintf("'s3://%s/topics/trips/*/*/*/*/*.snappy.parquet'", os.Getenv("LAKE_BUCKET_NAME"))
	stationsS3Table := fmt.Sprintf("'s3://%s/topics/stations/*/*/*/*/*.snappy.parquet'", os.Getenv("LAKE_BUCKET_NAME"))

	return strings.Replace(
		strings.Replace(query, "trips_tablename", tripsS3Table, -1),
		"stations_tablename", stationsS3Table, -1,
	)
}

func textToSql(prompt string) (*string, error) {
	ctx := context.Background()
	model := "anthropic.claude-3-5-sonnet-20241022-v2:0"

	schemaPath := "schema.json"

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("could not find schema file: %w", err)
	}
	schemaString := string(content)

	llm, err := bedrock.New(bedrock.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("couldn't initialise LLM: %w", err)
	}

	systemPrompt := fmt.Sprintf("You are an expert SQL assistant. Answer with raw unformatted queries only, based on the following database schema: %s", schemaString)
	humanPrompt := fmt.Sprintf("Write a query that answers the question: %s", prompt)

	response, err := llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, humanPrompt),
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't generate response: %w", err)
	}

	readyQuery := replaceTableNames(response.Choices[0].Content)

	return &readyQuery, nil
}
