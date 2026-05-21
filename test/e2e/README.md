# End-to-end tests

## How to run

```bash
./run_tests.sh v2.4.20
```

For testing you can instead just run

```bash
docker compose build exporter
docker compose up -d
curl localhost:10043/metric
# Clean up
docker compose down # there are no named volumes to remove
```

## Overview

1) Pulls minimal containers for IRIS
2) Builds exporter image
3) Run tests in `e2e_test.go`
4) Stores test results in `test_logs.csv`