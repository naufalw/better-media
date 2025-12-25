import { createFileRoute } from "@tanstack/react-router";
import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Upload as UploadIcon, FileVideo, Check, X } from "lucide-react";

export const Route = createFileRoute("/upload")({
  component: UploadPage,
});

type UploadState = "idle" | "uploading" | "transcoding" | "done" | "error";

function UploadPage() {
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [state, setState] = useState<UploadState>("idle");
  const [progress, setProgress] = useState(0);
  const [transcribe, setTranscribe] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [_videoId, setVideoId] = useState<string | null>(null);

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("No file selected");

      setState("uploading");
      setProgress(0);

      // Step 1: Get presigned URL
      const uploadData = await api.createUpload(file.name);
      const presignedUrl = uploadData.url;
      const vid = uploadData.videoId;
      setVideoId(vid);

      // Step 2: Upload to S3
      const xhr = new XMLHttpRequest();
      await new Promise((resolve, reject) => {
        xhr.upload.addEventListener("progress", (event) => {
          if (event.lengthComputable) {
            setProgress(Math.round((event.loaded / event.total) * 100));
          }
        });
        xhr.open("PUT", presignedUrl);
        xhr.setRequestHeader("Content-Type", file.type);
        xhr.onload = () =>
          xhr.status >= 200 && xhr.status < 300
            ? resolve(null)
            : reject(new Error("Upload failed"));
        xhr.onerror = () => reject(new Error("Upload failed"));
        xhr.send(file);
      });

      setState("transcoding");
      setProgress(0);

      // Step 3: Start transcoding
      const { job_id } = await api.startTranscoding(vid, file.name, transcribe);

      // Step 4: Poll
      while (true) {
        await new Promise((r) => setTimeout(r, 2000));
        const job = await api.getJob(job_id);
        setProgress(job.progress || 0);

        if (job.status === "completed" || job.status === "playable" || job.status === "ready") {
          setState("done");
          queryClient.invalidateQueries({ queryKey: ["videos"] });
          break;
        }
        if (job.status === "failed") {
          throw new Error(job.error || "Transcoding failed");
        }
      }
    },
    onError: (err) => {
      setState("error");
      setError(err instanceof Error ? err.message : "Unknown error");
    },
  });

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile?.type.startsWith("video/")) {
      setFile(droppedFile);
      setState("idle");
      setError(null);
    }
  }, []);

  const reset = () => {
    setFile(null);
    setState("idle");
    setProgress(0);
    setError(null);
    setVideoId(null);
  };

  return (
    <div className="max-w-3xl space-y-8">
      <div>
        <h1 className="text-xl font-semibold text-white">Upload Media</h1>
        <p className="text-sm text-zinc-500 mt-1">
          Upload video files for transcoding and streaming
        </p>
      </div>

      <div className="space-y-4">
        {/* Dropzone */}
        {state === "idle" && !file ? (
          <div
            onDrop={handleDrop}
            onDragOver={(e) => e.preventDefault()}
            className="border border-dashed border-zinc-700 rounded-lg p-16 text-center hover:bg-zinc-900/50 hover:border-zinc-500 transition-colors cursor-default"
          >
            <div className="w-12 h-12 bg-zinc-900 rounded-lg flex items-center justify-center mx-auto mb-4 border border-zinc-800">
              <UploadIcon className="w-6 h-6 text-zinc-400" />
            </div>
            <h3 className="text-sm font-medium text-white">Click or drag video to upload</h3>
            <p className="text-xs text-zinc-500 mt-1 max-w-xs mx-auto">
              Support for MP4, MOV, MKV. Maximum file size 2GB.
            </p>
            <input
              type="file"
              accept="video/*"
              className="hidden"
              id="file-upload"
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) setFile(f);
              }}
            />
            <label
              htmlFor="file-upload"
              className="mt-6 inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded-md cursor-pointer transition-colors"
            >
              Select File
            </label>
          </div>
        ) : (
          <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-zinc-900 rounded flex items-center justify-center border border-zinc-800 shrink-0">
                <FileVideo className="w-5 h-5 text-zinc-400" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-white truncate">{file?.name}</span>
                  <button
                    onClick={reset}
                    disabled={state !== "idle" && state !== "done" && state !== "error"}
                    className="text-zinc-500 hover:text-white"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
                <p className="text-xs text-zinc-500 mb-4">
                  {(file!.size / (1024 * 1024)).toFixed(2)} MB
                </p>

                {state === "idle" && (
                  <div className="space-y-4">
                    <label className="flex items-center gap-2 cursor-pointer select-none group">
                      <input
                        type="checkbox"
                        checked={transcribe}
                        onChange={(e) => setTranscribe(e.target.checked)}
                        className="rounded border-zinc-700 bg-zinc-800 text-zinc-200 focus:ring-0 w-4 h-4"
                      />
                      <span className="text-sm text-zinc-400 group-hover:text-zinc-300">
                        Generate AI Subtitles
                      </span>
                    </label>
                    <Button
                      onClick={() => uploadMutation.mutate()}
                      className="w-full bg-white text-black hover:bg-zinc-200 font-medium"
                    >
                      Start Upload
                    </Button>
                  </div>
                )}

                {(state === "uploading" || state === "transcoding") && (
                  <div className="space-y-2">
                    <div className="flex justify-between text-xs text-zinc-400">
                      <span className="capitalize">{state}...</span>
                      <span>{progress}%</span>
                    </div>
                    <div className="h-1.5 w-full bg-zinc-800 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-white transition-all duration-300 ease-out"
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                  </div>
                )}

                {state === "done" && (
                  <div className="flex items-center gap-2 text-sm text-emerald-500 mt-2">
                    <Check className="w-4 h-4" />
                    <span>Completed successfully</span>
                  </div>
                )}

                {state === "error" && (
                  <div className="flex items-center gap-2 text-sm text-red-500 mt-2">
                    <X className="w-4 h-4" />
                    <span>{error}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
