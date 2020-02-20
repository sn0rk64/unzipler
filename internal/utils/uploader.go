package utils

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const ARCH_DEST = "./assets/archives"

func UploadArchive(r *http.Request) (fpath string, fname string, err error) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("archive")
	if err != nil {
		return
	}
	defer file.Close()

	fname = RandString(10) + filepath.Ext(handler.Filename)
	fpath = filepath.Join(ARCH_DEST, fname)
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer outFile.Close()

	io.Copy(outFile, file)

	return
}
