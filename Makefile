export GOBIN = $(shell pwd)/toolbin

LINT = $(GOBIN)/golangci-lint
FORMAT = $(GOBIN)/goimports

MOCKGEN = $(GOBIN)/mockgen

.PHONY: require-tools
require-tools: tools
	@echo 'Checking installed tools...'
	@file $(LINT) > /dev/null
	@file $(FORMAT) > /dev/null

	@echo "All tools found in $(GOBIN)!"

.PHONY: tools
tools:
	@echo 'Installing tools...'
	@rm -rf toolbin
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
	@go install golang.org/x/tools/cmd/goimports@v0.1.11

	@go install github.com/golang/mock/mockgen@5b455625bd2c8ffbcc0de6a0873f864ba3820904

.PHONY: fmt
fmt: require-tools
	@go mod tidy
	@$(FORMAT) -w $$(go list -f {{.Dir}} ./...)

.PHONY: lint
lint: require-tools fmt
	@$(LINT) run ./...

.PHONY: mocks
mocks:
	$(MOCKGEN) -source etherclient/client.go -destination etherclient/mocks/mock_client.go
	$(MOCKGEN) -source etherclient/contractbackend/contract_backend.go -destination etherclient/contractbackend/mocks/mock_contract_backend.go
	$(MOCKGEN) -source aws/sns.go -destination aws/mocks/mock_sns.go
	$(MOCKGEN) -source aws/sqs.go -destination aws/mocks/mock_sqs.go
	$(MOCKGEN) -source aws/s3.go -destination aws/mocks/mock_s3.go
	$(MOCKGEN) -source aws/ses.go -destination aws/mocks/mock_ses.go
	$(MOCKGEN) -source aws/dynamodb.go -destination aws/mocks/mock_dynamodb.go
	$(MOCKGEN) -source store/dynamo/store.go -destination store/dynamo/mocks/mock_dynamo.go
	$(MOCKGEN) -source feeds/interfaces.go -destination feeds/mocks/mock_feeds.go

.PHONY: test
test:
	go test -v -count=1 -short -coverprofile=coverage.out ./...

.PHONY: coverage
coverage:
	go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'

.PHONY: coverage-func
coverage-func:
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html:
	go tool cover -html=coverage.out -o=coverage.html
