package entity

type LLMChat struct {
	Prompt string
	Model  string
}

type LLMResponse struct {
	Content string
	Model   string
}
