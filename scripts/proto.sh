#!/bin/bash
set -e
protoc \
	-I=. \
	-I=/usr/local/include \
	--go_out=. proto/*.proto \

go build -o ./bin/gen-cmdid cmd/proto/gen-cmdid/main.go

protoc \
  -I=. \
  -I=/usr/local/include \
  --plugin=protoc-gen-cmdid-gen=./bin/gen-cmdid \
  --cmdid-gen_out=proto/pb \
  proto/*.proto

echo "✅ 协议编译完成"


go run ./cmd/proto/gen-router/main.go
