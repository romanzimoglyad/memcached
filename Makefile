build:
	docker build -t service -f ./DockerFile .

run:
	docker-compose up -d
stop:
	docker-compose down