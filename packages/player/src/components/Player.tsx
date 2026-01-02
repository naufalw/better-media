import { usePlayer } from "../hooks/usePlayer";
import { Controls } from "./Controls";
import "../styles/player.css";
import { useEffect } from "react";

export interface PlayerProps {
  src: string;
  poster?: string;
  autoPlay?: boolean;
  muted?: boolean;
  controls?: boolean;

  theme?: "light" | "dark";
  accentColor?: string;
  borderRadius?: number;

  onPlay?: () => void;
  onPause?: () => void;
  onEnded?: () => void;
  onTimeUpdate?: (time: number) => void;
  onError?: (error: Error) => void;

  subtitleUrl?: string;
}

export function Player({
  src,
  poster,
  autoPlay = false,
  muted = false,
  controls = true,
  theme = "dark",
  accentColor = "#10b981", // this is better-media emerald color
  borderRadius = 0,
  onPlay,
  onPause,
  onEnded,
  onTimeUpdate,
  onError,
  subtitleUrl,
}: PlayerProps) {
  const { videoRef, state, actions } = usePlayer({
    src,
    autoPlay,
    muted,
    onPlay,
    onPause,
    onEnded,
    onTimeUpdate,
    onError,
  });

  // loading subtitle, wait for load first otherwise it will not be there
  useEffect(() => {
    if (subtitleUrl && !state.isLoading) {
      actions.loadCaptions([{ label: "English", lang: "en", src: subtitleUrl }]);
    }
  }, [subtitleUrl, state.isLoading]);

  // keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      switch (e.key) {
        case " ":
        case "k":
          e.preventDefault();
          actions.toggle();
          break;
        case "ArrowLeft":
          e.preventDefault();
          actions.seek(Math.max(0, state.currentTime - 5));
          break;
        case "ArrowRight":
          e.preventDefault();
          actions.seek(Math.min(state.duration, state.currentTime + 5));
          break;
        case "ArrowUp":
          e.preventDefault();
          actions.setVolume(Math.min(1, state.volume + 0.1));
          break;
        case "ArrowDown":
          e.preventDefault();
          actions.setVolume(Math.max(0, state.volume - 0.1));
          break;
        case "m":
          actions.toggleMute();
          break;
        case "f":
          actions.toggleFullscreen();
          break;
        case "c":
          actions.toggleCaptions();
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [state.currentTime, state.duration, state.volume, actions]);

  return (
    <div
      className={`bm-player ${theme}`}
      style={
        {
          "--bm-accent": accentColor,
          "--bm-radius": `${borderRadius}px`,
        } as React.CSSProperties
      }
    >
      <video
        ref={videoRef}
        className="bm-player-video"
        poster={poster}
        playsInline
        crossOrigin="anonymous"
        onClick={actions.toggle}
      />

      {state.isLoading && (
        <div className="bm-player-loader">
          <div className="bm-player-spinner" />
        </div>
      )}

      {controls && <Controls state={state} actions={actions} />}
    </div>
  );
}
