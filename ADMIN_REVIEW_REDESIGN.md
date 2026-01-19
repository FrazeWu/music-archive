# 审核队列界面重新设计

## 更新时间
2026-01-18 18:25

## 设计理念

**原设计问题**：
- 使用4个标签页分类（歌曲审核、纠错审核、专辑封面审核、专辑修正审核）
- 用户需要在多个标签页之间切换
- 审核流程分散，效率低

**新设计目标**：
- ✅ 统一展示所有待审核内容，无需分类标签页
- ✅ 参考上传界面的视觉风格
- ✅ 清晰标注"原信息"和"修改后信息"
- ✅ 简洁的"通过/驳回"操作

---

## 新界面特性

### 1. 统一列表视图

所有待审核内容按**时间倒序**排列在一个页面中：
- 最新提交的在最上方
- 每项都有类型标签（歌曲上传 / 歌曲纠错 / 专辑修正）
- 一目了然查看所有待处理任务

### 2. 三种内容类型

#### 类型A：歌曲批次上传
```
┌─────────────────────────────────────────┐
│ [歌曲上传] 2026-01-18 18:00            │
│ 2049                                    │
│ Kanye West                              │
│ 提交者: fazong                          │
│                                         │
│ [专辑封面缩略图] 共 12 首歌曲          │
│                                         │
│ 曲目列表：                              │
│  1  INTRO (CHARLIE HEAT VERSION) [▶试听]│
│  2  FATHER STRETCH MY HANDS PT. 1 [▶]  │
│  ...                                   │
│                                         │
│              [通过]  [驳回]             │
└─────────────────────────────────────────┘
```

**显示内容**：
- 专辑名称、艺术家、提交者、时间
- 专辑封面缩略图
- 完整曲目列表（带试听和下载按钮）
- 总曲目数

**操作**：
- 通过 → 调用 `/api/admin/approve-batch`
- 驳回 → 调用 `/api/admin/reject-batch`（删除所有音频文件）

---

#### 类型B：文本字段纠错
```
┌─────────────────────────────────────────┐
│ [歌曲纠错] 2026-01-18 17:30            │
│ Runaway                                 │
│ Kanye West                              │
│ 提交者: user123                         │
│                                         │
│ ┌─────────────┬─────────────┐          │
│ │原 标题       │修改后 标题   │          │
│ ├─────────────┼─────────────┤          │
│ │Runawy       │Runaway      │          │
│ │(无)         │(正确拼写)    │          │
│ └─────────────┴─────────────┘          │
│                                         │
│ 修改原因: 修正拼写错误                  │
│                                         │
│              [通过]  [驳回]             │
└─────────────────────────────────────────┘
```

**显示内容**：
- 歌曲/专辑名称、艺术家、提交者、时间
- 左右对比：原值 vs 修改后值
- 修改字段名称（标题/艺术家/专辑/年份/歌词等）
- 修改原因（如果有）

**操作**：
- 通过 → 调用 `/api/admin/approve-correction/:id` 或 `/api/admin/approve-album-correction/:id`
- 驳回 → 调用相应的 reject 接口

---

#### 类型C：专辑封面纠错
```
┌─────────────────────────────────────────┐
│ [专辑修正] 2026-01-18 16:45            │
│ My Beautiful Dark Twisted Fantasy       │
│ Kanye West                              │
│ 提交者: admin                           │
│                                         │
│ ┌─────────────┬─────────────┐          │
│ │原封面       │修改后封面    │          │
│ ├─────────────┼─────────────┤          │
│ │ [图片]      │ [图片]      │          │
│ │ (灰度显示)  │ (彩色显示)   │          │
│ └─────────────┴─────────────┘          │
│                                         │
│ 修改原因: 上传高分辨率版本              │
│                                         │
│              [通过]  [驳回]             │
└─────────────────────────────────────────┘
```

**显示内容**：
- 专辑名称、艺术家、提交者、时间
- 左右对比：原封面（灰度）vs 新封面（彩色）
- 修改原因（如果有）

