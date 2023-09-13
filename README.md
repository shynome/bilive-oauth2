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

| name         | url                                        |
| ------------ | ------------------------------------------ |
| Auth URL     | https://bilive-auth.remoon.cn/             |
| Token URL    | https://bilive-auth.remoon.cn/oauth/token  |
| User API URL | https://bilive-auth.remoon.cn/oauth/whoami |

ps: 如果你想使用该 OAuth Server, 可手动向我申请

# 运行依赖

- postgres 数据库
- nodejs 连接直播间弹幕
