import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Trash2, Calendar, HardDrive, FileVideo, CheckCircle2 } from "lucide-react";
import { Loader2 } from "lucide-react";
import { PaddedLayout } from "@/components/padded-layout";

export const Route = createFileRoute("/libraries/$libraryId/video/$videoId")({
  component: VideoDetailPage,
});

function VideoDetailPage() {
  const { libraryId, videoId } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: video, isLoading } = useQuery({
    queryKey: ["video", videoId],
    queryFn: () => api.getVideo(videoId),
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status !== "ready" && status !== "failed" ? 550 : false;
    },
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["library", libraryId] });
      navigate({ to: "/libraries/$libraryId", params: { libraryId }, search: { view: "list" } });
    },
  });

  const { data: transcription } = useQuery({
    queryKey: ["transcription", videoId, video?.subtitle_url],
    queryFn: async () => {
      if (!video?.subtitle_url) return null;
      const res = await fetch(`${API_BASE}${video.subtitle_url}`);
      if (!res.ok) return null;
      const text = await res.text();
      // Display raw VTT content as requested
      return text.trim();
    },
    enabled: !!video?.subtitle_url,
  });

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
      </div>
    );
  }

  if (!video) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <p className="text-zinc-500">Video not found.</p>
        <Link
          to="/libraries/$libraryId"
          params={{ libraryId }}
          search={{ view: "list" }}
          className="text-emerald-500 hover:underline"
        >
          Back to Library
        </Link>
      </div>
    );
  }

  return (
    <PaddedLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Link
            to="/libraries/$libraryId"
            params={{ libraryId }}
            search={{ view: "list" }}
            className="p-2 -ml-2 text-zinc-400 hover:text-white hover:bg-zinc-800 rounded-full transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div className="flex-1 min-w-0">
            <h1 className="text-xl font-semibold text-white truncate">{video.title}</h1>
            <div className="flex items-center gap-2 text-xs text-zinc-500 mt-1">
              <span className="font-mono">{video.id}</span>
            </div>
          </div>
          <Button
            variant="destructive"
            size="sm"
            className="bg-red-500/10 text-red-500 hover:bg-red-500/20 border-red-500/20 border"
            onClick={() => {
              if (confirm("Are you sure?")) deleteMutation.mutate(video.id);
            }}
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete Video
          </Button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content (Player) */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-black border border-zinc-800 rounded-lg overflow-hidden aspect-video relative group">
              {["playable", "completed", "ready"].includes(video.status) ? (
                <video
                  controls
                  autoPlay
                  className="w-full h-full"
                  src={`${API_BASE}${video.playback_url}`}
                  poster={`${API_BASE}${video.thumbnail_url}`}
                >
                  {video.subtitle_url && transcription && (
                    <track
                      kind="subtitles"
                      src={URL.createObjectURL(new Blob([transcription], { type: "text/vtt" }))}
                      srcLang="en"
                      label="English"
                      default
                    />
                  )}
                </video>
              ) : (
                <div className="absolute inset-0 flex items-center justify-center bg-zinc-900">
                  <div className="text-center w-full max-w-md px-6">
                    <div className="inline-block p-3 bg-zinc-800 rounded-full mb-3">
                      <FileVideo className="w-6 h-6 text-zinc-500" />
                    </div>
                    <p className="text-zinc-400 font-medium mb-2">
                      {video.status === "processing" ? "Processing..." : `Status: ${video.status}`}
                    </p>
                  </div>
                </div>
              )}
            </div>

            <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg p-6">
              <h3 className="font-medium text-white mb-4">Transcription</h3>
              <div className="text-zinc-400 text-sm leading-relaxed max-h-60 overflow-y-auto font-mono whitespace-pre-wrap">
                {transcription ? (
                  <p>{transcription}</p>
                ) : (
                  <p className="italic text-zinc-600 font-sans">
                    {video.transcription_status === "failed"
                      ? "Transcription failed."
                      : "No transcription available yet."}
                  </p>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar (Metadata) */}
          <div className="space-y-6">
            <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg overflow-hidden">
              <div className="px-4 py-3 border-b border-zinc-800 bg-zinc-900/50">
                <h3 className="text-sm font-medium text-white">Metadata</h3>
              </div>
              <div className="p-4 space-y-4">
                {/* Progress Bar in Sidebar */}
                {video.status !== "ready" && video.status !== "failed" && (
                  <div className="space-y-2 pb-4 border-b border-zinc-800">
                    <div className="flex justify-between text-xs text-zinc-400">
                      <span>Processing</span>
                      <span>{video.progress || 0}%</span>
                    </div>
                    <div className="w-full bg-zinc-800 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="bg-emerald-500 h-full transition-all duration-500 ease-out"
                        style={{ width: `${video.progress || 0}%` }}
                      />
                    </div>
                  </div>
                )}

                <div className="flex items-start gap-3">
                  <Calendar className="w-4 h-4 text-zinc-500 mt-0.5" />
                  <div>
                    <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                      Created At
                    </div>
                    <div className="text-sm text-zinc-300">
                      {new Date(video.created_at).toLocaleString()}
                    </div>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <HardDrive className="w-4 h-4 text-zinc-500 mt-0.5" />
                  <div>
                    <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                      Source
                    </div>
                    <div className="text-sm text-zinc-300 capitalize">{video.source}</div>
                  </div>
                </div>
                {video.file_size_bytes && (
                  <div className="flex items-start gap-3">
                    <HardDrive className="w-4 h-4 text-zinc-500 mt-0.5" />
                    <div>
                      <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                        Size
                      </div>
                      <div className="text-sm text-zinc-300">
                        {formatFileSize(video.file_size_bytes)}
                      </div>
                    </div>
                  </div>
                )}
                {video.resolution_width && video.resolution_height && (
                  <div className="flex items-start gap-3">
                    <FileVideo className="w-4 h-4 text-zinc-500 mt-0.5" />
                    <div>
                      <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                        Resolution
                      </div>
                      <div className="text-sm text-zinc-300">
                        {video.resolution_width}x{video.resolution_height}
                      </div>
                    </div>
                  </div>
                )}
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="w-4 h-4 text-zinc-500 mt-0.5" />
                  <div>
                    <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                      Status
                    </div>
                    <div className="text-sm">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${
                          ["ready", "completed", "playable"].includes(video.status)
                            ? "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20"
                            : "bg-amber-500/10 text-amber-500 border border-amber-500/20"
                        }`}
                      >
                        {video.status}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PaddedLayout>
  );
}
