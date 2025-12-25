import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Trash2, Play, Loader2, RefreshCw } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/")({
  component: VideosPage,
});

interface Video {
  id: string;
  title: string;
  status: string;
  source: string;
  created_at: string;
  playback_url: string;
  thumbnail_url: string;
}

function VideosPage() {
  const queryClient = useQueryClient();
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["videos"],
    queryFn: api.getVideos,
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["videos"] });
      if (selectedVideo) setSelectedVideo(null);
    },
  });

  const videos: Video[] = data?.videos || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">Videos</h1>
          <p className="text-sm text-zinc-500 mt-1">Manage and organize your video library</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="h-8 border-zinc-700 bg-transparent text-zinc-300 hover:bg-zinc-800 hover:text-white"
          onClick={() => queryClient.invalidateQueries({ queryKey: ["videos"] })}
        >
          <RefreshCw className="w-3.5 h-3.5 mr-2" />
          Refresh
        </Button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-6 h-6 animate-spin text-zinc-500" />
        </div>
      ) : videos.length === 0 ? (
        <div className="text-center py-20 border border-dashed border-zinc-800 rounded-lg">
          <p className="text-zinc-500 text-sm">No videos found in the library.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Player Section */}
          {selectedVideo && (
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg overflow-hidden">
              <div className="p-4 border-b border-zinc-800 flex justify-between items-center">
                <span className="text-sm font-medium text-white">{selectedVideo.title}</span>
                <button
                  onClick={() => setSelectedVideo(null)}
                  className="text-xs text-zinc-500 hover:text-zinc-300"
                >
                  Close Player
                </button>
              </div>
              <div className="relative aspect-video bg-black">
                <video
                  key={selectedVideo.id}
                  controls
                  autoPlay
                  className="w-full h-full"
                  src={`${API_BASE}${selectedVideo.playback_url}`}
                />
              </div>
            </div>
          )}

          {/* Table */}
          <div className="border border-zinc-800 rounded-lg overflow-hidden bg-zinc-900/20">
            <table className="w-full text-sm text-left">
              <thead className="bg-zinc-900/50 text-zinc-400 font-medium">
                <tr>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800 w-[80px]">Image</th>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800">Title</th>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800">Status</th>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800">Source</th>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800">Date</th>
                  <th className="px-4 py-3 font-medium border-b border-zinc-800 text-right">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800">
                {videos.map((video) => (
                  <tr key={video.id} className="hover:bg-zinc-900/30 transition-colors group">
                    <td className="px-4 py-3">
                      <div className="w-16 h-10 rounded bg-zinc-800 overflow-hidden relative border border-zinc-800">
                        <img
                          src={`${API_BASE}${video.thumbnail_url}`}
                          alt=""
                          className="w-full h-full object-cover"
                          onError={(e) => (e.currentTarget.style.display = "none")}
                        />
                      </div>
                    </td>
                    <td className="px-4 py-3 font-medium text-zinc-200">
                      {video.title}
                      <div className="text-xs text-zinc-500 mt-0.5 truncate max-w-[200px] font-mono opacity-50">
                        {video.id}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={video.status} />
                    </td>
                    <td className="px-4 py-3 text-zinc-400 capitalize">{video.source}</td>
                    <td className="px-4 py-3 text-zinc-400 tabular-nums">
                      {new Date(video.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="h-8 w-8 text-zinc-400 hover:text-white hover:bg-zinc-800"
                          disabled={video.status !== "ready"}
                          onClick={() => setSelectedVideo(video)}
                          title="Play"
                        >
                          <Play className="w-4 h-4 ml-0.5" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="h-8 w-8 text-zinc-400 hover:text-red-400 hover:bg-red-400/10"
                          onClick={() => {
                            if (confirm(`Delete "${video.title}"?`)) {
                              deleteMutation.mutate(video.id);
                            }
                          }}
                          title="Delete"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  if (status === "ready") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 capitalize">
        Ready
      </span>
    );
  }
  if (status === "processing") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-500/10 text-amber-500 border border-amber-500/20 capitalize">
        Processing
      </span>
    );
  }
  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-500 border border-red-500/20 capitalize">
      {status}
    </span>
  );
}
