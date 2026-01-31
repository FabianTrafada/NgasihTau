"use client";

import React from "react";
import { X, Lightbulb, BookOpen, Target, Users, Sparkles } from "lucide-react";
import { PredictPersonaResponse, Recommendation } from "@/lib/api/behavior";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface PersonaRecommendationPopupProps {
  isOpen: boolean;
  onClose: () => void;
  persona: PredictPersonaResponse | null;
  onActionClick?: (recommendation: Recommendation) => void;
}

// Map persona to friendly display name
const personaDisplayNames: Record<string, string> = {
  skimmer: "Quick Learner",
  struggler: "Determined Learner",
  anxious: "Careful Learner",
  burnout: "Resting Learner",
  master: "Advanced Learner",
  procrastinator: "Flexible Learner",
  deep_diver: "Thorough Learner",
  social_learner: "Collaborative Learner",
  perfectionist: "Detail-Oriented Learner",
  lost: "Exploring Learner",
  unknown: "New Learner",
};

// Map persona to color scheme
const personaColors: Record<string, { bg: string; text: string; border: string }> = {
  skimmer: { bg: "bg-blue-100", text: "text-blue-700", border: "border-blue-300" },
  struggler: { bg: "bg-orange-100", text: "text-orange-700", border: "border-orange-300" },
  anxious: { bg: "bg-purple-100", text: "text-purple-700", border: "border-purple-300" },
  burnout: { bg: "bg-gray-100", text: "text-gray-700", border: "border-gray-300" },
  master: { bg: "bg-green-100", text: "text-green-700", border: "border-green-300" },
  procrastinator: { bg: "bg-yellow-100", text: "text-yellow-700", border: "border-yellow-300" },
  deep_diver: { bg: "bg-indigo-100", text: "text-indigo-700", border: "border-indigo-300" },
  social_learner: { bg: "bg-pink-100", text: "text-pink-700", border: "border-pink-300" },
  perfectionist: { bg: "bg-teal-100", text: "text-teal-700", border: "border-teal-300" },
  lost: { bg: "bg-red-100", text: "text-red-700", border: "border-red-300" },
  unknown: { bg: "bg-gray-100", text: "text-gray-700", border: "border-gray-300" },
};

const actionIcons: Record<string, React.ReactNode> = {
  content: <BookOpen size={18} />,
  feature: <Target size={18} />,
  notification: <Lightbulb size={18} />,
  quiz: <Sparkles size={18} />,
  social: <Users size={18} />,
};

export default function PersonaRecommendationPopup({
  isOpen,
  onClose,
  persona,
  onActionClick,
}: PersonaRecommendationPopupProps) {
  if (!persona) return null;

  const displayName = personaDisplayNames[persona.persona] || "Learner";
  const colors = personaColors[persona.persona] || personaColors.unknown;
  const topRecommendations = persona.recommendations?.slice(0, 3) || [];

  const handleActionClick = (recommendation: Recommendation) => {
    if (onActionClick) {
      onActionClick(recommendation);
    }
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black] p-0 overflow-hidden">
        <div className={`${colors.bg} px-6 py-4 border-b-2 border-black`}>
          <DialogHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-full ${colors.bg} border-2 ${colors.border}`}>
                  <Lightbulb className={colors.text} size={20} />
                </div>
                <div>
                  <DialogTitle className="text-lg font-black text-black">
                    Learning Tips for You
                  </DialogTitle>
                  <p className={`text-sm font-medium ${colors.text}`}>
                    Based on your learning style: {displayName}
                  </p>
                </div>
              </div>
              <button
                onClick={onClose}
                className="p-1 rounded-full hover:bg-black/10 transition-colors"
              >
                <X size={20} className="text-black" />
              </button>
            </div>
          </DialogHeader>
        </div>

        {/* Feature Summary */}
        {persona.feature_summary && (
          <div className="px-6 py-3 bg-white border-b-2 border-gray-200">
            <div className="flex flex-wrap gap-2">
              <span className="text-xs font-medium text-gray-500">Your activity:</span>
              <span className={`text-xs px-2 py-0.5 rounded-full ${
                persona.feature_summary.chat_engagement === "high" 
                  ? "bg-green-100 text-green-700" 
                  : persona.feature_summary.chat_engagement === "medium"
                  ? "bg-yellow-100 text-yellow-700"
                  : "bg-gray-100 text-gray-700"
              }`}>
                Chat: {persona.feature_summary.chat_engagement}
              </span>
              <span className={`text-xs px-2 py-0.5 rounded-full ${
                persona.feature_summary.material_consumption === "high" 
                  ? "bg-green-100 text-green-700" 
                  : persona.feature_summary.material_consumption === "medium"
                  ? "bg-yellow-100 text-yellow-700"
                  : "bg-gray-100 text-gray-700"
              }`}>
                Reading: {persona.feature_summary.material_consumption}
              </span>
            </div>
          </div>
        )}

        {/* Recommendations */}
        <div className="px-6 py-4 space-y-3 max-h-[300px] overflow-y-auto">
          <p className="text-sm font-bold text-black">Recommended for you:</p>
          
          {topRecommendations.map((rec, index) => (
            <button
              key={rec.id}
              onClick={() => handleActionClick(rec)}
              className="w-full text-left p-4 rounded-xl border-2 border-black bg-white hover:bg-gray-50 shadow-[3px_3px_0_0_black] hover:shadow-[2px_2px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all"
            >
              <div className="flex items-start gap-3">
                <div className={`p-2 rounded-lg ${colors.bg} ${colors.text} flex-shrink-0`}>
                  {actionIcons[rec.action_type] || <Lightbulb size={18} />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="font-bold text-black text-sm">{rec.title}</h4>
                    {index === 0 && (
                      <span className="text-[10px] px-1.5 py-0.5 bg-orange-500 text-white rounded font-bold">
                        TOP
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-gray-600 mt-1 line-clamp-2">
                    {rec.description}
                  </p>
                </div>
              </div>
            </button>
          ))}
        </div>

        {/* Footer */}
        <div className="px-6 py-3 bg-gray-50 border-t-2 border-gray-200">
          <button
            onClick={onClose}
            className="w-full py-2 text-sm font-bold text-gray-500 hover:text-gray-700 transition-colors"
          >
            Maybe later
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
