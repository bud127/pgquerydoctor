# Installation

## Build from source

```bash
git clone https://github.com/budisuryadi/pgquerydoctor.git
cd pgquerydoctor
go build -o pgquerydoctor ./cmd/pgquerydoctor
./pgquerydoctor version
```

## Run sample analysis

```bash
./pgquerydoctor analyze --query examples/query.sql --explain examples/explain.txt
```

## Generate Markdown report

```bash
./pgquerydoctor report \
  --query examples/slow-pagination.sql \
  --explain examples/slow-pagination.explain.txt \
  --output report.md
```

## Docker

```bash
docker build -t pgquerydoctor .
docker run --rm -v "$PWD/examples:/workspace" pgquerydoctor analyze --query /workspace/query.sql --explain /workspace/explain.txt
```
