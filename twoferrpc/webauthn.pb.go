// Code generated by protoc-gen-go. DO NOT EDIT.
// source: webauthn.proto

package twoferrpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BeginRegisterRequest struct {
	User                 *UserInfo `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *BeginRegisterRequest) Reset()         { *m = BeginRegisterRequest{} }
func (m *BeginRegisterRequest) String() string { return proto.CompactTextString(m) }
func (*BeginRegisterRequest) ProtoMessage()    {}
func (*BeginRegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{0}
}
func (m *BeginRegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BeginRegisterRequest.Unmarshal(m, b)
}
func (m *BeginRegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BeginRegisterRequest.Marshal(b, m, deterministic)
}
func (dst *BeginRegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BeginRegisterRequest.Merge(dst, src)
}
func (m *BeginRegisterRequest) XXX_Size() int {
	return xxx_messageInfo_BeginRegisterRequest.Size(m)
}
func (m *BeginRegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BeginRegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BeginRegisterRequest proto.InternalMessageInfo

func (m *BeginRegisterRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type BeginRegisterResponse struct {
	PublicKey            string       `protobuf:"bytes,1,opt,name=publicKey" json:"publicKey,omitempty"`
	SessionData          *SessionData `protobuf:"bytes,2,opt,name=sessionData" json:"sessionData,omitempty"`
	User                 *UserInfo    `protobuf:"bytes,3,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *BeginRegisterResponse) Reset()         { *m = BeginRegisterResponse{} }
func (m *BeginRegisterResponse) String() string { return proto.CompactTextString(m) }
func (*BeginRegisterResponse) ProtoMessage()    {}
func (*BeginRegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{1}
}
func (m *BeginRegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BeginRegisterResponse.Unmarshal(m, b)
}
func (m *BeginRegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BeginRegisterResponse.Marshal(b, m, deterministic)
}
func (dst *BeginRegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BeginRegisterResponse.Merge(dst, src)
}
func (m *BeginRegisterResponse) XXX_Size() int {
	return xxx_messageInfo_BeginRegisterResponse.Size(m)
}
func (m *BeginRegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BeginRegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BeginRegisterResponse proto.InternalMessageInfo

func (m *BeginRegisterResponse) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *BeginRegisterResponse) GetSessionData() *SessionData {
	if m != nil {
		return m.SessionData
	}
	return nil
}

func (m *BeginRegisterResponse) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type FinishRegisterRequest struct {
	SessionData          *SessionData `protobuf:"bytes,1,opt,name=sessionData" json:"sessionData,omitempty"`
	User                 *UserInfo    `protobuf:"bytes,2,opt,name=user" json:"user,omitempty"`
	Blob                 string       `protobuf:"bytes,3,opt,name=blob" json:"blob,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *FinishRegisterRequest) Reset()         { *m = FinishRegisterRequest{} }
func (m *FinishRegisterRequest) String() string { return proto.CompactTextString(m) }
func (*FinishRegisterRequest) ProtoMessage()    {}
func (*FinishRegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{2}
}
func (m *FinishRegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinishRegisterRequest.Unmarshal(m, b)
}
func (m *FinishRegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinishRegisterRequest.Marshal(b, m, deterministic)
}
func (dst *FinishRegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinishRegisterRequest.Merge(dst, src)
}
func (m *FinishRegisterRequest) XXX_Size() int {
	return xxx_messageInfo_FinishRegisterRequest.Size(m)
}
func (m *FinishRegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FinishRegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FinishRegisterRequest proto.InternalMessageInfo

func (m *FinishRegisterRequest) GetSessionData() *SessionData {
	if m != nil {
		return m.SessionData
	}
	return nil
}

func (m *FinishRegisterRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *FinishRegisterRequest) GetBlob() string {
	if m != nil {
		return m.Blob
	}
	return ""
}

type FinishRegisterResponse struct {
	User                 *UserInfo `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *FinishRegisterResponse) Reset()         { *m = FinishRegisterResponse{} }
func (m *FinishRegisterResponse) String() string { return proto.CompactTextString(m) }
func (*FinishRegisterResponse) ProtoMessage()    {}
func (*FinishRegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{3}
}
func (m *FinishRegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinishRegisterResponse.Unmarshal(m, b)
}
func (m *FinishRegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinishRegisterResponse.Marshal(b, m, deterministic)
}
func (dst *FinishRegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinishRegisterResponse.Merge(dst, src)
}
func (m *FinishRegisterResponse) XXX_Size() int {
	return xxx_messageInfo_FinishRegisterResponse.Size(m)
}
func (m *FinishRegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FinishRegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FinishRegisterResponse proto.InternalMessageInfo

func (m *FinishRegisterResponse) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type SessionData struct {
	Challenge            string   `protobuf:"bytes,1,opt,name=challenge" json:"challenge,omitempty"`
	UserId               []byte   `protobuf:"bytes,2,opt,name=userId,proto3" json:"userId,omitempty"`
	AllowedCredentials   [][]byte `protobuf:"bytes,3,rep,name=AllowedCredentials,proto3" json:"AllowedCredentials,omitempty"`
	UserVerification     string   `protobuf:"bytes,4,opt,name=UserVerification" json:"UserVerification,omitempty"`
	Signature            string   `protobuf:"bytes,5,opt,name=Signature" json:"Signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionData) Reset()         { *m = SessionData{} }
func (m *SessionData) String() string { return proto.CompactTextString(m) }
func (*SessionData) ProtoMessage()    {}
func (*SessionData) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{4}
}
func (m *SessionData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionData.Unmarshal(m, b)
}
func (m *SessionData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionData.Marshal(b, m, deterministic)
}
func (dst *SessionData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionData.Merge(dst, src)
}
func (m *SessionData) XXX_Size() int {
	return xxx_messageInfo_SessionData.Size(m)
}
func (m *SessionData) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionData.DiscardUnknown(m)
}

var xxx_messageInfo_SessionData proto.InternalMessageInfo

func (m *SessionData) GetChallenge() string {
	if m != nil {
		return m.Challenge
	}
	return ""
}

func (m *SessionData) GetUserId() []byte {
	if m != nil {
		return m.UserId
	}
	return nil
}

func (m *SessionData) GetAllowedCredentials() [][]byte {
	if m != nil {
		return m.AllowedCredentials
	}
	return nil
}

func (m *SessionData) GetUserVerification() string {
	if m != nil {
		return m.UserVerification
	}
	return ""
}

func (m *SessionData) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

type BeginLoginRequest struct {
	User                 *UserInfo `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *BeginLoginRequest) Reset()         { *m = BeginLoginRequest{} }
func (m *BeginLoginRequest) String() string { return proto.CompactTextString(m) }
func (*BeginLoginRequest) ProtoMessage()    {}
func (*BeginLoginRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{5}
}
func (m *BeginLoginRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BeginLoginRequest.Unmarshal(m, b)
}
func (m *BeginLoginRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BeginLoginRequest.Marshal(b, m, deterministic)
}
func (dst *BeginLoginRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BeginLoginRequest.Merge(dst, src)
}
func (m *BeginLoginRequest) XXX_Size() int {
	return xxx_messageInfo_BeginLoginRequest.Size(m)
}
func (m *BeginLoginRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BeginLoginRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BeginLoginRequest proto.InternalMessageInfo

func (m *BeginLoginRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type BeginLoginResponse struct {
	PublicKey            string       `protobuf:"bytes,1,opt,name=publicKey" json:"publicKey,omitempty"`
	SessionData          *SessionData `protobuf:"bytes,2,opt,name=sessionData" json:"sessionData,omitempty"`
	User                 *UserInfo    `protobuf:"bytes,3,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *BeginLoginResponse) Reset()         { *m = BeginLoginResponse{} }
func (m *BeginLoginResponse) String() string { return proto.CompactTextString(m) }
func (*BeginLoginResponse) ProtoMessage()    {}
func (*BeginLoginResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{6}
}
func (m *BeginLoginResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BeginLoginResponse.Unmarshal(m, b)
}
func (m *BeginLoginResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BeginLoginResponse.Marshal(b, m, deterministic)
}
func (dst *BeginLoginResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BeginLoginResponse.Merge(dst, src)
}
func (m *BeginLoginResponse) XXX_Size() int {
	return xxx_messageInfo_BeginLoginResponse.Size(m)
}
func (m *BeginLoginResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BeginLoginResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BeginLoginResponse proto.InternalMessageInfo

func (m *BeginLoginResponse) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *BeginLoginResponse) GetSessionData() *SessionData {
	if m != nil {
		return m.SessionData
	}
	return nil
}

func (m *BeginLoginResponse) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type FinishLoginRequest struct {
	User                 *UserInfo    `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	Session              *SessionData `protobuf:"bytes,2,opt,name=session" json:"session,omitempty"`
	Blob                 string       `protobuf:"bytes,3,opt,name=blob" json:"blob,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *FinishLoginRequest) Reset()         { *m = FinishLoginRequest{} }
func (m *FinishLoginRequest) String() string { return proto.CompactTextString(m) }
func (*FinishLoginRequest) ProtoMessage()    {}
func (*FinishLoginRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{7}
}
func (m *FinishLoginRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinishLoginRequest.Unmarshal(m, b)
}
func (m *FinishLoginRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinishLoginRequest.Marshal(b, m, deterministic)
}
func (dst *FinishLoginRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinishLoginRequest.Merge(dst, src)
}
func (m *FinishLoginRequest) XXX_Size() int {
	return xxx_messageInfo_FinishLoginRequest.Size(m)
}
func (m *FinishLoginRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FinishLoginRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FinishLoginRequest proto.InternalMessageInfo

func (m *FinishLoginRequest) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *FinishLoginRequest) GetSession() *SessionData {
	if m != nil {
		return m.Session
	}
	return nil
}

func (m *FinishLoginRequest) GetBlob() string {
	if m != nil {
		return m.Blob
	}
	return ""
}

type FinishLoginResponse struct {
	User                 *UserInfo `protobuf:"bytes,1,opt,name=user" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *FinishLoginResponse) Reset()         { *m = FinishLoginResponse{} }
func (m *FinishLoginResponse) String() string { return proto.CompactTextString(m) }
func (*FinishLoginResponse) ProtoMessage()    {}
func (*FinishLoginResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{8}
}
func (m *FinishLoginResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinishLoginResponse.Unmarshal(m, b)
}
func (m *FinishLoginResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinishLoginResponse.Marshal(b, m, deterministic)
}
func (dst *FinishLoginResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinishLoginResponse.Merge(dst, src)
}
func (m *FinishLoginResponse) XXX_Size() int {
	return xxx_messageInfo_FinishLoginResponse.Size(m)
}
func (m *FinishLoginResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FinishLoginResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FinishLoginResponse proto.InternalMessageInfo

func (m *FinishLoginResponse) GetUser() *UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

type UserInfo struct {
	Id                   string               `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name                 string               `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	DisplayName          string               `protobuf:"bytes,3,opt,name=displayName" json:"displayName,omitempty"`
	AllowedCredentials   []*AllowedCredential `protobuf:"bytes,4,rep,name=allowedCredentials" json:"allowedCredentials,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *UserInfo) Reset()         { *m = UserInfo{} }
func (m *UserInfo) String() string { return proto.CompactTextString(m) }
func (*UserInfo) ProtoMessage()    {}
func (*UserInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{9}
}
func (m *UserInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserInfo.Unmarshal(m, b)
}
func (m *UserInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserInfo.Marshal(b, m, deterministic)
}
func (dst *UserInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserInfo.Merge(dst, src)
}
func (m *UserInfo) XXX_Size() int {
	return xxx_messageInfo_UserInfo.Size(m)
}
func (m *UserInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_UserInfo.DiscardUnknown(m)
}

var xxx_messageInfo_UserInfo proto.InternalMessageInfo

func (m *UserInfo) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UserInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserInfo) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

func (m *UserInfo) GetAllowedCredentials() []*AllowedCredential {
	if m != nil {
		return m.AllowedCredentials
	}
	return nil
}

type AuthenticatorAttestationResponse struct {
	ClientDataJSON       string   `protobuf:"bytes,1,opt,name=clientDataJSON" json:"clientDataJSON,omitempty"`
	AttestationObject    string   `protobuf:"bytes,2,opt,name=attestationObject" json:"attestationObject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AuthenticatorAttestationResponse) Reset()         { *m = AuthenticatorAttestationResponse{} }
func (m *AuthenticatorAttestationResponse) String() string { return proto.CompactTextString(m) }
func (*AuthenticatorAttestationResponse) ProtoMessage()    {}
func (*AuthenticatorAttestationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{10}
}
func (m *AuthenticatorAttestationResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthenticatorAttestationResponse.Unmarshal(m, b)
}
func (m *AuthenticatorAttestationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthenticatorAttestationResponse.Marshal(b, m, deterministic)
}
func (dst *AuthenticatorAttestationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthenticatorAttestationResponse.Merge(dst, src)
}
func (m *AuthenticatorAttestationResponse) XXX_Size() int {
	return xxx_messageInfo_AuthenticatorAttestationResponse.Size(m)
}
func (m *AuthenticatorAttestationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthenticatorAttestationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AuthenticatorAttestationResponse proto.InternalMessageInfo

func (m *AuthenticatorAttestationResponse) GetClientDataJSON() string {
	if m != nil {
		return m.ClientDataJSON
	}
	return ""
}

func (m *AuthenticatorAttestationResponse) GetAttestationObject() string {
	if m != nil {
		return m.AttestationObject
	}
	return ""
}

type AllowedCredential struct {
	ID                   []byte         `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	PublicKey            []byte         `protobuf:"bytes,2,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	AttestationType      string         `protobuf:"bytes,3,opt,name=AttestationType" json:"AttestationType,omitempty"`
	Authenticator        *Authenticator `protobuf:"bytes,4,opt,name=Authenticator" json:"Authenticator,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *AllowedCredential) Reset()         { *m = AllowedCredential{} }
func (m *AllowedCredential) String() string { return proto.CompactTextString(m) }
func (*AllowedCredential) ProtoMessage()    {}
func (*AllowedCredential) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{11}
}
func (m *AllowedCredential) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AllowedCredential.Unmarshal(m, b)
}
func (m *AllowedCredential) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AllowedCredential.Marshal(b, m, deterministic)
}
func (dst *AllowedCredential) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AllowedCredential.Merge(dst, src)
}
func (m *AllowedCredential) XXX_Size() int {
	return xxx_messageInfo_AllowedCredential.Size(m)
}
func (m *AllowedCredential) XXX_DiscardUnknown() {
	xxx_messageInfo_AllowedCredential.DiscardUnknown(m)
}

var xxx_messageInfo_AllowedCredential proto.InternalMessageInfo

func (m *AllowedCredential) GetID() []byte {
	if m != nil {
		return m.ID
	}
	return nil
}

func (m *AllowedCredential) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *AllowedCredential) GetAttestationType() string {
	if m != nil {
		return m.AttestationType
	}
	return ""
}

func (m *AllowedCredential) GetAuthenticator() *Authenticator {
	if m != nil {
		return m.Authenticator
	}
	return nil
}

type Authenticator struct {
	AAGUID               []byte   `protobuf:"bytes,1,opt,name=AAGUID,proto3" json:"AAGUID,omitempty"`
	SignCount            uint32   `protobuf:"varint,2,opt,name=SignCount" json:"SignCount,omitempty"`
	CloneWarning         bool     `protobuf:"varint,3,opt,name=CloneWarning" json:"CloneWarning,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Authenticator) Reset()         { *m = Authenticator{} }
func (m *Authenticator) String() string { return proto.CompactTextString(m) }
func (*Authenticator) ProtoMessage()    {}
func (*Authenticator) Descriptor() ([]byte, []int) {
	return fileDescriptor_webauthn_a2d53817f25388e0, []int{12}
}
func (m *Authenticator) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Authenticator.Unmarshal(m, b)
}
func (m *Authenticator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Authenticator.Marshal(b, m, deterministic)
}
func (dst *Authenticator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Authenticator.Merge(dst, src)
}
func (m *Authenticator) XXX_Size() int {
	return xxx_messageInfo_Authenticator.Size(m)
}
func (m *Authenticator) XXX_DiscardUnknown() {
	xxx_messageInfo_Authenticator.DiscardUnknown(m)
}

var xxx_messageInfo_Authenticator proto.InternalMessageInfo

func (m *Authenticator) GetAAGUID() []byte {
	if m != nil {
		return m.AAGUID
	}
	return nil
}

func (m *Authenticator) GetSignCount() uint32 {
	if m != nil {
		return m.SignCount
	}
	return 0
}

func (m *Authenticator) GetCloneWarning() bool {
	if m != nil {
		return m.CloneWarning
	}
	return false
}

func init() {
	proto.RegisterType((*BeginRegisterRequest)(nil), "twoferrpc.BeginRegisterRequest")
	proto.RegisterType((*BeginRegisterResponse)(nil), "twoferrpc.BeginRegisterResponse")
	proto.RegisterType((*FinishRegisterRequest)(nil), "twoferrpc.FinishRegisterRequest")
	proto.RegisterType((*FinishRegisterResponse)(nil), "twoferrpc.FinishRegisterResponse")
	proto.RegisterType((*SessionData)(nil), "twoferrpc.SessionData")
	proto.RegisterType((*BeginLoginRequest)(nil), "twoferrpc.BeginLoginRequest")
	proto.RegisterType((*BeginLoginResponse)(nil), "twoferrpc.BeginLoginResponse")
	proto.RegisterType((*FinishLoginRequest)(nil), "twoferrpc.FinishLoginRequest")
	proto.RegisterType((*FinishLoginResponse)(nil), "twoferrpc.FinishLoginResponse")
	proto.RegisterType((*UserInfo)(nil), "twoferrpc.UserInfo")
	proto.RegisterType((*AuthenticatorAttestationResponse)(nil), "twoferrpc.AuthenticatorAttestationResponse")
	proto.RegisterType((*AllowedCredential)(nil), "twoferrpc.AllowedCredential")
	proto.RegisterType((*Authenticator)(nil), "twoferrpc.Authenticator")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Webauthn service

type WebauthnClient interface {
	BeginRegister(ctx context.Context, in *BeginRegisterRequest, opts ...grpc.CallOption) (*BeginRegisterResponse, error)
	FinishRegister(ctx context.Context, in *FinishRegisterRequest, opts ...grpc.CallOption) (*FinishRegisterResponse, error)
	BeginLogin(ctx context.Context, in *BeginLoginRequest, opts ...grpc.CallOption) (*BeginLoginResponse, error)
	FinishLogin(ctx context.Context, in *FinishLoginRequest, opts ...grpc.CallOption) (*FinishLoginResponse, error)
}

type webauthnClient struct {
	cc *grpc.ClientConn
}

func NewWebauthnClient(cc *grpc.ClientConn) WebauthnClient {
	return &webauthnClient{cc}
}

func (c *webauthnClient) BeginRegister(ctx context.Context, in *BeginRegisterRequest, opts ...grpc.CallOption) (*BeginRegisterResponse, error) {
	out := new(BeginRegisterResponse)
	err := grpc.Invoke(ctx, "/twoferrpc.Webauthn/BeginRegister", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webauthnClient) FinishRegister(ctx context.Context, in *FinishRegisterRequest, opts ...grpc.CallOption) (*FinishRegisterResponse, error) {
	out := new(FinishRegisterResponse)
	err := grpc.Invoke(ctx, "/twoferrpc.Webauthn/FinishRegister", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webauthnClient) BeginLogin(ctx context.Context, in *BeginLoginRequest, opts ...grpc.CallOption) (*BeginLoginResponse, error) {
	out := new(BeginLoginResponse)
	err := grpc.Invoke(ctx, "/twoferrpc.Webauthn/BeginLogin", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webauthnClient) FinishLogin(ctx context.Context, in *FinishLoginRequest, opts ...grpc.CallOption) (*FinishLoginResponse, error) {
	out := new(FinishLoginResponse)
	err := grpc.Invoke(ctx, "/twoferrpc.Webauthn/FinishLogin", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Webauthn service

type WebauthnServer interface {
	BeginRegister(context.Context, *BeginRegisterRequest) (*BeginRegisterResponse, error)
	FinishRegister(context.Context, *FinishRegisterRequest) (*FinishRegisterResponse, error)
	BeginLogin(context.Context, *BeginLoginRequest) (*BeginLoginResponse, error)
	FinishLogin(context.Context, *FinishLoginRequest) (*FinishLoginResponse, error)
}

func RegisterWebauthnServer(s *grpc.Server, srv WebauthnServer) {
	s.RegisterService(&_Webauthn_serviceDesc, srv)
}

func _Webauthn_BeginRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BeginRegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebauthnServer).BeginRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/twoferrpc.Webauthn/BeginRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebauthnServer).BeginRegister(ctx, req.(*BeginRegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Webauthn_FinishRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FinishRegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebauthnServer).FinishRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/twoferrpc.Webauthn/FinishRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebauthnServer).FinishRegister(ctx, req.(*FinishRegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Webauthn_BeginLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BeginLoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebauthnServer).BeginLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/twoferrpc.Webauthn/BeginLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebauthnServer).BeginLogin(ctx, req.(*BeginLoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Webauthn_FinishLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FinishLoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebauthnServer).FinishLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/twoferrpc.Webauthn/FinishLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebauthnServer).FinishLogin(ctx, req.(*FinishLoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Webauthn_serviceDesc = grpc.ServiceDesc{
	ServiceName: "twoferrpc.Webauthn",
	HandlerType: (*WebauthnServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BeginRegister",
			Handler:    _Webauthn_BeginRegister_Handler,
		},
		{
			MethodName: "FinishRegister",
			Handler:    _Webauthn_FinishRegister_Handler,
		},
		{
			MethodName: "BeginLogin",
			Handler:    _Webauthn_BeginLogin_Handler,
		},
		{
			MethodName: "FinishLogin",
			Handler:    _Webauthn_FinishLogin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "webauthn.proto",
}

func init() { proto.RegisterFile("webauthn.proto", fileDescriptor_webauthn_a2d53817f25388e0) }

var fileDescriptor_webauthn_a2d53817f25388e0 = []byte{
	// 668 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x55, 0xcf, 0x6e, 0xd3, 0x4e,
	0x10, 0xfe, 0x39, 0xc9, 0xaf, 0x34, 0xe3, 0x34, 0xd0, 0x29, 0xad, 0xa2, 0xaa, 0x85, 0xe0, 0x03,
	0x44, 0x08, 0x45, 0x28, 0x5c, 0x38, 0xa0, 0xa2, 0xd0, 0x0a, 0x14, 0x5a, 0xb5, 0x68, 0xdb, 0xd2,
	0xf3, 0xc6, 0xd9, 0x26, 0x8b, 0xcc, 0x6e, 0xf0, 0xae, 0x55, 0x7a, 0xe6, 0xce, 0x0d, 0x89, 0x13,
	0x2f, 0xc0, 0x33, 0xf0, 0x1a, 0x3c, 0x0f, 0xf2, 0xc6, 0x71, 0xd6, 0x76, 0x4a, 0x09, 0x27, 0x6e,
	0xde, 0x6f, 0xfe, 0x7d, 0x33, 0xfe, 0x66, 0x17, 0xea, 0x17, 0xac, 0x4f, 0x23, 0x3d, 0x12, 0xed,
	0x71, 0x28, 0xb5, 0xc4, 0xaa, 0xbe, 0x90, 0xe7, 0x2c, 0x0c, 0xc7, 0xbe, 0xf7, 0x1c, 0x6e, 0xbf,
	0x60, 0x43, 0x2e, 0x08, 0x1b, 0x72, 0xa5, 0x59, 0x48, 0xd8, 0x87, 0x88, 0x29, 0x8d, 0x0f, 0xa0,
	0x12, 0x29, 0x16, 0x36, 0x9c, 0xa6, 0xd3, 0x72, 0x3b, 0x6b, 0xed, 0x34, 0xa2, 0x7d, 0xaa, 0x58,
	0xd8, 0x13, 0xe7, 0x92, 0x18, 0x07, 0xef, 0xab, 0x03, 0xeb, 0xb9, 0x0c, 0x6a, 0x2c, 0x85, 0x62,
	0xb8, 0x05, 0xd5, 0x71, 0xd4, 0x0f, 0xb8, 0xbf, 0xcf, 0x2e, 0x4d, 0x9e, 0x2a, 0x99, 0x01, 0xf8,
	0x14, 0x5c, 0xc5, 0x94, 0xe2, 0x52, 0xec, 0x51, 0x4d, 0x1b, 0x25, 0x53, 0x67, 0xc3, 0xaa, 0x73,
	0x3c, 0xb3, 0x12, 0xdb, 0x35, 0xa5, 0x56, 0xbe, 0x8e, 0xda, 0x67, 0x07, 0xd6, 0x5f, 0x72, 0xc1,
	0xd5, 0x28, 0xdf, 0x5d, 0xae, 0xb8, 0xb3, 0x78, 0xf1, 0xd2, 0x35, 0xc5, 0x11, 0xa1, 0xd2, 0x0f,
	0x64, 0xdf, 0xb0, 0xac, 0x12, 0xf3, 0xed, 0x75, 0x61, 0x23, 0xcf, 0x27, 0x99, 0xd5, 0x1f, 0x8f,
	0xfb, 0x87, 0x03, 0xae, 0x45, 0x2e, 0x1e, 0xb2, 0x3f, 0xa2, 0x41, 0xc0, 0xc4, 0x90, 0x4d, 0x87,
	0x9c, 0x02, 0xb8, 0x01, 0x4b, 0x71, 0x54, 0x6f, 0x60, 0xf8, 0xd6, 0x48, 0x72, 0xc2, 0x36, 0x60,
	0x37, 0x08, 0xe4, 0x05, 0x1b, 0xec, 0x86, 0x6c, 0xc0, 0x84, 0xe6, 0x34, 0x50, 0x8d, 0x72, 0xb3,
	0xdc, 0xaa, 0x91, 0x39, 0x16, 0x7c, 0x08, 0xb7, 0x62, 0x1e, 0x6f, 0x59, 0xc8, 0xcf, 0xb9, 0x4f,
	0x35, 0x97, 0xa2, 0x51, 0x31, 0xc5, 0x0a, 0x78, 0xcc, 0xe8, 0x98, 0x0f, 0x05, 0xd5, 0x51, 0xc8,
	0x1a, 0xff, 0x4f, 0x18, 0xa5, 0x80, 0xf7, 0x0c, 0x56, 0x8d, 0x5a, 0x0e, 0xa4, 0x91, 0xcc, 0x82,
	0x62, 0xfb, 0xe2, 0x00, 0xda, 0xe1, 0xff, 0x8a, 0xd2, 0x3e, 0x39, 0x80, 0x93, 0x3f, 0xfb, 0x57,
	0x7d, 0xe1, 0x63, 0xb8, 0x91, 0xd4, 0xbd, 0x86, 0xde, 0xd4, 0x6d, 0xae, 0xbc, 0x76, 0x60, 0x2d,
	0x43, 0x62, 0x51, 0x6d, 0x7d, 0x73, 0x60, 0x79, 0x0a, 0x61, 0x1d, 0x4a, 0x7c, 0x90, 0x0c, 0xb3,
	0xc4, 0x07, 0x71, 0x41, 0x41, 0xdf, 0x33, 0xc3, 0xaf, 0x4a, 0xcc, 0x37, 0x36, 0xc1, 0x1d, 0x70,
	0x35, 0x0e, 0xe8, 0xe5, 0x61, 0x6c, 0x9a, 0x70, 0xb1, 0x21, 0x3c, 0x00, 0xa4, 0x45, 0xa1, 0x55,
	0x9a, 0xe5, 0x96, 0xdb, 0xd9, 0xb2, 0x98, 0x14, 0x34, 0x47, 0xe6, 0xc4, 0x79, 0x1f, 0xa1, 0xd9,
	0x8d, 0xf4, 0x28, 0x3e, 0xfa, 0x54, 0xcb, 0xb0, 0xab, 0x35, 0x53, 0xda, 0xc8, 0x2e, 0xed, 0xf6,
	0x3e, 0xd4, 0xfd, 0x80, 0x33, 0xa1, 0xe3, 0x79, 0xbd, 0x3e, 0x3e, 0x3a, 0x4c, 0x7a, 0xc8, 0xa1,
	0xf8, 0x08, 0x56, 0xe9, 0x2c, 0xfc, 0xa8, 0xff, 0x8e, 0xf9, 0x3a, 0x69, 0xae, 0x68, 0xf0, 0xbe,
	0x3b, 0xb0, 0x5a, 0xe0, 0x18, 0xcf, 0xa8, 0xb7, 0x67, 0xf2, 0xd7, 0x48, 0xa9, 0xb7, 0x17, 0xeb,
	0xf0, 0x4d, 0xaa, 0xc3, 0xc9, 0xc6, 0xcd, 0x00, 0x6c, 0xc1, 0x4d, 0x8b, 0xf0, 0xc9, 0xe5, 0x78,
	0x3a, 0xb1, 0x3c, 0x8c, 0x3b, 0xb0, 0x92, 0xe9, 0xd3, 0xec, 0x9a, 0xdb, 0x69, 0xd8, 0x03, 0xb3,
	0xed, 0x24, 0xeb, 0xee, 0xf1, 0x5c, 0x7c, 0x7c, 0x0f, 0x74, 0xbb, 0xaf, 0x4e, 0x53, 0xb2, 0xc9,
	0x69, 0xba, 0xab, 0xbb, 0x32, 0x12, 0x93, 0xe6, 0x57, 0xc8, 0x0c, 0x40, 0x0f, 0x6a, 0xbb, 0x81,
	0x14, 0xec, 0x8c, 0x86, 0x82, 0x8b, 0xa1, 0x61, 0xbb, 0x4c, 0x32, 0x58, 0xe7, 0x67, 0x09, 0x96,
	0xcf, 0x92, 0xd7, 0x05, 0x4f, 0x60, 0x25, 0xf3, 0x14, 0xe0, 0x5d, 0x8b, 0xf1, 0xbc, 0x67, 0x66,
	0xb3, 0x79, 0xb5, 0xc3, 0xe4, 0x7f, 0x7a, 0xff, 0xe1, 0x19, 0xd4, 0xb3, 0xb7, 0x26, 0xda, 0x51,
	0x73, 0x2f, 0xf8, 0xcd, 0x7b, 0xbf, 0xf1, 0x48, 0x13, 0xef, 0x03, 0xcc, 0x2e, 0x13, 0xdc, 0xca,
	0x53, 0xb1, 0x57, 0x79, 0x73, 0xfb, 0x0a, 0x6b, 0x9a, 0xec, 0x10, 0x5c, 0x6b, 0xf9, 0x70, 0xbb,
	0x40, 0x20, 0x93, 0xee, 0xce, 0x55, 0xe6, 0x69, 0xbe, 0xfe, 0x92, 0x79, 0xaa, 0x9f, 0xfc, 0x0a,
	0x00, 0x00, 0xff, 0xff, 0x1e, 0xb8, 0xa9, 0x0b, 0xbc, 0x07, 0x00, 0x00,
}