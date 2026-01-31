# AI Chat API Documentation

Dokumentasi lengkap untuk menggunakan AI Chat API di frontend NgasihTau.

## 📋 Table of Contents

- [Overview](#overview)
- [API Functions](#api-functions)
- [Custom Hook](#custom-hook)
- [Usage Examples](#usage-examples)
- [Type Definitions](#type-definitions)

---

## 🎯 Overview

AI Chat API menyediakan fungsi-fungsi untuk:
- Chat dengan Material atau Pod
- Mendapatkan history chat
- Mendapatkan suggested questions
- Export chat history
- Submit feedback untuk chat messages

**Base URL:** `api/v1/...` (automatically handled by apiClient)

---

## 📡 API Functions

### Material Chat

#### `chatWithMaterial(materialId, request)`
Mengirim pesan ke material chat.

```typescript
import { chatWithMaterial } from "@/lib/api/ai";

const response = await chatWithMaterial("material-123", {
  message: "Jelaskan tentang topik ini",
  session_id: "optional-session-id"
});
```

#### `getMaterialChatHistory(materialId, limit?, offset?)`
Mendapatkan history chat material.

```typescript
import { getMaterialChatHistory } from "@/lib/api/ai";

const history = await getMaterialChatHistory("material-123", 20, 0);
```

#### `getMaterialChatSuggestions(materialId)`
Mendapatkan suggested questions untuk material.

```typescript
import { getMaterialChatSuggestions } from "@/lib/api/ai";

const suggestions = await getMaterialChatSuggestions("material-123");
```

#### `exportMaterialChatHistory(materialId, format?)`
Export chat history material.

```typescript
import { exportMaterialChatHistory, downloadChatExport } from "@/lib/api/ai";

const exportData = await exportMaterialChatHistory("material-123", "pdf");
downloadChatExport(exportData.data, "my-chat-history");
```

### Pod Chat

#### `chatWithPod(podId, request)`
Mengirim pesan ke pod chat.

```typescript
import { chatWithPod } from "@/lib/api/ai";

const response = await chatWithPod("pod-123", {
  message: "Apa isi pod ini?",
  session_id: "optional-session-id"
});
```

#### `getPodChatHistory(podId, limit?, offset?)`
Mendapatkan history chat pod.

```typescript
import { getPodChatHistory } from "@/lib/api/ai";

const history = await getPodChatHistory("pod-123", 20, 0);
```

#### `getPodChatSuggestions(podId)`
Mendapatkan suggested questions untuk pod.

```typescript
import { getPodChatSuggestions } from "@/lib/api/ai";

const suggestions = await getPodChatSuggestions("pod-123");
```

#### `exportPodChatHistory(podId, format?)`
Export chat history pod.

```typescript
import { exportPodChatHistory, downloadChatExport } from "@/lib/api/ai";

const exportData = await exportPodChatHistory("pod-123", "pdf");
downloadChatExport(exportData.data, "pod-chat-history");
```

### Feedback

#### `submitChatFeedback(messageId, feedback)`
Submit feedback untuk chat message.

```typescript
import { submitChatFeedback } from "@/lib/api/ai";

await submitChatFeedback("message-123", {
  rating: 5,
  comment: "Very helpful!",
  feedback_type: "helpful"
});
```

---

## 🎣 Custom Hook

### `useAIChat(options)`

Custom hook yang menyediakan interface lengkap untuk AI chat.

**Options:**
```typescript
{
  type: "material" | "pod";
  id: string;
  autoLoadHistory?: boolean;  // default: false
  historyLimit?: number;       // default: 20
}
```

**Returns:**
```typescript
{
  // State
  messages: ChatMessage[];
  suggestions: string[];
  isLoading: boolean;
  isSending: boolean;
  isLoadingHistory: boolean;
  isLoadingSuggestions: boolean;
  error: string | null;
  sessionId: string | null;

  // Actions
  sendMessage: (message: string) => Promise<void>;
  loadHistory: (limit?, offset?) => Promise<void>;
  loadSuggestions: () => Promise<void>;
  exportHistory: (format?) => Promise<void>;
  submitFeedback: (messageId, feedback) => Promise<void>;
  clearMessages: () => void;
  clearError: () => void;
}
```

---

## 💡 Usage Examples

### Example 1: Basic Chat Component

```typescript
"use client";

import { useAIChat } from "@/hooks/useAIChat";
import { useState } from "react";

export default function MaterialChatPage({ params }: { params: { id: string } }) {
  const [input, setInput] = useState("");
  
  const {
    messages,
    suggestions,
    isSending,
    isLoadingSuggestions,
    sendMessage,
    loadSuggestions,
  } = useAIChat({
    type: "material",
    id: params.id,
    autoLoadHistory: true,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      await sendMessage(input);
      setInput("");
    }
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {messages.map((msg) => (
          <div key={msg.id} className={msg.role === "user" ? "text-right" : "text-left"}>
            <p>{msg.content}</p>
          </div>
        ))}
      </div>

      {/* Suggestions */}
      {suggestions.length > 0 && (
        <div className="p-4 border-t">
          <p className="text-sm font-medium mb-2">Suggested Questions:</p>
          {suggestions.map((question, idx) => (
            <button
              key={idx}
              onClick={() => sendMessage(question)}
              className="block text-left text-sm p-2 hover:bg-gray-100 rounded"
            >
              {question}
            </button>
          ))}
        </div>
      )}

      {/* Input */}
      <form onSubmit={handleSubmit} className="p-4 border-t">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type your message..."
          disabled={isSending}
          className="w-full p-2 border rounded"
        />
      </form>
    </div>
  );
}
```

### Example 2: With Feedback

```typescript
"use client";

import { useAIChat } from "@/hooks/useAIChat";
import { ThumbsUp, ThumbsDown } from "lucide-react";

export default function PodChatPage({ params }: { params: { id: string } }) {
  const { messages, sendMessage, submitFeedback } = useAIChat({
    type: "pod",
    id: params.id,
  });

  const handleFeedback = async (messageId: string, isHelpful: boolean) => {
    await submitFeedback(messageId, {
      rating: isHelpful ? 5 : 1,
      feedback_type: isHelpful ? "helpful" : "not_helpful",
    });
  };

  return (
    <div className="space-y-4">
      {messages.map((msg) => (
        <div key={msg.id}>
          <p>{msg.content}</p>
          {msg.role === "assistant" && (
            <div className="flex gap-2 mt-2">
              <button onClick={() => handleFeedback(msg.id, true)}>
                <ThumbsUp size={16} />
              </button>
              <button onClick={() => handleFeedback(msg.id, false)}>
                <ThumbsDown size={16} />
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
```

### Example 3: Export Chat History

```typescript
"use client";

import { useAIChat } from "@/hooks/useAIChat";
import { Download } from "lucide-react";

export default function ChatWithExport({ materialId }: { materialId: string }) {
  const { messages, exportHistory, isLoading } = useAIChat({
    type: "material",
    id: materialId,
    autoLoadHistory: true,
  });

  return (
    <div>
      <button
        onClick={() => exportHistory("pdf")}
        disabled={isLoading}
        className="flex items-center gap-2"
      >
        <Download size={16} />
        {isLoading ? "Exporting..." : "Export Chat"}
      </button>

      <div className="mt-4">
        {messages.map((msg) => (
          <div key={msg.id}>{msg.content}</div>
        ))}
      </div>
    </div>
  );
}
```

### Example 4: Direct API Usage (Without Hook)

```typescript
import {
  chatWithMaterial,
  getMaterialChatHistory,
  submitChatFeedback,
} from "@/lib/api/ai";

// Send a message
async function sendChatMessage() {
  const response = await chatWithMaterial("material-123", {
    message: "Explain this concept",
  });
  console.log(response.data.message);
}

// Get history
async function loadHistory() {
  const history = await getMaterialChatHistory("material-123", 10, 0);
  console.log(history.data);
}

// Submit feedback
async function giveFeedback() {
  await submitChatFeedback("message-id", {
    rating: 5,
    feedback_type: "helpful",
  });
}
```

---

## 📝 Type Definitions

### ChatMessage
```typescript
interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  session_id?: string;
}
```

### ChatRequest
```typescript
interface ChatRequest {
  message: string;
  session_id?: string;
}
```

### ChatResponse
```typescript
interface ChatResponse {
  data: {
    message: string;
    session_id: string;
    timestamp: string;
  };
}
```

### SuggestedQuestion
```typescript
interface SuggestedQuestion {
  question: string;
  category?: string;
}
```

### FeedbackRequest
```typescript
interface FeedbackRequest {
  rating: number; // 1-5
  comment?: string;
  feedback_type?: "helpful" | "not_helpful" | "incorrect" | "other";
}
```

---

## 🔧 Error Handling

Semua API functions akan throw error jika request gagal. Gunakan try-catch untuk handling:

```typescript
try {
  await sendMessage("Hello");
} catch (error: any) {
  console.error("Error:", error?.response?.data?.message);
}
```

Hook `useAIChat` automatically handles errors dan menyimpannya di state `error`:

```typescript
const { error, clearError } = useAIChat({ ... });

if (error) {
  return <div>{error}</div>;
}
```

---

## 🎨 Best Practices

1. **Use the Hook:** Prefer `useAIChat` hook untuk component-level logic
2. **Error Handling:** Always handle errors gracefully
3. **Loading States:** Show loading indicators saat request sedang berjalan
4. **Optimistic UI:** Messages di-add immediately untuk better UX
5. **Session Management:** Hook automatically manages session IDs
6. **Clean Code:** Pisahkan business logic dari UI components

---

## 🚀 Quick Start

1. Import hook:
```typescript
import { useAIChat } from "@/hooks/useAIChat";
```

2. Initialize dalam component:
```typescript
const chat = useAIChat({
  type: "material",
  id: materialId,
  autoLoadHistory: true,
});
```

3. Use dalam JSX:
```typescript
<button onClick={() => chat.sendMessage("Hello")}>
  Send
</button>
```

Done! 🎉
