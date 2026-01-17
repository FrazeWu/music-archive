
export interface Song {
  id: number;
  title: string;
  artist: string;
  album: string;
  year: number;
  release_date: string;
  lyrics: string;
  audio_url: string;
  cover_url: string;
}

export type RepeatMode = 'none' | 'one' | 'all';

export interface PlayerState {
  currentSong: Song | null;
  isPlaying: boolean;
  isShuffled: boolean;
  repeatMode: RepeatMode;
  volume: number;
  currentTime: number;
  duration: number;
}

export interface User {
  username: string;
  email: string;
  role?: string;
}

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
}

export interface PendingSong extends Song {
  uploaded_by?: number;
  user?: User;
  status: 'pending' | 'approved' | 'rejected';
}
