# 审核功能测试指南

## 前置条件

### 1. 确认服务运行状态

**后端**:
```bash
lsof -i:8080  # 应该显示Go进程
```

**前端**:
```bash
lsof -i:5173  # 应该显示Node/Vite进程
```

### 2. 确认管理员账户

- **用户名**: `fazong`
- **邮箱**: `fz.wu@icloud.com`
- **角色**: `admin`
- **密码**: 您需要知道密码才能登录

如果不知道密码，重置方法：
```bash
cd /Users/fafa/Documents/projects/all_kanye/server/cmd/create_admin
go run main.go -promote fazong
# 或者创建新的管理员
go run main.go -username admin -email admin@test.com -password admin123
```

### 3. 确认有待审核内容

当前数据库中有一个待审核的专辑封面纠错：
```bash
cd /Users/fafa/Documents/projects/all_kanye/server
sqlite3 database.sqlite "SELECT id, album_id, field_name, status FROM Corrections WHERE status='pending';"
# 输出: 2|1|cover_url|pending
```

---

## 测试步骤

### 步骤1: 登录管理员账户

1. 打开浏览器访问: http://localhost:5173/login
2. 输入：
   - 邮箱: `fz.wu@icloud.com`
   - 密码: （您的密码）
3. 点击登录

**预期结果**: 
- 登录成功，跳转到首页
- 浏览器localStorage中保存了JWT token

**调试方法**（如果登录失败）:
- 打开浏览器开发者工具 → Network标签
- 查看 POST `/api/auth/login` 请求的响应
- 如果返回401，说明密码错误
- 如果返回其他错误，查看响应详情

---

### 步骤2: 访问审核队列页面

1. 访问: http://localhost:5173/admin/review
2. 或从顶部导航栏点击"审核队列"（如果有）

**预期结果**:
- 页面标题显示"审核队列"
- 显示待审核内容数量
- 看到一条待审核的专辑修正（封面更新）

**如果看不到内容**:
- 打开浏览器开发者工具 → Network标签
- 查看以下请求是否成功（状态码200）:
  - `GET /api/admin/pending`
  - `GET /api/admin/pending-corrections`
  - `GET /api/admin/pending-album-corrections`  ← 应该返回一条记录
  - `GET /api/admin/pending-albums`
- 如果某个请求返回401，说明token失效或角色不是admin
- 如果返回404，说明路由配置有问题

---

### 步骤3: 测试批准操作

**当前有的待审核项**:
- 类型: 专辑修正 (album_correction)
- 字段: cover_url（专辑封面）
- 专辑: 2049

**操作**:
1. 找到这条专辑修正记录
2. 查看左右对比的封面图片
3. 点击"通过"按钮

**预期结果**:
- 该项从列表中消失（无弹窗）
- 不会有"已批准"的alert（已删除）
- 后端日志中应该出现请求记录

**调试方法**:
- 打开浏览器开发者工具 → Network标签
- 点击"通过"后，查看 `POST /api/admin/approve-album-correction/2` 请求
- 检查请求状态码:
  - 200: 成功 ✅
  - 401: 未授权（token问题）
  - 404: 纠错记录不存在
  - 500: 服务器错误（查看后端日志）

**验证结果**:
```bash
cd /Users/fafa/Documents/projects/all_kanye/server
sqlite3 database.sqlite "SELECT id, status FROM Corrections WHERE id=2;"
# 应该显示: 2|approved
```

---

### 步骤4: 测试驳回操作

由于当前只有一条待审核记录，测试驳回需要先创建新的待审核内容。

**创建测试数据**:
1. 用普通用户登录（非admin）
2. 访问任意专辑详情页
3. 点击"编辑专辑"
4. 上传一张新封面（不填修改原因）
5. 提交

然后重新用管理员登录，应该能看到新的待审核项。

---

## 常见问题排查

### 问题1: 点击"通过"没有反应

**可能原因**:
1. **前端JS错误**: 
   - 打开浏览器Console标签，查看是否有红色错误
   - 常见错误: `Cannot read property 'filter' of undefined`

2. **网络请求失败**:
   - 打开Network标签，查看approve请求是否发送
   - 如果没有发送，说明前端代码有问题
   - 如果发送但失败，查看响应内容

3. **Token过期或无效**:
   - 检查localStorage中的token是否存在
   - 尝试重新登录

### 问题2: 提示"操作失败"

说明后端返回了非200状态码。

**检查后端日志**:
```bash
tail -50 /Users/fafa/Documents/projects/all_kanye/server/server.log
```

