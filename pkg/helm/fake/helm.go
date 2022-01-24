// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"bytes"
	"fmt"
	"strings"
)

// OCIPuller implements the ChartPuller interface
type OCIPuller struct {
	ExpectedName string
	Content      map[string]*bytes.Buffer
	Checksum     string
	Err          error
}

// PullOCIChart returns some fake content
func (f *OCIPuller) PullOCIChart(ociFullName string) (*bytes.Buffer, string, error) {
	tag := strings.Split(ociFullName, ":")[1]
	if f.ExpectedName != "" && f.ExpectedName != ociFullName {
		return nil, "", fmt.Errorf("expecting %s got %s", f.ExpectedName, ociFullName)
	}
	return f.Content[tag], f.Checksum, f.Err
}
