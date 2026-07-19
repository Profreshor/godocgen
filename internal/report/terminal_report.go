package report

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/Profreshor/godocgen/internal/model"
)

const logo string = `
        ▌               
▞▀▌▞▀▖▞▀▌▞▀▖▞▀▖▞▀▌▞▀▖▛▀▖
▚▄▌▌ ▌▌ ▌▌ ▌▌ ▖▚▄▌▛▀ ▌ ▌
▗▄▘▝▀ ▝▀▘▝▀ ▝▀ ▗▄▘▝▀▘▘ ▘`

// group project.Files by filepath.Dir(RelativePath)
// sort directory headings
// sort files inside each heading
// print filename using filepath.Base

func PrintTerminalReport(project model.Project) {
	fileGroup := make(map[string][]string, len(project.Files))
	for _, file := range project.Files {
		directory := filepath.Dir(file.RelativePath)
		fileName := filepath.Base(file.RelativePath)
		fileGroup[directory] = append(fileGroup[directory], fileName)
	}
	directories := make([]string, 0, len(fileGroup))
	for k := range fileGroup {
		directories = append(directories, k)
	}
	slices.Sort(directories)

	fmt.Println(logo)
	fmt.Printf("\nSelected Project Root: %s\n\n", project.RootPath)
	fmt.Printf("Project contains: %d Go files\n  Across %d directories\n",
		len(project.Files), len(directories))

	for _, dir := range directories {
		slices.Sort(fileGroup[dir])
		fmt.Printf("\nDirectory: %s\n", dir)
		fmt.Printf("Files -> %d\n", len(fileGroup[dir]))
		for _, file := range fileGroup[dir] {
			fmt.Printf("  %s\n", file)
		}
	}

}
