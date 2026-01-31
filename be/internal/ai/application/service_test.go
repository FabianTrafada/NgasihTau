// Package application contains unit tests for the AI Service.
package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"ngasihtau/internal/ai/domain"
)

// Mock implementations for repositories

type mockChatSessionRepo struct {
	sessions        map[uuid.UUID]*domain.ChatSession
	userMaterialIdx map[string]*domain.ChatSession
	userPodIdx      map[string]*domain.ChatSession
	createErr       error
	findErr         error
}

func newMockChatSessionRepo() *mockChatSessionRepo {
	return &mockChatSessionRepo{
		sessions:        make(map[uuid.UUID]*domain.ChatSession),
		userMaterialIdx: make(map[string]*domain.ChatSession),
		userPodIdx:      make(map[string]*domain.ChatSession),
	}
}

func (m *mockChatSessionRepo) Create(ctx context.Context, session *domain.ChatSession) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[session.ID] = session
	if session.MaterialID != nil {
		key := session.UserID.String() + ":" + session.MaterialID.String()
		m.userMaterialIdx[key] = session
	}
	if session.PodID != nil {
		key := session.UserID.String() + ":" + session.PodID.String()
		m.userPodIdx[key] = session
	}
	return nil
}

func (m *mockChatSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatSession, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	session, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (m *mockChatSessionRepo) FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*domain.ChatSession, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	key := userID.String() + ":" + materialID.String()
	session, ok := m.userMaterialIdx[key]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (m *mockChatSessionRepo) FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID) (*domain.ChatSession, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	key := userID.String() + ":" + podID.String()
	session, ok := m.userPodIdx[key]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (m *mockChatSessionRepo) Update(ctx context.Context, session *domain.ChatSession) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockChatSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.sessions, id)
	return nil
}

type mockChatMessageRepo struct {
	messages   map[uuid.UUID]*domain.ChatMessage
	sessionIdx map[uuid.UUID][]domain.ChatMessage
	createErr  error
}

func newMockChatMessageRepo() *mockChatMessageRepo {
	return &mockChatMessageRepo{
		messages:   make(map[uuid.UUID]*domain.ChatMessage),
		sessionIdx: make(map[uuid.UUID][]domain.ChatMessage),
	}
}

func (m *mockChatMessageRepo) Create(ctx context.Context, message *domain.ChatMessage) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.messages[message.ID] = message
	m.sessionIdx[message.SessionID] = append(m.sessionIdx[message.SessionID], *message)
	return nil
}

func (m *mockChatMessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	msg, ok := m.messages[id]
	if !ok {
		return nil, nil
	}
	return msg, nil
}

func (m *mockChatMessageRepo) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]domain.ChatMessage, error) {
	msgs := m.sessionIdx[sessionID]
	if offset >= len(msgs) {
		return []domain.ChatMessage{}, nil
	}
	end := offset + limit
	if end > len(msgs) {
		end = len(msgs)
	}
	return msgs[offset:end], nil
}

func (m *mockChatMessageRepo) UpdateFeedback(ctx context.Context, id uuid.UUID, feedback domain.FeedbackType, feedbackText *string) error {
	msg, ok := m.messages[id]
	if ok {
		msg.Feedback = &feedback
		msg.FeedbackText = feedbackText
	}
	return nil
}

func (m *mockChatMessageRepo) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error) {
	return len(m.sessionIdx[sessionID]), nil
}

type mockVectorRepo struct {
	chunks    []domain.MaterialChunk
	searchErr error
	upsertErr error
	deleteErr error
}

func newMockVectorRepo() *mockVectorRepo {
	return &mockVectorRepo{
		chunks: []domain.MaterialChunk{},
	}
}

func (m *mockVectorRepo) Upsert(ctx context.Context, chunks []domain.MaterialChunk) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.chunks = append(m.chunks, chunks...)
	return nil
}

func (m *mockVectorRepo) Search(ctx context.Context, embedding []float32, materialID *uuid.UUID, podID *uuid.UUID, limit int) ([]domain.MaterialChunk, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var result []domain.MaterialChunk
	for _, chunk := range m.chunks {
		if materialID != nil && chunk.MaterialID != *materialID {
			continue
		}
		if podID != nil && chunk.PodID != *podID {
			continue
		}
		result = append(result, chunk)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockVectorRepo) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	var remaining []domain.MaterialChunk
	for _, chunk := range m.chunks {
		if chunk.MaterialID != materialID {
			remaining = append(remaining, chunk)
		}
	}
	m.chunks = remaining
	return nil
}

