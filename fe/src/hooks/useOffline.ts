"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Device,
  License,
  DownloadResponse,
  registerDevice,
  listDevices,
  deregisterDevice,
  issueLicense,
  validateLicense,
  renewLicense,
  downloadMaterial,
  generateDeviceFingerprint,
  detectPlatform,
  getDeviceName,
} from "@/lib/api/offline";

// ============ Storage Keys ============
const DEVICE_STORAGE_KEY = "ngasihtau_offline_device";
const FINGERPRINT_STORAGE_KEY = "ngasihtau_device_fingerprint";
const LICENSES_STORAGE_KEY = "ngasihtau_offline_licenses";
const DOWNLOADS_STORAGE_KEY = "ngasihtau_offline_downloads";

// ============ Types ============
export interface OfflineDownload {
  materialId: string;
  materialTitle: string;
  licenseId: string;
  downloadUrl: string;
  manifest: DownloadResponse["manifest"];
  downloadedAt: string;
  expiresAt: string;
}

interface DeviceState {
  device: Device | null;
  fingerprint: string | null;
  loading: boolean;
  error: string | null;
}

interface OfflineState {
  devices: Device[];
  licenses: License[];
  downloads: OfflineDownload[];
  loading: boolean;
  error: string | null;
}

// ============ Device Hook ============

/**
 * Hook untuk mengelola device registration
 * Automatically registers device on first use
 */
export function useDevice() {
  const [state, setState] = useState<DeviceState>({
    device: null,
    fingerprint: null,
    loading: true,
    error: null,
  });

  // Load cached device and fingerprint
  const loadCachedDevice = useCallback(async () => {
    if (typeof window === "undefined") return;

    try {
      // Get or generate fingerprint
      let fingerprint = localStorage.getItem(FINGERPRINT_STORAGE_KEY);
      if (!fingerprint) {
        fingerprint = await generateDeviceFingerprint();
        localStorage.setItem(FINGERPRINT_STORAGE_KEY, fingerprint);
      }

      // Get cached device
      const cachedDevice = localStorage.getItem(DEVICE_STORAGE_KEY);
      if (cachedDevice) {
        const device = JSON.parse(cachedDevice) as Device;
        setState({
          device,
          fingerprint,
          loading: false,
          error: null,
        });
        return;
      }

      setState((prev) => ({
        ...prev,
        fingerprint,
        loading: false,
      }));
    } catch (err) {
      console.error("Failed to load cached device:", err);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: "Failed to load device",
      }));
    }
  }, []);

  // Register device
  const register = useCallback(async () => {
    if (!state.fingerprint) return null;

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const device = await registerDevice({
        fingerprint: state.fingerprint,
        name: getDeviceName(),
        platform: detectPlatform(),
      });

      localStorage.setItem(DEVICE_STORAGE_KEY, JSON.stringify(device));
      setState({
        device,
        fingerprint: state.fingerprint,
        loading: false,
        error: null,
      });

      return device;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to register device";
      setState((prev) => ({
        ...prev,
        loading: false,
        error: message,
      }));
      return null;
    }
  }, [state.fingerprint]);

  // Deregister device
  const deregister = useCallback(async () => {
    if (!state.device) return;

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      await deregisterDevice(state.device.id);
      localStorage.removeItem(DEVICE_STORAGE_KEY);
      setState((prev) => ({
        ...prev,
        device: null,
        loading: false,
      }));
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to deregister device";
      setState((prev) => ({
        ...prev,
        loading: false,
        error: message,
      }));
    }
  }, [state.device]);

  // Initial load
  useEffect(() => {
    loadCachedDevice();
  }, [loadCachedDevice]);

  return {
    ...state,
    register,
    deregister,
    isRegistered: !!state.device,
  };
}

// ============ Offline Materials Hook ============

/**
 * Hook untuk mengelola offline materials (licenses & downloads)
 */
