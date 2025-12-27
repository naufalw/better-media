import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";
import { LayoutDashboard, Loader2 } from "lucide-react";

export const Route = createFileRoute("/setup")({
  component: SetupPage,
});

function SetupPage() {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
  });
  const { login } = useAuth();
  const navigate = useNavigate();
  const [errorString, setErrorString] = useState("");

  const { data: status, isLoading: isChecking } = useQuery({
    queryKey: ["setup-status"],
    queryFn: api.setupStatus,
    retry: false,
  });

  const setupMutation = useMutation({
    mutationFn: api.setupAdmin,
    onSuccess: (data) => {
      login(data.token, data.user);
      navigate({ to: "/" });
    },
    onError: (err: any) => {
      setErrorString(err.message || "Setup failed");
    },
  });

  if (isChecking) {
    return (
      <div className="min-h-screen bg-[#09090b] flex items-center justify-center text-zinc-200">
        <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
      </div>
    );
  }

  if (status && !status.needs_setup) {
    // Already setup, redirect to login
    // We need to wait for render to use navigate effectively or useEffect
    // simpler: just render duplicate login redirect?
    // Better: use useEffect or handling it in component logic
    // Let's use useEffect to avoid "render verify" issues but since we are modifying invalid return paths...
    // We can just return null and navigate
    navigate({ to: "/login" });
    return null;
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setErrorString("");
    setupMutation.mutate(formData);
  };

  return (
    <div className="min-h-screen bg-[#09090b] flex items-center justify-center p-4 font-sans text-zinc-200 selection:bg-emerald-500/30">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center p-3 bg-emerald-500 text-black mb-4">
            <LayoutDashboard className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Setup Admin</h1>
          <p className="text-sm text-zinc-400 mt-2">Create the first administrator account</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium leading-none">Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b]"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium leading-none">Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b]"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium leading-none">Password</label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b]"
              required
              minLength={8}
            />
          </div>

          {errorString && (
            <div className="p-3 text-sm text-red-400 bg-red-900/10 border border-red-900/20">
              {errorString}
            </div>
          )}

          <button
            type="submit"
            disabled={setupMutation.isPending}
            className="inline-flex items-center justify-center w-full h-10 px-4 py-2 text-sm font-medium text-black transition-colors bg-emerald-500 hover:bg-emerald-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b] disabled:opacity-50"
          >
            {setupMutation.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              "Create Admin Account"
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
