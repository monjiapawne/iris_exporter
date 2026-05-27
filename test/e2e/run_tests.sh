#!/usr/bin/env bash
# Usage:
# ./run_tests.sh v2.4.20 

set -eo pipefail

cleanup() {
	echo "Cleaning up containers..."
	docker compose down
}

trap cleanup EXIT

# Main

main() {
    if [ -z "$1" ]; then
        echo "usage: $0 v2.4.20"
        exit 1
    fi

    cd "$(dirname "$0")"

    export APP_IMAGE_TAG=${APP_IMAGE_TAG:-$1}
    export DB_IMAGE_TAG=${DB_IMAGE_TAG:-$1}
    export APP_IMAGE_NAME=${APP_IMAGE_NAME:-ghcr.io/dfir-iris/iriswebapp_app}
    export DB_IMAGE_NAME=${DB_IMAGE_NAME:-ghcr.io/dfir-iris/iriswebapp_db}

    # Pull/build containers and run tests.
    docker compose down -v # ensure clean slate 
    docker compose up -d --build --wait

    # Run tests, store output
    if go test -count=1 -v ./...; then
        echo "$1,pass" >> test_log.csv
    else
        echo "$1,fail" >> test_log.csv
        exit 1
    fi
}

main "$@"
