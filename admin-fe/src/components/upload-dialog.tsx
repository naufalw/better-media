import { useState, useCallback, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Upload as UploadIcon, FileVideo, Check, X, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

type UploadState = "idle" | "uploading" | "transcoding" | "done" | "error";

interface UploadDialogProps {
  initialFile?: File | null;
  onOpenChange?: (open: boolean) => void;
  triggerClassName?: string;
  triggerSize?: "default" | "sm" | "lg" | "icon";
}

export function UploadDialog({
  initialFile,
  onOpenChange: parentOnOpenChange,
  triggerClassName,
  triggerSize = "default",
}: UploadDialogProps = {}) {
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [state, setState] = useState<UploadState>("idle");
  const [progress, setProgress] = useState(0);
  const [transcribe, setTranscribe] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (initialFile) {
      setFile(initialFile);
      setOpen(true);
    }
  }, [initialFile]);

  const reset = () => {
    setFile(null);
    setState("idle");
    setProgress(0);
    setError(null);
    if (parentOnOpenChange) parentOnOpenChange(false);
  };

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("No file selected");

      setState("uploading");
      setProgress(0);

      // Step 1: Get presigned URL
      const uploadData = await api.createUpload(file.name);
      const presignedUrl = uploadData.url;
      const vid = uploadData.videoId;

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

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Optional: warn if uploading?
      if (state === "uploading" || state === "transcoding") {
        if (
          !confirm(
            "Upload is in progress. Closing will not stop it but you will lose progress view. Close?",
          )
        ) {
          return;
        }
      }
      // Reset on close if done
      if (state === "done") {
        reset();
      }
      if (parentOnOpenChange) parentOnOpenChange(false);
    }
    setOpen(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button
          size={triggerSize}
          className={cn(
            "bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-900/20",
            triggerClassName,
          )}
        >
          <UploadIcon className="w-4 h-4 mr-2" />
          Upload
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-[#0c0c0e] border-zinc-800 text-zinc-100 sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Upload Media</DialogTitle>
          <DialogDescription className="text-zinc-500">
            Upload video files for transcoding and streaming.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-4 min-w-0">
          {state === "idle" && !file ? (
            <div
              onDrop={handleDrop}
              onDragOver={(e) => e.preventDefault()}
              className="border border-dashed border-zinc-800 p-10 text-center hover:bg-zinc-900/50 hover:border-zinc-700 transition-colors cursor-pointer bg-zinc-900/20"
              onClick={() => document.getElementById("dialog-file-upload")?.click()}
            >
              <div className="w-12 h-12 bg-zinc-900 flex items-center justify-center mx-auto mb-4 border border-zinc-800">
                <UploadIcon className="w-6 h-6 text-zinc-400" />
              </div>
              <h3 className="text-sm font-medium text-white">Click or drag video to upload</h3>
              <p className="text-xs text-zinc-500 mt-1 max-w-xs mx-auto">
                Support for MP4, MOV, MKV.
              </p>
              <input
                type="file"
                accept="video/*"
                className="hidden"
                id="dialog-file-upload"
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f) setFile(f);
                }}
              />
            </div>
          ) : (
            <div className="bg-zinc-900/30 border border-zinc-800 p-6">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 bg-zinc-900 flex items-center justify-center border border-zinc-800 shrink-0">
                  <FileVideo className="w-5 h-5 text-zinc-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <span
                      className="text-sm font-medium text-white truncate flex-1 min-w-0"
                      title={file?.name}
                    >
                      {file?.name}
                    </span>
                    <button
                      onClick={reset}
                      disabled={state !== "idle" && state !== "done" && state !== "error"}
                      className="text-zinc-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
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
                          className="border-zinc-700 bg-zinc-800 text-zinc-200 focus:ring-0 w-4 h-4"
                        />
                        <span className="text-sm text-zinc-400 group-hover:text-zinc-300">
                          Generate AI Subtitles
                        </span>
                      </label>
                      <Button
                        onClick={() => uploadMutation.mutate()}
                        disabled={uploadMutation.isPending}
                        className="w-full bg-white text-black hover:bg-zinc-200 font-medium"
                      >
                        {uploadMutation.isPending && (
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        )}
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
                      <div className="h-1.5 w-full bg-zinc-800 overflow-hidden">
                        <div
                          className="h-full bg-emerald-500 transition-all duration-300 ease-out"
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
      </DialogContent>
    </Dialog>
  );
}
