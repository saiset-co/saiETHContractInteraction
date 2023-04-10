up:
	docker-compose -f ./microservices/docker-compose.yml up -d

down:
	docker-compose -f ./microservices/docker-compose.yml down --remove-orphans

build:
	make service
	make docker

service:
		cd ./src/saiETHContractInteraction && go mod tidy && go build -o ./../../microservices/saiETHContractInteraction/build/sai-eth-interaction
		cp ./src/saiETHContractInteraction/config.yml ./microservices/saiETHContractInteraction/build/config.yml
		cp ./src/saiETHContractInteraction/contracts.json ./microservices/saiETHContractInteraction/build/contracts.json

docker:
	docker-compose -f ./microservices/docker-compose.yml up -d --build

log:
	docker-compose -f ./microservices/docker-compose.yml logs -f

logi:
	docker-compose -f ./microservices/docker-compose.yml logs -f sai-eth-interaction