type mockEmbeddingClient struct {
	embedding  []float32
	embeddings [][]float32
	err        error
}

func newMockEmbeddingClient() *mockEmbeddingClient {
	return &mockEmbeddingClient{
		embedding: make([]float32, 1536),
	}
}

func (m *mockEmbeddingClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.embedding, nil
}

func (m *mockEmbeddingClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.embeddings != nil {
		return m.embeddings, nil
	}
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 1536)
	}
	return result, nil
}

type mockChatClient struct {
	response    string
	suggestions []string
	questions   []domain.GeneratedQuestion
	err         error
}

func newMockChatClient() *mockChatClient {
	return &mockChatClient{
		response:    "This is a test response based on the provided context.",
		suggestions: []string{"What is the main topic?", "Can you explain more?", "What are the key points?"},
		questions: []domain.GeneratedQuestion{
			{
				Question:    "What is the main concept?",
				Type:        "multiple_choice",
				Options:     []string{"A. Option 1", "B. Option 2", "C. Option 3", "D. Option 4"},
				Answer:      "A. Option 1",
				Explanation: "This is the correct answer because...",
			},
		},
	}
}

func (m *mockChatClient) GenerateResponse(ctx context.Context, systemPrompt, userQuery string, contextChunks []string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockChatClient) GenerateSuggestions(ctx context.Context, content string, existingQuestions []string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.suggestions, nil
}

func (m *mockChatClient) GenerateQuestions(ctx context.Context, content string, count int, questionType string) ([]domain.GeneratedQuestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return up to count questions
	if count > len(m.questions) {
		return m.questions, nil
	}
	return m.questions[:count], nil
}

// mockAILimitChecker is a mock implementation of AILimitChecker for testing.
type mockAILimitChecker struct {
	checkErr        error
	incrementErr    error
	featureCheckErr error
}

func newMockAILimitChecker() *mockAILimitChecker {
	return &mockAILimitChecker{}
}

func (m *mockAILimitChecker) CheckAILimit(ctx context.Context, userID uuid.UUID) error {
	return m.checkErr
}

func (m *mockAILimitChecker) IncrementAIUsage(ctx context.Context, userID uuid.UUID) error {
	return m.incrementErr
}

func (m *mockAILimitChecker) CanAccessFeature(ctx context.Context, userID uuid.UUID, feature string) error {
	return m.featureCheckErr
}

// Helper to create a test service
func newTestService() (*Service, *mockChatSessionRepo, *mockChatMessageRepo, *mockVectorRepo, *mockEmbeddingClient, *mockChatClient) {
	sessionRepo := newMockChatSessionRepo()
	messageRepo := newMockChatMessageRepo()
	vectorRepo := newMockVectorRepo()
	embeddingClient := newMockEmbeddingClient()
	chatClient := newMockChatClient()
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

	return svc, sessionRepo, messageRepo, vectorRepo, embeddingClient, chatClient
}

// Test: Text Chunking
func TestChunker_ChunkText_BasicParagraphs(t *testing.T) {
	chunker := NewChunker(DefaultChunkerConfig())

	text := `This is the first paragraph with some content about machine learning.

This is the second paragraph discussing neural networks and deep learning concepts.

This is the third paragraph about natural language processing and transformers.`

	chunks := chunker.ChunkText(text)

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	for i, chunk := range chunks {
		if chunk.Index != i {
			t.Errorf("Expected chunk index %d, got %d", i, chunk.Index)
		}
		if chunk.Text == "" {
			t.Errorf("Chunk %d has empty text", i)
		}
		if chunk.TokenCount <= 0 {
			t.Errorf("Chunk %d has invalid token count: %d", i, chunk.TokenCount)
		}
	}
}

func TestChunker_ChunkText_EmptyInput(t *testing.T) {
	chunker := NewChunker(DefaultChunkerConfig())

	chunks := chunker.ChunkText("")

	if chunks != nil {
		t.Errorf("Expected nil for empty input, got %v", chunks)
	}
}

