 **webstack-golang 网址导航系统**
----------
#简介：

该系统是基于 Golang + mysql的开源网址导航系统

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

1.你可以直接从 [Releases ](https://github.com/skyzgh-cn/WebStack-golang/releases)下载预先编译好的二进制文件

2.修改config.json中的服务器端口号，mysql相关信息

3.上传webstack和config

4.执行webstack启动服务


----------


#方式二：下载源码编译部署（需Go和Mysql环境）

1.下载完整代码[源文件](https://github.com/skyzgh-cn/WebStack-golang/archive/refs/tags/v1.0.0.zip)并修改好config.json中的服务器端口号，mysql相关信息

2.上传服务器，在上传目录下

 执行 `go mod tidy` 拉取项目依赖库
 
执行编译 `go build -o webstack main.go` 编译生成可执行文件main

3.启动main

#通用BT面板部署步骤

获取可执行二进制文件后（方式1直接下载，方式2是编译生成），以BT面板为例：

1.确保上传服务器的有webstack和config，且config.json相关配置已经修改正确
![2025-05-06T14:44:23.png][3]

2.在Bt面板网站->GO项目->添加GO项目，相关配置如图，找到刚刚上传的文件即可
![2025-05-06T14:46:04.png][4]

3.正常情况就部署成功了、后台地址:你的地址/admin 默认管理员账户和密码均为admin


  [1]: https://blog.skyzgh.com/usr/uploads/2025/05/1645397260.png
  [2]: https://blog.skyzgh.com/usr/uploads/2025/05/3665417208.png
  [3]: https://blog.skyzgh.com/usr/uploads/2025/05/1055621062.png
  [4]: https://blog.skyzgh.com/usr/uploads/2025/05/31828951.png
