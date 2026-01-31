// Package application contains property-based tests for generated questions structure.
package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/ai/domain"
)

// **Feature: user-storage-limit, Property 12: Generated Questions Structure**
//
// *For any* question generation request with count N:
// - Output SHALL contain exactly min(N, 20) questions
// - Each question SHALL have non-empty `question`, `answer`, and `explanation` fields
// - Multiple choice questions SHALL have at least 2 options
//
// **Validates: Requirements 12.2, 12.3, 12.4, 12.5**

func TestProperty_GeneratedQuestionsStructure(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 12.1: Output contains exactly min(N, 20) questions
	// Validates: Requirement 12.3 - WHEN generating questions THEN THE Storage_Limit_System
	// SHALL return a configurable number of questions (default: 5, max: 20)
	properties.Property("output contains exactly min(N, MaxQuestionCount) questions", prop.ForAll(
		func(requestedCount int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			// Add content to vector repo
			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "This is sample content for question generation.",
				},
			}

			// Configure mock to return the requested number of questions
			expectedCount := requestedCount
			if expectedCount <= 0 {
				expectedCount = DefaultQuestionCount
			}
			if expectedCount > MaxQuestionCount {
				expectedCount = MaxQuestionCount
			}

			chatClient.questions = generateMockQuestions(expectedCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        requestedCount,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			return len(result.Questions) == expectedCount
		},
		gen.IntRange(-5, 30), // Test range including negative, zero, and above max
	))

	// Property 12.2: Each question has non-empty question field
	// Validates: Requirement 12.2 - WHEN a pro user requests question generation
	// THEN THE Storage_Limit_System SHALL generate questions based on material content
	properties.Property("each question has non-empty question field", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			// Ensure count is within valid range
			expectedCount := count
			if expectedCount <= 0 {
				expectedCount = DefaultQuestionCount
			}
			if expectedCount > MaxQuestionCount {
				expectedCount = MaxQuestionCount
			}

			chatClient.questions = generateMockQuestions(expectedCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			for _, q := range result.Questions {
				if q.Question == "" {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 12.3: Each question has non-empty answer field
	// Validates: Requirement 12.5 - WHEN generating questions THEN THE Storage_Limit_System
	// SHALL include the correct answer and explanation for each question
	properties.Property("each question has non-empty answer field", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			expectedCount := count
			if expectedCount <= 0 {
				expectedCount = DefaultQuestionCount
			}
			if expectedCount > MaxQuestionCount {
				expectedCount = MaxQuestionCount
			}

			chatClient.questions = generateMockQuestions(expectedCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			for _, q := range result.Questions {
				if q.Answer == "" {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 12.4: Each question has non-empty explanation field
	// Validates: Requirement 12.5 - WHEN generating questions THEN THE Storage_Limit_System
	// SHALL include the correct answer and explanation for each question
	properties.Property("each question has non-empty explanation field", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			expectedCount := count
			if expectedCount <= 0 {
				expectedCount = DefaultQuestionCount
			}
			if expectedCount > MaxQuestionCount {
				expectedCount = MaxQuestionCount
			}

			chatClient.questions = generateMockQuestions(expectedCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			for _, q := range result.Questions {
				if q.Explanation == "" {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 12.5: Multiple choice questions have at least 2 options
	// Validates: Requirement 12.4 - WHEN generating questions THEN THE Storage_Limit_System
	// SHALL support multiple question types (multiple_choice, true_false, short_answer)
	properties.Property("multiple choice questions have at least 2 options", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			expectedCount := count
			if expectedCount <= 0 {
				expectedCount = DefaultQuestionCount
			}
			if expectedCount > MaxQuestionCount {
				expectedCount = MaxQuestionCount
			}

			// Generate only multiple choice questions
			chatClient.questions = generateMockQuestions(expectedCount, QuestionTypeMultipleChoice)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMultipleChoice,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			for _, q := range result.Questions {
				if q.Type == QuestionTypeMultipleChoice && len(q.Options) < 2 {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 12.6: Default count is applied when count is zero or negative
	// Validates: Requirement 12.3 - default: 5
	properties.Property("default count is applied when count is zero or negative", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			// For zero or negative count, expect default count
			chatClient.questions = generateMockQuestions(DefaultQuestionCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			return len(result.Questions) == DefaultQuestionCount
		},
		gen.IntRange(-10, 0), // Only test zero and negative values
	))

	// Property 12.7: Max count is enforced when count exceeds maximum
	// Validates: Requirement 12.3 - max: 20
	properties.Property("max count is enforced when count exceeds maximum", prop.ForAll(
		func(count int) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			// For count > max, expect max count
			chatClient.questions = generateMockQuestions(MaxQuestionCount, QuestionTypeMixed)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        count,
				QuestionType: QuestionTypeMixed,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			return len(result.Questions) == MaxQuestionCount
		},
		gen.IntRange(21, 100), // Only test values above max
	))

	// Property 12.8: Question type is correctly set for each question
	// Validates: Requirement 12.4 - support multiple question types
	properties.Property("question type is correctly set for each question", prop.ForAll(
		func(questionType string) bool {
			svc, vectorRepo, chatClient := newTestQuestionService()
			ctx := context.Background()

			userID := uuid.New()
			materialID := uuid.New()
			podID := uuid.New()

			vectorRepo.chunks = []domain.MaterialChunk{
				{
					ID:         "chunk1",
					MaterialID: materialID,
					PodID:      podID,
					ChunkIndex: 0,
					Text:       "Sample content for testing.",
				},
			}

			chatClient.questions = generateMockQuestions(5, questionType)

			input := GenerateQuestionsInput{
				UserID:       userID,
				MaterialID:   materialID,
				Count:        5,
				QuestionType: questionType,
			}

			result, err := svc.GenerateQuestions(ctx, input)
			if err != nil {
				return false
			}

			// For non-mixed types, all questions should have the specified type
			if questionType != QuestionTypeMixed {
				for _, q := range result.Questions {
					if q.Type != questionType {
						return false
					}
				}
			}
			return true
		},
		genValidQuestionType(),
	))

	properties.TestingRun(t)
}

// Helper function to create test service for question generation
func newTestQuestionService() (*Service, *mockVectorRepo, *mockChatClient) {
	sessionRepo := newMockChatSessionRepo()
	messageRepo := newMockChatMessageRepo()
	vectorRepo := newMockVectorRepo()
	embeddingClient := newMockEmbeddingClient()
	chatClient := newMockChatClient()
	// Use a mock that allows all operations (returns nil for all checks)
	aiLimitChecker := newMockAILimitChecker()

	svc := NewService(
		sessionRepo,
		messageRepo,
		vectorRepo,
		embeddingClient,
		chatClient,
		"http://localhost:8000",
		aiLimitChecker,
	)

	return svc, vectorRepo, chatClient
}

// Helper function to generate mock questions
func generateMockQuestions(count int, questionType string) []domain.GeneratedQuestion {
	questions := make([]domain.GeneratedQuestion, count)
	for i := 0; i < count; i++ {
		qType := questionType
		if questionType == QuestionTypeMixed {
			// Rotate through types for mixed
			types := []string{QuestionTypeMultipleChoice, QuestionTypeTrueFalse, QuestionTypeShortAnswer}
			qType = types[i%len(types)]
		}

		q := domain.GeneratedQuestion{
			Question:    "What is the main concept discussed in the material?",
			Type:        qType,
			Answer:      "The correct answer is A",
			Explanation: "This is the explanation for why this answer is correct.",
		}

		// Add options for multiple choice questions
		if qType == QuestionTypeMultipleChoice {
			q.Options = []string{"A. Option 1", "B. Option 2", "C. Option 3", "D. Option 4"}
		} else if qType == QuestionTypeTrueFalse {
			q.Options = []string{"True", "False"}
		}

		questions[i] = q
	}
	return questions
}

// Generator for valid question types
func genValidQuestionType() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		types := []string{QuestionTypeMultipleChoice, QuestionTypeTrueFalse, QuestionTypeShortAnswer, QuestionTypeMixed}
		idx := params.Rng.Intn(len(types))
		return gopter.NewGenResult(types[idx], gopter.NoShrinker)
	})
}
