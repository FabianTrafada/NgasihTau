package application

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/ai/domain"
	"ngasihtau/internal/ai/infrastructure/learningpulse"
)

// Feature constants define the premium and pro features for AI service.
// These are used by the HTTP handler to check tier-based access.
const (
	// FeatureChatExport allows exporting chat conversations.
	// Requires premium tier or higher.
	// Implements requirement 11.1.
	FeatureChatExport = "chat_export"

	// FeaturePodWideChat allows AI chat across all materials in a pod.
	// Requires pro tier.
	// Implements requirement 11.2.
	FeaturePodWideChat = "pod_wide_chat"

	// FeatureQuestionGeneration allows generating quiz questions from materials.
	// Requires pro tier.
	// Implements requirement 12.1.
	FeatureQuestionGeneration = "question_generation"
)

// Question type constants for generated questions.
// Implements requirement 12.4.
const (
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse      = "true_false"
	QuestionTypeShortAnswer    = "short_answer"
	QuestionTypeMixed          = "mixed"
)

// Default and maximum values for question generation.
// Implements requirement 12.3.
const (
	DefaultQuestionCount = 5
	MaxQuestionCount     = 20
)

// AILimitChecker defines the interface for checking and tracking AI usage limits.
// This interface is implemented by the User Service's AIService.
type AILimitChecker interface {
	// CheckAILimit checks if a user has remaining AI messages for today.
	// Returns nil if the user can send AI messages, or an error if the limit is exceeded.
	CheckAILimit(ctx context.Context, userID uuid.UUID) error

	// IncrementAIUsage increments the user's daily AI usage count.
	// Should be called after a successful AI chat message is processed.
	IncrementAIUsage(ctx context.Context, userID uuid.UUID) error

	// CanAccessFeature checks if a user's tier allows access to a specific feature.
	// Returns nil if access is allowed, or an appropriate error if not.
	// Implements requirements 11.1, 11.2, 11.3, 11.4, 12.1, 12.6.
	CanAccessFeature(ctx context.Context, userID uuid.UUID, feature string) error
}

// LearningPulseClient defines the interface for fetching user learning personas.
type LearningPulseClient interface {
	// GetPersona fetches the user's learning persona based on their behavior data.
	GetPersona(ctx context.Context, userID string, behaviorData *learningpulse.BehaviorData) (learningpulse.Persona, error)
}

// BehaviorDataProvider defines the interface for fetching user behavior data.
// This is typically implemented by a repository that aggregates user interactions.
type BehaviorDataProvider interface {
	// GetBehaviorData retrieves aggregated behavior data for a user.
	GetBehaviorData(ctx context.Context, userID uuid.UUID) (*learningpulse.BehaviorData, error)
}

type Service struct {
	chatSessionRepo      ChatSessionRepository
	chatMessageRepo      ChatMessageRepository
	vectorRepo           VectorRepository
	embeddingClient      EmbeddingClient
	chatClient           ChatClient
	fileProcessorURL     string
	aiLimitChecker       AILimitChecker
	learningPulseClient  LearningPulseClient
	behaviorDataProvider BehaviorDataProvider
}

type ChatSessionRepository interface {
	Create(ctx context.Context, session *domain.ChatSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatSession, error)
	FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*domain.ChatSession, error)
	FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID) (*domain.ChatSession, error)
	Update(ctx context.Context, session *domain.ChatSession) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ChatMessageRepository interface {
	Create(ctx context.Context, message *domain.ChatMessage) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]domain.ChatMessage, error)
	UpdateFeedback(ctx context.Context, id uuid.UUID, feedback domain.FeedbackType, feedbackText *string) error
	CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error)
}

type VectorRepository interface {
	Upsert(ctx context.Context, chunks []domain.MaterialChunk) error
	Search(ctx context.Context, embedding []float32, materialID *uuid.UUID, podID *uuid.UUID, limit int) ([]domain.MaterialChunk, error)
	DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error
}

type EmbeddingClient interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

type ChatClient interface {
	GenerateResponse(ctx context.Context, systemPrompt, userQuery string, contextChunks []string) (string, error)
	GenerateSuggestions(ctx context.Context, content string, existingQuestions []string) ([]string, error)
	GenerateQuestions(ctx context.Context, content string, count int, questionType string) ([]domain.GeneratedQuestion, error)
}

type ExportChatInput struct {
	UserID     uuid.UUID
	MaterialID uuid.UUID
	Format     domain.ExportFormat
}

