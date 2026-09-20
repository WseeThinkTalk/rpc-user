## rpc 生成命令
rpc:
	goctl rpc protoc user.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=./user --go-grpc_out=./user --zrpc_out=./ -m --verbose --style=go_zero

run:
	go run user.go

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/rpc-user user.go
