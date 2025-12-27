import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAuth } from "../lib/auth";
import { api } from "../lib/api";
import { LayoutDashboard, Loader2 } from "lucide-react";

export const Route = createFileRoute("/login")({
  component: LoginComponent,
});

function LoginComponent() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errorString, setErrorString] = useState("");
  const { login } = useAuth();
  const navigate = useNavigate();

  const { data: setupStatus } = useQuery({
    queryKey: ["setup-status"],
    queryFn: api.setupStatus,
    retry: false,
  });

  if (setupStatus?.needs_setup) {
    navigate({ to: "/setup" });
    return null;
  }

  const loginMutation = useMutation({
    mutationFn: async () => {
      return api.login(email, password);
    },
    onSuccess: (data) => {
      login(data.token, data.user);
      navigate({ to: "/" });
    },
    onError: (err: any) => {
      setErrorString(err.message || "Failed to login");
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorString("");
    loginMutation.mutate();
  };

  return (
    <div className="min-h-screen bg-[#09090b] flex items-center justify-center p-4 font-sans text-zinc-200 selection:bg-emerald-500/30">
      <div className="w-full max-w-sm">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center p-3 bg-emerald-500 text-black mb-4">
            <LayoutDashboard className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Welcome back</h1>
          <p className="text-sm text-zinc-400 mt-2">
            Enter your credentials to access the admin panel
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label
              htmlFor="email"
              className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
              Email
            </label>
            <input
              id="email"
              type="email"
              placeholder="name@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b] disabled:cursor-not-allowed disabled:opacity-50"
              required
            />
          </div>
          <div className="space-y-2">
            <label
              htmlFor="password"
              className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b] disabled:cursor-not-allowed disabled:opacity-50"
              required
            />
          </div>

          {errorString && (
            <div className="p-3 text-sm text-red-400 bg-red-900/10 border border-red-900/20">
              {errorString}
            </div>
          )}

          <button
            type="submit"
            disabled={loginMutation.isPending}
            className="inline-flex items-center justify-center w-full h-10 px-4 py-2 text-sm font-medium text-black transition-colors bg-emerald-500 hover:bg-emerald-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b] disabled:opacity-50 disabled:pointer-events-none"
          >
            {loginMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : "Sign In"}
          </button>
        </form>
      </div>
    </div>
  );
}
