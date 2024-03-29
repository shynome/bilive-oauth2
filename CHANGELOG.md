# Changelog

## [0.5.2] - 2024-03-26

- 修复: B 站弹幕链接现在需要应用心跳保活了

## [0.5.1] - 2024-03-04

- 修复代码错误: err 变量在错误的作用域引用

## [0.5.0] - 2024-02-15

- 添加 `/ws-info-keep`接口, 连接该 ws 接口时会持续发送心跳

## [0.4.4] - 2023-12-19

- 优化手机端登录指示文本

## [0.4.3] - 2023-11-11

- change: 改用 sse 推送事件, 因为阿里云 ws 配置没通过

## [0.4.2] - 2023-11-11

- fix: 修复 Dockerfile

## [0.4.1] - 2023-11-11

- 将大部分配置移到配置文件里了

## [0.4.0] - 2023-11-11

- 移除了数据库依赖, 使用配置文件加载 OAuth Clients
- 清除了不再使用的 danmu js bridge
- 升级 openapi-bilibili

## [0.3.0] - 2023-09-22

- 添加 '/bilibili/ws-info' 接口获取直播长连信息, 该接口需验证 jwt token

## [0.2.1] - 2023-09-21

### Improve

- 修复 css 没有预先加载好的问题

## [0.2.0] - 2023-09-21

### Change

- 使用[B 站官方开发平台的接口](https://open-live.bilibili.com/document/f9ce25be-312e-1f4a-85fd-fef21f1637f8)获取弹幕
- 运行依赖移除了 nodejs

## [0.1.1] - 2023-09-13

### Improve

- 登录服务页面已优化完成

## [0.1.0] - 2023-09-12

### Add

- 支持作为 PocketBase oidc provider 服务器使用

## [0.0.8] - 2023-07-03

### Fix

- `KeepLiveTCP` 也不行, 添加认证信息. 引入了 bilipage 服务依赖

## [0.0.7] - 2023-07-02

### Fix

- `KeepLiveTCP` 相较于 `KeepLiveWS` 貌似没有受到 b 站隐藏 UID 和昵称的影响

## [0.0.6] - 2023-07-02

### Fix

- 修复弹幕验证只有首次可被验证的问题

## [0.0.5] - 2023-07-02

### Improve

- 使用成熟的 js 直播弹幕库替换自己写的 golang 版, 额外要求 node 存在于服务器上
- 添加版本号

## [0.0.4] - 2023-03-13

### Change

- token 使用 ed25519 进行签名, 资源服务器可直接对 token 验证无需通过网络

## [0.0.3] - 2023-03-13

### Fix

- logout 不应清除 ReturnUri

## [0.0.2] - 2023-03-13

### Fix

- 支持 PCKE 认证
