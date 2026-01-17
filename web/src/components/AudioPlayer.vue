
<script setup lang="ts">
import { computed } from 'vue';
import { usePlayerStore } from '@/stores/player.ts';

const player = usePlayerStore();

const formatTime = (time: number) => {
  const mins = Math.floor(time / 60);
  const secs = Math.floor(time % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
};

const repeatLabel = computed(() => {
  const map: Record<string, string> = {
    none: '不循环',
    all: '列表循环',
    one: '单曲循环'
  };
  return map[player.repeatMode];
});

const handleSeek = (e: MouseEvent) => {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  const x = e.clientX - rect.left;
  const pct = x / rect.width;
  player.seek(pct * player.duration);
};
</script>

<template>
  <footer v-if="player.currentSong" class="fixed bottom-0 w-full bg-white border-t-2 border-black shadow-[0_-10px_25px_-5px_rgba(0,0,0,0.1)] z-50 p-4">
    <div class="max-w-6xl mx-auto flex flex-col gap-3">
      <div class="flex items-center gap-6">
        <div class="flex items-center gap-3">
          <img :src="player.currentSong.cover_url || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%2250%22 height=%2250%22%3E%3Crect width=%2250%22 height=%2250%22 fill=%22%23000%22/%3E%3C/svg%3E'" 
               :alt="player.currentSong.title" 
               class="w-14 h-14 border border-black grayscale flex-shrink-0" />
          <div class="overflow-hidden min-w-[180px]">
            <h4 class="font-bold truncate">{{ player.currentSong.title }}</h4>
            <p class="text-xs text-gray-500 truncate">{{ player.currentSong.artist }}</p>
          </div>
        </div>

        <div class="flex items-center gap-[5px] flex-1 justify-center">
          <button 
            @click="player.toggleShuffle"
            :class="['px-3 py-1 text-xs border border-black hover:bg-black hover:text-white transition-colors', player.isShuffled ? 'bg-black text-white' : '']"
          >
            随机
          </button>
          <button 
            @click="player.playPrevious"
            class="px-3 py-1 text-xs border border-black hover:bg-black hover:text-white transition-colors"
          >
            上一首
          </button>
          <button 
            @click="player.togglePlay"
            class="px-6 py-2 text-sm font-black border-2 border-black bg-black text-white hover:bg-white hover:text-black transition-all"
          >
            {{ player.isPlaying ? '暂停' : '播放' }}
          </button>
          <button 
            @click="player.playNext"
            class="px-3 py-1 text-xs border border-black hover:bg-black hover:text-white transition-colors"
          >
            下一首
          </button>
          <button 
            @click="player.toggleRepeat"
            class="px-3 py-1 text-xs border border-black hover:bg-black hover:text-white transition-colors min-w-[80px]"
          >
            {{ repeatLabel }}
          </button>
        </div>

        <div class="flex items-center gap-3 min-w-[140px] justify-end">
          <span class="text-xs font-bold">音量</span>
          <input 
            type="range" 
            min="0" 
            max="1" 
            step="0.01" 
            :value="player.volume" 
            @input="(e) => player.setVolume(parseFloat((e.target as HTMLInputElement).value))"
            class="accent-black h-1 bg-gray-200 cursor-pointer w-20"
          />
        </div>
      </div>

      <div class="flex items-center gap-4 text-xs font-bold">
        <span class="w-10 text-right">{{ formatTime(player.currentTime) }}</span>
        <div class="relative flex-1 h-1 bg-gray-200 cursor-pointer" @click="handleSeek">
          <div 
            class="absolute top-0 left-0 h-full bg-black transition-all"
            :style="{ width: `${(player.currentTime / player.duration) * 100}%` }"
          />
        </div>
        <span class="w-10">{{ formatTime(player.duration) }}</span>
      </div>
    </div>
  </footer>
</template>
