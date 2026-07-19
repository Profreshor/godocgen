package model

type SourceFile struct {
	AbsolutePath string
	RelativePath string
	FileExt      string
	Content      []byte
	LoadErr      error
}
