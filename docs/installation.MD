
# Lepus

“老师，谢谢你!” 活动专用程序

#  安装与运行

## 下载软件包

在一台和互联网联通的电脑上点击下面的链接下载 lepus-win64.zip 或 lepus-win64.7z

https://github.com/jusongchen/lepus/releases

## 安装   
1. 将上面下载的 lepus-win64.zip(或lepus-win64.7z) 拷贝到任何一台个人机或Windows服务器的一个文件夹（这个文件夹就是Lepus的安装目录）
2. 将 lepus-win64.zip(或lepus-win64.7z) 解压缩。 解压后会生成 lepus.exe 和 其他文件夹
3. 运行 runLepus.bat  这个批命令会启动lepus.exe服务器程序（默认监听 TCP 端口 8080）。另外 runLepus.bat把服务器程序的日志转存到文件 Lepus.log

## 本机测试

* 在运行 lepus.exe 的机器上,通过网络浏览器访问网址 http://127.0.0.1:8080
* 如一切顺利， 网络浏览器会显示 "老师,谢谢你!" 和其他内容。

## 在网址为lsnh.fjyxyz.net的机器上运行caddy.exe 来支持https

在 lsnh.fjyxyz.net 这台机器上：
* 打开一个新的命令行窗口
* 用 cd 命令将当前目录换到Lepus的安装目录
* 运行 caddy.exe
* 如一切顺利，caddy.exe 将会驻留（不会自己退出）
* 测试用https协议访问网站： 从任何联网的手机或机器访问 https://lsnh.fjyxyz.net 

