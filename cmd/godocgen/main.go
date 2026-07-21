package main

import (
	"fmt"
	"os"
	"path/filepath"

	// "github.com/Profreshor/godocgen/internal/report"
	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/walker"
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

	project, err := walker.WalkFiles(absPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, file := range project.Files {
		if file.LoadErr != nil {
			fmt.Printf("Skipping %s: due to load error: %v\n", file.RelativePath, file.LoadErr)
		}
		lex, err := lexer.CreateLexer(file.Content, file.FileExt)
		if err != nil {
			fmt.Printf("codedocgen: %v\n", err)
		}
		lex.Tokenize()
	}
	// report.PrintTerminalReport(project)
}
