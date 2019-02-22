package ast

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq"
	"os"
	"runtime"
	"strings"
)

type GoPathEnv interface {
	IsWindows() bool
	GetGoPath() string
	ConcatFileNameWithPath(path string, fileName string) string
	GetPkgPath(systemPath string) (string, error)
}

type envImpl struct {
}

func NewGoPathEnv() GoPathEnv {
	return &envImpl{}
}

func (*envImpl) IsWindows() bool {
	return runtime.GOOS == "windows"
}

func (*envImpl) GetGoPath() string {
	return os.Getenv("GOPATH")
}

func (e *envImpl) ConcatFileNameWithPath(path string, fileName string) string {
	return concatFileNameWithPath(e.IsWindows(), path, fileName)
}

func concatFileNameWithPath(isWindows bool, path string, fileName string) string {
	if isWindows {
		return fmt.Sprintf("%s\\%s", path, fileName)
	}
	return fmt.Sprintf("%s/%s", path, fileName)
}

func (e *envImpl) GetPkgPath(systemPath string) (string, error) {
	return getPkgPath(e, systemPath)
}

func getPkgPath(env GoPathEnv, systemPath string) (string, error) {
	isWindows := env.IsWindows()
	goPaths, err := getGoPaths(env)
	if err != nil {
		return "", err
	}
	return getPkgPathFromSystemPathUsingGoPath(isWindows, goPaths, systemPath)
}

func getGoPaths(env GoPathEnv) (gopaths []string, err error) {
	sep := ":"
	if env.IsWindows() {
		sep = ";"
	}
	goPath := env.GetGoPath()
	if goPath == "" {
		err = errors.New("no go path detected")
		return
	}
	gopaths = strings.Split(goPath, sep)
	return
}

func getPkgPathFromSystemPathUsingGoPath(isWindows bool, goPaths []string, systemPath string) (pkgPath string, err error) {
	goSrcPath := linq.From(goPaths).Select(func(path interface{}) interface{} {
		srcTemplate := "%s/src"
		if isWindows {
			srcTemplate = "%s\\src"
		}
		return fmt.Sprintf(srcTemplate, path)
	}).FirstWith(func(path interface{}) bool {
		stringPath := path.(string)
		return isInGoPath(stringPath, systemPath)
	})
	if goSrcPath == nil {
		err = errors.New(fmt.Sprintf("%s is not in go src path", systemPath))
		return
	} else {
		goPath := goSrcPath.(string)
		if isGoPathRoot(goPath, systemPath) {
			pkgPath = ""
			return
		}
		pkgPath = strings.TrimPrefix(systemPath, goPath)
		pkgPath = strings.TrimPrefix(pkgPath, "/")
		pkgPath = strings.TrimPrefix(pkgPath, "\\")
		if isWindows {
			pkgPath = strings.Replace(pkgPath, "\\", "/", -1)
		}
	}
	return
}

func isInGoPath(stringPath string, systemPath string) bool {
	return isGoPathRoot(stringPath, systemPath) || strings.HasPrefix(systemPath, stringPath)
}

func isGoPathRoot(stringPath string, systemPath string) bool {
	return stringPath == (systemPath)
}
