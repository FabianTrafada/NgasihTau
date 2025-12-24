import React from "react";
import { MessageSquare } from "lucide-react";

interface ChatbotLogCardProps {
  title: string;
  snippet: string;
}

const ChatbotLogCard = ({ title, snippet }: ChatbotLogCardProps) => {
  return (
    <div className="bg-white border-2 border-[#2B2D42] p-5 flex items-center justify-between shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all group">
      <div className="flex gap-4 items-start">
        <div className="w-10 h-10 bg-gray-50 flex items-center justify-center flex-shrink-0 border-2 border-[#2B2D42]">
          <MessageSquare className="w-5 h-5 text-[#2B2D42]" />
        </div>
        <div>
          <h3 className="font-bold text-[#2B2D42] mb-1">{title}</h3>
          <div className="flex items-center gap-2 text-xs text-gray-500 font-mono bg-gray-50 px-2 py-1 border border-gray-200">
            <span className="text-[#FF8811] font-bold">{">"}</span>
            <span className="truncate max-w-[200px]">{snippet}</span>
          </div>
        </div>
      </div>

      <button className="px-4 py-2 border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-[2px] hover:translate-y-[2px]">
        Continue
      </button>
    </div>
  );
};

export default ChatbotLogCard;