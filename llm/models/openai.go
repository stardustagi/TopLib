package models

type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // e.g., "Hello!"
}

type ChatRequest struct {
	Model       string        `json:"model"`                 // e.g., "gpt-4"
	Messages    []ChatMessage `json:"messages"`              // message history
	Temperature float32       `json:"temperature,omitempty"` // creativity level
	TopP        float32       `json:"top_p,omitempty"`       // optional
	MaxTokens   int           `json:"max_tokens,omitempty"`  // optional
	Stream      bool          `json:"stream,omitempty"`      // for streaming
	User        string        `json:"user,omitempty"`        // user identifier
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
