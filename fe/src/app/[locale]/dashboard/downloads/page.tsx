"use client";

import { useState } from "react";
import { useOfflineMaterials } from "@/hooks/useOffline";
import { ProtectedRoute } from "@/components/auth";
import Link from "next/link";
import {
  Download,
  FileText,
  Trash2,
  Smartphone,
  Monitor,
  RefreshCw,
  Shield,
  Clock,
  AlertCircle,
  CheckCircle,
  Loader,
} from "lucide-react";
import { Device } from "@/lib/api/offline";

// Device icon based on platform
function DeviceIcon({ platform }: { platform: string }) {
  if (platform === "desktop") {
    return <Monitor size={16} />;
  }
  return <Smartphone size={16} />;
}

// License status badge
function LicenseStatusBadge({ status }: { status: string }) {
  const statusConfig = {
    active: { color: "bg-green-100 text-green-700", icon: CheckCircle, label: "Active" },
    expired: { color: "bg-yellow-100 text-yellow-700", icon: Clock, label: "Expired" },
    revoked: { color: "bg-red-100 text-red-700", icon: AlertCircle, label: "Revoked" },
  };

  const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.expired;
  const Icon = config.icon;

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-bold ${config.color}`}>
      <Icon size={12} />
      {config.label}
    </span>
  );
}

// Device card component
function DeviceCard({
  device,
  isCurrentDevice,
  onDeregister,
  isDeregistering,
}: {
  device: Device;
  isCurrentDevice: boolean;
  onDeregister: () => void;
  isDeregistering: boolean;
}) {
  return (
    <div
      className={`p-4 rounded-lg border-2 ${
        isCurrentDevice
          ? "border-[#FF8811] bg-orange-50"
          : "border-[#2B2D42] bg-white"
      } shadow-[3px_3px_0px_0px_${isCurrentDevice ? "#FF8811" : "#2B2D42"}]`}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg ${isCurrentDevice ? "bg-[#FF8811] text-white" : "bg-gray-100 text-gray-600"}`}>
            <DeviceIcon platform={device.platform} />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-bold text-[#2B2D42]">{device.name}</h3>
              {isCurrentDevice && (
                <span className="text-[10px] px-1.5 py-0.5 bg-[#FF8811] text-white rounded font-bold">
                  This Device
                </span>
              )}
            </div>
            <p className="text-xs text-gray-500 capitalize">{device.platform}</p>
          </div>
        </div>

        {!isCurrentDevice && (
          <button
            onClick={onDeregister}
            disabled={isDeregistering}
            className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
            title="Remove device"
          >
            {isDeregistering ? (
              <Loader size={16} className="animate-spin" />
            ) : (
              <Trash2 size={16} />
            )}
          </button>
        )}
      </div>

      <div className="mt-3 flex items-center gap-4 text-xs text-gray-500">
        <span>Last used: {new Date(device.last_used_at).toLocaleDateString()}</span>
        <span className={`px-2 py-0.5 rounded-full ${
          device.status === "active" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
        }`}>
          {device.status}
        </span>
      </div>
    </div>
  );
}

