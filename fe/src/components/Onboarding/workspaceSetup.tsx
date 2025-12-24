"use client";

import { useEffect, useState } from "react";
import { Plus, X, Check, Loader2, ChevronDown, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { Interest } from "@/types";
import { completeOnboarding, getAllInterests, setUserInterests } from "@/lib/interests";
import { Checkbox } from "@/components/ui/checkbox";

interface WorkspaceSetupProps {
    onComplete?: () => void;
}

export function WorkspaceSetup({ onComplete }: WorkspaceSetupProps) {
    const router = useRouter();

    const [interests, setInterests] = useState<Record<string, Interest[]>>({});
    const [expandedCategories, setExpandedCategories] = useState<string[]>([]);
    const [selectedInterestIds, setSelectedInterestIds] = useState<string[]>([]);
    const [isAdding, setIsAdding] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [newInterest, setNewInterest] = useState("");
    const [isSaving, setIsSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);


    const [viewMode, setViewMode] = useState<"category" | "predefined">("category");


    useEffect(() => {
        const loadInterests = async () => {
            try {
                setIsLoading(true);

                const fetchedInterests = await getAllInterests();

                setInterests(fetchedInterests);
                // Expand the first category by default
                const categories = Object.keys(fetchedInterests);
                if (categories.length > 0) {
                    setExpandedCategories([categories[0]]);
                }
            } catch (err) {
                console.error("Error fetching interests:", err);
                setError("Failed to load interests. Please try again.");
            } finally {
                setIsLoading(false);
            }
        };
        loadInterests();
    }, []);


    const toggleInterest = (interestId: string) => {
        if (selectedInterestIds.includes(interestId)) {
            setSelectedInterestIds(selectedInterestIds.filter((id) => id !== interestId));
        } else {
            setSelectedInterestIds([...selectedInterestIds, interestId]);
        }
    }

    const toggleCategory = (category: string) => {
        if (expandedCategories.includes(category)) {
            setExpandedCategories(expandedCategories.filter((c) => c !== category));
        } else {
            setExpandedCategories([...expandedCategories, category]);
        }
    }

    const handleAddInterest = (e: React.FormEvent) => {
        e?.preventDefault();
        if (newInterest.trim()) {
            const interestName = newInterest.trim();

            // Search in all categories
            let existingInterest: Interest | undefined;
            for (const category in interests) {
                const found = interests[category].find(i => i.name.toLowerCase() === interestName.toLowerCase());
                if (found) {
                    existingInterest = found;
                    break;
                }
            }

            if (existingInterest) {
                // If it exists, just select it
                if (!selectedInterestIds.includes(existingInterest.id)) {
                    setSelectedInterestIds((prev) => [...prev, existingInterest!.id]);
                }
            } else {
                // Create a temporary interest in "Custom" category
                const tempId = `temp-${Date.now()}`;
                const newInterestObj: Interest = {
                    id: tempId,
                    name: interestName,
                    category: "Custom",
                };

                setInterests(prev => ({
                    ...prev,
                    "Custom": [...(prev["Custom"] || []), newInterestObj]
                }));

                if (!expandedCategories.includes("Custom")) {
                    setExpandedCategories(prev => [...prev, "Custom"]);
                }

                setSelectedInterestIds((prev) => [...prev, tempId]);
            }

            setNewInterest("");
            setIsAdding(false);
        }
    };

    const handleSubmit = async () => {
        if (selectedInterestIds.length === 0) {
            setError("Please select at least one interest");
            return;
        }

        try {
            setIsSaving(true);
            setError(null);

            // Separate predefined interests from custom interests
            const predefinedInterestIds = selectedInterestIds.filter(
                (id) => !id.startsWith("temp-")
            );

            // Get custom interest names from temp IDs
            const customInterests: string[] = [];
            selectedInterestIds
                .filter((id) => id.startsWith("temp-"))
                .forEach((tempId) => {
                    // Find the custom interest in the "Custom" category
                    const customCategory = interests["Custom"] || [];
                    const customInterest = customCategory.find((i) => i.id === tempId);
                    if (customInterest) {
                        customInterests.push(customInterest.name);
                    }
                });

            // Save selected interests (both predefined and custom)
            await setUserInterests(predefinedInterestIds, customInterests);

            // Mark onboarding as complete
            await completeOnboarding();

            // Trigger completion callback (shows loading screen)
            if (onComplete) {
                onComplete();
            } else {
                // Fallback: Redirect to dashboard directly
                router.push("/dashboard");
            }
        } catch (err: any) {
            setError(err.response?.data?.message || "Failed to save interests. Please try again.");
            console.error("Error saving interests:", err);
        } finally {
            setIsSaving(false);
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen bg-[#FFFBF5]">
                <Loader2 className="w-8 h-8 animate-spin text-[#FF8811]" />
            </div>
        );
    }





    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-[#FFFBF5] p-4">
            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="w-full max-w-3xl"
            >
                <h1 className="text-3xl md:text-4xl font-bold text-[#1A1A1A] mb-12 text-center">
                    Initialize Workspace
                </h1>

                <div className="relative">
                    {/* Main Card */}
                    <div className="relative z-10 bg-white rounded-lg border-2 border-black p-8">
                        <div className="mb-6">
                            <h2 className="text-xl font-bold text-black mb-2">
                                1. Tell us What you're interested in
                            </h2>
                            <p className="text-sm text-gray-500 leading-relaxed">
                                Initialize your workspace by identifying your interested material
                            </p>
                        </div>

                        <div className="flex justify-end h-8 items-center mb-4">
                            {isAdding ? (
                                <form onSubmit={handleAddInterest} className="flex items-center gap-2">
                                    <input
                                        type="text"
                                        value={newInterest}
                                        onChange={(e) => setNewInterest(e.target.value)}
                                        placeholder="Add interest..."
                                        className="px-2 py-1 text-sm border-2 border-black rounded-md focus:outline-none focus:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] w-32 font-medium"
                                        autoFocus
                                    />
                                    <button
                                        type="submit"
                                        className="p-1 hover:bg-green-100 cursor-pointer text-green-600 rounded-full transition-colors border-2 border-transparent hover:border-black"
                                    >
                                        <Check className="w-4 h-4" strokeWidth={3} />
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setIsAdding(false);
                                            setNewInterest("");
                                        }}
                                        className="p-1 hover:bg-red-100 cursor-pointer text-red-600 rounded-full transition-colors border-2 border-transparent hover:border-black"
                                    >
                                        <X className="w-4 h-4" strokeWidth={3} />
                                    </button>
                                </form>
                            ) : (
                                <div className="flex items-center gap-4">
                                    <button
                                        onClick={() => setIsAdding(true)}
                                        className="p-1 hover:bg-gray-100 cursor-pointer rounded-full transition-colors group flex items-center gap-2"
                                    >
                                        <span className="text-xs font-bold opacity-0 group-hover:opacity-100 transition-opacity text-gray-500 ml-2">Add New</span>
                                        <Plus className="w-6 h-6 text-black group-hover:scale-110 transition-transform" strokeWidth={3} />
                                    </button>

                                    <div className="flex items-center border-2 border-black rounded-md overflow-hidden">
                                        <button
                                            onClick={() => setViewMode("category")}
                                            className={cn(
                                                "px-3 py-1 text-xs font-bold transition-colors",
                                                viewMode === "category"
                                                    ? "bg-[#FF8811] text-white"
                                                    : "bg-white text-black hover:bg-gray-100"
                                            )}
                                        >
                                            Category
                                        </button>
                                        <div className="w-[2px] bg-black self-stretch"></div>
                                        <button
                                            onClick={() => setViewMode("predefined")}
                                            className={cn(
                                                "px-3 py-1 text-xs font-bold transition-colors",
                                                viewMode === "predefined"
                                                    ? "bg-[#FF8811] text-white"
                                                    : "bg-white text-black hover:bg-gray-100"
                                            )}
                                        >
                                            Predefined
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>

                        <div className="space-y-4 mb-8">
                            {viewMode === "category" ? (
                                Object.entries(interests).map(([category, categoryInterests]) => {
                                    const isExpanded = expandedCategories.includes(category);
                                    return (
                                        <div key={category} className="">
                                            <button
                                                onClick={() => toggleCategory(category)}
                                                className={cn(
                                                    "w-full flex items-center justify-between px-4 py-3 font-bold text-lg text-left border-2 border-black rounded-md transition-all duration-200",
                                                    isExpanded
                                                        ? "bg-[#FF8811] text-white shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] translate-x-[2px] translate-y-[2px]"
                                                        : "bg-white text-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[3px_3px_0px_0px_rgba(0,0,0,1)]"
                                                )}
                                            >
                                                <span className="flex items-center gap-2">
                                                    {category}
                                                </span>
                                                {isExpanded ? <ChevronDown className="w-6 h-6" strokeWidth={3} /> : <ChevronRight className="w-6 h-6" strokeWidth={3} />}
                                            </button>

                                            {isExpanded && (
                                                <div className="mt-2 mx-1 p-4 border-2 border-black rounded-md bg-white shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                                        {categoryInterests.map((interest, index) => (
                                                            <div key={interest.id || `${category}-${index}`} className="flex items-center space-x-3 p-2 rounded-md hover:bg-gray-50 transition-colors border-2 border-transparent hover:border-black/10">
                                                                <Checkbox
                                                                    id={interest.id}
                                                                    checked={selectedInterestIds.includes(interest.id)}
                                                                    onCheckedChange={() => toggleInterest(interest.id)}
                                                                    className="w-5 h-5 border-2 border-black data-[state=checked]:bg-[#FF8811] data-[state=checked]:text-white data-[state=checked]:border-black rounded-sm"
                                                                />
                                                                <label
                                                                    htmlFor={interest.id}
                                                                    className="text-sm font-bold leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer flex-1"
                                                                >
                                                                    {interest.name}
                                                                </label>
                                                            </div>
                                                        ))}
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    );
                                })
                            ) : (
                                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                                    {Object.values(interests).flat().map((interest) => {
                                        const isSelected = selectedInterestIds.includes(interest.id);
                                        return (
                                            <button
                                                key={interest.id}
                                                onClick={() => toggleInterest(interest.id)}
                                                className={cn(
                                                    "px-4 py-3 text-sm font-bold border-2 border-black rounded-md transition-all duration-200",
                                                    isSelected
                                                        ? "bg-[#FF8811] text-white"
                                                        : "bg-white text-black hover:bg-gray-50"
                                                )}
                                                style={{
                                                    boxShadow: isSelected
                                                        ? "2px 2px 0px 0px rgba(0,0,0,1)"
                                                        : "4px 4px 0px 0px rgba(0,0,0,1)",
                                                    transform: isSelected
                                                        ? "translate(2px, 2px)"
                                                        : "translate(0px, 0px)",
                                                }}
                                            >
                                                {interest.name}
                                            </button>
                                        );
                                    })}
                                </div>
                            )}
                        </div>


                        <button className="w-full py-4 bg-[#FF8811] text-white font-bold rounded-md border-2 border-black hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all active:translate-x-[4px] active:translate-y-[4px] active:shadow-none disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-x-0 disabled:hover:translate-y-0 disabled:hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] flex items-center justify-center gap-2"
                            onClick={handleSubmit}
                            disabled={isSaving || selectedInterestIds.length === 0}
                        >
                            {isSaving ? (
                                <>
                                    <Loader2 className="w-5 h-5 animate-spin" />
                                    <span>Saving...</span>
                                </>
                            ) : (
                                "Save and Continue"
                            )}

                        </button>



                    </div>

                    {/* Card Shadow Effect */}
                    <div className="absolute top-2 -right-2 w-full h-full bg-[#FF8811] rounded-lg border-2 border-black z-0" />
                </div>
            </motion.div>
        </div>
    );
}
