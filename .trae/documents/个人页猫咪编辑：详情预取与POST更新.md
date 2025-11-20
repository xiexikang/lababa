## 目标
- 点击“编辑”时，先拉取猫咪详情填充表单。
- 猫咪信息保存改为 POST 请求，并在请求体中传递宠物 `id`。

## 后端改造（server-go）
- 新增详情接口：`GET /api/cats/detail/:id`
  - 鉴权（从 `Authorization` 解析当前用户）。
  - 仅允许访问本人猫：`SELECT ... FROM cats WHERE id=? AND userId=?`；返回 `cat`。
  - 参考已存在列表/更新接口的返回结构（如 `cats/list` 与 `cats/update`）。
- 新增更新接口：`POST /api/cats/update`
  - 鉴权。
  - 请求体：`{ id, name?, breedId?, avatarUrl?, gender?, birthDate?, weightKg?, neutered?, notes? }`
  - 校验 `id` 属于当前用户，构造动态 `UPDATE` 语句更新传入字段，返回最新 `cat`。
  - 保留现有 `PUT /api/cats/update/:id`（或标记为废弃），前端迁移到 POST 版本。

## 前端改造
- 页面：`src/pages/cats/index.vue`
  - 详情预取：
    - 原逻辑：`loadForEdit()` 使用 `GET /api/cats/list` 并在列表中查找目标项（文件: g:\www-xxk\xcx-lababa\src\pages\cats\index.vue:128-144）。
    - 修改为：`GET /api/cats/detail/${editId}` 直接获取详情；将返回字段映射到表单（名称/性别/生日/体重/绝育/备注）。
  - 保存请求：
    - 原逻辑：编辑分支使用 `PUT /api/cats/update/:id`（文件: g:\www-xxk\xcx-lababa\src\pages\cats\index.vue:163-164）。
    - 修改为：`POST /api/cats/update`，body 包含 `id: editId` 与各字段（生日按当前格式转换为时间戳）。
- 入口：`src/pages/profile/index.vue`
  - 保持导航到 `/pages/cats/index?id=...`（文件: g:\www-xxk\xcx-lababa\src\pages\profile\index.vue:266-269），详情预取在编辑页完成，无需在点击前调用接口，避免双请求和不必要耦合。

## 校验与兼容
- 请求鉴权：复用 `src/utils/request.ts` 自动添加 `Authorization`。
- 字段一致性：沿用现有页面字段命名；生日字符串转毫秒时间戳。
- 兼容旧接口：后端保留 `PUT /api/cats/update/:id`（可标注 deprecated），前端统一采用 POST；后端将新增 POST，以免影响旧版本客户端。

## 交付与测试
- 后端：实现并编译运行，手动调用：
  - `GET /api/cats/detail/:id` 返回对应 `cat`。
  - `POST /api/cats/update` 成功更新并返回 `cat`。
- 前端：
  - 编辑页进入时看到表单已被详情填充。
  - 修改后保存成功并返回之前页面（按已有 `redirect` 逻辑）。
- 边界：
  - 非本人猫或不存在 `id`→返回 403/404；前端提示“保存失败/加载失败”。
