"use client";

import { useState, useEffect, useCallback } from "react";
import { fetchUserPersona, PredictPersonaResponse } from "@/lib/api/behavior";

const PERSONA_STORAGE_KEY = "ngasihtau_persona";
const PERSONA_TIMESTAMP_KEY = "ngasihtau_persona_timestamp";
const PERSONA_CACHE_DURATION = 30 * 60 * 1000; // 30 minutes

interface PersonaState {
  persona: PredictPersonaResponse | null;
  loading: boolean;
  error: string | null;
}

interface UsePersonaReturn extends PersonaState {
  refreshPersona: () => Promise<void>;
  clearPersona: () => void;
}

/**
 * Hook untuk mengelola persona user dengan caching di localStorage
 * 
 * Flow:
 * 1. Cek localStorage untuk cached persona
 * 2. Jika cache valid (< 30 menit), gunakan cached data
 * 3. Jika cache expired atau tidak ada, fetch dari API
 * 4. Simpan hasil ke localStorage
 */
export function usePersona(userId: string | undefined): UsePersonaReturn {
  const [state, setState] = useState<PersonaState>({
    persona: null,
    loading: true,
    error: null,
  });

  // Load cached persona from localStorage
  const loadCachedPersona = useCallback((): PredictPersonaResponse | null => {
    if (typeof window === "undefined") return null;

    try {
      const cached = localStorage.getItem(PERSONA_STORAGE_KEY);
      const timestamp = localStorage.getItem(PERSONA_TIMESTAMP_KEY);

      if (!cached || !timestamp) return null;

      const cacheAge = Date.now() - parseInt(timestamp, 10);
      if (cacheAge > PERSONA_CACHE_DURATION) {
        // Cache expired
        return null;
      }

      const parsed = JSON.parse(cached) as PredictPersonaResponse;
      
      // Verify it's for the same user
      if (parsed.user_id !== userId) {
        return null;
      }

      return parsed;
    } catch {
      return null;
    }
  }, [userId]);

  // Save persona to localStorage
  const savePersonaToCache = useCallback((persona: PredictPersonaResponse) => {
    if (typeof window === "undefined") return;

    try {
      localStorage.setItem(PERSONA_STORAGE_KEY, JSON.stringify(persona));
      localStorage.setItem(PERSONA_TIMESTAMP_KEY, Date.now().toString());
    } catch (err) {
      console.warn("Failed to cache persona:", err);
    }
  }, []);

  // Clear cached persona
  const clearPersona = useCallback(() => {
    if (typeof window === "undefined") return;

    localStorage.removeItem(PERSONA_STORAGE_KEY);
    localStorage.removeItem(PERSONA_TIMESTAMP_KEY);
    setState({ persona: null, loading: false, error: null });
  }, []);

  // Fetch persona from API
  const fetchPersona = useCallback(async () => {
    if (!userId) {
      setState({ persona: null, loading: false, error: null });
      return;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const persona = await fetchUserPersona(userId);
      savePersonaToCache(persona);
      setState({ persona, loading: false, error: null });
    } catch (err) {
      console.error("Failed to fetch persona:", err);
      setState({
        persona: null,
        loading: false,
        error: err instanceof Error ? err.message : "Failed to fetch persona",
      });
    }
  }, [userId, savePersonaToCache]);

  // Refresh persona (force fetch)
  const refreshPersona = useCallback(async () => {
    clearPersona();
    await fetchPersona();
  }, [clearPersona, fetchPersona]);

  // Initial load
  useEffect(() => {
    if (!userId) {
      setState({ persona: null, loading: false, error: null });
      return;
    }

    // Try to load from cache first
    const cached = loadCachedPersona();
    if (cached) {
      setState({ persona: cached, loading: false, error: null });
      return;
    }

    // Fetch from API
    fetchPersona();
  }, [userId, loadCachedPersona, fetchPersona]);

  return {
    ...state,
    refreshPersona,
    clearPersona,
  };
}

/**
 * Hook untuk mengecek apakah popup recommendation harus ditampilkan
 * 
 * Logic:
 * - Tampilkan 1x per session (menggunakan sessionStorage)
 * - Hanya tampilkan jika ada recommendations
 */
export function useRecommendationTrigger(persona: PredictPersonaResponse | null) {
  const [shouldShow, setShouldShow] = useState(false);
  const [dismissed, setDismissed] = useState(false);

  const SESSION_KEY = "ngasihtau_recommendation_shown";

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (!persona || !persona.recommendations?.length) return;
    if (dismissed) return;

    // Check if already shown this session
    const alreadyShown = sessionStorage.getItem(SESSION_KEY);
    if (alreadyShown) {
      setShouldShow(false);
      return;
    }

    // Show popup
    setShouldShow(true);
  }, [persona, dismissed]);

  const dismiss = useCallback(() => {
    if (typeof window === "undefined") return;
    
    sessionStorage.setItem(SESSION_KEY, "true");
    setShouldShow(false);
    setDismissed(true);
  }, []);

  const reset = useCallback(() => {
    if (typeof window === "undefined") return;
    
    sessionStorage.removeItem(SESSION_KEY);
    setDismissed(false);
  }, []);

  return {
    shouldShow,
    dismiss,
    reset,
  };
}
