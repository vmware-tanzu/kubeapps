/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/response"
	log "github.com/sirupsen/logrus"
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
const filesCollection = "files"

type apiResponse struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Attributes    interface{} `json:"attributes"`
	Links         interface{} `json:"links"`
	Relationships relMap      `json:"relationships"`
}

type apiListResponse []*apiResponse

type apiChartCategoryListResponse []*models.ChartCategory

type selfLink struct {
	Self string `json:"self"`
}

type relMap map[string]rel
type rel struct {
	Data  interface{} `json:"data"`
	Links selfLink    `json:"links"`
}

type meta struct {
	TotalPages int `json:"totalPages"`
}

// count is used to parse the result of a $count operation in the database
type count struct {
	Count int
}

// getPageAndSizeParams extracts the page number and the page size of a request. Default (page,size) = (1, 0) if not set
func getPageAndSizeParams(req *http.Request) (int, int) {
	pageNumber := req.FormValue("page")
	pageSize := req.FormValue("size")

	// if page is a non-positive int or 0, defaults to 1
	pageNumberInt, err := strconv.ParseUint(pageNumber, 10, 64)
	if err != nil || pageNumberInt == 0 {
		pageNumberInt = 1
	}
	// ParseUint will return 0 if size is a not positive integer
	pageSizeInt, _ := strconv.ParseUint(pageSize, 10, 64)

	return int(pageNumberInt), int(pageSizeInt)
}

func extractDecodedNamespaceAndRepoAndVersionParams(params Params) (string, string, string, string, error) {
	namespace, err := url.PathUnescape(params["namespace"])
	if err != nil {
		return "", "", "", params["namespace"], err
	}

	repo, err := url.PathUnescape(params["repo"])
	if err != nil {
		return "", "", "", params["repo"], err
	}

	version, err := url.PathUnescape(params["version"])
	if err != nil {
		return "", "", "", params["version"], err
	}

	return namespace, repo, version, "", nil
}

func extractChartQueryFromRequest(namespace, repo string, req *http.Request) utils.ChartQuery {
	repos := []string{}
	if repo != "" {
		repos = append(repos, repo)
	}

	if req.FormValue("repos") != "" {
		repos = append(repos, strings.Split(strings.TrimSpace(req.FormValue("repos")), ",")...)
	}
	categories := []string{}
	if req.FormValue("categories") != "" {
		categories = strings.Split(strings.TrimSpace(req.FormValue("categories")), ",")
	}

	return utils.ChartQuery{
		Namespace:   namespace,
		ChartName:   req.FormValue("name"), // chartName remains encoded
		Version:     req.FormValue("version"),
		AppVersion:  req.FormValue("appversion"),
		Repos:       repos,
		Categories:  categories,
		SearchQuery: req.FormValue("q"),
	}
}

func getAllChartCategories(cq utils.ChartQuery) (apiChartCategoryListResponse, error) {
	chartCategories, err := manager.GetAllChartCategories(cq)
	return newChartCategoryListResponse(chartCategories), err
}

// getChartCategories returns all the distinct chart categories name and count
func getChartCategories(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, _, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	cq := extractChartQueryFromRequest(namespace, repo, req)

	chartCategories, err := getAllChartCategories(cq)
	if err != nil {
		log.WithError(err).Error("could not fetch categories")
		response.NewErrorResponse(http.StatusInternalServerError, "could not fetch chart categories").Write(w)
		return
	}
	response.NewDataResponse(chartCategories).Write(w)
}

// getChart returns the chart from the given repo
func getChart(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, _, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	chartID := getChartID(repo, params["chartName"]) // chartName remains encoded

	chart, err := manager.GetChart(namespace, chartID)
	if err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart").Write(w)
		return
	}

	cr := newChartResponse(&chart)
	response.NewDataResponse(cr).Write(w)
}

