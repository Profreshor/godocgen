package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Profreshor/godocgen/internal/report"
	"github.com/Profreshor/godocgen/internal/scanner"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("godocgen: Please enter <project-directory>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		os.Exit(1)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("godocgen: %s: No such directory\n", inputPath)
			os.Exit(1)
		}
		fmt.Printf("godocgen: %s\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Println("godocgen: Please enter a directory path")
		os.Exit(1)
	}

	project, err := scanner.Scan(absPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	report.PrintTerminalReport(project)
}
