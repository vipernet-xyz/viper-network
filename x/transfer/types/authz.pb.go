// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ibc/applications/transfer/v1/authz.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/vipernet-xyz/viper-network/types"
	types "github.com/vipernet-xyz/viper-network/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Allocation defines the spend limit for a particular port and channel
type Allocation struct {
	// the port on which the packet will be sent
	SourcePort string `protobuf:"bytes,1,opt,name=source_port,json=sourcePort,proto3" json:"source_port,omitempty" yaml:"source_port"`
	// the channel by which the packet will be sent
	SourceChannel string `protobuf:"bytes,2,opt,name=source_channel,json=sourceChannel,proto3" json:"source_channel,omitempty" yaml:"source_channel"`
	// spend limitation on the channel
	SpendLimit github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,3,rep,name=spend_limit,json=spendLimit,proto3,castrepeated=github.com/vipernet-xyz/viper-network/types.Coins" json:"spend_limit"`
	// allow list of receivers, an empty allow list permits any receiver address
	AllowList []string `protobuf:"bytes,4,rep,name=allow_list,json=allowList,proto3" json:"allow_list,omitempty"`
}

func (m *Allocation) Reset()         { *m = Allocation{} }
func (m *Allocation) String() string { return proto.CompactTextString(m) }
func (*Allocation) ProtoMessage()    {}
func (*Allocation) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1a28b55d17325aa, []int{0}
}
func (m *Allocation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Allocation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Allocation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Allocation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Allocation.Merge(m, src)
}
func (m *Allocation) XXX_Size() int {
	return m.Size()
}
func (m *Allocation) XXX_DiscardUnknown() {
	xxx_messageInfo_Allocation.DiscardUnknown(m)
}

var xxx_messageInfo_Allocation proto.InternalMessageInfo

func (m *Allocation) GetSourcePort() string {
	if m != nil {
		return m.SourcePort
	}
	return ""
}

func (m *Allocation) GetSourceChannel() string {
	if m != nil {
		return m.SourceChannel
	}
	return ""
}

func (m *Allocation) GetSpendLimit() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.SpendLimit
	}
	return nil
}

func (m *Allocation) GetAllowList() []string {
	if m != nil {
		return m.AllowList
	}
	return nil
}

// TransferAuthorization allows the grantee to spend up to spend_limit coins from
// the granter's account for ibc transfer on a specific channel
type TransferAuthorization struct {
	// port and channel amounts
	Allocations []Allocation `protobuf:"bytes,1,rep,name=allocations,proto3" json:"allocations"`
}

func (m *TransferAuthorization) Reset()         { *m = TransferAuthorization{} }
func (m *TransferAuthorization) String() string { return proto.CompactTextString(m) }
func (*TransferAuthorization) ProtoMessage()    {}
func (*TransferAuthorization) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1a28b55d17325aa, []int{1}
}
func (m *TransferAuthorization) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TransferAuthorization) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TransferAuthorization.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TransferAuthorization) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransferAuthorization.Merge(m, src)
}
func (m *TransferAuthorization) XXX_Size() int {
	return m.Size()
}
func (m *TransferAuthorization) XXX_DiscardUnknown() {
	xxx_messageInfo_TransferAuthorization.DiscardUnknown(m)
}

var xxx_messageInfo_TransferAuthorization proto.InternalMessageInfo

func (m *TransferAuthorization) GetAllocations() []Allocation {
	if m != nil {
		return m.Allocations
	}
	return nil
}

func init() {
	proto.RegisterType((*Allocation)(nil), "ibc.applications.transfer.v1.Allocation")
	proto.RegisterType((*TransferAuthorization)(nil), "ibc.applications.transfer.v1.TransferAuthorization")
}

func init() {
	proto.RegisterFile("ibc/applications/transfer/v1/authz.proto", fileDescriptor_b1a28b55d17325aa)
}

