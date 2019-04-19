rm -fr lepus
git clone https://github.com/jusongchen/lepus.git
cd lepus/dist-win32
unzip lepus-win32.zip -d lepus-win32
open -o http://localhost:8082
lepus-win32/lepus.exe -port 8082
