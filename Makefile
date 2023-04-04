up:
	docker-compose -f ./microservices/docker-compose.yml up -d

down:
	docker-compose -f ./microservices/docker-compose.yml down --remove-orphans

build:
	make service
	make docker

service:
		cd ./src/saiETHContractInteraction && go mod tidy && go build -o ./../../microservices/saiETHContractInteraction/build/sai-eth-interaction

docker:
	docker-compose -f ./microservices/docker-compose.yml up -d --build

log:
	docker-compose -f ./microservices/docker-compose.yml logs -f

logi:
	docker-compose -f ./microservices/docker-compose.yml logs -f sai-eth-interaction


