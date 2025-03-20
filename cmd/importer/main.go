package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/OTakumi/data-importer/internal/config"
	"github.com/OTakumi/data-importer/internal/domain"
	"github.com/OTakumi/data-importer/internal/repository"
	"github.com/OTakumi/data-importer/internal/service"
	"github.com/OTakumi/data-importer/internal/utils"
)

func main() {
	// Parse command line arguments
	var showHelp bool
	var envFile string
	flag.BoolVar(&showHelp, "help", false, "Show usage information")
	flag.BoolVar(&showHelp, "h", false, "Show usage information (shorthand)")
	flag.StringVar(&envFile, "env", ".env", "Path to .env file")
	flag.Parse()

	// Display help
	if showHelp || flag.NArg() == 0 {
		printUsage()
		os.Exit(0)
	}

	// Set the env file path to use
	if envFile != ".env" {
		os.Setenv("DOTENV_PATH", envFile)
	}

	// Get the first argument as import path
	importPath := flag.Arg(0)

	// Initialize configuration
	cfg := config.NewConfig()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Setup signal handling (for graceful shutdown when Ctrl+C is pressed)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		cancel()
		os.Exit(1)
	}()

	// Initialize MongoDB repository
	repo, err := repository.NewMongoRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize file utilities
	fileUtils := utils.NewFileUtils(nil) // Use actual file system

	// Initialize importer service
	importer := service.NewMongoImporter(ctx, fileUtils, repo, cfg.BatchSize)

	// Execute import process
	startTime := time.Now()
	fmt.Printf("Starting import: %s\n", importPath)
	fmt.Printf("Using MongoDB: %s, Database: %s\n", cfg.MongoURI, cfg.DatabaseName)

	result, err := importer.ImportPath(importPath)
	if err != nil {
		log.Fatalf("Error during import process: %v", err)
	}

	// Display results
	displayResults(result, time.Since(startTime))
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("MongoDB JSON Importer")
	fmt.Println("Usage: importer [options] <file-path or directory-path>")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nEnvironment Variables (can be set in .env file):")
	fmt.Println("  MONGODB_URI        - MongoDB connection URI (default: mongodb://mongodb:27017)")
	fmt.Println("  MONGODB_DATABASE   - Database name (default: test_db)")
	fmt.Println("  MONGODB_TIMEOUT    - Timeout in seconds (default: 10)")
	fmt.Println("  MONGODB_BATCH_SIZE - Batch size for imports (default: 1000)")
}

// displayResults displays the results of the import process
func displayResults(result any, duration time.Duration) {
	switch r := result.(type) {
	case *domain.ImportResult:
		// Display results for a single file
		fmt.Printf("\nImport results for file '%s':\n", r.FileName)
		fmt.Printf("  Collection: %s\n", r.CollectionName)
		fmt.Printf("  Documents inserted: %d\n", r.InsertedCount)
		fmt.Printf("  Processing time: %v\n", r.Duration)
		if r.Error != nil {
			fmt.Printf("  Error: %v\n", r.Error)
		}

	case []*domain.ImportResult:
		// Display results for a directory (multiple files)
		fmt.Printf("\nDirectory import results (%d files):\n", len(r))

		totalDocuments := 0
		successCount := 0
		errorCount := 0

		for _, res := range r {
			totalDocuments += res.InsertedCount
			if res.Error == nil {
				successCount++
				fmt.Printf("  ✓ %s -> %s (%d documents, %v)\n",
					res.FileName, res.CollectionName, res.InsertedCount, res.Duration)
			} else {
				errorCount++
				fmt.Printf("  ✗ %s -> Error: %v\n", res.FileName, res.Error)
			}
		}

		fmt.Printf("\nTotal: %d documents, %d files succeeded, %d files failed\n",
			totalDocuments, successCount, errorCount)
	}

	fmt.Printf("\nTotal processing time: %v\n", duration)
}
