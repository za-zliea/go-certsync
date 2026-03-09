MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
OUTPUT_DIR := $(MKFILE_DIR)output

.PHONY: build-all build docker docker-alpine push push-alpine clean version all release image image-push image-alpine image-alpine-push

build-all:
	if [ ! -d $(OUTPUT_DIR) ]; then mkdir $(OUTPUT_DIR); else rm -Rf $(OUTPUT_DIR)/*; fi
	go mod download
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-client_windows_x64.exe src/client/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o $(OUTPUT_DIR)/certsync-client_windows_x86.exe src/client/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-client_windows_arm64.exe src/client/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-client_linux_x64 src/client/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o $(OUTPUT_DIR)/certsync-client_linux_x86 src/client/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-client_linux_arm64 src/client/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-client_darwin_x64 src/client/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-client_darwin_arm64 src/client/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-server_windows_x64.exe src/server/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o $(OUTPUT_DIR)/certsync-server_windows_x86.exe src/server/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-server_windows_arm64.exe src/server/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-server_linux_x64 src/server/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o $(OUTPUT_DIR)/certsync-server_linux_x86 src/server/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-server_linux_arm64 src/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-server_darwin_x64 src/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_DIR)/certsync-server_darwin_arm64 src/server/main.go
build:
	if [ ! -d $(OUTPUT_DIR) ]; then mkdir $(OUTPUT_DIR); else rm -Rf $(OUTPUT_DIR)/*; fi
	go mod download
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-client src/client/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/certsync-server src/server/main.go
docker:
	docker pull ubuntu:focal
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client:$(VERSION)-ubuntu -f docker/Dockerfile-Client .; fi
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client:$(VERSION) -f docker/Dockerfile-Client .; fi
	docker build -t zliea/certsync-client:ubuntu -f docker/Dockerfile-Client .
	docker build -t zliea/certsync-client:latest -f docker/Dockerfile-Client .
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client-docker:$(VERSION)-ubuntu -f docker/Dockerfile-Client-Docker .; fi
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client-docker:$(VERSION) -f docker/Dockerfile-Client-Docker .; fi
	docker build -t zliea/certsync-client-docker:ubuntu -f docker/Dockerfile-Client-Docker .
	docker build -t zliea/certsync-client-docker:latest -f docker/Dockerfile-Client-Docker .
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-server:$(VERSION)-ubuntu -f docker/Dockerfile-Server .; fi
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-server:$(VERSION) -f docker/Dockerfile-Server .; fi
	docker build -t zliea/certsync-server:ubuntu -f docker/Dockerfile-Server .
	docker build -t zliea/certsync-server:latest -f docker/Dockerfile-Server .
docker-alpine:
	docker pull alpine:latest
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client:$(VERSION)-alpine -f docker/Dockerfile-Client-Alpine .; fi
	docker build -t zliea/certsync-client:alpine -f docker/Dockerfile-Client-Alpine .
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-client-docker:$(VERSION)-alpine -f docker/Dockerfile-Client-Docker-Alpine .; fi
	docker build -t zliea/certsync-client-docker:alpine -f docker/Dockerfile-Client-Docker-Alpine .
	if [ -n "$(VERSION)" ]; then docker build -t zliea/certsync-server:$(VERSION)-alpine -f docker/Dockerfile-Server-Alpine .; fi
	docker build -t zliea/certsync-server:alpine -f docker/Dockerfile-Server-Alpine .
push:
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client:$(VERSION)-ubuntu; fi
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client:$(VERSION); fi
	docker push zliea/certsync-client:ubuntu
	docker push zliea/certsync-client:latest
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client-docker:$(VERSION)-ubuntu; fi
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client-docker:$(VERSION); fi
	docker push zliea/certsync-client-docker:ubuntu
	docker push zliea/certsync-client-docker:latest
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-server:$(VERSION)-ubuntu; fi
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-server:$(VERSION); fi
	docker push zliea/certsync-server:ubuntu
	docker push zliea/certsync-server:latest
push-alpine:
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client:$(VERSION)-alpine; fi
	docker push zliea/certsync-client:alpine
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-client-docker:$(VERSION)-alpine; fi
	docker push zliea/certsync-client-docker:alpine
	if [ -n "$(VERSION)" ]; then docker push zliea/certsync-server:$(VERSION)-alpine; fi
	docker push zliea/certsync-server:alpine
clean:
	rm -Rf $(OUTPUT_DIR)
	go clean --cache
version:
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_windows_x64.exe $(OUTPUT_DIR)/certsync-client_$(VERSION)_windows_x64.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_windows_x86.exe $(OUTPUT_DIR)/certsync-client_$(VERSION)_windows_x86.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_windows_arm64.exe $(OUTPUT_DIR)/certsync-client_$(VERSION)_windows_arm64.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_linux_x64 $(OUTPUT_DIR)/certsync-client_$(VERSION)_linux_x64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_linux_x86 $(OUTPUT_DIR)/certsync-client_$(VERSION)_linux_x86; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_linux_arm64 $(OUTPUT_DIR)/certsync-client_$(VERSION)_linux_arm64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_darwin_x64 $(OUTPUT_DIR)/certsync-client_$(VERSION)_darwin_x64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-client_darwin_arm64 $(OUTPUT_DIR)/certsync-client_$(VERSION)_darwin_arm64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_windows_x64.exe $(OUTPUT_DIR)/certsync-server_$(VERSION)_windows_x64.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_windows_x86.exe $(OUTPUT_DIR)/certsync-server_$(VERSION)_windows_x86.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_windows_arm64.exe $(OUTPUT_DIR)/certsync-server_$(VERSION)_windows_arm64.exe; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_linux_x64 $(OUTPUT_DIR)/certsync-server_$(VERSION)_linux_x64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_linux_x86 $(OUTPUT_DIR)/certsync-server_$(VERSION)_linux_x86; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_linux_arm64 $(OUTPUT_DIR)/certsync-server_$(VERSION)_linux_arm64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_darwin_x64 $(OUTPUT_DIR)/certsync-server_$(VERSION)_darwin_x64; fi
	if [ -n "$(VERSION)" ]; then mv $(OUTPUT_DIR)/certsync-server_darwin_arm64 $(OUTPUT_DIR)/certsync-server_$(VERSION)_darwin_arm64; fi
all: clean build-all
release: clean build-all version
image: clean build docker
image-push: image push
image-alpine: clean build docker-alpine
image-alpine-push: image-alpine push-alpine