// listChartVersions returns a list of chart versions for the given chart
func listChartVersions(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, _, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	chartID := getChartID(repo, params["chartName"]) // chartName remains encoded

	chart, err := manager.GetChart(namespace, chartID)
	if err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart").Write(w)
		return
	}

	cvl := newChartVersionListResponse(&chart)
	response.NewDataResponse(cvl).Write(w)
}

// getChartVersion returns the given chart version
func getChartVersion(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, version, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	chartID := getChartID(repo, params["chartName"]) // chartName remains encoded

	chart, err := manager.GetChartVersion(namespace, chartID, version)
	if err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		response.NewErrorResponse(http.StatusNotFound, "could not find chart version").Write(w)
		return
	}

	cvr := newChartVersionResponse(&chart, chart.ChartVersions[0])
	response.NewDataResponse(cvr).Write(w)
}

// getChartIcon returns the icon for a given chart
func getChartIcon(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, _, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	chartID := getChartID(repo, params["chartName"]) // chartName remains encoded

	chart, err := manager.GetChart(namespace, chartID)
	if err != nil {
		log.WithError(err).Errorf("could not find chart with id %s", chartID)
		http.NotFound(w, req)
		return
	}

	if len(chart.RawIcon) == 0 {
		http.NotFound(w, req)
		return
	}

	if chart.IconContentType != "" {
		// Force the Content-Type header because the autogenerated type does not work for
		// image/svg+xml. It is detected as plain text
		w.Header().Set("Content-Type", chart.IconContentType)
	}

	w.Write(chart.RawIcon)
}

// getChartVersionReadme returns the README for a given chart
func getChartVersionReadme(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, version, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	fileID := fmt.Sprintf("%s-%s", getChartID(repo, params["chartName"]), version) // chartName remains encoded

	files, err := manager.GetChartFiles(namespace, fileID)
	if err != nil {
		log.WithError(err).Errorf("could not find files with id %s", fileID)
		http.NotFound(w, req)
		return
	}
	readme := []byte(files.Readme)
	if len(readme) == 0 {
		log.Errorf("could not find a README for id %s", fileID)
		http.NotFound(w, req)
		return
	}
	w.Write(readme)
}

// getChartVersionValues returns the values.yaml for a given chart
func getChartVersionValues(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, version, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	fileID := fmt.Sprintf("%s-%s", getChartID(repo, params["chartName"]), version) // chartName remains encoded

	files, err := manager.GetChartFiles(namespace, fileID)
	if err != nil {
		log.WithError(err).Errorf("could not find values.yaml with id %s", fileID)
		http.NotFound(w, req)
		return
	}

	w.Write([]byte(files.Values))
}

// getChartVersionSchema returns the values.schema.json for a given chart
func getChartVersionSchema(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, version, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	fileID := fmt.Sprintf("%s-%s", getChartID(repo, params["chartName"]), version) // chartName remains encoded

	files, err := manager.GetChartFiles(namespace, fileID)
	if err != nil {
		log.WithError(err).Errorf("could not find values.schema.json with id %s", fileID)
		http.NotFound(w, req)
		return
	}

	w.Write([]byte(files.Schema))
}

// listChartsWithFilters returns the list of repos that contains the given chart and the latest version found
func listChartsWithFilters(w http.ResponseWriter, req *http.Request, params Params) {
	namespace, repo, _, paramErr, err := extractDecodedNamespaceAndRepoAndVersionParams(params)
	if err != nil {
		handleDecodeError(paramErr, w, err)
		return
	}
	cq := extractChartQueryFromRequest(namespace, repo, req)

	pageNumber, pageSize := getPageAndSizeParams(req)
	charts, totalPages, err := manager.GetPaginatedChartListWithFilters(cq, pageNumber, pageSize)
	if err != nil {
		log.WithError(err).Errorf("could not find charts with the given namespace=%s, chartName=%s, version=%s, appversion=%s, repos=%s, categories=%s, searchQuery=%s",
			cq.Namespace, cq.ChartName, cq.Version, cq.AppVersion, cq.Repos, cq.Categories, cq.SearchQuery,
		)
		// continue to return empty list
	}

	chartResponse := charts
	cl := newChartListResponse(chartResponse)

	response.NewDataResponseWithMeta(cl, meta{TotalPages: totalPages}).Write(w)
}

