// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hapi/chart/metadata.proto

package chart

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Metadata_Engine int32

const (
	Metadata_UNKNOWN Metadata_Engine = 0
	Metadata_GOTPL   Metadata_Engine = 1
)

var Metadata_Engine_name = map[int32]string{
	0: "UNKNOWN",
	1: "GOTPL",
}
var Metadata_Engine_value = map[string]int32{
	"UNKNOWN": 0,
	"GOTPL":   1,
}

func (x Metadata_Engine) String() string {
	return proto.EnumName(Metadata_Engine_name, int32(x))
}
func (Metadata_Engine) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{1, 0} }

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	// Url is an optional URL to an address for the named maintainer
	Url string `protobuf:"bytes,3,opt,name=url" json:"url,omitempty"`
}

func (m *Maintainer) Reset()                    { *m = Maintainer{} }
func (m *Maintainer) String() string            { return proto.CompactTextString(m) }
func (*Maintainer) ProtoMessage()               {}
func (*Maintainer) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *Maintainer) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Maintainer) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *Maintainer) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

// 	Metadata for a Chart file. This models the structure of a Chart.yaml file.
//
// 	Spec: https://k8s.io/helm/blob/master/docs/design/chart_format.md#the-chart-file
type Metadata struct {
	// The name of the chart
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `protobuf:"bytes,2,opt,name=home" json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `protobuf:"bytes,3,rep,name=sources" json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `protobuf:"bytes,6,rep,name=keywords" json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `protobuf:"bytes,7,rep,name=maintainers" json:"maintainers,omitempty"`
	// The name of the template engine to use. Defaults to 'gotpl'.
	Engine string `protobuf:"bytes,8,opt,name=engine" json:"engine,omitempty"`
	// The URL to an icon file.
	Icon string `protobuf:"bytes,9,opt,name=icon" json:"icon,omitempty"`
	// The API Version of this chart.
	ApiVersion string `protobuf:"bytes,10,opt,name=apiVersion" json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `protobuf:"bytes,11,opt,name=condition" json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `protobuf:"bytes,12,opt,name=tags" json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `protobuf:"bytes,13,opt,name=appVersion" json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `protobuf:"varint,14,opt,name=deprecated" json:"deprecated,omitempty"`
	// TillerVersion is a SemVer constraints on what version of Tiller is required.
	// See SemVer ranges here: https://github.com/Masterminds/semver#basic-comparisons
	TillerVersion string `protobuf:"bytes,15,opt,name=tillerVersion" json:"tillerVersion,omitempty"`
	// Annotations are additional mappings uninterpreted by Tiller,
	// made available for inspection by other applications.
	Annotations map[string]string `protobuf:"bytes,16,rep,name=annotations" json:"annotations,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `protobuf:"bytes,17,opt,name=kubeVersion" json:"kubeVersion,omitempty"`
}

func (m *Metadata) Reset()                    { *m = Metadata{} }
func (m *Metadata) String() string            { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()               {}
func (*Metadata) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *Metadata) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Metadata) GetHome() string {
	if m != nil {
		return m.Home
	}
	return ""
}

func (m *Metadata) GetSources() []string {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *Metadata) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Metadata) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Metadata) GetKeywords() []string {
	if m != nil {
		return m.Keywords
	}
	return nil
}

func (m *Metadata) GetMaintainers() []*Maintainer {
	if m != nil {
		return m.Maintainers
	}
	return nil
}

func (m *Metadata) GetEngine() string {
	if m != nil {
		return m.Engine
	}
	return ""
}

func (m *Metadata) GetIcon() string {
	if m != nil {
		return m.Icon
	}
	return ""
}

func (m *Metadata) GetApiVersion() string {
	if m != nil {
		return m.ApiVersion
	}
	return ""
}

func (m *Metadata) GetCondition() string {
	if m != nil {
		return m.Condition
	}
	return ""
}

func (m *Metadata) GetTags() string {
	if m != nil {
		return m.Tags
	}
	return ""
}

func (m *Metadata) GetAppVersion() string {
	if m != nil {
		return m.AppVersion
	}
	return ""
}

func (m *Metadata) GetDeprecated() bool {
	if m != nil {
		return m.Deprecated
	}
	return false
}

func (m *Metadata) GetTillerVersion() string {
	if m != nil {
		return m.TillerVersion
	}
	return ""
}

func (m *Metadata) GetAnnotations() map[string]string {
	if m != nil {
		return m.Annotations
	}
	return nil
}

func (m *Metadata) GetKubeVersion() string {
	if m != nil {
		return m.KubeVersion
	}
	return ""
}

func init() {
	proto.RegisterType((*Maintainer)(nil), "hapi.chart.Maintainer")
	proto.RegisterType((*Metadata)(nil), "hapi.chart.Metadata")
	proto.RegisterEnum("hapi.chart.Metadata_Engine", Metadata_Engine_name, Metadata_Engine_value)
}

