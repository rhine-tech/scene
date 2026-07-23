# Storage

Storage 用 `StorageKey`（`{provider}://{identifier}`）定位文件。Provider 只负责存储能力，
不知道 HTTP API 的挂载路径。

## 文件访问

`GET /storage/data/:provider/*fileid` 始终由 Scene 代理传输文件。

`GET /storage/url/:provider/*fileid` 用于获取可访问 URL：

- 不传 `mode` 或 `mode=proxy`：返回对应的 `/storage/data/...` 路径。返回值会保留
  Gin 容器的实际挂载前缀，例如 `/api/storage/url/...` 会得到
  `/api/storage/data/...`。
- `mode=direct`：调用 `IStorageService.GetDirectURL`，由 provider 返回直连 URL。
  不支持直连的 provider 返回 `ErrDirectURLUnsupported`；生成失败返回
  `ErrGetDirectURLFailed`。

Local provider 不支持 direct URL。S3 provider 的 direct URL 是有时效的预签名 URL，
有效期由 `storage.s3.presigned_url_ttl_seconds` 配置。

Provider 不需要配置 URL prefix；HTTP 路径只由 delivery 层生成。

### 访问权限

当前 `FileMeta` 没有 Public/Private 或 ACL 字段，因此所有文件都按 **Private** 处理；
storage 暂不提供匿名公开文件。

`/data` 的实际代理下载和 `/url` 的 URL 获取都要求
`storage:file:download`（`PermFileDownload`）权限：

- `mode=proxy` 返回的 `/data` 路径只表示文件位置，不代表授权；客户端请求该路径时会
  再次鉴权。
- `mode=direct` 只在鉴权通过后生成预签名 URL。预签名 URL 在有效期内是 bearer
  credential，后续由 provider 直接校验签名，不再经过 Scene 权限中间件。

### Proxy 传输

`/data` 使用请求级 `io.ReadSeekCloser` 适配 `http.ServeContent`。第一次 `Read` 才调用
provider 的 ranged `Load`；后续顺序读取复用同一个 reader，并在请求结束时统一关闭。
`Seek` 只更新逻辑位置，下一次 `Read` 才按新位置重新打开 reader。

因此普通 GET 和单 Range 请求各只会 open/close 一次；HEAD 不会打开数据流。只有客户端
显式请求多个 Range 时，才会按 Range 分别打开 reader。delivery 会在调用
`ServeContent` 前设置元数据中的 `Content-Type`（为空时使用
`application/octet-stream`），避免 MIME sniff 触发一次额外的读取和回退。

## Context

Service、provider、metadata repository 和 upload session tracker 的 I/O 方法都以
`context.Context` 为第一个参数。HTTP 调用使用 `Request.Context()`，并一直传递到
GORM、S3 和 Redis。Provider 和 repository 可以返回标准 context error；service
边界会将其映射为当前操作对应的 storage errcode，保持 service 只返回 errcode 的约定。

Context 只表达取消和超时，不负责并发一致性，也不替代锁。Local filesystem 的 syscall
本身不接收 context，因此只能在操作开始前、流式复制的分块之间和提交点前观察取消，
无法中断一个已经阻塞的 filesystem syscall。

远端写入、删除或 multipart 完成已经成功后，元数据和 session 的必要收尾会使用保留
request value、脱离 request cancellation 且最多持续 5 秒的 context。正常前向 I/O
始终服从请求取消。

## todo

- [x] local
- [ ] Aliyun oss
- [ ] Tencent cos
- [ ] Qiniu
- [x] aws s3 (S3-compatible, e.g. RustFS)
