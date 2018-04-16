/*
Copyright (c) 2017 Bitnami

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

package main

import (
	"time"
)

type repo struct {
	Name        string
	URL         string
	AccessToken string `bson:"-"`
}

type maintainer struct {
	Name  string
	Email string
}

type chart struct {
	ID            string `bson:"_id"`
	Name          string
	Repo          repo
	Description   string
	Home          string
	Keywords      []string
	Maintainers   []maintainer
	Sources       []string
	Icon          string
	ChartVersions []chartVersion
}

type chartVersion struct {
	Version    string
	AppVersion string
	Created    time.Time
	Digest     string
	URLs       []string
}

type chartFiles struct {
	ID     string `bson:"_id"`
	Readme string
	Values string
	Repo   repo
}
