package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ngasihtau/internal/ai/domain"
	"strings"
	"time"

	"google.golang.org/genai"
)

var (
	ErrRateLimited        = errors.New("rate limited by Gemini API")
	ErrContextTooLong     = errors.New("context length exceeded")
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrServiceUnavailable = errors.New("Gemini service unavailable")
)

type Config struct {
	APIKey         string
	ChatModel      string
	EmbeddingModel string
	MaxRetries     int
	RetryDelay     time.Duration
}

type Client struct {
	client         *genai.Client
	chatModel      string
	embeddingModel string
	maxRetries     int
	retryDelay     time.Duration
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	chatModel := cfg.ChatModel
	if chatModel == "" {
		chatModel = "gemini-1.5-flash"
	}

	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-004"
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client:         client,
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
		maxRetries:     maxRetries,
		retryDelay:     retryDelay,
	}, nil
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

	errStr := err.Error()
	// Check for rate limiting or server errors
	if strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") ||
		strings.Contains(errStr, "RESOURCE_EXHAUSTED") {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

func wrapGeminiError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	if strings.Contains(errStr, "INVALID_ARGUMENT") && strings.Contains(errStr, "API key") {
		return fmt.Errorf("%w: %s", ErrInvalidAPIKey, errStr)
	}
	if strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "429") {
		return fmt.Errorf("%w: %s", ErrRateLimited, errStr)
	}
	if strings.Contains(errStr, "context_length") || strings.Contains(errStr, "token limit") {
		return fmt.Errorf("%w: %s", ErrContextTooLong, errStr)
	}
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return fmt.Errorf("%w: %s", ErrServiceUnavailable, errStr)
	}

	return err
}

func (c *Client) GenerateResponse(ctx context.Context, systemPrompt, userQuery string, contextChunks []string) (string, error) {
	var response string

	contextText := ""
	if len(contextChunks) > 0 {
		contextText = "Context:\n" + strings.Join(contextChunks, "\n\n---\n\n")
	}

	// Build the prompt with system instruction and context
	var parts []*genai.Part

	if contextText != "" {
		parts = append(parts, &genai.Part{Text: contextText})
	}
	parts = append(parts, &genai.Part{Text: userQuery})

	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr[float32](0.7),
		MaxOutputTokens: 1000,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	err := c.retryWithBackoff(ctx, func() error {
		result, err := c.client.Models.GenerateContent(
			ctx,
			c.chatModel,
			[]*genai.Content{{Parts: parts}},
			config,
		)
		if err != nil {
			return err
		}

		if result == nil || len(result.Candidates) == 0 {
			return fmt.Errorf("no response returned")
		}

		response = result.Text()
		return nil
	})

	if err != nil {
		return "", wrapGeminiError(err)
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

	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr[float32](0.8),
		MaxOutputTokens: 500,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	err := c.retryWithBackoff(ctx, func() error {
		result, err := c.client.Models.GenerateContent(
			ctx,
			c.chatModel,
			[]*genai.Content{{Parts: []*genai.Part{{Text: userPrompt}}}},
			config,
		)
		if err != nil {
			return err
		}

		if result == nil || len(result.Candidates) == 0 {
			return fmt.Errorf("no suggestions returned")
		}

		lines := strings.Split(result.Text(), "\n")
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
		return nil, wrapGeminiError(err)
	}

	return questions, nil
}

// GenerateEmbedding generates a vector embedding for the given text using Gemini.
func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var embedding []float32

	err := c.retryWithBackoff(ctx, func() error {
		result, err := c.client.Models.EmbedContent(
			ctx,
			c.embeddingModel,
			[]*genai.Content{{Parts: []*genai.Part{{Text: text}}}},
			nil,
		)
		if err != nil {
			return err
		}

		if result == nil || len(result.Embeddings) == 0 {
			return fmt.Errorf("no embedding returned")
		}

		embedding = result.Embeddings[0].Values
		return nil
	})

	if err != nil {
		return nil, wrapGeminiError(err)
	}

	return embedding, nil
}

// GenerateEmbeddings generates vector embeddings for multiple texts using Gemini.
func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	var embeddings [][]float32

	// Build contents for batch embedding
	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = &genai.Content{Parts: []*genai.Part{{Text: text}}}
	}

	err := c.retryWithBackoff(ctx, func() error {
		result, err := c.client.Models.EmbedContent(
			ctx,
			c.embeddingModel,
			contents,
			nil,
		)
		if err != nil {
			return err
		}

		if result == nil || len(result.Embeddings) == 0 {
			return fmt.Errorf("no embeddings returned")
		}

		embeddings = make([][]float32, len(result.Embeddings))
		for i, emb := range result.Embeddings {
			embeddings[i] = emb.Values
		}

		return nil
	})

	if err != nil {
		return nil, wrapGeminiError(err)
	}

	return embeddings, nil
}

// GenerateQuestions generates quiz questions from material content.
// Implements requirements 12.2, 12.3, 12.4, 12.5.
func (c *Client) GenerateQuestions(ctx context.Context, content string, count int, questionType string) ([]domain.GeneratedQuestion, error) {
	var questions []domain.GeneratedQuestion

	systemPrompt := buildQuestionGenerationPrompt(questionType)
	userPrompt := fmt.Sprintf("Based on this learning material content, generate %d quiz questions:\n\n%s", count, content)

	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr[float32](0.7),
		MaxOutputTokens: 2000,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		ResponseMIMEType: "application/json",
	}

	err := c.retryWithBackoff(ctx, func() error {
		result, err := c.client.Models.GenerateContent(
			ctx,
			c.chatModel,
			[]*genai.Content{{Parts: []*genai.Part{{Text: userPrompt}}}},
			config,
		)
		if err != nil {
			return err
		}

		if result == nil || len(result.Candidates) == 0 {
			return fmt.Errorf("no questions returned")
		}

		// Parse JSON response
		responseText := result.Text()
		var parsed struct {
			Questions []domain.GeneratedQuestion `json:"questions"`
		}
		if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
			return fmt.Errorf("failed to parse questions response: %w", err)
		}

		questions = parsed.Questions
		return nil
	})

	if err != nil {
		return nil, wrapGeminiError(err)
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
