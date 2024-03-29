// Code generated by protoc-gen-go. DO NOT EDIT.
// source: node.proto

/*
Package store is a generated protocol buffer package.

It is generated from these files:
	node.proto

It has these top-level messages:
	Node
	NodeList
*/
package store

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Node struct {
	ID       string   `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	Hash     string   `protobuf:"bytes,2,opt,name=Hash" json:"Hash,omitempty"`
	Alias    string   `protobuf:"bytes,3,opt,name=Alias" json:"Alias,omitempty"`
	Title    string   `protobuf:"bytes,4,opt,name=Title" json:"Title,omitempty"`
	Assigned string   `protobuf:"bytes,5,opt,name=Assigned" json:"Assigned,omitempty"`
	Body     string   `protobuf:"bytes,6,opt,name=Body" json:"Body,omitempty"`
	Tags     []string `protobuf:"bytes,7,rep,name=Tags" json:"Tags,omitempty"`
	Createdt string   `protobuf:"bytes,8,opt,name=Createdt" json:"Createdt,omitempty"`
	Updatedt string   `protobuf:"bytes,9,opt,name=Updatedt" json:"Updatedt,omitempty"`
}

func (m *Node) Reset()                    { *m = Node{} }
func (m *Node) String() string            { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()               {}
func (*Node) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Node) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Node) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *Node) GetAlias() string {
	if m != nil {
		return m.Alias
	}
	return ""
}

func (m *Node) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Node) GetAssigned() string {
	if m != nil {
		return m.Assigned
	}
	return ""
}

func (m *Node) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *Node) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *Node) GetCreatedt() string {
	if m != nil {
		return m.Createdt
	}
	return ""
}

func (m *Node) GetUpdatedt() string {
	if m != nil {
		return m.Updatedt
	}
	return ""
}

type NodeList struct {
	Items []*Node `protobuf:"bytes,1,rep,name=Items" json:"Items,omitempty"`
}

func (m *NodeList) Reset()                    { *m = NodeList{} }
func (m *NodeList) String() string            { return proto.CompactTextString(m) }
func (*NodeList) ProtoMessage()               {}
func (*NodeList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *NodeList) GetItems() []*Node {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
	proto.RegisterType((*Node)(nil), "store.Node")
	proto.RegisterType((*NodeList)(nil), "store.NodeList")
}

func init() { proto.RegisterFile("node.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 212 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x3c, 0x90, 0xb1, 0x4e, 0x87, 0x30,
	0x10, 0xc6, 0xc3, 0x1f, 0x8a, 0x70, 0x24, 0x0e, 0x8d, 0xc3, 0xc5, 0x09, 0x99, 0x58, 0x64, 0xd0,
	0x27, 0x40, 0x1d, 0x24, 0x31, 0x0e, 0x04, 0x1f, 0xa0, 0xa6, 0x0d, 0x36, 0x41, 0x4a, 0xb8, 0x2e,
	0xbe, 0xa6, 0x4f, 0x24, 0xd7, 0x06, 0xb7, 0xef, 0xfb, 0xfd, 0x92, 0x2f, 0xb9, 0x03, 0x58, 0x9d,
	0x36, 0xdd, 0xb6, 0x3b, 0xef, 0xa4, 0x20, 0xef, 0x76, 0xd3, 0xfc, 0x26, 0x90, 0xbd, 0x1f, 0x54,
	0x5e, 0xc3, 0x65, 0x78, 0xc1, 0xa4, 0x4e, 0xda, 0x72, 0x3c, 0x92, 0x94, 0x90, 0xbd, 0x2a, 0xfa,
	0xc2, 0x4b, 0x20, 0x21, 0xcb, 0x1b, 0x10, 0xfd, 0x62, 0x15, 0x61, 0x1a, 0x60, 0x2c, 0x4c, 0x27,
	0xeb, 0x17, 0x83, 0x59, 0xa4, 0xa1, 0xc8, 0x5b, 0x28, 0x7a, 0x22, 0x3b, 0xaf, 0x46, 0xa3, 0x08,
	0xe2, 0xbf, 0xf3, 0xf6, 0x93, 0xd3, 0x3f, 0x98, 0xc7, 0x6d, 0xce, 0xcc, 0x26, 0x35, 0x13, 0x5e,
	0xd5, 0x29, 0x33, 0xce, 0xbc, 0xf1, 0xbc, 0x1b, 0xe5, 0x8d, 0xf6, 0x58, 0xc4, 0x8d, 0xb3, 0xb3,
	0xfb, 0xd8, 0x74, 0x74, 0x65, 0x74, 0x67, 0x6f, 0xee, 0xa1, 0xe0, 0x9b, 0xde, 0x2c, 0x79, 0x79,
	0x07, 0x62, 0xf0, 0xe6, 0x9b, 0x8e, 0xd3, 0xd2, 0xb6, 0x7a, 0xa8, 0xba, 0x70, 0x77, 0xc7, 0x7e,
	0x8c, 0xe6, 0x33, 0x0f, 0x1f, 0x79, 0xfc, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x84, 0xa4, 0x4d, 0xc6,
	0x1f, 0x01, 0x00, 0x00,
}
