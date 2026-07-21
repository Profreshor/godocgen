package walker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Profreshor/godocgen/internal/model"
)

var ignoreList = map[string]bool{
	".git":         true,
	"vendor":       true,
	"venv":         true,
	"node_modules": true,
	"bin":          true,
}

var captureList = map[string]bool{
	".go": true,
	// ".py":   true,
	// ".md":   true,
	// ".sh":   true,
	// ".yml":  true,
	// ".yaml": true,
}

func WalkFiles(rootPath string) (model.Project, error) {
	project := model.Project{RootPath: rootPath}
	fileSystem := os.DirFS(project.RootPath)
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		ext := filepath.Ext(name)

		if d.IsDir() && ignoreList[name] {
			return fs.SkipDir
		}
		if !d.IsDir() && captureList[ext] {
			fmt.Println(name)
			absPath := filepath.Join(
				project.RootPath,
				path,
			)
			content, ReadErr := os.ReadFile(absPath)
			project.Files = append(project.Files, model.SourceFile{
				AbsolutePath: absPath,
				RelativePath: path,
				FileExt:      ext,
				Content:      content,
				LoadErr:      ReadErr,
			})
		}
		return nil
	})
	if err != nil {
		return model.Project{}, err
	}
	return project, nil
}
