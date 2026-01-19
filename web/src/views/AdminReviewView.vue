<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useApi } from '@/composables/useApi';
import type { Song, User } from '@/types';

// 统一的审核项接口
interface ReviewItem {
  id: string; // 唯一标识
  type: 'song_batch' | 'song_correction' | 'album_correction' | 'album_upload';
  created_at: string;
  user: User;
  
  // 歌曲批次上传
  batch_id?: string;
  songs?: Song[];
  song_ids?: number[];
  
  // 纠错
  correction_id?: number;
  field_name?: string;
  current_value?: string;
  corrected_value?: string;
  reason?: string;
  target_title?: string; // 歌曲或专辑名称
  
  // 元数据
  artist?: string;
  album?: string;
  album_id?: number;
  cover_url?: string;
  cover_source?: 'local' | 's3';
  count?: number;
}

const authStore = useAuthStore();
const api = useApi();
const reviewItems = ref<ReviewItem[]>([]);
const loading = ref(true);
const processingItems = ref<Set<string>>(new Set());

// 获取所有待审核内容并统一格式
const fetchAllPendingItems = async () => {
  try {
    const headers = { 'Authorization': `Bearer ${authStore.token}` };
    
    // 并行获取所有待审核数据
    const [songsRes, songCorrectionsRes, albumCorrectionsRes, albumsRes] = await Promise.all([
      fetch(`${api.url}/admin/pending`, { headers }),
      fetch(`${api.url}/admin/pending-song-corrections`, { headers }),
      fetch(`${api.url}/admin/pending-album-corrections`, { headers }),
      fetch(`${api.url}/admin/pending-albums`, { headers })
    ]);

    const items: ReviewItem[] = [];

    // 处理歌曲批次上传 - 按 batch_id 分组
    if (songsRes.ok) {
      const songs = await songsRes.json();
      if (Array.isArray(songs) && songs.length > 0) {
        // 按 batch_id 分组
        const batchGroups = new Map<string, any[]>();
        
        songs.forEach((song: any) => {
          const batchId = song.batch_id || `single_${song.id}`;
          if (!batchGroups.has(batchId)) {
            batchGroups.set(batchId, []);
          }
          batchGroups.get(batchId)!.push(song);
        });
        
        // 为每个批次创建一个 ReviewItem
        batchGroups.forEach((batchSongs, batchId) => {
          // 使用第一首歌的信息作为批次信息
          const firstSong = batchSongs[0];
          
          items.push({
            id: `batch_${batchId}`,
            type: 'song_batch',
            batch_id: batchId,
            created_at: firstSong.created_at,
            user: firstSong.user || { username: 'Unknown' },
            songs: batchSongs,
            song_ids: batchSongs.map(s => s.id),
            artist: firstSong.artist,
            album: firstSong.album,
            count: batchSongs.length,
            cover_url: firstSong.cover_url,
            cover_source: firstSong.cover_source,
            target_title: firstSong.album // 使用专辑名作为标题
          });
        });
      }
    }

    // 处理歌曲纠错
    if (songCorrectionsRes.ok) {
      const corrections = await songCorrectionsRes.json();
      corrections.forEach((corr: any) => {
        items.push({
          id: `song_corr_${corr.id}`,
          type: 'song_correction',
          created_at: corr.created_at,
          user: corr.user,
          correction_id: corr.id,
          field_name: corr.field_name,
          current_value: corr.current_value,
          corrected_value: corr.corrected_value,
          reason: corr.reason,
          target_title: corr.song?.title,
          artist: corr.song?.artist,
          album: corr.song?.album
        });
      });
    }

    // 处理专辑纠错（包括封面和元数据）
    if (albumCorrectionsRes.ok) {
      const albumCorr = await albumCorrectionsRes.json();
      albumCorr.forEach((corr: any) => {
        const artistName = corr.album?.artists?.length > 0 
          ? corr.album.artists[0].name 
          : 'Unknown Artist';
        
        items.push({
          id: `album_corr_${corr.id}`,
          type: 'album_correction',
          created_at: corr.created_at,
          user: corr.user,
          correction_id: corr.id,
          field_name: corr.corrected_cover_url ? 'cover_url' : 'title',
          current_value: corr.original_title || corr.original_cover_url,
          corrected_value: corr.corrected_title || corr.corrected_cover_url,
          reason: corr.reason,
          target_title: corr.album?.title,
          artist: artistName,
          cover_url: corr.corrected_cover_url || corr.album?.cover_url
        });
      });
    }

    // 处理专辑上传
    if (albumsRes.ok) {
      const albums = await albumsRes.json();
      albums.forEach((album: any) => {
        const artistName = album.artists?.length > 0 
          ? album.artists[0].name 
          : 'Unknown Artist';
        
        items.push({
          id: `album_${album.id}`,
          type: 'album_upload',
          created_at: album.created_at,
          user: album.user || { username: 'Unknown' },
          album_id: album.id,
          target_title: album.title,
          artist: artistName,
          cover_url: album.cover_url,
          cover_source: album.cover_source
        });
      });
    }

    // 按时间倒序排列（最新的在前）
    items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    
    reviewItems.value = items;
  } catch (e) {
    console.error('Failed to fetch pending items:', e);
  } finally {
    loading.value = false;
  }
};

