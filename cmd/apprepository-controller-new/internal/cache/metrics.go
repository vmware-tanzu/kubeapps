/*
Copyright 2022 The Flux authors

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

package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// CacheEventTypeMiss is the event type for cache misses.
	CacheEventTypeMiss = "cache_miss"
	// CacheEventTypeHit is the event type for cache hits.
	CacheEventTypeHit = "cache_hit"
)

// CacheRecorder is a recorder for cache events.
type CacheRecorder struct {
	// cacheEventsCounter is a counter for cache events.
	cacheEventsCounter *prometheus.CounterVec
}

// NewCacheRecorder returns a new CacheRecorder.
// The configured labels are: event_type, name, namespace.
// The event_type is one of:
//   - "miss"
//   - "hit"
//   - "update"
// The name is the name of the reconciled resource.
// The namespace is the namespace of the reconciled resource.
func NewCacheRecorder() *CacheRecorder {
	return &CacheRecorder{
		cacheEventsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gotk_cache_events_total",
				Help: "Total number of cache retrieval events for a Gitops Toolkit resource reconciliation.",
			},
			[]string{"event_type", "name", "namespace"},
		),
	}
}

// Collectors returns the metrics.Collector objects for the CacheRecorder.
func (r *CacheRecorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.cacheEventsCounter,
	}
}

// IncCacheEventCount increment by 1 the cache event count for the given event type, name and namespace.
func (r *CacheRecorder) IncCacheEvents(event, name, namespace string) {
	r.cacheEventsCounter.WithLabelValues(event, name, namespace).Inc()
}

// MustMakeMetrics creates a new CacheRecorder, and registers the metrics collectors in the controller-runtime metrics registry.
func MustMakeMetrics() *CacheRecorder {
	r := NewCacheRecorder()
	metrics.Registry.MustRegister(r.Collectors()...)

	return r
}