func TestChunker_ChunkText_LongText(t *testing.T) {
	chunker := NewChunker(ChunkerConfig{
		TargetChunkSize: 50,
		MaxChunkSize:    100,
		OverlapSize:     10,
	})

	// Generate long text
	var longText string
	for i := 0; i < 100; i++ {
		longText += "This is sentence number " + string(rune('0'+i%10)) + ". "
	}

	chunks := chunker.ChunkText(longText)

	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks for long text, got %d", len(chunks))
	}

	// Verify all chunks have content
	for i, chunk := range chunks {
		if chunk.Text == "" {
			t.Errorf("Chunk %d is empty", i)
		}
	}
}

func TestChunker_ChunkText_PreservesParagraphBoundaries(t *testing.T) {
	chunker := NewChunker(ChunkerConfig{
		TargetChunkSize: 100,
		MaxChunkSize:    200,
		OverlapSize:     20,
	})

	text := `First paragraph with important content.

Second paragraph with more details.

Third paragraph concluding the topic.`

	chunks := chunker.ChunkText(text)

	// With small target size, paragraphs should be preserved where possible
	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}
}

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int
		maxCount int
	}{
		{"empty", "", 0, 0},
		{"single word", "hello", 1, 2},
		{"short sentence", "Hello world", 2, 4},
		{"longer text", "This is a longer sentence with multiple words", 8, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := estimateTokenCount(tt.text)
			if count < tt.minCount || count > tt.maxCount {
				t.Errorf("estimateTokenCount(%q) = %d, want between %d and %d", tt.text, count, tt.minCount, tt.maxCount)
			}
		})
	}
}

// Test: Embedding Generation (via mock)
func TestService_Chat_GeneratesEmbedding(t *testing.T) {
	svc, _, _, vectorRepo, embeddingClient, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	// Add some chunks to vector repo
	vectorRepo.chunks = []domain.MaterialChunk{
		{
			ID:         "chunk1",
			MaterialID: materialID,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "This is relevant content about the topic.",
		},
	}

	// Set expected embedding
	embeddingClient.embedding = make([]float32, 1536)
	for i := range embeddingClient.embedding {
		embeddingClient.embedding[i] = 0.1
	}

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this about?",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if result.Message == nil {
		t.Fatal("Expected message in result")
	}
	if result.Message.Role != domain.MessageRoleAssistant {
		t.Errorf("Expected assistant role, got %s", result.Message.Role)
	}
}

func TestService_Chat_EmbeddingError(t *testing.T) {
	svc, _, _, _, embeddingClient, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Set embedding error
	embeddingClient.err = context.DeadlineExceeded

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this about?",
	}

	_, err := svc.Chat(ctx, input)
	if err == nil {
		t.Fatal("Expected error when embedding fails")
	}
}

// Test: RAG Retrieval
func TestService_Chat_RetrievesRelevantChunks(t *testing.T) {
	svc, _, messageRepo, vectorRepo, _, chatClient := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	// Add chunks to vector repo
	vectorRepo.chunks = []domain.MaterialChunk{
		{
			ID:         "chunk1",
			MaterialID: materialID,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "Machine learning is a subset of artificial intelligence.",
		},
		{
			ID:         "chunk2",
			MaterialID: materialID,
			PodID:      podID,
			ChunkIndex: 1,
			Text:       "Deep learning uses neural networks with many layers.",
		},
	}

	chatClient.response = "Based on the context, machine learning is a subset of AI."

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is machine learning?",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Verify response was generated
	if result.Message.Content == "" {
		t.Error("Expected non-empty response")
	}

	// Verify sources are included
	if len(result.Message.Sources) == 0 {
		t.Error("Expected sources in response")
	}

	// Verify messages were stored
	if len(messageRepo.messages) != 2 { // user + assistant
		t.Errorf("Expected 2 messages stored, got %d", len(messageRepo.messages))
	}
}

func TestService_Chat_NoRelevantChunks(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Empty vector repo - no chunks
	vectorRepo.chunks = []domain.MaterialChunk{}

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is quantum physics?",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should return a message indicating no relevant content
	if result.Message.Content == "" {
		t.Error("Expected response even with no chunks")
	}
	if len(result.Message.Sources) != 0 {
		t.Error("Expected no sources when no chunks found")
	}
}

