// Copyright 2017-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"
	"time"

	"database/sql/driver"

	"helm.sh/helm/v3/pkg/chart"
)

// Repo holds the App repository basic details
type Repo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

// RepoInternal holds the App repository details including auth
type RepoInternal struct {
	Namespace           string `json:"namespace"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Type                string `json:"type"`
	AuthorizationHeader string `bson:"-"`
}

// Chart is a higher-level representation of a chart package
type Chart struct {
	ID              string             `json:"ID" bson:"chart_id"`
	Name            string             `json:"name"`
	Repo            *Repo              `json:"repo"`
	Description     string             `json:"description"`
	Home            string             `json:"home"`
	Keywords        []string           `json:"keywords"`
	Maintainers     []chart.Maintainer `json:"maintainers"`
	Sources         []string           `json:"sources"`
	Icon            string             `json:"icon"`
	RawIcon         []byte             `json:"raw_icon" bson:"raw_icon"`
	IconContentType string             `json:"icon_content_type" bson:"icon_content_type,omitempty"`
	Category        string             `json:"category"`
	ChartVersions   []ChartVersion     `json:"chartVersions"`
}

// ChartCategory is the representation of the chart category query
// Note "name" and "count" names are not columns but select aliases
type ChartCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ChartIconString is a higher-level representation of a chart package
// TODO(andresmgot) Replace this type when the icon is stored as a binary
type ChartIconString struct {
	Chart
	RawIcon string `json:"raw_icon" bson:"raw_icon"`
}

// ChartVersion is a representation of a specific version of a chart
type ChartVersion struct {
	Version    string    `json:"version"`
	AppVersion string    `json:"app_version"`
	Created    time.Time `json:"created"`
	Digest     string    `json:"digest"`
	URLs       []string  `json:"urls"`
	// The following three fields get set with the URL paths to the respective
	// chart files (as opposed to the similar fields on ChartFiles which
	// contain the actual content).
	Readme string `json:"readme" bson:"-"`
	Values string `json:"values" bson:"-"`
	Schema string `json:"schema" bson:"-"`
}

// ChartFiles holds the README and values for a given chart version
type ChartFiles struct {
	ID     string `bson:"file_id"`
	Readme string
	Values string
	Schema string
	Repo   *Repo
	Digest string
}

// Allow to convert ChartFiles to a sql JSON
func (a ChartFiles) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// some constant strings used as keys in maps in several modules
const (
	ReadmeKey    = "readme"
	ValuesKey    = "values"
	SchemaKey    = "schema"
	ChartYamlKey = "chartYaml"
)
