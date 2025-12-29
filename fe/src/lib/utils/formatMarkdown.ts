/**
 * Formatter utility untuk mengubah markdown-like text menjadi formatted object
 * Handles: \n, **, *, numbered lists, dll
 */

export interface MarkdownBlock {
  type: "paragraph" | "list" | "heading";
  content: string | string[] | ListItem[];
}

export interface ListItem {
  text: string;
  nested?: ListItem[];
}

/**
 * Parse markdown-like string menjadi array of blocks
 * Handles:
 * - Line breaks (\n)
 * - Bold text (**text**)
 * - Bullet points (* item)
 * - Numbered lists (1. item)
 * - Headings (### heading)
 */
export function parseMarkdown(text: string): MarkdownBlock[] {
  if (!text) return [];

  // Split by double newline untuk paragraphs/blocks
  const blocks = text.split(/\n(?=\n)|\n\n/).filter(Boolean);

  return blocks.flatMap((block) => {
    const trimmedBlock = block.trim();
    const lines = trimmedBlock.split(/\n/);
    const subBlocks: MarkdownBlock[] = [];
    let currentListItems: ListItem[] = [];
    let currentParagraph: string[] = [];

    lines.forEach((line) => {
      const trimmedLine = line.trim();
      const leadingSpaces = line.search(/\S/); // Count leading spaces

      // Check if line is a list item
      if (/^[\*\-]\s+/.test(trimmedLine) || /^\d+\.\s+/.test(trimmedLine)) {
        // If we have paragraph content, save it first
        if (currentParagraph.length > 0) {
          subBlocks.push({
            type: "paragraph",
            content: currentParagraph.join(" ").trim(),
          });
          currentParagraph = [];
        }

        // Extract list item content (remove bullet/number)
        const itemContent = trimmedLine.replace(/^[\*\-\d+.]\s+/, "").trim();
        if (itemContent) {
          // Create new list item with indentation info
          const listItem: ListItem = { text: itemContent };

          // Determine if this is nested (more than 2 spaces of indentation)
          if (leadingSpaces > 4 && currentListItems.length > 0) {
            // Add as nested item to last parent
            const lastItem = currentListItems[currentListItems.length - 1];
            if (!lastItem.nested) {
              lastItem.nested = [];
            }
            lastItem.nested.push(listItem);
          } else {
            // Add as parent item
            currentListItems.push(listItem);
          }
        }
      } else if (trimmedLine.length > 0) {
        // Non-empty line that's not a list item
        // If we have list items, save them first
        if (currentListItems.length > 0) {
          subBlocks.push({
            type: "list",
            content: currentListItems,
          });
          currentListItems = [];
        }
        currentParagraph.push(trimmedLine);
      }
    });

    // Save remaining paragraph content
    if (currentParagraph.length > 0) {
      subBlocks.push({
        type: "paragraph",
        content: currentParagraph.join(" ").trim(),
      });
    }

    // Save remaining list items
    if (currentListItems.length > 0) {
      subBlocks.push({
        type: "list",
        content: currentListItems,
      });
    }

    return subBlocks.length > 0
      ? subBlocks
      : [
          {
            type: "paragraph",
            content: trimmedBlock,
          },
        ];
  });
}

/**
 * Format AI response untuk display yang lebih readable
 * Main entry point untuk digunakan di React components
 */
export function formatAIResponse(text: string): {
  blocks: MarkdownBlock[];
  formatted: boolean;
} {
  try {
    // Remove excessive newlines
    const cleaned = text
      .replace(/\\n/g, "\n") // Convert literal \n to actual newlines
      .replace(/\n{3,}/g, "\n\n") // Replace 3+ newlines with 2
      .trim();

    const blocks = parseMarkdown(cleaned);

    return {
      blocks,
      formatted: true,
    };
  } catch (error) {
    console.error("Error formatting markdown:", error);
    return {
      blocks: [
        {
          type: "paragraph",
          content: text,
        },
      ],
      formatted: false,
    };
  }
}

/**
 * Helper: Check apakah text contains markdown formatting
 */
export function hasMarkdownFormatting(text: string): boolean {
  return /\*\*.*\*\*|\n\*\s|\n\d+\.\s|^#+\s/m.test(text);
}

/**
 * Helper: Convert block to plain text (for debugging)
 */
export function blockToPlainText(block: MarkdownBlock): string {
  if (block.type === "list") {
    return (block.content as string[]).map((item) => `• ${item}`).join("\n");
  }
  return block.content as string;
}
