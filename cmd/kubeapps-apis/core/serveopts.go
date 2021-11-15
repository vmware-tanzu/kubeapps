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

package core

import (
	"context"

	"flag"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	klogv2 "k8s.io/klog/v2"
)

// ServeOptions encapsulates the available command-line options.
type ServeOptions struct {
	Port                     int
	PluginDirs               []string
	ClustersConfigPath       string
	PluginConfigPath         string
	PinnipedProxyURL         string
	UnsafeLocalDevKubeconfig bool
}

// GatewayHandlerArgs is a helper struct just encapsulating all the args
// required when registering an HTTP handler for the gateway.
type GatewayHandlerArgs struct {
	Ctx         context.Context
	Mux         *runtime.ServeMux
	Addr        string
	DialOptions []grpc.DialOption
}

// KubernetesConfigGetter is a function type used throughout the apis server so
// that call-sites don't need to know how to obtain an authenticated client, but
// rather can just pass the request context and the cluster to get one.
type KubernetesConfigGetter func(ctx context.Context, cluster string) (*rest.Config, error)

type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
}

type KubeappsApisLogger struct {
	//use "k8s.io/klog/v2" as default
}

func NewBuiltinKlogger() KubeappsApisLogger {
	klogv2.InitFlags(nil) // initializing the flags
	flag.Set("v", "5")
	flag.Parse()
	//defer klogv2.Flush() // flushes all pending log I/O
	return KubeappsApisLogger{}
}
func (l KubeappsApisLogger) Error(args ...interface{}) {
	klogv2.Error(args...)
}

func (l KubeappsApisLogger) Errorf(format string, args ...interface{}) {
	klogv2.Errorf(format, args...)
}
func (l KubeappsApisLogger) Info(args ...interface{}) {
	klogv2.Info(args...)
}

func (l KubeappsApisLogger) Infof(format string, args ...interface{}) {
	klogv2.Infof(format, args...)
}
func (l KubeappsApisLogger) Fatal(args ...interface{}) {
	klogv2.Fatal(args...)
}

func (l KubeappsApisLogger) Fatalf(format string, args ...interface{}) {
	klogv2.Fatalf(format, args...)
}
