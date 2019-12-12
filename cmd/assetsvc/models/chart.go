/*
Copyright (c) 2017 The Helm Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"time"

	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Repo holds the App repository details
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Chart is a higher-level representation of a chart package
type Chart struct {
	ID              string             `json:"-" bson:"_id"`
	Name            string             `json:"name"`
	Repo            Repo               `json:"repo"`
	Description     string             `json:"description"`
	Home            string             `json:"home"`
	Keywords        []string           `json:"keywords"`
	Maintainers     []chart.Maintainer `json:"maintainers"`
	Sources         []string           `json:"sources"`
	Icon            string             `json:"icon"`
	RawIcon         []byte             `json:"-" bson:"raw_icon"`
	IconContentType string             `json:"-" bson:"icon_content_type,omitempty"`
	ChartVersions   []ChartVersion     `json:"-"`
}

// ChartVersion is a representation of a specific version of a chart
type ChartVersion struct {
	Version    string    `json:"version"`
	AppVersion string    `json:"app_version"`
	Created    time.Time `json:"created"`
	Digest     string    `json:"digest"`
	URLs       []string  `json:"urls"`
	Readme     string    `json:"readme" bson:"-"`
	Values     string    `json:"values" bson:"-"`
	Schema     string    `json:"schema" bson:"-"`
}

// ChartFiles holds the README and values for a given chart version
type ChartFiles struct {
	ID     string `bson:"_id"`
	Readme string
	Values string
	Schema string
}
