import React, { useState, useEffect } from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { MaterialVersion } from '@/types/version';
import { getMaterialVersions, restoreMaterialVersion, getMaterialDownloadUrl } from '@/lib/api/material';
import { History, Upload, User, FileText, Loader2, RotateCcw, Download } from 'lucide-react';
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

interface VersionHistoryDialogProps {
  materialId: string;
  currentVersion?: string | number; // This might be version number (1, 2, 3) or string ("v1")
  onRestore?: () => void;
}

const VersionHistoryDialog: React.FC<VersionHistoryDialogProps> = ({ materialId, currentVersion, onRestore }) => {
  const [versions, setVersions] = useState<MaterialVersion[]>([]);
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  // Upload State
  const [isUploading, setIsUploading] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [commitMessage, setCommitMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Restore/Download state
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchVersions = async () => {
    setLoading(true);
    try {
      const data = await getMaterialVersions(materialId);
      setVersions(data);
    } catch (error) {
      console.error("Failed to fetch versions", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchVersions();
    }
  }, [isOpen, materialId]);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setUploadFile(e.target.files[0]);
    }
  };

  const handleUploadSubmit = async () => {
    if (!uploadFile || !commitMessage.trim()) return;

    setIsSubmitting(true);

    // Simulate API Call - TODO: Replace with real upload logic if needed
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Create mock new version
    const newVersion: MaterialVersion = {
      id: `v-new-${Date.now()}`,
      version_number: `v${versions.length + 1}`, // Simple increment mock
      version: versions.length + 1,
      created_at: new Date().toISOString(),
      commit_message: commitMessage,
      author_name: "You",
      author_avatar: "https://github.com/shadcn.png"
    };

    setVersions([newVersion, ...versions]);

    // Reset form
    setUploadFile(null);
    setCommitMessage("");
    setIsUploading(false);
    setIsSubmitting(false);
  };

  const handleRestore = async (version: MaterialVersion) => {
    if (!confirm(`Are you sure you want to restore to version ${version.version_number}?`)) return;

    try {
      setActionLoading(version.id);
      await restoreMaterialVersion(materialId, version.version); // Assuming version.version is the number needed
      alert(`Successfully restored to ${version.version_number}`);
      if (onRestore) onRestore();
      setIsOpen(false);
    } catch (err) {
      console.error("Failed to restore version:", err);
      alert("Failed to restore version");
    } finally {
      setActionLoading(null);
    }
  };

  const handleDownload = async (version: MaterialVersion) => {
    try {
      setActionLoading(version.id);
      const url = await getMaterialDownloadUrl(materialId, version.version);
      window.open(url, '_blank');
    } catch (err) {
      console.error("Failed to download version:", err);
      alert("Failed to download version");
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <button
          className="px-6 py-2 ml-2 max-sm:px-4 max-sm:text-xs bg-white border-2 border-[#2B2D42] text-sm font-bold text-[#2B2D42] hover:bg-[#FF8811] hover:text-white transition-all shadow-[2px_2px_0px_0px_#2B2D42] hover:shadow-none hover:translate-x-0.5 hover:translate-y-0.5 flex items-center gap-2"
        >
          <History size={16} />
          History
        </button>
      </SheetTrigger>
      <SheetContent className="w-[400px] sm:w-[540px] border-l-4 border-black bg-[#FFFBF7] p-0 overflow-hidden flex flex-col z-[100]">
        <SheetHeader className="p-6 border-b-2 border-black bg-white flex flex-row items-center justify-between">
          <SheetTitle className="text-2xl font-black flex items-center gap-2">
            <History />
            Version History
          </SheetTitle>
          {!isUploading && (
            <button
              onClick={() => setIsUploading(true)}
              className="px-3 py-1.5 bg-[#FF8811] text-white text-xs font-bold rounded-md shadow-lg flex items-center gap-2 hover:bg-[#e67a0f] transition-all"
            >
              <Upload size={14} /> Upload New
            </button>
          )}
        </SheetHeader>

        <div className="flex-1 overflow-y-auto p-6">

          {/* Upload Form */}
          {isUploading && (
            <div className="mb-8 bg-white border-2 border-dashed border-[#FF8811] rounded-xl p-4 animate-in slide-in-from-top-4 fade-in duration-300">
              <div className="flex justify-between items-center mb-3">
                <h3 className="font-bold text-lg text-[#2B2D42]">Upload New Version</h3>
                <button onClick={() => setIsUploading(false)} className="text-gray-400 hover:text-black">
                  Cancel
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-xs font-bold text-gray-700 mb-1">Select File</label>
                  <div className="relative">
                    <input
                      type="file"
                      onChange={handleFileChange}
                      className="block w-full text-sm text-slate-500
                                                file:mr-4 file:py-2 file:px-4
                                                file:rounded-full file:border-0
                                                file:text-xs file:font-semibold
                                                file:bg-orange-50 file:text-orange-700
                                                hover:file:bg-orange-100
                                            "
                    />
                  </div>
                  {uploadFile && (
                    <div className="mt-2 flex items-center gap-2 text-xs font-semibold text-green-600 bg-green-50 p-2 rounded">
                      <FileText size={14} /> {uploadFile.name}
                    </div>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-bold text-gray-700 mb-1">Commit Message</label>
                  <textarea
                    value={commitMessage}
                    onChange={(e) => setCommitMessage(e.target.value)}
                    placeholder="Describe what changed in this version..."
                    className="w-full text-sm p-3 border-2 border-gray-200 rounded-lg focus:border-[#FF8811] focus:outline-none min-h-[80px]"
                  />
                </div>

                <button
                  onClick={handleUploadSubmit}
                  disabled={!uploadFile || !commitMessage || isSubmitting}
                  className="w-full py-2 bg-[#2B2D42] text-white font-bold rounded-lg hover:bg-black transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? <Loader2 className="animate-spin" size={16} /> : <Upload size={16} />}
                  {isSubmitting ? "Uploading..." : "Commit Changes"}
                </button>
              </div>
            </div>
          )}

          {loading ? (
            <div className="flex flex-col gap-4">
              {[1, 2, 3].map(i => (
                <div key={i} className="h-24 bg-gray-200 animate-pulse rounded-lg border-2 border-gray-300" />
              ))}
            </div>
          ) : (
            <div className="relative pl-4 space-y-8">
              {/* Vertical Line */}
              <div className="absolute left-[19px] top-2 bottom-2 w-[2px] bg-gray-300 -z-10" />

              {versions.map((version, index) => {
                const isCurrent = currentVersion == version.version || currentVersion == version.version_number;

                return (
                  <div key={version.id} className="relative flex gap-4 group">
                    {/* Timeline Node */}
                    <div className={`w-3 h-3 mt-1.5 rounded-full border-2 border-black z-10 shrink-0 ${index === 0 ? 'bg-[#FF8811]' : 'bg-white'}`} />

                    {/* Content Card */}
                    <div className={`flex-1 bg-white border-2 border-black rounded-lg p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] group-hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] group-hover:translate-x-[2px] group-hover:translate-y-[2px] transition-all ${isCurrent ? "ring-2 ring-offset-2 ring-[#FF8811]" : ""}`}>
                      <div className="flex justify-between items-start mb-2">
                        <div className="flex flex-col">
                          <span className="bg-black text-white text-xs font-bold px-2 py-1 rounded w-fit mb-1">
                            {version.version_number}
                          </span>
                          {isCurrent && <span className="text-[10px] font-bold text-[#FF8811]">Current Version</span>}
                        </div>

                        <span className="text-xs font-mono text-gray-500 font-bold">
                          {new Date(version.created_at).toLocaleDateString()}
                        </span>
                      </div>

                      <p className="font-bold text-lg mb-3 leading-tight">
                        {version.commit_message}
                      </p>

                      <div className="flex items-center gap-2 border-t border-gray-100 pt-3 justify-between">
                        <div className="flex items-center gap-2">
                          <Avatar className="h-6 w-6 border border-black">
                            <AvatarImage src={version.author_avatar} />
                            <AvatarFallback><User size={12} /></AvatarFallback>
                          </Avatar>
                          <span className="text-sm font-semibold text-gray-700">
                            {version.author_name}
                          </span>
                        </div>

                        <div className="flex gap-2">
                          <button
                            onClick={() => handleDownload(version)}
                            className="p-1.5 hover:bg-gray-100 rounded text-gray-600 transition"
                            title="Download this version"
                          >
                            <Download size={16} />
                          </button>

                          {!isCurrent && (
                            <button
                              onClick={() => handleRestore(version)}
                              disabled={actionLoading === version.id}
                              className="p-1.5 hover:bg-red-50 hover:text-red-600 rounded text-gray-600 transition disabled:opacity-50"
                              title="Restore to this version"
                            >
                              {actionLoading === version.id ? <Loader2 className="animate-spin" size={16} /> : <RotateCcw size={16} />}
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
};

export default VersionHistoryDialog;
