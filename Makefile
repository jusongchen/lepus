
PROJECT?=github.com/jusongchen/lepus
PORT?=8080

RELEASE?=0.3.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

win32:
	GOOS=windows GOARCH=386 go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o lepus.exe

	zip -r lepus-win32.zip lepus.exe
	# return 0 if when no file to rm
	rm public/images/* || true 
	zip -r lepus-win32.zip public/
	zip -r lepus-win32.zip views/
	mv lepus-win32.zip ./dist-win32/lepus-win32.zip

osx:
	GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o lepus

	zip -r lepus-osx.zip lepus
	# return 0 if when no file to rm
	rm public/images/* || true 
	zip -r lepus-osx.zip public/
	zip -r lepus-osx.zip views/
	mv lepus-osx.zip ./dist-osx/lepus-osx.zip
run: osx
	mkdir ../lepus-tmp || true
	rm -fr ../lepus-tmp/lepus-osx || true 
	unzip ./dist-osx/lepus-osx.zip -d ../lepus-tmp/lepus-osx
	open http://localhost:8081
	cd ../lepus-tmp && 	lepus-osx/lepus -port 8081


win32run: 
	git clone https://github.com/jusongchen/lepus.git
	cd dist-win32
	unzip lepus-win32.zip -d lepus-win32
	open -o http://localhost:8082
	lepus-win32\lepus.exe -p 8082
	
win64run: 
	git clone https://github.com/jusongchen/lepus.git
	cd dist-win64
	unzip lepus-win64.zip -d lepus-win64 
	open -o http://localhost:8082
	lepus-win64\lepus.exe -p 8082
	

test:
	go test -v -race ./...
