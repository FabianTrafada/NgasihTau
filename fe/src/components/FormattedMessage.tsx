/**
 * FormattedMessage Component
 * Render AI chat messages dengan proper markdown formatting
 */

import React from "react";
import { formatAIResponse, MarkdownBlock, ListItem } from "@/lib/utils/formatMarkdown";

interface FormattedMessageProps {
  content: string;
  role: "user" | "assistant";
  className?: string;
}

/**
 * Component untuk render list items (dengan support nested lists)
 */
const ListItemComponent: React.FC<{ item: ListItem }> = ({ item }) => {
  return (
    <>
      <li className="text-xs leading-relaxed">
        <InlineFormattedText text={item.text} />
      </li>
      {item.nested && item.nested.length > 0 && (
        <ul className="list-disc list-inside space-y-1 pl-4 mt-1">
          {item.nested.map((nestedItem, idx) => (
            <ListItemComponent key={`nested-${idx}`} item={nestedItem} />
          ))}
        </ul>
      )}
    </>
  );
};

/**
 * Component untuk render formatted markdown content
 */
const FormattedContent: React.FC<{ blocks: MarkdownBlock[] }> = ({ blocks }) => {
  return (
    <div className="space-y-2">
      {blocks.map((block, blockIdx) => {
        if (block.type === "heading") {
          return (
            <h4 key={`heading-${blockIdx}`} className="font-bold text-sm text-gray-800">
              {typeof block.content === "string" ? block.content : ""}
            </h4>
          );
        }

        if (block.type === "list") {
          // Check if content is ListItem[] (new format) or string[] (old format)
          const isListItems = Array.isArray(block.content) && block.content.length > 0 && typeof block.content[0] === "object" && "text" in block.content[0];

          return (
            <ul key={`list-${blockIdx}`} className="list-disc list-inside space-y-1 pl-2">
              {isListItems
                ? (block.content as ListItem[]).map((item, itemIdx) => <ListItemComponent key={`item-${blockIdx}-${itemIdx}`} item={item} />)
                : (block.content as string[]).map((item, itemIdx) => (
                    <li key={`item-${blockIdx}-${itemIdx}`} className="text-xs leading-relaxed">
                      <InlineFormattedText text={item} />
                    </li>
                  ))}
            </ul>
          );
        }

        if (block.type === "paragraph") {
          return (
            <p key={`para-${blockIdx}`} className="text-xs leading-relaxed">
              <InlineFormattedText text={typeof block.content === "string" ? block.content : ""} />
            </p>
          );
        }

        return null;
      })}
    </div>
  );
};

/**
 * Component untuk handle inline formatting (bold, italic, dll)
 */
const InlineFormattedText: React.FC<{ text: string }> = ({ text }) => {
  if (!text) return null;

  // Split by ** untuk bold
  const parts = text.split(/(\*\*[^*]+\*\*)/);

  return (
    <span className="overflow-wrap break-word">
      {parts.map((part, idx) => {
        if (part.startsWith("**") && part.endsWith("**")) {
          // Bold text
          return (
            <strong key={`bold-${idx}`} className="font-bold">
              {part.slice(2, -2)}
            </strong>
          );
        }
        return <span key={`text-${idx}`}>{part}</span>;
      })}
    </span>
  );
};

/**
 * Main FormattedMessage Component
 */
export const FormattedMessage: React.FC<FormattedMessageProps> = ({ content, role, className = "" }) => {
  const { blocks, formatted } = formatAIResponse(content);

  return (
    <div className={`p-3 rounded-lg text-xs max-w-[85%] leading-relaxed ${role === "user" ? "bg-[#2B2D42] text-white" : "bg-gray-100 text-gray-700 border-2 border-[#2B2D42]"} ${className}`}>
      {formatted ? <FormattedContent blocks={blocks} /> : <p>{content}</p>}
    </div>
  );
};

export default FormattedMessage;
