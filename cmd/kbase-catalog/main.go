package main

import (
	"context"
	"flag"
	"fmt"
	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
	"kbase-catalog/internal/web"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse flags for web service
	webPort := flag.Int("port", 8080, "Port to run the web server on")
	archiveDir := flag.String("archive-dir", "archive", "Directory to use for archive files (default: 'archive')")

	flag.Parse()

	// Create processor
	catalogProcessor := processor.NewCatalogProcessor(cfg, *archiveDir)
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("KBase Image Catalog")
		fmt.Println("Usage: kbase-catalog [command]")
		fmt.Println("Commands:")
		fmt.Println("  process <root_dir> - Process the catalog starting from root directory")
		fmt.Println("  test <image_path>  - Test single image processing")
		fmt.Println("  version           - Show version information")
		fmt.Println("  web               - Start web interface")
		return
	}

	command := args[0]
	switch command {
	case "process":
		fmt.Printf("Processing catalog in: %s\n", *archiveDir)

		err = catalogProcessor.ProcessCatalog(ctx, *archiveDir)
		if err != nil {
			log.Fatalf("Failed to process catalog: %v", err)
		}

	case "test":
		if len(args) < 2 {
			fmt.Println("Error: test command requires an image path")
			return
		}
		imagePath := args[1]
		fmt.Printf("Testing single image: %s\n", imagePath)

		response, err := catalogProcessor.TestSingleImage(ctx, imagePath)
		if err != nil {
			log.Fatalf("Failed to test image: %v", err)
		}

		if response != nil {
			fmt.Printf("\n✅ Successfully obtained result:\n")
			fmt.Printf("Short name: %s\n", response.ShortName)
			fmt.Printf("Description: %s\n", response.Description)
		}

	case "version":
		fmt.Println("KBase Image Catalog v0.1.0")

	case "web":
		fmt.Println("Starting web interface...")

		server := web.NewServer(cfg, catalogProcessor, *webPort, *archiveDir)

		err := server.Start()
		if err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}

		// Wait for shutdown signal
		<-ctx.Done()
		fmt.Println("Shutting down gracefully...")
		err = server.Stop(ctx)
		if err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Use 'help' for available commands")
	}
}
