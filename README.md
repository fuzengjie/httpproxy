[![Build Status](https://travis-ci.org/yryz/httpproxy.svg?branch=master)](https://travis-ci.org/yryz/httpproxy)

使用Golang实现的HTTP代理转shadowsocks，主要为命令行下`go get`、`docker pull`、`npm install`、`pip install`、`gem install`、`curl`等程序提供HTTP代理服务，解决安装总是失败的问题。这些服务不支持shadowsocks，但对http代理都有支持。

## 原理
应用 <-------HTTP/1.1-------> httpproxy <-------加密-------> 你的shadowsocks服务器

## 使用
**安装**

`go get github.com/fuzengjie/httpproxy`


配置文件 ~/.httpproxy/config.json（这里是默认配置，也可以通过 httpproxy -c config.json 来指定）

```
{
        "listen": "127.0.0.1:6666",
        "ss_server": "ip:port",
        "ss_cipher": "aes-128-cfb",
        "ss_password": "your password",
	"auth":[{"user":"your username","pwd":"your password"}]
}
```
启动 `httpproxy`

**使用代理**

如果想命令行一直走代理，下面配置加入到 ~/.bash_profile

```
http_proxy=http://127.0.0.1:6666
https_proxy=http://127.0.0.1:6666
```

如果只是想临时使用，可以手动设置http_proxy环境变量或者 使用`httpproxy set` 快速就地设置（不影响全局，推荐！）。
推荐方式二：修改~/.bash_profile设置别名`alias proxy="http_proxy=http://127.0.0.1:6666"`，使用时可以 proxy curl ip.cn 

## 特点

* 支持与shadowsocks服务桥接
* 支持CONNECT，支持HTTPS、HTTP2代理
* 简单易用、命令行友好
* 支持用户验证
