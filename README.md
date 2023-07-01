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

# 运行依赖

- postgres 数据库
- bun 连接直播间弹幕

# Todo

- [x] 不应该每次用户验证时创建一个直播间弹幕实例, 应该共用一个, 节省资源, 但当前为了实现简单这样做了
- [ ] 实现 OAuth Server .well-known, 公开公钥方便第三方验证 JWT Token
      (目前的展示服务器速度不是很理想, 等服务速度稳定后实现该接口, 让其他人也可以使用这个验证服务, 目前的话需要手动向我申请)
