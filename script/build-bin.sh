protoc \
	--proto_path=./internal/igrpc \
	--go_out=./internal/igrpc \
	--go_opt=paths=source_relative \
	--go-grpc_out=./internal/igrpc \
	--go-grpc_opt=paths=source_relative \
	./internal/igrpc/igrpc.proto

go build \
	-o ./bin/ ./cmd/*
