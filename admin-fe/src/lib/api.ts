const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";

export const api = {
  // Videos
  async getVideos() {
    const res = await fetch(`${API_BASE}/v1/videos`);
    if (!res.ok) throw new Error("Failed to fetch videos");
    return res.json();
  },

  async deleteVideo(id: string) {
    const res = await fetch(`${API_BASE}/v1/videos/${id}`, { method: "DELETE" });
    if (!res.ok) throw new Error("Failed to delete video");
    return res.json();
  },

  async getVideoPlaybackUrl(id: string) {
    return `${API_BASE}/v1/videos/${id}/playback/hls/master.m3u8`;
  },

  // Upload
  async createUpload(fileName: string) {
    const res = await fetch(`${API_BASE}/v1/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ file_name: fileName }),
    });
    if (!res.ok) throw new Error("Failed to create upload");
    return res.json();
  },

  async startTranscoding(videoId: string, inputFile: string, transcribe = false) {
    const res = await fetch(`${API_BASE}/v1/jobs/transcoding`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        video_id: videoId,
        input_file: inputFile,
        resolutions: [360, 720, 1080],
        transcribe,
      }),
    });
    if (!res.ok) throw new Error("Failed to start transcoding");
    return res.json();
  },

  async getJob(jobId: string) {
    const res = await fetch(`${API_BASE}/v1/jobs/${jobId}`);
    if (!res.ok) throw new Error("Failed to get job status");
    return res.json();
  },

  // Stream Keys
  async getStreamKeys() {
    const res = await fetch(`${API_BASE}/v1/stream-keys`);
    if (!res.ok) throw new Error("Failed to fetch stream keys");
    return res.json();
  },

  async createStreamKey(name: string) {
    const res = await fetch(`${API_BASE}/v1/stream-keys`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    if (!res.ok) throw new Error("Failed to create stream key");
    return res.json();
  },

  // Live Streams
  async getLiveStreams() {
    const res = await fetch(`${API_BASE}/v1/live`);
    if (!res.ok) throw new Error("Failed to fetch live streams");
    return res.json();
  },
};

export { API_BASE };
