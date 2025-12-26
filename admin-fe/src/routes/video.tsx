import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Trash2, Play, Loader2, RefreshCw, List, Grid, Upload } from "lucide-react";
import { useState } from "react";
import { UploadDialog } from "@/components/upload-dialog";

export const Route = createFileRoute("/video")({
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
  const [droppedFile, setDroppedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.types.some((t) => t === "Files")) {
      setIsDragging(true);
    }
  };

  const handleOverlayDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // Prevent flickering when dragging over children
    if (e.currentTarget.contains(e.relatedTarget as Node)) return;
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const file = e.dataTransfer.files[0];
    if (file?.type.startsWith("video/")) {
      setDroppedFile(file);
    }
  };
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
    <div className="h-full flex flex-col">
      {/* Page Header */}
      <div className="flex items-center justify-between p-8 border-b border-zinc-900">
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
            className="h-8 border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800 hover:text-white"
            onClick={() => queryClient.invalidateQueries({ queryKey: ["videos"] })}
          >
            <RefreshCw className="w-3.5 h-3.5 mr-2" />
            Refresh
          </Button>
          <UploadDialog
            triggerSize="sm"
            initialFile={droppedFile}
            onOpenChange={(open) => !open && setDroppedFile(null)}
          />
        </div>
      </div>
      <div
        className="flex-1 overflow-hidden flex flex-col relative"
        onDragEnter={handleDragEnter}
        onDragOver={(e) => e.preventDefault()}
        onDrop={handleDrop}
      >
        {isDragging && (
          <div
            className="absolute inset-0 z-[100] bg-[#09090b]/95 border border-zinc-800 m-2 flex flex-col items-center justify-start pt-[20vh] animate-in fade-in zoom-in-95 duration-200"
            onDragOver={(e) => {
              e.preventDefault();
              e.stopPropagation();
            }}
            onDragLeave={handleOverlayDragLeave}
            onDrop={handleDrop}
          >
            <div className="bg-zinc-900 border border-zinc-700 p-4 mb-4 pointer-events-none">
              <Upload className="w-8 h-8 text-white" />
            </div>
            <p className="text-lg font-medium text-white pointer-events-none">
              Drop video to upload
            </p>
          </div>
        )}
        {isLoading ? (
          <div className="flex-1 flex items-center justify-center">
            <Loader2 className="w-6 h-6 animate-spin text-zinc-500" />
          </div>
        ) : videos.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center p-20">
            <div className="p-4 border border-zinc-800 border-dashed rounded-full bg-zinc-900/50 mb-4">
              <List className="w-6 h-6 text-zinc-500" />
            </div>
            <p className="text-zinc-400 font-medium">No videos found</p>
            <p className="text-zinc-600 text-sm mt-1">Upload a video to get started</p>
          </div>
        ) : (
          <div className="flex-1 overflow-hidden flex flex-col">
            {viewMode === "list" ? (
              /* List View */
              <div className="flex-1 overflow-auto">
                <table className="w-full text-sm text-left border-collapse">
                  <thead className="text-zinc-400 font-medium bg-[#09090b] sticky top-0 z-30 shadow-sm">
                    <tr>
                      <th className="sticky border-r left-0 z-20 bg-[#09090b] px-6 py-3 border-b border-zinc-900 min-w-[350px] max-w-[350px]">
                        Video
                      </th>
                      <th className="px-6 py-3 border-b border-zinc-900 min-w-[120px] bg-[#09090b]">
                        Status
                      </th>
                      <th className="px-6 py-3 border-b border-zinc-900 min-w-[280px] bg-[#09090b]">
                        ID
                      </th>
                      <th className="px-6 py-3 border-b border-zinc-900 min-w-[150px] bg-[#09090b]">
                        Source
                      </th>
                      <th className="px-6 py-3 border-b border-zinc-900 min-w-[150px] bg-[#09090b]">
                        Date
                      </th>
                      <th className="px-6 py-3 border-b border-zinc-900 min-w-[150px] bg-[#09090b]">
                        Size
                      </th>
                      <th className="px-6 py-3 border-b border-l border-zinc-900 bg-[#09090b] text-right sticky right-0 z-20">
                        Action
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-zinc-900 bg-[#09090b]">
                    {videos.map((video) => (
                      <tr key={video.id} className="hover:bg-zinc-900/50 transition-colors group">
                        <td className="sticky left-0  z-10 bg-[#09090b] border-r border-zinc-900 group-hover:bg-zinc-900 px-6 py-4 max-w-[260px]">
                          <div className="flex items-center gap-3">
                            <Link
                              to="/video/$videoId"
                              params={{ videoId: video.id }}
                              className="block  w-20 aspect-video rounded-sm bg-zinc-900 overflow-hidden relative shrink-0 border border-zinc-800"
                            >
                              <img
                                src={`${API_BASE}${video.thumbnail_url}`}
                                alt=""
                                className="w-full h-full object-cover"
                                onError={(e) => (e.currentTarget.style.display = "none")}
                              />
                            </Link>
                            <div className="min-w-0">
                              <Link
                                to="/video/$videoId"
                                params={{ videoId: video.id }}
                                className="font-medium text-zinc-200 hover:text-white hover:underline line-clamp-2 text-base"
                              >
                                {video.title}
                              </Link>
                              <div className="mt-1 flex items-center gap-2">
                                {/* Extra meta if needed */}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <StatusBadge status={video.status} />
                        </td>
                        <td className="px-6 py-4 text-zinc-500 font-mono text-xs select-all">
                          {video.id}
                        </td>
                        <td className="px-6 py-4 text-zinc-400 capitalize">{video.source}</td>
                        <td className="px-6 py-4 text-zinc-400 tabular-nums">
                          {new Date(video.created_at).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 text-zinc-500 font-mono text-xs">
                          {/* Mock Size */}
                          -- MB
                        </td>
                        <td className="px-6 py-4 text-right sticky right-0 z-10 bg-[#09090b] border-l border-zinc-900 group-hover:bg-zinc-900">
                          <div className="flex items-center justify-end gap-3">
                            <Link
                              to="/video/$videoId"
                              params={{ videoId: video.id }}
                              className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium rounded transition-colors shadow-lg shadow-emerald-900/20"
                            >
                              View
                            </Link>
                            <button
                              onClick={() => {
                                if (confirm(`Delete "${video.title}"?`)) {
                                  deleteMutation.mutate(video.id);
                                }
                              }}
                              className="p-1.5 text-zinc-500 hover:text-red-400 hover:bg-red-400/10 rounded transition-colors"
                              title="Delete"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              /* Grid View */
              <div className="p-6 overflow-auto bg-[#09090b]">
                <div className="grid grid-cols-2  md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5 gap-6">
                  {videos.map((video) => (
                    /* Existing Grid Item Code ... need to rewrite it since I replaced the whole block */
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
              </div>
            )}
          </div>
        )}
      </div>
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
