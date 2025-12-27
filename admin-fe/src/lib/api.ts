const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";


interface SetupAdminData {
  name: string;
  email: string;
  password: string;
}

interface ChangePasswordData {
  current_password: string;
  new_password: string;
}

interface CreateUserData {
  name: string;
  email: string;
  password: string;
}


export const api = {
  // Videos
  async getVideos() {
    const res = await fetchWithAuth(`${API_BASE}/v1/videos`);
    if (!res.ok) throw new Error("Failed to fetch videos");
    return res.json();
  },

  async deleteVideo(id: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/videos/${id}`, { method: "DELETE" });
    if (!res.ok) throw new Error("Failed to delete video");
    return res.json();
  },

  async getVideoPlaybackUrl(id: string) {
    return `${API_BASE}/v1/videos/${id}/playback/hls/master.m3u8`;
  },

  // Upload
  async createUpload(fileName: string, libraryId?: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ file_name: fileName, library_id: libraryId }),
    });
    if (!res.ok) throw new Error("Failed to create upload");
    return res.json();
  },

  async startTranscoding(videoId: string, inputFile: string, libraryId?: string, transcribe = false) {
    const res = await fetchWithAuth(`${API_BASE}/v1/jobs/transcoding`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        video_id: videoId,
        input_file: inputFile,
        library_id: libraryId,
        resolutions: [360, 720, 1080],
        transcribe,
      }),
    });
    if (!res.ok) throw new Error("Failed to start transcoding");
    return res.json();
  },

  async getJob(jobId: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/jobs/${jobId}`);
    if (!res.ok) throw new Error("Failed to get job status");
    return res.json();
  },

  // Stream Keys
  async getStreamKeys() {
    const res = await fetchWithAuth(`${API_BASE}/v1/stream-keys`);
    if (!res.ok) throw new Error("Failed to fetch stream keys");
    return res.json();
  },

  async createStreamKey(name: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/stream-keys`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    if (!res.ok) throw new Error("Failed to create stream key");
    return res.json();
  },

  // Live Streams
  async getLiveStreams() {
    const res = await fetchWithAuth(`${API_BASE}/v1/live`);
    if (!res.ok) throw new Error("Failed to fetch live streams");
    return res.json();
  },

  // Auth
  async login(email: string, password: string) {
    const res = await fetch(`${API_BASE}/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
       const err = await res.json();
       throw new Error(err.error || "Login failed");
    }
    return res.json();
  },

  async getMe() {
    const res = await fetchWithAuth(`${API_BASE}/v1/auth/me`);
    if (!res.ok) throw new Error("Failed to get current user");
    return res.json();
  },

  async setupStatus() {
    const res = await fetch(`${API_BASE}/v1/setup/status`);
    if (!res.ok) throw new Error("Failed to check setup status");
    return res.json();
  },



  async setupAdmin(data: SetupAdminData) {
    const res = await fetch(`${API_BASE}/v1/setup/admin`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error("Failed to setup admin");
    return res.json();
  },

  async changePassword(data: ChangePasswordData) {
      const res = await fetchWithAuth(`${API_BASE}/v1/auth/change-password`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
      });
      if (!res.ok) {
          const err = await res.json();
          throw new Error(err.error || "Failed to change password");
      }
      return res.json();
  },

  // User Management
  async listUsers() {
      const res = await fetchWithAuth(`${API_BASE}/v1/users`);
      if (!res.ok) throw new Error("Failed to list users");
      return res.json();
  },

  async createUser(data: CreateUserData) {

      const res = await fetch(`${API_BASE}/v1/auth/register`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
      });
      if (!res.ok) {
          const err = await res.json();
          throw new Error(err.error || "Failed to create user");
      }
      return res.json();
  },


  async deleteUser(id: string) {
      const res = await fetchWithAuth(`${API_BASE}/v1/users/${id}`, { method: "DELETE" });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to delete user");
      }
      return res.json();
  },

  // Libraries
  async listLibraries() {
    const res = await fetchWithAuth(`${API_BASE}/v1/libraries`);
    if (!res.ok) throw new Error("Failed to list libraries");
    return res.json();
  },

  async createLibrary(name: string, description: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/libraries`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description }),
    });
    if (!res.ok) throw new Error("Failed to create library");
    return res.json();
  },

  async getLibrary(id: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/libraries/${id}`);
    if (!res.ok) throw new Error("Failed to get library");
    return res.json();
  },

  async deleteLibrary(id: string) {
    const res = await fetchWithAuth(`${API_BASE}/v1/libraries/${id}`, {
      method: "DELETE",
    });
    if (!res.ok) throw new Error("Failed to delete library");
    return res.json();
  }
};

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem("auth_token");
  const headers = {
    ...options.headers,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
  return fetch(url, { ...options, headers });
}

export { API_BASE };
