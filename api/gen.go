//go:generate protoc --go_out=../go-auth/internal/generated --go_opt=paths=source_relative --go-grpc_out=../go-auth/internal/generated --go-grpc_opt=paths=source_relative auth.proto

package api
