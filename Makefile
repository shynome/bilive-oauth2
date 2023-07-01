build-fo:
	npm run --prefix frontend build && \
	rm -rf cmd/bilive-auth/build/ && \
	cp -r frontend/build/ cmd/bilive-auth/build/
build: build-fo
	go generate ./...
	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$$(git describe --tags --always --dirty | cut -c2-)' -s -w" ./cmd/bilive-auth/
sync: build
	rsync -rP ./frontend/build/ remoon-sh-1:/opt/bilive-auth/frontend
	rsync -rP bilive-auth remoon-sh-1:/opt/bilive-auth/bilive-auth
