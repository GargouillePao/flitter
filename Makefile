gobuild:
	go install ./core
gotest:
	go test ./test -v
proto:
	protoc --gofast_out=. share/proto/*.proto && go run tool/msggen.go share/proto/msg.json
docker:
	[ "$$(docker ps -a | grep flitter-test)" ] && (echo "rebuild" && docker rm flitter-test -f && docker rmi flitter:v1 -f) || (echo "new build")
	docker build -t flitter:v1 -f Dockerfile .
	docker run --name flitter-test -it -p 8080:8080 -p 8081:8081 flitter:v1
rt:
	go run client/main.go
front:
	go run server/main.go