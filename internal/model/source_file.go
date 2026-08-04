package model

type SourceFile struct {
	RelativePath string
	FileExt      string
	Content      []byte
	LoadErr      error
}
