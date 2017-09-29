.PHONY: clean build

build: psyche psyche.linux

psyche: main.go
	go build -o $@

psyche.linux: GOOS=linux
psyche.linux: GOARCH=amd64
psyche.linux: main.go
	go build -o $@

docker-build: psyche.linux
	docker build -t docker.atl-paas.net/dkrishnamurthy/psyche:0.0.0 .

docker-run: docker-build
	docker run -ti --rm -p 8080:8080 docker.atl-paas.net/dkrishnamurthy/psyche:0.0.0

docker-push: docker-build
	docker push docker.atl-paas.net/dkrishnamurthy/psyche:0.0.0

deploy: docker-push
	DOCKER_IMAGE=docker.atl-paas.net/dkrishnamurthy/psyche:0.0.0 DOCKER_TAG=0.0.0 micros service:deploy psyche -f psyche.sd.yml

clean:
	rm -f psyche psyche.linux
