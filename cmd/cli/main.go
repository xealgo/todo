package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"charm.land/huh/v2"
	"github.com/xealgo/todo/internal/config"
	"github.com/xealgo/todo/internal/scanner"
	"github.com/xealgo/todo/internal/scanner/generators"
)

const banner = `
 ______   ______     _____     ______    
/\__  _\ /\  __ \   /\  __-.  /\  __ \   
\/_/\ \/ \ \ \/\ \  \ \ \/\ \ \ \ \/\ \  
   \ \_\  \ \_____\  \ \____-  \ \_____\ 
    \/_/   \/_____/   \/____/   \/_____/ 
-------------------------------------------
`

// Options struct holds the command-line options for the application.
type Options struct {
	sourcePath string
	run        bool
	outputs    []string
}

func main() {
	fmt.Print("\033[H\033[2J")
	fmt.Printf(banner)

	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("Failed to load configuration: " + err.Error())
		os.Exit(1)
	}

	if err = os.Mkdir(cfg.ArtifactsPath, 0644); err != nil && !os.IsExist(err) {
		slog.Error("Failed to create artifacts directory: " + err.Error())
		os.Exit(1)
	}

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle shutdown signal
	go func() {
		<-sigChan
		cancel()
	}()

	var options Options

	form := newForm(&options)

	if err = form.RunWithContext(ctx); err != nil {
		log.Fatal(err)
	}

	if !options.run {
		fmt.Println("Scan cancelled by user.")
		return
	}

	scanner := scanner.NewScanner(options.sourcePath)

	for _, outputType := range options.outputs {
		switch outputType {
		case "console":
			scanner.WithGenerator(generators.NewConsoleWriter())
		case "markdown":
			scanner.WithGenerator(generators.NewMarkdownGenerator(options.sourcePath, cfg.ArtifactsPath))
		}
	}

	if len(cfg.Keywords) > 0 {
		scanner.SetKeywords(cfg.Keywords...)
	}

	if err := scanner.Process(cancelCtx, 30*time.Second); err != nil {
		slog.Error(err.Error())
	}
}

// newForm creates a new interactive form using the huh library.
func newForm(options *Options) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter the root source path to your codebase").
				Prompt("📂 ").
				Placeholder("/home/morty/projects/myproject/src").
				Value(&options.sourcePath).
				Validate(func(value string) error {
					if len(value) == 0 {
						return fmt.Errorf("source path cannot be empty")
					}

					info, ok := os.Stat(value)
					if os.IsNotExist(ok) {
						return fmt.Errorf("source path does not exist")
					}

					if !info.IsDir() {
						return fmt.Errorf("source path is not a directory")
					}

					return nil
				}),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Output Formats").
				Options(
					huh.NewOption("Stdout", "console"),
					huh.NewOption("Markdown", "markdown").Selected(true),
				).
				Limit(50).
				Height(6).
				Value(&options.outputs).
				Validate(func(selection []string) error {
					if len(selection) == 0 {
						return fmt.Errorf("at least one output type is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Ready to start scanning?").
				Affirmative("Make it so!").
				Negative("No").
				Value(&options.run),
		),
	)

	return form
}