var fileDescriptor_b1a28b55d17325aa = []byte{
	// 435 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0x4d, 0x6e, 0xd3, 0x40,
	0x14, 0x8e, 0x9b, 0x0a, 0x29, 0x13, 0xc1, 0xc2, 0xa2, 0xc8, 0xa9, 0xc0, 0x89, 0xbc, 0x40, 0xde,
	0x64, 0x86, 0xc0, 0x22, 0x52, 0x57, 0x34, 0xdd, 0x76, 0x51, 0x2c, 0x56, 0x6c, 0xa2, 0xf1, 0x64,
	0xb0, 0x47, 0x8c, 0xfd, 0x2c, 0xcf, 0xd8, 0xa8, 0x3d, 0x05, 0x48, 0x9c, 0x82, 0x35, 0x87, 0xa8,
	0x58, 0x75, 0xc9, 0x2a, 0xa0, 0xe4, 0x06, 0x3d, 0x01, 0xf2, 0xcc, 0x14, 0x5c, 0x21, 0xb1, 0xb2,
	0xdf, 0xcf, 0xf7, 0xde, 0xf7, 0xbe, 0xf9, 0x50, 0x2c, 0x52, 0x46, 0x68, 0x55, 0x49, 0xc1, 0xa8,
	0x16, 0x50, 0x2a, 0xa2, 0x6b, 0x5a, 0xaa, 0xf7, 0xbc, 0x26, 0xed, 0x82, 0xd0, 0x46, 0xe7, 0x57,
	0xb8, 0xaa, 0x41, 0x83, 0xff, 0x54, 0xa4, 0x0c, 0xf7, 0x3b, 0xf1, 0x5d, 0x27, 0x6e, 0x17, 0xc7,
	0x13, 0x06, 0xaa, 0x00, 0xb5, 0x36, 0xbd, 0xc4, 0x06, 0x16, 0x78, 0xfc, 0x38, 0x83, 0x0c, 0x6c,
	0xbe, 0xfb, 0x73, 0xd9, 0xd0, 0xf6, 0x90, 0x94, 0x2a, 0x4e, 0xda, 0x45, 0xca, 0x35, 0x5d, 0x10,
	0x06, 0xa2, 0xb4, 0xf5, 0xe8, 0xcb, 0x01, 0x42, 0xa7, 0x52, 0x82, 0x5d, 0xe6, 0x2f, 0xd1, 0x58,
	0x41, 0x53, 0x33, 0xbe, 0xae, 0xa0, 0xd6, 0x81, 0x37, 0xf3, 0xe2, 0xd1, 0xea, 0xc9, 0xed, 0x76,
	0xea, 0x5f, 0xd2, 0x42, 0x9e, 0x44, 0xbd, 0x62, 0x94, 0x20, 0x1b, 0x5d, 0x40, 0xad, 0xfd, 0xd7,
	0xe8, 0x91, 0xab, 0xb1, 0x9c, 0x96, 0x25, 0x97, 0xc1, 0x81, 0xc1, 0x4e, 0x6e, 0xb7, 0xd3, 0xa3,
	0x7b, 0x58, 0x57, 0x8f, 0x92, 0x87, 0x36, 0x71, 0x66, 0x63, 0x5f, 0xa2, 0xb1, 0xaa, 0x78, 0xb9,
	0x59, 0x4b, 0x51, 0x08, 0x1d, 0x0c, 0x67, 0xc3, 0x78, 0xfc, 0x72, 0x82, 0xdd, 0x8d, 0x1d, 0x7f,
	0xec, 0xf8, 0xe3, 0x33, 0x10, 0xe5, 0xea, 0xc5, 0xf5, 0x76, 0x3a, 0xf8, 0xfa, 0x73, 0x1a, 0x67,
	0x42, 0xe7, 0x4d, 0x8a, 0x19, 0x14, 0x4e, 0x10, 0xf7, 0x99, 0xab, 0xcd, 0x07, 0xa2, 0x2f, 0x2b,
	0xae, 0x0c, 0x40, 0x25, 0xc8, 0xcc, 0x3f, 0xef, 0xc6, 0xfb, 0xcf, 0x10, 0xa2, 0x52, 0xc2, 0xc7,
	0xb5, 0x14, 0x4a, 0x07, 0x87, 0xb3, 0x61, 0x3c, 0x4a, 0x46, 0x26, 0x73, 0x2e, 0x94, 0x8e, 0x3e,
	0x7b, 0xe8, 0xe8, 0xad, 0xd3, 0xfd, 0xb4, 0xd1, 0x39, 0xd4, 0xe2, 0xca, 0x2a, 0x74, 0x81, 0xc6,
	0xf4, 0x8f, 0x5e, 0x2a, 0xf0, 0x0c, 0xcd, 0x18, 0xff, 0xef, 0xd5, 0xf0, 0x5f, 0x81, 0x57, 0x87,
	0x1d, 0xeb, 0xa4, 0x3f, 0xe2, 0xe4, 0xf9, 0xf7, 0x6f, 0xf3, 0xc8, 0x9d, 0x69, 0x9d, 0x70, 0x77,
	0xe7, 0xbd, 0xcd, 0xab, 0x37, 0xd7, 0xbb, 0xd0, 0xbb, 0xd9, 0x85, 0xde, 0xaf, 0x5d, 0xe8, 0x7d,
	0xda, 0x87, 0x83, 0x9b, 0x7d, 0x38, 0xf8, 0xb1, 0x0f, 0x07, 0xef, 0x96, 0xff, 0x4a, 0x20, 0x52,
	0x36, 0xcf, 0x80, 0xb4, 0x4b, 0x52, 0xc0, 0xa6, 0x91, 0x5c, 0x75, 0xee, 0xeb, 0xb9, 0xce, 0xe8,
	0x92, 0x3e, 0x30, 0x26, 0x78, 0xf5, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x61, 0xe8, 0x65, 0x9c, 0x9f,
	0x02, 0x00, 0x00,
}

