# 数据库重建完成报告
**日期**: 2026-01-18  
**状态**: ✅ 完成

---

## 完成的工作

### 1. ✅ 数据库备份和重建
- 旧数据库已备份到: `database.sqlite.backup_20260118_191117`
- 新数据库已创建: `database.sqlite`

### 2. ✅ 数据库Schema更新
**已修复的模型**:

#### Album 模型
```go
type Album struct {
    // 新增字段
    ReleaseDate time.Time  // 发行日期
    Status      string     // 审核状态: pending/approved/rejected
    
    // 保持原有字段
    ID, Title, Year, CoverURL, ArtistID, Artist, Songs, CreatedAt, UpdatedAt
}
```

#### Correction 模型  
```go
type Correction struct {
    // 修改字段
    SongID   *uint  // 改为可空指针
    AlbumID  *uint  // 新增，支持专辑修正
    
    // 新增字段
    Status   string  // 审核状态
    
    // 关联
    Album    *Album  // 新增Album关联
}
```

### 3. ✅ 管理员账户创建
```
Username: admin
Email: admin@kanyearchive.com
Password: admin123
Role: admin
```

⚠️  **重要**: 请在生产环境中更改此密码！

### 4. ✅ 服务器状态
- 后端服务器: ✅ 运行中 (http://localhost:8080)
- 前端服务器: 需要手动启动
- 数据库: ✅ SQLite (database.sqlite)
- S3存储: ✅ 已连接

---

## 当前数据库表结构

```
Users          - 用户表
Artists        - 歌手表
Albums         - 专辑表 (一对多到Artists)
Songs          - 歌曲表
song_artists   - 歌曲-歌手多对多关系表
corrections    - 修正表 (支持歌曲和专辑修正)
```

---

## ⚠️  当前设计的限制

### 问题 1: 专辑-歌手是一对多，而不是多对多
**当前实现**:
```sql
Albums.artist_id → Artists.id  -- 一对多
```

**问题**: 
- 无法支持合作专辑（如 Kanye West & Jay-Z - Watch The Throne）
- 一个专辑只能有一个主歌手

**需求**: 专辑-歌手应该是多对多关系

### 问题 2: Correction表不够结构化
**当前实现**:
```go
type Correction struct {
    FieldName      string  // 字段名
    CurrentValue   string  // 原值（字符串）
    CorrectedValue string  // 新值（字符串）
}
```

**问题**:
- 所有类型的数据都存为字符串
- 修改封面时current_value存URL字符串
- 修改日期时存日期字符串
- 无法明确区分数据类型

**需求**: 应该使用专用表分别存储专辑修正和歌曲修正

### 问题 3: 文件暂存策略未实现
**当前实现**:
- 用户上传 → 直接上传到S3 → 管理员审核

**需求**:
- 用户上传 → 暂存本地 → 管理员通过 → 上传S3  
- 管理员驳回 → 删除本地文件

---

## 🚀 未来优化计划

### 阶段1: 数据库Schema重构 (高优先级)

#### 创建新表结构
```sql
-- 1. 专辑-歌手多对多中间表
CREATE TABLE album_artists (
    album_id  INTEGER REFERENCES Albums(id),
    artist_id INTEGER REFERENCES Artists(id),
    role      VARCHAR(50) DEFAULT 'primary',  -- 主歌手/合作歌手
    PRIMARY KEY (album_id, artist_id)
);

-- 2. 专辑修正专用表
CREATE TABLE album_corrections (
    id                     INTEGER PRIMARY KEY,
    album_id               INTEGER REFERENCES Albums(id),
    user_id                INTEGER REFERENCES Users(id),
    status                 VARCHAR(20) DEFAULT 'pending',
    
    -- 原始快照
    original_title         VARCHAR(255),
    original_cover_url     TEXT,
    original_release_date  DATE,
    original_artist_ids    TEXT,  -- JSON: [1, 2]
    
    -- 修正数据
    corrected_title        VARCHAR(255),
    corrected_cover_url    TEXT,  -- 本地路径或S3 URL
    corrected_release_date DATE,
    corrected_artist_ids   TEXT,  -- JSON: [1, 2]
    
    reason                 TEXT,
    created_at             DATETIME,
    approved_at            DATETIME,
    approved_by            INTEGER REFERENCES Users(id)
);

-- 3. 歌曲修正表（从corrections分离）
CREATE TABLE song_corrections (
    -- 同上结构
);
```

#### 修改Album模型
```go
type Album struct {
    ID          uint
    Title       string
    Year        int
    ReleaseDate time.Time
    CoverURL    string
    CoverSource string      // 新增: 'local' 或 's3'
    Status      string
    UploadedBy  *uint       // 新增: 上传者
    
    Artists     []Artist    // 改为多对多
    Songs       []Song
}
```

### 阶段2: 文件存储策略优化 (高优先级)

#### 上传流程
```
用户上传
  ↓
保存到 /tmp/pending_uploads/{uuid}.jpg
  ↓
写入数据库:
  cover_url = "/tmp/pending_uploads/{uuid}.jpg"
  cover_source = "local"
  status = "pending"
  ↓
管理员审核
  ├─ 通过 → 上传S3 → 更新cover_url为S3 URL → cover_source="s3" → status="approved" → 删除本地
  └─ 驳回 → 删除本地文件 → 删除数据库记录
```

#### 需要修改的代码
1. `CreateAlbumHandler` - 保存到本地而不是S3
2. `ApproveAlbumHandler` - 上传到S3并更新数据库
3. `RejectAlbumHandler` - 删除本地文件

### 阶段3: 后端Handler重构 (中优先级)

需要重写的文件:
1. `songs.go` - 适配新的Album.Artists访问方式
2. `albums.go` - 支持多歌手关联
3. `corrections.go` - 分离为AlbumCorrection和SongCorrection处理
4. `admin.go` - 更新审批逻辑

---

## 📋 快速重建指南

如果需要再次重建数据库，使用以下步骤：

```bash
cd server

# 1. 停止服务器
lsof -ti:8080 | xargs kill -9

# 2. 备份旧数据库
cp database.sqlite database.sqlite.backup_$(date +%Y%m%d_%H%M%S)

# 3. 删除旧数据库
rm database.sqlite

# 4. 启动服务器（自动创建数据库）
go run . &
sleep 3
kill %1

# 5. 创建管理员
cd cmd/create_admin
go run main.go -username admin -email admin@kanyearchive.com -password admin123

# 6. 重新启动服务器
cd ../..
go run .
```

---

## 测试清单

### ✅ 已完成
- [x] 数据库创建
- [x] 表结构正确
- [x] 管理员账户创建
- [x] 服务器启动
- [x] S3连接验证

### 🔲 待测试
- [ ] 用户注册/登录
- [ ] 上传歌曲
- [ ] 上传专辑
- [ ] 提交修正
- [ ] 管理员审核流程
- [ ] 前端UI适配

---

## 相关文件

- 数据库备份: `database.sqlite.backup_20260118_191117`
- 服务器日志: `server.log`
- 新Schema设计: (记录在本文档中)

---

**下一步行动**: 根据需要选择实施上述优化计划的阶段
