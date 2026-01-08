"use client";

import React, { useState, useEffect, useRef } from "react";
import { ProtectedRoute } from "@/components/auth";
import { useRouter } from "next/navigation";
import { Bell, ChevronLeft, Download, Loader, Plus, Search, Send, MessageCircle, X, Minimize2, Maximize2 } from "lucide-react";
import { getMaterialDetail, getMaterialChatHistory, sendMaterialChatMessage, getMaterialPreviewUrl, getMaterialDownloadUrl } from "@/lib/api/material";
import { getUserDetail } from "@/lib/api/user";
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

  const chatEndRef = useRef<HTMLDivElement>(null);
  const previewRef = useRef<HTMLDivElement>(null);

  const [isFullscreen, setIsFullscreen] = useState(false);

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
        }

        // Fetch user detail untuk validasi username
        const userData = await getUserDetail(materialData.uploader_id);
        console.log("User Detail:", userData);

        // Validate: username dari URL harus sesuai dengan user.username
        // if (userData.username !== username) {
        //   console.warn("Username mismatch:", { urlUsername: username, userUsername: userData.name });
        //   setIsNotFound(true);
        //   setLoading(false);
        //   return;
        // }

        setMaterial(materialData);

        // Fetch preview URL
        try {
          const previewUrl = await getMaterialPreviewUrl(material_id);
          console.log("Preview URL:", previewUrl);
          if (previewUrl) {
            setDocUrl(previewUrl);
          } else {
            // Fallback: construct URL if preview API doesn't return anything
            if (materialData.file_url.startsWith("http")) {
              setDocUrl(materialData.file_url);
            } else {
              setDocUrl("http://localhost:9000/" + materialData.file_url);
            }
          }

        } catch (err) {
          console.warn("Failed to fetch preview URL, using fallback:", err);
          // Fallback URL construction
          if (materialData.file_url.startsWith("http")) {
            setDocUrl(materialData.file_url);
          } else {
            setDocUrl("http://localhost:9000/" + materialData.file_url);
          }
        }
        console.log("Constructed DocURL:", docUrl);

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
      router.push(downloadUrl);
      return;
    } catch (err) {
      console.error("Error downloading material:", err);
      setError("Failed to download material");
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
              <button onClick={() => setIsChatOpen(false)} className="hover:bg-[#e67a0f] p-1 rounded transition">
                <X size={18} />
              </button>
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
                      <div key={msg.id} className={`flex ${msg.role === "user" ? "justify-end wrap-break-word" : "gap-2"}`}>
                        {msg.role === "assistant" && <div className="w-6 h-6 rounded-full bg-[#FF8811] shrink-0 flex items-center justify-center text-[9px] text-white font-bold">AI</div>}
                        <FormattedMessage content={msg.content} role={msg.role} />
                      </div>
                    );
                  })
                  .filter(Boolean)
              )}
              <div ref={chatEndRef} />
            </div>

            {/* Chat Input */}
            <div className="border-t-2 border-gray-200 p-3">
              <div className="flex gap-2">
                <textarea
                  ref={textareaRef}
                  data-lenis-prevent
                  className="flex-1 min-h-10 overflow-y-auto resize-none text-xs p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#FF8811]"
                  style={{ height: '20px' }}
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
      </div>
    </ProtectedRoute>
  );
}
