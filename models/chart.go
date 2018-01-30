package models

import "time"

// Repo holds the App repository details
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Maintainer holds maintainer details of a chart
type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Chart is a higher-level representation of a chart package
type Chart struct {
	ID            string         `json:"-" bson:"_id"`
	Name          string         `json:"name"`
	Repo          Repo           `json:"repo"`
	Description   string         `json:"description"`
	Home          string         `json:"home"`
	Keywords      []string       `json:"keywords"`
	Maintainers   []Maintainer   `json:"maintainers"`
	Sources       []string       `json:"sources"`
	Icon          string         `json:"icon"`
	RawIcon       []byte         `json:"-" bson:"raw_icon"`
	ChartVersions []ChartVersion `json:"-"`
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
}

// ChartReadme holds the README for a given chart version
type ChartReadme struct {
	ID     string `bson:"_id"`
	Readme string
}

// ChartValues holds the values.yaml for a given chart version
type ChartValues struct {
	ID     string `bson:"_id"`
	Values string
}