func TestService_Chat_PodWideMode(t *testing.T) {
	svc, sessionRepo, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	podID := uuid.New()
	material1 := uuid.New()
	material2 := uuid.New()

	// Add chunks from multiple materials in the same pod
	vectorRepo.chunks = []domain.MaterialChunk{
		{
			ID:         "chunk1",
			MaterialID: material1,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "Content from first material.",
		},
		{
			ID:         "chunk2",
			MaterialID: material2,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "Content from second material.",
		},
	}

	input := ChatInput{
		UserID:  userID,
		PodID:   &podID,
		Message: "Summarize all materials",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Verify pod-wide session was created
	if result.Session.Mode != domain.ChatModePod {
		t.Errorf("Expected pod mode, got %s", result.Session.Mode)
	}

	// Verify session was stored
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessionRepo.sessions))
	}
}

// Test: Chat Response Generation
func TestService_Chat_GeneratesResponse(t *testing.T) {
	svc, _, _, vectorRepo, _, chatClient := newTestService()
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
			Text:       "Python is a programming language.",
		},
	}

	expectedResponse := "Python is a high-level programming language known for its simplicity."
	chatClient.response = expectedResponse

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is Python?",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if result.Message.Content != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, result.Message.Content)
	}
}

func TestService_Chat_ChatClientError(t *testing.T) {
	svc, _, _, vectorRepo, _, chatClient := newTestService()
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
			Text:       "Some content.",
		},
	}

	chatClient.err = context.DeadlineExceeded

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this?",
	}

	_, err := svc.Chat(ctx, input)
	if err == nil {
		t.Fatal("Expected error when chat client fails")
	}
}

// Test: Chat History
func TestService_GetChatHistory(t *testing.T) {
	svc, sessionRepo, messageRepo, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Create session
	session := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[session.ID] = session
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = session

	// Add messages
	msg1 := domain.NewChatMessage(session.ID, domain.MessageRoleUser, "Hello", nil)
	msg2 := domain.NewChatMessage(session.ID, domain.MessageRoleAssistant, "Hi there!", nil)
	messageRepo.messages[msg1.ID] = msg1
	messageRepo.messages[msg2.ID] = msg2
	messageRepo.sessionIdx[session.ID] = []domain.ChatMessage{*msg1, *msg2}

	messages, total, err := svc.GetChatHistory(ctx, userID, &materialID, nil, 10, 0)
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestService_GetChatHistory_NoSession(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	messages, total, err := svc.GetChatHistory(ctx, userID, &materialID, nil, 10, 0)
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
}

// Test: Feedback
func TestService_SubmitFeedback(t *testing.T) {
	svc, _, messageRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Create message
	sessionID := uuid.New()
	msg := domain.NewChatMessage(sessionID, domain.MessageRoleAssistant, "Response", nil)
	messageRepo.messages[msg.ID] = msg

	feedback := domain.FeedbackThumbsUp
	feedbackText := "Very helpful!"

	err := svc.SubmitFeedback(ctx, msg.ID, feedback, &feedbackText)
	if err != nil {
		t.Fatalf("SubmitFeedback failed: %v", err)
	}

	// Verify feedback was stored
	storedMsg := messageRepo.messages[msg.ID]
	if storedMsg.Feedback == nil || *storedMsg.Feedback != feedback {
		t.Error("Expected feedback to be stored")
	}
	if storedMsg.FeedbackText == nil || *storedMsg.FeedbackText != feedbackText {
		t.Error("Expected feedback text to be stored")
	}
}

// Test: Suggestions
func TestService_GetSuggestions(t *testing.T) {
	svc, _, _, vectorRepo, _, chatClient := newTestService()
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
			Text:       "Introduction to machine learning concepts.",
		},
	}

	chatClient.suggestions = []string{
		"What are the main ML algorithms?",
		"How does supervised learning work?",
		"What is the difference between ML and AI?",
	}

	input := GetSuggestionsInput{
		UserID:     userID,
		MaterialID: materialID,
	}

	suggestions, err := svc.GetSuggestions(ctx, input)
	if err != nil {
		t.Fatalf("GetSuggestions failed: %v", err)
	}

	if len(suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(suggestions))
	}
}

