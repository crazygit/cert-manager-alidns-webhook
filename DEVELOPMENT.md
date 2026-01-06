# 开发指南

## 📋 项目概述

基于 [cert-manager/webhook-example](https://github.com/cert-manager/webhook-example) 模板，开发适用于阿里云 DNS (AliDNS) 的 cert-manager webhook。

### 核心特性

- ✅ 支持 **RRSA** (RAM Roles for Service Accounts) - 生产环境推荐
- ✅ 支持 **AccessKey/SecretKey** - 开发/测试环境
- ✅ 支持 **ECS 实例 RAM 角色** - ACK 自动支持
- ✅ 使用 **V2.0 Tea SDK** - 官方推荐版本
- ✅ **幂等性** DNS 记录管理
- ✅ **Helm Chart** 一键部署

## 运行测试用例

### 单元测试

```
$ go test -v ./pkg/alidns/...
```

### 集成测试

⚠️ **注意**：

集成测试会通过 API 操作阿里云解析的域名记录，运行时最好使用一个**非生产环境**的域名测试。

前提条件：

- 已经有域名托管在阿里云解析
- 参考[管理访问凭证](https://help.aliyun.com/zh/sdk/developer-reference/v2-manage-go-access-credentials), 在本地配置好了访问凭证的环境变量或`config.json`文件

```shell
TEST_ZONE_NAME=example.com. make test
```

替换上面命令中 `example.com.` 为你当前托管在阿里云用于测试的域名（不要忘记域名后面的 `.`）

## 🔗 参考资源

- [阿里云 Golang SDK 配置](https://next.api.aliyun.com/api-tools/sdk/Alidns?version=2015-01-09&language=go-tea&tab=primer-doc)
- [管理访问凭证](https://help.aliyun.com/zh/sdk/developer-reference/v2-manage-go-access-credentials)
- [Endpoint 设置](https://api.aliyun.com/product/Alidns)
- [Cert-Manager Creating DNS Providers](https://cert-manager.io/docs/contributing/dns-providers/)
- [Cert-Manager webhook-example](https://github.com/cert-manager/webhook-example)
