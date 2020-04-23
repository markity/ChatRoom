# 简单介绍

这是一个简单的聊天室程序，通信方面基于`TCP协议`, 支持心跳检测。服务端使用`Golang`开发，客户端采用`C++ Qt`开发。希望通过这个项目练练手

![程序图片](https://cloud-netdisk.oss-cn-chengdu.aliyuncs.com/%E9%A1%B9%E7%9B%AE%E5%9B%BE%E7%89%87ChatRoom1.png)

# 如何运行?

**软件安装**

* Golang
* Qt Creator以及相应C++编译器

**服务端运行**

```sh
$ cd ./Server
$ go run ./main.go
2020/04/23 08:23:06 server started at 127.0.0.1:8000...
```

**客户端运行**

1. 打开`Client`文件夹
2. 使用`Qt Creator`打开`ChatClient.pro`文件
3. 点击`Qt Creator`左下角的`运行`按钮