func TestService_GetSuggestions_NoChunks(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	vectorRepo.chunks = []domain.MaterialChunk{}

	input := GetSuggestionsInput{
		UserID:     userID,
		MaterialID: materialID,
	}

	suggestions, err := svc.GetSuggestions(ctx, input)
	if err != nil {
		t.Fatalf("GetSuggestions failed: %v", err)
	}

	// Should return default suggestions
	if len(suggestions) == 0 {
		t.Error("Expected default suggestions")
	}
}

// Test: Delete Material Chunks
func TestService_DeleteMaterialChunks(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	otherMaterialID := uuid.New()
	podID := uuid.New()

	vectorRepo.chunks = []domain.MaterialChunk{
		{ID: "chunk1", MaterialID: materialID, PodID: podID, ChunkIndex: 0, Text: "Content 1"},
		{ID: "chunk2", MaterialID: materialID, PodID: podID, ChunkIndex: 1, Text: "Content 2"},
		{ID: "chunk3", MaterialID: otherMaterialID, PodID: podID, ChunkIndex: 0, Text: "Other content"},
	}

	err := svc.DeleteMaterialChunks(ctx, materialID)
	if err != nil {
		t.Fatalf("DeleteMaterialChunks failed: %v", err)
	}

	// Verify only target material chunks were deleted
	if len(vectorRepo.chunks) != 1 {
		t.Errorf("Expected 1 chunk remaining, got %d", len(vectorRepo.chunks))
	}
	if vectorRepo.chunks[0].MaterialID != otherMaterialID {
		t.Error("Wrong chunk was deleted")
	}
}

// Test: Helper functions
func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{"short text", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello..."},
		{"empty", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	materialPrompt := buildSystemPrompt(domain.ChatModeMaterial)
	if materialPrompt == "" {
		t.Error("Expected non-empty material prompt")
	}

	podPrompt := buildSystemPrompt(domain.ChatModePod)
	if podPrompt == "" {
		t.Error("Expected non-empty pod prompt")
	}

	if materialPrompt == podPrompt {
		t.Error("Expected different prompts for material and pod modes")
	}
}

func TestDefaultSuggestions(t *testing.T) {
	suggestions := defaultSuggestions()

	if len(suggestions) == 0 {
		t.Error("Expected default suggestions")
	}

	for i, s := range suggestions {
		if s.Question == "" {
			t.Errorf("Suggestion %d has empty question", i)
		}
	}
}

// Test: Normalize Whitespace
func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello world", "hello world"},
		{"multiple spaces", "hello    world", "hello world"},
		{"newlines", "hello\n\n\nworld", "hello\n\nworld"},
		{"mixed", "hello  \n  world", "hello   world"},
		{"tabs", "hello\t\tworld", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test: Split Into Paragraphs
func TestSplitIntoParagraphs(t *testing.T) {
	text := `First paragraph.

Second paragraph.

Third paragraph.`

	paragraphs := splitIntoParagraphs(text)

	if len(paragraphs) != 3 {
		t.Errorf("Expected 3 paragraphs, got %d", len(paragraphs))
	}

	expected := []string{"First paragraph.", "Second paragraph.", "Third paragraph."}
	for i, p := range paragraphs {
		if p != expected[i] {
			t.Errorf("Paragraph %d: expected %q, got %q", i, expected[i], p)
		}
	}
}

// Test: Split Into Sentences
func TestSplitIntoSentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			"simple",
			"Hello world. How are you?",
			[]string{"Hello world.", "How are you?"},
		},
		{
			"exclamation",
			"Wow! That is amazing.",
			[]string{"Wow!", "That is amazing."},
		},
		{
			"no punctuation",
			"Hello world",
			[]string{"Hello world"},
		},
		{
			"multiple sentences",
			"First. Second. Third.",
			[]string{"First.", "Second.", "Third."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoSentences(tt.text)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d sentences, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, s := range result {
				if s != tt.expected[i] {
					t.Errorf("Sentence %d: expected %q, got %q", i, tt.expected[i], s)
				}
			}
		})
	}
}

