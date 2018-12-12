FROM golang as build
WORKDIR /go/src/github.com/gargous/flitter
COPY . .
RUN go get ./server
RUN cd ./server && go build main.go

FROM debian:stable-slim
COPY --from=build /go/src/github.com/gargous/flitter/server/main /app/
ENTRYPOINT ["/app/main"]