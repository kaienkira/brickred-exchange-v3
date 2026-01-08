BUILD_RELEASE_FLAGS = -trimpath -ldflags="-s -w"

build:
	@cd src && go build -o ../bin/brexc

build-linux:
	@cd src && GOOS=linux arch=amd64 go build $(BUILD_RELEASE_FLAGS) -o ../bin/brexc

build-windows:
	@cd src && GOOS=windows arch=amd64 go build $(BUILD_RELEASE_FLAGS) -o ../bin/brexc.exe

fmt:
	@cd src && go fmt ./...

clean:
	@rm -rf bin/
