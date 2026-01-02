import { useState } from "react";

interface ControlsProps {
  state: {
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    volume: number;
    isMuted: boolean;
    isFullscreen: boolean;
    buffered: number;
    playbackRate: number;
    qualities: {
      height: number;
      index: number;
    }[];
    currentQuality: number;
    captionsEnabled: boolean;
    captionTracks: { label: string; lang: string; src: string }[];
  };
  actions: {
    toggle: () => void;
    seek: (time: number) => void;
    setVolume: (level: number) => void;
    toggleMute: () => void;
    toggleFullscreen: () => void;
    setSpeed: (rate: number) => void;
    setQuality: (levelIndex: number) => void;
    toggleCaptions: () => void;
  };
}

export function Controls({ state, actions }: ControlsProps) {
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };
  const speeds = [0.5, 0.75, 1, 1.25, 1.5, 2];
  const [showSettings, setShowSettings] = useState(false);
  const [settingsTab, setSettingsTab] = useState<"main" | "speed" | "quality">("main");

  const progress = state.duration ? (state.currentTime / state.duration) * 100 : 0;
  const bufferedPercent = state.duration ? (state.buffered / state.duration) * 100 : 0;

  const closeSettings = () => {
    setShowSettings(false);
    setSettingsTab("main");
  };

  return (
    <div className="bm-controls">
      {/* Progress bar */}
      <div
        className="bm-progress-container"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect();
          const percent = (e.clientX - rect.left) / rect.width;
          actions.seek(percent * state.duration);
        }}
      >
        <div className="bm-progress-buffered" style={{ width: `${bufferedPercent}%` }} />
        <div className="bm-progress-bar" style={{ width: `${progress}%` }} />
      </div>

      <div className="bm-controls-row">
        {/* Play/Pause */}
        <button className="bm-control-btn" onClick={actions.toggle}>
          {state.isPlaying ? <PauseIcon /> : <PlayIcon />}
        </button>

        {/* Volume */}
        <button className="bm-control-btn" onClick={actions.toggleMute}>
          {state.isMuted ? <MutedIcon /> : <VolumeIcon />}
        </button>

        {/* Time */}
        <span className="bm-time">
          {formatTime(state.currentTime)} / {formatTime(state.duration)}
        </span>

        {/* Spacer */}
        <div className="bm-spacer" />

        {state.captionTracks.length > 0 && (
          <button
            className={`bm-control-btn ${state.captionsEnabled ? "active" : ""}`}
            onClick={actions.toggleCaptions}
          >
            <CaptionsIcon />
          </button>
        )}

        {/* Settings */}
        <div className="bm-menu-container">
          <button className="bm-control-btn" onClick={() => setShowSettings(!showSettings)}>
            <SettingsIcon />
          </button>

          {showSettings && (
            <div className="bm-menu">
              {settingsTab === "main" && (
                <>
                  <button
                    className="bm-menu-item bm-menu-item-nav"
                    onClick={() => setSettingsTab("speed")}
                  >
                    <span>Speed</span>
                    <span className="bm-menu-value">{state.playbackRate}x</span>
                  </button>
                  <button
                    className="bm-menu-item bm-menu-item-nav"
                    onClick={() => setSettingsTab("quality")}
                  >
                    <span>Quality</span>
                    <span className="bm-menu-value">
                      {state.currentQuality === -1
                        ? "Auto"
                        : `${state.qualities.find((q) => q.index === state.currentQuality)?.height}p`}
                    </span>
                  </button>
                </>
              )}

              {settingsTab === "speed" && (
                <>
                  <button
                    className="bm-menu-item bm-menu-back"
                    onClick={() => setSettingsTab("main")}
                  >
                    <ChevronLeftIcon />
                    <span>Speed</span>
                  </button>
                  <div className="bm-menu-divider" />
                  {speeds.map((speed) => (
                    <button
                      key={speed}
                      className={`bm-menu-item ${state.playbackRate === speed ? "active" : ""}`}
                      onClick={() => {
                        actions.setSpeed(speed);
                        closeSettings();
                      }}
                    >
                      {speed}x
                    </button>
                  ))}
                </>
              )}

              {settingsTab === "quality" && (
                <>
                  <button
                    className="bm-menu-item bm-menu-back"
                    onClick={() => setSettingsTab("main")}
                  >
                    <ChevronLeftIcon /> Quality
                  </button>
                  <div className="bm-menu-divider" />
                  <button
                    className={`bm-menu-item ${state.currentQuality === -1 ? "active" : ""}`}
                    onClick={() => {
                      actions.setQuality(-1);
                      closeSettings();
                    }}
                  >
                    Auto
                  </button>
                  {state.qualities.map((q) => (
                    <button
                      key={q.index}
                      className={`bm-menu-item ${state.currentQuality === q.index ? "active" : ""}`}
                      onClick={() => {
                        actions.setQuality(q.index);
                        closeSettings();
                      }}
                    >
                      {q.height}p
                    </button>
                  ))}
                </>
              )}
            </div>
          )}
        </div>

        {/* Fullscreen */}
        <button className="bm-control-btn" onClick={actions.toggleFullscreen}>
          {state.isFullscreen ? <ExitFullscreenIcon /> : <FullscreenIcon />}
        </button>
      </div>
    </div>
  );
}