// 批准歌曲批次
const approveSongBatch = async (item: ReviewItem) => {
  if (!item.song_ids || item.song_ids.length === 0) return;
  
  // 防止重复点击
  if (processingItems.value.has(item.id)) return;
  
  processingItems.value.add(item.id);
  
  try {
    let failedCount = 0;
    
    // 循环批准每首歌
    for (const songId of item.song_ids) {
      const response = await fetch(`${api.url}/admin/approve/${songId}`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authStore.token}` }
      });
      
      if (!response.ok) {
        failedCount++;
      }
    }
    
    if (failedCount > 0) {
      alert(`批准失败：${failedCount} 首歌曲处理失败`);
    } else {
      // 成功后从列表移除
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    }
  } catch (e) {
    console.error('批准歌曲批次失败:', e);
    alert('操作失败，请重试');
  } finally {
    processingItems.value.delete(item.id);
  }
};

// 驳回歌曲批次
const rejectSongBatch = async (item: ReviewItem) => {
  if (!item.song_ids || item.song_ids.length === 0) return;
  
  // 防止重复点击
  if (processingItems.value.has(item.id)) return;
  
  processingItems.value.add(item.id);
  
  try {
    let failedCount = 0;
    
    // 循环驳回每首歌
    for (const songId of item.song_ids) {
      const response = await fetch(`${api.url}/admin/reject/${songId}`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authStore.token}` }
      });
      
      if (!response.ok) {
        failedCount++;
      }
    }
    
    if (failedCount > 0) {
      alert(`驳回失败：${failedCount} 首歌曲处理失败`);
    } else {
      // 成功后从列表移除
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    }
  } catch (e) {
    console.error('驳回歌曲批次失败:', e);
    alert('操作失败，请重试');
  } finally {
    processingItems.value.delete(item.id);
  }
};

