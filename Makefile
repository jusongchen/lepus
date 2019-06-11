
PROJECT?=github.com/jusongchen/lepus
PORT?=8080

RELEASE?=0.6.0
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

win64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o lepus.exe

	7z a lepus-win64.7z lepus.exe caddy.exe runLepus.bat CaddyFile views/
	# return 0 if when no file to rm
	rm public/images/* || true 
	7z a lepus-win64.7z public/  -xr!*DS_Store
	mv lepus-win64.7z ./dist-win64/lepus-win64.7z

	# deliver as zip format as well
	# return 0 if when no file to rm
	rm public/images/* || true 

	7z a  -tzip lepus-win64.zip lepus.exe caddy.exe runLepus.bat CaddyFile  views/
	7z a  lepus-win64.zip public/  -xr!*DS_Store
	mv lepus-win64.zip ./dist-win64/lepus-win64.zip

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
	mkdir ./dist-osx || true
	mv lepus-osx.zip ./dist-osx/lepus-osx.zip


run: osx
	mkdir ../lepus-tmp || true
	rm -fr ../lepus-tmp/lepus-osx || true 
	unzip ./dist-osx/lepus-osx.zip -d ../lepus-tmp/lepus-osx
	open http://localhost:8081
	cd ../lepus-tmp && LEPUS_SESSION_KEY=IUY!YHG@GBE#VFR4ytk5nhs6jwh7hni8	lepus-osx/lepus -port 8081


test:
	go test -v -race ./...
