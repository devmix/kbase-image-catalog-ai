package main

import (
	"context"
	"fmt"
	"kbase-catalog/internal/images"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
	"kbase-catalog/internal/webserver"
	"kbase-catalog/web"

	"github.com/spf13/cobra"
)

var (
	archiveDirFlag string
	useFilesystem  bool
	// web flags
	portFlag int

	// Convert images flags
	qualityFlag   int
	originDirFlag string

	rootCmd = &cobra.Command{
		Use:   "kbase-catalog",
		Short: "KBase Image Catalog tool",
		Long:  `A tool for managing image catalogs with LLM-powered processing and image conversion capabilities.`,
	}

	processCmd = &cobra.Command{
		Use:   "process <path to images catalog>",
		Short: "Process the catalog starting from root directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Load configuration
			cfg, err := config.LoadConfig("")
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			imagesCatalog := args[0]

			// Create processor
			catalogProcessor := processor.NewCatalogProcessor(cfg, imagesCatalog)

			fmt.Printf("Processing catalog in: %s\n", imagesCatalog)

			err = catalogProcessor.ProcessCatalog(ctx)
			if err != nil {
				log.Fatalf("Failed to process catalog: %v", err)
			}

			err = catalogProcessor.RebuildRootIndex(ctx)
			if err != nil {
				log.Fatalf("Failed to rebuild root index: %v", err)
			}
		},
	}

	rebuildIndexCmd = &cobra.Command{
		Use:   "rebuild-index",
		Short: "Rebuild the root index.json file",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Load configuration
			cfg, err := config.LoadConfig("")
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			// Create processor
			catalogProcessor := processor.NewCatalogProcessor(cfg, archiveDirFlag)

			fmt.Printf("Rebuilding root index in: %s\n", archiveDirFlag)

			err = catalogProcessor.RebuildRootIndex(ctx)
			if err != nil {
				log.Fatalf("Failed to rebuild root index: %v", err)
			}
		},
	}

	testCmd = &cobra.Command{
		Use:   "test <image_path>",
		Short: "Test single image processing",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Load configuration
			cfg, err := config.LoadConfig("")
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			// Create processor
			catalogProcessor := processor.NewCatalogProcessor(cfg, archiveDirFlag)

			imagePath := args[0]
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
		},
	}

	convertImagesCmd = &cobra.Command{
		Use:   "convert-images",
		Short: "Convert images to WebP format",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Load configuration
			cfg, err := config.LoadConfig("")
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			// Create converter
			imageConverter := images.NewImageConverter(cfg)

			fmt.Printf("Converting images in: %s\n", archiveDirFlag)

			err = imageConverter.ConvertImages(ctx, archiveDirFlag, originDirFlag, qualityFlag)
			if err != nil {
				log.Fatalf("Failed to convert images: %v", err)
			}
		},
	}

	webCmd = &cobra.Command{
		Use:   "web",
		Short: "Start web interface",
		Run: func(cmd *cobra.Command, args []string) {
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

			// Create processor
			catalogProcessor := processor.NewCatalogProcessor(cfg, archiveDirFlag)

			fmt.Println("Starting web interface...")

			web.InitTemplateFS(useFilesystem)

			server := webserver.NewServer(cfg, catalogProcessor, portFlag, archiveDirFlag)

			err = server.Start()
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
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("KBase Image Catalog v0.1.0")
		},
	}
)

func init() {
	descriptionArchiveDir := "Directory to use for archive files"

	// Convert images flags
	convertImagesCmd.Flags().IntVarP(&qualityFlag, "quality", "q", 85, "WebP compression quality (0-100, default: 85)")
	convertImagesCmd.Flags().StringVarP(&originDirFlag, "origin-dir", "o", "origin", "Directory to move original files to")
	convertImagesCmd.Flags().StringVarP(&archiveDirFlag, "archive-dir", "a", "archive", descriptionArchiveDir)

	// web flags
	webCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "Port to run the web server on")
	webCmd.Flags().BoolVarP(&useFilesystem, "use-fs", "l", false, "Use real filesystem for static resources instead of embedded")
	webCmd.Flags().StringVarP(&archiveDirFlag, "archive-dir", "a", "archive", descriptionArchiveDir)

	// rebuild index flags
	rebuildIndexCmd.Flags().StringVarP(&archiveDirFlag, "archive-dir", "a", "archive", descriptionArchiveDir)

	// Add commands
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(rebuildIndexCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(convertImagesCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