// 批准纠错
const approveCorrection = async (item: ReviewItem) => {
  const endpoint = item.type === 'song_correction' 
    ? `${api.url}/admin/approve-song-correction/${item.correction_id}`
    : `${api.url}/admin/approve-album-correction/${item.correction_id}`;
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 驳回纠错
const rejectCorrection = async (item: ReviewItem) => {
  const endpoint = item.type === 'song_correction'
    ? `${api.url}/admin/reject-song-correction/${item.correction_id}`
    : `${api.url}/admin/reject-album-correction/${item.correction_id}`;
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 批准专辑上传
const approveAlbum = async (item: ReviewItem) => {
  try {
    const response = await fetch(`${api.url}/admin/approve-album/${item.album_id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 驳回专辑上传
const rejectAlbum = async (item: ReviewItem) => {
  try {
    const response = await fetch(`${api.url}/admin/reject-album/${item.album_id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    
    if (response.ok) {
      reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
    } else {
      alert('操作失败');
    }
  } catch (e) {
    alert('操作失败');
  }
};

// 试听歌曲
const playAudio = (url: string) => {
  const audio = new Audio(url);
  audio.play();
};

// 类型标签文本
const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    'song_batch': '歌曲上传',
    'song_correction': '歌曲纠错',
    'album_correction': '专辑修正',
    'album_upload': '专辑上传'
  };
  return labels[type] || type;
};

// 字段名称中文映射
const getFieldLabel = (field: string) => {
  const labels: Record<string, string> = {
    'title': '标题',
    'artist': '艺术家',
    'album': '专辑',
    'year': '年份',
    'release_date': '发行日期',
    'lyrics': '歌词',
    'cover_url': '专辑封面',
    'track_number': '曲目编号'
  };
  return labels[field] || field;
};

// 格式化日期
const formatDate = (dateString: string | undefined) => {
  if (!dateString) return '未知';
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// 检查是否本地存储
const isLocalStorage = (item: any) => {
  // Check if any song in the batch has local storage
  if (item.songs && item.songs.length > 0) {
    return item.songs.some((song: any) => 
      song.audio_source === 'local' || 
      song.cover_source === 'local' ||
      (song.audio_url && song.audio_url.startsWith('/uploads/'))
    );
  }
  // Check for album uploads
  if (item.type === 'album_upload' && item.cover_url) {
    return item.cover_url.startsWith('/uploads/');
  }
  return false;
};

// 格式化曲目编号
const formatTrackNumber = (num: number | null | undefined) => {
  if (!num) return '—';
  return String(num).padStart(2, '0');
};

onMounted(async () => {
  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    alert('需要管理员权限');
    return;
  }
  await fetchAllPendingItems();
});
</script>

<template>
  <div class="max-w-5xl mx-auto px-8 py-20">
    <div class="mb-12">
      <h1 class="text-4xl font-black tracking-tighter mb-4">审核队列</h1>
      <p class="text-gray-500">
        审核用户提交的内容修正和新增。共 <strong>{{ reviewItems.length }}</strong> 项待审核。
      </p>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-20">
      <p class="text-gray-400 font-medium">加载中...</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="reviewItems.length === 0" 
      class="text-center py-20 bg-gray-50 border-2 border-dashed border-gray-200">
      <p class="text-gray-400 font-medium">暂无待审核内容</p>
    </div>

    <!-- 审核列表 -->
    <div v-else class="space-y-6">
      <div 
        v-for="item in reviewItems" 
        :key="item.id"
        class="bg-white border-2 border-black p-8 hover:shadow-[15px_15px_0px_0px_rgba(0,0,0,1)] transition-shadow"
      >
        <!-- 头部：类型标签 + 元信息 -->
        <div class="flex justify-between items-start mb-6 pb-4 border-b-2 border-gray-100">
          <div class="flex-1">
            <div class="flex items-center gap-3 mb-2">
              <span class="bg-black text-white px-3 py-1 text-xs font-black uppercase tracking-widest">
                {{ getTypeLabel(item.type) }}
              </span>
              <span class="text-gray-400 text-sm">
                {{ new Date(item.created_at).toLocaleString('zh-CN') }}
              </span>
            </div>
            
            <h2 class="text-2xl font-black mb-1">
              {{ item.target_title || item.album || '未命名' }}
            </h2>
            
            <p class="text-gray-500 font-bold">
              {{ item.artist }}
            </p>
            
            <p class="text-sm text-gray-400 mt-2">
              提交者: {{ item.user?.username }}
            </p>
          </div>

          <!-- 操作按钮 -->
          <div class="flex gap-3">
            <button 
              @click="
                item.type === 'song_batch' ? approveSongBatch(item) :
                item.type === 'album_upload' ? approveAlbum(item) :
                approveCorrection(item)
              "
              :disabled="processingItems.has(item.id)"
              class="bg-black text-white px-6 py-3 font-bold hover:bg-green-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ processingItems.has(item.id) ? '处理中...' : '通过' }}
            </button>
            <button 
              @click="
                item.type === 'song_batch' ? rejectSongBatch(item) :
                item.type === 'album_upload' ? rejectAlbum(item) :
                rejectCorrection(item)
              "
              :disabled="processingItems.has(item.id)"
              class="bg-white text-black border-2 border-black px-6 py-3 font-bold hover:bg-red-600 hover:text-white hover:border-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ processingItems.has(item.id) ? '处理中...' : '驳回' }}
            </button>
          </div>
        </div>

        <!-- 内容区：根据类型显示不同内容 -->
        
        <!-- 歌曲批次上传 -->
        <div v-if="item.type === 'song_batch'" class="space-y-4">
          <div class="flex items-center gap-6 mb-4">
            <div v-if="item.cover_url" class="w-32 h-32 border-2 border-black shrink-0">
              <img :src="item.cover_url" class="w-full h-full object-cover grayscale" alt="专辑封面" />
            </div>
            <div v-else class="w-32 h-32 border-2 border-gray-300 bg-gray-100 flex items-center justify-center shrink-0">
              <span class="text-gray-400 font-bold text-xs">无封面</span>
            </div>
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-1">专辑信息</p>
              <p class="text-xl font-black mb-1">{{ item.album }}</p>
              <p class="font-bold text-gray-600">{{ item.artist }}</p>
              <p class="text-sm text-gray-500 mt-2">
                <span class="font-black">{{ item.count }}</span> 首歌曲
              </p>
            </div>
          </div>

          <!-- 上传信息 -->
          <div class="grid grid-cols-3 gap-4 p-4 bg-gray-50 border border-gray-200 text-sm">
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">上传时间</p>
              <p class="font-bold">{{ formatDate(item.created_at) }}</p>
            </div>
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">上传者</p>
              <p class="font-bold">{{ item.user?.username || '匿名用户' }}</p>
            </div>
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">存储位置</p>
              <div class="flex items-center gap-2">
                <span v-if="isLocalStorage(item)" class="inline-block px-2 py-1 bg-yellow-100 text-yellow-800 text-xs font-black">
                  🟡 本地暂存
                </span>
                <span v-else class="inline-block px-2 py-1 bg-blue-100 text-blue-800 text-xs font-black">
                  🔵 云端存储
                </span>
              </div>
            </div>
          </div>

          <!-- 歌曲列表 -->
          <div class="space-y-2">
            <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-2">曲目列表</p>
            <div 
              v-for="song in item.songs" 
              :key="song.id"
              class="flex items-center gap-4 p-3 bg-gray-50 hover:bg-gray-100 transition-colors"
            >
              <span class="text-gray-400 font-mono w-8 text-right text-sm">
                {{ formatTrackNumber(song.track_number) }}
              </span>
              <div class="flex-1 font-bold">{{ song.title }}</div>
              <div class="flex items-center gap-2 text-xs">
                <span 
                  v-if="song.audio_source === 'local'"
                  class="px-2 py-1 bg-yellow-50 text-yellow-700 font-bold border border-yellow-200"
                  title="本地暂存"
                >
                  LOCAL
                </span>
                <span 
                  v-else
                  class="px-2 py-1 bg-blue-50 text-blue-700 font-bold border border-blue-200"
                  title="云端存储"
                >
                  S3
                </span>
              </div>
              <button 
                @click="playAudio(song.audio_url)"
                class="text-xs font-black uppercase tracking-widest bg-white border border-black px-3 py-1 hover:bg-black hover:text-white transition-colors"
              >
                ▶ 试听
              </button>
              <a 
                :href="song.audio_url" 
                target="_blank" 
                class="text-gray-400 hover:text-black text-sm transition-colors"
                title="下载"
              >
                ⬇
              </a>
            </div>
          </div>
        </div>

        <!-- 专辑上传 -->
        <div v-else-if="item.type === 'album_upload'" class="space-y-4">
          <div class="flex items-center gap-6 mb-4">
            <div v-if="item.cover_url" class="w-48 h-48 border-2 border-black shrink-0">
              <img :src="item.cover_url" class="w-full h-full object-cover" alt="专辑封面" />
            </div>
            <div v-else class="w-48 h-48 border-2 border-gray-300 bg-gray-100 flex items-center justify-center shrink-0">
              <span class="text-gray-400 font-bold">无封面</span>
            </div>
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-2">专辑信息</p>
              <p class="text-2xl font-black mb-2">{{ item.target_title }}</p>
              <p class="text-lg font-bold text-gray-600">{{ item.artist }}</p>
            </div>
          </div>

          <!-- 上传信息 -->
          <div class="grid grid-cols-3 gap-4 p-4 bg-gray-50 border border-gray-200 text-sm">
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">上传时间</p>
              <p class="font-bold">{{ formatDate(item.created_at) }}</p>
            </div>
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">上传者</p>
              <p class="font-bold">{{ item.user?.username || '匿名用户' }}</p>
            </div>
            <div>
              <p class="text-xs font-black uppercase tracking-widest text-gray-500 mb-1">存储位置</p>
              <div class="flex items-center gap-2">
                <span v-if="isLocalStorage(item)" class="inline-block px-2 py-1 bg-yellow-100 text-yellow-800 text-xs font-black">
                  🟡 本地暂存
                </span>
                <span v-else class="inline-block px-2 py-1 bg-blue-100 text-blue-800 text-xs font-black">
                  🔵 云端存储
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 纠错：封面类型 -->
        <div v-else-if="item.field_name === 'cover_url'" class="space-y-4">
          <div class="grid grid-cols-2 gap-6">
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-3">原封面</p>
              <div class="border-2 border-gray-300">
                <img 
                  v-if="item.current_value"
                  :src="item.current_value" 
                  class="w-full aspect-square object-cover" 
                  alt="原封面"
                />
                <div v-else class="w-full aspect-square bg-gray-200 flex items-center justify-center">
                  <span class="text-gray-400">无封面</span>
                </div>
              </div>
            </div>
            
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-3">修改后</p>
              <div class="border-2 border-black">
                <img 
                  :src="item.corrected_value" 
                  class="w-full aspect-square object-cover" 
                  alt="新封面"
                />
              </div>
            </div>
          </div>

          <div v-if="item.reason" class="mt-4 p-4 bg-gray-50 border-l-4 border-black">
            <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-1">修改原因</p>
            <p class="text-gray-700">{{ item.reason }}</p>
          </div>
        </div>

        <!-- 纠错：文本类型 -->
        <div v-else class="space-y-4">
          <div class="grid grid-cols-2 gap-6">
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-3">
                原 {{ getFieldLabel(item.field_name || '') }}
              </p>
              <div class="p-4 bg-gray-50 border-2 border-gray-300 font-mono text-sm min-h-20">
                {{ item.current_value || '（无）' }}
              </div>
            </div>
            
            <div>
              <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-3">
                修改后 {{ getFieldLabel(item.field_name || '') }}
              </p>
              <div class="p-4 bg-green-50 border-2 border-black font-mono text-sm min-h-20">
                {{ item.corrected_value }}
              </div>
            </div>
          </div>

          <div v-if="item.reason" class="p-4 bg-gray-50 border-l-4 border-black">
            <p class="text-sm font-black uppercase tracking-widest text-gray-500 mb-1">修改原因</p>
            <p class="text-gray-700">{{ item.reason }}</p>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
/* 额外样式如果需要 */
</style>
