gobuild:
	go install .
	go install ./fltool
gotest:
	go test ./test -v
proto:
	protoc --gofast_out=. share/proto/*.proto && go run tool/msggen.go share/proto/msg.json && go install ./share/proto
rt:
	go run client/main.go
front:
	go run server/main.go