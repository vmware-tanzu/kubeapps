package cmd

import (
	"bytes"
	_ "github.com/kubeapps/installer/generated/statik"
	"github.com/rakyll/statik/fs"
	"io"
	"log"
	"net/http"
)

var statikFS http.FileSystem

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatalf("ERROR initializing statikFS")
	}
}

func fsGetFile(fname string) (string, error) {
	buf := bytes.NewBuffer(nil)
	statik_file, err := statikFS.Open(fname)
	if err != nil {
		log.Fatalf("ERROR: Static file '%s' not found", fname)
		return "", err
	}
	io.Copy(buf, statik_file)
	content := string(buf.Bytes())
	return content, nil
}
