# Default recipe to display help information
default:
    @just --list

# Build the dontrm binary
build:
    go build -ldflags="-s -w" -o dontrm .

# Install dontrm to system (requires sudo)
install: build
    sudo mv dontrm /usr/bin/dontrm
    @echo "dontrm installed to /usr/bin/dontrm"

# Run tests in Docker container (TESTS MUST RUN IN DOCKER FOR SAFETY)
test:
    @just _check-docker-rebuild
    @echo "Running tests in Docker container..."
    docker run --rm \
        -v $(pwd):/app \
        -w /app \
        dontrm-test:latest \
        go test -v -coverprofile=coverage.out -covermode=atomic ./...
    @echo "\nTest completed successfully!"

# Run tests with coverage report and enforce 85% coverage (main() excluded from realistic coverage)
coverage:
    @just _check-docker-rebuild
    @echo "Running tests with coverage in Docker..."
    docker run --rm \
        -v $(pwd):/app \
        -w /app \
        dontrm-test:latest \
        sh -c 'go test -v -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -func=coverage.out'
    @just _check-coverage

# Run End-to-End tests in Docker (tests actual binary with multiple shells)
e2e:
    @just _check-docker-e2e-rebuild
    @echo "Running E2E tests in Docker container..."
    docker run --rm dontrm-e2e:latest
    @echo "\nE2E tests completed successfully!"

# Run all tests (unit tests + E2E tests)
test-all:
    @echo "Running all tests..."
    @just test
    @just e2e
    @echo "\n✅ All tests passed!"

# Run linting (fmt + golangci-lint)
lint:
    @echo "Running go fmt..."
    go fmt ./...
    @echo "Running golangci-lint..."
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run; \
    else \
        echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
        exit 1; \
    fi

# Clean build artifacts and coverage reports
clean:
    rm -f dontrm coverage.out
    rm -f .dockertest.hash .dockere2e.hash
    @echo "Cleaned build artifacts"

# Force rebuild of Docker test image
rebuild-test-image:
    @echo "Force rebuilding Docker test image..."
    docker build -f Dockerfile.test -t dontrm-test:latest .
    @just _save-docker-hash

# Force rebuild of Docker E2E test image
rebuild-e2e-image:
    @echo "Force rebuilding Docker E2E test image..."
    docker build -f Dockerfile.e2e -t dontrm-e2e:latest .
    @just _save-docker-e2e-hash

# Internal: Check if Docker test image needs rebuilding
_check-docker-rebuild:
    #!/usr/bin/env bash
    set -euo pipefail

    DOCKERFILE="Dockerfile.test"
    HASHFILE=".dockertest.hash"

    # Calculate current hash of Dockerfile
    if command -v sha256sum >/dev/null 2>&1; then
        CURRENT_HASH=$(sha256sum "$DOCKERFILE" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        CURRENT_HASH=$(shasum -a 256 "$DOCKERFILE" | cut -d' ' -f1)
    else
        echo "Error: Neither sha256sum nor shasum found"
        exit 1
    fi

    # Check if we need to rebuild
    NEEDS_REBUILD=false

    if [ ! -f "$HASHFILE" ]; then
        NEEDS_REBUILD=true
    else
        SAVED_HASH=$(cat "$HASHFILE")
        if [ "$CURRENT_HASH" != "$SAVED_HASH" ]; then
            NEEDS_REBUILD=true
        fi
    fi

    # Check if image exists
    if ! docker image inspect dontrm-test:latest >/dev/null 2>&1; then
        NEEDS_REBUILD=true
    fi

    if [ "$NEEDS_REBUILD" = true ]; then
        echo "Dockerfile changed or image missing. Rebuilding Docker test image..."
        docker build -f Dockerfile.test -t dontrm-test:latest .
        echo "$CURRENT_HASH" > "$HASHFILE"
        echo "Docker test image rebuilt successfully!"
    else
        echo "Using cached Docker test image"
    fi

# Internal: Save Docker hash after manual rebuild
_save-docker-hash:
    #!/usr/bin/env bash
    set -euo pipefail

    DOCKERFILE="Dockerfile.test"
    HASHFILE=".dockertest.hash"

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$DOCKERFILE" | cut -d' ' -f1 > "$HASHFILE"
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$DOCKERFILE" | cut -d' ' -f1 > "$HASHFILE"
    fi

# Internal: Check coverage meets 85% threshold (main() excluded from realistic coverage)
_check-coverage:
    #!/usr/bin/env bash
    set -euo pipefail

    if [ ! -f coverage.out ]; then
        echo "Error: coverage.out not found"
        exit 1
    fi

    # Extract total coverage percentage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

    # Check if coverage meets threshold (85% accounts for untestable main() function)
    THRESHOLD=85.0

    if command -v bc >/dev/null 2>&1; then
        if [ $(echo "$COVERAGE < $THRESHOLD" | bc) -eq 1 ]; then
            echo ""
            echo "❌ Coverage is ${COVERAGE}% - below required ${THRESHOLD}%"
            exit 1
        else
            echo ""
            echo "✅ Coverage is ${COVERAGE}% - meets required ${THRESHOLD}%"
        fi
    else
        # Fallback for systems without bc
        COVERAGE_INT=$(echo "$COVERAGE" | cut -d'.' -f1)
        if [ "$COVERAGE_INT" -lt 85 ]; then
            echo ""
            echo "❌ Coverage is ${COVERAGE}% - below required ${THRESHOLD}%"
            exit 1
        else
            echo ""
            echo "✅ Coverage is ${COVERAGE}% - meets required ${THRESHOLD}%"
        fi
    fi

# Internal: Check if E2E Docker image needs rebuilding
_check-docker-e2e-rebuild:
    #!/usr/bin/env bash
    set -euo pipefail

    DOCKERFILE="Dockerfile.e2e"
    HASHFILE=".dockere2e.hash"

    # Calculate current hash of Dockerfile
    if command -v sha256sum >/dev/null 2>&1; then
        CURRENT_HASH=$(sha256sum "$DOCKERFILE" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        CURRENT_HASH=$(shasum -a 256 "$DOCKERFILE" | cut -d' ' -f1)
    else
        echo "Error: Neither sha256sum nor shasum found"
        exit 1
    fi

    # Check if we need to rebuild
    NEEDS_REBUILD=false

    if [ ! -f "$HASHFILE" ]; then
        NEEDS_REBUILD=true
    else
        SAVED_HASH=$(cat "$HASHFILE")
        if [ "$CURRENT_HASH" != "$SAVED_HASH" ]; then
            NEEDS_REBUILD=true
        fi
    fi

    # Check if image exists
    if ! docker image inspect dontrm-e2e:latest >/dev/null 2>&1; then
        NEEDS_REBUILD=true
    fi

    if [ "$NEEDS_REBUILD" = true ]; then
        echo "Dockerfile.e2e changed or image missing. Rebuilding Docker E2E image..."
        docker build -f Dockerfile.e2e -t dontrm-e2e:latest .
        echo "$CURRENT_HASH" > "$HASHFILE"
        echo "Docker E2E image rebuilt successfully!"
    else
        echo "Using cached Docker E2E image"
    fi

# Internal: Save E2E Docker hash after manual rebuild
_save-docker-e2e-hash:
    #!/usr/bin/env bash
    set -euo pipefail

    DOCKERFILE="Dockerfile.e2e"
    HASHFILE=".dockere2e.hash"

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$DOCKERFILE" | cut -d' ' -f1 > "$HASHFILE"
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$DOCKERFILE" | cut -d' ' -f1 > "$HASHFILE"
    fi