func init() { proto.RegisterFile("hapi/chart/metadata.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 435 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x52, 0x5d, 0x6b, 0xd4, 0x40,
	0x14, 0x35, 0xcd, 0x66, 0x77, 0x73, 0x63, 0x35, 0x0e, 0x52, 0xc6, 0x22, 0x12, 0x16, 0x85, 0x7d,
	0xda, 0x82, 0xbe, 0x14, 0x1f, 0x04, 0x85, 0x52, 0x41, 0xbb, 0x95, 0xe0, 0x07, 0xf8, 0x36, 0x4d,
	0x2e, 0xdd, 0x61, 0x93, 0x99, 0x30, 0x99, 0xad, 0xec, 0xaf, 0xf0, 0x2f, 0xcb, 0xdc, 0x64, 0x9a,
	0xac, 0xf4, 0xed, 0x9e, 0x73, 0x66, 0xce, 0xcc, 0xbd, 0xf7, 0xc0, 0x8b, 0x8d, 0x68, 0xe4, 0x59,
	0xb1, 0x11, 0xc6, 0x9e, 0xd5, 0x68, 0x45, 0x29, 0xac, 0x58, 0x35, 0x46, 0x5b, 0xcd, 0xc0, 0x49,
	0x2b, 0x92, 0x16, 0x9f, 0x01, 0xae, 0x84, 0x54, 0x56, 0x48, 0x85, 0x86, 0x31, 0x98, 0x28, 0x51,
	0x23, 0x0f, 0xb2, 0x60, 0x19, 0xe7, 0x54, 0xb3, 0xe7, 0x10, 0x61, 0x2d, 0x64, 0xc5, 0x8f, 0x88,
	0xec, 0x00, 0x4b, 0x21, 0xdc, 0x99, 0x8a, 0x87, 0xc4, 0xb9, 0x72, 0xf1, 0x37, 0x82, 0xf9, 0x55,
	0xff, 0xd0, 0x83, 0x46, 0x0c, 0x26, 0x1b, 0x5d, 0x63, 0xef, 0x43, 0x35, 0xe3, 0x30, 0x6b, 0xf5,
	0xce, 0x14, 0xd8, 0xf2, 0x30, 0x0b, 0x97, 0x71, 0xee, 0xa1, 0x53, 0xee, 0xd0, 0xb4, 0x52, 0x2b,
	0x3e, 0xa1, 0x0b, 0x1e, 0xb2, 0x0c, 0x92, 0x12, 0xdb, 0xc2, 0xc8, 0xc6, 0x3a, 0x35, 0x22, 0x75,
	0x4c, 0xb1, 0x53, 0x98, 0x6f, 0x71, 0xff, 0x47, 0x9b, 0xb2, 0xe5, 0x53, 0xb2, 0xbd, 0xc7, 0xec,
	0x1c, 0x92, 0xfa, 0xbe, 0xe1, 0x96, 0xcf, 0xb2, 0x70, 0x99, 0xbc, 0x3d, 0x59, 0x0d, 0x23, 0x59,
	0x0d, 0xf3, 0xc8, 0xc7, 0x47, 0xd9, 0x09, 0x4c, 0x51, 0xdd, 0x4a, 0x85, 0x7c, 0x4e, 0x4f, 0xf6,
	0xc8, 0xf5, 0x25, 0x0b, 0xad, 0x78, 0xdc, 0xf5, 0xe5, 0x6a, 0xf6, 0x0a, 0x40, 0x34, 0xf2, 0x67,
	0xdf, 0x00, 0x90, 0x32, 0x62, 0xd8, 0x4b, 0x88, 0x0b, 0xad, 0x4a, 0x49, 0x1d, 0x24, 0x24, 0x0f,
	0x84, 0x73, 0xb4, 0xe2, 0xb6, 0xe5, 0x8f, 0x3b, 0x47, 0x57, 0x77, 0x8e, 0x8d, 0x77, 0x3c, 0xf6,
	0x8e, 0x9e, 0x71, 0x7a, 0x89, 0x8d, 0xc1, 0x42, 0x58, 0x2c, 0xf9, 0x93, 0x2c, 0x58, 0xce, 0xf3,
	0x11, 0xc3, 0x5e, 0xc3, 0xb1, 0x95, 0x55, 0x85, 0xc6, 0x5b, 0x3c, 0x25, 0x8b, 0x43, 0x92, 0x5d,
	0x42, 0x22, 0x94, 0xd2, 0x56, 0xb8, 0x7f, 0xb4, 0x3c, 0xa5, 0xe9, 0xbc, 0x39, 0x98, 0x8e, 0xcf,
	0xd2, 0xc7, 0xe1, 0xdc, 0x85, 0xb2, 0x66, 0x9f, 0x8f, 0x6f, 0xba, 0x25, 0x6d, 0x77, 0x37, 0xe8,
	0x1f, 0x7b, 0xd6, 0x2d, 0x69, 0x44, 0x9d, 0x7e, 0x80, 0xf4, 0x7f, 0x0b, 0x97, 0xaa, 0x2d, 0xee,
	0xfb, 0xd4, 0xb8, 0xd2, 0xa5, 0xef, 0x4e, 0x54, 0x3b, 0x9f, 0x9a, 0x0e, 0xbc, 0x3f, 0x3a, 0x0f,
	0x16, 0x19, 0x4c, 0x2f, 0xba, 0x05, 0x24, 0x30, 0xfb, 0xb1, 0xfe, 0xb2, 0xbe, 0xfe, 0xb5, 0x4e,
	0x1f, 0xb1, 0x18, 0xa2, 0xcb, 0xeb, 0xef, 0xdf, 0xbe, 0xa6, 0xc1, 0xa7, 0xd9, 0xef, 0x88, 0xfe,
	0x7c, 0x33, 0xa5, 0xdc, 0xbf, 0xfb, 0x17, 0x00, 0x00, 0xff, 0xff, 0x36, 0xf9, 0x0d, 0xa6, 0x14,
	0x03, 0x00, 0x00,
}