type ExportChatOutput struct {
	Content     []byte
	Filename    string
	ContentType string
}

// GenerateQuestionsInput represents the input for question generation.
// Implements requirements 12.2, 12.3, 12.4.
type GenerateQuestionsInput struct {
	UserID       uuid.UUID
	MaterialID   uuid.UUID
	Count        int    // default 5, max 20
	QuestionType string // multiple_choice, true_false, short_answer, mixed
}

// GenerateQuestionsOutput represents the output of question generation.
// Implements requirement 12.5.
type GenerateQuestionsOutput struct {
	Questions []domain.GeneratedQuestion `json:"questions"`
}

type GetSuggestionsInput struct {
	UserID     uuid.UUID
	MaterialID uuid.UUID
}

func NewService(
	chatSessionRepo ChatSessionRepository,
	chatMessageRepo ChatMessageRepository,
	vectorRepo VectorRepository,
	embeddingClient EmbeddingClient,
	chatClient ChatClient,
	fileProcessorURL string,
	aiLimitChecker AILimitChecker,
	learningPulseClient LearningPulseClient,
	behaviorDataProvider BehaviorDataProvider,
) *Service {
	return &Service{
		chatSessionRepo:      chatSessionRepo,
		chatMessageRepo:      chatMessageRepo,
		vectorRepo:           vectorRepo,
		embeddingClient:      embeddingClient,
		chatClient:           chatClient,
		fileProcessorURL:     fileProcessorURL,
		aiLimitChecker:       aiLimitChecker,
		learningPulseClient:  learningPulseClient,
		behaviorDataProvider: behaviorDataProvider,
	}
}

type ChatInput struct {
	UserID     uuid.UUID
	MaterialID *uuid.UUID
	PodID      *uuid.UUID
	Message    string
}

type ChatOutput struct {
	Message *domain.ChatMessage
	Session *domain.ChatSession
}

