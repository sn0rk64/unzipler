package arch

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DEST = "./assets/files"

var garbage = []string{
	"__MACOSX",
	".DS_Store",
}

type Arch struct {
	Src  string
	Dest string
}

type resource map[string]map[string]FileInfo

type FileInfo struct {
	Path string `json:"path"`
	Ext  string `json:"ext"`
}

type FileTree struct {
	Name     string               `json:"name"`
	Files    map[string]FileInfo  `json:"files"`
	Children map[string]*FileTree `json:"children"`
}

func New(src string, fname string) *Arch {
	return &Arch{
		Src:  src,
		Dest: filepath.Join(DEST, fname),
	}
}

func (a *Arch) Unzip() (*FileTree, error) {
	rs := make(resource)

	r, err := zip.OpenReader(a.Src)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(a.Dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(a.Dest)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("%s: illegal file path", fpath)
		}

		if hasGarbage(fpath) {
			continue
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return nil, err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return nil, err
		}

		fillResource(&rs, fpath, a.Dest)

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return nil, err
		}

		fd, err := f.Open()
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(outFile, fd)

		outFile.Close()
		fd.Close()

		if err != nil {
			return nil, err
		}
	}

	tree := assembleFileTree(rs)

	return tree, nil
}

func newFileTree(name string) *FileTree {
	return &FileTree{
		Name:     name,
		Files:    make(map[string]FileInfo),
		Children: make(map[string]*FileTree),
	}
}

func fillResource(rs *resource, fpath string, dest string) {
	dir, fname := filepath.Split(fpath)
	dir = strings.Replace(dir, dest, "", -1)
	dir = strings.TrimPrefix(dir, "/")
	dir = strings.TrimSuffix(dir, "/")

	// Empty dir means that there is no dir and files are in the root directory.
	if dir == "" {
		// So dir will be mean as root.
		dir = "."
	}

	if _, ok := (*rs)[dir]; !ok {
		(*rs)[dir] = make(map[string]FileInfo)
	}

	(*rs)[dir][fname] = FileInfo{
		Path: fpath,
		Ext:  getFileExt(fname),
	}
}

func assembleFileTree(rs resource) *FileTree {
	tree := newFileTree(".")
	tree.Files = rs["."]
	delete(rs, ".")

	for key, val := range rs {
		// If we have only one dir.
		if !strings.Contains(key, "/") {
			// Check tree has initialized children.
			if _, ok := tree.Children[key]; !ok {
				// If it's not, initialize children.
				tree.Children[key] = newFileTree(key)
			}

			tree.Children[key].Files = val

			continue
		}

		// If we have many dirs split them and handle one by one.
		dirs := strings.Split(key, "/")
		var tmp *FileTree
		for idx, dir := range dirs {
			// Check for tmp is empty.
			if tmp == nil {
				// Check tree has initialized children.
				if _, ok := tree.Children[dir]; !ok {
					// If it's not, initialize children.
					tree.Children[dir] = newFileTree(dir)
				}

				// This children will be used as parent for its children.
				tmp = tree.Children[dir]
				continue
			}

			// Check parent FileTree has children FileTree.
			if _, ok := tmp.Children[dir]; !ok {
				// If it's not, initialize children FileTree.
				tmp.Children[dir] = newFileTree(dir)
			}

			// If last iteration assign val (that contains files) to FileTree.Files
			if idx == len(dirs)-1 {
				tmp.Children[dir].Files = val
			}

			// This children will be used as parent for its children.
			tmp = tmp.Children[dir]
		}
	}

	return tree
}

func hasGarbage(s string) bool {
	for _, g := range garbage {
		if strings.Contains(s, g) {
			return true
		}
	}

	return false
}

func getFileExt(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	if string(ext[0]) == "." {
		ext = ext[1:]
	}
	return ext
}
