import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { Loader2, Plus, Library, Trash2, Folder } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/_padded/libraries")({
  component: LibrariesPage,
});

function LibrariesPage() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [formData, setFormData] = useState({ name: "", description: "" });
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["libraries"],
    queryFn: api.listLibraries,
  });

  const createMutation = useMutation({
    mutationFn: (data: { name: string; description: string }) =>
      api.createLibrary(data.name, data.description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["libraries"] });
      setIsCreateOpen(false);
      setFormData({ name: "", description: "" });
      setError("");
    },
    onError: (err: any) => {
      setError(err.message || "Failed to create library");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteLibrary,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["libraries"] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
      </div>
    );
  }

  const libraries = data?.libraries || [];

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between pb-8 border-b border-zinc-900">
        <div>
          <h1 className="text-xl font-semibold text-white">Libraries</h1>
          <p className="text-sm text-zinc-500 mt-1">Manage your media collections</p>
        </div>
        <button
          onClick={() => setIsCreateOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-black bg-emerald-500 hover:bg-emerald-400 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 rounded-sm"
        >
          <Plus className="w-4 h-4" />
          New Library
        </button>
      </div>

      <div className="flex-1 overflow-auto py-8">
        {isCreateOpen && (
          <div className="p-4 border border-zinc-800 bg-zinc-900/50 space-y-4 mb-6 rounded-sm">
            <h2 className="text-lg font-medium text-white">Create New Library</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-400">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  required
                  placeholder="e.g. Movies, TV Shows"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-400">Description</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="flex h-10 w-full bg-[#0c0c0e] border border-zinc-800 px-3 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  placeholder="Optional description"
                />
              </div>
              {error && <div className="text-sm text-red-400">{error}</div>}
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className="px-4 py-2 text-sm font-medium text-black bg-emerald-500 hover:bg-emerald-400 rounded-sm"
                >
                  {createMutation.isPending ? "Creating..." : "Create"}
                </button>
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white bg-transparent border border-zinc-800 hover:bg-zinc-800 rounded-sm"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {libraries.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-zinc-500">
            <Library className="w-12 h-12 mb-4 opacity-20" />
            <p className="text-lg font-medium">No libraries yet</p>
            <p className="text-sm">Create one to start organizing your videos</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {libraries.map((lib: any) => (
              <div
                key={lib.id}
                className="group relative bg-[#09090b] border border-zinc-800 hover:border-zinc-700 transition-all rounded-sm p-4 flex flex-col gap-4"
              >
                <div className="flex items-start justify-between">
                  <div className="p-2 bg-zinc-900 rounded-sm">
                    <Folder className="w-6 h-6 text-emerald-500" />
                  </div>
                  <button
                    onClick={(e) => {
                      e.preventDefault();
                      if (
                        confirm(
                          `Delete library "${lib.name}"? This might delete all videos inside it.`,
                        )
                      ) {
                        deleteMutation.mutate(lib.id);
                      }
                    }}
                    className="p-1.5 text-zinc-500 hover:text-red-400 hover:bg-red-900/10 rounded-sm transition-colors opacity-0 group-hover:opacity-100"
                    title="Delete Library"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>

                <div>
                  <h3 className="text-lg font-medium text-zinc-200 group-hover:text-white transition-colors">
                    <Link
                      to="/libraries/$libraryId"
                      params={{ libraryId: lib.id }}
                      search={{ view: "list" }}
                    >
                      {lib.name}
                    </Link>
                  </h3>
                  <p className="text-sm text-zinc-500 line-clamp-2 mt-1">
                    {lib.description || "No description"}
                  </p>
                </div>

                <div className="text-xs text-zinc-600 mt-auto pt-4 border-t border-zinc-900 flex justify-between">
                  <span>{new Date(lib.created_at).toLocaleDateString()}</span>
                  {/* We could show video count here if the list API returned it, but currently list endpoint doesn't return count I think? Check handlers_libraries.go: handleListLibraries -> db.ListLibraries -> SELECT id, name, description... no count. */}
                </div>

                <Link
                  to="/libraries/$libraryId"
                  params={{ libraryId: lib.id }}
                  search={{ view: "list" }}
                  className="absolute inset-0 z-0"
                />

                {/* Ensure buttons are above the link */}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
