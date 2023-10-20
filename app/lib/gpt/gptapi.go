package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const COMPLETIONS_URL = "chat/completions"

var char_gpt_prompt = `
Ниже указан текст новости по MMORPG игре World of Warcraft, подготовленный в формате discord на английском языке.
Его нужно перевести на русский, адаптировав учитывая вселенную. 
После перевода к тексту нужно добавить обзац с общей информацией - выжимкой (Не более 400 символов) по этой новости на русском языке.
Также нужно отформатировать discord теги, чтобы было более читабельно для человека.
Вот текст: `

type OpenaiApiClient struct {
	http_client *http.Client
	baseUrl     string
	api_key     string
	prompt      string
}

func InitOpenaiApiClient(api_key string) (client *OpenaiApiClient) {
	client = &OpenaiApiClient{&http.Client{}, "https://api.openai.com/v1/", api_key, char_gpt_prompt}
	return client
}

func (client *OpenaiApiClient) GetCompletion(prompt string) (string, error) {
	promt_text := client.prompt + prompt
	r := completionsRequest{
		Model: "gpt-3.5-turbo-16k-0613",
		Messages: []Message{
			{
				Role:    "user",
				Content: promt_text,
			},
		},
		Temperature: 1.0,
	}

	rJson, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", client.baseUrl+COMPLETIONS_URL, bytes.NewReader(rJson))
	req.Header.Add("Authorization", "Bearer "+client.api_key)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.http_client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	response, err := io.ReadAll(resp.Body)
	fmt.Println(string(response))
	var rResp completionsResponse
	err = json.Unmarshal(response, &rResp)
	if err != nil {
		return "", err
	}
	return rResp.Choices[0].Message.Content, nil
}
