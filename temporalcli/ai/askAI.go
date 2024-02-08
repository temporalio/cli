package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ResponseBody struct {
	Responses []Message `json:"responses"`
}

func SendQuestion(question string) {
	requestBody := RequestBody{
		Messages: []Message{
			{Role: "system", Content: "Provide factually correct information on Temporal.io, the durable execution framework."},
			{Role: "user", Content: question},
		},
		Temperature: 0.7,
		MaxTokens:   -1,
		Stream:      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling request data:", err)
		return
	}

	resp, err := http.Post("http://localhost:1234/v1/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		fmt.Println("Error unmarshaling response body:", err)
		return
	}

	if len(chatResponse.Choices) > 0 {
		fmt.Printf("Your assistant says: %s\n", chatResponse.Choices[0].Message.Content)
	}
}
