// Code generated by protoc-gen-go. DO NOT EDIT.
// source: test_hello.proto

/*
Package __test_service is a generated protocol buffer package.

It is generated from these files:
	test_hello.proto

It has these top-level messages:
	HelloRequest
	HelloReply
*/
package test_hello_service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/any"
import _ "github.com/golang/protobuf/ptypes/empty"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type HelloReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *HelloReply) Reset()                    { *m = HelloReply{} }
func (m *HelloReply) String() string            { return proto.CompactTextString(m) }
func (*HelloReply) ProtoMessage()               {}
func (*HelloReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *HelloReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloRequest)(nil), "test_service.HelloRequest")
	proto.RegisterType((*HelloReply)(nil), "test_service.HelloReply")
}

func init() { proto.RegisterFile("test_hello.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 186 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x28, 0x49, 0x2d, 0x2e,
	0x89, 0xcf, 0x48, 0xcd, 0xc9, 0xc9, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x01, 0x8b,
	0x14, 0xa7, 0x16, 0x95, 0x65, 0x26, 0xa7, 0x4a, 0x49, 0xa6, 0xe7, 0xe7, 0xa7, 0xe7, 0xa4, 0xea,
	0x83, 0xe5, 0x92, 0x4a, 0xd3, 0xf4, 0x13, 0xf3, 0x2a, 0x21, 0x0a, 0xa5, 0xa4, 0xd1, 0xa5, 0x52,
	0x73, 0x0b, 0x4a, 0xa0, 0x92, 0x4a, 0x4a, 0x5c, 0x3c, 0x1e, 0x20, 0x43, 0x83, 0x52, 0x0b, 0x4b,
	0x53, 0x8b, 0x4b, 0x84, 0x84, 0xb8, 0x58, 0xf2, 0x12, 0x73, 0x53, 0x25, 0x18, 0x15, 0x18, 0x35,
	0x38, 0x83, 0xc0, 0x6c, 0x25, 0x35, 0x2e, 0x2e, 0xa8, 0x9a, 0x82, 0x9c, 0x4a, 0x21, 0x09, 0x2e,
	0xf6, 0xdc, 0xd4, 0xe2, 0xe2, 0xc4, 0x74, 0x98, 0x22, 0x18, 0xd7, 0xc8, 0x9f, 0x8b, 0x3b, 0x24,
	0xb5, 0xb8, 0x24, 0x18, 0xe2, 0x24, 0x21, 0x07, 0x2e, 0x8e, 0xe0, 0xc4, 0x4a, 0xb0, 0x4e, 0x21,
	0x29, 0x3d, 0x64, 0xd7, 0xea, 0x21, 0x5b, 0x29, 0x25, 0x81, 0x55, 0xae, 0x20, 0xa7, 0xd2, 0x49,
	0x20, 0x8a, 0x4f, 0xcf, 0x1a, 0x59, 0x32, 0x89, 0x0d, 0xec, 0x6a, 0x63, 0x40, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x12, 0x2b, 0xb6, 0xc3, 0x0f, 0x01, 0x00, 0x00,
}