export default function DownloadedMaterialsPage() {
  const {
    downloads,
    devices,
    device: currentDevice,
    isRegistered,
    loading,
    error,
    registerDevice,
    fetchDevices,
    removeDownload,
    getLicense,
    renewLicense,
  } = useOfflineMaterials();

  const [isRegistering, setIsRegistering] = useState(false);
  const [isFetchingDevices, setIsFetchingDevices] = useState(false);
  const [renewingLicenseId, setRenewingLicenseId] = useState<string | null>(null);
  const [deregisteringDeviceId, setDeregisteringDeviceId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"downloads" | "devices">("downloads");

  const handleRegisterDevice = async () => {
    setIsRegistering(true);
    await registerDevice();
    setIsRegistering(false);
  };

  const handleFetchDevices = async () => {
    setIsFetchingDevices(true);
    await fetchDevices();
    setIsFetchingDevices(false);
  };

  const handleRenewLicense = async (licenseId: string) => {
    setRenewingLicenseId(licenseId);
    await renewLicense(licenseId);
    setRenewingLicenseId(null);
  };

  return (
    <ProtectedRoute>
      <div className="p-4 sm:p-6 lg:p-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-[#2B2D42]">Offline Materials</h1>

          {/* Device Registration Status */}
          {isRegistered ? (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-green-100 text-green-700 rounded-lg border border-green-300">
              <Shield size={16} />
              <span className="text-sm font-bold">Device Registered</span>
            </div>
          ) : (
            <button
              onClick={handleRegisterDevice}
              disabled={isRegistering || loading}
              className="flex items-center gap-2 px-4 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f] transition-all shadow-[2px_2px_0px_0px_#2B2D42] disabled:opacity-50"
            >
              {isRegistering ? (
                <Loader size={16} className="animate-spin" />
              ) : (
                <Shield size={16} />
              )}
              Register Device
            </button>
          )}
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-100 border-2 border-red-300 rounded-lg text-red-700 text-sm">
            <div className="flex items-center gap-2">
              <AlertCircle size={16} />
              {error}
            </div>
          </div>
        )}

        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          <button
            onClick={() => setActiveTab("downloads")}
            className={`px-4 py-2 rounded-lg font-bold transition-all ${
              activeTab === "downloads"
                ? "bg-[#FF8811] text-white shadow-[2px_2px_0px_0px_#2B2D42]"
                : "bg-white text-[#2B2D42] border-2 border-[#2B2D42] hover:bg-gray-50"
            }`}
          >
            <div className="flex items-center gap-2">
              <Download size={16} />
              Downloads ({downloads.length})
            </div>
          </button>
          <button
            onClick={() => {
              setActiveTab("devices");
              handleFetchDevices();
            }}
            className={`px-4 py-2 rounded-lg font-bold transition-all ${
              activeTab === "devices"
                ? "bg-[#FF8811] text-white shadow-[2px_2px_0px_0px_#2B2D42]"
                : "bg-white text-[#2B2D42] border-2 border-[#2B2D42] hover:bg-gray-50"
            }`}
          >
            <div className="flex items-center gap-2">
              <Smartphone size={16} />
              Devices ({devices.length})
            </div>
          </button>
        </div>

        {/* Downloads Tab */}
        {activeTab === "downloads" && (
          <>
            {downloads.length === 0 ? (
              <div className="flex flex-col items-center justify-center p-12 bg-white border-2 border-[#2B2D42] rounded-lg shadow-[4px_4px_0px_0px_#2B2D42]">
                <Download size={48} className="text-[#FF8811] mb-4" />
                <p className="text-lg font-bold text-[#2B2D42]">No materials downloaded yet</p>
                <p className="text-sm text-gray-500 mb-6">
                  {isRegistered
                    ? "Materials you download will appear here for offline access."
                    : "Register your device first to download materials for offline access."}
                </p>
                <Link
                  href="/dashboard/pod"
                  className="px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f] transition-all shadow-[2px_2px_0px_0px_#2B2D42]"
                >
                  Browse Materials
                </Link>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {downloads.map((download) => {
                  const license = getLicense(download.materialId);
                  const isExpired = license && new Date(license.expires_at) < new Date();

                  return (
                    <div
                      key={download.materialId}
                      className="bg-white border-2 border-[#2B2D42] p-4 rounded-lg shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_0px_#2B2D42] transition-all"
                    >
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <div className="p-2 bg-gray-100 rounded-lg border border-[#2B2D42]">
                            <FileText size={20} className="text-[#FF8811]" />
                          </div>
                          <div>
                            <h3 className="font-bold text-[#2B2D42] line-clamp-1" title={download.materialTitle}>
                              {download.materialTitle}
                            </h3>
                            <p className="text-xs text-gray-500">
                              {download.manifest.file_type.toUpperCase()} •{" "}
                              {(download.manifest.total_size / 1024 / 1024).toFixed(2)} MB
                            </p>
                          </div>
                        </div>
                      </div>

                      {/* License Status */}
                      {license && (
                        <div className="flex items-center justify-between mt-3 mb-2">
                          <LicenseStatusBadge status={license.status} />
                          {isExpired && license.status === "active" && (
                            <button
                              onClick={() => handleRenewLicense(license.id)}
                              disabled={renewingLicenseId === license.id}
                              className="flex items-center gap-1 text-xs font-bold text-[#FF8811] hover:underline disabled:opacity-50"
                            >
                              {renewingLicenseId === license.id ? (
                                <Loader size={12} className="animate-spin" />
                              ) : (
                                <RefreshCw size={12} />
                              )}
                              Renew
                            </button>
                          )}
                        </div>
                      )}

                      <div className="flex items-center justify-between mt-4">
                        <Link
                          href={`/dashboard/pod/${download.materialId}`}
                          className="text-xs font-bold text-[#FF8811] hover:underline"
                        >
                          View Details
                        </Link>
                        <button
                          onClick={() => removeDownload(download.materialId)}
                          className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                          title="Remove from downloads"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>

                      <div className="mt-2 flex items-center justify-between text-[10px] text-gray-400">
                        <span>Downloaded {new Date(download.downloadedAt).toLocaleDateString()}</span>
                        {license && (
                          <span>Expires {new Date(license.expires_at).toLocaleDateString()}</span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </>
        )}

        {/* Devices Tab */}
        {activeTab === "devices" && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-500">
                You can register up to 5 devices for offline access.
              </p>
              <button
                onClick={handleFetchDevices}
                disabled={isFetchingDevices}
                className="flex items-center gap-2 px-3 py-1.5 text-sm font-bold text-[#2B2D42] hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
              >
                {isFetchingDevices ? (
                  <Loader size={14} className="animate-spin" />
                ) : (
                  <RefreshCw size={14} />
                )}
                Refresh
              </button>
            </div>

            {devices.length === 0 ? (
              <div className="flex flex-col items-center justify-center p-12 bg-white border-2 border-[#2B2D42] rounded-lg shadow-[4px_4px_0px_0px_#2B2D42]">
                <Smartphone size={48} className="text-[#FF8811] mb-4" />
                <p className="text-lg font-bold text-[#2B2D42]">No devices registered</p>
                <p className="text-sm text-gray-500 mb-6">
                  Register this device to start downloading materials for offline access.
                </p>
                {!isRegistered && (
                  <button
                    onClick={handleRegisterDevice}
                    disabled={isRegistering}
                    className="px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f] transition-all shadow-[2px_2px_0px_0px_#2B2D42] disabled:opacity-50"
                  >
                    {isRegistering ? (
                      <Loader size={16} className="animate-spin" />
                    ) : (
                      "Register This Device"
                    )}
                  </button>
                )}
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {devices.map((device) => (
                  <DeviceCard
                    key={device.id}
                    device={device}
                    isCurrentDevice={currentDevice?.id === device.id}
                    onDeregister={() => {
                      setDeregisteringDeviceId(device.id);
                      // TODO: Implement deregister
                      setDeregisteringDeviceId(null);
                    }}
                    isDeregistering={deregisteringDeviceId === device.id}
                  />
                ))}
              </div>
            )}

            {/* Device Limit Info */}
            <div className="p-4 bg-blue-50 border-2 border-blue-200 rounded-lg">
              <div className="flex items-start gap-3">
                <Shield size={20} className="text-blue-600 flex-shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-bold text-blue-800">Device Security</h4>
                  <p className="text-sm text-blue-700 mt-1">
                    Each device is uniquely identified to protect your offline content.
                    Removing a device will revoke all licenses associated with it.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </ProtectedRoute>
  );
}