export function useOfflineMaterials() {
  const { device, fingerprint, isRegistered, register } = useDevice();
  const [state, setState] = useState<OfflineState>({
    devices: [],
    licenses: [],
    downloads: [],
    loading: true,
    error: null,
  });

  // Load cached data
  const loadCachedData = useCallback(() => {
    if (typeof window === "undefined") return;

    try {
      const cachedLicenses = localStorage.getItem(LICENSES_STORAGE_KEY);
      const cachedDownloads = localStorage.getItem(DOWNLOADS_STORAGE_KEY);

      setState((prev) => ({
        ...prev,
        licenses: cachedLicenses ? JSON.parse(cachedLicenses) : [],
        downloads: cachedDownloads ? JSON.parse(cachedDownloads) : [],
        loading: false,
      }));
    } catch (err) {
      console.error("Failed to load cached offline data:", err);
      setState((prev) => ({ ...prev, loading: false }));
    }
  }, []);

  // Fetch devices from server
  const fetchDevices = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const devices = await listDevices();
      setState((prev) => ({
        ...prev,
        devices,
        loading: false,
      }));
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to fetch devices";
      setState((prev) => ({
        ...prev,
        loading: false,
        error: message,
      }));
    }
  }, []);

  // Request license for a material
  const requestLicense = useCallback(
    async (materialId: string): Promise<License | null> => {
      if (!device || !fingerprint) {
        // Try to register device first
        const newDevice = await register();
        if (!newDevice) {
          setState((prev) => ({
            ...prev,
            error: "Please register your device first",
          }));
          return null;
        }
      }

      const currentDevice = device;
      const currentFingerprint = fingerprint;

      if (!currentDevice || !currentFingerprint) {
        setState((prev) => ({
          ...prev,
          error: "Device not registered",
        }));
        return null;
      }

      setState((prev) => ({ ...prev, loading: true, error: null }));

      try {
        const license = await issueLicense(materialId, {
          device_id: currentDevice.id,
          fingerprint: currentFingerprint,
        });

        // Update cached licenses
        const updatedLicenses = [...state.licenses.filter((l) => l.material_id !== materialId), license];
        localStorage.setItem(LICENSES_STORAGE_KEY, JSON.stringify(updatedLicenses));

        setState((prev) => ({
          ...prev,
          licenses: updatedLicenses,
          loading: false,
        }));

        return license;
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to request license";
        setState((prev) => ({
          ...prev,
          loading: false,
          error: message,
        }));
        return null;
      }
    },
    [device, fingerprint, register, state.licenses]
  );

  // Download material for offline access
  const downloadForOffline = useCallback(
    async (materialId: string, materialTitle: string): Promise<OfflineDownload | null> => {
      if (!device || !fingerprint) {
        setState((prev) => ({
          ...prev,
          error: "Device not registered",
        }));
        return null;
      }

      setState((prev) => ({ ...prev, loading: true, error: null }));

      try {
        // First, get or request license
        let license = state.licenses.find((l) => l.material_id === materialId);
        if (!license || license.status !== "active") {
          license = await requestLicense(materialId);
          if (!license) return null;
        }

        // Download material
        const downloadResponse = await downloadMaterial(
          materialId,
          device.id,
          fingerprint,
          license.id
        );

        const offlineDownload: OfflineDownload = {
          materialId,
          materialTitle,
          licenseId: license.id,
          downloadUrl: downloadResponse.download_url,
          manifest: downloadResponse.manifest,
          downloadedAt: new Date().toISOString(),
          expiresAt: downloadResponse.expires_at,
        };

        // Update cached downloads
        const updatedDownloads = [
          ...state.downloads.filter((d) => d.materialId !== materialId),
          offlineDownload,
        ];
        localStorage.setItem(DOWNLOADS_STORAGE_KEY, JSON.stringify(updatedDownloads));

        setState((prev) => ({
          ...prev,
          downloads: updatedDownloads,
          loading: false,
        }));

        return offlineDownload;
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to download material";
        setState((prev) => ({
          ...prev,
          loading: false,
          error: message,
        }));
        return null;
      }
    },
    [device, fingerprint, state.licenses, state.downloads, requestLicense]
  );

  // Remove offline download
  const removeDownload = useCallback(
    (materialId: string) => {
      const updatedDownloads = state.downloads.filter((d) => d.materialId !== materialId);
      localStorage.setItem(DOWNLOADS_STORAGE_KEY, JSON.stringify(updatedDownloads));

      setState((prev) => ({
        ...prev,
        downloads: updatedDownloads,
      }));
    },
    [state.downloads]
  );

  // Check if material is downloaded
  const isDownloaded = useCallback(
    (materialId: string): boolean => {
      return state.downloads.some((d) => d.materialId === materialId);
    },
    [state.downloads]
  );

  // Get license for material
  const getLicense = useCallback(
    (materialId: string): License | undefined => {
      return state.licenses.find((l) => l.material_id === materialId);
    },
    [state.licenses]
  );

  // Validate license
  const validateMaterialLicense = useCallback(
    async (licenseId: string, nonce: string): Promise<License | null> => {
      if (!fingerprint) return null;

      try {
        const license = await validateLicense(licenseId, {
          fingerprint,
          nonce,
        });

        // Update cached license
        const updatedLicenses = state.licenses.map((l) =>
          l.id === licenseId ? license : l
        );
        localStorage.setItem(LICENSES_STORAGE_KEY, JSON.stringify(updatedLicenses));

        setState((prev) => ({
          ...prev,
          licenses: updatedLicenses,
        }));

        return license;
      } catch (err) {
        console.error("Failed to validate license:", err);
        return null;
      }
    },
    [fingerprint, state.licenses]
  );

  // Renew license
  const renewMaterialLicense = useCallback(
    async (licenseId: string): Promise<License | null> => {
      if (!fingerprint) return null;

      try {
        const license = await renewLicense(licenseId, {
          fingerprint,
        });

        // Update cached license
        const updatedLicenses = state.licenses.map((l) =>
          l.id === licenseId ? license : l
        );
        localStorage.setItem(LICENSES_STORAGE_KEY, JSON.stringify(updatedLicenses));

        setState((prev) => ({
          ...prev,
          licenses: updatedLicenses,
        }));

        return license;
      } catch (err) {
        console.error("Failed to renew license:", err);
        return null;
      }
    },
    [fingerprint, state.licenses]
  );

  // Initial load
  useEffect(() => {
    loadCachedData();
  }, [loadCachedData]);

  return {
    ...state,
    device,
    fingerprint,
    isRegistered,
    registerDevice: register,
    fetchDevices,
    requestLicense,
    downloadForOffline,
    removeDownload,
    isDownloaded,
    getLicense,
    validateLicense: validateMaterialLicense,
    renewLicense: renewMaterialLicense,
  };
}
