export interface Artist {
  id: number;
  name: string;
  bio?: string;
  image_url?: string;
  created_at?: string;
  updated_at?: string;
}

export interface Album {
  id: number;
  title: string;
  year?: number;
  release_date?: string;
  cover_url?: string;
  cover_source?: 'local' | 's3';
  status: 'pending' | 'approved' | 'rejected';
  uploaded_by?: number;
  artists?: Artist[];
  created_at?: string;
  updated_at?: string;
}

export interface Song {
  id: number;
  title: string;
  artist: string;
  album: string;
  album_id: number;
  year: number;
  release_date: string;
  lyrics: string;
  audio_url: string;
  audio_source?: 'local' | 's3';
  cover_url: string;
  cover_source?: 'local' | 's3';
  track_number?: number;
  status: 'pending' | 'approved' | 'rejected';
  uploaded_by?: number;
  artists?: Artist[];
}

export interface SongCorrection {
  id: number;
  song_id: number;
  song?: Song;
  user_id?: number;
  user?: User;
  status: 'pending' | 'approved' | 'rejected';
  field_name: string;
  current_value: string;
  corrected_value: string;
  reason?: string;
  created_at: string;
  approved_at?: string;
  approved_by?: number;
  rejected_at?: string;
  rejected_by?: number;
}

export interface AlbumCorrection {
  id: number;
  album_id: number;
  album?: Album;
  user_id?: number;
  user?: User;
  status: 'pending' | 'approved' | 'rejected';
  original_title?: string;
  original_cover_url?: string;
  original_release_date?: string;
  original_artist_ids?: string;
  corrected_title?: string;
  corrected_cover_url?: string;
  corrected_cover_source?: 'local' | 's3';
  corrected_release_date?: string;
  corrected_artist_ids?: string;
  reason?: string;
  created_at: string;
  approved_at?: string;
  approved_by?: number;
  rejected_at?: string;
  rejected_by?: number;
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
  id?: number;
  username: string;
  email: string;
  role?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
}
