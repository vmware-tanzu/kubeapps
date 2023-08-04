// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: kubeappsapis/core/plugins/v1alpha1/plugins.proto

package v1alpha1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// GetConfiguredPluginsRequest
//
// Request for GetConfiguredPlugins
type GetConfiguredPluginsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetConfiguredPluginsRequest) Reset() {
	*x = GetConfiguredPluginsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfiguredPluginsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfiguredPluginsRequest) ProtoMessage() {}

func (x *GetConfiguredPluginsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfiguredPluginsRequest.ProtoReflect.Descriptor instead.
func (*GetConfiguredPluginsRequest) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescGZIP(), []int{0}
}

// GetConfiguredPluginsResponse
//
// Response for GetConfiguredPlugins
type GetConfiguredPluginsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Plugins
	//
	// List of Plugin
	Plugins []*Plugin `protobuf:"bytes,1,rep,name=plugins,proto3" json:"plugins,omitempty"`
}

func (x *GetConfiguredPluginsResponse) Reset() {
	*x = GetConfiguredPluginsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfiguredPluginsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfiguredPluginsResponse) ProtoMessage() {}

func (x *GetConfiguredPluginsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfiguredPluginsResponse.ProtoReflect.Descriptor instead.
func (*GetConfiguredPluginsResponse) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescGZIP(), []int{1}
}

func (x *GetConfiguredPluginsResponse) GetPlugins() []*Plugin {
	if x != nil {
		return x.Plugins
	}
	return nil
}

// Plugin
//
// A plugin can implement multiple services and multiple versions of a service.
type Plugin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Plugin name
	//
	// The name of the plugin, such as `fluxv2.packages` or `kapp_controller.packages`.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Plugin version
	//
	// The version of the plugin, such as v1alpha1
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *Plugin) Reset() {
	*x = Plugin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugin) ProtoMessage() {}

func (x *Plugin) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plugin.ProtoReflect.Descriptor instead.
func (*Plugin) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescGZIP(), []int{2}
}

func (x *Plugin) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Plugin) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

var File_kubeappsapis_core_plugins_v1alpha1_plugins_proto protoreflect.FileDescriptor

var file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDesc = []byte{
	0x0a, 0x30, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x22, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e,
	0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1d, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0xb5, 0x01, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73,
	0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x52, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x3a, 0x4f, 0x92, 0x41, 0x4c, 0x32,
	0x4a, 0x7b, 0x22, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0x7b, 0x22,
	0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x22, 0x2c, 0x20, 0x22, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x22, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0x7d, 0x5d, 0x7d, 0x22, 0x78, 0x0a, 0x06, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x3a, 0x40, 0x92, 0x41, 0x3d, 0x32, 0x3b, 0x7b, 0x22, 0x6e, 0x61, 0x6d, 0x65,
	0x22, 0x3a, 0x20, 0x22, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x22, 0x2c, 0x20, 0x22,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x22, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x22, 0x7d, 0x32, 0xdf, 0x01, 0x0a, 0x0e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0xcc, 0x01, 0x0a, 0x14, 0x47, 0x65, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x12, 0x3f, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x40, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69,
	0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x65, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x31, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x2b, 0x12, 0x29, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x2d,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x42, 0x4e, 0x5a, 0x4c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d, 0x74, 0x61, 0x6e,
	0x7a, 0x75, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x63, 0x6d, 0x64, 0x2f,
	0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescOnce sync.Once
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescData = file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDesc
)

func file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescGZIP() []byte {
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescOnce.Do(func() {
		file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescData = protoimpl.X.CompressGZIP(file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescData)
	})
	return file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDescData
}

var file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_goTypes = []interface{}{
	(*GetConfiguredPluginsRequest)(nil),  // 0: kubeappsapis.core.plugins.v1alpha1.GetConfiguredPluginsRequest
	(*GetConfiguredPluginsResponse)(nil), // 1: kubeappsapis.core.plugins.v1alpha1.GetConfiguredPluginsResponse
	(*Plugin)(nil),                       // 2: kubeappsapis.core.plugins.v1alpha1.Plugin
}
var file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_depIdxs = []int32{
	2, // 0: kubeappsapis.core.plugins.v1alpha1.GetConfiguredPluginsResponse.plugins:type_name -> kubeappsapis.core.plugins.v1alpha1.Plugin
	0, // 1: kubeappsapis.core.plugins.v1alpha1.PluginsService.GetConfiguredPlugins:input_type -> kubeappsapis.core.plugins.v1alpha1.GetConfiguredPluginsRequest
	1, // 2: kubeappsapis.core.plugins.v1alpha1.PluginsService.GetConfiguredPlugins:output_type -> kubeappsapis.core.plugins.v1alpha1.GetConfiguredPluginsResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_init() }
func file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_init() {
	if File_kubeappsapis_core_plugins_v1alpha1_plugins_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetConfiguredPluginsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetConfiguredPluginsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plugin); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_goTypes,
		DependencyIndexes: file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_depIdxs,
		MessageInfos:      file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_msgTypes,
	}.Build()
	File_kubeappsapis_core_plugins_v1alpha1_plugins_proto = out.File
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_rawDesc = nil
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_goTypes = nil
	file_kubeappsapis_core_plugins_v1alpha1_plugins_proto_depIdxs = nil
}