func (s *Service) Chat(ctx context.Context, input ChatInput) (*ChatOutput, error) {
	// Check AI limit before processing chat
	// Implements Requirements 9.4, 9.5
	if s.aiLimitChecker != nil {
		if err := s.aiLimitChecker.CheckAILimit(ctx, input.UserID); err != nil {
			return nil, err
		}
	}

	// Determine chat mode
	mode := domain.ChatModeMaterial
	if input.PodID != nil && input.MaterialID == nil {
		mode = domain.ChatModePod
	}

	// Fetch user's learning persona for personalized responses
	persona := s.fetchUserPersona(ctx, input.UserID)

	// Find or create session
	var session *domain.ChatSession
	var err error

	if mode == domain.ChatModeMaterial && input.MaterialID != nil {
		session, err = s.chatSessionRepo.FindByUserAndMaterial(ctx, input.UserID, *input.MaterialID)
	} else if mode == domain.ChatModePod && input.PodID != nil {
		session, err = s.chatSessionRepo.FindByUserAndPod(ctx, input.UserID, *input.PodID)
	}

	if err != nil || session == nil {
		session = domain.NewChatSession(input.UserID, input.MaterialID, input.PodID, mode)
		if err := s.chatSessionRepo.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create chat session: %w", err)
		}
	}

	// Save user message
	userMessage := domain.NewChatMessage(session.ID, domain.MessageRoleUser, input.Message, nil)
	if err := s.chatMessageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Generate embedding for the query
	queryEmbedding, err := s.embeddingClient.GenerateEmbedding(ctx, input.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for relevant chunks
	chunks, err := s.vectorRepo.Search(ctx, queryEmbedding, input.MaterialID, input.PodID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search for relevant chunks: %w", err)
	}

	// Build context from chunks
	var contextTexts []string
	var sources []domain.ChunkSource
	for _, chunk := range chunks {
		contextTexts = append(contextTexts, chunk.Text)
		sources = append(sources, domain.ChunkSource{
			MaterialID: chunk.MaterialID,
			ChunkIndex: chunk.ChunkIndex,
			Text:       truncateText(chunk.Text, 200),
			Score:      0, // Score would come from search results
		})
	}

	// Generate response with personalized system prompt
	systemPrompt := buildPersonalizedSystemPrompt(mode, persona)
	response, err := s.chatClient.GenerateResponse(ctx, systemPrompt, input.Message, contextTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Handle case where no relevant context found
	if len(chunks) == 0 {
		response = "I couldn't find relevant information in the material to answer your question. Please try asking something related to the content."
		sources = nil
	}

	// Save assistant message
	assistantMessage := domain.NewChatMessage(session.ID, domain.MessageRoleAssistant, response, sources)
	if err := s.chatMessageRepo.Create(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Increment AI usage after successful processing
	// Implements Requirement 9.3
	if s.aiLimitChecker != nil {
		if err := s.aiLimitChecker.IncrementAIUsage(ctx, input.UserID); err != nil {
			// Log the error but don't fail the request since the chat was successful
			log.Warn().Err(err).Str("user_id", input.UserID.String()).Msg("failed to increment AI usage")
		}
	}

	return &ChatOutput{
		Message: assistantMessage,
		Session: session,
	}, nil
}

// fetchUserPersona retrieves the user's learning persona from Learning Pulse.
// Returns PersonaUnknown if the service is unavailable or data is missing.
func (s *Service) fetchUserPersona(ctx context.Context, userID uuid.UUID) learningpulse.Persona {
	if s.learningPulseClient == nil {
		log.Debug().Str("user_id", userID.String()).Msg("learningPulseClient is nil - persona personalization disabled")
		return learningpulse.PersonaUnknown
	}

	if s.behaviorDataProvider == nil {
		log.Debug().Str("user_id", userID.String()).Msg("behaviorDataProvider is nil - persona personalization disabled")
		return learningpulse.PersonaUnknown
	}

	// Fetch behavior data
	behaviorData, err := s.behaviorDataProvider.GetBehaviorData(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to fetch behavior data for persona")
		return learningpulse.PersonaUnknown
	}

	if behaviorData == nil {
		log.Debug().Str("user_id", userID.String()).Msg("behavior data is nil - not enough data for persona prediction")
		return learningpulse.PersonaUnknown
	}

	log.Debug().
		Str("user_id", userID.String()).
		Int("chat_messages", behaviorData.Chat.TotalMessages).
		Int("material_views", behaviorData.Material.TotalViews).
		Int("active_days", behaviorData.Activity.ActiveDays).
		Msg("fetched behavior data for persona prediction")

	// Get persona from Learning Pulse
	persona, err := s.learningPulseClient.GetPersona(ctx, userID.String(), behaviorData)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to fetch persona from learning pulse")
		return learningpulse.PersonaUnknown
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("persona", string(persona)).
		Msg("fetched user persona for personalized chat")

	return persona
}

func (s *Service) GetChatHistory(ctx context.Context, userID uuid.UUID, materialID *uuid.UUID, podID *uuid.UUID, limit, offset int) ([]domain.ChatMessage, int, error) {
	var session *domain.ChatSession
	var err error

	if materialID != nil {
		session, err = s.chatSessionRepo.FindByUserAndMaterial(ctx, userID, *materialID)
	} else if podID != nil {
		session, err = s.chatSessionRepo.FindByUserAndPod(ctx, userID, *podID)
	}

	if err != nil || session == nil {
		return []domain.ChatMessage{}, 0, nil
	}

	messages, err := s.chatMessageRepo.FindBySessionID(ctx, session.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat history: %w", err)
	}

	total, err := s.chatMessageRepo.CountBySessionID(ctx, session.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return messages, total, nil
}

func (s *Service) SubmitFeedback(ctx context.Context, messageID uuid.UUID, feedback domain.FeedbackType, feedbackText *string) error {
	return s.chatMessageRepo.UpdateFeedback(ctx, messageID, feedback, feedbackText)
}

func (s *Service) GetSuggestions(ctx context.Context, input GetSuggestionsInput) ([]domain.SuggestedQuestion, error) {
	chunks, err := s.vectorRepo.Search(ctx, nil, &input.MaterialID, nil, 3)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get chunks for suggestions")
		return defaultSuggestions(), nil
	}

	if len(chunks) == 0 {
		return defaultSuggestions(), nil
	}

	// Build content summary
	var contentParts []string
	for _, chunk := range chunks {
		contentParts = append(contentParts, chunk.Text)
	}
	content := strings.Join(contentParts, "\n\n")

	var existingQuestions []string
	session, err := s.chatSessionRepo.FindByUserAndMaterial(ctx, input.UserID, input.MaterialID)
	if err == nil && session != nil {
		messages, err := s.chatMessageRepo.FindBySessionID(ctx, session.ID, 20, 0)
		if err == nil {
			for _, msg := range messages {
				if msg.Role == domain.MessageRoleUser {
					existingQuestions = append(existingQuestions, msg.Content)
				}
			}
		}
	}

	// Generate suggestions
	suggestions, err := s.chatClient.GenerateSuggestions(ctx, content, existingQuestions)
	if err != nil {
		log.Warn().Err(err).Msg("failed to generate suggestions")
		return defaultSuggestions(), nil
	}

	var result []domain.SuggestedQuestion
	for _, q := range suggestions {
		result = append(result, domain.SuggestedQuestion{Question: q})
	}

	return result, nil
}

func (s *Service) ProcessMaterial(ctx context.Context, materialID, podID uuid.UUID, fileURL, fileType string) error {
	log.Info().
		Str("material_id", materialID.String()).
		Str("file_url", fileURL).
		Msg("processing material")

	// This will be implemented in task 7.7
	// 1. Call file processor to extract text
	// 2. Chunk the text
	// 3. Generate embeddings
	// 4. Store in Qdrant

	return nil
}

func (s *Service) DeleteMaterialChunks(ctx context.Context, materialID uuid.UUID) error {
	return s.vectorRepo.DeleteByMaterialID(ctx, materialID)
}

// CheckFeatureAccess checks if a user has access to a specific premium feature.
// Returns nil if access is allowed, or an appropriate error if not.
// Implements requirements 11.1, 11.2, 11.3, 11.4.
func (s *Service) CheckFeatureAccess(ctx context.Context, userID uuid.UUID, feature string) error {
	if s.aiLimitChecker == nil {
		return nil // No limit checker configured, allow access
	}
	return s.aiLimitChecker.CanAccessFeature(ctx, userID, feature)
}

// GenerateQuestions generates quiz questions from material content.
// Implements requirements 12.1, 12.2, 12.3, 12.4, 12.5, 12.6.
func (s *Service) GenerateQuestions(ctx context.Context, input GenerateQuestionsInput) (*GenerateQuestionsOutput, error) {
	// Check user has pro tier via CanAccessFeature
	// Implements requirement 12.1, 12.6
	if s.aiLimitChecker != nil {
		if err := s.aiLimitChecker.CanAccessFeature(ctx, input.UserID, FeatureQuestionGeneration); err != nil {
			return nil, err
		}
	}

	// Validate and set default count
	// Implements requirement 12.3
	count := input.Count
	if count <= 0 {
		count = DefaultQuestionCount
	}
	if count > MaxQuestionCount {
		count = MaxQuestionCount
	}

	// Validate question type
	// Implements requirement 12.4
	questionType := input.QuestionType
	if questionType == "" {
		questionType = QuestionTypeMixed
	}
	if !isValidQuestionType(questionType) {
		questionType = QuestionTypeMixed
	}

	// Get material content from vector store
	// We retrieve more chunks to have enough context for question generation
	chunks, err := s.vectorRepo.Search(ctx, nil, &input.MaterialID, nil, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get material content: %w", err)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no content found for material")
	}

	// Build content from chunks
	var contentParts []string
	for _, chunk := range chunks {
		contentParts = append(contentParts, chunk.Text)
	}
	content := strings.Join(contentParts, "\n\n")

	// Call AI provider to generate questions
	// Implements requirements 12.2, 12.4, 12.5
	questions, err := s.chatClient.GenerateQuestions(ctx, content, count, questionType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	return &GenerateQuestionsOutput{
		Questions: questions,
	}, nil
}

// isValidQuestionType checks if the question type is valid.
func isValidQuestionType(questionType string) bool {
	switch questionType {
	case QuestionTypeMultipleChoice, QuestionTypeTrueFalse, QuestionTypeShortAnswer, QuestionTypeMixed:
		return true
	default:
		return false
	}
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func defaultSuggestions() []domain.SuggestedQuestion {
	return []domain.SuggestedQuestion{
		{Question: "What is the main topic of this material?"},
		{Question: "Can you summarize the key points?"},
		{Question: "What are the most important concepts covered?"},
	}
}

func (s *Service) ExportChat(ctx context.Context, input ExportChatInput) (*ExportChatOutput, error) {
	session, err := s.chatSessionRepo.FindByUserAndMaterial(ctx, input.UserID, input.MaterialID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("no chat history found for this material")
	}

	messages, err := s.chatMessageRepo.FindBySessionID(ctx, session.ID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages to export")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	switch input.Format {
	case domain.ExportFormatMarkdown:
		content := s.generateMarkdownExport(session, messages)
		return &ExportChatOutput{
			Content:     []byte(content),
			Filename:    fmt.Sprintf("chat_export_%s.md", timestamp),
			ContentType: "text/markdown",
		}, nil

	case domain.ExportFormatPDF:
		content, err := s.generatePDFExport(session, messages)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %w", err)
		}
		return &ExportChatOutput{
			Content:     content,
			Filename:    fmt.Sprintf("chat_export_%s.pdf", timestamp),
			ContentType: "application/pdf",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported export format: %s", input.Format)
	}
}

func (s *Service) generateMarkdownExport(session *domain.ChatSession, messages []domain.ChatMessage) string {
	var sb strings.Builder

	sb.WriteString("# Chat Export\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID:** %s\n", session.ID.String()))
	sb.WriteString(fmt.Sprintf("**Exported at:** %s\n", time.Now().Format(time.RFC3339)))
	if session.MaterialID != nil {
		sb.WriteString(fmt.Sprintf("**Material ID:** %s\n", session.MaterialID.String()))
	}
	sb.WriteString(fmt.Sprintf("**Mode:** %s\n\n", session.Mode))
	sb.WriteString("---\n\n")

	for _, msg := range messages {
		roleLabel := "🧑 User"
		if msg.Role == domain.MessageRoleAssistant {
			roleLabel = "🤖 Assistant"
		}

		sb.WriteString(fmt.Sprintf("### %s\n", roleLabel))
		sb.WriteString(fmt.Sprintf("*%s*\n\n", msg.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")

		if len(msg.Sources) > 0 {
			sb.WriteString("**Sources:**\n")
			for i, source := range msg.Sources {
				sb.WriteString(fmt.Sprintf("- [%d] Chunk %d: %s\n", i+1, source.ChunkIndex, source.Text))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

func (s *Service) generatePDFExport(session *domain.ChatSession, messages []domain.ChatMessage) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString("Chat Export\n")
	sb.WriteString("===========\n\n")
	sb.WriteString(fmt.Sprintf("Session ID: %s\n", session.ID.String()))
	sb.WriteString(fmt.Sprintf("Exported at: %s\n", time.Now().Format(time.RFC3339)))
	if session.MaterialID != nil {
		sb.WriteString(fmt.Sprintf("Material ID: %s\n", session.MaterialID.String()))
	}
	sb.WriteString(fmt.Sprintf("Mode: %s\n\n", session.Mode))
	sb.WriteString("----------------------------------------\n\n")

	for _, msg := range messages {
		roleLabel := "USER"
		if msg.Role == domain.MessageRoleAssistant {
			roleLabel = "ASSISTANT"
		}

		sb.WriteString(fmt.Sprintf("[%s] %s\n", roleLabel, msg.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("%s\n\n", msg.Content))

		if len(msg.Sources) > 0 {
			sb.WriteString("Sources:\n")
			for i, source := range msg.Sources {
				sb.WriteString(fmt.Sprintf("  [%d] Chunk %d: %s\n", i+1, source.ChunkIndex, source.Text))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("----------------------------------------\n\n")
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 10)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Chat Export")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Session ID: %s", session.ID.String()))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Exported at: %s", time.Now().Format(time.RFC3339)))
	pdf.Ln(6)
	if session.MaterialID != nil {
		pdf.Cell(0, 6, fmt.Sprintf("Material ID: %s", session.MaterialID.String()))
		pdf.Ln(6)
	}
	pdf.Cell(0, 6, fmt.Sprintf("Mode: %s", session.Mode))
	pdf.Ln(12)

	for _, msg := range messages {
		roleLabel := "User"
		if msg.Role == domain.MessageRoleAssistant {
			roleLabel = "Assistant"
		}

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 8, fmt.Sprintf("%s - %s", roleLabel, msg.CreatedAt.Format("2006-01-02 15:04:05")))
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, msg.Content, "", "", false)
		pdf.Ln(4)

		if len(msg.Sources) > 0 {
			pdf.SetFont("Arial", "I", 9)
			pdf.Cell(0, 5, "Sources:")
			pdf.Ln(5)
			for i, source := range msg.Sources {
				text := truncateText(source.Text, 100)
				pdf.Cell(0, 5, fmt.Sprintf("  [%d] Chunk %d: %s", i+1, source.ChunkIndex, text))
				pdf.Ln(5)
			}
		}

		pdf.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
