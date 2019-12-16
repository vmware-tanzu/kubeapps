package handlerutil

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"k8s.io/helm/pkg/proto/hapi/chart"
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
	return strings.Contains(err.Error(), "not found")
}

func isAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "is still in use") || strings.Contains(err.Error(), "already exists")
}

func isForbidden(err error) bool {
	return strings.Contains(err.Error(), "Unauthorized")
}

func isUnprocessable(err error) bool {
	re := regexp.MustCompile(`release.*failed`)
	return re.MatchString(err.Error())
}

func ErrorCode(err error) int {
	return ErrorCodeWithDefault(err, http.StatusInternalServerError)
}

func ErrorCodeWithDefault(err error, defaultCode int) int {
	errCode := defaultCode
	if isAlreadyExists(err) {
		errCode = http.StatusConflict
	} else if isNotFound(err) {
		errCode = http.StatusNotFound
	} else if isForbidden(err) {
		errCode = http.StatusForbidden
	} else if isUnprocessable(err) {
		errCode = http.StatusUnprocessableEntity
	}
	return errCode
}

func ParseAndGetChart(req *http.Request, cu chartUtils.Resolver) (*chartUtils.Details, *chart.Chart, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}
	chartDetails, err := cu.ParseDetails(body)
	if err != nil {
		return nil, nil, err
	}
	netClient, err := cu.InitNetClient(chartDetails)
	if err != nil {
		return nil, nil, err
	}
	requireV1Support := true
	ch, err := cu.GetChart(chartDetails, netClient, requireV1Support)
	if err != nil {
		return nil, nil, err
	}
	return chartDetails, ch.Helm2Chart, nil
}
