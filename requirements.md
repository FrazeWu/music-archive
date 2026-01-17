# 需求清单 (Requirements List) for All Kanye Song Collection Website

## 概述
这是一个收集 Kanye West 全部歌曲的 Vue 3 网站，包含用户认证、上传、勘误功能，使用极简黑白档案馆风格 UI。

## UI/UX 设计规范

### 设计风格
**极简主义 + 档案馆美学 (Minimalist Archive Aesthetic)**

#### 色彩系统
- 主色调：纯黑 (#000000) + 纯白 (#FFFFFF)
- 辅助色：灰度系统 (gray-100, gray-400, gray-500, gray-600, gray-700)
- 强调色：绿色 (text-green-600) 用于状态显示，红色 (text-red-600) 用于管理入口

#### 排版系统
- 主标题：`text-5xl ~ text-6xl`, `font-black`, `tracking-tighter` (极紧字距)
- 次级标题：`text-2xl ~ text-4xl`, `font-black`, `tracking-tight`
- 正文：`text-sm ~ text-xl`, `font-medium` / `font-bold`
- 标签/标识：`text-xs`, `font-black`, `uppercase`, `tracking-widest` (极宽字距)

#### 视觉语言
- **边框美学**：所有卡片、按钮、输入框使用 `border-2 border-black` 或 `border-4 border-black`
- **投影效果**：使用硬投影 `shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]` 替代柔和阴影
- **灰度图像**：所有封面图片使用 `grayscale` 滤镜，统一视觉调性
- **过渡动效**：使用 `transition-all` + `duration-300 / duration-500`，强调状态变化
- **悬停反转**：按钮悬停时执行黑白反转 `hover:bg-white hover:text-black`

#### 组件样式规范

**按钮 (Buttons)**
```css
/* 主要按钮 */
bg-black text-white border-2 border-black px-8 py-4 font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all

/* 次要按钮 */
border-2 border-black px-4 py-2 font-black text-xs uppercase tracking-widest hover:bg-black hover:text-white transition-colors
```

**输入框 (Inputs)**
```css
w-full bg-white border-2 border-black p-4 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all
```

**卡片 (Cards)**
```css
bg-white border-2 border-black p-6 hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all duration-300
```

### 布局结构

#### 1. Topbar (固定顶栏)
- **位置**：`sticky top-0 z-50`
- **高度**：`h-16` (64px)
- **背景**：白色背景 + 黑色底边 `border-b-2 border-black`
- **内容布局**：
  - 左侧：`KANYE ARCHIVE` 品牌标识 (`text-2xl font-black tracking-tighter`)
  - 右侧：导航链接 + 用户菜单
    - 时间线 | 上传歌曲 | 审核队列 (admin) | 用户名下拉菜单 / 登录

#### 2. 主内容区域

**首页时间线 (HomeView)**
- **布局**：垂直时间轴 + 卡片展示
- **时间轴线**：`absolute left-1/3` 位置，`w-1 bg-black` 贯穿全页
- **年份标签**：悬浮于时间轴上方 `bg-black text-white px-4 py-1 font-bold tracking-widest`
- **时间节点**：圆形节点 `w-6 h-6 rounded-full border-4 border-white bg-black`
  - 当前播放歌曲节点放大 `scale-125`
- **歌曲卡片**：
  - 位置：`ml-[calc(33.333%+2rem)] w-[calc(66.666%-2rem)]`
  - 封面：`w-32 h-32 border-2 border-black grayscale`
  - 悬停效果：`hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]`

**歌曲详情页 (SongDetailView)**
- **布局**：双栏网格 `grid grid-cols-1 md:grid-cols-2 gap-16`
- **左栏**：大封面图 `border-4 border-black grayscale shadow-2xl`
- **右栏**：
  - 档案编号标签：`YE-0001` 格式
  - 标题 + 艺术家信息
  - 操作按钮：播放 + 修正数据
  - 歌词预览区：`italic text-gray-700`
  - 技术规格表格：发行年份、音频格式、档案编号、状态

**上传页面 (UploadView)**
- **布局**：居中单栏表单 `max-w-2xl mx-auto`
- **标题**：`text-4xl font-black tracking-tighter`
- **文件上传区**：虚线边框 `border-2 border-dashed border-black p-12`
  - 悬停时背景变化 `hover:bg-gray-100`
  - 文件选中后显示文件名

**登录/注册页面 (LoginView)**
- **布局**：垂直居中 `min-h-[calc(100vh-64px)] flex items-center justify-center`
- **表单容器**：`max-w-md border-2 border-black p-12 shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]`
- **输入框**：`focus:bg-gray-50` 聚焦时轻微背景变化

#### 3. 底部播放器 (AudioPlayer)
- **位置**：`fixed bottom-0 w-full z-50`
- **背景**：白色背景 + 顶部黑色边框 `border-t-2 border-black`
- **布局**：三列布局 (歌曲信息 | 播放控制 | 音量控制)
- **进度条**：
  - 容器：`h-1 bg-gray-200`
  - 进度：`bg-black` 动态宽度
  - 交互：点击跳转到指定位置
- **控制按钮**：
  - 播放/暂停：主按钮样式 `px-6 py-2 border-2 border-black bg-black text-white`
  - 辅助按钮：次要按钮样式 `px-3 py-1 border border-black`
  - 激活状态：背景反转 `bg-black text-white`

### 动画系统
- **淡入动画**：`animate-in fade-in slide-in-from-bottom-4 duration-500`
- **加载状态**：`animate-pulse`
- **下拉菜单**：`animate-in fade-in slide-in-from-top-1 duration-200`
- **旋转图标**：`transition-transform duration-200 rotate-180`

### 字体权重使用规范
- `font-black` (900)：标题、标签、按钮、品牌名
- `font-bold` (700)：次要标题、强调文本
- `font-medium` (500)：正文、描述性文字

### 间距系统
- 页面边距：`px-8` (水平) + `py-20` (垂直)
- 卡片内边距：`p-6` (内容卡片) / `p-12` (表单容器)
- 元素间距：`gap-4`, `gap-8`, `gap-16` (根据层级递增)

---

## 功能需求

### 1. 用户认证 (User Authentication)
- 实现登录/注册系统（邮箱 + 密码）
- 未登录用户可以查看和播放歌曲
- 登录用户可以上传、编辑和勘误
- 管理员角色可以访问审核队列

### 2. 数据库架构 (Database Architecture)
- **Artists 表**：歌手信息 (id, name, bio, image_url)
- **Albums 表**：专辑信息 (id, title, year, cover_url, artist_id)
- **Songs 表**：歌曲信息 (id, title, year, lyrics, audio_url, cover_url, album_id, uploaded_by)
- **SongArtists 中间表**：歌曲-歌手多对多关系 (song_id, artist_id, role)
- **Users 表**：用户账户 (id, username, email, password_hash, role)
- **Corrections 表**：勘误记录 (id, song_id, field_name, current_value, corrected_value, reason, user_id, status)

**关系定义**：
- Artist (1) ↔ Album (N)：一个歌手拥有多张专辑
- Album (1) ↔ Song (N)：一张专辑包含多首歌曲
- Artist (N) ↔ Song (M)：歌手和歌曲多对多 (支持 feat、合作等场景)
- User (1) → Song (N)：用户上传的歌曲

### 3. 歌曲/专辑管理 (Song/Album Management)
- 上传歌曲时自动创建或关联歌手、专辑
- 显示时间线视图，按年份时间顺序排列
- 详情页面：显示歌曲信息、播放选项、技术规格
- 档案编号系统：`YE-{id.padStart(4, '0')}` 格式

### 4. 勘误功能 (Correction Feature)
- 登录用户可以提交歌曲信息修正
- 如果有错误，可以重新上传媒体文件
- 提交勘误后，管理员审核队列处理
- 勘误状态：pending / approved / rejected

### 5. 媒体存储 (Media Storage)
- 使用 AWS S3 存储音频文件和封面图片
- **S3 路径格式**：`music/{ArtistName}/{AlbumName}/{FileName}`
- 支持流媒体播放，确保安全访问
- 文件上传时对路径进行安全处理（斜杠替换）

### 6. 后端 API (Backend API)
- **技术栈**：Go (Gin framework) + GORM
- **数据库**：MySQL / PostgreSQL
- **认证**：JWT Token
- **端点**：
  - `/api/auth/register` - 用户注册
  - `/api/auth/login` - 用户登录
  - `/api/songs` - 获取所有歌曲 (带关联数据 Preload)
  - `/api/songs/:id` - 获取单个歌曲详情
  - `/api/songs` POST - 上传新歌曲 (需认证)
  - `/api/corrections` - 勘误管理

### 7. 播放器 (Player)
- 使用 HTML5 Audio API
- 功能支持：
  - 播放 / 暂停
  - 上一首 / 下一首
  - 进度条拖拽
  - 音量控制
  - 随机播放
  - 循环模式：不循环 / 列表循环 / 单曲循环
- 播放器状态管理使用 Pinia Store
- 固定在页面底部，不影响内容滚动

### 8. 审核系统 (Admin Review)
- 管理员查看待审核的上传内容和勘误
- 批准/拒绝操作
- 状态追踪和通知

---

## 技术栈

### 前端
- **框架**：Vue 3 (Composition API)
- **语言**：TypeScript
- **构建工具**：Vite
- **状态管理**：Pinia
- **路由**：Vue Router
- **样式**：Tailwind CSS
- **动画**：Tailwind 动画工具类

### 后端
- **语言**：Go
- **框架**：Gin
- **ORM**：GORM
- **认证**：JWT (jsonwebtoken)
- **文件上传**：AWS SDK for Go (S3)
- **数据库**：MySQL / PostgreSQL

### 基础设施
- **云存储**：AWS S3
- **部署**：Docker (可选)
- **测试**：Vitest (前端), Go testing (后端)

---

## 优先级
- **高**：用户认证、数据库架构、歌曲管理、后端 API、S3 上传
- **中**：UI 细节优化、勘误功能、审核系统
- **低**：性能优化、SEO、国际化