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

package error

import "time"

// Stalling is the reconciliation stalled state error. It contains an error
// and a reason for the stalled condition.
type Stalling struct {
	// Reason is the stalled condition reason string.
	Reason string
	// Err is the error that caused stalling. This can be used as the message in
	// stalled condition.
	Err error
}

// Error implements error interface.
func (se *Stalling) Error() string {
	return se.Err.Error()
}

// Unwrap returns the underlying error.
func (se *Stalling) Unwrap() error {
	return se.Err
}

// Event is an error event. It can be used to construct an event to be
// recorded.
type Event struct {
	// Reason is the reason for the event error.
	Reason string
	// Error is the actual error for the event.
	Err error
}

// Error implements error interface.
func (ee *Event) Error() string {
	return ee.Err.Error()
}

// Unwrap returns the underlying error.
func (ee *Event) Unwrap() error {
	return ee.Err
}

// Waiting is the reconciliation wait state error. It contains an error, wait
// duration and a reason for the wait.
type Waiting struct {
	// RequeueAfter is the wait duration after which to requeue.
	RequeueAfter time.Duration
	// Reason is the reason for the wait.
	Reason string
	// Err is the error that caused the wait.
	Err error
}

// Error implement error interface.
func (we *Waiting) Error() string {
	return we.Err.Error()
}

// Unwrap returns the underlying error.
func (we *Waiting) Unwrap() error {
	return we.Err
}
