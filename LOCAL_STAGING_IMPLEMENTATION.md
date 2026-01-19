# 本地文件暂存实施完成报告

## 概述

成功实现了本地文件暂存系统，普通用户上传的文件现在会先保存到本地，等待管理员审批后再上传到 S3。

---

## 实施内容

### ✅ 1. 后端文件存储系统

**新增工具函数** (`server/storage.go`):
- `SaveFileLocally()` - 保存文件到 `uploads/music/艺术家/专辑/`
- `UploadLocalFileToS3()` - 上传本地文件到 S3
- `DeleteLocalFile()` - 删除本地文件
- `GetLocalPathFromURL()` - URL 转本地路径

**文件组织结构**:
```
server/uploads/music/
  └── Kanye West/
      └── 808s & Heartbreak/
          ├── 01 Say You Will.mp3
          ├── 02 Welcome To Heartbreak.mp3
          └── cover_808s.jpg
```

---

### ✅ 2. 上传逻辑修改

**songs.go - CreateSongHandler**:
- 判断用户角色
- **管理员**: 文件直接上传到 S3，`audio_source = 's3'`, `status = 'approved'`
- **普通用户**: 文件保存到本地，`audio_source = 'local'`, `status = 'pending'`
- 音频和封面都支持本地暂存

**albums.go - CreateAlbumHandler**:
- 同样的双路径逻辑
- 专辑封面支持本地暂存

**文件来源追踪**:
- `audio_source`: `'local'` 或 `'s3'`
- `cover_source`: `'local'` 或 `'s3'`

---

### ✅ 3. 管理员审批流程

**admin.go - ApproveSongHandler**:
```
1. 检查 audio_source 是否为 'local'
2. 如果是，上传本地文件到 S3
3. 更新 audio_url 为 S3 URL
4. 更新 audio_source 为 's3'
5. 删除本地文件
6. 更新 status 为 'approved'
```

**admin.go - RejectSongHandler**:
```
1. 检查 audio_source 是否为 'local'
2. 如果是，删除本地文件
3. 从数据库删除记录
```

同样的逻辑应用于封面文件和专辑审批。

---

### ✅ 4. 静态文件服务

**main.go**:
```go
router.Static("/uploads", "./uploads")
```

本地文件可通过 HTTP 访问：
```
http://localhost:8080/uploads/music/Kanye West/808s & Heartbreak/01 Say You Will.mp3
```

---

### ✅ 5. 前端审核队列优化

**AdminReviewView.vue** - 批次合并显示:

**之前**:
```
[ Song 1 ] [ Approve ] [ Reject ]
[ Song 2 ] [ Approve ] [ Reject ]
[ Song 3 ] [ Approve ] [ Reject ]
```

**现在**:
```
┌─────────────────────────────────────────┐
│ [封面] 808s & Heartbreak               │
│        Kanye West                       │
│        3 songs                          │
│                                         │
│        1. Say You Will                  │
│        2. Welcome To Heartbreak         │
│        3. Heartless                     │
│                                         │
│        [ Approve All ] [ Reject All ]   │
└─────────────────────────────────────────┘
```

**实现细节**:
- 按 `batch_id` 分组歌曲
- 单次操作审批/拒绝整个专辑
- 显示歌曲数量和列表
- 添加加载状态防止重复提交
- 错误处理显示部分失败数量

---

## 工作流程

### 普通用户上传流程

```
1. 用户选择文件并上传
   ↓
2. 后端接收文件
   ↓
3. 检测用户角色 = 'user'
   ↓
4. 保存文件到 server/uploads/music/艺术家/专辑/
   ↓
5. 数据库创建记录:
   - audio_url = "/uploads/music/..."
   - audio_source = "local"
   - status = "pending"
   ↓
6. 文件通过 /uploads/ 端点可访问（用于预览）
   ↓
7. 等待管理员审批
```

### 管理员审批流程

```
1. 管理员在审核页面看到待审批内容
   ↓
2. 点击 "Approve All"
   ↓
3. 后端处理每首歌:
   a. 打开本地文件
   b. 上传到 S3
   c. 更新 audio_url 为 S3 URL
   d. 更新 audio_source 为 "s3"
   e. 删除本地文件
   f. 更新 status 为 "approved"
   ↓
4. 歌曲现在对所有用户可见
```

### 管理员拒绝流程

```
1. 管理员点击 "Reject All"
   ↓
2. 后端处理每首歌:
   a. 删除本地音频文件
   b. 删除本地封面文件
   c. 从数据库删除记录
   ↓
3. 文件和数据完全清除
```

### 管理员上传流程

```
1. 管理员上传文件
   ↓
2. 后端检测用户角色 = 'admin'
   ↓
3. 文件直接上传到 S3
   ↓
4. 数据库创建记录:
   - audio_url = "https://s3.../music/..."
   - audio_source = "s3"
   - status = "approved"
   ↓
5. 立即对所有用户可见
```

