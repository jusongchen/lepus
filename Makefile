win32:
	GOOS=windows GOARCH=386 go build -o lepus.exe

	zip -r lepus-win32.zip lepus.exe
	rm public/images/*
	zip -r lepus-win32.zip public/
	zip -r lepus-win32.zip views/
	mv lepus-win32.zip ./dist-win32/lepus-win32.zip
