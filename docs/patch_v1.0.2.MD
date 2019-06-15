补丁 Patch V1.0.2 安装说明
==============

补丁的目的是更新views子目录里的文件 selectphoto.html. 具体操作如下：

## 下载新版文件

1) 点击下载这个文件
https://github.com/jusongchen/lepus/releases/download/v1.0.2/patch_v1.0.2.zip

 2) 把下载文件解压缩。 解压缩会生成文件 selectphoto.html

## 用新版文件替换现有文件
3) 在运行Lepus这台服务器上用文件浏览器打开 Lepus 安装目录
4) 打开子目录 views
5) 把文件 selectphoto.html  改名为 selectphoto.v1.0.1.html 
6) 把步骤 2）解压缩生成的 selectphoto.html 拷贝到 子目录 views 中

## 重启动 Lepus
7) 退出运行中的 Lepus
8）重新运行 runLepus.bat 来启动Lepus

## 上传一张相片测试一下

注：可以不用重启 Caddy.exe


