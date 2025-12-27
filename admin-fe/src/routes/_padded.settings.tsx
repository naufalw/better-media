import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { api } from "@/lib/api";
import { Loader2, KeyRound } from "lucide-react";

export const Route = createFileRoute("/_padded/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  const [formData, setFormData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [errorFragment, setErrorFragment] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (formData.newPassword !== formData.confirmPassword) {
      setErrorFragment("New passwords do not match");
      setStatus("error");
      return;
    }

    setStatus("loading");
    setErrorFragment("");

    try {
      await api.changePassword({
        current_password: formData.currentPassword,
        new_password: formData.newPassword,
      });
      setStatus("success");
      setFormData({ currentPassword: "", newPassword: "", confirmPassword: "" });
    } catch (err: any) {
      setErrorFragment(err.message || "Failed to update password");
      setStatus("error");
    }
  };

  return (
    <div className="max-w-xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-white tracking-tight">Settings</h1>
        <p className="text-sm text-zinc-400 mt-1">Manage your account preferences</p>
      </div>

      <div className="p-6 border border-zinc-800 bg-[#0c0c0e]">
        <div className="flex items-center gap-3 mb-6">
          <div className="p-2 bg-emerald-500/10 text-emerald-500">
            <KeyRound className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-lg font-medium text-white">Change Password</h2>
            <p className="text-sm text-zinc-400">Update your access credentials</p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-zinc-400">Current Password</label>
            <input
              type="password"
              value={formData.currentPassword}
              onChange={(e) => setFormData({ ...formData, currentPassword: e.target.value })}
              className="flex h-10 w-full bg-[#09090b] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-zinc-400">New Password</label>
            <input
              type="password"
              value={formData.newPassword}
              onChange={(e) => setFormData({ ...formData, newPassword: e.target.value })}
              className="flex h-10 w-full bg-[#09090b] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              required
              minLength={8}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-zinc-400">Confirm New Password</label>
            <input
              type="password"
              value={formData.confirmPassword}
              onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
              className="flex h-10 w-full bg-[#09090b] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              required
              minLength={8}
            />
          </div>

          {status === "error" && (
            <div className="p-3 text-sm text-red-400 bg-red-900/10 border border-red-900/20">
              {errorFragment}
            </div>
          )}

          {status === "success" && (
            <div className="p-3 text-sm text-emerald-400 bg-emerald-900/10 border border-emerald-900/20">
              Password updated successfully
            </div>
          )}

          <div className="pt-2">
            <button
              type="submit"
              disabled={status === "loading"}
              className="px-4 py-2 text-sm font-medium text-black bg-emerald-500 hover:bg-emerald-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:opacity-50"
            >
              {status === "loading" ? (
                <div className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Updating...
                </div>
              ) : (
                "Update Password"
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
