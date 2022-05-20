/*
Copyright 2021 The Flux authors

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

package helm

// This list defines a set of global variables used to ensure Helm files loaded
// into memory during runtime do not exceed defined upper bound limits.
var (
	// MaxIndexSize is the max allowed file size in bytes of a ChartRepository.
	MaxIndexSize int64 = 50 << 20
	// MaxChartSize is the max allowed file size in bytes of a Helm Chart.
	MaxChartSize int64 = 10 << 20
	// MaxChartFileSize is the max allowed file size in bytes of any arbitrary
	// file originating from a chart.
	MaxChartFileSize int64 = 5 << 20
)
