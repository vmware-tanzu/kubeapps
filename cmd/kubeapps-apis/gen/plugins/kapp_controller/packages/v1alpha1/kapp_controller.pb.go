// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller.proto

package v1alpha1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	v1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	v1alpha11 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
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

// GetPackageRepositories
//
// Request for GetPackageRepositories
type GetPackageRepositoriesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The context (cluster/namespace) for the request
	Context *v1alpha1.Context `protobuf:"bytes,1,opt,name=context,proto3" json:"context,omitempty"` // TODO: Add standard filters.
}

func (x *GetPackageRepositoriesRequest) Reset() {
	*x = GetPackageRepositoriesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPackageRepositoriesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPackageRepositoriesRequest) ProtoMessage() {}

func (x *GetPackageRepositoriesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPackageRepositoriesRequest.ProtoReflect.Descriptor instead.
func (*GetPackageRepositoriesRequest) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescGZIP(), []int{0}
}

func (x *GetPackageRepositoriesRequest) GetContext() *v1alpha1.Context {
	if x != nil {
		return x.Context
	}
	return nil
}

// GetPackageRepositories
//
// Response for GetPackageRepositories
type GetPackageRepositoriesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repositories
	//
	// List of PackageRepository
	Repositories []*PackageRepository `protobuf:"bytes,1,rep,name=repositories,proto3" json:"repositories,omitempty"`
}

func (x *GetPackageRepositoriesResponse) Reset() {
	*x = GetPackageRepositoriesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPackageRepositoriesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPackageRepositoriesResponse) ProtoMessage() {}

func (x *GetPackageRepositoriesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPackageRepositoriesResponse.ProtoReflect.Descriptor instead.
func (*GetPackageRepositoriesResponse) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescGZIP(), []int{1}
}

func (x *GetPackageRepositoriesResponse) GetRepositories() []*PackageRepository {
	if x != nil {
		return x.Repositories
	}
	return nil
}

// PackageRepository
//
// A PackageRepository defines a repository of packages for installation.
type PackageRepository struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Package repository name
	//
	// The name identifying package repository on the cluster.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Package repository namespace
	//
	// An optional namespace for namespaced package repositories.
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// Package repository URL
	//
	// A url identifying the package repository location.
	Url string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	// Package repository plugin
	//
	// The plugin used to interact with this package repository.
	Plugin *v1alpha11.Plugin `protobuf:"bytes,4,opt,name=plugin,proto3" json:"plugin,omitempty"`
}

func (x *PackageRepository) Reset() {
	*x = PackageRepository{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PackageRepository) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackageRepository) ProtoMessage() {}

func (x *PackageRepository) ProtoReflect() protoreflect.Message {
	mi := &file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackageRepository.ProtoReflect.Descriptor instead.
func (*PackageRepository) Descriptor() ([]byte, []int) {
	return file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescGZIP(), []int{2}
}

func (x *PackageRepository) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PackageRepository) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *PackageRepository) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *PackageRepository) GetPlugin() *v1alpha11.Plugin {
	if x != nil {
		return x.Plugin
	}
	return nil
}

var File_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto protoreflect.FileDescriptor

