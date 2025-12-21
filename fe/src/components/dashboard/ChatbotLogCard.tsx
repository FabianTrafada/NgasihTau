import React from "react";
import { MessageSquare } from "lucide-react";

interface ChatbotLogCardProps {
  title: string;
  snippet: string;
}

const ChatbotLogCard = ({ title, snippet }: ChatbotLogCardProps) => {
  return (
    <div className="bg-white border border-gray-200 rounded-xl p-5 flex items-center justify-between hover:border-[#FF8811]/50 transition-colors group">
      <div className="flex gap-4 items-start">
        <div className="w-10 h-10 rounded-lg bg-gray-50 flex items-center justify-center flex-shrink-0 border border-gray-100">
          <MessageSquare className="w-5 h-5 text-gray-600" />
        </div>
        <div>
          <h3 className="font-bold text-gray-900 mb-1">{title}</h3>
          <div className="flex items-center gap-2 text-xs text-gray-500 font-mono bg-gray-50 px-2 py-1 rounded border border-gray-100">
            <span className="text-gray-400">{">"}</span>
            <span className="truncate max-w-[200px]">{snippet}</span>
          </div>
        </div>
      </div>
      
      <button className="px-4 py-2 rounded-lg border border-gray-200 text-sm font-bold text-gray-700 hover:bg-gray-900 hover:text-white hover:border-gray-900 transition-all">
        Continue
      </button>
    </div>
  );
};

export default ChatbotLogCard;