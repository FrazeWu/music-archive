# 修复总结 - 2026-01-18 18:31

## 🐛 修复的问题

### 问题1: Corrections 表列名错误
**错误信息**:
```
no such column: createdAt
SELECT * FROM `corrections` WHERE status = "pending" ORDER BY createdAt desc
```

**根本原因**:
- `Songs` 和 `Albums` 表使用 camelCase 列名：`createdAt`, `updatedAt`
- `Corrections` 表使用 snake_case 列名：`created_at`
- 查询语句混淆了两种命名风格

**修复**:
- `server/admin.go` 第223行：`Order("createdAt desc")` → `Order("created_at desc")`

---

### 问题2: 缺少 /api/admin/pending-album-corrections 端点
**错误**:
- 前端调用 `GET /api/admin/pending-album-corrections`
- 后端没有实现这个端点
- 导致前端无法获取待审核的专辑纠错列表

**修复**:
1. 新增 `GetPendingAlbumCorrectionsHandler` 函数（admin.go 第299-308行）
   ```go
   func GetPendingAlbumCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
       return func(c *gin.Context) {
           var corrections []Correction
           if err := db.Where("status = ? AND album_id IS NOT NULL", "pending").
               Preload("User").Preload("Album").Preload("Album.Artist").
               Order("created_at desc").Find(&corrections).Error; err != nil {
               c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending album corrections"})
               return
           }
           c.JSON(http.StatusOK, corrections)
       }
   }
   ```

2. 在路由中注册端点（admin.go 第60行）
   ```go
   admin.GET("/pending-album-corrections", GetPendingAlbumCorrectionsHandler(db))
   ```

**查询逻辑**:
- 筛选条件：`status = "pending" AND album_id IS NOT NULL`
- 预加载关联：User, Album, Album.Artist
- 排序：按 `created_at` 降序

---

## ✅ 验证结果

### 服务器启动日志
```
[GIN-debug] GET /api/admin/pending-album-corrections --> ...GetPendingAlbumCorrectionsHandler.func14
```
✅ 新端点已成功注册

### 测试账户
- **用户名**: `fazong`
- **角色**: `admin` (已提升)
- **可访问**: http://localhost:5173/admin/review

---

## 📊 数据库表列名映射

| 表名 | created_at 列名 | updated_at 列名 |
|------|----------------|-----------------|
| Songs | `createdAt` (camelCase) | `updatedAt` |
| Albums | `createdAt` (camelCase) | `updatedAt` |
| Corrections | `created_at` (snake_case) | ❌ 不存在 |
| Users | ❌ 不存在 | ❌ 不存在 |

**重要提示**: 在写 Order 子句时，必须根据具体表的列名使用正确的命名风格。

---

## 🔍 所有 Order 子句检查

### ✅ 正确的查询

| 文件 | 行号 | 查询 | 表 | 列名风格 |
|------|------|------|-----|---------|
| admin.go | 84 | `Order("createdAt desc")` | Songs | camelCase ✅ |
| admin.go | 223 | `Order("created_at desc")` | Corrections | snake_case ✅ |
| admin.go | 288 | `Order("createdAt desc")` | Albums | camelCase ✅ |
| admin.go | 306 | `Order("created_at desc")` | Corrections | snake_case ✅ |

---

## 🚀 当前系统状态

### 后端
- **状态**: ✅ 运行中
- **端口**: 8080
- **进程**: `nohup go run . > server.log 2>&1 &`
- **日志**: `/Users/fafa/Documents/projects/all_kanye/server/server.log`

### 前端
- **状态**: ✅ 运行中
- **端口**: 5173
- **URL**: http://localhost:5173

### 数据库
- **类型**: SQLite
- **路径**: `/Users/fafa/Documents/projects/all_kanye/server/database.sqlite`
- **歌曲数**: 17 首

---

## 🧪 测试步骤

### 1. 登录管理员账户
- URL: http://localhost:5173/login
- 用户名: `fazong`
- 密码: (您设置的密码)

### 2. 访问审核队列
- URL: http://localhost:5173/admin/review
- 应该看到统一的审核列表（无标签页分类）