---

## 优势

### 1. 节省云存储成本
- 只有审批通过的内容才上传到 S3
- 拒绝的内容只占用临时本地空间

### 2. 更快的拒绝处理
- 删除本地文件比 S3 删除更快
- 无需 S3 API 调用

### 3. 管理员优先级
- 管理员上传立即生效
- 无需等待审批流程

### 4. 内容预览
- 本地文件通过 HTTP 可访问
- 可在审批前预览/播放

### 5. 更好的用户体验
- 批次上传显示为单个项目
- 一键审批/拒绝整张专辑
- 清晰的状态反馈

---

## 技术细节

### 文件路径映射

| 存储类型 | 数据库 URL | 实际路径 |
|---------|-----------|---------|
| 本地 | `/uploads/music/Kanye West/808s/song.mp3` | `server/uploads/music/Kanye West/808s/song.mp3` |
| S3 | `https://objectstorage.../music/Kanye West/808s/song.mp3` | S3: `music/Kanye West/808s/song.mp3` |

### API 端点

| 端点 | 方法 | 功能 |
|-----|------|------|
| `/api/songs` | POST | 创建歌曲（根据角色决定存储位置）|
| `/api/admin/approve/:id` | POST | 审批歌曲（上传到 S3）|
| `/api/admin/reject/:id` | POST | 拒绝歌曲（删除本地文件）|
| `/uploads/music/...` | GET | 访问本地暂存文件 |

### 数据库字段

| 字段 | 类型 | 可能值 |
|-----|------|--------|
| `audio_url` | string | `/uploads/...` 或 `https://s3.../...` |
| `audio_source` | string | `'local'` 或 `'s3'` |
| `cover_url` | string | `/uploads/...` 或 `https://s3.../...` |
| `cover_source` | string | `'local'` 或 `'s3'` |
| `status` | string | `'pending'` 或 `'approved'` 或 `'rejected'` |

---

## 测试建议

### 测试场景 1: 普通用户上传
1. 以普通用户身份登录
2. 上传一张专辑（3首歌）
3. 检查 `server/uploads/music/` 目录是否有文件
4. 检查数据库 `audio_source = 'local'`, `status = 'pending'`
5. 尝试访问 `http://localhost:8080/uploads/music/...` 确认可播放

### 测试场景 2: 管理员审批
1. 以管理员身份登录审核页面
2. 看到批次合并显示（一个专辑显示为一个项目）
3. 点击 "Approve All"
4. 检查本地文件被删除
5. 检查 S3 有新文件
6. 检查数据库 `audio_source = 's3'`, `status = 'approved'`
7. 前端主页应显示新专辑

### 测试场景 3: 管理员拒绝
1. 上传新专辑
2. 管理员点击 "Reject All"
3. 检查本地文件被删除
4. 检查数据库记录被删除
5. 检查 S3 没有新文件

### 测试场景 4: 管理员直传
1. 以管理员身份上传专辑
2. 检查文件直接到 S3（本地无文件）
3. 检查 `audio_source = 's3'`, `status = 'approved'`
4. 立即在主页可见

---

## 注意事项

### 1. 磁盘空间管理
- 定期清理长时间 pending 的文件
- 考虑添加自动清理脚本（例如 7 天未审批自动删除）

### 2. 文件访问权限
- `server/uploads/` 目录需要读写权限
- Gin router 配置了静态文件服务

### 3. 并发审批
- 当前实现支持单个管理员审批
- 多管理员同时审批可能需要加锁机制

### 4. 错误恢复
- 如果 S3 上传失败，本地文件仍保留
- 可重试审批操作

---

## 文件清单

### 后端修改
- ✅ `server/storage.go` - 添加本地文件工具函数
- ✅ `server/songs.go` - 修改上传逻辑支持本地暂存
- ✅ `server/albums.go` - 修改专辑上传逻辑
- ✅ `server/admin.go` - 审批/拒绝时处理本地文件
- ✅ `server/main.go` - 添加静态文件服务
- ✅ `server/uploads/` - 创建本地暂存目录

### 前端修改
- ✅ `web/src/views/AdminReviewView.vue` - 批次合并显示

### 构建验证
- ✅ 后端编译成功（`go build`）
- ✅ 前端构建成功（`npm run build`）
- ✅ LSP 诊断无错误

---

## 下一步

1. **启动服务器测试**:
   ```bash
   # 后端
   cd server
   ./all_kanye_server
   
   # 前端
   cd web
   npm run dev
   ```

2. **手动测试上传流程**

3. **可选优化**:
   - 添加上传进度显示
   - 添加批量拒绝确认对话框
   - 添加审批历史记录
   - 实现自动清理脚本

---

## 状态

**实施状态**: ✅ 完成  
**测试状态**: ⏳ 待手动测试  
**部署状态**: 🔄 准备就绪

---

**实施日期**: 2026-01-18  
**实施人**: Sisyphus (AI Agent)
