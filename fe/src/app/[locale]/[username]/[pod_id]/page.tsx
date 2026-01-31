"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Eye, BadgeCheck, FileText, Loader, Plus, Edit2, Trash2, X, UploadCloud, UserMinus } from "lucide-react";
import { useTranslations } from "next-intl";
import FileListItem from "@/components/knowledge-pod/FileListItem";
import { getPodDetail, getPodMaterials, UpdatePod } from "@/lib/api/pod";
import { updateMaterial, deleteMaterial } from "@/lib/api/material";
import { Pod } from "@/types/pod";
import { Material } from "@/types/material";
import { ProtectedRoute } from "@/components/auth";
import { useAuth } from "@/lib/auth-context";
import { usePersona, useRecommendationTrigger } from "@/hooks/usePersona";
import { PersonaRecommendationPopup } from "@/components/persona";
import { useCollaborators } from "@/hooks/useCollaborators";
import AddCollaboratorModal from "@/components/knowledge-pod/AddCollaboratorModal";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import { uploadMaterial } from "@/lib/api/uploadMaterial";
import Link from "next/link";

interface PageProps {
  params: Promise<{
    username: string;
    pod_id: string;
  }>;
}

interface EditingMaterial {
  id: string;
  title: string;
  description: string;
}

export default function KnowledgePodDetail({ params }: PageProps) {
  const router = useRouter();
  const { username, pod_id } = React.use(params);
  const { user } = useAuth();

  // Persona hooks
  const { persona, loading: personaLoading } = usePersona(user?.id);
  const { shouldShow: showRecommendation, dismiss: dismissRecommendation } = useRecommendationTrigger(persona);
  const t = useTranslations("collaborator");
  const tPod = useTranslations("podDetail");
  const tCommon = useTranslations("common");

  const [pod, setPod] = useState<Pod | null>(null);
  const [materials, setMaterials] = useState<Material[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isNotFound, setIsNotFound] = useState(false);

  const [podVisibility, setPodVisibility] = useState<"public" | "private">("public");
  const [isUpdatingVisibility, setIsUpdatingVisibility] = useState(false);

  // Collaborator state
  const [isAddCollaboratorOpen, setIsAddCollaboratorOpen] = useState(false);

  // Collaborator hook
  const {
    collaborators,
    loading: collaboratorsLoading,
    error: collaboratorError,
    inviting,
    removing,
    fetchCollaborators,
    invite,
    remove,
    clearError: clearCollaboratorError,
    getPermissions,
    verifiedCollaborators,
  } = useCollaborators(pod_id);

  // Compute permissions (use email for reliable cross-service validation)
  const currentUserId = user?.id || "";
  const currentUserEmail = user?.email || "";
  const ownerId = pod?.owner_id || "";
  const permissions = getPermissions(currentUserId, ownerId, currentUserEmail);
  const isOwner = permissions.isOwner;

  // Edit material states
  const [editingMaterial, setEditingMaterial] = useState<EditingMaterial | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isUpdatingMaterial, setIsUpdatingMaterial] = useState(false);
  const [isDeletingMaterial, setIsDeletingMaterial] = useState<string | null>(null);

  const handleUpdateVisibility = async (newVisibility: boolean) => {
    // newVisibility: true = public, false = private
    const visibility = newVisibility ? "public" : "private";

    try {
      setIsUpdatingVisibility(true);
      await UpdatePod(pod_id, { visibility });
      await new Promise(resolve => setTimeout(resolve, 2000));
      setPodVisibility(visibility);
    } catch (err) {
      console.error("Failed to update visibility:", err);
      alert("Failed to update visibility");
      // Revert state jika error
      setPodVisibility(visibility === "public" ? "private" : "public");
    } finally {
      setIsUpdatingVisibility(false);
    }
  };



  // Handle Edit Material
  const handleEditMaterial = (material: Material) => {
    setEditingMaterial({
      id: material.id,
      title: material.title,
      description: material.description || "",
    });
    setIsEditDialogOpen(true);
  };

  // Handle Save Material Changes
  const handleSaveMaterialChanges = async () => {
    if (!editingMaterial) return;

    try {
      setIsUpdatingMaterial(true);
      await updateMaterial(editingMaterial.id, {
        title: editingMaterial.title,
        description: editingMaterial.description,
      });

      // Update local materials list
      setMaterials((prev) =>
        prev.map((m) =>
          m.id === editingMaterial.id
            ? {
              ...m,
              title: editingMaterial.title,
              description: editingMaterial.description,
            }
            : m
        )
      );

      setIsEditDialogOpen(false);
      setEditingMaterial(null);
    } catch (err) {
      console.error("Failed to update material:", err);
      alert("Failed to update material");
    } finally {
      setIsUpdatingMaterial(false);
    }
  };

  // Handle Delete Material
  const handleDeleteMaterial = async (materialId: string) => {
    if (!confirm("Are you sure you want to delete this material?")) return;

    try {
      setIsDeletingMaterial(materialId);
      await deleteMaterial(materialId);

      // Remove from local materials list
      setMaterials((prev) => prev.filter((m) => m.id !== materialId));
    } catch (err) {
      console.error("Failed to delete material:", err);
      alert("Failed to delete material");
    } finally {
      setIsDeletingMaterial(null);
    }
  };

  // Handle Remove Collaborator
  const handleRemoveCollaborator = async (userId: string, userName: string) => {
    if (!confirm(t("confirmRemove"))) return;
    await remove(userId);
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const podData = await getPodDetail(pod_id);
        setPod(podData);
        setPodVisibility(podData.visibility as "public" | "private");

        const podMaterials = await getPodMaterials(pod_id);
        setMaterials(podMaterials);

        // Fetch collaborators
        await fetchCollaborators();
      } catch (err) {
        setError("Failed to load pod");
        setIsNotFound(true);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [pod_id, username, fetchCollaborators]);

  if (loading) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen items-center justify-center">
          <Loader className="animate-spin text-orange-500" size={32} />
        </div>
      </ProtectedRoute>
    );
  }

  if (isNotFound || !pod) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen flex-col items-center justify-center gap-4">
          <h1 className="text-5xl font-black">404</h1>
          <p className="text-sm text-zinc-500">{error}</p>
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      {/* Persona Recommendation Popup */}
      <PersonaRecommendationPopup
        isOpen={showRecommendation}
        onClose={dismissRecommendation}
        persona={persona}
        onActionClick={(rec) => {
          console.log("Recommendation clicked:", rec);
          // TODO: Handle different action types
        }}
      />

      {/* Edit Material Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-md border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black]">
          <DialogHeader>
            <DialogTitle className="text-xl font-black">Edit Material</DialogTitle>
          </DialogHeader>

          {editingMaterial && (
            <div className="space-y-4">
              {/* Title Input */}
              <div>
                <label className="block text-sm font-bold text-black mb-2">
                  Title
                </label>
                <input
                  type="text"
                  value={editingMaterial.title}
                  onChange={(e) =>
                    setEditingMaterial({ ...editingMaterial, title: e.target.value })
                  }
                  className="w-full px-4 py-2 border-2 border-black rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  placeholder="Material title"
                />
              </div>

              {/* Description Input */}
              <div>
                <label className="block text-sm font-bold text-black mb-2">
                  Description
                </label>
                <textarea
                  value={editingMaterial.description}
                  onChange={(e) =>
                    setEditingMaterial({
                      ...editingMaterial,
                      description: e.target.value,
                    })
                  }
                  className="w-full px-4 py-2 border-2 border-black rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                  rows={4}
                  placeholder="Material description"
                />
              </div>
            </div>
          )}

          <DialogFooter className="gap-2">
            <button
              onClick={() => setIsEditDialogOpen(false)}
              className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-white hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSaveMaterialChanges}
              disabled={isUpdatingMaterial}
              className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-[#FF8811] text-white shadow-[2px_2px_0_0_black] hover:shadow-[1px_1px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all disabled:opacity-50"
            >
              {isUpdatingMaterial ? "Saving..." : "Save Changes"}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Collaborator Modal */}
      <AddCollaboratorModal
        open={isAddCollaboratorOpen}
        onOpenChange={setIsAddCollaboratorOpen}
        onInvite={invite}
        inviting={inviting}
        error={collaboratorError}
        clearError={clearCollaboratorError}
        currentUserEmail={user?.email || ""}
        existingCollaboratorEmails={collaborators.map((c) => c.user_email || "").filter(Boolean)}
      />

      <div className="mx-auto max-w-6xl px-4 md:px-6 lg:px-8 py-6 space-y-6">
        {/* ================= HEADER ================= */}
        <header className="space-y-4">
          {/* TOP ROW — TITLE + USE + UPLOAD */}
          <div className="grid grid-cols-1 md:grid-cols-[1fr_auto_auto] gap-3 items-center">
            <h1 className="text-2xl md:text-3xl font-black tracking-tight text-black">
              {pod.name}
            </h1>

            {/* Upload button - only show if user can upload */}
            {permissions.canUpload && (
              <Link href={`/${pod.owner_id}/${pod.id}/upload`} >
                <button
                  className="w-full md:w-auto flex items-center justify-center gap-2 rounded-lg border-2 border-black bg-[#FF8811] px-6 py-2 text-sm font-bold text-white shadow-[3px_3px_0_0_black] hover:shadow-[2px_2px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all">
                  <UploadCloud size={16} />
                  Upload
                </button>
              </Link>
            )}

          </div>

          {/* BOTTOM ROW — COLLAB + SWITCH + SAVE */}
          <div className="grid grid-cols-1 md:grid-cols-[1fr_auto] gap-3 items-center">
            {/* LEFT SIDE */}
            <div className="flex flex-wrap items-center gap-3">
              {/* COLLABORATOR AVATARS (max 2 + overflow) */}
              <div className="flex items-center -space-x-2">
                {/* Show up to 2 collaborator avatars + overflow indicator */}
                {verifiedCollaborators.slice(0, 2).map((collab) => (
                  <Avatar
                    key={collab.id}
                    className="h-9 w-9 border-2 border-black"
                    title={collab.user_name || collab.user_email || "Collaborator"}
                  >
                    <AvatarImage src={collab.user_avatar_url} />
                    <AvatarFallback className="bg-blue-100 text-blue-700 text-xs font-bold">
                      {(collab.user_name || "U").substring(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                ))}
                {/* Overflow indicator */}
                {verifiedCollaborators.length > 2 && (
                  <div
                    className="flex h-9 w-9 items-center justify-center rounded-full border-2 border-black bg-gray-100 text-xs font-bold text-gray-700"
                    title={`+${verifiedCollaborators.length - 2} more collaborators`}
                  >
                    +{verifiedCollaborators.length - 2}
                  </div>
                )}
              </div>

              {/* ADD COLLABORATOR BUTTON */}
              <div className="flex items-center -space-x-2">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      className="z-20 flex h-9 w-9 items-center justify-center rounded-full border-2 border-black bg-orange-500 text-white shadow-[2px_2px_0_0_black]"
                      title={permissions.canManageCollaborators ? t("addCollaborator") : t("title")}
                    >
                      <Plus size={16} />
                    </button>
                  </DropdownMenuTrigger>

                  <DropdownMenuContent className="w-72 border-2 border-black bg-[#FFFBF7] p-3 shadow-[4px_4px_0_0_black]">
                    <DropdownMenuLabel className="text-sm font-bold text-center">
                      {t("title")}
                    </DropdownMenuLabel>

                    {/* Add Collaborator Button - only if user can manage */}
                    {permissions.canManageCollaborators && (
                      <>
                        <DropdownMenuItem
                          className="flex items-center gap-2 rounded-md px-2 py-2 cursor-pointer bg-orange-50 hover:bg-orange-100 mt-2"
                          onClick={() => setIsAddCollaboratorOpen(true)}
                        >
                          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-orange-500 text-white">
                            <Plus size={14} />
                          </div>
                          <span className="text-xs font-bold text-orange-700">{t("addCollaborator")}</span>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator className="my-2" />
                      </>
                    )}

                    <div className="mt-2 space-y-1 max-h-60 overflow-y-auto">
                      {collaboratorsLoading ? (
                        <div className="flex items-center justify-center py-4">
                          <Loader className="w-4 h-4 animate-spin text-orange-500" />
                        </div>
                      ) : verifiedCollaborators.length === 0 ? (
                        <p className="text-xs text-gray-500 text-center py-2">
                          {t("noCollaborators")}
                        </p>
                      ) : (
                        verifiedCollaborators.map((collab) => (
                          <DropdownMenuItem
                            key={collab.id}
                            className="flex items-center gap-2 rounded-md px-2 py-1 cursor-default"
                          >
                            <Avatar className="h-7 w-7 border-2 border-black">
                              <AvatarFallback className="bg-gray-100 text-xs font-bold">
                                {(collab.user_name || "U").substring(0, 2).toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div className="flex-1 min-w-0">
                              <span className="text-xs font-bold block truncate">
                                {collab.user_name || collab.user_id.substring(0, 8)}
                              </span>
                              <span className="text-[10px] text-gray-500">
                                {t(`roles.${collab.role}`)}
                              </span>
                            </div>
                            {/* Remove button - only if user can manage and not removing self */}
                            {permissions.canManageCollaborators && collab.user_id !== currentUserId && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleRemoveCollaborator(collab.user_id, collab.user_name || "");
                                }}
                                disabled={removing === collab.user_id}
                                className="p-1 rounded hover:bg-red-100 text-red-500 transition-colors"
                                title={t("remove")}
                              >
                                {removing === collab.user_id ? (
                                  <Loader className="w-3 h-3 animate-spin" />
                                ) : (
                                  <UserMinus size={14} />
                                )}
                              </button>
                            )}
                          </DropdownMenuItem>
                        ))
                      )}
                    </div>
                  </DropdownMenuContent>
                </DropdownMenu>

                {/* Display first few collaborator avatars */}
                {verifiedCollaborators.slice(0, 2).map((collab) => (
                  <Avatar key={collab.id} className="h-9 w-9 border-2 border-black bg-white">
                    <AvatarFallback className="text-xs font-bold">
                      {(collab.user_name || "U").substring(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                ))}

                {/* Show count if more than 2 collaborators */}
                {verifiedCollaborators.length > 2 && (
                  <div className="flex h-9 w-9 items-center justify-center rounded-full border-2 border-black bg-[#FFFBF7] text-xs font-bold shadow-[2px_2px_0_0_black]">
                    +{verifiedCollaborators.length - 2}
                  </div>
                )}
              </div>

              {/* PRIVATE / PUBLIC SWITCH - only show if user is owner */}
              {permissions.isOwner && (
                <div className="flex items-center gap-2 rounded-full border-2 border-black bg-white px-3 py-1 shadow-[2px_2px_0_0_black] relative">
                  <span className={`text-xs font-bold ${podVisibility === "private" ? "text-black" : "text-zinc-400"}`}>
                    Private
                  </span>

                  {isUpdatingVisibility ? (
                    <div className="flex items-center justify-center h-6 w-12 mx-1">
                      <div className="w-4 h-4 border-2 border-orange-500 border-t-transparent rounded-full animate-spin" />
                    </div>
                  ) : (
                    <Switch
                      checked={podVisibility === "public"}
                      onCheckedChange={handleUpdateVisibility}
                      disabled={isUpdatingVisibility}
                      className="data-[state=checked]:bg-orange-500"
                    />
                  )}

                  <span className={`text-xs font-bold ${podVisibility === "public" ? "text-black" : "text-zinc-400"}`}>
                    Public
                  </span>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* ================= DESCRIPTION ================= */}
        <div className="rounded-xl border-2 border-black bg-white p-4 shadow-[4px_4px_0_0_black]">
          <p className="text-xs font-semibold text-zinc-400 mb-1">Description</p>
          <p className="text-sm font-medium text-black">
            {pod.description || "No description available"}
          </p>
        </div>

        {/* ================= FILE LIST ================= */}
        <div className="overflow-hidden rounded-xl border-2 border-black bg-white shadow-[4px_4px_0_0_black]">
          {materials.length === 0 ? (
            <p className="p-6 text-center text-sm text-zinc-500">
              No materials in this pod yet
            </p>
          ) : (
            <div>
              {materials.map((material, index) => (
                <div
                  key={material.id || index}
                  className={`flex items-center justify-between p-4 hover:bg-gray-50 transition-colors ${index !== materials.length - 1 ? "border-b-2 border-gray-200" : ""
                    }`}
                >
                  {/* Material Content */}
                  <div
                    className="flex-1 cursor-pointer"
                    onClick={() =>
                      router.push(
                        `/${pod.owner_id}/${pod.id}/${material.id}`
                      )
                    }
                  >
                    <FileListItem
                      variant="file"
                      materialId={material.id}
                      userId={pod.owner_id}
                      podId={pod.id}
                      title={material.title}
                      description={material.description || ""}
                      likes={material.download_count.toString()}
                      date={new Date(material.created_at).toLocaleDateString()}
                      isLast={index === materials.length - 1}
                    />
                  </div>

                  {/* Edit & Delete Actions - only show if user has permission */}
                  {permissions.canEdit && (
                    <div className="flex items-center gap-2 ml-4">
                      <button
                        onClick={() => handleEditMaterial(material)}
                        className="p-2 rounded-lg border-2 border-black bg-blue-100 hover:bg-blue-200 transition-colors"
                        title="Edit material"
                      >
                        <Edit2 size={16} className="text-blue-600" />
                      </button>

                      {permissions.canDelete && (
                        <button
                          onClick={() => handleDeleteMaterial(material.id)}
                          disabled={isDeletingMaterial === material.id}
                          className="p-2 rounded-lg border-2 border-black bg-red-100 hover:bg-red-200 transition-colors disabled:opacity-50"
                          title="Delete material"
                        >
                          <Trash2 size={16} className="text-red-600" />
                        </button>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}