import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Trash2, Play, Loader2, RefreshCw, List, Grid } from "lucide-react";
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
  // const [selectedVideo, setSelectedVideo] = useState<Video | null>(null); // Removed

  const { data, isLoading } = useQuery({
    queryKey: ["videos"],
    queryFn: api.getVideos,
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["videos"] });
      // if (selectedVideo) setSelectedVideo(null); // Removed
    },
  });

  const [viewMode, setViewMode] = useState<"list" | "grid">("list");
  const videos: Video[] = data?.videos || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">Videos</h1>
          <p className="text-sm text-zinc-500 mt-1">Manage and organize your video library</p>
        </div>
        <div className="flex gap-3">
          <div className="flex h-8 bg-zinc-900 border border-zinc-800 rounded-md p-1 gap-1 items-center">
            <button
              onClick={() => setViewMode("list")}
              className={`h-full aspect-square flex items-center justify-center rounded-sm transition-all ${viewMode === "list" ? "bg-zinc-800 text-white shadow-sm" : "text-zinc-500 hover:text-zinc-300"}`}
              title="List View"
            >
              <List className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode("grid")}
              className={`h-full aspect-square flex items-center justify-center rounded-sm transition-all ${viewMode === "grid" ? "bg-zinc-800 text-white shadow-sm" : "text-zinc-500 hover:text-zinc-300"}`}
              title="Grid View"
            >
              <Grid className="w-4 h-4" />
            </button>
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
          {/* Player Section - REMOVED */}
          {/* {selectedVideo && (
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg overflow-hidden animate-in fade-in slide-in-from-top-4 duration-300">
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
          )} */}

          {/* Views */}
          {viewMode === "list" ? (
            /* List View */
            <div className="border border-zinc-800 rounded-lg overflow-hidden bg-zinc-900/20">
              <table className="w-full text-sm text-left">
                <thead className="bg-zinc-900/50 text-zinc-400 font-medium">
                  <tr>
                    <th className="px-4 py-3 font-medium border-b border-zinc-800 w-[80px]">
                      Image
                    </th>
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
                        <Link
                          to="/video/$videoId"
                          params={{ videoId: video.id }}
                          className="block w-16 h-10 rounded bg-zinc-800 overflow-hidden relative border border-zinc-800 cursor-pointer"
                        >
                          <img
                            src={`${API_BASE}${video.thumbnail_url}`}
                            alt=""
                            className="w-full h-full object-cover"
                            onError={(e) => (e.currentTarget.style.display = "none")}
                          />
                        </Link>
                      </td>
                      <td className="px-4 py-3 font-medium text-zinc-200">
                        <Link
                          to="/video/$videoId"
                          params={{ videoId: video.id }}
                          className="cursor-pointer hover:underline"
                        >
                          {video.title}
                        </Link>
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
                          <Link
                            to="/video/$videoId"
                            params={{ videoId: video.id }}
                            className="inline-flex items-center justify-center h-8 w-8 rounded-md hover:bg-zinc-800 hover:text-white text-zinc-400 transition-colors"
                            title="Play"
                          >
                            <Play className="w-4 h-4 ml-0.5" />
                          </Link>
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
          ) : (
            /* Grid View */
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
              {videos.map((video) => (
                <div
                  key={video.id}
                  className="group bg-zinc-900/30 border border-zinc-800 rounded-lg overflow-hidden hover:border-zinc-700 transition-colors"
                >
                  <div className="aspect-video bg-zinc-900 relative">
                    <Link
                      to="/video/$videoId"
                      params={{ videoId: video.id }}
                      className="block w-full h-full"
                    >
                      <img
                        src={`${API_BASE}${video.thumbnail_url}`}
                        alt=""
                        className="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity"
                        onError={(e) => (e.currentTarget.style.display = "none")}
                      />
                      {["playable", "completed", "ready"].includes(video.status) && (
                        <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/40">
                          <div className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center">
                            <Play className="w-4 h-4 text-white ml-0.5" />
                          </div>
                        </div>
                      )}
                    </Link>
                  </div>
                  <div className="p-3">
                    <div className="flex justify-between items-start gap-2 mb-3">
                      <Link
                        to="/video/$videoId"
                        params={{ videoId: video.id }}
                        className="font-medium text-sm text-zinc-200 line-clamp-2 hover:underline leading-snug"
                      >
                        {video.title}
                      </Link>
                      <button
                        className="text-zinc-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                        onClick={() => {
                          if (confirm(`Delete "${video.title}"?`)) {
                            deleteMutation.mutate(video.id);
                          }
                        }}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                    <div className="flex justify-between items-center text-xs text-zinc-500">
                      <StatusBadge status={video.status} />
                      <div className="flex gap-2">
                        <span>{new Date(video.created_at).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
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