func newChartResponse(c *models.Chart) *apiResponse {
	latestCV := c.ChartVersions[0]
	namespace := c.Repo.Namespace
	chartPath := fmt.Sprintf("%s/ns/%s/charts/", pathPrefix, namespace)
	return &apiResponse{
		Type:       "chart",
		ID:         c.ID,
		Attributes: blankRawIconAndChartVersions(chartAttributes(namespace, *c)),
		Links:      selfLink{chartPath + c.ID},
		Relationships: relMap{
			"latestChartVersion": rel{
				Data:  chartVersionAttributes(namespace, c.Repo.Name, c.Name, latestCV),
				Links: selfLink{chartPath + c.ID + "/versions/" + latestCV.Version},
			},
		},
	}
}

func newChartCategoryResponse(c *models.ChartCategory) *models.ChartCategory {
	return &models.ChartCategory{
		Name:  c.Name,
		Count: c.Count,
	}
}

// blankRawIconAndChartVersions returns the same chart data but with a blank raw icon field and no chartversions.
// TODO(mnelson): The raw icon data should be stored in a separate postgresql column
// rather than the json field so that this isn't necessary.
func blankRawIconAndChartVersions(c models.Chart) models.Chart {
	c.RawIcon = nil
	c.ChartVersions = []models.ChartVersion{}
	return c
}

func newChartListResponse(charts []*models.Chart) apiListResponse {
	cl := apiListResponse{}
	for _, c := range charts {
		cl = append(cl, newChartResponse(c))
	}
	return cl
}

func newChartCategoryListResponse(charts []*models.ChartCategory) apiChartCategoryListResponse {
	cl := apiChartCategoryListResponse{}
	for _, c := range charts {
		cl = append(cl, newChartCategoryResponse(c))
	}
	return cl
}

func chartVersionAttributes(namespace, chartRepoName, chartNameUnencoded string, cv models.ChartVersion) models.ChartVersion {
	versionPath := fmt.Sprintf("%s/ns/%s/assets/%s/versions/%s/", pathPrefix, namespace, getChartID(chartRepoName, chartNameUnencoded), cv.Version)
	cv.Readme = versionPath + "README.md"
	cv.Values = versionPath + "values.yaml"
	return cv
}

func chartAttributes(namespace string, c models.Chart) models.Chart {
	if c.RawIcon != nil {
		c.Icon = pathPrefix + "/ns/" + namespace + "/assets/" + c.ID + "/logo"
	} else {
		// If the icon wasn't processed, it is either not set or invalid
		c.Icon = ""
	}
	return c
}

func newChartVersionResponse(c *models.Chart, cv models.ChartVersion) *apiResponse {
	namespace := c.Repo.Namespace
	chartPath := fmt.Sprintf("%s/ns/%s/charts/%s", pathPrefix, namespace, c.ID)
	return &apiResponse{
		Type:       "chartVersion",
		ID:         fmt.Sprintf("%s-%s", c.ID, cv.Version),
		Attributes: chartVersionAttributes(namespace, c.Repo.Name, c.Name, cv),
		Links:      selfLink{chartPath + "/versions/" + cv.Version},
		Relationships: relMap{
			"chart": rel{
				Data:  blankRawIconAndChartVersions(chartAttributes(namespace, *c)),
				Links: selfLink{chartPath},
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

func handleDecodeError(paramErr string, w http.ResponseWriter, err error) {
	log.WithError(err).Errorf("could not decode param %s", paramErr)
	response.NewErrorResponse(http.StatusBadRequest, "could not decode params").Write(w)
}

func getChartID(chartRepoName, chartName string) string {
	return fmt.Sprintf("%s/%s", chartRepoName, chartName)
}
