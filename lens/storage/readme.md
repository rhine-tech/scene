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

## todo

- [x] local
- [ ] Aliyun oss
- [ ] Tencent cos
- [ ] Qiniu
- [x] aws s3 (S3-compatible, e.g. RustFS)
