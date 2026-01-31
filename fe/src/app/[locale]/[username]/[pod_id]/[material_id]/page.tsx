"use client";

import React, { useState, useEffect, useRef } from "react";
import { ProtectedRoute } from "@/components/auth";
import { useRouter } from "next/navigation";
import { Bell, ChevronLeft, Download, Loader, Plus, Search, Send, MessageCircle, X, Minimize2, Maximize2, ThumbsUp, ThumbsDown, FileDown, Sparkles, Lock } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { getMaterialDetail, getMaterialChatHistory, sendMaterialChatMessage, getMaterialPreviewUrl, getMaterialDownloadUrl } from "@/lib/api/material";
import { getUserDetail } from "@/lib/api/user";
import { Material, ChatMessage } from "@/types/material";
import { FormattedMessage } from "@/components/FormattedMessage";
import { useAIFeatures } from "@/hooks/useAIFeatures";
import VersionHistoryDialog from "@/components/knowledge-pod/VersionHistoryDialog";
import { useDownloads } from "@/hooks/useDownloads";
import { useOfflineMaterials } from "@/hooks/useOffline";

interface PageProps {
  params: Promise<{
    username: string;
    pod_id: string;
    material_id: string;
  }>;
}

export default function MaterialDetailPage({ params }: PageProps) {
  const router = useRouter();
  const { username, pod_id, material_id } = React.use(params);
  const { addMaterial, isDownloaded } = useDownloads();
  const { 
    downloadForOffline, 
    isDownloaded: isOfflineDownloaded, 
    isRegistered,
    loading: offlineLoading 
  } = useOfflineMaterials();

  // State untuk offline download
  const [savingOffline, setSavingOffline] = useState(false);

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
  const [isPremiumModalOpen, setIsPremiumModalOpen] = useState(false);

  // AI Features integration
  const {
    suggestions,
    suggestionsLoading,
    loadSuggestions,
    exportLoading,
    handleExport,
    feedbackLoading,
    handleFeedback,
  } = useAIFeatures({ materialId: material_id });

  const [showSuggestions, setShowSuggestions] = useState(false);
  const [messageFeedback, setMessageFeedback] = useState<Record<string, "thumbs_up" | "thumbs_down">>({});

  // Debug: Monitor suggestions state
  useEffect(() => {
    console.log("[Page] Suggestions state changed:", suggestions);
    console.log("[Page] Suggestions length:", suggestions.length);
    console.log("[Page] Show suggestions:", showSuggestions);
  }, [suggestions, showSuggestions]);

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
          console.log("Preview URL from API:", previewUrl);
          if (previewUrl) {
            setDocUrl(previewUrl);
            console.log("Set DocURL to preview URL:", previewUrl);
          } else {
            // Fallback: construct URL if preview API doesn't return anything
            let fallbackUrl = "";
            if (materialData.file_url.startsWith("http")) {
              fallbackUrl = materialData.file_url;
            } else {
              fallbackUrl = "http://localhost:9000/" + materialData.file_url;
            }
            setDocUrl(fallbackUrl);
            console.log("Set DocURL to fallback:", fallbackUrl);
          }
        } catch (err) {
          console.warn("Failed to fetch preview URL, using fallback:", err);
          // Fallback URL construction
          let fallbackUrl = "";
          if (materialData.file_url.startsWith("http")) {
            fallbackUrl = materialData.file_url;
          } else {
            fallbackUrl = "http://localhost:9000/" + materialData.file_url;
          }
          setDocUrl(fallbackUrl);
          console.log("Set DocURL to error fallback:", fallbackUrl);
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
      textareaRef.current.style.height = "auto";

      // Set height sesuai dengan scrollHeight (tinggi konten asli)
      // Kita batasi maksimalnya (misal 150px)
      const nextHeight = Math.min(textareaRef.current.scrollHeight, 150);
      textareaRef.current.style.height = `${nextHeight}px`;
    }
  }, [messageInput]);

  const toggleFullscreen = () => {
    if (!previewRef.current) return;

    if (!document.fullscreenElement) {
      previewRef.current.requestFullscreen().catch((err) => {
        console.error("Error attempting to enable fullscreen mode:", err);
      });
    } else {
      document.exitFullscreen();
    }
  };

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
    };
  }, []);

  // Handle send chat message
  const handleSendMessage = async () => {
    if (!messageInput.trim()) return;

    try {
      setSendingMessage(true);
      const newMessage = await sendMaterialChatMessage(material_id, messageInput);
      setChatMessages([...chatMessages, newMessage]);
      setMessageInput("");
    } catch (err) {
      console.error("Error sending message:", err);
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

  // Handle save for offline
  const handleSaveOffline = async () => {
    if (!material) return;
    
    setSavingOffline(true);
    try {
      await downloadForOffline(material.id, material.title);
    } catch (err) {
      console.error("Error saving for offline:", err);
      setError("Failed to save for offline");
    } finally {
      setSavingOffline(false);
    }
  };

  // Handle suggestion click
  const handleSuggestionClick = (question: string) => {
    setMessageInput(question);
    setShowSuggestions(false);
  };

  // Handle export chat
  const handleExportChat = async (format: "pdf" | "markdown") => {
    // Show premium modal instead of exporting
    setIsPremiumModalOpen(true);
    /* 
    try {
      await handleExport(format);
    } catch (err) {
      console.error("Export failed:", err);
      setError("Failed to export chat");
    } 
    */
  };

  // Handle message feedback
  const handleMessageFeedback = async (
    messageId: string,
    feedback: "thumbs_up" | "thumbs_down"
  ) => {
    try {
      await handleFeedback(messageId, feedback);
      setMessageFeedback((prev) => ({ ...prev, [messageId]: feedback }));
    } catch (err) {
      console.error("Feedback submission failed:", err);
      setError("Failed to submit feedback");
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
          <div className="flex items-center gap-2">
            {/* Save Offline Button */}
            <button
              onClick={handleSaveOffline}
              disabled={savingOffline || isOfflineDownloaded(material.id) || !isRegistered}
              className={`px-4 py-2 max-sm:px-3 max-sm:text-xs bg-white border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] transition-all shadow-[2px_2px_0px_0px_#2B2D42] ${
                savingOffline || isOfflineDownloaded(material.id) || !isRegistered
                  ? "opacity-50 cursor-not-allowed"
                  : "hover:bg-[#FF8811] hover:text-white hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5"
              }`}
              title={!isRegistered ? "Register device first in Downloads page" : "Save for offline access"}
            >
              {savingOffline ? (
                <Loader size={16} className="animate-spin" />
              ) : isOfflineDownloaded(material.id) ? (
                "Saved Offline"
              ) : (
                <Download size={16} />
              )}
            </button>
            <button
              onClick={handleDownload}
              disabled={material && isDownloaded(material.id)}
              className={`px-6 py-2 max-sm:px-4 max-sm:text-xs bg-white border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] transition-all shadow-[2px_2px_0px_0px_#2B2D42] ${material && isDownloaded(material.id)
                ? "opacity-50 cursor-not-allowed"
                : "hover:bg-[#FF8811] hover:text-white hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5"
                }`}
            >
              {material && isDownloaded(material.id) ? "Saved" : "Download"}
            </button>
            <VersionHistoryDialog
              materialId={material.id}
              currentVersion={material.current_version}
              onRestore={() => {
                getMaterialDetail(material_id).then(setMaterial);
              }}
            />
          </div>
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
          <div ref={previewRef} className="absolute p-1">
            <button
              onClick={toggleFullscreen}
              className="px-3 py-1 bg-white border-2 rounded-lg border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5"
            >
              {isFullscreen ? <Minimize2 size={20} /> : <Maximize2 size={20} />}
            </button>
          </div>

          {material.file_type.toLowerCase() === "pdf" &&
            (docUrl ? (
              <iframe key={docUrl} src={docUrl} className="w-full bg-yellow-500 h-full" title="PDF Preview" />
            ) : (
              <div className="flex items-center justify-center h-full bg-gray-50">
                <div className="text-center">
                  <Loader className="animate-spin text-[#FF8811] mx-auto mb-4" size={40} />
                  <p className="text-gray-600 font-semibold">Loading PDF preview...</p>
                </div>
              </div>
            ))}
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
            title="Open Chat"
          >
            <MessageCircle size={24} />
          </button>
        )}

        {/* Floating Chat Widget */}
        {isChatOpen && (
          <div className="fixed bottom-8 right-8 w-3/10 h-3/4 max-sm:w-4/5 max-sm: bg-white border-2 border-[#2B2D42] rounded-lg shadow-xl flex flex-col overflow-hidden">
            {/* Chat Header */}
            <div className="bg-[#FF8811] text-white px-4 py-3 flex items-center justify-between">
              <h3 className="font-bold text-sm">Chat with AI</h3>
              <div className="flex items-center gap-2">
                {/* Export Dropdown */}
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      className="hover:bg-[#e67a0f] p-1 rounded transition disabled:opacity-50"
                      title="Export Chat"
                      disabled={exportLoading}
                    >
                      {exportLoading ? (
                        <Loader size={18} className="animate-spin" />
                      ) : (
                        <FileDown size={18} />
                      )}
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="bg-white border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] rounded-none p-0 min-w-[160px]">
                    <DropdownMenuItem
                      onClick={() => handleExportChat("pdf")}
                      className="px-4 py-3 text-xs font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white rounded-none cursor-pointer focus:bg-[#FF8811] focus:text-white"
                      disabled={exportLoading}
                    >
                      Export as PDF
                    </DropdownMenuItem>
                    <div className="h-[2px] bg-[#2B2D42]"></div>
                    <DropdownMenuItem
                      onClick={() => handleExportChat("markdown")}
                      className="px-4 py-3 text-xs font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white rounded-none cursor-pointer focus:bg-[#FF8811] focus:text-white"
                      disabled={exportLoading}
                    >
                      Export as Markdown
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <button onClick={() => setIsChatOpen(false)} className="hover:bg-[#e67a0f] p-1 rounded transition">
                  <X size={18} />
                </button>
              </div>
            </div>

            {/* Chat Messages */}
            <div data-lenis-prevent className="flex-1 overflow-y-auto space-y-3 p-4 pr-2 bg-gray-50">
              {chatMessages.length === 0 ? (
                <div className="text-center text-gray-500 text-xs py-6">No messages yet. Start a conversation!</div>
              ) : (
                chatMessages
                  .map((msg, idx) => {
                    if (idx === 0 && !msg.content.trim()) {
                      return null;
                    }
                    return (
                      <div key={msg.id} className={`flex flex-col ${msg.role === "user" ? "items-end" : "gap-2"}`}>
                        <div className={`flex ${msg.role === "user" ? "justify-end wrap-break-word" : "gap-2"}`}>
                          {msg.role === "assistant" && <div className="w-6 h-6 rounded-full bg-[#FF8811] shrink-0 flex items-center justify-center text-[9px] text-white font-bold">AI</div>}
                          <FormattedMessage content={msg.content} role={msg.role} />
                        </div>
                        {/* Feedback buttons for assistant messages */}
                        {msg.role === "assistant" && (
                          <div className="flex gap-1 ml-8 mt-1">
                            <button
                              onClick={() => handleMessageFeedback(msg.id, "thumbs_up")}
                              disabled={feedbackLoading[msg.id] || !!messageFeedback[msg.id]}
                              className={`p-1 rounded transition ${messageFeedback[msg.id] === "thumbs_up"
                                ? "bg-green-100 text-green-600"
                                : "hover:bg-gray-200 text-gray-500"
                                } disabled:opacity-50`}
                              title="Good response"
                            >
                              <ThumbsUp size={12} />
                            </button>
                            <button
                              onClick={() => handleMessageFeedback(msg.id, "thumbs_down")}
                              disabled={feedbackLoading[msg.id] || !!messageFeedback[msg.id]}
                              className={`p-1 rounded transition ${messageFeedback[msg.id] === "thumbs_down"
                                ? "bg-red-100 text-red-600"
                                : "hover:bg-gray-200 text-gray-500"
                                } disabled:opacity-50`}
                              title="Bad response"
                            >
                              <ThumbsDown size={12} />
                            </button>
                          </div>
                        )}
                      </div>
                    );
                  })
                  .filter(Boolean)
              )}
              <div ref={chatEndRef} />
            </div>

            {/* Chat Input */}
            <div className="border-t-2 border-gray-200 p-3">
              {/* Suggestions Panel - Above Input */}
              {showSuggestions && (
                <div className="mb-3 bg-gray-50 border border-gray-300 rounded-lg p-2 max-h-32 overflow-y-auto">
                  <div className="text-xs font-bold text-[#2B2D42] mb-2">Suggested Questions:</div>
                  {suggestionsLoading ? (
                    <div className="text-center py-2">
                      <Loader size={14} className="animate-spin inline-block" />
                      <span className="ml-2 text-xs text-gray-500">Loading suggestions...</span>
                    </div>
                  ) : suggestions.length > 0 ? (
                    <div className="space-y-1">
                      {suggestions.map((question, idx) => (
                        <button
                          key={idx}
                          onClick={() => handleSuggestionClick(question)}
                          className="block w-full text-left text-xs p-2 bg-white border border-gray-300 rounded hover:bg-[#FF8811] hover:text-white transition"
                        >
                          {question}
                        </button>
                      ))}
                    </div>
                  ) : (
                    <div className="text-xs text-gray-500 text-center py-2">
                      No suggestions available. Try asking a question!
                    </div>
                  )}
                </div>
              )}

              <div className="flex gap-2">
                <button
                  onClick={() => {
                    console.log("[Page] Suggestions button clicked");
                    console.log("[Page] Current showSuggestions:", showSuggestions);
                    console.log("[Page] Current suggestions:", suggestions);
                    console.log("[Page] Suggestions length:", suggestions.length);
                    if (!showSuggestions && suggestions.length === 0) {
                      console.log("[Page] Calling loadSuggestions...");
                      loadSuggestions();
                    }
                    setShowSuggestions(!showSuggestions);
                    console.log("[Page] Toggled showSuggestions to:", !showSuggestions);
                  }}
                  disabled={suggestionsLoading}
                  className="px-2 h-10 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-100 transition disabled:opacity-50"
                  title="Get suggestions"
                >
                  {suggestionsLoading ? (
                    <Loader size={14} className="animate-spin" />
                  ) : (
                    <Sparkles size={14} className={showSuggestions ? "text-[#FF8811]" : ""} />
                  )}
                </button>
                <textarea
                  ref={textareaRef}
                  data-lenis-prevent
                  className="flex-1 min-h-10 overflow-y-auto resize-none text-xs p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#FF8811]"
                  style={{ height: "20px" }}
                  placeholder="Type your question..."
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && e.ctrlKey) {
                      handleSendMessage();
                    }
                  }}
                ></textarea>
                <button onClick={handleSendMessage} disabled={sendingMessage || !messageInput.trim()} className="px-3 h-10 py-2 bg-[#FF8811] text-white rounded-lg hover:bg-[#e67a0f] transition disabled:opacity-50">
                  {sendingMessage ? <Loader size={14} className="animate-spin" /> : <Send size={14} />}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Premium Modal */}
        <Dialog open={isPremiumModalOpen} onOpenChange={setIsPremiumModalOpen}>
          <DialogContent className="bg-white border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] rounded-none sm:max-w-md p-0 overflow-hidden gap-0">
            <div className="bg-[#FF8811] p-4 flex items-center justify-between border-b-2 border-[#2B2D42]">
              <div className="flex items-center gap-3">
                <div className="bg-white p-1 border-2 border-[#2B2D42] shadow-[2px_2px_0px_0px_#2B2D42]">
                  <Lock className="w-5 h-5 text-[#2B2D42]" />
                </div>
                <DialogTitle className="text-xl font-bold text-white uppercase tracking-wider">LOCKED</DialogTitle>
              </div>
              <button onClick={() => setIsPremiumModalOpen(false)} className="text-white hover:text-[#2B2D42] transition-colors">
                <X size={24} />
              </button>
            </div>

            <div className="p-8 bg-white">
              <div className="mb-6 bg-yellow-50 border-2 border-[#2B2D42] p-4 shadow-[2px_2px_0px_0px_#2B2D42]">
                <h4 className="text-lg font-bold text-[#2B2D42] mb-1">PREMIUM ONLY!</h4>
                <DialogDescription className="text-sm font-medium text-[#2B2D42] opacity-100">
                  You need to upgrade to Premium to export your chat history. Don&apos;t miss out!
                </DialogDescription>
              </div>

              <DialogFooter className="sm:justify-center">
                <button
                  onClick={() => setIsPremiumModalOpen(false)}
                  className="w-full px-6 py-3 bg-[#2B2D42] text-white font-bold text-lg uppercase border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#FF8811] hover:shadow-none hover:translate-x-1 hover:translate-y-1 transition-all"
                >
                  I Understand
                </button>
              </DialogFooter>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </ProtectedRoute>
  );
}