const PlayIcon = () => (
  // lucide play
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-play-icon lucide-play"
  >
    <path d="M5 5a2 2 0 0 1 3.008-1.728l11.997 6.998a2 2 0 0 1 .003 3.458l-12 7A2 2 0 0 1 5 19z" />
  </svg>
);

const PauseIcon = () => (
  // lucide pause
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-pause-icon lucide-pause"
  >
    <rect x="14" y="3" width="5" height="18" rx="1" />
    <rect x="5" y="3" width="5" height="18" rx="1" />
  </svg>
);

const VolumeIcon = () => (
  // lucide volume 2
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-volume2-icon lucide-volume-2"
  >
    <path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z" />
    <path d="M16 9a5 5 0 0 1 0 6" />
    <path d="M19.364 18.364a9 9 0 0 0 0-12.728" />
  </svg>
);

const MutedIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-volume-x-icon lucide-volume-x"
  >
    <path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z" />
    <line x1="22" x2="16" y1="9" y2="15" />
    <line x1="16" x2="22" y1="9" y2="15" />
  </svg>
);

const FullscreenIcon = () => (
  // this is lucide maximize
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-maximize-icon lucide-maximize"
  >
    <path d="M8 3H5a2 2 0 0 0-2 2v3" />
    <path d="M21 8V5a2 2 0 0 0-2-2h-3" />
    <path d="M3 16v3a2 2 0 0 0 2 2h3" />
    <path d="M16 21h3a2 2 0 0 0 2-2v-3" />
  </svg>
);

const ExitFullscreenIcon = () => (
  // lucide minimize
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-minimize-icon lucide-minimize"
  >
    <path d="M8 3v3a2 2 0 0 1-2 2H3" />
    <path d="M21 8h-3a2 2 0 0 1-2-2V3" />
    <path d="M3 16h3a2 2 0 0 1 2 2v3" />
    <path d="M16 21v-3a2 2 0 0 1 2-2h3" />
  </svg>
);

const SettingsIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-settings-icon lucide-settings"
  >
    <path d="M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);

const ChevronLeftIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-chevron-left-icon lucide-chevron-left"
  >
    <path d="m15 18-6-6 6-6" />
  </svg>
);

const CaptionsIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className="lucide lucide-captions-icon lucide-captions"
  >
    <rect width="18" height="14" x="3" y="5" rx="2" ry="2" />
    <path d="M7 15h4M15 15h2M7 11h2M13 11h4" />
  </svg>
);
