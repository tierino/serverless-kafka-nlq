package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/marcboeker/go-duckdb"
)

func extractQuestion(body string) (*string, error) {
	var payload struct {
		Question string `json:"question"`
	}

	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, err
	}

	if payload.Question == "" {
		return nil, errors.New("expected a 'question' field")
	}

	return &payload.Question, nil
}

func stringifyResults(results []map[string]any) (*string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		return nil, err
	}
	formatted := buf.String()
	return &formatted, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	os.Setenv("HOME", "/tmp")

	question, err := extractQuestion(request.Body)
	if err != nil {
		log.Fatal(err)
	}

	rawQuery, err := textToSql(*question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("LLM generated raw query: %s", *rawQuery)

	results, err := runQuery(*rawQuery)
	if err != nil {
		log.Fatal(err)
	}

	stringified, err := stringifyResults(results)
	if err != nil {
		log.Fatal(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       *stringified,
	}, nil
}

func main() {
	lambda.Start(handler)
}
