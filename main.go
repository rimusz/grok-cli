package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/sashabaranov/go-openai"
)

const apiKeyEnv = "GROK_API_KEY"
const modelEnv = "GROK_MODEL"

type ToolFunction func(map[string]interface{}) string

var tools = []openai.Tool{
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "calculate",
			Description: "Execute simple math expressions.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"expression": {"type": "string", "description": "The math expression to evaluate (e.g., '2 + 2')."}
				},
				"required": ["expression"]
			}`),
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {"type": "string", "description": "The path to the file to read."}
				},
				"required": ["path"]
			}`),
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "write_file",
			Description: "Write content to a file.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {"type": "string", "description": "The path to the file to write."},
					"content": {"type": "string", "description": "The content to write to the file."}
				},
				"required": ["path", "content"]
			}`),
		},
	},
	// Add more tools here, e.g., for other file operations or external integrations.
}

var toolFunctions = map[string]ToolFunction{
	"calculate": func(params map[string]interface{}) string {
		exprStr, ok := params["expression"].(string)
		if !ok {
			return "Error: invalid expression"
		}
		expr, err := govaluate.NewEvaluableExpression(exprStr)
		if err != nil {
			return "Error: " + err.Error()
		}
		result, err := expr.Evaluate(nil)
		if err != nil {
			return "Error: " + err.Error()
		}
		return fmt.Sprint(result)
	},
	"read_file": func(params map[string]interface{}) string {
		path, ok := params["path"].(string)
		if !ok {
			return "Error: invalid path"
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "Error: " + err.Error()
		}
		return string(content)
	},
	"write_file": func(params map[string]interface{}) string {
		path, ok := params["path"].(string)
		if !ok {
			return "Error: invalid path"
		}
		content, ok := params["content"].(string)
		if !ok {
			return "Error: invalid content"
		}
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return "Error: " + err.Error()
		}
		return "File written successfully"
	},
}

// CustomChatCompletionRequest extends the standard request to include Grok-specific search_parameters
type CustomChatCompletionRequest struct {
	openai.ChatCompletionRequest
	SearchParameters map[string]interface{} `json:"search_parameters,omitempty"`
}

func main() {
	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		fmt.Println("Please set the GROK_API_KEY environment variable.")
		return
	}

	model := os.Getenv(modelEnv)
	if model == "" {
		model = "grok-4" // Default to grok-4
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.x.ai/v1"

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful AI agent."},
	}

	fmt.Println("Grok CLI: Enter your query (type 'exit' to quit). Tools and live search are enabled.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		for {
			// Marshal the custom request
			req := CustomChatCompletionRequest{
				ChatCompletionRequest: openai.ChatCompletionRequest{
					Model:      model,
					Messages:   messages,
					Tools:      tools,
					ToolChoice: "auto",
				},
				SearchParameters: map[string]interface{}{
					"mode":             "auto",
					"return_citations": true,
				},
			}
			reqBody, err := json.Marshal(req)
			if err != nil {
				fmt.Println("Error marshaling request:", err)
				break
			}

			// Create HTTP request
			httpReq, err := http.NewRequestWithContext(context.Background(), "POST", config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
			if err != nil {
				fmt.Println("Error creating request:", err)
				break
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+apiKey)

			// Send request
			httpResp, err := http.DefaultClient.Do(httpReq)
			if err != nil {
				fmt.Println("Error sending request:", err)
				break
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(httpResp.Body)
				fmt.Println("Error:", httpResp.Status, string(body))
				break
			}

			var resp openai.ChatCompletionResponse
			if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
				fmt.Println("Error decoding response:", err)
				break
			}

			choice := resp.Choices[0]
			if choice.FinishReason == "tool_calls" {
				messages = append(messages, choice.Message)

				for _, toolCall := range choice.Message.ToolCalls {
					funcName := toolCall.Function.Name
					var params map[string]interface{}
					json.Unmarshal([]byte(toolCall.Function.Arguments), &params)
					result := toolFunctions[funcName](params)
					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						ToolCallID: toolCall.ID,
						Name:       funcName,
						Content:    result,
					})
				}
			} else {
				messages = append(messages, choice.Message)
				fmt.Println("Grok:", choice.Message.Content)
				// Handle citations if present (e.g., parse from response or content)
				break
			}
		}
	}
}
