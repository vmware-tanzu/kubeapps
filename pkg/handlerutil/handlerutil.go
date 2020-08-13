package handlerutil

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
)

// Params a key-value map of path params
type Params map[string]string

// WithParams can be used to wrap handlers to take an extra arg for path params
type WithParams func(http.ResponseWriter, *http.Request, Params)

func (h WithParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(w, req, vars)
}

// WithoutParams can be used to wrap handlers that doesn't take params
type WithoutParams func(http.ResponseWriter, *http.Request)

func (h WithoutParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h(w, req)
}

func isNotFound(err error) bool {
	// TODO(mnelson): When helm updates to use golang wrapped errors, switch these
	// to use errors.is(err, ...) etc.
	return strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no revision for release")
}

func isAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "is still in use") || strings.Contains(err.Error(), "already exists")
}

func isForbidden(err error) bool {
	return strings.Contains(err.Error(), "Unauthorized") || strings.Contains(err.Error(), "forbidden")
}

func isUnprocessable(err error) bool {
	re := regexp.MustCompile(`[rR]elease.*failed`)
	return re.MatchString(err.Error())
}

// ErrorCode returns the int representing an error.
func ErrorCode(err error) int {
	return ErrorCodeWithDefault(err, http.StatusInternalServerError)
}

// ErrorCodeWithDefault returns the int representing an error with a default value.
func ErrorCodeWithDefault(err error, defaultCode int) int {
	if isAlreadyExists(err) {
		return http.StatusConflict
	} else if isForbidden(err) {
		return http.StatusForbidden
	} else if isNotFound(err) {
		return http.StatusNotFound
	} else if isUnprocessable(err) {
		return http.StatusUnprocessableEntity
	}
	return defaultCode
}

// ParseAndGetChart request and parse a chart.
func ParseAndGetChart(req *http.Request, cu chartUtils.Resolver, requireV1Support bool) (*chartUtils.Details, *chartUtils.ChartMultiVersion, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}
	chartDetails, err := cu.ParseDetails(body)
	if err != nil {
		return nil, nil, err
	}

	netClient, err := cu.InitNetClient(chartDetails, auth.ExtractToken(req.Header.Get("Authorization")))
	if err != nil {
		return nil, nil, err
	}
	ch, err := cu.GetChart(chartDetails, netClient, requireV1Support)
	if err != nil {
		return nil, nil, err
	}
	return chartDetails, ch, nil
}

func QueryParamIsTruthy(param string, req *http.Request) bool {
	value := req.URL.Query().Get(param)
	return value == "1" || value == "true"
}
