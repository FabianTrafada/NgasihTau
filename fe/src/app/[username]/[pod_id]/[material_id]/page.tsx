"use client";

import React, { useState, useEffect } from "react";
import { ProtectedRoute } from "@/components/auth";
import { useRouter } from "next/navigation";
import { Bell, ChevronLeft, Download, Loader, Plus, Search, Send } from "lucide-react";
import { getMaterialDetail, getMaterialChatHistory, sendMaterialChatMessage, getMaterialPreviewUrl, } from "@/lib/api/material";
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
      <div className="p-4 sm:p-6 lg:p-8">
        <div className="flex flex-col gap-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <h1 className="text-2xl sm:text-3xl font-bold text-[#2B2D42]">{material.title}</h1>
            <button onClick={() => router.back()} className="px-6 py-2 border-2 border-[#2B2D42] rounded-lg font-bold text-[#2B2D42] hover:bg-[#2B2D42] hover:text-white transition">
              Back
            </button>
          </div>

          {/* Content Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
            {/* Material Viewer - Left Side */}
            <div className="lg:col-span-3 space-y-4">
              {/* Material Info Card */}
              <div className="bg-white border-2 border-[#2B2D42] p-6 shadow-[2px_2px_0px_0px_#2B2D42]">
                <h2 className="text-lg font-bold text-[#2B2D42] mb-4">Material Information</h2>

                <div className="space-y-3 text-xs">
                  <div className="flex items-center justify-between pb-3 border-b-2 border-gray-100">
                    <span className="font-bold text-[#2B2D42]">Status:</span>
                    <span className={`px-3 py-1 rounded font-bold ${material.status === "ready" ? "bg-green-200 text-green-800" : material.status === "processing" ? "bg-blue-200 text-blue-800" : "bg-red-200 text-red-800"}`}>
                      {material.status}
                    </span>
                  </div>

                  <div className="flex items-center justify-between pb-3 border-b-2 border-gray-100">
                    <span className="font-bold text-[#2B2D42]">File Type:</span>
                    <span className="text-gray-600">{material.file_type.toUpperCase()}</span>
                  </div>

                  <div className="flex items-center justify-between pb-3 border-b-2 border-gray-100">
                    <span className="font-bold text-[#2B2D42]">Size:</span>
                    <span className="text-gray-600">{(material.file_size / 1024 / 1024).toFixed(2)} MB</span>
                  </div>

                  <div className="flex items-center justify-between">
                    <span className="font-bold text-[#2B2D42]">Stats:</span>
                    <span className="text-xs text-gray-500">
                      Views: {material.view_count} | Downloads: {material.download_count} | Rating: {material.average_rating.toFixed(1)}/5.0
                    </span>
                  </div>
                </div>
              </div>

              {/* Download Button */}
              <button className="w-full px-4 py-2 bg-white border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-[2px] hover:translate-y-[2px]">
                Download
              </button>

              {/* Debug Section - Temporary
              <div className="bg-yellow-100 border-2 border-yellow-400 p-4 rounded-lg text-xs">
                <p className="font-bold text-yellow-800 mb-2">Debug Info:</p>
                <p className="text-yellow-700 break-all">
                  <strong>File URL:</strong> {material.file_url}
                </p>
                <p className="text-yellow-700 mt-2">
                  <strong>File Type:</strong> {material.file_type}
                </p>
              </div> */}

              {/* Preview Card */}
              <div className="bg-white border-2 border-[#2B2D42] p-6 shadow-[2px_2px_0px_0px_#2B2D42]">
                <h3 className="text-lg font-bold text-[#2B2D42] mb-4">Preview</h3>
                <div className="w-full h-[500px] bg-gray-100 rounded-lg overflow-hidden border border-gray-300">
                  {material.file_type.toLowerCase() === "pdf" && <iframe src={docUrl} className="w-full h-full" title="PDF Preview" />}
                  {material.file_type.toLowerCase() === "docx" && (
                    <div className="flex items-center justify-center h-full bg-gray-50">
                      <div className="text-center">
                        <div className="text-4xl mb-2">📄</div>
                        <p className="text-gray-600 font-semibold">DOCX Preview</p>
                        <p className="text-xs text-gray-500 mt-2">Download to view full document</p>
                      </div>
                    </div>
                  )}
                  {material.file_type.toLowerCase() === "pptx" && (
                    <div className="flex items-center justify-center h-full bg-gray-50">
                      <div className="text-center">
                        <div className="text-4xl mb-2">📊</div>
                        <p className="text-gray-600 font-semibold">PPTX Preview</p>
                        <p className="text-xs text-gray-500 mt-2">Download to view full presentation</p>
                      </div>
                    </div>
                  )}
                  {!["pdf", "docx", "pptx"].includes(material.file_type.toLowerCase()) && (
                    <div className="flex items-center justify-center h-full bg-gray-50">
                      <div className="text-center">
                        <div className="text-4xl mb-2">📁</div>
                        <p className="text-gray-600 font-semibold">Preview Unavailable</p>
                        <p className="text-xs text-gray-500 mt-2">File type: {material.file_type.toUpperCase()}</p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* AI Chat Panel - Right Side */}
            <div className="col-span-2 bg-white border-2 border-[#2B2D42] p-6 shadow-[2px_2px_0px_0px_#2B2D42] flex flex-col max-h-[600px]">
              <h3 className="font-bold text-lg text-[#2B2D42] mb-4">Chatbot</h3>

              {/* Chat Messages */}
              <div className="flex-1 overflow-y-auto space-y-3 mb-4 pr-2">
                {chatMessages.length === 0 ? (
                  <div className="text-center text-gray-500 text-xs py-6">No messages yet. Start a conversation!</div>
                ) : (
                  chatMessages
                    .map((msg, idx) => {
                      // Skip first message (index 0) if it's empty
                      if (idx === 0 && !msg.content.trim()) {
                        return null;
                      }
                      return (
                        <div key={msg.id} className={`flex ${msg.role === "user" ? "justify-end" : "gap-2"}`}>
                          {msg.role === "assistant" && <div className="w-7 h-7 rounded-full bg-[#FF8811] shrink-0 flex items-center justify-center text-[10px] text-white font-bold">AI</div>}
                          <FormattedMessage content={msg.content} role={msg.role} />
                        </div>
                      );
                    })
                    .filter(Boolean) // Filter out null values
                )}
              </div>

              {/* Input Area */}
              <div className="border-t-2 border-gray-200 mt-3 border-2 border-[#2B2D42] rounded-lg">
                <textarea
                  className="w-full h-12 resize-none text-xs p-2 rounded-lg focus:outline-none"
                  placeholder="Type your question..."
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && e.ctrlKey) {
                      handleSendMessage();
                    }
                  }}
                ></textarea>
                <button onClick={handleSendMessage} disabled={sendingMessage || !messageInput.trim()} className="flex justify-end w-full px-4 py-2 text-sm font-bold text-[#2B2D42] transition-all">
                  {sendingMessage ? (
                    <>
                      <Loader size={14} className="animate-spin" />
                    </>
                  ) : (
                    <Send size={16} className="rotate-45 fill-[#2B2D42] hover:stroke-[#FF8811] hover:fill-[#FF8811]" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
