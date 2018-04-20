package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	jsonnet "github.com/google/go-jsonnet"
	log "github.com/sirupsen/logrus"
)

var errNotFound = errors.New("Not found")

//go:generate go-bindata -nometadata -ignore .*_test\.|~$DOLLAR -pkg $GOPACKAGE -o bindata.go -prefix ../ ../lib/...
func newInternalFS(prefix string) http.FileSystem {
	// Asset/AssetDir returns `fmt.Errorf("Asset %s not found")`,
	// which does _not_ get mapped to 404 by `http.FileSystem`.
	// Need to convert to `os.ErrNotExist` explicitly ourselves.
	mapNotFound := func(err error) error {
		if err != nil && strings.Contains(err.Error(), "not found") {
			err = os.ErrNotExist
		}
		return err
	}
	return &assetfs.AssetFS{
		Asset: func(path string) ([]byte, error) {
			ret, err := Asset(path)
			return ret, mapNotFound(err)
		},
		AssetDir: func(path string) ([]string, error) {
			ret, err := AssetDir(path)
			return ret, mapNotFound(err)
		},
		Prefix: prefix,
	}
}

/*
MakeUniversalImporter creates an importer that handles resolving imports from the filesystem and http/s.

In addition to the standard importer, supports:
  - URLs in import statements
  - URLs in library search paths

A real-world example:
  - you have https://raw.githubusercontent.com/ksonnet/ksonnet-lib/master in your search URLs
  - you evaluate a local file which calls `import "ksonnet.beta.2/k.libsonnet"`
  - if the `ksonnet.beta.2/k.libsonnet`` is not located in the current workdir, an attempt
    will be made to follow the search path, i.e. to download
    https://raw.githubusercontent.com/ksonnet/ksonnet-lib/master/ksonnet.beta.2/k.libsonnet
  - since the downloaded `k.libsonnet`` file turn in contains `import "k8s.libsonnet"`, the import
    will be resolved as https://raw.githubusercontent.com/ksonnet/ksonnet-lib/master/ksonnet.beta.2/k8s.libsonnet
	and downloaded from that location
*/
func MakeUniversalImporter(searchUrls []*url.URL) jsonnet.Importer {
	// Reconstructed copy of http.DefaultTransport (to avoid
	// modifying the default)
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	t.RegisterProtocol("internal", http.NewFileTransport(newInternalFS("lib")))

	return &universalImporter{
		BaseSearchURLs: searchUrls,
		HTTPClient:     &http.Client{Transport: t},
	}
}

type universalImporter struct {
	BaseSearchURLs []*url.URL
	HTTPClient     *http.Client
}

func (importer *universalImporter) Import(dir, importedPath string) (*jsonnet.ImportedData, error) {
	log.Debugf("Importing %q from %q", importedPath, dir)

	candidateURLs, err := importer.expandImportToCandidateURLs(dir, importedPath)
	if err != nil {
		return nil, fmt.Errorf("Could not get candidate URLs for when importing %s (import dir is %s)", importedPath, dir)
	}

	var tried []string
	for _, u := range candidateURLs {
		tried = append(tried, u.String())
		importedData, err := importer.tryImport(u)
		if err == nil {
			return importedData, nil
		} else if err != errNotFound {
			return nil, err
		}
	}

	return nil, fmt.Errorf("Couldn't open import %q, no match locally or in library search paths. Tried: %s",
		importedPath,
		strings.Join(tried, ";"),
	)
}

func (importer *universalImporter) tryImport(url *url.URL) (*jsonnet.ImportedData, error) {
	res, err := importer.HTTPClient.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	log.Debugf("GET %s -> %s", url, res.Status)
	if res.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading content: %s", res.Status)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return &jsonnet.ImportedData{
		FoundHere: url.String(),
		Content:   string(bodyBytes),
	}, nil
}

func (importer *universalImporter) expandImportToCandidateURLs(dir, importedPath string) ([]*url.URL, error) {
	importedPathURL, err := url.Parse(importedPath)
	if err != nil {
		return nil, fmt.Errorf("Import path %q is not valid", importedPath)
	}
	if importedPathURL.IsAbs() {
		return []*url.URL{importedPathURL}, nil
	}

	importDirURL, err := url.Parse(dir)
	if err != nil {
		return nil, fmt.Errorf("Invalid import dir %q", dir)
	}

	candidateURLs := make([]*url.URL, 0, len(importer.BaseSearchURLs)+1)

	candidateURLs = append(candidateURLs, importDirURL.ResolveReference(importedPathURL))

	for _, u := range importer.BaseSearchURLs {
		candidateURLs = append(candidateURLs, u.ResolveReference(importedPathURL))
	}

	return candidateURLs, nil
}
