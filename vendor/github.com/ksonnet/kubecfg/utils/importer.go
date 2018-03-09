package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
	log "github.com/sirupsen/logrus"
)

var errNotFound = errors.New("Not found")

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
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

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
