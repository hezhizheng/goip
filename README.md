# GOIP - IP 地址查询服务

一个基于 Merged-IP-Data 数据库，极简的 IP 地址查询服务，支持并发以及多语言、简化输出格式。

## 功能特性
- 完全免费，下载即使用
- 毫秒级响应，准确度高
- 两种输出格式：完整版 / 简化版
- 多语言支持：简体中文、英语、德语、法语、日语等
- 支持多 IP 并发查询（逗号分隔）
- 支持下载/更新 IP 数据库
- 下载进度条显示
- 提供免费服务 [https://ip.hzzio.top](https://ip.hzzio.top) 、 [https://ip.hzzio.top/s](https://ip.hzzio.top/s)（请勿滥用）

## 使用步骤

### 1. [下载编译好的文件启动](https://github.com/hezhizheng/goip/releases/)

### 2. 下载 IP 数据库

首次运行服务时，如果检测到本地没有 IP 数据库，会自动下载默认数据库：

```bash
# 启动服务 并自动下载数据库(不存在的情况下)
./goip.exe
```

国内服务器可使用 gh-proxy.com 代理下载，速度飞快
```bash
# 该命令仅下载数据库 服务需额外启动 
./goip.exe -d https://gh-proxy.com/https://github.com/NetworkCats/Merged-IP-Data/releases/latest/download/Merged-IP.mmdb
```

![](https://iili.io/CHaatDB.png)

**手动下载/更新方式**：

**方式一：下载默认数据库**
```bash
./goip.exe -d 1
```

**方式二：下载指定 URL 的数据库**
```bash
./goip.exe -d https://example.com/custom.mmdb
```

下载完成后会自动退出，数据库文件保存为 `Merged-IP.mmdb`。
可结合 crontab + [shell脚本](./goip.sh) 定时自动更新数据库

### 3. 启动服务

默认监听端口 8066，可通过 `-p` 参数指定端口：

```bash
./goip.exe -p 8080
```

### 4. API 接口

#### 完整版查询（默认）

**路由**：`/`

**参数**：
- `ip`：要查询的 IP 地址（多个用逗号分隔）

**示例**：
```bash
# 查询单个 IP
curl "http://127.0.0.1:8066/?ip=8.8.8.8"

# 批量查询
curl "http://127.0.0.1:8066/?ip=8.8.8.8,1.1.1.1,114.114.114.114"
```

**响应示例**：
```json
[
  {
    "ip": "8.8.8.8",
    "data": {
      "city": {...},
      "continent": {...},
      "country": {...},
      "location": {...},
      "subdivisions": [...],
      "asn": {...}
    }
  }
]
```

#### 简化版查询

**路由**：`/s` 或 `/s/{lang}`

**参数**：
- `ip`：要查询的 IP 地址（多个用逗号分隔）
- `{lang}`：语言代码（可选，默认为 `zh-CN`）

**支持的语言**：

| 语言代码 | 说明 |
|---------|------|
| zh-CN | 简体中文（默认） |
| en | 英语 |
| de | 德语 |
| es | 西班牙语 |
| fr | 法语 |
| ja | 日语 |
| pt-BR | 巴西葡萄牙语 |
| ru | 俄语 |

**示例**：
```bash
# 默认简体中文
curl "http://127.0.0.1:8066/s?ip=188.253.117.144"

# 英文输出
curl "http://127.0.0.1:8066/s/en?ip=8.8.8.8"

# 日语输出
curl "http://127.0.0.1:8066/s/ja?ip=1.1.1.1"

# 批量查询
curl "http://127.0.0.1:8066/s/zh-CN?ip=8.8.8.8,1.1.1.1,114.114.114.114"
```

**响应示例**：
```json
[
  {
    "organization": "Google LLC",
    "city": "Mountain View",
    "isp": "google.com",
    "asn_organization": "Google LLC",
    "latitude": 37.40599,
    "asn": 15169,
    "continent_code": "NA",
    "country": "美国",
    "timezone": "America/Los_Angeles",
    "country_code": "US",
    "longitude": -122.078514,
    "region": "加利福尼亚州",
    "ip": "8.8.8.8",
    "region_code": "CA"
  }
]
```

**简化版字段说明**：

| 字段 | 说明 |
|------|------|
| ip | IP 地址 |
| asn | AS 号码 |
| asn_organization | AS 组织名称 |
| organization | 组织名称（同 asn_organization） |
| isp | ISP 域名 |
| continent_code | 大洲代码 |
| country | 国家名称 |
| country_code | 国家代码 |
| region | 地区/省份 |
| region_code | 地区代码 |
| city | 城市名称 |
| latitude | 纬度 |
| longitude | 经度 |
| timezone | 时区 |

## 编译运行
### 可自行编译可执行文件(跨平台)
```
# 用法参考 https://github.com/mitchellh/gox
# 生成文件可直接执行 
gox -osarch="windows/amd64" -ldflags "-s -w" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}"
gox -osarch="darwin/amd64" -ldflags "-s -w" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}"
gox -osarch="linux/amd64" -ldflags "-s -w" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}"
gox -osarch="linux/arm64" -ldflags "-s -w" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}"
```

## 许可证
本项目采用 [MIT License](./LICENSE.txt) 开源许可证。

## 相关项目
- [Merged-IP-Data](https://github.com/NetworkCats/Merged-IP-Data) (特别感谢)