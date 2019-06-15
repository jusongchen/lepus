补丁 Patch V1.0.2 安装说明
==============

补丁的目的是用下载的新版的selectphoto.html文件更新views子目录里的selectphoto.html。
具体操作如下：

## 下载新版文件

1) 点击下载这个文件
https://github.com/jusongchen/lepus/releases/download/v1.0.2/patch_v1.0.2.zip

 2) 把下载文件解压缩。 解压缩会生成文件 selectphoto.html

## 用新版文件替换现有文件
3) 在运行Lepus这台服务器上用文件浏览器打开 Lepus 安装文件夹
4) 打开子文件夹 views
5) 把文件 selectphoto.html  改名为 selectphoto.v1.0.1.html （备份用，如新版不工作，需要用这个来恢复到原版本）
6) 把步骤 2）解压缩生成的 selectphoto.html 拷贝到当前文件夹 （也就是文件夹 views ）

## 重启动 Lepus
7) 退出运行中的 Lepus
8) 重新运行 runLepus.bat 来启动Lepus

## 测试
* 上传一张相片测试一下 (测试用姓名：TEST)
* 如测试失败， 把views文件夹内selectphoto.html 换成原来的selectphoto.v1.0.1.html。 然后重启 Lepus.

注：可以不用重启 Caddy.exe


