GOBINARY=craurl

all: build

.PHONY: build
build:
	go build -o ./dist/$(GOBINARY) ./cmd

.PHONY: run
run: build
	./dist/$(GOBINARY) ./testdata/urls.txt

.PHONY: docker
docker:
	docker build -t craurl .

.PHONY: generate-mocks
generate-mocks: ## Generates mocks for the tests, using mockery tool
	mockery --name=storer --dir=./crawler --output=./crawler --outpkg=crawler --inpackage --structname=mockStorer --filename=mock_storer.go

.PHONY: test
test:
	go test -race ./...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...

.PHONY: cover-html
cover-html: cover
	go tool cover -html=coverage.out -o coverage.html

.PHONY: dep
dep:
	go mod download

.PHONY: distclean
distclean:
	rm -rf ./dist

