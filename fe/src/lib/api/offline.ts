import apiClient from "@/lib/api-client";

/**
 * Offline Material API Client
 * Handles device registration, license management, and material downloads
 */

// ============ Types ============

export interface Device {
  id: string;
  user_id: string;
  fingerprint: string;
  name: string;
  platform: "ios" | "android" | "desktop";
  status: "active" | "revoked";
  last_used_at: string;
  created_at: string;
}

export interface License {
  id: string;
  user_id: string;
  material_id: string;
  device_id: string;
  status: "active" | "expired" | "revoked";
  expires_at: string;
  offline_grace_period: string;
  last_validated_at: string;
  nonce: string;
  created_at: string;
}

export interface DownloadChunk {
  index: number;
  offset: number;
  size: number;
}

export interface DownloadManifest {
  material_id: string;
  license_id: string;
  total_chunks: number;
  total_size: number;
  original_hash: string;
  encrypted_hash: string;
  chunk_size: number;
  chunks: DownloadChunk[];
  file_type: string;
  created_at: string;
}

export interface DownloadResponse {
  manifest: DownloadManifest;
  download_url: string;
  expires_at: string;
}

export interface RegisterDeviceInput {
  fingerprint: string;
  name: string;
  platform: "ios" | "android" | "desktop";
}

export interface IssueLicenseInput {
  device_id: string;
  fingerprint: string;
}

export interface ValidateLicenseInput {
  fingerprint: string;
  nonce: string;
}

export interface RenewLicenseInput {
  fingerprint: string;
}

// ============ Device API ============

/**
 * Register a new device for offline access
 * Maximum 5 devices per user
 */
export async function registerDevice(input: RegisterDeviceInput): Promise<Device> {
  const response = await apiClient.post<{ data: Device }>(
    "/api/v1/offline/devices",
    input
  );
  return response.data.data;
}

/**
 * List all registered devices for the current user
 */
export async function listDevices(): Promise<Device[]> {
  const response = await apiClient.get<{ data: Device[] }>(
    "/api/v1/offline/devices"
  );
  return response.data.data;
}

/**
 * Deregister a device and revoke all associated licenses
 */
export async function deregisterDevice(deviceId: string): Promise<void> {
  await apiClient.delete(`/api/v1/offline/devices/${deviceId}`);
}

// ============ License API ============

/**
 * Issue a license for offline access to a material
 */
export async function issueLicense(
  materialId: string,
  input: IssueLicenseInput
): Promise<License> {
  const response = await apiClient.post<{ data: License }>(
    `/api/v1/offline/materials/${materialId}/license`,
    input
  );
  return response.data.data;
}

/**
 * Validate a license (online check)
 * Updates last_validated_at timestamp
 */
export async function validateLicense(
  licenseId: string,
  input: ValidateLicenseInput
): Promise<License> {
  const response = await apiClient.post<{ data: License }>(
    `/api/v1/offline/licenses/${licenseId}/validate`,
    input
  );
  return response.data.data;
}

/**
 * Renew a license before it expires
 */
export async function renewLicense(
  licenseId: string,
  input: RenewLicenseInput
): Promise<License> {
  const response = await apiClient.post<{ data: License }>(
    `/api/v1/offline/licenses/${licenseId}/renew`,
    input
  );
  return response.data.data;
}

// ============ Download API ============

/**
 * Get download manifest and URL for encrypted material
 */
export async function downloadMaterial(
  materialId: string,
  deviceId: string,
  fingerprint: string,
  licenseId: string
): Promise<DownloadResponse> {
  const params = new URLSearchParams({
    device_id: deviceId,
    fingerprint: fingerprint,
    license_id: licenseId,
  });

  const response = await apiClient.get<{ data: DownloadResponse }>(
    `/api/v1/offline/materials/${materialId}/download?${params.toString()}`
  );
  return response.data.data;
}

// ============ Device Fingerprint ============

/**
 * Generate a device fingerprint based on browser/device characteristics
 * This is a simplified version - in production, use a library like FingerprintJS
 */
export async function generateDeviceFingerprint(): Promise<string> {
  const components: string[] = [];

  // User agent
  components.push(navigator.userAgent);

  // Screen info
  components.push(`${screen.width}x${screen.height}x${screen.colorDepth}`);

  // Timezone
  components.push(Intl.DateTimeFormat().resolvedOptions().timeZone);

  // Language
  components.push(navigator.language);

  // Platform
  components.push(navigator.platform);

  // Hardware concurrency
  components.push(String(navigator.hardwareConcurrency || 0));

  // Device memory (if available)
  const nav = navigator as Navigator & { deviceMemory?: number };
  if (nav.deviceMemory) {
    components.push(String(nav.deviceMemory));
  }

  // Canvas fingerprint
  try {
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    if (ctx) {
      ctx.textBaseline = "top";
      ctx.font = "14px Arial";
      ctx.fillText("NgasihTau Fingerprint", 2, 2);
      components.push(canvas.toDataURL());
    }
  } catch {
    // Canvas not available
  }

  // Create hash from components
  const data = components.join("|");
  const encoder = new TextEncoder();
  const dataBuffer = encoder.encode(data);

  // Use SubtleCrypto for hashing
  const hashBuffer = await crypto.subtle.digest("SHA-256", dataBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");

  return hashHex;
}

/**
 * Detect the current platform
 */
export function detectPlatform(): "ios" | "android" | "desktop" {
  const userAgent = navigator.userAgent.toLowerCase();

  if (/iphone|ipad|ipod/.test(userAgent)) {
    return "ios";
  }

  if (/android/.test(userAgent)) {
    return "android";
  }

  return "desktop";
}

/**
 * Get device name based on platform and browser
 */
export function getDeviceName(): string {
  const platform = detectPlatform();
  const browser = detectBrowser();

  if (platform === "desktop") {
    return `${browser} on ${getOS()}`;
  }

  return `${browser} on ${platform === "ios" ? "iOS" : "Android"}`;
}

function detectBrowser(): string {
  const userAgent = navigator.userAgent;

  if (userAgent.includes("Chrome")) return "Chrome";
  if (userAgent.includes("Firefox")) return "Firefox";
  if (userAgent.includes("Safari")) return "Safari";
  if (userAgent.includes("Edge")) return "Edge";
  if (userAgent.includes("Opera")) return "Opera";

  return "Browser";
}

function getOS(): string {
  const platform = navigator.platform.toLowerCase();

  if (platform.includes("win")) return "Windows";
  if (platform.includes("mac")) return "macOS";
  if (platform.includes("linux")) return "Linux";

  return "Desktop";
}
