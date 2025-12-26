import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Trash2, Calendar, HardDrive, FileVideo, CheckCircle2 } from "lucide-react";
import { Loader2 } from "lucide-react";

export const Route = createFileRoute("/_padded/video/$videoId")({
  component: VideoDetailPage,
});

function VideoDetailPage() {
  const { videoId } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // We reuse getVideos for now and find the item, or we could fetch single if API supported it
  // Since we don't have getSingleVideo yet, we'll fetch all and find.
  // Ideally, we should add api.getVideo(id).
  const { data, isLoading } = useQuery({
    queryKey: ["videos"],
    queryFn: api.getVideos,
  });

  const video = data?.videos.find((v: any) => v.id === videoId);

  const deleteMutation = useMutation({
    mutationFn: api.deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["videos"] });
      navigate({ to: "/video" });
    },
  });

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
        <Link to="/video" className="text-white hover:underline">
          Go back home
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link
          to="/video"
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
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center bg-zinc-900">
                <div className="text-center">
                  <div className="inline-block p-3 bg-zinc-800 rounded-full mb-3">
                    <FileVideo className="w-6 h-6 text-zinc-500" />
                  </div>
                  <p className="text-zinc-400 font-medium">Video is processing...</p>
                </div>
              </div>
            )}
          </div>

          <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg p-6">
            <h3 className="font-medium text-white mb-4">Transcription</h3>
            <div className="text-zinc-400 text-sm leading-relaxed">
              {/* Placeholder for transcription */}
              <p className="italic text-zinc-600">No transcription available yet.</p>
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
              <div className="flex items-start gap-3">
                <CheckCircle2 className="w-4 h-4 text-zinc-500 mt-0.5" />
                <div>
                  <div className="text-xs text-zinc-500 uppercase tracking-wider font-medium">
                    Status
                  </div>
                  <div className="text-sm">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${
                        video.status === "ready"
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
  );
}
