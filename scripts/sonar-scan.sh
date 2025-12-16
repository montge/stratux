#!/bin/bash
# Run SonarCloud analysis locally
# Requires: sonar-scanner installed and SONAR_TOKEN environment variable set

set -e

cd "$(dirname "$0")/.."

# Check for sonar-scanner
if ! command -v sonar-scanner &> /dev/null; then
    echo "Error: sonar-scanner not installed" >&2
    echo "Install with: brew install sonar-scanner (macOS) or download from sonarcloud.io" >&2
    exit 1
fi

# Check for token
if [[ -z "$SONAR_TOKEN" ]]; then
    echo "Error: SONAR_TOKEN environment variable not set" >&2
    echo "Add to ~/.bashrc: export SONAR_TOKEN=\"your-token\"" >&2
    exit 1
fi

# Generate coverage report
echo "Running tests with coverage..."
cd main
go test -run "^(Test[^EH]|TestHandle[A-O]|TestHandle(Set|Sa|Si|St|Ti|To|Tr|We))" -coverprofile=../coverage.out -coverpkg=. 2>&1 || true
cd ..

# Run sonar-scanner
echo "Running SonarCloud analysis..."
sonar-scanner -Dsonar.token="$SONAR_TOKEN"

echo "Analysis complete! View results at https://sonarcloud.io/project/overview?id=montge_stratux"
