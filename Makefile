GOBINARY=craurl

.PHONY: build
build:
	go build -o ./dist/$(GOBINARY) ./cmd

.PHONY: run
run: build
	./dist/$(GOBINARY) ./testdata/urls.txt

.PHONY: generate-mocks
generate-mocks: ## Generates mocks for the tests, using mockery tool
	mockery --name=storer --dir=./crawler --output=./crawler --outpkg=crawler --inpackage --structname=mockStorer --filename=mock_storer.go

.PHONY: test
test:
	go test -v -race ./crawler/...

.PHONY: distclean
distclean:
	rm -rf ./dist