// Test: Session Reuse
func TestService_Chat_ReusesExistingSession(t *testing.T) {
	svc, sessionRepo, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	// Create existing session
	existingSession := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[existingSession.ID] = existingSession
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = existingSession

	vectorRepo.chunks = []domain.MaterialChunk{
		{ID: "chunk1", MaterialID: materialID, PodID: podID, ChunkIndex: 0, Text: "Content"},
	}

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "Hello",
	}

	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Should reuse existing session
	if result.Session.ID != existingSession.ID {
		t.Error("Expected to reuse existing session")
	}

	// Should still have only 1 session
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessionRepo.sessions))
	}
}

// Test: Export Chat
func TestService_ExportChat_Markdown(t *testing.T) {
	svc, sessionRepo, messageRepo, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Create session
	session := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[session.ID] = session
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = session

	// Add messages
	msg1 := domain.NewChatMessage(session.ID, domain.MessageRoleUser, "What is AI?", nil)
	msg2 := domain.NewChatMessage(session.ID, domain.MessageRoleAssistant, "AI is artificial intelligence.", nil)
	messageRepo.messages[msg1.ID] = msg1
	messageRepo.messages[msg2.ID] = msg2
	messageRepo.sessionIdx[session.ID] = []domain.ChatMessage{*msg1, *msg2}

	input := ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     domain.ExportFormatMarkdown,
	}

	result, err := svc.ExportChat(ctx, input)
	if err != nil {
		t.Fatalf("ExportChat failed: %v", err)
	}

	if result.ContentType != "text/markdown" {
		t.Errorf("Expected content type text/markdown, got %s", result.ContentType)
	}
	if len(result.Content) == 0 {
		t.Error("Expected non-empty content")
	}
	if result.Filename == "" {
		t.Error("Expected non-empty filename")
	}
}

func TestService_ExportChat_PDF(t *testing.T) {
	svc, sessionRepo, messageRepo, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Create session
	session := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[session.ID] = session
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = session

	// Add messages
	msg1 := domain.NewChatMessage(session.ID, domain.MessageRoleUser, "Question", nil)
	msg2 := domain.NewChatMessage(session.ID, domain.MessageRoleAssistant, "Answer", nil)
	messageRepo.messages[msg1.ID] = msg1
	messageRepo.messages[msg2.ID] = msg2
	messageRepo.sessionIdx[session.ID] = []domain.ChatMessage{*msg1, *msg2}

	input := ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     domain.ExportFormatPDF,
	}

	result, err := svc.ExportChat(ctx, input)
	if err != nil {
		t.Fatalf("ExportChat failed: %v", err)
	}

	if result.ContentType != "application/pdf" {
		t.Errorf("Expected content type application/pdf, got %s", result.ContentType)
	}
	if len(result.Content) == 0 {
		t.Error("Expected non-empty content")
	}
}

func TestService_ExportChat_NoSession(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	input := ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     domain.ExportFormatMarkdown,
	}

	_, err := svc.ExportChat(ctx, input)
	if err == nil {
		t.Fatal("Expected error when no session exists")
	}
}

func TestService_ExportChat_NoMessages(t *testing.T) {
	svc, sessionRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Create session with no messages
	session := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[session.ID] = session
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = session

	input := ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     domain.ExportFormatMarkdown,
	}

	_, err := svc.ExportChat(ctx, input)
	if err == nil {
		t.Fatal("Expected error when no messages exist")
	}
}

func TestService_ExportChat_UnsupportedFormat(t *testing.T) {
	svc, sessionRepo, messageRepo, _, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()

	// Create session
	session := domain.NewChatSession(userID, &materialID, nil, domain.ChatModeMaterial)
	sessionRepo.sessions[session.ID] = session
	key := userID.String() + ":" + materialID.String()
	sessionRepo.userMaterialIdx[key] = session

	// Add message
	msg := domain.NewChatMessage(session.ID, domain.MessageRoleUser, "Hello", nil)
	messageRepo.messages[msg.ID] = msg
	messageRepo.sessionIdx[session.ID] = []domain.ChatMessage{*msg}

	input := ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     "invalid",
	}

	_, err := svc.ExportChat(ctx, input)
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
}

