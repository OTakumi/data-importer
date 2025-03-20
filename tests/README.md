# Integration Tests for MongoDB JSON Importer

This directory contains integration tests that test the full functionality of the MongoDB JSON Importer by connecting to a real MongoDB instance.

## Setup

### Option 1: Using .env.test File

1. Create a `.env.test` file in the project root by copying the provided sample:
   ```bash
   cp .env.test.sample .env.test
   ```

2. Edit `.env.test` with your MongoDB test instance credentials:
   ```
   MONGODB_USERNAME=your_test_user
   MONGODB_PASSWORD=your_test_password
   MONGODB_HOST=localhost
   MONGODB_PORT=27017
   ```

### Option 2: Using Environment Variables

You can directly set environment variables when running the tests:

```bash
TEST_MONGODB_URI="mongodb://username:password@localhost:27017" go test ./tests/integration
```

## Running the Tests

### Run all integration tests:

```bash
# Using go test directly
go test -v ./tests/integration DOTENV_PATH=dotenv/filepath/.env.test

# Or using the Makefile target
make test-integration DOTENV_PATH=dotenv/filepath/.env.test
```

### Run a specific test:

```bash
go test -v ./tests/integration -run TestIntegration/ImportArrayJSON
```

## Test Data

The tests look for JSON test files in the `testdata` directory. Make sure that directory contains:

- `users_array.json` - A valid array-formatted JSON file
- `product_object.json` - A valid single-object JSON file
- `invalid.json` - An invalid JSON file for testing error handling

## Troubleshooting

### Authentication Errors

If you see authentication errors like:

```
integration_test.go:92: Failed to import array JSON: error importing documents to collection users_array: Command insert requires authentication
```

Make sure your MongoDB credentials in `.env.test` are correct and have write permissions to the test database.

### Test Data Not Found

If the test data files are not found, the tests will be skipped. Ensure the `testdata` directory is accessible and contains the required files.
