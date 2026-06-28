#!/bin/bash
protoc \
	-I=. \
	--plugin=protoc-gen-go=/root/go/bin/protoc-gen-go \
	--go_out=. proto/*.proto \

go build -o ./bin/gen-cmdid cmd/proto/gen-cmdid/main.go

protoc \
  -I=. \
  --plugin=protoc-gen-cmdid-gen=./bin/gen-cmdid \
  --cmdid-gen_out=proto/pb \
  proto/*.proto

echo "Success 协议编译完成"


go run ./cmd/proto/gen-router/main.go
