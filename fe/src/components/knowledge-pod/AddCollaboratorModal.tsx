"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2, UserPlus, Check, AlertCircle, Mail } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import { CollaboratorRole, CollaboratorError } from "@/types/collaborator";
import { getCollaboratorErrorKey } from "@/hooks/useCollaborators";

// =============================================================================
// TYPES
// =============================================================================

interface AddCollaboratorModalProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onInvite: (email: string, role: CollaboratorRole) => Promise<boolean>;
    inviting: boolean;
    error: CollaboratorError | null;
    clearError: () => void;
    currentUserEmail: string;
    existingCollaboratorEmails: string[];
}

// =============================================================================
// COMPONENT
// =============================================================================

export default function AddCollaboratorModal({
    open,
    onOpenChange,
    onInvite,
    inviting,
    error,
    clearError,
    currentUserEmail,
    existingCollaboratorEmails,
}: AddCollaboratorModalProps) {
    const t = useTranslations("collaborator");

    // State
    const [email, setEmail] = useState("");
    const [selectedRole, setSelectedRole] = useState<CollaboratorRole>("contributor");
    const [localSuccess, setLocalSuccess] = useState(false);
    const [validationError, setValidationError] = useState<string | null>(null);

    // ==========================================================================
    // VALIDATION
    // ==========================================================================

    const validateEmail = (emailValue: string): string | null => {
        if (!emailValue.trim()) {
            return null; // Empty is not an error, just not ready
        }

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailValue.trim())) {
            return t("errors.invalidEmail");
        }

        if (emailValue.trim().toLowerCase() === currentUserEmail.toLowerCase()) {
            return t("errors.cannotInviteSelf");
        }

        if (existingCollaboratorEmails.some(e => e.toLowerCase() === emailValue.trim().toLowerCase())) {
            return t("errors.alreadyCollaborator");
        }

        return null;
    };

    // ==========================================================================
    // HANDLE EMAIL INPUT CHANGE
    // ==========================================================================

    const handleEmailChange = (value: string) => {
        setEmail(value);
        setLocalSuccess(false);
        clearError();
        setValidationError(validateEmail(value));
    };

    // ==========================================================================
    // HANDLE INVITE
    // ==========================================================================

    const handleInvite = async () => {
        const validationErr = validateEmail(email);
        if (validationErr) {
            setValidationError(validationErr);
            return;
        }

        const success = await onInvite(email.trim(), selectedRole);

        if (success) {
            setLocalSuccess(true);
            // Reset form after short delay
            setTimeout(() => {
                setEmail("");
                setLocalSuccess(false);
                setValidationError(null);
                onOpenChange(false);
            }, 1500);
        }
    };

    // ==========================================================================
    // HANDLE CLOSE
    // ==========================================================================

    const handleClose = () => {
        setEmail("");
        setValidationError(null);
        setLocalSuccess(false);
        clearError();
        onOpenChange(false);
    };

    // Check if form is ready to submit
    const isValidEmail = email.trim() && !validateEmail(email);

    // ==========================================================================
    // RENDER
    // ==========================================================================

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="sm:max-w-md border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black]">
                <DialogHeader>
                    <DialogTitle className="text-xl font-black flex items-center gap-2">
                        <UserPlus size={20} />
                        {t("addCollaborator")}
                    </DialogTitle>
                </DialogHeader>

                <div className="space-y-4 py-4">
                    {/* Email Input */}
                    <div>
                        <label className="block text-sm font-bold text-black mb-2">
                            {t("enterEmail")}
                        </label>
                        <div className="relative">
                            <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => handleEmailChange(e.target.value)}
                                placeholder={t("emailPlaceholder")}
                                className="w-full pl-10 pr-4 py-3 border-2 border-black rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                                disabled={inviting || localSuccess}
                            />
                        </div>
                        <p className="text-xs text-gray-500 mt-1">
                            {t("emailHint")}
                        </p>
                    </div>

                    {/* Validation Error */}
                    {validationError && (
                        <div className="flex items-center gap-2 p-3 bg-red-50 border-2 border-red-200 rounded-lg text-red-700 text-sm">
                            <AlertCircle size={16} />
                            {validationError}
                        </div>
                    )}

                    {/* API Error */}
                    {error && (
                        <div className="flex items-center gap-2 p-3 bg-red-50 border-2 border-red-200 rounded-lg text-red-700 text-sm">
                            <AlertCircle size={16} />
                            {t(getCollaboratorErrorKey(error).replace("collaborator.", ""))}
                        </div>
                    )}

                    {/* Success Message */}
                    {localSuccess && (
                        <div className="flex items-center gap-2 p-3 bg-green-50 border-2 border-green-200 rounded-lg text-green-700 text-sm">
                            <Check size={16} />
                            {t("success.invited")}
                        </div>
                    )}

                    {/* Role Selection - Show when valid email entered */}
                    {isValidEmail && !localSuccess && (
                        <div className="border-2 border-black rounded-lg p-4 bg-white">
                            <label className="block text-xs font-bold text-gray-500 mb-2 uppercase">
                                {t("selectRole")}
                            </label>
                            <div className="grid grid-cols-3 gap-2">
                                {(["contributor", "viewer", "admin"] as CollaboratorRole[]).map(
                                    (role) => (
                                        <button
                                            key={role}
                                            type="button"
                                            onClick={() => setSelectedRole(role)}
                                            className={`px-3 py-2 text-sm font-bold rounded-lg border-2 transition-all ${selectedRole === role
                                                ? "border-orange-500 bg-orange-50 text-orange-700"
                                                : "border-gray-200 bg-white text-gray-700 hover:border-gray-300"
                                                }`}
                                            disabled={inviting}
                                        >
                                            {t(`roles.${role}`)}
                                        </button>
                                    )
                                )}
                            </div>
                            <p className="text-xs text-gray-500 mt-2">
                                {t(`roleDescriptions.${selectedRole}`)}
                            </p>
                        </div>
                    )}
                </div>

                <DialogFooter className="gap-2">
                    <button
                        onClick={handleClose}
                        className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-white hover:bg-gray-50 transition-colors"
                        disabled={inviting}
                    >
                        {t("cancel")}
                    </button>

                    <button
                        onClick={handleInvite}
                        disabled={!isValidEmail || inviting || localSuccess}
                        className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-[#FF8811] text-white shadow-[2px_2px_0_0_black] hover:shadow-[1px_1px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        {inviting && <Loader2 className="w-4 h-4 animate-spin" />}
                        {inviting ? t("inviting") : t("add")}
                    </button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
