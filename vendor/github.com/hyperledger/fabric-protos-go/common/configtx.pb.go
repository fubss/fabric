// Code generated by protoc-gen-go. DO NOT EDIT.
// source: common/configtx.proto

package common

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// ConfigEnvelope is designed to contain _all_ configuration for a chain with no dependency
// on previous configuration transactions.
//
// It is generated with the following scheme:
//  1. Retrieve the existing configuration
//  2. Note the config properties (ConfigValue, ConfigPolicy, ConfigGroup) to be modified
//  3. Add any intermediate ConfigGroups to the ConfigUpdate.read_set (sparsely)
//  4. Add any additional desired dependencies to ConfigUpdate.read_set (sparsely)
//  5. Modify the config properties, incrementing each version by 1, set them in the ConfigUpdate.write_set
//     Note: any element not modified but specified should already be in the read_set, so may be specified sparsely
//  6. Create ConfigUpdate message and marshal it into ConfigUpdateEnvelope.update and encode the required signatures
//     a) Each signature is of type ConfigSignature
//     b) The ConfigSignature signature is over the concatenation of signature_header and the ConfigUpdate bytes (which includes a ChainHeader)
//  5. Submit new Config for ordering in Envelope signed by submitter
//     a) The Envelope Payload has data set to the marshaled ConfigEnvelope
//     b) The Envelope Payload has a header of type Header.Type.CONFIG_UPDATE
//
// The configuration manager will verify:
//  1. All items in the read_set exist at the read versions
//  2. All items in the write_set at a different version than, or not in, the read_set have been appropriately signed according to their mod_policy
//  3. The new configuration satisfies the ConfigSchema
type ConfigEnvelope struct {
	Config               *Config   `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	LastUpdate           *Envelope `protobuf:"bytes,2,opt,name=last_update,json=lastUpdate,proto3" json:"last_update,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ConfigEnvelope) Reset()         { *m = ConfigEnvelope{} }
func (m *ConfigEnvelope) String() string { return proto.CompactTextString(m) }
func (*ConfigEnvelope) ProtoMessage()    {}
func (*ConfigEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{0}
}

func (m *ConfigEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigEnvelope.Unmarshal(m, b)
}
func (m *ConfigEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigEnvelope.Marshal(b, m, deterministic)
}
func (m *ConfigEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigEnvelope.Merge(m, src)
}
func (m *ConfigEnvelope) XXX_Size() int {
	return xxx_messageInfo_ConfigEnvelope.Size(m)
}
func (m *ConfigEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigEnvelope proto.InternalMessageInfo

func (m *ConfigEnvelope) GetConfig() *Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *ConfigEnvelope) GetLastUpdate() *Envelope {
	if m != nil {
		return m.LastUpdate
	}
	return nil
}

// Config represents the config for a particular channel
type Config struct {
	Sequence             uint64       `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	ChannelGroup         *ConfigGroup `protobuf:"bytes,2,opt,name=channel_group,json=channelGroup,proto3" json:"channel_group,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{1}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetSequence() uint64 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

func (m *Config) GetChannelGroup() *ConfigGroup {
	if m != nil {
		return m.ChannelGroup
	}
	return nil
}

type ConfigUpdateEnvelope struct {
	ConfigUpdate         []byte             `protobuf:"bytes,1,opt,name=config_update,json=configUpdate,proto3" json:"config_update,omitempty"`
	Signatures           []*ConfigSignature `protobuf:"bytes,2,rep,name=signatures,proto3" json:"signatures,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ConfigUpdateEnvelope) Reset()         { *m = ConfigUpdateEnvelope{} }
func (m *ConfigUpdateEnvelope) String() string { return proto.CompactTextString(m) }
func (*ConfigUpdateEnvelope) ProtoMessage()    {}
func (*ConfigUpdateEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{2}
}

func (m *ConfigUpdateEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigUpdateEnvelope.Unmarshal(m, b)
}
func (m *ConfigUpdateEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigUpdateEnvelope.Marshal(b, m, deterministic)
}
func (m *ConfigUpdateEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigUpdateEnvelope.Merge(m, src)
}
func (m *ConfigUpdateEnvelope) XXX_Size() int {
	return xxx_messageInfo_ConfigUpdateEnvelope.Size(m)
}
func (m *ConfigUpdateEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigUpdateEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigUpdateEnvelope proto.InternalMessageInfo

func (m *ConfigUpdateEnvelope) GetConfigUpdate() []byte {
	if m != nil {
		return m.ConfigUpdate
	}
	return nil
}

func (m *ConfigUpdateEnvelope) GetSignatures() []*ConfigSignature {
	if m != nil {
		return m.Signatures
	}
	return nil
}

// ConfigUpdate is used to submit a subset of config and to have the orderer apply to Config
// it is always submitted inside a ConfigUpdateEnvelope which allows the addition of signatures
// resulting in a new total configuration.  The update is applied as follows:
//  1. The versions from all of the elements in the read_set is verified against the versions in the existing config.
//     If there is a mismatch in the read versions, then the config update fails and is rejected.
//  2. Any elements in the write_set with the same version as the read_set are ignored.
//  3. The corresponding mod_policy for every remaining element in the write_set is collected.
//  4. Each policy is checked against the signatures from the ConfigUpdateEnvelope, any failing to verify are rejected
//  5. The write_set is applied to the Config and the ConfigGroupSchema verifies that the updates were legal
type ConfigUpdate struct {
	ChannelId            string            `protobuf:"bytes,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	ReadSet              *ConfigGroup      `protobuf:"bytes,2,opt,name=read_set,json=readSet,proto3" json:"read_set,omitempty"`
	WriteSet             *ConfigGroup      `protobuf:"bytes,3,opt,name=write_set,json=writeSet,proto3" json:"write_set,omitempty"`
	IsolatedData         map[string][]byte `protobuf:"bytes,5,rep,name=isolated_data,json=isolatedData,proto3" json:"isolated_data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ConfigUpdate) Reset()         { *m = ConfigUpdate{} }
func (m *ConfigUpdate) String() string { return proto.CompactTextString(m) }
func (*ConfigUpdate) ProtoMessage()    {}
func (*ConfigUpdate) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{3}
}

func (m *ConfigUpdate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigUpdate.Unmarshal(m, b)
}
func (m *ConfigUpdate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigUpdate.Marshal(b, m, deterministic)
}
func (m *ConfigUpdate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigUpdate.Merge(m, src)
}
func (m *ConfigUpdate) XXX_Size() int {
	return xxx_messageInfo_ConfigUpdate.Size(m)
}
func (m *ConfigUpdate) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigUpdate.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigUpdate proto.InternalMessageInfo

func (m *ConfigUpdate) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *ConfigUpdate) GetReadSet() *ConfigGroup {
	if m != nil {
		return m.ReadSet
	}
	return nil
}

func (m *ConfigUpdate) GetWriteSet() *ConfigGroup {
	if m != nil {
		return m.WriteSet
	}
	return nil
}

func (m *ConfigUpdate) GetIsolatedData() map[string][]byte {
	if m != nil {
		return m.IsolatedData
	}
	return nil
}

// ConfigGroup is the hierarchical data structure for holding config
type ConfigGroup struct {
	Version              uint64                   `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Groups               map[string]*ConfigGroup  `protobuf:"bytes,2,rep,name=groups,proto3" json:"groups,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Values               map[string]*ConfigValue  `protobuf:"bytes,3,rep,name=values,proto3" json:"values,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Policies             map[string]*ConfigPolicy `protobuf:"bytes,4,rep,name=policies,proto3" json:"policies,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ModPolicy            string                   `protobuf:"bytes,5,opt,name=mod_policy,json=modPolicy,proto3" json:"mod_policy,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ConfigGroup) Reset()         { *m = ConfigGroup{} }
func (m *ConfigGroup) String() string { return proto.CompactTextString(m) }
func (*ConfigGroup) ProtoMessage()    {}
func (*ConfigGroup) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{4}
}

func (m *ConfigGroup) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigGroup.Unmarshal(m, b)
}
func (m *ConfigGroup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigGroup.Marshal(b, m, deterministic)
}
func (m *ConfigGroup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigGroup.Merge(m, src)
}
func (m *ConfigGroup) XXX_Size() int {
	return xxx_messageInfo_ConfigGroup.Size(m)
}
func (m *ConfigGroup) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigGroup.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigGroup proto.InternalMessageInfo

func (m *ConfigGroup) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *ConfigGroup) GetGroups() map[string]*ConfigGroup {
	if m != nil {
		return m.Groups
	}
	return nil
}

func (m *ConfigGroup) GetValues() map[string]*ConfigValue {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *ConfigGroup) GetPolicies() map[string]*ConfigPolicy {
	if m != nil {
		return m.Policies
	}
	return nil
}

func (m *ConfigGroup) GetModPolicy() string {
	if m != nil {
		return m.ModPolicy
	}
	return ""
}

// ConfigValue represents an individual piece of config data
type ConfigValue struct {
	Version              uint64   `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	ModPolicy            string   `protobuf:"bytes,3,opt,name=mod_policy,json=modPolicy,proto3" json:"mod_policy,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfigValue) Reset()         { *m = ConfigValue{} }
func (m *ConfigValue) String() string { return proto.CompactTextString(m) }
func (*ConfigValue) ProtoMessage()    {}
func (*ConfigValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{5}
}

func (m *ConfigValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigValue.Unmarshal(m, b)
}
func (m *ConfigValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigValue.Marshal(b, m, deterministic)
}
func (m *ConfigValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigValue.Merge(m, src)
}
func (m *ConfigValue) XXX_Size() int {
	return xxx_messageInfo_ConfigValue.Size(m)
}
func (m *ConfigValue) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigValue.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigValue proto.InternalMessageInfo

func (m *ConfigValue) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *ConfigValue) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *ConfigValue) GetModPolicy() string {
	if m != nil {
		return m.ModPolicy
	}
	return ""
}

type ConfigPolicy struct {
	Version              uint64   `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Policy               *Policy  `protobuf:"bytes,2,opt,name=policy,proto3" json:"policy,omitempty"`
	ModPolicy            string   `protobuf:"bytes,3,opt,name=mod_policy,json=modPolicy,proto3" json:"mod_policy,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfigPolicy) Reset()         { *m = ConfigPolicy{} }
func (m *ConfigPolicy) String() string { return proto.CompactTextString(m) }
func (*ConfigPolicy) ProtoMessage()    {}
func (*ConfigPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{6}
}

func (m *ConfigPolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigPolicy.Unmarshal(m, b)
}
func (m *ConfigPolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigPolicy.Marshal(b, m, deterministic)
}
func (m *ConfigPolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigPolicy.Merge(m, src)
}
func (m *ConfigPolicy) XXX_Size() int {
	return xxx_messageInfo_ConfigPolicy.Size(m)
}
func (m *ConfigPolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigPolicy.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigPolicy proto.InternalMessageInfo

func (m *ConfigPolicy) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *ConfigPolicy) GetPolicy() *Policy {
	if m != nil {
		return m.Policy
	}
	return nil
}

func (m *ConfigPolicy) GetModPolicy() string {
	if m != nil {
		return m.ModPolicy
	}
	return ""
}

type ConfigSignature struct {
	SignatureHeader      []byte   `protobuf:"bytes,1,opt,name=signature_header,json=signatureHeader,proto3" json:"signature_header,omitempty"`
	Signature            []byte   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfigSignature) Reset()         { *m = ConfigSignature{} }
func (m *ConfigSignature) String() string { return proto.CompactTextString(m) }
func (*ConfigSignature) ProtoMessage()    {}
func (*ConfigSignature) Descriptor() ([]byte, []int) {
	return fileDescriptor_5190bbf196fa7499, []int{7}
}

func (m *ConfigSignature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigSignature.Unmarshal(m, b)
}
func (m *ConfigSignature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigSignature.Marshal(b, m, deterministic)
}
func (m *ConfigSignature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigSignature.Merge(m, src)
}
func (m *ConfigSignature) XXX_Size() int {
	return xxx_messageInfo_ConfigSignature.Size(m)
}
func (m *ConfigSignature) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigSignature.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigSignature proto.InternalMessageInfo

func (m *ConfigSignature) GetSignatureHeader() []byte {
	if m != nil {
		return m.SignatureHeader
	}
	return nil
}

func (m *ConfigSignature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*ConfigEnvelope)(nil), "common.ConfigEnvelope")
	proto.RegisterType((*Config)(nil), "common.Config")
	proto.RegisterType((*ConfigUpdateEnvelope)(nil), "common.ConfigUpdateEnvelope")
	proto.RegisterType((*ConfigUpdate)(nil), "common.ConfigUpdate")
	proto.RegisterMapType((map[string][]byte)(nil), "common.ConfigUpdate.IsolatedDataEntry")
	proto.RegisterType((*ConfigGroup)(nil), "common.ConfigGroup")
	proto.RegisterMapType((map[string]*ConfigGroup)(nil), "common.ConfigGroup.GroupsEntry")
	proto.RegisterMapType((map[string]*ConfigPolicy)(nil), "common.ConfigGroup.PoliciesEntry")
	proto.RegisterMapType((map[string]*ConfigValue)(nil), "common.ConfigGroup.ValuesEntry")
	proto.RegisterType((*ConfigValue)(nil), "common.ConfigValue")
	proto.RegisterType((*ConfigPolicy)(nil), "common.ConfigPolicy")
	proto.RegisterType((*ConfigSignature)(nil), "common.ConfigSignature")
}

func init() { proto.RegisterFile("common/configtx.proto", fileDescriptor_5190bbf196fa7499) }

var fileDescriptor_5190bbf196fa7499 = []byte{
	// 645 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x55, 0x61, 0x4f, 0xd4, 0x4c,
	0x10, 0x0e, 0xd7, 0x5e, 0xe9, 0xcd, 0xf5, 0xe0, 0xde, 0x85, 0x37, 0x36, 0x17, 0x8d, 0x58, 0x0d,
	0x01, 0x13, 0x8a, 0xe2, 0x07, 0x88, 0x89, 0x31, 0x51, 0x89, 0x82, 0x89, 0xd1, 0x12, 0xf9, 0x40,
	0x4c, 0x9a, 0xa5, 0x5d, 0x7a, 0x95, 0x5e, 0xb7, 0x6e, 0xb7, 0x68, 0x7f, 0x92, 0x7f, 0xcd, 0x5f,
	0x61, 0xba, 0xbb, 0x2d, 0x2d, 0x1e, 0x67, 0xfc, 0x02, 0xcc, 0xcc, 0xf3, 0x3c, 0x33, 0xcf, 0xce,
	0x76, 0x81, 0xff, 0x03, 0x3a, 0x9b, 0xd1, 0x74, 0x37, 0xa0, 0xe9, 0x45, 0x1c, 0xf1, 0x1f, 0x6e,
	0xc6, 0x28, 0xa7, 0xc8, 0x90, 0xe9, 0xc9, 0x5a, 0x53, 0xae, 0x7e, 0xc9, 0xe2, 0xa4, 0xe6, 0x64,
	0x34, 0x89, 0x83, 0x98, 0xe4, 0x32, 0xed, 0x5c, 0xc2, 0xca, 0x6b, 0xa1, 0x72, 0x98, 0x5e, 0x91,
	0x84, 0x66, 0x04, 0x6d, 0x82, 0x21, 0x75, 0xed, 0xa5, 0x8d, 0xa5, 0xad, 0xe1, 0xde, 0x8a, 0xab,
	0x74, 0x24, 0xce, 0x53, 0x55, 0xf4, 0x14, 0x86, 0x09, 0xce, 0xb9, 0x5f, 0x64, 0x21, 0xe6, 0xc4,
	0xee, 0x09, 0xf0, 0xb8, 0x06, 0xd7, 0x72, 0x1e, 0x54, 0xa0, 0xcf, 0x02, 0xe3, 0x7c, 0x05, 0x43,
	0x8a, 0xa0, 0x09, 0x98, 0x39, 0xf9, 0x56, 0x90, 0x34, 0x20, 0xa2, 0x8d, 0xee, 0x35, 0x31, 0x3a,
	0x80, 0x51, 0x30, 0xc5, 0x69, 0x4a, 0x12, 0x3f, 0x62, 0xb4, 0xc8, 0x94, 0xf4, 0x5a, 0x77, 0x8e,
	0xb7, 0x55, 0xc9, 0xb3, 0x14, 0x52, 0x44, 0xc7, 0xba, 0xa9, 0x8d, 0x75, 0x4f, 0xe7, 0x65, 0x46,
	0x1c, 0x0e, 0xeb, 0x12, 0x28, 0x7b, 0x37, 0xf6, 0x1e, 0xc2, 0x48, 0x1a, 0xa8, 0x07, 0xaf, 0xda,
	0x5b, 0x9e, 0x15, 0xb4, 0xc0, 0x68, 0x1f, 0x20, 0x8f, 0xa3, 0x14, 0xf3, 0x82, 0x91, 0xdc, 0xee,
	0x6d, 0x68, 0x5b, 0xc3, 0xbd, 0x3b, 0xdd, 0xfe, 0x27, 0x75, 0xdd, 0x6b, 0x41, 0x9d, 0x9f, 0x3d,
	0xb0, 0xda, 0x6d, 0xd1, 0x3d, 0x80, 0xda, 0x4c, 0x1c, 0x8a, 0x5e, 0x03, 0x6f, 0xa0, 0x32, 0x47,
	0x21, 0x72, 0xc1, 0x64, 0x04, 0x87, 0x7e, 0x4e, 0xf8, 0x22, 0x9b, 0xcb, 0x15, 0xe8, 0x84, 0x70,
	0xf4, 0x04, 0x06, 0xdf, 0x59, 0xcc, 0x89, 0x20, 0x68, 0xb7, 0x13, 0x4c, 0x81, 0xaa, 0x18, 0xef,
	0x61, 0x14, 0xe7, 0x34, 0xc1, 0x9c, 0x84, 0x7e, 0x88, 0x39, 0xb6, 0xfb, 0xc2, 0xcd, 0x66, 0x97,
	0x25, 0xa7, 0x75, 0x8f, 0x14, 0xf2, 0x0d, 0xe6, 0xf8, 0x30, 0xe5, 0xac, 0xf4, 0xac, 0xb8, 0x95,
	0x9a, 0xbc, 0x84, 0xff, 0xfe, 0x80, 0xa0, 0x31, 0x68, 0x97, 0xa4, 0x54, 0xde, 0xaa, 0x3f, 0xd1,
	0x3a, 0xf4, 0xaf, 0x70, 0x52, 0xc8, 0x4b, 0x61, 0x79, 0x32, 0x78, 0xde, 0x3b, 0x58, 0x3a, 0xd6,
	0x4d, 0x7d, 0xdc, 0x57, 0x1b, 0xfa, 0xa5, 0xc1, 0xb0, 0x35, 0x33, 0xb2, 0x61, 0xf9, 0x8a, 0xb0,
	0x3c, 0xa6, 0xa9, 0xba, 0x12, 0x75, 0x88, 0xf6, 0xc1, 0x10, 0x37, 0xa1, 0x5e, 0xc5, 0xfd, 0x39,
	0x96, 0x5d, 0xf1, 0x33, 0x97, 0x53, 0x2b, 0x78, 0x45, 0x14, 0xbd, 0x73, 0x5b, 0xbb, 0x9d, 0x78,
	0x2a, 0x10, 0x8a, 0x28, 0xe1, 0xe8, 0x05, 0x98, 0xf5, 0x87, 0x62, 0xeb, 0x82, 0xfa, 0x60, 0x1e,
	0xf5, 0xa3, 0xc2, 0x48, 0x72, 0x43, 0xa9, 0xb6, 0x3e, 0xa3, 0xa1, 0x2f, 0xe2, 0xd2, 0xee, 0xcb,
	0xad, 0xcf, 0x68, 0x28, 0xf0, 0xe5, 0xe4, 0x03, 0x0c, 0x5b, 0xd3, 0xce, 0x39, 0xc0, 0xed, 0xf6,
	0x01, 0xde, 0xb2, 0xe2, 0xeb, 0x53, 0xad, 0xf4, 0x5a, 0x26, 0xfe, 0x59, 0x4f, 0x70, 0xdb, 0x7a,
	0x9f, 0x60, 0xd4, 0x71, 0x36, 0x47, 0xf1, 0x71, 0x57, 0x71, 0xbd, 0xab, 0x28, 0x7d, 0xb6, 0x24,
	0x9d, 0x2f, 0xf5, 0xae, 0x45, 0xb3, 0x05, 0xbb, 0x9e, 0x7b, 0x77, 0x6e, 0x1c, 0xa8, 0x76, 0xe3,
	0x40, 0x1d, 0x5a, 0x7f, 0x75, 0x32, 0x5e, 0x20, 0xbf, 0x09, 0x86, 0x12, 0xe9, 0x75, 0x5f, 0x37,
	0x35, 0xb2, 0xaa, 0xfe, 0xad, 0xe1, 0x19, 0xac, 0xde, 0x78, 0x06, 0xd0, 0x36, 0x8c, 0x9b, 0x87,
	0xc0, 0x9f, 0x12, 0x1c, 0x12, 0xa6, 0xde, 0x96, 0xd5, 0x26, 0xff, 0x4e, 0xa4, 0xd1, 0x5d, 0x18,
	0x34, 0x29, 0xe5, 0xf3, 0x3a, 0xf1, 0xea, 0x14, 0x1e, 0x51, 0x16, 0xb9, 0xd3, 0x32, 0x23, 0x2c,
	0x21, 0x61, 0x44, 0x98, 0x7b, 0x81, 0xcf, 0x59, 0x1c, 0xc8, 0x27, 0x3b, 0x57, 0x13, 0x9f, 0xb9,
	0x51, 0xcc, 0xa7, 0xc5, 0x79, 0x15, 0xee, 0xb6, 0xc0, 0xbb, 0x12, 0xbc, 0x23, 0xc1, 0x3b, 0x11,
	0x55, 0xff, 0x07, 0xce, 0x0d, 0x91, 0x79, 0xf6, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x4e, 0x9b, 0x6e,
	0xa7, 0x3e, 0x06, 0x00, 0x00,
}