**操作**：
- 通过 → 调用 `/api/admin/approve-album-correction/:id`（更新封面URL，删除旧封面）
- 驳回 → 删除上传的新封面文件

---

## 视觉设计

### 布局风格
参考 `UploadView.vue` 的极简黑白风格：
- 白色背景 + 黑色边框
- 鼠标悬停时出现黑色硬阴影：`shadow-[15px_15px_0px_0px_rgba(0,0,0,1)]`
- 标题使用 `font-black tracking-tighter`
- 标签使用大写字母 + 宽字距：`uppercase tracking-widest`

### 按钮样式
- **通过按钮**：黑底白字，悬停变绿色
  ```css
  bg-black text-white hover:bg-green-600
  ```
- **驳回按钮**：白底黑边，悬停变红色
  ```css
  bg-white text-black border-2 border-black hover:bg-red-600 hover:text-white hover:border-red-600
  ```

### 类型标签
```html
<span class="bg-black text-white px-3 py-1 text-xs font-black uppercase tracking-widest">
  歌曲上传
</span>
```

### 对比区域
- **原值区域**：灰色背景 `bg-gray-50` + 灰色边框 `border-gray-300`
- **修改后区域**：绿色背景 `bg-green-50` + 黑色边框 `border-black`
- 使用 Grid 布局并排显示

---

## 技术实现

### 数据获取流程

```typescript
// 1. 并行获取三种待审核数据
const [batches, songCorrections, albumCorrections] = await Promise.all([
  fetch('/api/admin/pending'),
  fetch('/api/admin/pending-corrections'),
  fetch('/api/admin/pending-album-corrections')
]);

// 2. 统一转换为 ReviewItem 接口
interface ReviewItem {
  id: string;
  type: 'song_batch' | 'song_correction' | 'album_correction';
  created_at: string;
  user: User;
  // ... 其他字段
}

// 3. 按时间倒序排列
items.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
```

### 条件渲染逻辑

```vue
<div v-for="item in reviewItems" :key="item.id">
  <!-- 歌曲批次上传 -->
  <div v-if="item.type === 'song_batch'">
    <!-- 显示曲目列表 + 试听按钮 -->
  </div>
  
  <!-- 封面纠错 -->
  <div v-else-if="item.field_name === 'cover_url'">
    <!-- 显示左右对比图片 -->
  </div>
  
  <!-- 文本字段纠错 -->
  <div v-else>
    <!-- 显示左右对比文本 -->
  </div>
</div>
```

---

## 字段名称映射

后端返回的字段名是英文，前端显示时需要中文化：

```typescript
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
```

---

## API 端点使用

| 操作 | 端点 | 方法 | 请求体 |
|------|------|------|--------|
| 批准歌曲批次 | `/api/admin/approve-batch` | POST | `{ ids: [1,2,3] }` |
| 驳回歌曲批次 | `/api/admin/reject-batch` | POST | `{ ids: [1,2,3] }` |
| 批准歌曲纠错 | `/api/admin/approve-correction/:id` | POST | 无 |
| 驳回歌曲纠错 | `/api/admin/reject-correction/:id` | POST | 无 |
| 批准专辑纠错 | `/api/admin/approve-album-correction/:id` | POST | 无 |
| 驳回专辑纠错 | `/api/admin/reject-album-correction/:id` | POST | 无 |

---

## 用户体验优化

### 1. 操作确认
所有批准/驳回操作都需要确认：
```typescript
if (!confirm('确认通过此修正？')) return;
```

### 2. 即时反馈
操作成功后：
- 从列表中移除该项（无需刷新页面）
- 显示 `alert('已批准')` 或 `alert('已驳回')`

### 3. 试听功能
歌曲批次上传中每首歌都有试听按钮：
```typescript
const playAudio = (url: string) => {
  const audio = new Audio(url);
  audio.play();
};
```

