import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { api, API_BASE } from "@/lib/api";
import { Loader2, Radio, Signal } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/live")({
  component: LivePage,
});

interface LiveStream {
  stream_key: string;
  started_at: string;
  playback_url: string;
}

function LivePage() {
  const [selectedStream, setSelectedStream] = useState<LiveStream | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["live-streams"],
    queryFn: api.getLiveStreams,
    refetchInterval: 5000,
  });

  const streams: LiveStream[] = data?.streams || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-white">Live Monitoring</h1>
        <p className="text-sm text-zinc-500 mt-1">Real-time status of active streams</p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-6 h-6 animate-spin text-zinc-500" />
        </div>
      ) : streams.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 border border-zinc-800 border-dashed rounded-lg opacity-50">
          <Signal className="w-10 h-10 text-zinc-600 mb-4" />
          <h3 className="text-lg font-medium text-zinc-400">No signals detected</h3>
          <p className="text-sm text-zinc-600 mt-1">Waiting for incoming RTMP streams...</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {streams.map((stream) => (
            <div
              key={stream.stream_key}
              className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden group"
            >
              {/* Player / Preview */}
              <div className="aspect-video bg-black relative">
                <video
                  autoPlay
                  muted
                  controls
                  className="w-full h-full object-contain"
                  src={`${API_BASE}${stream.playback_url}`}
                />
                <div className="absolute top-3 left-3 flex items-center gap-1.5 px-2 py-1 bg-red-500/90 text-white text-[10px] font-bold uppercase tracking-wider rounded">
                  <span className="w-1.5 h-1.5 bg-white rounded-full animate-pulse" />
                  Live
                </div>
              </div>

              {/* Info */}
              <div className="p-4">
                <div className="flex items-start justify-between">
                  <div>
                    <h4 className="font-medium text-white text-sm">{stream.stream_key}</h4>
                    <p className="text-xs text-zinc-500 mt-1">
                      Started {new Date(stream.started_at).toLocaleTimeString()}
                    </p>
                  </div>
                  <div className="flex items-center gap-1 text-xs text-emerald-500 bg-emerald-500/10 px-2 py-1 rounded border border-emerald-500/20">
                    <Signal className="w-3 h-3" />
                    <span>Stable</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
