package application

import (
	"strings"
	"unicode"
)

type ChunkerConfig struct {
	TargetChunkSize int
	MaxChunkSize    int
	OverlapSize     int
}

func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		TargetChunkSize: 750, // Middle of 500-1000 range
		MaxChunkSize:    1000,
		OverlapSize:     100, // ~10-15% overlap for context
	}
}

type Chunker struct {
	config ChunkerConfig
}

func NewChunker(config ChunkerConfig) *Chunker {
	return &Chunker{config: config}
}

type Chunk struct {
	Index      int
	Text       string
	TokenCount int
}

func (c *Chunker) ChunkText(text string) []Chunk {
	if text == "" {
		return nil
	}

	text = normalizeWhitespace(text)

	paragraphs := splitIntoParagraphs(text)

	var chunks []Chunk
	var currentChunk strings.Builder
	currentTokens := 0
	chunkIndex := 0

	for _, para := range paragraphs {
		paraTokens := estimateTokenCount(para)

		if paraTokens > c.config.MaxChunkSize {
			if currentTokens > 0 {
				chunks = append(chunks, Chunk{
					Index:      chunkIndex,
					Text:       strings.TrimSpace(currentChunk.String()),
					TokenCount: currentTokens,
				})
				chunkIndex++
				currentChunk.Reset()
				currentTokens = 0
			}

			sentenceChunks := c.splitBySentences(para)
			for _, sc := range sentenceChunks {
				sc.Index = chunkIndex
				chunks = append(chunks, sc)
				chunkIndex++
			}
			continue
		}

		if currentTokens+paraTokens > c.config.TargetChunkSize && currentTokens > 0 {
			chunks = append(chunks, Chunk{
				Index:      chunkIndex,
				Text:       strings.TrimSpace(currentChunk.String()),
				TokenCount: currentTokens,
			})
			chunkIndex++

			overlapText := c.getOverlapText(currentChunk.String())
			currentChunk.Reset()
			if overlapText != "" {
				currentChunk.WriteString(overlapText)
				currentChunk.WriteString("\n\n")
				currentTokens = estimateTokenCount(overlapText)
			} else {
				currentTokens = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentTokens += paraTokens
	}

	if currentTokens > 0 {
		chunks = append(chunks, Chunk{
			Index:      chunkIndex,
			Text:       strings.TrimSpace(currentChunk.String()),
			TokenCount: currentTokens,
		})
	}

	return chunks
}

func (c *Chunker) splitBySentences(text string) []Chunk {
	sentences := splitIntoSentences(text)
	var chunks []Chunk
	var currentChunk strings.Builder
	currentTokens := 0

	for _, sentence := range sentences {
		sentenceTokens := estimateTokenCount(sentence)

		if sentenceTokens > c.config.MaxChunkSize {
			if currentTokens > 0 {
				chunks = append(chunks, Chunk{
					Text:       strings.TrimSpace(currentChunk.String()),
					TokenCount: currentTokens,
				})
				currentChunk.Reset()
				currentTokens = 0
			}

			wordChunks := c.splitByWords(sentence)
			chunks = append(chunks, wordChunks...)
			continue
		}

		if currentTokens+sentenceTokens > c.config.TargetChunkSize && currentTokens > 0 {
			chunks = append(chunks, Chunk{
				Text:       strings.TrimSpace(currentChunk.String()),
				TokenCount: currentTokens,
			})

			overlapText := c.getOverlapText(currentChunk.String())
			currentChunk.Reset()
			if overlapText != "" {
				currentChunk.WriteString(overlapText)
				currentChunk.WriteString(" ")
				currentTokens = estimateTokenCount(overlapText)
			} else {
				currentTokens = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
		currentTokens += sentenceTokens
	}

	if currentTokens > 0 {
		chunks = append(chunks, Chunk{
			Text:       strings.TrimSpace(currentChunk.String()),
			TokenCount: currentTokens,
		})
	}

	return chunks
}

func (c *Chunker) splitByWords(text string) []Chunk {
	words := strings.Fields(text)
	var chunks []Chunk
	var currentChunk strings.Builder
	currentTokens := 0

	for _, word := range words {
		wordTokens := estimateTokenCount(word)

		if currentTokens+wordTokens > c.config.TargetChunkSize && currentTokens > 0 {
			chunks = append(chunks, Chunk{
				Text:       strings.TrimSpace(currentChunk.String()),
				TokenCount: currentTokens,
			})

			overlapText := c.getOverlapText(currentChunk.String())
			currentChunk.Reset()
			if overlapText != "" {
				currentChunk.WriteString(overlapText)
				currentChunk.WriteString(" ")
				currentTokens = estimateTokenCount(overlapText)
			} else {
				currentTokens = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(word)
		currentTokens += wordTokens
	}

	if currentTokens > 0 {
		chunks = append(chunks, Chunk{
			Text:       strings.TrimSpace(currentChunk.String()),
			TokenCount: currentTokens,
		})
	}

	return chunks
}

func (c *Chunker) getOverlapText(text string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	overlapWords := c.config.OverlapSize / 2 // Rough estimate: ~2 tokens per word
	if overlapWords > len(words) {
		overlapWords = len(words) / 4 // At most 25% of the chunk
	}
	if overlapWords < 5 {
		overlapWords = 5
	}
	if overlapWords > len(words) {
		return ""
	}

	return strings.Join(words[len(words)-overlapWords:], " ")
}

func estimateTokenCount(text string) int {
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	charCount := len(text)

	wordCount := len(words)
	if wordCount == 0 {
		return charCount / 4
	}

	avgWordLen := float64(charCount) / float64(wordCount)

	var tokenEstimate float64
	if avgWordLen <= 4 {
		tokenEstimate = float64(wordCount)
	} else if avgWordLen <= 8 {
		tokenEstimate = float64(wordCount) * 1.3
	} else {
		tokenEstimate = float64(wordCount) * 1.5
	}

	return int(tokenEstimate)
}

func normalizeWhitespace(text string) string {
	var result strings.Builder
	prevSpace := false
	prevNewline := false
	newlineCount := 0

	for _, r := range text {
		if r == '\n' || r == '\r' {
			if !prevNewline {
				newlineCount = 1
				prevNewline = true
			} else {
				newlineCount++
			}
			prevSpace = false
			continue
		}

		// Flush newlines
		if prevNewline {
			if newlineCount >= 2 {
				result.WriteString("\n\n")
			} else {
				result.WriteString(" ")
			}
			prevNewline = false
			newlineCount = 0
		}

		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

func splitIntoParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")
	var paragraphs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			paragraphs = append(paragraphs, p)
		}
	}
	return paragraphs
}

func splitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])

		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) || unicode.IsUpper(runes[i+1]) {
				sentence := strings.TrimSpace(current.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				current.Reset()
			}
		}
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		sentences = append(sentences, remaining)
	}

	return sentences
}
