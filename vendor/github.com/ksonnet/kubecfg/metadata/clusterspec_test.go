package metadata

import (
	"path/filepath"
	"testing"
)

type parseSuccess struct {
	input  string
	target ClusterSpec
}

var successTests = []parseSuccess{
	{"version:v1.7.1", &clusterSpecVersion{"v1.7.1"}},
	{"file:swagger.json", &clusterSpecFile{"swagger.json", testFS}},
	{"url:file:///some_file", &clusterSpecLive{"file:///some_file"}},
}

func TestClusterSpecParsingSuccess(t *testing.T) {
	for _, test := range successTests {
		parsed, err := parseClusterSpec(test.input, testFS)
		if err != nil {
			t.Errorf("Failed to parse spec: %v", err)
		}

		parsedResource := parsed.resource()
		targetResource := test.target.resource()

		switch pt := parsed.(type) {
		case *clusterSpecLive:
		case *clusterSpecVersion:
			if parsedResource != targetResource {
				t.Errorf("Expected version '%v', got '%v'", parsedResource, targetResource)
			}
		case *clusterSpecFile:
			// Techncially we're cheating here by passing a *relative path*
			// into `newPathSpec` instead of an absolute one. This is to
			// make it work on multiple machines. We convert it here, after
			// the fact.
			absPath, err := filepath.Abs(targetResource)
			if err != nil {
				t.Errorf("Failed to convert `file:` spec to an absolute path: %v", err)
			}

			if parsedResource != absPath {
				t.Errorf("Expected path '%v', got '%v'", absPath, parsedResource)
			}
		default:
			t.Errorf("Unknown cluster spec type '%v'", pt)
		}
	}
}

type parseFailure struct {
	input    string
	errorMsg string
}

var failureTests = []parseFailure{
	{"fakeprefix:foo", "Could not parse cluster spec 'fakeprefix:foo'"},
	{"foo:", "Invalid API specification 'foo:'"},
	{"version:", "Invalid API specification 'version:'"},
	{"file:", "Invalid API specification 'file:'"},
	{"url:", "Invalid API specification 'url:'"},
}

func TestClusterSpecParsingFailure(t *testing.T) {
	for _, test := range failureTests {
		_, err := parseClusterSpec(test.input, testFS)
		if err == nil {
			t.Errorf("Cluster spec parse for '%s' should have failed, but succeeded", test.input)
		} else if msg := err.Error(); msg != test.errorMsg {
			t.Errorf("Expected cluster spec parse error: '%s', got: '%s'", test.errorMsg, msg)
		}
	}
}
