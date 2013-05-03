package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var exitCode = 0
var (
	target  = flag.String("t", "C:\\mygosrc", "Source is saved in the directory")
	pkgPath = flag.String("p", "", "Execution of the program package name")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gosrc [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
func main() {
	flag.Usage = usage
	flag.Parse()
	if *target == "" {
		fmt.Println("target Can not be empty")
		exitCode = 2
		return
	}
	if *pkgPath == "" {
		fmt.Println("target Can not be empty")
		exitCode = 2
		return
	}
	out, err := exec.Command("go", "list", "-json", fmt.Sprint(*pkgPath)).Output()
	if err != nil {
		log.Println(err)
		return
	}
	var pk Package
	err = json.Unmarshal(out, &pk)
	if err != nil {
		log.Println(err)
		return
	}
	var Deps []string
	var SysDeps []string
	var OutDeps []string
	for _, v := range pk.Deps {
		Deps = append(Deps, v)
	}
	dir, file := filepath.Split(fmt.Sprint(runtime.GOROOT(), "src\\pkg\\"))
	fs := http.Dir(dir)
	f, err := fs.Open(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil {
		log.Println(err)
		return
	}
	//获取系统库src\pkg下的文件夹列表
	if d.IsDir() {
		for {
			dirs, err := f.Readdir(100)
			if err != nil || len(dirs) == 0 {
				break
			}
			for _, dd := range dirs {
				SysDeps = append(SysDeps, dd.Name())
			}
		}
	}
	//获取外部引用的包列表
	for _, v := range Deps {
		isSys := false
		for _, vv := range SysDeps {
			if strings.Index(v, vv) == 0 {
				isSys = true
				break
			}
		}
		if !isSys {
			OutDeps = append(OutDeps, v)
		}
	}
	//加入自身的源码包
	OutDeps = append(OutDeps, pk.ImportPath)
	base := fmt.Sprint(*target, "\\src")
	//删除指定的目录
	os.RemoveAll(base)
	//建立指定的目录
	ttbase := ""
	ts := strings.Split(base, "\\")

	for i := 0; i < len(ts); i++ {
		if i == 0 {
			ttbase = fmt.Sprint(ts[i])
		} else {
			ttbase = fmt.Sprint(ttbase, "\\", ts[i])
			os.Mkdir(ttbase, 0600)
		}
	}
	//复制引用的源码到指定目录
	srcbase := fmt.Sprint(pk.Root, "\\src")
	fmt.Println(srcbase)
	for _, v := range OutDeps {
		s := strings.Split(v, "/")
		tbase := fmt.Sprint(base)
		for i := 0; i < len(s); i++ {
			tbase = fmt.Sprint(tbase, "\\", s[i])
			os.Mkdir(tbase, 0600)
		}
		tsrcbase := strings.Replace(tbase, base, srcbase, -1)
		cmd := exec.Command("xcopy", fmt.Sprint(tsrcbase, "\\*.go"), tbase, "/f", "/e")
		cmd.Run()
		fmt.Println(tbase)
		fmt.Println(tsrcbase)
	}

}

type Package struct {
	Dir        string // directory containing package sources
	ImportPath string // import path of package in dir
	Name       string // package name
	Doc        string // package documentation string
	Target     string // install path
	Goroot     bool   // is this package in the Go root?
	Standard   bool   // is this package part of the standard Go library?
	Stale      bool   // would 'go install' do anything for this package?
	Root       string // Go root or Go path dir containing this package

	// Source files
	GoFiles   []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles  []string // .go sources files that import "C"
	CFiles    []string // .c source files
	HFiles    []string // .h source files
	SFiles    []string // .s source files
	SysoFiles []string // .syso object files to add to archive

	// Cgo directives
	CgoCFLAGS    []string // cgo: flags for C compiler
	CgoLDFLAGS   []string // cgo: flags for linker
	CgoPkgConfig []string // cgo: pkg-config names

	// Dependency information
	Imports []string // import paths used by this package
	Deps    []string // all (recursively) imported dependencies

	// Error information
	Incomplete bool            // this package or a dependency has an error
	Error      *PackageError   // error loading package
	DepsErrors []*PackageError // errors loading dependencies

	TestGoFiles  []string // _test.go files in package
	TestImports  []string // imports from TestGoFiles
	XTestGoFiles []string // _test.go files outside package
	XTestImports []string // imports from XTestGoFiles
}
type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error
	Err         string   // the error itself
}