var file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDesc = []byte{
	0x0a, 0x4c, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x36,
	0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2e, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x32, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70,
	0x69, 0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x30, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x67, 0x0a, 0x1d, 0x47, 0x65,
	0x74, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x6b,
	0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x22, 0xd9, 0x02, 0x0a, 0x1e, 0x47, 0x65, 0x74, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6d, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x49, 0x2e, 0x6b,
	0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x0c, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x65, 0x73, 0x3a, 0xc7, 0x01, 0x92, 0x41, 0xc3, 0x01, 0x32, 0xc0, 0x01, 0x7b,
	0x22, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x22, 0x3a, 0x20,
	0x5b, 0x7b, 0x22, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x72, 0x65, 0x70, 0x6f, 0x2d,
	0x6e, 0x61, 0x6d, 0x65, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
	0x22, 0x2c, 0x20, 0x22, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x3a, 0x20,
	0x22, 0x22, 0x2c, 0x20, 0x22, 0x75, 0x72, 0x6c, 0x22, 0x3a, 0x20, 0x22, 0x66, 0x6f, 0x6f, 0x2e,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x2d, 0x6e, 0x61, 0x6d, 0x65, 0x2f, 0x6d,
	0x61, 0x69, 0x6e, 0x40, 0x73, 0x68, 0x61, 0x32, 0x35, 0x36, 0x3a, 0x63, 0x65, 0x63, 0x64, 0x39,
	0x62, 0x35, 0x31, 0x62, 0x31, 0x66, 0x32, 0x39, 0x61, 0x37, 0x37, 0x33, 0x61, 0x35, 0x32, 0x32,
	0x38, 0x66, 0x65, 0x30, 0x34, 0x66, 0x61, 0x65, 0x63, 0x31, 0x32, 0x31, 0x63, 0x39, 0x66, 0x62,
	0x64, 0x32, 0x39, 0x36, 0x39, 0x64, 0x65, 0x35, 0x35, 0x62, 0x30, 0x63, 0x33, 0x38, 0x30, 0x34,
	0x32, 0x36, 0x39, 0x61, 0x31, 0x64, 0x35, 0x37, 0x61, 0x61, 0x35, 0x22, 0x7d, 0x5d, 0x7d, 0x22,
	0x9b, 0x01, 0x0a, 0x11, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d,
	0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x42, 0x0a, 0x06, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x6b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x32, 0xc4, 0x12,
	0x0a, 0x1d, 0x4b, 0x61, 0x70, 0x70, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x81, 0x02, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73,
	0x12, 0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61,
	0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72,
	0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x49, 0x2e, 0x6b, 0x75, 0x62,
	0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x4c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x46, 0x12, 0x44, 0x2f,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61,
	0x62, 0x6c, 0x65, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72,
	0x69, 0x65, 0x73, 0x12, 0xf6, 0x01, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x12, 0x45, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x46, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61,
	0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x4a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x44, 0x12, 0x42, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x8f, 0x02, 0x0a,
	0x16, 0x47, 0x65, 0x74, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x12, 0x55, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x6b,
	0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x56,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x46, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x40, 0x12, 0x3e,
	0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x12, 0xfd,
	0x01, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x47,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x4b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x45, 0x12, 0x43, 0x2f, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x81,
	0x02, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x12,
	0x48, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x49, 0x2e, 0x6b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x4c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x46, 0x12, 0x44, 0x2f, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x65, 0x64, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x69,
	0x65, 0x73, 0x12, 0xf5, 0x01, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c,
	0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x12, 0x45, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c,
	0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x46, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70,
	0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x49, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x43, 0x12, 0x41, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65,
	0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0xea, 0x01, 0x0a, 0x16, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x42, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73,
	0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x43, 0x2e, 0x6b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x47,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x41, 0x22, 0x3c, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x3a, 0x01, 0x2a, 0x12, 0xea, 0x01, 0x0a, 0x16, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x12, 0x42, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69,
	0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x43, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70,
	0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x47, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x41, 0x32, 0x3c, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61,
	0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x3a, 0x01, 0x2a, 0x12, 0xbd, 0x02, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12,
	0x42, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x43, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x61, 0x70,
	0x69, 0x73, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x99, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x92, 0x01, 0x2a, 0x8c, 0x01, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x6b, 0x61,
	0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x2f, 0x6e, 0x73, 0x2f, 0x7b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x5f,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x5f, 0x72, 0x65, 0x66, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x2e, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x7d, 0x2f, 0x7b,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x5f, 0x72, 0x65, 0x66, 0x2e, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x7d, 0x3a, 0x01, 0x2a, 0x42, 0x5e, 0x5a, 0x5c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x6b, 0x75, 0x62, 0x65,
	0x61, 0x70, 0x70, 0x73, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x61, 0x70, 0x70,
	0x73, 0x2d, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x2f, 0x6b, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescOnce sync.Once
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescData = file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDesc
)

func file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescGZIP() []byte {
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescOnce.Do(func() {
		file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescData = protoimpl.X.CompressGZIP(file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescData)
	})
	return file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDescData
}

