# Variables
BINARY=build/paye-ton-kawa--products
DOCKER_IMAGE=ghcr.io/payetonkawa-epsi-2025/products/paye-ton-kawa--products
SRC=cmd/main.go

build: $(BINARY)

$(BINARY): $(SRC)
	@mkdir -p build
	GOPRIVATE=github.com/PayeTonKawa-EPSI-2025/* GOOS=linux GOARCH=amd64 go build -o $@ $<

build-image: build
	@if [ -z "$(VERSION)" ]; then \
	  echo "Error: VERSION variable is required. Use: make build-image VERSION=<tag>" >&2; \
	  exit 1; \
	fi

	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

clean:
	rm -rf build
