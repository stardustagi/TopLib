package clients

import (
	"encoding/json"
	"os"

	"github.com/stardustagi/TopLib/llm/models"
	"resty.dev/v3"
)

func SendChatRequest(req *models.ChatRequest) (*models.ChatResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	url := "https://api.openai.com/v1/chat/completions"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var chatResp models.ChatResponse
	err = json.Unmarshal(resp.Bytes(), &chatResp)
	if err != nil {
		return nil, err
	}

	return &chatResp, nil
}
