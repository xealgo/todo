package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/xealgo/todo/internal/codescanner"
)

func main() {
	ctx := context.Background()
	timeout, _ := context.WithTimeout(ctx, 30*time.Second)

	scanner := codescanner.NewScanner("/root/path/to/your/codebase")
	scanner.WithWriter(codescanner.NewSimpleConsoleWriter())
	// scanner.SetKeywords("hello", "WORLD")

	// files, err := scanner.Enumerate(timeout)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	os.Exit(1)
	// }

	// fmt.Println(files)

	err := scanner.Process(timeout)
	if err != nil {
		slog.Error(err.Error())
	}
}
