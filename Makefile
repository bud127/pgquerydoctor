APP_NAME=pgquerydoctor

.PHONY: build test run example clean

build:
	go build -o $(APP_NAME) ./cmd/pgquerydoctor

test:
	go test ./...

run:
	go run ./cmd/pgquerydoctor analyze --query examples/query.sql --explain examples/explain.txt

example:
	go run ./cmd/pgquerydoctor report --query examples/query.sql --explain examples/explain.txt --output report.md

clean:
	rm -f $(APP_NAME) report.md