### 4. 空状态
没有待审核内容时显示友好提示：
```vue
<div class="text-center py-20 bg-gray-50 border-2 border-dashed border-gray-200">
  <p class="text-gray-400 font-medium">暂无待审核内容</p>
</div>
```

---

## 与旧版对比

| 特性 | 旧版（标签页设计） | 新版（统一列表） |
|------|-------------------|-----------------|
| 内容分类 | 4个标签页 | 统一列表 + 类型标签 |
| 查看效率 | 需要切换标签 | 一次查看全部 |
| 视觉风格 | 分散的卡片 | 统一的极简风格 |
| 信息对比 | 文本描述 | 左右并排对比 |
| 操作便捷性 | 每个标签单独操作 | 所有操作在同一页面 |
| 时间排序 | 每个标签内排序 | 全局统一排序 |

---

## 测试清单

### 前置条件
- [x] 后端服务运行在 `http://localhost:8080`
- [x] 前端服务运行在 `http://localhost:5173`
- [ ] 已创建管理员账户
- [ ] 有测试数据（待审核的歌曲/纠错）

### 测试步骤

#### 1. 访问审核页面
- URL: `http://localhost:5173/admin/review`
- 预期：显示所有待审核内容的统一列表

#### 2. 测试歌曲批次审核
- 找到类型为"歌曲上传"的项目
- 验证显示：专辑名、艺术家、曲目列表、封面
- 点击"试听"按钮 → 音频播放
- 点击"通过" → 确认提示 → 项目消失
- 或点击"驳回" → 确认提示 → 项目消失

#### 3. 测试文本纠错审核
- 找到类型为"歌曲纠错"或"专辑修正"的文本修改
- 验证左右对比显示正确
- 验证字段名称显示为中文
- 验证修改原因显示（如果有）
- 执行批准/驳回操作

#### 4. 测试封面纠错审核
- 找到封面修改项
- 验证左右图片对比显示
- 原封面应为灰度，新封面为彩色
- 执行批准/驳回操作

#### 5. 测试空状态
- 批准/驳回所有项目
- 验证显示"暂无待审核内容"

#### 6. 测试权限控制
- 退出登录
- 尝试访问 `/admin/review`
- 预期：显示"需要管理员权限"提示

---

## 已知限制

1. **无搜索/过滤功能**：如果待审核项目过多（>50），可能需要添加搜索或过滤
2. **无批量操作**：不能一次性批准/驳回多个项目
3. **无撤销功能**：批准/驳回后无法撤销

---

## 未来优化方向

### 短期（如有需求）
- [ ] 添加搜索框（按歌曲名/专辑名/提交者搜索）
- [ ] 添加过滤器（按类型/提交者/日期过滤）
- [ ] 批量操作功能（批量批准/驳回）

### 中期
- [ ] 审核历史记录页面
- [ ] 审核统计仪表板
- [ ] 邮件通知（审核结果通知提交者）

### 长期
- [ ] 审核评论功能（驳回时可留言说明原因）
- [ ] 分配审核任务（多个管理员协作）
- [ ] 审核优先级排序

---

## 文件变更记录

### 修改的文件
- `web/src/views/AdminReviewView.vue` - 完全重写

### 删除的代码
- 4个标签页组件（songs/corrections/album_corrections/albums）
- 分散的状态管理（`activeTab` ref）
- 重复的空状态组件

### 新增的代码
- `ReviewItem` 统一接口
- `fetchAllPendingItems()` 统一数据获取
- `getTypeLabel()` 类型标签文本
- `getFieldLabel()` 字段名称中文化
- 统一的条件渲染逻辑

---

## 总结

新的审核队列界面实现了：
✅ 统一、直观的审核体验  
✅ 清晰的原值/修改后对比  
✅ 简洁的操作流程  
✅ 与上传界面一致的视觉风格  
✅ 无需在多个标签页间切换  

符合用户需求：**不需要分类，只需标注好原信息和修改后的信息，以及是否通过**。
