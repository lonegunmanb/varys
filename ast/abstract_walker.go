package ast

import (
	"github.com/ahmetb/go-linq"
	"github.com/lonegunmanb/johnnie"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"regexp"
	"strings"
)

type abstractWalker struct {
	johnnie.DefaultWalker
	osEnv         GoPathEnv
	pkgPath       string
	pkgName       string
	physicalPath  string
	actualWalker  johnnie.Walker
	analyzedTypes *types.Info
}

func (walker *abstractWalker) SetDir(dir string) {
	walker.physicalPath = dir
}

func (walker *abstractWalker) setAnalyzedTypes(i *types.Info) {
	walker.analyzedTypes = i
}

func (walker *abstractWalker) walkAsts(fileMap map[string][]*ast.File) error {
	for path, fileAsts := range fileMap {
		walker.SetDir(path)
		pkgPath, err := walker.osEnv.GetPkgPath(path)
		if err != nil {
			return err
		}
		for _, fileAst := range fileAsts {
			err := walkAst(walker, pkgPath, fileAst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func walkAst(walker *abstractWalker, pkgPath string, astFile *ast.File) error {
	walker.pkgPath = pkgPath
	johnnie.Visit(walker.actualWalker, astFile)
	return nil
}

func (*abstractWalker) analyzeTypes(pkgPath string, fileSet *token.FileSet,
	astFile *ast.File) (*types.Info, error) {
	analyzedTypes := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	_, err := (&types.Config{Importer: importer.For("source", nil)}).
		Check(pkgPath, fileSet, []*ast.File{astFile}, analyzedTypes)
	return analyzedTypes, err
}

func (*abstractWalker) getFiles(dirPath string, ignorePattern string) ([]FileInfo, error) {
	fileRetrieverKey := (*FileRetriever)(nil)
	fileRetriever := getOrRegister(fileRetrieverKey, func() interface{} {
		return NewFileRetriever()
	}).(FileRetriever)

	ignoreRegex, err := parseIgnorePattern(ignorePattern)
	if err != nil {
		return nil, err
	}
	files, err := fileRetriever.GetFiles(dirPath)
	if err != nil {
		return nil, err
	}
	filteredFiles := make([]FileInfo, 0)
	linq.From(files).Where(func(fileInfo interface{}) bool {
		info := fileInfo.(FileInfo)
		if ignoreRegex != nil {
			return isGoFile(info) && !ignoreRegex.MatchString(info.Name())
		}
		return isGoFile(info)
	}).ToSlice(&filteredFiles)
	return filteredFiles, nil
}

func (walker *abstractWalker) parseFileAsts(dirPath string, ignorePattern string, fSet *token.FileSet,
	osEnv GoPathEnv) (map[string][]*ast.File, error) {
	files, err := walker.getFiles(dirPath, ignorePattern)
	if err != nil {
		return nil, err
	}
	fileMap := make(map[string][]*ast.File)
	for _, file := range files {
		fileAst, err := parser.ParseFile(fSet, osEnv.ConcatFileNameWithPath(file.Dir(), file.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		fileMap[file.Dir()] = append(fileMap[file.Dir()], fileAst)
	}
	return fileMap, nil
}

func (walker *abstractWalker) parse(pkgPath string, fileName string, sourceCode string) error {
	fileset := token.NewFileSet()

	astFile, err := parser.ParseFile(fileset, fileName, sourceCode, 0)
	if err != nil {
		return err
	}
	fileAstMap := make(map[string][]*ast.File)
	fileAstMap[walker.physicalPath] = []*ast.File{astFile}
	if walker.analyzedTypes == nil {
		analyzedTypes, err := walker.analyzeTypes(pkgPath, fileset, astFile)
		if err != nil {
			return err
		}
		walker.setAnalyzedTypes(analyzedTypes)
	}

	return walker.walkAsts(fileAstMap)
}

func (walker *abstractWalker) Parse(pkgPath string, sourceCode string) error {
	return walker.parse(pkgPath, "src.go", sourceCode)
}

func (walker *abstractWalker) ParseDir(dirPath string, ignorePattern string) error {
	fSet := token.NewFileSet()
	osEnv := getOsEnv()
	fileAstMap, err := walker.parseFileAsts(dirPath, ignorePattern, fSet, osEnv)
	if err != nil {
		return err
	}
	info, err := parseTypes(fileAstMap, fSet, osEnv)
	if err != nil {
		return err
	}
	walker.setAnalyzedTypes(info)
	return walker.walkAsts(fileAstMap)
}

func isGoFile(info os.FileInfo) bool {
	return !info.IsDir() && isGoSrcFile(info.Name()) && !isTestFile(info.Name())
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(strings.TrimSuffix(fileName, ".go"), "test")
}

func isGoSrcFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".go")
}

func parseTypes(fileMap map[string][]*ast.File, fSet *token.FileSet, osEnv GoPathEnv) (*types.Info, error) {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	for path, fileAsts := range fileMap {
		var conf = &types.Config{Importer: importer.For("source", nil)}
		goPath, err := osEnv.GetPkgPath(path)
		if err != nil {
			return nil, err
		}
		_, err = conf.Check(goPath, fSet, fileAsts, info)
		if err != nil {
			return nil, err
		}
	}
	return info, nil
}

func getOsEnv() GoPathEnv {
	return getOrRegister((*GoPathEnv)(nil), func() interface{} {
		return NewGoPathEnv()
	}).(GoPathEnv)
}

func parseIgnorePattern(ignorePattern string) (*regexp.Regexp, error) {
	var regex *regexp.Regexp
	if ignorePattern != "" {
		reg, err := regexp.Compile(ignorePattern)
		if err != nil {
			return nil, err
		}
		regex = reg
	}
	return regex, nil
}