**常见错误**:
- `Correction not found`: 纠错ID不存在或已被处理
- `Failed to update album`: 数据库更新失败
- `Failed to delete old cover from S3`: S3删除失败（不影响主流程）

### 问题3: 页面显示"暂无待审核内容"但数据库有数据

**可能原因**:
1. **前端没有正确获取数据**:
   ```javascript
   // 检查AdminReviewView.vue的fetchAllPendingItems函数
   // 确认所有API调用都成功了
   ```

2. **数据格式问题**:
   - 后端返回的数据结构与前端期望的不一致
   - 打开Network标签，查看API响应的JSON结构

3. **状态过滤问题**:
   - 确认数据库中状态确实是`pending`
   ```bash
   sqlite3 database.sqlite "SELECT status FROM Corrections WHERE id=2;"
   ```

---

## 后端API测试（命令行）

如果前端有问题，可以直接用curl测试后端：

### 1. 登录获取Token
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fz.wu@icloud.com","password":"您的密码"}' \
  | jq -r '.token')

echo "Token: ${TOKEN:0:50}..."
```

### 2. 获取待审核列表
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/pending-album-corrections \
  | jq '.'
```

### 3. 批准纠错
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/approve-album-correction/2 \
  | jq '.'
```

**预期响应**:
```json
{
  "message": "Correction approved and applied"
}
```

### 4. 验证结果
```bash
# 检查纠错状态
curl -s http://localhost:8080/api/albums/1 | jq '.cover_url'
# 应该显示新的封面URL
```

---

## 前端代码检查点

如果问题仍然存在，检查以下代码：

### AdminReviewView.vue

**1. API URL构造**:
```typescript
// 行180-182附近
const endpoint = item.type === 'song_correction' 
  ? `${api.url}/admin/approve-correction/${item.correction_id}`
  : `${api.url}/admin/approve-album-correction/${item.correction_id}`;
```

**2. Token传递**:
```typescript
// 行185-187附近
headers: { 'Authorization': `Bearer ${authStore.token}` }
```

**3. 成功后的列表更新**:
```typescript
// 行190-191附近
if (response.ok) {
  reviewItems.value = reviewItems.value.filter(i => i.id !== item.id);
}
```

### auth.ts (Store)

**检查token存储**:
```typescript
// 确认login成功后保存了token
localStorage.setItem('token', response.data.token);
```

---

## 预期的完整流程日志

**后端日志** (`tail -f server.log`):
```
[GIN] 2026/01/18 - 18:45:00 | 200 |  1.234ms | ::1 | POST /api/auth/login
[GIN] 2026/01/18 - 18:45:05 | 200 |  5.678ms | ::1 | GET /api/admin/pending-album-corrections
[GIN] 2026/01/18 - 18:45:10 | 200 | 12.345ms | ::1 | POST /api/admin/approve-album-correction/2
```

**浏览器Console**:
```
(无错误信息)
```

**浏览器Network**:
```
POST /api/admin/approve-album-correction/2  Status: 200 OK
Response: {"message":"Correction approved and applied"}
```

---

## 如果所有测试都失败

1. **重启所有服务**:
   ```bash
   # 停止后端
   lsof -ti:8080 | xargs kill -9
   
   # 重启后端
   cd /Users/fafa/Documents/projects/all_kanye/server
   nohup go run . > server.log 2>&1 &
   
   # 前端通常不需要重启,但如果需要:
   # Ctrl+C停止当前的npm run dev
   # 然后重新运行: npm run dev
   ```

2. **清除浏览器缓存和LocalStorage**:
   - 打开开发者工具 → Application → Local Storage
   - 删除所有存储的数据
   - 刷新页面并重新登录

3. **检查TypeScript编译**:
   ```bash
   cd /Users/fafa/Documents/projects/all_kanye/web
   npx vue-tsc --noEmit
   ```

4. **查看完整错误栈**:
   - 浏览器Console中的完整错误信息
   - 后端server.log的完整错误栈
   - 提供这些信息以便进一步诊断

---

## 成功标志

✅ **批准操作成功**的标志:
1. 该项从审核列表中消失
2. 没有alert弹窗（成功和确认都已删除）
3. 数据库中Correction状态变为`approved`
4. Album的cover_url更新为新的URL
5. 后端日志显示200响应

✅ **驳回操作成功**的标志:
1. 该项从审核列表中消失
2. 没有alert弹窗
3. 数据库中Correction状态变为`rejected`
4. 后端日志显示200响应

---

**测试结果请告诉我，以便进一步诊断问题。**
