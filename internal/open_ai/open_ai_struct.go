package open_ai

type Request struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
	MaxTokens int `json:"max_tokens"`
	Input string `json:"input"`
}

type Prompt struct {
	Text string
	ThreadId string
}