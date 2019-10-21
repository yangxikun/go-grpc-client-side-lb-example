build:
	cd greeter_client && GOOS=linux GOARCH=amd64 go build -mod=vendor . && cd ..
	cd greeter_server && GOOS=linux GOARCH=amd64 go build -mod=vendor . && cd ..
	docker build -t docker.io/yangxikun/go-grpc-client-side-lb-example:0.1 .
	docker push docker.io/yangxikun/go-grpc-client-side-lb-example:0.1
