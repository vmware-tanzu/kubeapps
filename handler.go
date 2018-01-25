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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubeapps/chartsvc/models"
	"github.com/kubeapps/common/response"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// Params a key-value map of path params
type Params map[string]string

// WithParams can be used to wrap handlers to take an extra arg for path params
type WithParams func(http.ResponseWriter, *http.Request, Params)

func (h WithParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(w, req, vars)
}

const chartCollection = "charts"
const readmeCollection = "readmes"

type apiResponse struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Attributes    interface{} `json:"attributes"`
	Links         interface{} `json:"links"`
	Relationships relMap      `json:"relationships"`
}

type apiListResponse []*apiResponse

type selfLink struct {
	Self string `json:"self"`
}

type relMap map[string]rel
type rel struct {
	Data  interface{} `json:"data"`
	Links selfLink    `json:"links"`
}

// listCharts returns a list of charts
func listCharts(w http.ResponseWriter, req *http.Request) {
	db, closer := dbSession.DB()
	defer closer()
	var charts []*models.Chart
	if err := db.C(chartCollection).Find(nil).Sort("name").All(&charts); err != nil {
		log.WithError(err).Error("could not fetch charts")
		response.NewErrorResponse(http.StatusInternalServerError, "could not fetch all charts").Write(w)
		return
	}

	cl := newChartListResponse(charts)
	response.NewDataResponse(cl).Write(w)
}

// listRepoCharts returns a list of charts in the given repo
func listRepoCharts(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var charts []*models.Chart
	if err := db.C(chartCollection).Find(bson.M{"repo.name": params["repo"]}).Sort("_id").All(&charts); err != nil {
		log.WithError(err).Error("could not fetch charts")
		response.NewErrorResponse(http.StatusInternalServerError, "could not fetch all charts").Write(w)
		return
	}

	cl := newChartListResponse(charts)
	response.NewDataResponse(cl).Write(w)
}

// getChart returns the chart from the given repo
func getChart(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var chart *models.Chart
	chartID := fmt.Sprintf("%s/%s", params["repo"], params["chartName"])
	if err := db.C(chartCollection).FindId(chartID).One(&chart); err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart").Write(w)
		return
	}

	cr := newChartResponse(chart)
	response.NewDataResponse(cr).Write(w)
}

// listChartVersions returns a list of chart versions for the given chart
func listChartVersions(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var chart *models.Chart
	chartID := fmt.Sprintf("%s/%s", params["repo"], params["chartName"])
	if err := db.C(chartCollection).FindId(chartID).One(&chart); err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart").Write(w)
		return
	}

	cvl := newChartVersionListResponse(chart)
	response.NewDataResponse(cvl).Write(w)
}

// getChartVersion returns the given chart version
func getChartVersion(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var chart *models.Chart
	chartID := fmt.Sprintf("%s/%s", params["repo"], params["chartName"])
	if err := db.C(chartCollection).Find(bson.M{
		"_id":           chartID,
		"chartversions": bson.M{"$elemMatch": bson.M{"version": params["version"]}},
	}).Select(bson.M{
		"name": 1, "repo": 1, "description": 1, "home": 1, "keywords": 1, "maintainers": 1, "sources": 1,
		"chartversions.$": 1,
	}).One(&chart); err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart version").Write(w)
		return
	}

	cvr := newChartVersionResponse(chart, chart.ChartVersions[0])
	response.NewDataResponse(cvr).Write(w)
}

// getChartIcon returns the icon for a given chart
func getChartIcon(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var chart *models.Chart
	chartID := fmt.Sprintf("%s/%s", params["repo"], params["chartName"])
	if err := db.C(chartCollection).FindId(chartID).One(&chart); err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		http.NotFound(w, req)
		return
	}

	if chart.RawIcon == nil {
		http.NotFound(w, req)
		return
	}

	w.Write(chart.RawIcon)
}

// getChartVersionReadme returns the README for a given chart
func getChartVersionReadme(w http.ResponseWriter, req *http.Request, params Params) {
	db, closer := dbSession.DB()
	defer closer()
	var readme *models.ChartReadme
	readmeID := fmt.Sprintf("%s/%s-%s", params["repo"], params["chartName"], params["version"])
	if err := db.C(readmeCollection).FindId(readmeID).One(&readme); err != nil {
		log.WithError(err).Errorf("could not find readme with id %s", readmeID)
		http.NotFound(w, req)
		return
	}

	w.Write([]byte(readme.Readme))
}

func newChartResponse(c *models.Chart) *apiResponse {
	latestCV := c.ChartVersions[0]
	return &apiResponse{
		Type:       "chart",
		ID:         c.ID,
		Attributes: chartAttributes(*c),
		Links:      selfLink{"/v1/charts/" + c.ID},
		Relationships: relMap{
			"latestChartVersion": rel{
				Data:  chartVersionAttributes(c.ID, latestCV),
				Links: selfLink{"/v1/charts/" + c.ID + "/versions/" + latestCV.Version},
			},
		},
	}
}

func newChartListResponse(charts []*models.Chart) apiListResponse {
	var cl apiListResponse
	for _, c := range charts {
		cl = append(cl, newChartResponse(c))
	}

	return cl
}

func chartVersionAttributes(cid string, cv models.ChartVersion) models.ChartVersion {
	cv.Readme = "/v1/assets/" + cid + "/versions/" + cv.Version + "/README.md"
	return cv
}

func chartAttributes(c models.Chart) models.Chart {
	if c.RawIcon != nil {
		c.Icon = "/v1/assets/" + c.ID + "/logo-160x160-fit.png"
	} else {
		// If the icon wasn't processed, it is either not set or invalid
		c.Icon = ""
	}
	return c
}

func newChartVersionResponse(c *models.Chart, cv models.ChartVersion) *apiResponse {
	return &apiResponse{
		Type:       "chartVersion",
		ID:         fmt.Sprintf("%s-%s", c.ID, cv.Version),
		Attributes: chartVersionAttributes(c.ID, cv),
		Links:      selfLink{"/v1/charts/" + c.ID + "/versions/" + cv.Version},
		Relationships: relMap{
			"chart": rel{
				Data:  chartAttributes(*c),
				Links: selfLink{"/v1/charts/" + c.ID},
			},
		},
	}
}

func newChartVersionListResponse(c *models.Chart) apiListResponse {
	var cvl apiListResponse
	for _, cv := range c.ChartVersions {
		cvl = append(cvl, newChartVersionResponse(c, cv))
	}

	return cvl
}
