// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func PrettyPrint(o interface{}) string {
	prettyBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	return string(prettyBytes)
}

func PreferObjectName(o interface{}) string {
	if o == nil {
		return "<nil>"
	} else if obj, ok := o.(ctrlclient.Object); ok {
		name := obj.GetName()
		namespace := obj.GetNamespace()
		return fmt.Sprintf("%s/%s", namespace, name)
	} else {
		return PrettyPrint(o)
	}
}

func NamespacedName(obj ctrlclient.Object) (*types.NamespacedName, error) {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	if name != "" && namespace != "" {
		return &types.NamespacedName{Name: name, Namespace: namespace}, nil
	} else {
		return nil,
			status.Errorf(codes.Internal,
				"required fields 'metadata.name' and/or 'metadata.namespace' not found on resource: %v",
				PrettyPrint(obj))
	}
}