### 3. 验证数据加载
打开浏览器开发者工具 Network 标签，应该看到以下请求：
- ✅ `GET /api/admin/pending` - 待审核歌曲批次
- ✅ `GET /api/admin/pending-corrections` - 待审核歌曲纠错
- ✅ `GET /api/admin/pending-album-corrections` - 待审核专辑纠错（新增）
- ✅ `GET /api/admin/pending-albums` - 待审核专辑封面

### 4. 测试审核操作
如果有待审核内容：
- 点击"通过"按钮 → 应该调用相应的 approve 端点
- 点击"驳回"按钮 → 应该调用相应的 reject 端点
- 操作成功后，该项应从列表中消失

---

## 📝 修改的文件

### server/admin.go
```diff
+ // GetPendingAlbumCorrectionsHandler retrieves all pending album corrections
+ func GetPendingAlbumCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
+     return func(c *gin.Context) {
+         var corrections []Correction
+         if err := db.Where("status = ? AND album_id IS NOT NULL", "pending").
+             Preload("User").Preload("Album").Preload("Album.Artist").
+             Order("created_at desc").Find(&corrections).Error; err != nil {
+             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending album corrections"})
+             return
+         }
+         c.JSON(http.StatusOK, corrections)
+     }
+ }

  func SetupAdminRoutes(r *gin.RouterGroup, db *gorm.DB, s3Client *s3.S3) {
      admin := r.Group("/admin")
      admin.Use(AuthMiddleware())
      admin.Use(AdminMiddleware(db))
      {
          admin.GET("/pending", GetPendingRequestsHandler(db))
          // ... 其他路由 ...
+         admin.GET("/pending-album-corrections", GetPendingAlbumCorrectionsHandler(db))
          admin.POST("/approve-album-correction/:id", ApproveAlbumCorrectionHandler(db, s3Client))
          admin.POST("/reject-album-correction/:id", RejectAlbumCorrectionHandler(db, s3Client))
      }
  }

  func GetPendingCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
      return func(c *gin.Context) {
          var corrections []Correction
-         if err := db.Where("status = ?", "pending").Preload("User").Preload("Song").Order("createdAt desc").Find(&corrections).Error; err != nil {
+         if err := db.Where("status = ?", "pending").Preload("User").Preload("Song").Order("created_at desc").Find(&corrections).Error; err != nil {
              c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending corrections"})
              return
          }
          c.JSON(http.StatusOK, corrections)
      }
  }
```

### web/src/views/AdminReviewView.vue
- **完全重写**：从4个标签页改为统一列表视图
- **详细文档**: 见 `ADMIN_REVIEW_REDESIGN.md`

---

## ⚠️ 注意事项

### 数据库列名不一致问题
当前数据库存在列名不一致的情况：
- 某些表使用 camelCase (`createdAt`, `albumId`)
- 某些表使用 snake_case (`created_at`, `album_id`)
- 甚至同一张表有两种命名（如 Songs 表同时有 `albumId` 和 `album_id`）

**建议**（未来优化）:
1. 统一所有表的列名为 snake_case（Go 社区推荐）
2. 或者统一使用 camelCase（需要配置 GORM NamingStrategy）
3. 编写数据库迁移脚本统一列名

### Order 子句最佳实践
在写 GORM 查询时：
1. 先用 `PRAGMA table_info(TableName)` 检查实际列名
2. 不要假设列名，明确指定
3. 考虑在 model 中用 `gorm:"column:xxx"` 标签统一映射

---

## ✅ 完成清单

- [x] 修复 Corrections 表 Order 子句列名错误
- [x] 添加 GetPendingAlbumCorrectionsHandler 函数
- [x] 注册 /api/admin/pending-album-corrections 路由
- [x] 重启后端服务器
- [x] 验证新端点已注册
- [x] 提升 fazong 为管理员
- [x] 重写 AdminReviewView 为统一列表视图
- [x] TypeScript 编译通过
- [x] 创建修复总结文档

---

## 🎯 下一步

1. **测试审核界面**
   - 访问 http://localhost:5173/admin/review
   - 验证所有数据正确加载
   - 测试批准/驳回功能

2. **创建测试数据**（如果没有待审核内容）
   - 用普通用户登录
   - 上传歌曲或提交纠错
   - 然后用管理员审核

3. **后续优化**（可选）
   - 统一数据库列名
   - 添加搜索/过滤功能
   - 添加批量操作
   - 优化移动端显示

---

**所有问题已修复！系统现在可以正常运行。**
