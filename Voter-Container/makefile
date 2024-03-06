SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build                            Build the Voter executable"
	@echo "	   build-up                         Build the Voter container and run it from docker-compose"
	@echo "	   build-container                  Build the Voter container"
	@echo "	   build-multi-arch-containers      (Extra-Credit) Build multiple Voter containers for multiple architectures"
	@echo "	   run                              Run the voter program from code"
	@echo "	   run-compose                      Run the voter program from docker-compose"
	@echo "	   run-bin                          Run the voter executable"
	@echo "	   load-voters                      Add sample voters via curl"
	@echo "	   get-by-id                        Get a todo by id pass id=<id> on command line"
	@echo "	   get-all                          Get all todos"
	@echo "	   update-3                         Update record 3, pass a new name in using name=<name> on command line"
	@echo "	   delete-all                       Delete all voter"
	@echo "	   delete-by-id                     Delete a voter by id pass id=<id> on command line"
	@echo "	   build-amd64-linux                Build amd64/Linux executable"
	@echo "	   build-arm64-linux                Build arm64/Linux executable"





.PHONY: build
build:
	go build -C voter-api

.PHONY: build-container
build-container:
	docker build -t voter-api:latest .

.PHONY: build-up
build-up:
	docker build -t voter-api:latest .; docker-compose up

.PHONY: build-multi-arch-containers
build-multi-arch-containers: 
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 -t voter-api:latest .

.PHONY: build-amd64-linux
build-amd64-linux:
	GOOS=linux GOARCH=amd64 go build -C voter-api -o ./voter-linux-amd64 .

.PHONY: build-arm64-linux
build-arm64-linux:
	GOOS=linux GOARCH=arm64 go build -C voter-api -o ./voter-linux-arm64 .

	
.PHONY: run
run:
	go run -C voter-api main.go

.PHONY: run-compose
run-compose:
	docker-compose up

.PHONY: run-bin
run-bin:
	./voter-api/voter-api

.PHONY: load-voters
load-voters:
	curl -d '{"voterId":1,"name":"Ayden Corwin","email":"pedrostroman@damore.io"}' -H "Content-Type: application/json" -X POST http://localhost:8080/voters/1
	curl -d '{"voterId":2,"name":"Ila Stehr","email":"hazlelarson@mante.io"}' -H "Content-Type: application/json" -X POST http://localhost:8080/voters/2
	curl -d '{"voterId":3,"name":"John Doe","email":"john.doe@gmail.com"}' -H "Content-Type: application/json" -X POST http://localhost:8080/voters/3
	

.PHONY: update-3
update-3:
	curl -d '{"name": "${name}"}' -H "Content-Type: application/json" -X PUT http://localhost:8080/voters/3

.PHONY: get-by-id
get-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:8080/voters/$(id) 

.PHONY: get-all
get-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:8080/voters 

.PHONY: delete-all
delete-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:8080/voters

.PHONY: delete-by-id
delete-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:8080/voters/$(id) 
