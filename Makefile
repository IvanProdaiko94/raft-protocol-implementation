.PHONY: gen

gen:
#	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
#	@go get -u github.com/golang/protobuf/protoc-gen-go

	@protoc -I/usr/local/include -I. \
      -I$$GOPATH/src \
      -I$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
      --go_out=plugins=grpc:. \
      --grpc-gateway_out=logtostderr=true:. \
      schema/service.proto