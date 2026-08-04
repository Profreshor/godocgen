package walker

import (
	"io/fs"
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

func WalkFiles(fileSystem fs.FS) (model.Project, error) {
	project := model.Project{}

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
			content, readErr := fs.ReadFile(fileSystem, path)
			project.Files = append(project.Files, model.SourceFile{
				RelativePath: path,
				FileExt:      ext,
				Content:      content,
				LoadErr:      readErr,
			})
		}

		return nil
	})
	if err != nil {
		return model.Project{}, err
	}
	return project, nil
}
