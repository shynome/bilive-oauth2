build-fo:
	npm run --prefix frontend build && \
	rm -rf cmd/bilive-auth/build/ && \
	cp -r frontend/build/ cmd/bilive-auth/build/
build: build-fo
	go generate ./...
	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$$(git describe --tags --always --dirty)' -s -w" -o bilive-auth ./cmd/bilive-auth/
docker: build
	docker build . -t shynome/bilive-auth:$$(git describe --tags --always --dirty)
push: docker
	docker push shynome/bilive-auth:$$(git describe --tags --always --dirty)
