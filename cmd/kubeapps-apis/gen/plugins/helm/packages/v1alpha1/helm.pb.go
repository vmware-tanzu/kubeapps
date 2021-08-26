// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.1
// source: kubeappsapis/plugins/helm/packages/v1alpha1/helm.proto

package v1alpha1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	v1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	_ "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
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

// InstalledPackageDetailCustomDataHelm
//
// InstalledPackageDetailCustomDataHelm is a message type used for the
// InstalledPackageDetail.CustomDetail field by the helm plugin.
type InstalledPackageDetailCustomDataHelm struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ReleaseRevision
	//
	// A number identifying the Helm revision
	ReleaseRevision int32 `protobuf:"varint,1,opt,name=release_revision,json=releaseRevision,proto3" json:"release_revision,omitempty"`
}

func (x *InstalledPackageDetailCustomDataHelm) Reset() {
	*x = InstalledPackageDetailCustomDataHelm{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstalledPackageDetailCustomDataHelm) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstalledPackageDetailCustomDataHelm) ProtoMessage() {}

func (x *InstalledPackageDetailCustomDataHelm) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstalledPackageDetailCustomDataHelm.ProtoReflect.Descriptor instead.
func (*InstalledPackageDetailCustomDataHelm) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescGZIP(), []int{0}
}

func (x *InstalledPackageDetailCustomDataHelm) GetReleaseRevision() int32 {
	if x != nil {
		return x.ReleaseRevision
	}
	return 0
}

var File_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto protoreflect.FileDescriptor

var file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDesc = []byte{
	0x0a, 0x36, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x68, 0x65,
	0x6c, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2b, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x68,
	0x65, 0x6c, 0x6d, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x32, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69,
	0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x30, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70,
	0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51, 0x0a, 0x24, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x44, 0x61, 0x74, 0x61, 0x48, 0x65, 0x6c,
	0x6d, 0x12, 0x29, 0x0a, 0x10, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x72, 0x65, 0x6c,
	0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x32, 0xd7, 0x09, 0x0a,
	0x13, 0x48, 0x65, 0x6c, 0x6d, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0xf6, 0x01, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69,
	0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d,
	0x61, 0x72, 0x69, 0x65, 0x73, 0x12, 0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73,
	0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41,
	0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x49, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62,
	0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x41, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x3b, 0x12, 0x39, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c,
	0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x12, 0xeb, 0x01,
	0x0a, 0x19, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x45, 0x2e, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x46, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69,
	0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69,
	0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61,
	0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3f, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x39, 0x12, 0x37, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c,
	0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0xf2, 0x01, 0x0a, 0x1b,
	0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x47, 0x2e, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61,
	0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76,
	0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x40,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x3a, 0x12, 0x38, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0xf6, 0x01, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65,
	0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65,
	0x73, 0x12, 0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61,
	0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x49, 0x2e, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x41, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x3b, 0x12, 0x39,
	0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x12, 0xea, 0x01, 0x0a, 0x19, 0x47, 0x65,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x45, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x46,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65,
	0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x38, 0x12, 0x36,
	0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x42, 0x53, 0x5a, 0x51, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x6b, 0x75,
	0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61,
	0x70, 0x70, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescOnce sync.Once
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescData = file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDesc
)

func file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescGZIP() []byte {
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescOnce.Do(func() {
		file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescData = protoimpl.X.CompressGZIP(file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescData)
	})
	return file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDescData
}

var file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_goTypes = []interface{}{
	(*InstalledPackageDetailCustomDataHelm)(nil),          // 0: kubeappsapis.plugins.helm.packages.v1alpha1.InstalledPackageDetailCustomDataHelm
	(*v1alpha1.GetAvailablePackageSummariesRequest)(nil),  // 1: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesRequest
	(*v1alpha1.GetAvailablePackageDetailRequest)(nil),     // 2: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailRequest
	(*v1alpha1.GetAvailablePackageVersionsRequest)(nil),   // 3: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsRequest
	(*v1alpha1.GetInstalledPackageSummariesRequest)(nil),  // 4: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesRequest
	(*v1alpha1.GetInstalledPackageDetailRequest)(nil),     // 5: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailRequest
	(*v1alpha1.GetAvailablePackageSummariesResponse)(nil), // 6: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesResponse
	(*v1alpha1.GetAvailablePackageDetailResponse)(nil),    // 7: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailResponse
	(*v1alpha1.GetAvailablePackageVersionsResponse)(nil),  // 8: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsResponse
	(*v1alpha1.GetInstalledPackageSummariesResponse)(nil), // 9: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesResponse
	(*v1alpha1.GetInstalledPackageDetailResponse)(nil),    // 10: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailResponse
}
var file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_depIdxs = []int32{
	1,  // 0: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageSummaries:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesRequest
	2,  // 1: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageDetail:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailRequest
	3,  // 2: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageVersions:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsRequest
	4,  // 3: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetInstalledPackageSummaries:input_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesRequest
	5,  // 4: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetInstalledPackageDetail:input_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailRequest
	6,  // 5: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageSummaries:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesResponse
	7,  // 6: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageDetail:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailResponse
	8,  // 7: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetAvailablePackageVersions:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsResponse
	9,  // 8: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetInstalledPackageSummaries:output_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesResponse
	10, // 9: kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService.GetInstalledPackageDetail:output_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailResponse
	5,  // [5:10] is the sub-list for method output_type
	0,  // [0:5] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_init() }
func file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_init() {
	if File_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstalledPackageDetailCustomDataHelm); i {
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
			RawDescriptor: file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_goTypes,
		DependencyIndexes: file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_depIdxs,
		MessageInfos:      file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_msgTypes,
	}.Build()
	File_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto = out.File
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_rawDesc = nil
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_goTypes = nil
	file_kubeappsapis_plugins_helm_packages_v1alpha1_helm_proto_depIdxs = nil
}
