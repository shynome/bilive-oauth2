VERSION := $(shell git describe --tags --always --dirty)

build-fo:
	npm run --prefix frontend build && \
	rm -rf cmd/bilive-auth/build/ && \
	cp -r frontend/build/ cmd/bilive-auth/build/
build:
	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=${VERSION}' -s -w" -o bilive-auth .
docker: build
	docker build . -t shynome/bilive-auth:${VERSION}
push: docker
	docker push shynome/bilive-auth:${VERSION}