func (m *Allocation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Allocation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Allocation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AllowList) > 0 {
		for iNdEx := len(m.AllowList) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowList[iNdEx])
			copy(dAtA[i:], m.AllowList[iNdEx])
			i = encodeVarintAuthz(dAtA, i, uint64(len(m.AllowList[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.SpendLimit) > 0 {
		for iNdEx := len(m.SpendLimit) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SpendLimit[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintAuthz(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.SourceChannel) > 0 {
		i -= len(m.SourceChannel)
		copy(dAtA[i:], m.SourceChannel)
		i = encodeVarintAuthz(dAtA, i, uint64(len(m.SourceChannel)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SourcePort) > 0 {
		i -= len(m.SourcePort)
		copy(dAtA[i:], m.SourcePort)
		i = encodeVarintAuthz(dAtA, i, uint64(len(m.SourcePort)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *TransferAuthorization) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TransferAuthorization) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TransferAuthorization) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Allocations) > 0 {
		for iNdEx := len(m.Allocations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Allocations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintAuthz(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintAuthz(dAtA []byte, offset int, v uint64) int {
	offset -= sovAuthz(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Allocation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SourcePort)
	if l > 0 {
		n += 1 + l + sovAuthz(uint64(l))
	}
	l = len(m.SourceChannel)
	if l > 0 {
		n += 1 + l + sovAuthz(uint64(l))
	}
	if len(m.SpendLimit) > 0 {
		for _, e := range m.SpendLimit {
			l = e.Size()
			n += 1 + l + sovAuthz(uint64(l))
		}
	}
	if len(m.AllowList) > 0 {
		for _, s := range m.AllowList {
			l = len(s)
			n += 1 + l + sovAuthz(uint64(l))
		}
	}
	return n
}

func (m *TransferAuthorization) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Allocations) > 0 {
		for _, e := range m.Allocations {
			l = e.Size()
			n += 1 + l + sovAuthz(uint64(l))
		}
	}
	return n
}

func sovAuthz(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAuthz(x uint64) (n int) {
	return sovAuthz(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Allocation) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthz
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Allocation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Allocation: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourcePort", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuthz
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuthz
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourcePort = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceChannel", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuthz
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuthz
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourceChannel = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendLimit", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthz
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthz
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SpendLimit = append(m.SpendLimit, types.Coin{})
			if err := m.SpendLimit[len(m.SpendLimit)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowList", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuthz
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuthz
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowList = append(m.AllowList, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthz(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthz
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TransferAuthorization) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthz
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TransferAuthorization: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TransferAuthorization: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Allocations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthz
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthz
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Allocations = append(m.Allocations, Allocation{})
			if err := m.Allocations[len(m.Allocations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthz(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthz
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAuthz(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAuthz
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAuthz
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthAuthz
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAuthz
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAuthz
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAuthz        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAuthz          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAuthz = fmt.Errorf("proto: unexpected end of group")
)
