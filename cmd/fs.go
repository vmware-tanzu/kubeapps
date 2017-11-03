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
	manifest_f, err := statikFS.Open(fname)
	if err != nil {
		log.Fatalf("ERROR: Static file '%s' not found", fname)
		return "", err
	}
	io.Copy(buf, manifest_f)
	manifest_s := string(buf.Bytes())
	return manifest_s, nil
}
