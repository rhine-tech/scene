SCENE_EXECUTABLE = scene-server

protogen:
	@find . -type d | while read -r dir; do \
		if ls $$dir/*.proto 1> /dev/null 2>&1; then \
			cd $$dir; \
            protoc --go_out=. --go-grpc_out=. *.proto; \
            cd ..; \
		fi \
	done

prebuild: protogen
	mkdir -p dist

build:
	go build -o ./$(SCENE_EXECUTABLE) ./app

build-linux: prebuild
	GOOS=linux GOARCH=amd64 go build -o ./dist/$(SCENE_EXECUTABLE)-linux-amd64 ./app

build-windows: prebuild
	GOOS=windows GOARCH=amd64 go build -o ./dist/$(SCENE_EXECUTABLE)-windows-amd64.exe ./app

build-macos: prebuild
	GOOS=darwin GOARCH=amd64 go build -o ./dist/$(SCENE_EXECUTABLE)-darwin-amd64 ./app

build-all: build-linux build-windows build-macos

run: build
	./$(SCENE_EXECUTABLE)

clean:
	rm -f ./$(SCENE_EXECUTABLE)
	rm -rf ./release
	rm -rf ./dist

