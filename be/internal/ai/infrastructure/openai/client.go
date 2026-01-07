package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ngasihtau/internal/ai/domain"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

var (
	ErrRateLimited        = errors.New("rate limited by OpenAI API")
	ErrContextTooLong     = errors.New("context length exceeded")
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrServiceUnavailable = errors.New("OpenAI service unavailable")
)

type Config struct {
	APIKey         string
	EmbeddingModel string
	ChatModel      string
	MaxRetries     int
	RetryDelay     time.Duration
}

type Client struct {
	client         *openai.Client
	embeddingModel string
	chatModel      string
	maxRetries     int
	retryDelay     time.Duration
}

func NewClient(cfg Config) *Client {
	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	chatModel := cfg.ChatModel
	if chatModel == "" {
		chatModel = "gpt-4"
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second
	}

	return &Client{
		client:         openai.NewClient(cfg.APIKey),
		embeddingModel: embeddingModel,
		chatModel:      chatModel,
		maxRetries:     maxRetries,
		retryDelay:     retryDelay,
	}
}

func (c *Client) retryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := c.retryDelay

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay *= 2 // exponential backoff
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 429:
			return true
		case 500, 502, 503, 504:
			return true
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

func wrapOpenAIError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 401:
			return fmt.Errorf("%w: %s", ErrInvalidAPIKey, apiErr.Message)
		case 429:
			return fmt.Errorf("%w: %s", ErrRateLimited, apiErr.Message)
		case 400:
			if strings.Contains(apiErr.Message, "context_length") ||
				strings.Contains(apiErr.Message, "maximum context") {
				return fmt.Errorf("%w: %s", ErrContextTooLong, apiErr.Message)
			}
		case 500, 502, 503, 504:
			return fmt.Errorf("%w: %s", ErrServiceUnavailable, apiErr.Message)
		}
	}

	return err
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var embedding []float32

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(c.embeddingModel),
			Input: []string{text},
		})
		if err != nil {
			return err
		}

		if len(resp.Data) == 0 {
			return fmt.Errorf("no embedding returned")
		}

		embedding = resp.Data[0].Embedding
		return nil
	})

	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	return embedding, nil
}

func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	var embeddings [][]float32

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(c.embeddingModel),
			Input: texts,
		})
		if err != nil {
			return err
		}

		embeddings = make([][]float32, len(resp.Data))
		for i, data := range resp.Data {
			embeddings[i] = data.Embedding
		}

		return nil
	})

	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	return embeddings, nil
}

func (c *Client) GenerateResponse(ctx context.Context, systemPrompt, userQuery string, contextChunks []string) (string, error) {
	var response string

	contextText := ""
	if len(contextChunks) > 0 {
		contextText = "Context:\n" + strings.Join(contextChunks, "\n\n---\n\n")
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	if contextText != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: contextText,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userQuery,
	})

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:               c.chatModel,
			Messages:            messages,
			MaxCompletionTokens: 1000,
			Temperature:         0.7,
		})
		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no response returned")
		}

		response = resp.Choices[0].Message.Content
		return nil
	})

	if err != nil {
		return "", wrapOpenAIError(err)
	}

	return response, nil
}

func (c *Client) GenerateSuggestions(ctx context.Context, content string, existingQuestions []string) ([]string, error) {
	var questions []string

	systemPrompt := `You are a helpful assistant that generates study questions based on learning material content.
Generate 3-5 thoughtful questions that would help a student understand the material better.
Questions should:
- Be specific to the content provided
- Progress from basic understanding to deeper analysis
- Encourage critical thinking
- Be clear and concise

Return only the questions, one per line, without numbering or bullet points.`

	userPrompt := fmt.Sprintf("Based on this learning material content, generate study questions:\n\n%s", content)
	if len(existingQuestions) > 0 {
		userPrompt += fmt.Sprintf("\n\nThe student has already asked these questions, so generate different ones that explore other aspects of the material:\n%s", strings.Join(existingQuestions, "\n"))
	}

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: c.chatModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userPrompt},
			},
			MaxCompletionTokens: 500,
			Temperature:         0.8,
		})
		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no suggestions returned")
		}

		lines := strings.Split(resp.Choices[0].Message.Content, "\n")
		questions = nil
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.TrimLeft(line, "0123456789.-•) ")
			line = strings.TrimSpace(line)
			if line != "" && len(line) > 10 { // Filter out very short lines
				questions = append(questions, line)
			}
		}

		if len(questions) > 5 {
			questions = questions[:5]
		}

		return nil
	})

	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	return questions, nil
}

// GenerateQuestions generates quiz questions from material content.
// Implements requirements 12.2, 12.3, 12.4, 12.5.
func (c *Client) GenerateQuestions(ctx context.Context, content string, count int, questionType string) ([]domain.GeneratedQuestion, error) {
	var questions []domain.GeneratedQuestion

	systemPrompt := buildQuestionGenerationPrompt(questionType)
	userPrompt := fmt.Sprintf("Based on this learning material content, generate %d quiz questions:\n\n%s", count, content)

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: c.chatModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userPrompt},
			},
			MaxCompletionTokens: 2000,
			Temperature:         0.7,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		})
		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no questions returned")
		}

		// Parse JSON response
		var result struct {
			Questions []domain.GeneratedQuestion `json:"questions"`
		}
		if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
			return fmt.Errorf("failed to parse questions response: %w", err)
		}

		questions = result.Questions
		return nil
	})

	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	// Ensure we don't return more than requested
	if len(questions) > count {
		questions = questions[:count]
	}

	return questions, nil
}

// buildQuestionGenerationPrompt builds the system prompt for question generation.
func buildQuestionGenerationPrompt(questionType string) string {
	basePrompt := `You are an expert educator that generates high-quality quiz questions from learning materials.
Generate questions that test understanding of the key concepts in the provided content.

Each question must include:
- A clear, well-formed question
- The correct answer
- A brief explanation of why the answer is correct

Return your response as a JSON object with a "questions" array containing objects with these fields:
- "question": the question text
- "type": the question type (multiple_choice, true_false, or short_answer)
- "options": array of options (only for multiple_choice, must have at least 2 options)
- "answer": the correct answer
- "explanation": explanation of the correct answer

`

	switch questionType {
	case "multiple_choice":
		return basePrompt + `Generate ONLY multiple choice questions. Each question must have 4 options (A, B, C, D) with exactly one correct answer.
The "options" field should contain the 4 options as strings.
The "answer" field should contain the correct option letter and text (e.g., "A. The correct answer").`

	case "true_false":
		return basePrompt + `Generate ONLY true/false questions.
The "options" field should be ["True", "False"].
The "answer" field should be either "True" or "False".`

	case "short_answer":
		return basePrompt + `Generate ONLY short answer questions that require a brief written response.
Do not include an "options" field for short answer questions.
The "answer" field should contain the expected answer.`

	default: // mixed
		return basePrompt + `Generate a mix of question types: multiple choice, true/false, and short answer.
For multiple choice: include 4 options and specify the correct one.
For true/false: include ["True", "False"] as options.
For short answer: do not include options.`
	}
}
