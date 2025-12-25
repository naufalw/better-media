import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Plus, Copy, Loader2 } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/streams")({
  component: StreamsPage,
});

interface StreamKey {
  id: string;
  name: string;
  key: string;
  created_at: string;
}

function StreamsPage() {
  const queryClient = useQueryClient();
  const [newKeyName, setNewKeyName] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["stream-keys"],
    queryFn: api.getStreamKeys,
  });

  const createMutation = useMutation({
    mutationFn: (name: string) => api.createStreamKey(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["stream-keys"] });
      setNewKeyName("");
      setIsCreating(false);
    },
  });

  const streamKeys: StreamKey[] = data?.stream_keys || [];
  const rtmpUrl = "rtmp://localhost:1935/live";

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-xl font-semibold text-white">Stream Keys</h1>
        <p className="text-sm text-zinc-500 mt-1">
          Manage RTMP credentials for streaming software (OBS, etc)
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Create Form */}
          <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg p-4">
            <div className="flex gap-3">
              <input
                type="text"
                placeholder="Key name (e.g. OBS Main)"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="flex-1 bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-zinc-600 focus:ring-1 focus:ring-zinc-600 transition-all"
              />
              <Button
                onClick={() => createMutation.mutate(newKeyName)}
                disabled={!newKeyName || createMutation.isPending}
                className="bg-white text-black hover:bg-zinc-200"
              >
                {createMutation.isPending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  "Create New Key"
                )}
              </Button>
            </div>
          </div>

          {/* List */}
          {isLoading ? (
            <div className="py-10 flex justify-center">
              <Loader2 className="w-5 h-5 animate-spin text-zinc-600" />
            </div>
          ) : streamKeys.length === 0 ? (
            <div className="text-center py-10 text-sm text-zinc-500">No stream keys found.</div>
          ) : (
            <div className="space-y-3">
              {streamKeys.map((key) => (
                <div
                  key={key.id}
                  className="bg-zinc-900/20 border border-zinc-800 rounded-lg p-4 flex items-center justify-between group"
                >
                  <div>
                    <div className="font-medium text-white text-sm">{key.name}</div>
                    <div className="text-xs text-zinc-500 mt-0.5 font-mono opacity-60">
                      Created {new Date(key.created_at).toLocaleDateString()}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <code className="px-2 py-1 bg-zinc-950 border border-zinc-800 rounded text-xs text-zinc-400 font-mono">
                      {key.key}
                    </code>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="h-7 w-7 text-zinc-500 hover:text-white"
                      onClick={() => copyToClipboard(key.key)}
                    >
                      <Copy className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Sidebar Info */}
        <div className="space-y-6">
          <div className="bg-zinc-900/20 border border-zinc-800 rounded-lg p-5">
            <h3 className="font-medium text-white text-sm mb-3">Connection Details</h3>
            <div className="space-y-4">
              <div>
                <label className="text-xs font-medium text-zinc-500 uppercase tracking-wider block mb-1.5">
                  Server URL
                </label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 bg-zinc-950 border border-zinc-800 rounded px-2 py-1.5 text-xs text-zinc-300 font-mono overflow-auto whitespace-nowrap scrollbar-hide">
                    {rtmpUrl}
                  </code>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 text-zinc-500 hover:text-white"
                    onClick={() => copyToClipboard(rtmpUrl)}
                  >
                    <Copy className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
              <p className="text-xs text-zinc-500 leading-relaxed">
                Enter this URL into your streaming software (OBS, vMix, etc) along with one of your
                stream keys.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
