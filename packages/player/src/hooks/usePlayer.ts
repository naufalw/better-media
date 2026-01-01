import Hls from "hls.js";
import { useCallback, useEffect, useRef, useState } from "react";

interface UsePlayerOptions {
  src: string;
  autoPlay?: boolean;
  muted?: boolean;
  onPlay?: () => void;
  onPause?: () => void;
  onEnded?: () => void;
  onTimeUpdate?: (time: number) => void;
  onError?: (error: Error) => void;
}

export function usePlayer(options: UsePlayerOptions) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(options.muted ?? false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [buffered, setBuffered] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [playbackRate, setPlaybackRate] = useState(1);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !options.src) return;

    if (Hls.isSupported()) {
      const hls = new Hls();
      hlsRef.current = hls;

      hls.loadSource(options.src);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        setIsLoading(false);
        if (options.autoPlay) video.play();
      });

      hls.on(Hls.Events.ERROR, (_, data) => {
        if (data.fatal) {
          options.onError?.(new Error(data.details));
        }
      });

      return () => {
        hls.destroy();
        hlsRef.current = null;
      };
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // here is safari native

      video.src = options.src;

      video.addEventListener("loadedmetadata", () => {
        setIsLoading(false);
        if (options.autoPlay) video.play();
      });
    }
  }, [options.src]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const onPlay = () => {
      setIsPlaying(true);
      options.onPlay?.();
    };
    const onPause = () => {
      setIsPlaying(false);
      options.onPause?.();
    };
    const onEnded = () => {
      setIsPlaying(false);
      options.onEnded?.();
    };
    const onTimeUpdate = () => {
      setCurrentTime(video.currentTime);
      options.onTimeUpdate?.(video.currentTime);
    };
    const onDurationChange = () => setDuration(video.duration);
    const onProgress = () => {
      if (video.buffered.length > 0) {
        setBuffered(video.buffered.end(video.buffered.length - 1));
      }
    };

    video.addEventListener("play", onPlay);
    video.addEventListener("pause", onPause);
    video.addEventListener("ended", onEnded);
    video.addEventListener("timeupdate", onTimeUpdate);
    video.addEventListener("durationchange", onDurationChange);
    video.addEventListener("progress", onProgress);

    return () => {
      video.removeEventListener("play", onPlay);
      video.removeEventListener("pause", onPause);
      video.removeEventListener("ended", onEnded);
      video.removeEventListener("timeupdate", onTimeUpdate);
      video.removeEventListener("durationchange", onDurationChange);
      video.removeEventListener("progress", onProgress);
    };
  }, []);

  // action defintion
  const play = useCallback(() => videoRef.current?.play(), []);
  const pause = useCallback(() => videoRef.current?.pause(), []);
  const toggle = useCallback(() => {
    isPlaying ? pause() : play();
  }, [isPlaying]);

  const seek = useCallback((time: number) => {
    if (videoRef.current) videoRef.current.currentTime = time;
  }, []);

  const setSpeed = useCallback((rate: number) => {
    if (videoRef.current) {
      videoRef.current.playbackRate = rate;
      setPlaybackRate(rate);
    }
  }, []);

  const setVolumeLevel = useCallback((level: number) => {
    if (videoRef.current) {
      videoRef.current.volume = level;
      setVolume(level);
    }
  }, []);

  const toggleMute = useCallback(() => {
    if (videoRef.current) {
      videoRef.current.muted = !videoRef.current.muted;
      setIsMuted(videoRef.current.muted);
    }
  }, []);

  const toggleFullscreen = useCallback(() => {
    const container = videoRef.current?.parentElement;
    if (!container) return;

    if (document.fullscreenElement) {
      document.exitFullscreen();
      setIsFullscreen(false);
    } else {
      container.requestFullscreen();
      setIsFullscreen(true);
    }
  }, []);

  return {
    videoRef,
    state: {
      isPlaying,
      currentTime,
      duration,
      volume,
      isMuted,
      isFullscreen,
      buffered,
      isLoading,
      playbackRate,
    },
    actions: {
      play,
      pause,
      toggle,
      seek,
      setVolume: setVolumeLevel,
      toggleMute,
      toggleFullscreen,
      setSpeed,
    },
  };
}
