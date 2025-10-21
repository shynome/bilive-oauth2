# bilive-oauth2

bilibili 直播间弹幕验证, 这是一个 oauth2 server

# 使用/用途

- 第三方应用发起 OAuth2 登录请求
- 进入到 BiliveAuth 网站进行验证
  - BiliveAuth 与用户网页创建一个 ws 连接
  - 用户前往指定的 Bilibili 直播间发送指定弹幕
  - BiliveAuth 收到指定弹幕后显示登录按钮
  - 用户点击登录按钮返回第三方应用
- 第三方应用收到 OAuth Code 进行交换获取 JWT Token
- 第三方应用使用公钥验证 JWT Token 的有效性后提供服务

# 作为 OAuth Server 使用

以 PocketBase 的 OpenID Connect provider 为例

| name          | url                                                                                |
| ------------- | ---------------------------------------------------------------------------------- |
| Auth URL      | https://bilive-auth.remoon.cn/                                                     |
| Token URL     | https://bilive-auth.remoon.cn/oauth/token                                          |
| User API URL  | https://bilive-auth.remoon.cn/oauth/whoami (uid@bilibili.com)                      |
| User API2 URL | https://bilive-auth.remoon.cn/oauth/live-open/user (openid@live-open.bilibili.com) |

ps: 如果你想使用该 OAuth Server, 可手动向我申请

# 使用身份码进行认证 (ID Token 中获取用户信息, 主播认证)

**如果使用了这个认证方式, 那么用户认证需要使用 `User API2 URL`**

| name      | url                                                                         |
| --------- | --------------------------------------------------------------------------- |
| Auth URL  | https://bilive-auth.remoon.cn/                                              |
| Token URL | https://bilive-auth.remoon.cn/oauth/id_code (openid@live-open.bilibili.com) |
| JWKS URL  | https://bilive-auth.remoon.cn/oauth/jwks.json                               |

[参考 PocketBase 的 Manual code exchange](https://pocketbase.io/docs/authentication/#authenticate-with-oauth2)

```js
pb.collection("users").authWithOAuth2Code(
  provider.name,
  `IDCode`,
  provider.codeVerifier,
  redirectURL, // 需是来自B站的链接, 包含 CodeSign, Timestamp, 超过20s将会认证失败
  {
    emailVisibility: false,
  }
)
```
