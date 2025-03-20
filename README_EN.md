# MongoDB JSON Importer

A Go tool for efficiently importing JSON files into MongoDB. Supports processing individual files or entire directories, with batch processing capabilities for large datasets.

## Features

- Import single JSON files into MongoDB collections
- Recursively process multiple JSON files within directories
- Automatically set collection names based on filenames
- Support for both array-format and single-object JSON documents
- Efficient batch processing for document imports
- Flexible configuration via environment variables or .env files
- Easy execution in Docker environments

## Requirements

- Go 1.20 or higher
- MongoDB (local or remote)
- Docker and Docker Compose (optional)

## Installation

```bash
# Clone the repository
git clone https://github.com/OTakumi/data-importer.git
cd data-importer

# Build
make build
```

## Usage

### Command Line

```bash
# Import a single JSON file
./mongodb-importer path/to/file.json

# Import all JSON files in a directory
./mongodb-importer path/to/directory

# Use a custom environment config file
./mongodb-importer -env=custom.env path/to/file.json

# Show help
./mongodb-importer --help
```

### Running with Docker

```bash
# Build and run
make docker-build
make docker-up

# Import a JSON file
make import-file file=path/to/file.json

# Import a directory
make import-dir dir=path/to/directory

# Stop
make docker-down
```

## Configuration

You can customize settings using environment variables or a .env file.

### Environment Variables

#### MongoDB Connection Settings (Option 1: Individual Components)
- `MONGODB_USERNAME`: MongoDB username
- `MONGODB_PASSWORD`: MongoDB password
- `MONGODB_HOST`: MongoDB hostname (default: "mongodb")
- `MONGODB_PORT`: MongoDB port number (default: "27017")
- `MONGODB_AUTH_DATABASE`: Authentication database name (optional)
- `MONGODB_REPLICA_SET`: Replica set name (optional)

#### MongoDB Connection Settings (Option 2: Direct URI)
- `MONGODB_URI`: MongoDB connection URI (default: `mongodb://mongodb:27017`)
- `MONGODB_DATABASE`: Database name to use (default: `test_db`)

#### Application Settings
- `MONGODB_TIMEOUT`: Timeout in seconds (default: `10`)
- `MONGODB_BATCH_SIZE`: Batch size for imports (default: `1000`)

### .env File

You can create a .env file to set environment variables:

```
# MongoDB Connection Settings (Option 1: Individual Components)
MONGODB_USERNAME=admin
MONGODB_PASSWORD=password123
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_AUTH_DATABASE=admin

# MongoDB Connection Settings (Option 2: Direct URI)
#MONGODB_URI=mongodb://admin:password123@localhost:27017/?authSource=admin

# Database and Application Settings
MONGODB_DATABASE=import_db
MONGODB_TIMEOUT=30
MONGODB_BATCH_SIZE=500
```

## Testing

```bash
# Run all tests
make test

# Run integration tests
make integration-test

# Generate coverage report
make coverage
```

For integration tests, create a `.env.test` file or set environment variables. See [integration test README](tests/integration/README.md) for details.

## Project Structure

```
mongodb-importer/
├── cmd/
│   └── importer/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── domain/
│   │   └── models.go            # Domain models
│   ├── repository/
│   │   └── mongodb.go           # Data access layer
│   ├── service/
│   │   └── importer.go          # Business logic layer
│   └── utils/
│       └── fileutils.go         # File operation utilities
├── tests/
│   ├── integration/             # Integration tests
│   └── testdata/                # Test data files
├── Dockerfile
├── docker-compose.yaml
├── .env.sample                  # Environment variables sample
├── Makefile
└── README.md
```

## Future Development Plans

### Short-term Improvements
1. **Implement Streaming Processing**
   - Improve memory efficiency for large JSON files
   - Introduce JSON streaming parser

2. **Optimize Parallel Processing**
   - Improve efficiency of parallel processing for multiple files
   - Optimize resource usage

3. **Detailed Progress Reporting**
   - Implement real-time progress bars
   - Display processing speed and statistics

### Medium to Long-term Expansion Plans
1. **Data Transformation Features**
   - JSON schema transformation capabilities
   - Field name and data type conversion

2. **Support for Other Data Sources**
   - Support for other formats like CSV, XML, YAML
   - Export capabilities to various databases

3. **Add GUI Interface**
   - Web-based management interface
   - Drag-and-drop file operations

## License

[MIT License](LICENSE)