var file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_goTypes = []interface{}{
	(*GetPackageRepositoriesRequest)(nil),                 // 0: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesRequest
	(*GetPackageRepositoriesResponse)(nil),                // 1: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesResponse
	(*PackageRepository)(nil),                             // 2: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.PackageRepository
	(*v1alpha1.Context)(nil),                              // 3: kubeappsapis.core.packages.v1alpha1.Context
	(*v1alpha11.Plugin)(nil),                              // 4: kubeappsapis.core.plugins.v1alpha1.Plugin
	(*v1alpha1.GetAvailablePackageSummariesRequest)(nil),  // 5: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesRequest
	(*v1alpha1.GetAvailablePackageDetailRequest)(nil),     // 6: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailRequest
	(*v1alpha1.GetAvailablePackageVersionsRequest)(nil),   // 7: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsRequest
	(*v1alpha1.GetInstalledPackageSummariesRequest)(nil),  // 8: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesRequest
	(*v1alpha1.GetInstalledPackageDetailRequest)(nil),     // 9: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailRequest
	(*v1alpha1.CreateInstalledPackageRequest)(nil),        // 10: kubeappsapis.core.packages.v1alpha1.CreateInstalledPackageRequest
	(*v1alpha1.UpdateInstalledPackageRequest)(nil),        // 11: kubeappsapis.core.packages.v1alpha1.UpdateInstalledPackageRequest
	(*v1alpha1.DeleteInstalledPackageRequest)(nil),        // 12: kubeappsapis.core.packages.v1alpha1.DeleteInstalledPackageRequest
	(*v1alpha1.GetAvailablePackageSummariesResponse)(nil), // 13: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesResponse
	(*v1alpha1.GetAvailablePackageDetailResponse)(nil),    // 14: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailResponse
	(*v1alpha1.GetAvailablePackageVersionsResponse)(nil),  // 15: kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsResponse
	(*v1alpha1.GetInstalledPackageSummariesResponse)(nil), // 16: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesResponse
	(*v1alpha1.GetInstalledPackageDetailResponse)(nil),    // 17: kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailResponse
	(*v1alpha1.CreateInstalledPackageResponse)(nil),       // 18: kubeappsapis.core.packages.v1alpha1.CreateInstalledPackageResponse
	(*v1alpha1.UpdateInstalledPackageResponse)(nil),       // 19: kubeappsapis.core.packages.v1alpha1.UpdateInstalledPackageResponse
	(*v1alpha1.DeleteInstalledPackageResponse)(nil),       // 20: kubeappsapis.core.packages.v1alpha1.DeleteInstalledPackageResponse
}
var file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_depIdxs = []int32{
	3,  // 0: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesRequest.context:type_name -> kubeappsapis.core.packages.v1alpha1.Context
	2,  // 1: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesResponse.repositories:type_name -> kubeappsapis.plugins.kapp_controller.packages.v1alpha1.PackageRepository
	4,  // 2: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.PackageRepository.plugin:type_name -> kubeappsapis.core.plugins.v1alpha1.Plugin
	5,  // 3: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageSummaries:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesRequest
	6,  // 4: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageDetail:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailRequest
	0,  // 5: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetPackageRepositories:input_type -> kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesRequest
	7,  // 6: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageVersions:input_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsRequest
	8,  // 7: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetInstalledPackageSummaries:input_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesRequest
	9,  // 8: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetInstalledPackageDetail:input_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailRequest
	10, // 9: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.CreateInstalledPackage:input_type -> kubeappsapis.core.packages.v1alpha1.CreateInstalledPackageRequest
	11, // 10: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.UpdateInstalledPackage:input_type -> kubeappsapis.core.packages.v1alpha1.UpdateInstalledPackageRequest
	12, // 11: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.DeleteInstalledPackage:input_type -> kubeappsapis.core.packages.v1alpha1.DeleteInstalledPackageRequest
	13, // 12: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageSummaries:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageSummariesResponse
	14, // 13: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageDetail:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageDetailResponse
	1,  // 14: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetPackageRepositories:output_type -> kubeappsapis.plugins.kapp_controller.packages.v1alpha1.GetPackageRepositoriesResponse
	15, // 15: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetAvailablePackageVersions:output_type -> kubeappsapis.core.packages.v1alpha1.GetAvailablePackageVersionsResponse
	16, // 16: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetInstalledPackageSummaries:output_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageSummariesResponse
	17, // 17: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.GetInstalledPackageDetail:output_type -> kubeappsapis.core.packages.v1alpha1.GetInstalledPackageDetailResponse
	18, // 18: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.CreateInstalledPackage:output_type -> kubeappsapis.core.packages.v1alpha1.CreateInstalledPackageResponse
	19, // 19: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.UpdateInstalledPackage:output_type -> kubeappsapis.core.packages.v1alpha1.UpdateInstalledPackageResponse
	20, // 20: kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService.DeleteInstalledPackage:output_type -> kubeappsapis.core.packages.v1alpha1.DeleteInstalledPackageResponse
	12, // [12:21] is the sub-list for method output_type
	3,  // [3:12] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_init() }
func file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_init() {
	if File_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPackageRepositoriesRequest); i {
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
		file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPackageRepositoriesResponse); i {
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
		file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PackageRepository); i {
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
			RawDescriptor: file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_goTypes,
		DependencyIndexes: file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_depIdxs,
		MessageInfos:      file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_msgTypes,
	}.Build()
	File_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto = out.File
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_rawDesc = nil
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_goTypes = nil
	file_kubeappsapis_plugins_kapp_controller_packages_v1alpha1_kapp_controller_proto_depIdxs = nil
}
