win32:
	GOOS=windows GOARCH=386 go build -o lepus.exe 
	zip -r lepus-win32.zip lepus.exe 
	zip -r lepus-win32.zip public/
	mv lepus-win32.zip ./dist-win32
