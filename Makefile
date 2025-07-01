.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/metrics.proto

.PHONY: testcoverage
testcoverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out   

