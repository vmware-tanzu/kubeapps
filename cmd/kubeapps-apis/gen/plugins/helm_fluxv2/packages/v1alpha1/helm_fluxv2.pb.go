// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.2
// source: kubeappsapis/plugins/helm_fluxv2/packages/v1alpha1/helm_fluxv2.proto

package v1alpha1

import (
	v1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

var File_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto protoreflect.FileDescriptor

var file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_rawDesc = []byte{
	0x0a, 0x44, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x5f, 0x66, 0x6c, 0x75, 0x78,
	0x76, 0x32, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x5f, 0x66, 0x6c, 0x75, 0x78, 0x76, 0x32,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x32, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73,
	0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x68, 0x65, 0x6c,
	0x6d, 0x5f, 0x66, 0x6c, 0x75, 0x78, 0x76, 0x32, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x32, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xd7, 0x01, 0x0a,
	0x0f, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0xc3, 0x01, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x12, 0x40, 0x2e, 0x6b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x41, 0x2e, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x26,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x12, 0x1e, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2d, 0x66, 0x6c,
	0x75, 0x78, 0x76, 0x32, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x5a, 0x5a, 0x58, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61,
	0x70, 0x70, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x5f, 0x66, 0x6c, 0x75, 0x78, 0x76, 0x32,
	0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_goTypes = []interface{}{
	(*v1alpha1.GetAvailablePackagesRequest)(nil),  // 0: kubeappsapis.core.packages.v1alpha1.GetAvailablePackagesRequest
	(*v1alpha1.GetAvailablePackagesResponse)(nil), // 1: kubeappsapis.core.packages.v1alpha1.GetAvailablePackagesResponse
}
var file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_depIdxs = []int32{
	0, // 0: kubeappsapis.plugins.helm_fluxv2.packages.v1alpha1.PackagesService.GetAvailablePackages:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackagesRequest
	1, // 1: kubeappsapis.plugins.helm_fluxv2.packages.v1alpha1.PackagesService.GetAvailablePackages:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackagesResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_init() }
func file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_init() {
	if File_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_goTypes,
		DependencyIndexes: file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_depIdxs,
	}.Build()
	File_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto = out.File
	file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_rawDesc = nil
	file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_goTypes = nil
	file_kubeappsapis_plugins_helm_fluxv2_packages_v1alpha1_helm_fluxv2_proto_depIdxs = nil
}
