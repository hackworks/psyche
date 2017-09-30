.PHONY: all psyche

version=0.0.1

all: psyche

psyche:
	go build -o $@
	go test ./...

psyche.linux: GOOS=linux
psyche.linux: GOARCH=amd64
psyche.linux: main.go
	go build -o $@

docker-build: psyche.linux
	docker build -t docker.atl-paas.net/dkrishnamurthy/psyche:$(version) .

docker-run: docker-build
	docker run -ti --rm -p 8080:8080 docker.atl-paas.net/dkrishnamurthy/psyche:$(version)

docker-push: docker-build
	docker push docker.atl-paas.net/dkrishnamurthy/psyche:$(version)

deploy: docker-push
	DOCKER_IMAGE=docker.atl-paas.net/dkrishnamurthy/psyche:$(version) DOCKER_TAG=$(version) micros service:deploy psyche -f psyche.sd.yml

clean:
	rm -f psyche psyche.linux
