"use client";

import React, { useState, useEffect, useRef } from "react";
import { ProtectedRoute } from "@/components/auth";
import { useRouter } from "next/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ChevronLeft, Loader, Send, MessageCircle, X, Minimize2, Maximize2, Sparkles, Download as DownloadIcon, FileText } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { getMaterialDetail, getMaterialChatHistory, sendMaterialChatMessage, getMaterialPreviewUrl, getMaterialDownloadUrl } from "@/lib/api/material";
import { getUserDetail } from "@/lib/api/user";
import { getMaterialChatSuggestions } from "@/lib/api/ai";
import { Material, ChatMessage } from "@/types/material";
import { FormattedMessage } from "@/components/FormattedMessage";

interface PageProps {
  params: Promise<{
    username: string;
    pod_id: string;
    material_id: string;
  }>;
}

export default function MaterialDetailPage({ params }: PageProps) {
  const router = useRouter();
  const locale = useLocale();
  const t = useTranslations("material");
  const tCommon = useTranslations("common");
  const { username, pod_id, material_id } = React.use(params);

  // State untuk material data
  const [material, setMaterial] = useState<Material | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(true);

  const [error, setError] = useState<string | null>(null);
  const [messageInput, setMessageInput] = useState("");
  const [sendingMessage, setSendingMessage] = useState(false);
  const [docUrl, setDocUrl] = useState("");

  const [isNotFound, setIsNotFound] = useState(false);
  const [isChatOpen, setIsChatOpen] = useState(false);

  const [isFullscreen, setIsFullscreen] = useState(false);

  // State untuk suggested questions
  const [suggestedQuestions, setSuggestedQuestions] = useState<string[]>([]);
  const [loadingSuggestions, setLoadingSuggestions] = useState(false);

  // State untuk export chat
  const [showExportDialog, setShowExportDialog] = useState(false);
  const [exportFormat, setExportFormat] = useState<"pdf" | "markdown">("pdf");
  const [isExporting, setIsExporting] = useState(false);
  const [showPremiumDialog, setShowPremiumDialog] = useState(false);


  const chatEndRef = useRef<HTMLDivElement>(null);
  const previewRef = useRef<HTMLDivElement>(null);

  // Fetch material detail dan chat history
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);
        setIsNotFound(false);

        // Fetch material detail
        const materialData = await getMaterialDetail(material_id);
        console.log("Material Detail API Response:", materialData);
        console.log("File URL:", materialData.file_url);
        console.log("File Type:", materialData.file_type);

        // Validate: pod_id dari URL harus sesuai dengan material.pod_id
        if (materialData.pod_id !== pod_id) {
          console.warn("Pod ID mismatch:", { urlPodId: pod_id, materialPodId: materialData.pod_id });
          setIsNotFound(true);
          setLoading(false);
          return;
        }

        // Fetch user detail untuk validasi username
        const userData = await getUserDetail(materialData.uploader_id);
        console.log("User Detail:", userData);



        setMaterial(materialData);


        // Fetch preview URL
        try {
          const previewUrl = await getMaterialPreviewUrl(material_id);
          console.log("Preview URL:", previewUrl);
          if (previewUrl) {
            setDocUrl(previewUrl);
            console.log("Using preview URL:", previewUrl);
          } else {
            // Fallback: construct URL if preview API doesn't return anything
            const fallbackUrl = materialData.file_url.startsWith("http")
              ? materialData.file_url
              : `http://localhost:9000/${materialData.file_url}`;
            setDocUrl(fallbackUrl);
            console.log("Fallback DocURL:", fallbackUrl);
          }
        } catch (err) {
          console.warn("Failed to fetch preview URL, using fallback:", err);
          // Fallback URL construction
          const fallbackUrl = materialData.file_url.startsWith("http")
            ? materialData.file_url
            : `http://localhost:9000/${materialData.file_url}`;
          setDocUrl(fallbackUrl);
          console.log("Error fallback DocURL:", fallbackUrl);
        }

        // Fetch chat history - optional, don't break if fails
        try {
          const chatHistory = await getMaterialChatHistory(material_id);
          console.log("Chat History:", chatHistory);
          setChatMessages(chatHistory);
        } catch (err) {
          console.warn("Chat history endpoint not available or error:", err);
          // Don't set error state, just continue without chat history
          setChatMessages([]);
        }

        // Fetch suggested questions - optional
        try {
          setLoadingSuggestions(true);
          const suggestions = await getMaterialChatSuggestions(material_id, locale);
          console.log("Suggested Questions Response:", suggestions);

          // Handle different response structures
          let questions: string[] = [];
          if (Array.isArray(suggestions.data?.questions)) {
            // Map objects to strings if needed
            questions = suggestions.data.questions.map((q: any) =>
              typeof q === 'string' ? q : q.question || String(q)
            );
          } else if (Array.isArray(suggestions.data)) {
            // Map objects to strings if needed
            questions = suggestions.data.map((q: any) =>
              typeof q === 'string' ? q : q.question || String(q)
            );
          }

          console.log("Parsed Questions Array:", questions);
          setSuggestedQuestions(questions);
        } catch (err) {
          console.warn("Failed to load suggestions:", err);
          setSuggestedQuestions([]);
        } finally {
          setLoadingSuggestions(false);
        }
      } catch (err) {
        console.error("Error loading material:", err);
        setError(err instanceof Error ? err.message : "Failed to load material");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [material_id, username, pod_id]);

  // Auto-scroll chat messages
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [chatMessages]);

  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (textareaRef.current) {
      // Reset height dulu agar saat teks dihapus, box-nya bisa mengecil lagi
      textareaRef.current.style.height = 'auto';

      // Set height sesuai dengan scrollHeight (tinggi konten asli)
      // Kita batasi maksimalnya (misal 150px)
      const nextHeight = Math.min(textareaRef.current.scrollHeight, 150);
      textareaRef.current.style.height = `${nextHeight}px`;
    }
  }, [messageInput]);

  const toggleFullscreen = () => {
    if (!previewRef.current) return;

    if (!document.fullscreenElement) {
      previewRef.current.requestFullscreen().catch(err => {
        console.error("Error attempting to enable fullscreen mode:", err);
      });
    } else {
      document.exitFullscreen();
    }
  }

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    }
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
    };
  }, []);


  // Handle send chat message
  const handleSendMessage = async () => {
    const messageToSend = messageInput.trim();
    if (!messageToSend || sendingMessage) return;

    try {
      setSendingMessage(true);
      const originalInput = messageInput;
      setMessageInput(""); // Clear input immediately

      // Add user message to chat immediately for better UX
      const tempUserMessage: ChatMessage = {
        id: `temp-user-${Date.now()}`,
        role: "user",
        content: originalInput,
        created_at: new Date().toISOString(),
        session_id: ""
      };
      setChatMessages(prev => [...prev, tempUserMessage]);

      // Send to backend - backend returns assistant response
      const assistantResponse = await sendMaterialChatMessage(material_id, originalInput);

      // Add assistant response (backend should return assistant message)
      setChatMessages(prev => [...prev, assistantResponse]);
    } catch (err) {
      console.error("Error sending message:", err);
      setError("Failed to send message");
      // Restore input on error
      setMessageInput(messageInput);
    } finally {
      setSendingMessage(false);
    }
  };

  // Handle suggested question click
  const handleSuggestedQuestionClick = async (question: string) => {
    if (!question.trim() || sendingMessage) return;

    try {
      setSendingMessage(true);

      // Add user message to chat immediately
      const tempUserMessage: ChatMessage = {
        id: `temp-user-${Date.now()}`,
        role: "user",
        content: question,
        created_at: new Date().toISOString(),
        session_id: ""
      };
      setChatMessages(prev => [...prev, tempUserMessage]);

      // Send to backend - backend returns assistant response
      const assistantResponse = await sendMaterialChatMessage(material_id, question);

      // Add assistant response
      setChatMessages(prev => [...prev, assistantResponse]);
    } catch (err) {
      console.error("Error sending suggested question:", err);
      setError("Failed to send message");
    } finally {
      setSendingMessage(false);
    }
  };

  const handleDownload = async () => {
    try {
      const downloadUrl = await getMaterialDownloadUrl(material_id);
      // console.log("Download URL:", downloadUrl);
      window.open(downloadUrl, "_blank");
      return;
    } catch (err) {
      console.error("Error downloading material:", err);
      setError("Failed to download material");
    }
  };

  const handleExportChat = async () => {
    if (chatMessages.length === 0) {
      setError(t("exportChat.noHistory"));
      return;
    }

    try {
      setIsExporting(true);
      const { exportMaterialChat } = await import("@/lib/api/material");
      const blob = await exportMaterialChat(material_id, exportFormat);

      // Download the file
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `chat-export-${material_id}.${exportFormat === "pdf" ? "pdf" : "md"}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      setShowExportDialog(false);
    } catch (err: any) {
      console.error("Error exporting chat:", err);
      if (err.message === "PREMIUM_REQUIRED") {
        setShowExportDialog(false);
        setShowPremiumDialog(true);
      } else {
        setError("Failed to export chat history");
      }
    } finally {
      setIsExporting(false);
    }
  };

  // Show loading state
  if (loading) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen items-center justify-center">
          <Loader className="animate-spin text-[#FF8811]" size={40} />
        </div>
      </ProtectedRoute>
    );
  }

  // Show 404 state
  if (isNotFound) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen flex-col items-center justify-center gap-4 text-[#2B2D42]">
          <div className="text-6xl font-bold">404</div>
          <div className="text-lg font-bold">Material Not Found</div>
          <div className="text-sm text-gray-500">The material you are looking for does not exist or has been moved.</div>
          <button onClick={() => router.back()} className="px-4 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f]">
            Go Back
          </button>
        </div>
      </ProtectedRoute>
    );
  }

  // Show error state
  if (error || !material) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen flex-col items-center justify-center gap-4 text-[#2B2D42]">
          <div className="text-lg font-bold">Error Loading Material</div>
          <div className="text-sm text-gray-500">{error || "Material not found"}</div>
          <button onClick={() => router.back()} className="px-4 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f]">
            Go Back
          </button>
        </div>
      </ProtectedRoute>
    );
  }
  return (
    <ProtectedRoute>
      <div className="p-4 sm:p-6 lg:p-8 h-screen flex flex-col">
        {/* Header */}
        <div className="flex mb-4 justify-between">
          <div className="flex items-center gap-3">
            <button onClick={() => router.back()} className="hover:text-[#FF8811] transition">
              <ChevronLeft></ChevronLeft>
            </button>
            <h1 className="text-lg sm:text-xl font-bold text-[#2B2D42] truncate">{material.title}</h1>
          </div>
          <button
            onClick={handleDownload}
            className="px-6 py-2 ml-2 max-sm:px-4 max-sm:text-xs bg-white border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5"
          >
            Download
          </button>
        </div>

        {/* Material Information - Horizontal Columns */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 relative">



          {/* Status */}
          <div className="bg-white border-2 border-[#2B2D42] p-4 pb-2 pt-3 shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all">
            <p className="text-xs font-bold text-[#2B2D42]">Status</p>
            <span className={`inline-block px-3 pb-0.5 rounded font-bold text-xs ${material.status === "ready" ? "bg-green-200 text-green-800" : material.status === "processing" ? "bg-blue-200 text-blue-800" : "bg-red-200 text-red-800"}`}>
              {material.status}
            </span>
          </div>


          {/* File Type */}
          <div className="bg-white border-2 border-[#2B2D42] p-4 pb-2 pt-3 shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all">
            <p className="text-xs font-bold text-[#2B2D42] mb-1">File Type</p>
            <p className="text-sm text-gray-600 font-semibold">{material.file_type.toUpperCase()}</p>
          </div>

          {/* Size */}
          <div className="bg-white border-2 border-[#2B2D42] p-4 pb-2 pt-3 shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all">
            <p className="text-xs font-bold text-[#2B2D42] mb-1">Size</p>
            <p className="text-sm text-gray-600 font-semibold">{(material.file_size / 1024 / 1024).toFixed(2)} MB</p>
          </div>

          {/* Rating */}
          <div className="bg-white border-2 border-[#2B2D42] p-4 pb-2 pt-3 shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all">
            <p className="text-xs font-bold text-[#2B2D42] mb-1">Rating</p>
            <p className="text-sm text-gray-600 font-semibold">{material.average_rating.toFixed(1)}/5.0</p>
          </div>
        </div>

        {/* Preview Area - Full Width */}
        <div className="flex-1 bg-white border-2 border-[#2B2D42] shadow-[2px_2px_0px_0px_#2B2D42] rounded-lg overflow-hidden relative">
          <div ref={previewRef} className="w-full h-full relative p-1">
            <button
              onClick={toggleFullscreen}
              className="px-3 py-1 bg-white border-2 rounded-lg border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5"
            >
              {isFullscreen ? <Minimize2 size={20} /> : <Maximize2 size={20} />}
            </button>
          </div>

          {material.file_type.toLowerCase() === "pdf" && <iframe src={docUrl} className="w-full h-full" title="PDF Preview" />}
          {material.file_type.toLowerCase() === "docx" && (
            <div className="flex items-center justify-center h-full bg-gray-50">
              <div className="text-center">
                <div className="text-6xl mb-4">📄</div>
                <p className="text-gray-600 font-semibold text-lg">DOCX Preview</p>
                <p className="text-sm text-gray-500 mt-2">Download to view full document</p>
              </div>
            </div>
          )}
          {material.file_type.toLowerCase() === "pptx" && (
            <div className="flex items-center justify-center h-full bg-gray-50">
              <div className="text-center">
                <div className="text-6xl mb-4">📊</div>
                <p className="text-gray-600 font-semibold text-lg">PPTX Preview</p>
                <p className="text-sm text-gray-500 mt-2">Download to view full presentation</p>
              </div>
            </div>
          )}
          {!["pdf", "docx", "pptx"].includes(material.file_type.toLowerCase()) && (
            <div className="flex items-center justify-center h-full bg-gray-50">
              <div className="text-center">
                <div className="text-6xl mb-4">📁</div>
                <p className="text-gray-600 font-semibold text-lg">Preview Unavailable</p>
                <p className="text-sm text-gray-500 mt-2">File type: {material.file_type.toUpperCase()}</p>
              </div>
            </div>
          )}

        </div>
        {/* Floating Chat Widget Button */}
        {!isChatOpen && (
          <button
            onClick={() => setIsChatOpen(true)}
            className="fixed bottom-8 right-8 w-14 h-14 bg-[#FF8811] text-white rounded-full shadow-lg hover:bg-[#e67a0f] transition-all flex items-center justify-center hover:scale-110"
            title={t("chatWidget.openChat")}
          >
            <MessageCircle size={24} />
          </button>
        )}

        {/* Floating Chat Widget */}
        {isChatOpen && (
          <div className="fixed bottom-8 right-8 w-3/10 h-3/4 max-sm:w-4/5 max-sm: bg-white border-2 border-[#2B2D42] rounded-lg shadow-xl flex flex-col overflow-hidden">
            {/* Chat Header */}
            <div className="bg-[#FF8811] text-white px-4 py-3 flex items-center justify-between">
              <h3 className="font-bold text-sm">{t("chatWidget.title")}</h3>
              <div className="flex items-center gap-2">
                {chatMessages.length > 0 && (
                  <button
                    onClick={() => setShowExportDialog(true)}
                    className="hover:bg-[#e67a0f] p-1 rounded transition"
                    title={t("chatWidget.exportChat")}
                  >
                    <DownloadIcon size={16} />
                  </button>
                )}
                <button onClick={() => setIsChatOpen(false)} className="hover:bg-[#e67a0f] p-1 rounded transition">
                  <X size={18} />
                </button>
              </div>
            </div>

            {/* Chat Messages */}
            <div data-lenis-prevent className="flex-1 overflow-y-auto space-y-3 p-4 pr-2 bg-gray-50">
              {chatMessages.length === 0 ? (
                <div className="text-center text-gray-500 text-xs py-6">
                  <div className="flex items-center justify-center gap-2 mb-2">
                    <Sparkles size={16} className="text-[#FF8811]" />
                    <span className="font-semibold text-gray-700">{t("chatWidget.startConversation")}</span>
                  </div>
                  <p className="text-gray-500">{t("chatWidget.askQuestion")}</p>
                </div>
              ) : (
                <>
                  {chatMessages
                    .map((msg, idx) => {
                      if (idx === 0 && !msg.content.trim()) {
                        return null;
                      }
                      return (
                        <div key={msg.id} className={`flex ${msg.role === "user" ? "justify-end wrap-break-word" : "gap-2"}`}>
                          {msg.role === "assistant" && <div className="w-6 h-6 rounded-full bg-[#FF8811] shrink-0 flex items-center justify-center text-[9px] text-white font-bold">AI</div>}
                          <FormattedMessage content={msg.content} role={msg.role} />
                        </div>
                      );
                    })
                    .filter(Boolean)}
                </>
              )}
              <div ref={chatEndRef} />
            </div>

            {/* Suggested Questions - Above Chat Input */}
            {suggestedQuestions.length > 0 && (
              <div className="border-t border-gray-200 px-3 pt-3 pb-2 bg-gradient-to-b from-white to-gray-50 max-h-48 overflow-y-auto">
                {loadingSuggestions ? (
                  <div className="flex justify-center py-2">
                    <Loader size={14} className="animate-spin text-[#FF8811]" />
                  </div>
                ) : (
                  <div className="space-y-2">
                    <p className="text-[11px] font-bold text-gray-700 mb-2 flex items-center gap-1.5">
                      <Sparkles size={13} className="text-[#FF8811]" />
                      {t("chatWidget.suggestedQuestions")}
                    </p>
                    {suggestedQuestions.map((question, idx) => (
                      <button
                        key={idx}
                        onClick={() => handleSuggestedQuestionClick(question)}
                        disabled={sendingMessage}
                        className="w-full text-left text-[11px] px-3 py-2 bg-white border border-gray-300 rounded-lg hover:border-[#FF8811] hover:bg-orange-50 hover:shadow-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed group"
                      >
                        <span className="text-gray-700 line-clamp-2 group-hover:text-[#FF8811]">{question}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Chat Input */}
            <div className="border-t-2 border-gray-200 p-3">
              <div className="flex gap-2">
                <textarea
                  ref={textareaRef}
                  data-lenis-prevent
                  className="flex-1 min-h-10 overflow-y-auto resize-none text-xs p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#FF8811]"
                  style={{ height: '20px' }}
                  placeholder={t("chatWidget.placeholder")}
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && e.ctrlKey) {
                      handleSendMessage();
                    }
                  }}
                ></textarea>
                <button onClick={() => handleSendMessage()} disabled={sendingMessage || !messageInput.trim()} className="px-3 h-10 py-2 bg-[#FF8811] text-white rounded-lg hover:bg-[#e67a0f] transition disabled:opacity-50">
                  {sendingMessage ? <Loader size={14} className="animate-spin" /> : <Send size={14} />}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Export Chat Dialog - Standard Style */}
        <Dialog open={showExportDialog} onOpenChange={setShowExportDialog}>
          <DialogContent className="sm:max-w-md border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-xl font-black">
                <FileText size={20} className="text-black" />
                {t("exportChat.title")}
              </DialogTitle>
            </DialogHeader>

            <div className="space-y-4">
              <p className="text-sm text-gray-600">
                {t("exportChat.description")}
              </p>

              {/* Format Selection */}
              <div className="space-y-3">
                <label className="text-sm font-bold text-black">{t("exportChat.formatLabel")}</label>
                <div className="space-y-2">
                  <button
                    onClick={() => setExportFormat("pdf")}
                    className={`w-full text-left px-4 py-3 border-2 border-black rounded-lg font-semibold transition-all ${exportFormat === "pdf"
                      ? "bg-[#FF8811] text-white shadow-[2px_2px_0_0_black]"
                      : "bg-white text-black hover:bg-gray-50"
                      }`}
                  >
                    <div className="flex items-center justify-between">
                      <span>{t("exportChat.pdfFormat")}</span>
                      <span className="text-xs opacity-75">{t("exportChat.recommended")}</span>
                    </div>
                  </button>
                  <button
                    onClick={() => setExportFormat("markdown")}
                    className={`w-full text-left px-4 py-3 border-2 border-black rounded-lg font-semibold transition-all ${exportFormat === "markdown"
                      ? "bg-[#FF8811] text-white shadow-[2px_2px_0_0_black]"
                      : "bg-white text-black hover:bg-gray-50"
                      }`}
                  >
                    <div className="flex items-center justify-between">
                      <span>{t("exportChat.markdownFormat")}</span>
                      <span className="text-xs opacity-75">{t("exportChat.plainText")}</span>
                    </div>
                  </button>
                </div>
              </div>
            </div>

            <DialogFooter className="gap-2 sm:justify-end">
              <button
                onClick={() => setShowExportDialog(false)}
                disabled={isExporting}
                className="px-4 py-2 border-2 border-black bg-white text-black font-bold rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
              >
                {tCommon("cancel")}
              </button>
              <button
                onClick={handleExportChat}
                disabled={isExporting}
                className="px-4 py-2 border-2 border-black bg-[#FF8811] text-white font-bold rounded-lg shadow-[2px_2px_0_0_black] hover:shadow-[1px_1px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all disabled:opacity-50 flex items-center gap-2"
              >
                {isExporting ? (
                  <>
                    <Loader size={16} className="animate-spin" />
                    {t("exportChat.exporting")}
                  </>
                ) : (
                  <>
                    <DownloadIcon size={16} />
                    {t("exportChat.export")}
                  </>
                )}
              </button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Premium Required Dialog - Standard Style */}
        <Dialog open={showPremiumDialog} onOpenChange={setShowPremiumDialog}>
          <DialogContent className="sm:max-w-md border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-xl font-black">
                <span className="text-2xl">⭐</span>
                {t("premiumFeature.title")}
              </DialogTitle>
            </DialogHeader>

            <div className="space-y-4">
              <p className="text-gray-700 font-semibold">
                {t("premiumFeature.chatExportPremium")}
              </p>
              <p className="text-sm text-gray-600">
                {t("premiumFeature.upgradeDescription")}
              </p>

              {/* Premium Benefits */}
              <div className="bg-white border-2 border-black rounded-lg p-4 space-y-2">
                <p className="text-xs font-bold text-black uppercase">
                  {t("premiumFeature.benefitsTitle")}
                </p>
                <ul className="text-sm text-gray-700 space-y-1">
                  <li className="flex items-center gap-2">
                    <span className="text-[#FF8811]">✓</span>
                    {t("premiumFeature.benefit1")}
                  </li>
                  <li className="flex items-center gap-2">
                    <span className="text-[#FF8811]">✓</span>
                    {t("premiumFeature.benefit2")}
                  </li>
                  <li className="flex items-center gap-2">
                    <span className="text-[#FF8811]">✓</span>
                    {t("premiumFeature.benefit3")}
                  </li>
                </ul>
              </div>
            </div>

            <DialogFooter className="gap-2 sm:justify-end">
              <button
                onClick={() => setShowPremiumDialog(false)}
                className="px-4 py-2 border-2 border-black bg-white text-black font-bold rounded-lg hover:bg-gray-50 transition-colors"
              >
                {t("premiumFeature.maybeLater")}
              </button>
              <button
                onClick={() => {
                  setShowPremiumDialog(false);
                  router.push("/dashboard/upgrade");
                }}
                className="px-4 py-2 border-2 border-black bg-[#FF8811] text-white font-bold rounded-lg shadow-[2px_2px_0_0_black] hover:shadow-[1px_1px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all"
              >
                {t("premiumFeature.upgradeNow")}
              </button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </ProtectedRoute>
  );
}
