proto:
	protoc --gofast_out=. res/proto/*.proto && go run tool/msggen.go res/proto/msg.json