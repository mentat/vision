test:
	docker run -d --name=dev-consul consul
	go test
	docker stop dev-consul