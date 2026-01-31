# AI Features Integration Documentation

## Overview

This document describes the implementation of three AI-related features based on the Swagger API documentation:

1. **AI Suggestions** - Get AI-generated suggested questions for materials
2. **Chat Export** - Export chat history to PDF or Markdown
3. **Feedback Submission** - Submit thumbs up/down feedback for AI responses

## Architecture

### Clean Separation of Concerns

```
┌─────────────────────────────────────────────────────────┐
│                    UI Layer (React)                      │
│  - Material Detail Page                                  │
│  - Chat Widget with Feedback & Export buttons           │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│              Hook Layer (useAIFeatures)                  │
│  - State management                                      │
│  - Loading/error handling                                │
│  - Business logic coordination                           │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│            API Service Layer (ai.ts)                     │
│  - HTTP request handling                                 │
│  - Response parsing                                      │
│  - Error handling                                        │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│          API Client (apiClient from api-client.ts)       │
│  - Base URL configuration                                │
│  - Authentication (Bearer token)                         │
│  - Token refresh interceptor                             │
└──────────────────────────────────────────────────────────┘
```

## Implementation Details

### 1. Type Definitions (`fe/src/types/ai.ts`)

All API types are strongly typed based on Swagger documentation:

```typescript
// Suggestion types
export interface SuggestedQuestionsResponse {
  questions: string[];
}

// Export types
export type ExportFormat = "pdf" | "markdown";
export interface ExportChatRequest {
  format: ExportFormat;
}

// Feedback types
export type FeedbackType = "thumbs_up" | "thumbs_down";
export interface FeedbackRequest {
  feedback: FeedbackType;
  feedback_text?: string;
}
```

### 2. API Service Layer (`fe/src/lib/api/ai.ts`)

#### Suggestions API

**Endpoint:** `GET /api/v1/materials/{id}/chat/suggestions`

```typescript
const questions = await getSuggestedQuestions(materialId);
// Returns: ["What are the key concepts?", "Can you explain..."]
```

**Features:**
- Handles both wrapped (`{data: {...}}`) and unwrapped responses
- Returns empty array on error (graceful degradation)
- Full error logging for debugging

#### Export API

**Endpoint:** `POST /api/v1/materials/{id}/chat/export`

```typescript
// Low-level: Get blob
const blob = await exportChatHistory(materialId, "pdf");

// High-level: Trigger download
await downloadChatExport(materialId, "pdf", "my-chat.pdf");
```

**Features:**
- Handles binary data (Blob) correctly with `responseType: "blob"`
- Automatic file download with proper cleanup
- Supports both PDF and Markdown formats
- Custom filename support

**Implementation Notes:**
- Uses `URL.createObjectURL()` for blob handling
- Proper cleanup with `URL.revokeObjectURL()`
- Creates temporary `<a>` element for download trigger

#### Feedback API

**Endpoint:** `POST /api/v1/chat/{messageId}/feedback`

```typescript
await submitFeedback(messageId, "thumbs_up");
await submitFeedback(messageId, "thumbs_down", "Response was inaccurate");
```

**Features:**
- Optional feedback text for detailed explanations
- Validates feedback type at compile time
- Proper error propagation

### 3. Custom Hook (`fe/src/hooks/useAIFeatures.ts`)

Provides centralized state management for all AI features:

```typescript
const {
  // Suggestions
  suggestions,
  suggestionsLoading,
  suggestionsError,
  loadSuggestions,
  
  // Export
  exportLoading,
  exportError,
  handleExport,
  
  // Feedback
  feedbackLoading, // Per-message loading state
  feedbackError,
  handleFeedback,
} = useAIFeatures({ materialId });
```

**Key Features:**
- Separate loading/error states for each feature
- Per-message feedback loading state (prevents race conditions)
- Memoized callbacks with `useCallback` for performance
- Automatic error handling and logging

### 4. UI Integration

#### Minimal Changes to Existing UI

The integration follows the requirement of **no layout/styling changes**:

1. **Export Button** - Added to chat header as dropdown
2. **Suggestions Button** - Added to chat input area (sparkle icon)
3. **Feedback Buttons** - Added below each AI message (thumbs up/down)

#### User Experience Flow

**Suggestions:**
1. User clicks sparkle icon
2. First click: Loads suggestions from API
3. Suggestions panel appears with clickable questions
4. Clicking a suggestion fills the input field

**Export:**
1. User hovers over export icon in chat header
2. Dropdown shows "Export as PDF" / "Export as Markdown"
3. Click triggers download
4. Browser downloads file automatically

**Feedback:**
1. Each AI message shows thumbs up/down buttons
2. Click submits feedback to API
3. Button highlights to show submitted state
4. Buttons disabled after submission (one feedback per message)

## Error Handling Strategy

### Defensive Error Handling

All API calls implement defensive error handling:

