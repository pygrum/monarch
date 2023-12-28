.PHONY: server linux macos windows pb

LDFLAGS = -s -w
PROTOS = "pkg/protobuf"

server:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o monarch-linux_x64 cmd/server/monarch-server.go
	GOOS=linux GOARCH=386 go build -ldflags="$(LDFLAGS)" -o monarch-linux_x86 cmd/server/monarch-server.go
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o monarch-linux_arm64 cmd/server/monarch-server.go
	GOOS=linux GOARCH=arm go build -ldflags="$(LDFLAGS)" -o monarch-linux_arm cmd/server/monarch-server.go

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o monarch-client-linux_x64 cmd/client/monarch-client.go
	GOOS=linux GOARCH=386 go build -ldflags="$(LDFLAGS)" -o monarch-client-linux_x86 cmd/client/monarch-client.go
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o monarch-client-linux_arm64 cmd/client/monarch-client.go
	GOOS=linux GOARCH=arm go build -ldflags="$(LDFLAGS)" -o monarch-client-linux_arm cmd/client/monarch-client.go

macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o monarch-client-macos_x64 cmd/client/monarch-client.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o monarch-client-macos_arm cmd/client/monarch-client.go

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o monarch-client-win_x64.exe cmd/client/monarch-client.go
	GOOS=windows GOARCH=386 go build -ldflags="$(LDFLAGS)" -o monarch-client-win_x86.exe cmd/client/monarch-client.go

pb:
	cd $(PROTOS) && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		clientpb/client.proto
	cd $(PROTOS) && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		rpcpb/services.proto
	cd $(PROTOS) && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		builderpb/builder.proto