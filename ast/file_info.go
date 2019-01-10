package ast

import "os"

type FileInfo interface {
	os.FileInfo
	Dir() string
}

type fileInfo struct {
	os.FileInfo
	fileDir string
}

func (f *fileInfo) Dir() string {
	return f.fileDir
}
