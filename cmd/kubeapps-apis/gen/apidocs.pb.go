// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: kubeappsapis/apidocs/v1alpha1/apidocs.proto

package gen

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_kubeappsapis_apidocs_v1alpha1_apidocs_proto protoreflect.FileDescriptor

var file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x61,
	0x70, 0x69, 0x64, 0x6f, 0x63, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x61, 0x70, 0x69, 0x64, 0x6f, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1d, 0x6b,
	0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x64,
	0x6f, 0x63, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69,
	0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x42, 0xff, 0x0d, 0x92,
	0x41, 0xc3, 0x0d, 0x12, 0xb0, 0x0b, 0x0a, 0x0c, 0x4b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73,
	0x20, 0x41, 0x50, 0x49, 0x12, 0x9d, 0x0a, 0x5b, 0x21, 0x5b, 0x4d, 0x61, 0x69, 0x6e, 0x20, 0x50,
	0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f,
	0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61,
	0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70,
	0x73, 0x2f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x73, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2d, 0x6d, 0x61, 0x69,
	0x6e, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x2f, 0x62, 0x61, 0x64, 0x67, 0x65, 0x2e, 0x73, 0x76, 0x67,
	0x29, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e,
	0x7a, 0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2d, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x79, 0x61, 0x6d, 0x6c,
	0x29, 0x0a, 0x20, 0x0a, 0x20, 0x5b, 0x4b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x5d, 0x28,
	0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f,
	0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x29, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x77,
	0x65, 0x62, 0x2d, 0x62, 0x61, 0x73, 0x65, 0x64, 0x20, 0x55, 0x49, 0x20, 0x66, 0x6f, 0x72, 0x20,
	0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x69, 0x6e, 0x67, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x20, 0x69, 0x6e, 0x20, 0x4b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65,
	0x73, 0x20, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x0a, 0x20, 0x0a, 0x20, 0x4e,
	0x6f, 0x74, 0x65, 0x3a, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x64, 0x6f,
	0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x69, 0x73, 0x20, 0x73,
	0x74, 0x69, 0x6c, 0x6c, 0x20, 0x69, 0x6e, 0x20, 0x61, 0x6e, 0x20, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x61, 0x6c, 0x20, 0x73, 0x74, 0x61, 0x67, 0x65, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x69, 0x73, 0x20,
	0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x20, 0x74, 0x6f, 0x20, 0x63, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x2e, 0x20, 0x42, 0x65, 0x66, 0x6f, 0x72, 0x65, 0x20, 0x63, 0x6f, 0x75, 0x70, 0x6c, 0x69,
	0x6e, 0x67, 0x20, 0x74, 0x6f, 0x20, 0x69, 0x74, 0x2c, 0x20, 0x70, 0x6c, 0x65, 0x61, 0x73, 0x65,
	0x20, 0x5b, 0x64, 0x72, 0x6f, 0x70, 0x20, 0x75, 0x73, 0x20, 0x61, 0x6e, 0x20, 0x69, 0x73, 0x73,
	0x75, 0x65, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61,
	0x6e, 0x7a, 0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x69, 0x73, 0x73,
	0x75, 0x65, 0x73, 0x2f, 0x6e, 0x65, 0x77, 0x2f, 0x63, 0x68, 0x6f, 0x6f, 0x73, 0x65, 0x29, 0x20,
	0x6f, 0x72, 0x20, 0x72, 0x65, 0x61, 0x63, 0x68, 0x20, 0x75, 0x73, 0x20, 0x5b, 0x76, 0x69, 0x61,
	0x20, 0x53, 0x6c, 0x61, 0x63, 0x6b, 0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f,
	0x6b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x2e, 0x73, 0x6c, 0x61, 0x63, 0x6b,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x29, 0x20, 0x74, 0x6f, 0x20, 0x6b, 0x6e, 0x6f, 0x77, 0x20,
	0x6d, 0x6f, 0x72, 0x65, 0x20, 0x61, 0x62, 0x6f, 0x75, 0x74, 0x20, 0x79, 0x6f, 0x75, 0x72, 0x20,
	0x75, 0x73, 0x65, 0x20, 0x63, 0x61, 0x73, 0x65, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x73, 0x65, 0x65,
	0x20, 0x68, 0x6f, 0x77, 0x20, 0x77, 0x65, 0x20, 0x63, 0x61, 0x6e, 0x20, 0x61, 0x73, 0x73, 0x69,
	0x73, 0x74, 0x20, 0x79, 0x6f, 0x75, 0x2e, 0x0a, 0x20, 0x23, 0x23, 0x23, 0x23, 0x20, 0x44, 0x65,
	0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x72, 0x20, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x0a, 0x20, 0x2d, 0x20, 0x54, 0x68, 0x65, 0x20, 0x5b, 0x4b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x20, 0x61, 0x72, 0x63, 0x68, 0x69, 0x74, 0x65, 0x63, 0x74,
	0x75, 0x72, 0x65, 0x20, 0x6f, 0x76, 0x65, 0x72, 0x76, 0x69, 0x65, 0x77, 0x5d, 0x28, 0x68, 0x74,
	0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f, 0x6d, 0x61, 0x69, 0x6e,
	0x2f, 0x73, 0x69, 0x74, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2f, 0x64, 0x6f,
	0x63, 0x73, 0x2f, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x67, 0x72,
	0x6f, 0x75, 0x6e, 0x64, 0x2f, 0x61, 0x72, 0x63, 0x68, 0x69, 0x74, 0x65, 0x63, 0x74, 0x75, 0x72,
	0x65, 0x2e, 0x6d, 0x64, 0x29, 0x2e, 0x0a, 0x20, 0x2d, 0x20, 0x54, 0x68, 0x65, 0x20, 0x5b, 0x4b,
	0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x20, 0x44, 0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65,
	0x72, 0x20, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5d,
	0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75,
	0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f, 0x6d,
	0x61, 0x69, 0x6e, 0x2f, 0x73, 0x69, 0x74, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x2f, 0x64, 0x6f, 0x63, 0x73, 0x2f, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x72,
	0x2f, 0x52, 0x45, 0x41, 0x44, 0x4d, 0x45, 0x2e, 0x6d, 0x64, 0x29, 0x20, 0x66, 0x6f, 0x72, 0x20,
	0x69, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x20, 0x6f, 0x6e, 0x20,
	0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x75, 0x70, 0x20, 0x74, 0x68, 0x65, 0x20, 0x64,
	0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x72, 0x20, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e,
	0x6d, 0x65, 0x6e, 0x74, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x64, 0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70,
	0x69, 0x6e, 0x67, 0x20, 0x6f, 0x6e, 0x20, 0x4b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x20,
	0x61, 0x6e, 0x64, 0x20, 0x69, 0x74, 0x73, 0x20, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x0a, 0x20, 0x2d, 0x20, 0x54, 0x68, 0x65, 0x20, 0x5b, 0x4b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x20, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x20, 0x47, 0x75, 0x69, 0x64, 0x65,
	0x5d, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a,
	0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f,
	0x6d, 0x61, 0x69, 0x6e, 0x2f, 0x73, 0x69, 0x74, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x2f, 0x64, 0x6f, 0x63, 0x73, 0x2f, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x72, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65,
	0x72, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x6d, 0x64, 0x29, 0x20, 0x66, 0x6f, 0x72, 0x20,
	0x69, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x20, 0x6f, 0x6e, 0x20,
	0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x75, 0x70, 0x20, 0x74, 0x68, 0x65, 0x20, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x20, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74,
	0x20, 0x61, 0x6e, 0x64, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x69, 0x6e, 0x67, 0x20, 0x4b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x20, 0x66, 0x72, 0x6f, 0x6d, 0x20, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x2e, 0x0a, 0x1a, 0x3a, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d,
	0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x62,
	0x6c, 0x6f, 0x62, 0x2f, 0x6d, 0x61, 0x69, 0x6e, 0x2f, 0x4c, 0x49, 0x43, 0x45, 0x4e, 0x53, 0x45,
	0x2a, 0x3d, 0x0a, 0x0a, 0x41, 0x70, 0x61, 0x63, 0x68, 0x65, 0x2d, 0x32, 0x2e, 0x30, 0x12, 0x2f,
	0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x77, 0x77, 0x77, 0x2e, 0x61, 0x70, 0x61, 0x63, 0x68,
	0x65, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x73, 0x2f, 0x4c,
	0x49, 0x43, 0x45, 0x4e, 0x53, 0x45, 0x2d, 0x32, 0x2e, 0x30, 0x2e, 0x68, 0x74, 0x6d, 0x6c, 0x32,
	0x05, 0x30, 0x2e, 0x31, 0x2e, 0x30, 0x1a, 0x0e, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e,
	0x31, 0x3a, 0x38, 0x30, 0x38, 0x30, 0x22, 0x05, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x2a, 0x02, 0x01,
	0x02, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a,
	0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x52, 0x50, 0x0a, 0x03, 0x34, 0x30, 0x31, 0x12, 0x49, 0x0a, 0x47,
	0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x20, 0x77, 0x68, 0x65, 0x6e, 0x20, 0x74, 0x68,
	0x65, 0x20, 0x75, 0x73, 0x65, 0x72, 0x20, 0x64, 0x6f, 0x65, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20,
	0x68, 0x61, 0x76, 0x65, 0x20, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x20,
	0x74, 0x6f, 0x20, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x20, 0x74, 0x68, 0x65, 0x20, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x5a, 0x23, 0x0a, 0x21, 0x0a, 0x0a, 0x41, 0x70, 0x69,
	0x4b, 0x65, 0x79, 0x41, 0x75, 0x74, 0x68, 0x12, 0x13, 0x08, 0x02, 0x1a, 0x0d, 0x41, 0x75, 0x74,
	0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x02, 0x62, 0x10, 0x0a, 0x0e,
	0x0a, 0x0a, 0x41, 0x70, 0x69, 0x4b, 0x65, 0x79, 0x41, 0x75, 0x74, 0x68, 0x12, 0x00, 0x72, 0x46,
	0x0a, 0x1a, 0x4b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x20, 0x47, 0x69, 0x74, 0x48, 0x75,
	0x62, 0x20, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x28, 0x68, 0x74,
	0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f,
	0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x6b, 0x75, 0x62,
	0x65, 0x61, 0x70, 0x70, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67, 0x65, 0x6e, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_goTypes = []any{}
var file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_init() }
func file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_init() {
	if File_kubeappsapis_apidocs_v1alpha1_apidocs_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_goTypes,
		DependencyIndexes: file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_depIdxs,
	}.Build()
	File_kubeappsapis_apidocs_v1alpha1_apidocs_proto = out.File
	file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_rawDesc = nil
	file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_goTypes = nil
	file_kubeappsapis_apidocs_v1alpha1_apidocs_proto_depIdxs = nil
}
