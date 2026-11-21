 **webstack-golang 网址导航系统**
----------
#简介：

该系统是基于 Golang + sqlite（或mysql）的开源网址导航系统

使用前端：[WebStackPage](https://github.com/WebStackPage/WebStackPage.github.io)作为前端模板,作者：[viggo](https://www.viggoz.com/)

#演示地址：
[https://site.skyzgh.com](https://site.skyzgh.com)

#开源地址：
[https://github.com/skyzgh-cn/WebStack-golang](https://github.com/skyzgh-cn/WebStack-golang)

----------


#前台效果图：
![2025-05-06T14:17:46.png][1]

#后台效果图：
![2025-05-06T14:19:04.png][2]


----------
#部署方式

#方式一：下载可执行文件部署：

1.你可以直接从 [Releases ](https://github.com/skyzgh-cn/WebStack-golang/releases)下载预先编译好的二进制文件(压缩包)

2.解压后修改config.json中的服务器端口号，默认是sqlite数据库，可以改为mysql数据库

3.上传解压后的文件到服务器上

4.执行webstack启动服务（可参考下面如何用BT面板启动）


----------


#方式二：下载源码编译部署（需运行环境）

1.下载完整代码[源文件](https://github.com/skyzgh-cn/WebStack-golang/archive/refs/heads/master.zip)并修改好config.json中的服务器端口号，数据库相关信息

2.上传服务器，在上传目录下

执行 `go mod tidy` 拉取项目依赖库
 
执行编译 `go build -o webstack main.go` 编译生成可执行文件webstack

3.启动webstack（可参考下面如何用BT面板启动）



#通用BT面板部署步骤（方式一和方式二）

获取可执行二进制文件后（方式1直接下载，方式2是编译生成），以BT面板为例：

1.确保上传服务器的有webstack、config、default.sql三个文件，且config.json相关配置已经修改正确，首次运行后成功后，default.sql可以删除
![image](https://github.com/user-attachments/assets/4422594b-6ce5-4fe0-a493-c7a530765218)


2.在Bt面板网站->GO项目->添加GO项目，相关配置如图，找到刚刚上传的文件即可
![2025-05-06T14:46:04.png][4]

3.正常情况就部署成功了、后台地址:你的地址/admin 默认管理员账户和密码均为admin

#安全配置

**首次部署后请立即修改默认管理员密码**（后台 → 用户管理）。

生产环境建议通过环境变量覆盖敏感配置，避免写入版本库：

| 环境变量 | 说明 | 示例 |
| --- | --- | --- |
| `WEBSTACK_CONFIG` | 指定 config.json 的绝对路径 | `/opt/webstack/config.json` |
| `WEBSTACK_MYSQL_PASSWORD` | MySQL 密码 | `s3cret` |
| `WEBSTACK_MYSQL_USER` / `WEBSTACK_MYSQL_HOST` / `WEBSTACK_MYSQL_PORT` / `WEBSTACK_MYSQL_DATABASE` | MySQL 连接信息 | |
| `WEBSTACK_DB_TYPE` | 数据库类型（`sqlite` / `mysql`） | `mysql` |
| `WEBSTACK_SQLITE_FILE` | SQLite 数据库文件路径 | `./db/webstack.db` |
| `WEBSTACK_SECURE_COOKIE` | 站点启用 HTTPS 时置为 `true` | `true` |
| `WEBSTACK_TRUSTED_PROXIES` | 可信反向代理 IP/CIDR，多个用逗号分隔 | `127.0.0.1` |
| `WEBSTACK_HOST` / `WEBSTACK_PORT` | 监听地址与端口 | `0.0.0.0` |

说明：

- `trustedProxies` 留空表示不信任任何代理，此时 `X-Forwarded-*` 头一律被忽略。若站点部署在 Nginx 等反向代理之后，请把代理的 IP 填入该列表，否则客户端 IP 会被识别为代理地址（影响登录限流）。
- `secureCookie` 为 `true` 时会话 Cookie 仅通过 HTTPS 传输，站点启用 HTTPS 后应开启。
- 站内"抓取网页信息"功能已内置 SSRF 防护，禁止访问内网地址与非 80/443 端口。

#方式三：Docker部署

从 Docker Hub 拉取已发布镜像运行（`skyzgh/webstack:latest`，始终指向最新版本；也可改用固定版本号如 `v2.1.2`）。

仓库内已提供部署文件 [docker/webstack/docker-compose.yml](docker/webstack/docker-compose.yml)，下载后按以下步骤操作：

 ```
# 1. 进入部署目录
cd /www/wwwroot/site.skyzgh.com/app/docker/webstack

# 2. 生成配置文件并填写数据库密码等实际值
cp config.json.example config.json
vi config.json

# 3. 创建持久化目录并修正属主（容器内 webstack 用户为 uid=100, gid=101）
mkdir -p data/upload data/db
sudo chown -R 100:101 data

# 4. 启动
docker compose up -d
docker compose logs -f
  ```

部署文件使用自定义 bridge 网络，容器固定 IP 为 `172.28.0.2`，网关为 `172.28.0.1`。
应用端口 `8080` 只绑定到宿主机的 `127.0.0.1`，外部无法直接访问，由宿主机 Nginx 反向代理进入。

MySQL 授权（容器 IP 固定，只需授权这一个来源）：

```
CREATE USER IF NOT EXISTS 'webstack'@'172.28.0.2' IDENTIFIED BY '密码';
GRANT ALL PRIVILEGES ON webstack.* TO 'webstack'@'172.28.0.2';
FLUSH PRIVILEGES;
```

同时确认 MySQL 的 `bind-address` 不是 `127.0.0.1`（否则会拒绝来自容器的连接），并保证防火墙未放行 3306 公网访问。

Nginx 反向代理配置：

```
server {
    listen 443 ssl http2;
    server_name site.skyzgh.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用 HTTPS 后，需将 `config.json` 中的 `secureCookie` 改为 `true`，
并将 `trustedProxies` 保持为网关地址 `172.28.0.1`（容器内观察到的代理来源 IP）。

持久化说明：

| 宿主机路径 | 容器内路径 | 说明 |
| --- | --- | --- |
| `./config.json` | `/app/config.json` | 站点配置（含数据库密码），只读挂载 |
| `./data/upload` | `/app/upload` | 用户上传的文件，如站点 Logo |
| `./data/db` | `/app/db` | SQLite 数据库文件（使用 MySQL 时为空目录） |

[1]: https://blog.skyzgh.com/usr/uploads/2025/05/1645397260.png
  [2]: https://blog.skyzgh.com/usr/uploads/2025/05/3665417208.png
  [3]: https://blog.skyzgh.com/usr/uploads/2025/05/1055621062.png
  [4]: https://blog.skyzgh.com/usr/uploads/2025/05/31828951.png