// Test: AI Limit Checking
func TestService_Chat_AILimitExceeded(t *testing.T) {
	sessionRepo := newMockChatSessionRepo()
	messageRepo := newMockChatMessageRepo()
	vectorRepo := newMockVectorRepo()
	embeddingClient := newMockEmbeddingClient()
	chatClient := newMockChatClient()
	aiLimitChecker := newMockAILimitChecker()

	// Set AI limit exceeded error
	aiLimitChecker.checkErr = fmt.Errorf("daily AI message limit exceeded")

	svc := NewService(
		sessionRepo,
		messageRepo,
		vectorRepo,
		embeddingClient,
		chatClient,
		"http://localhost:8000",
		aiLimitChecker,
	)

	ctx := context.Background()
	userID := uuid.New()
	materialID := uuid.New()

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this about?",
	}

	_, err := svc.Chat(ctx, input)
	if err == nil {
		t.Fatal("Expected error when AI limit is exceeded")
	}
	if err.Error() != "daily AI message limit exceeded" {
		t.Errorf("Expected AI limit exceeded error, got: %v", err)
	}
}

func TestService_Chat_AILimitIncrementsUsage(t *testing.T) {
	sessionRepo := newMockChatSessionRepo()
	messageRepo := newMockChatMessageRepo()
	vectorRepo := newMockVectorRepo()
	embeddingClient := newMockEmbeddingClient()
	chatClient := newMockChatClient()
	aiLimitChecker := &mockAILimitCheckerWithTracking{}

	svc := NewService(
		sessionRepo,
		messageRepo,
		vectorRepo,
		embeddingClient,
		chatClient,
		"http://localhost:8000",
		aiLimitChecker,
	)

	ctx := context.Background()
	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	// Add some chunks to vector repo
	vectorRepo.chunks = []domain.MaterialChunk{
		{
			ID:         "chunk1",
			MaterialID: materialID,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "This is relevant content about the topic.",
		},
	}

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this about?",
	}

	_, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Verify that CheckAILimit was called
	if !aiLimitChecker.checkCalled {
		t.Error("Expected CheckAILimit to be called")
	}

	// Verify that IncrementAIUsage was called after successful chat
	if !aiLimitChecker.incrementCalled {
		t.Error("Expected IncrementAIUsage to be called after successful chat")
	}
}

func TestService_Chat_NilAILimitChecker(t *testing.T) {
	sessionRepo := newMockChatSessionRepo()
	messageRepo := newMockChatMessageRepo()
	vectorRepo := newMockVectorRepo()
	embeddingClient := newMockEmbeddingClient()
	chatClient := newMockChatClient()

	// Create service with nil AILimitChecker
	svc := NewService(
		sessionRepo,
		messageRepo,
		vectorRepo,
		embeddingClient,
		chatClient,
		"http://localhost:8000",
		nil,
	)

	ctx := context.Background()
	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	// Add some chunks to vector repo
	vectorRepo.chunks = []domain.MaterialChunk{
		{
			ID:         "chunk1",
			MaterialID: materialID,
			PodID:      podID,
			ChunkIndex: 0,
			Text:       "This is relevant content about the topic.",
		},
	}

	input := ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    "What is this about?",
	}

	// Should work without AI limit checker
	result, err := svc.Chat(ctx, input)
	if err != nil {
		t.Fatalf("Chat failed with nil AILimitChecker: %v", err)
	}
	if result.Message == nil {
		t.Fatal("Expected message in result")
	}
}

// mockAILimitCheckerWithTracking tracks calls to CheckAILimit and IncrementAIUsage.
type mockAILimitCheckerWithTracking struct {
	checkCalled     bool
	incrementCalled bool
	checkErr        error
	incrementErr    error
}

func (m *mockAILimitCheckerWithTracking) CheckAILimit(ctx context.Context, userID uuid.UUID) error {
	m.checkCalled = true
	return m.checkErr
}

func (m *mockAILimitCheckerWithTracking) IncrementAIUsage(ctx context.Context, userID uuid.UUID) error {
	m.incrementCalled = true
	return m.incrementErr
}

func (m *mockAILimitCheckerWithTracking) CanAccessFeature(ctx context.Context, userID uuid.UUID, feature string) error {
	return nil
}
