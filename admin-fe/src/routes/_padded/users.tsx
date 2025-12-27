import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { Loader2, Trash2, UserPlus, Shield, User } from "lucide-react";
import { useState } from "react";
import { useAuth } from "../../lib/auth";

export const Route = createFileRoute("/_padded/users")({
  component: UsersPage,
});

function UsersPage() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
  });
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["users"],
    queryFn: api.listUsers,
  });

  const createMutation = useMutation({
    mutationFn: api.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setIsCreateOpen(false);
      setFormData({ name: "", email: "", password: "" });
      setError("");
    },
    onError: (err: any) => {
      setError(err.message || "Failed to create user");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  if (isLoading) {
    return <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">User Management</h1>
          <p className="text-sm text-zinc-400 mt-1">Manage system access and roles</p>
        </div>
        <button
          onClick={() => setIsCreateOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-black bg-emerald-500 hover:bg-emerald-400 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
        >
          <UserPlus className="w-4 h-4" />
          Add User
        </button>
      </div>

      {isCreateOpen && (
        <div className="p-4 border border-zinc-800 bg-zinc-900/50 space-y-4">
          <h2 className="text-lg font-medium text-white">Create New User</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-400">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-400">Email</label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-400">Password</label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  required
                  minLength={8}
                />
              </div>
            </div>
            {error && <div className="text-sm text-red-400">{error}</div>}
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={createMutation.isPending}
                className="px-4 py-2 text-sm font-medium text-black bg-emerald-500 hover:bg-emerald-400"
              >
                {createMutation.isPending ? "Creating..." : "Create User"}
              </button>
              <button
                type="button"
                onClick={() => setIsCreateOpen(false)}
                className="px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white bg-transparent border border-zinc-800 hover:bg-zinc-800"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Users Table */}
      <div className="border border-zinc-800 overflow-hidden">
        <table className="w-full text-sm text-left">
          <thead className="text-xs text-zinc-400 uppercase bg-zinc-900/50 border-b border-zinc-800">
            <tr>
              <th className="px-6 py-3">User</th>
              <th className="px-6 py-3">Role</th>
              <th className="px-6 py-3">Created</th>
              <th className="px-6 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800">
            {data?.users?.map((u: any) => (
              <tr key={u.id} className="hover:bg-zinc-900/50 transition-colors">
                <td className="px-6 py-4">
                  <div className="flex flex-col">
                    <span className="font-medium text-white">{u.name}</span>
                    <span className="text-zinc-500">{u.email}</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    {u.role === "admin" ? (
                      <Shield className="w-4 h-4 text-emerald-500" />
                    ) : (
                      <User className="w-4 h-4 text-zinc-500" />
                    )}
                    <span className="capitalize text-zinc-300">{u.role}</span>
                  </div>
                </td>
                <td className="px-6 py-4 text-zinc-500">
                  {new Date(u.created_at).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 text-right">
                  {currentUser?.id !== u.id && (
                    <button
                      onClick={() => {
                        if (confirm("Are you sure you want to delete this user?")) {
                          deleteMutation.mutate(u.id);
                        }
                      }}
                      className="p-2 text-zinc-500 hover:text-red-400 hover:bg-red-400/10 transition-colors"
                      title="Delete User"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
