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

  useEffect(() => {
    if (subtitleUrl && !state.isLoading) {
      actions.loadCaptions([{ label: "English", lang: "en", src: subtitleUrl }]);
    }
  }, [subtitleUrl, state.isLoading]);

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