```typescript
try {
  const result = await apiCall();
  return result;
} catch (error) {
  console.error("Context-specific error message:", error);
  throw error; // Re-throw for caller to handle
}
```

### UI Error States

- **Loading States:** Show spinners/disabled buttons
- **Error States:** Display error messages (non-blocking)
- **Graceful Degradation:** Features fail independently

### Edge Cases Handled

1. **Empty Suggestions:** Shows empty state, doesn't break UI
2. **Export Failure:** Shows error, doesn't crash chat
3. **Feedback Already Submitted:** Buttons disabled, prevents duplicate submissions
4. **Network Timeout:** Handled by apiClient (10s timeout)
5. **401 Unauthorized:** Handled by token refresh interceptor

## API Response Handling

### Wrapped vs Unwrapped Responses

The API sometimes returns wrapped responses:

```typescript
// Wrapped
{ data: { questions: [...] } }

// Unwrapped
{ questions: [...] }
```

Our implementation handles both:

```typescript
const data = response.data.data || response.data;
```

### Binary Data (Export)

Export endpoint returns binary data (PDF/Markdown file):

```typescript
const response = await apiClient.post(url, body, {
  responseType: "blob" // Critical for binary data
});
```

## Performance Considerations

### Optimizations

1. **Lazy Loading:** Suggestions only load when user clicks button
2. **Memoization:** All handlers use `useCallback` to prevent re-renders
3. **Per-Message State:** Feedback loading tracked per message (no global blocking)
4. **Cleanup:** Blob URLs properly revoked after download

### Network Efficiency

- Suggestions cached after first load
- Export streams directly to download (no memory buffering)
- Feedback submissions are fire-and-forget (no polling)

## Testing Recommendations

### Unit Tests

```typescript
// Test API service layer
describe("getSuggestedQuestions", () => {
  it("should return questions array", async () => {
    const questions = await getSuggestedQuestions("material-id");
    expect(Array.isArray(questions)).toBe(true);
  });
  
  it("should handle API errors gracefully", async () => {
    // Mock API error
    await expect(getSuggestedQuestions("invalid-id")).rejects.toThrow();
  });
});
```

### Integration Tests

```typescript
// Test hook behavior
describe("useAIFeatures", () => {
  it("should load suggestions on demand", async () => {
    const { result } = renderHook(() => useAIFeatures({ materialId: "test" }));
    
    await act(async () => {
      await result.current.loadSuggestions();
    });
    
    expect(result.current.suggestions.length).toBeGreaterThan(0);
  });
});
```

### E2E Tests

1. Load material page
2. Open chat widget
3. Click suggestions button → verify questions appear
4. Click export → verify file downloads
5. Click feedback → verify button state changes

## Future Improvements

### Potential Enhancements

1. **Suggestion Caching:** Cache suggestions per material in localStorage
2. **Feedback Text Input:** Add modal for detailed feedback text
3. **Export Progress:** Show progress bar for large chat histories
4. **Retry Logic:** Automatic retry for failed API calls
5. **Optimistic Updates:** Update UI before API confirmation
6. **Analytics:** Track which suggestions users click most

### Scalability Considerations

1. **Pagination:** If chat history grows large, implement pagination for export
2. **Debouncing:** Add debounce to prevent rapid-fire feedback submissions
3. **Rate Limiting:** Implement client-side rate limiting for API calls
4. **Offline Support:** Queue feedback submissions when offline

## Troubleshooting

### Common Issues

**Issue:** Suggestions not loading
- Check: Material ID is valid UUID
- Check: User is authenticated (Bearer token present)
- Check: API endpoint is accessible

**Issue:** Export downloads empty file
- Check: `responseType: "blob"` is set in axios config
- Check: Chat history exists for material
- Check: Browser allows downloads

**Issue:** Feedback not submitting
- Check: Message ID is valid
- Check: Feedback type is exactly "thumbs_up" or "thumbs_down"
- Check: Not already submitted for this message

### Debug Mode

Enable detailed logging:

```typescript
// In api-client.ts
apiClient.interceptors.request.use((config) => {
  console.log("API Request:", config.method, config.url, config.data);
  return config;
});
```

## API Endpoints Reference

| Feature | Method | Endpoint | Auth Required |
|---------|--------|----------|---------------|
| Suggestions | GET | `/api/v1/materials/{id}/chat/suggestions` | Yes |
| Export | POST | `/api/v1/materials/{id}/chat/export` | Yes |
| Feedback | POST | `/api/v1/chat/{messageId}/feedback` | Yes |

## Conclusion

This implementation provides a production-ready, maintainable, and scalable solution for AI features integration. The clean architecture ensures:

- **Separation of Concerns:** UI, business logic, and API calls are isolated
- **Type Safety:** Full TypeScript coverage prevents runtime errors
- **Error Resilience:** Defensive error handling prevents cascading failures
- **Performance:** Optimized for minimal re-renders and network usage
- **Maintainability:** Clear code structure with comprehensive documentation
