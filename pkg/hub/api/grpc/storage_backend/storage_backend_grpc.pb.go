// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: storage_backend.proto

package storage_backend

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StorageBackendClient is the client API for StorageBackend service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StorageBackendClient interface {
	// value
	GetValue(ctx context.Context, in *GetValueRequest, opts ...grpc.CallOption) (*GetValueResponse, error)
	OnCreate(ctx context.Context, in *OnCreateRequest, opts ...grpc.CallOption) (*OnCreateResponse, error)
	OnUpdate(ctx context.Context, in *OnUpdateRequest, opts ...grpc.CallOption) (*OnUpdateResponse, error)
	OnDelete(ctx context.Context, in *OnDeleteRequest, opts ...grpc.CallOption) (*OnDeleteResponse, error)
	// lock
	GetLockedBy(ctx context.Context, in *GetLockedByRequest, opts ...grpc.CallOption) (*GetLockedByResponse, error)
	OnLock(ctx context.Context, in *OnLockRequest, opts ...grpc.CallOption) (*OnLockResponse, error)
	OnUnlock(ctx context.Context, in *OnUnlockRequest, opts ...grpc.CallOption) (*OnUnlockResponse, error)
}

type storageBackendClient struct {
	cc grpc.ClientConnInterface
}

func NewStorageBackendClient(cc grpc.ClientConnInterface) StorageBackendClient {
	return &storageBackendClient{cc}
}

func (c *storageBackendClient) GetValue(ctx context.Context, in *GetValueRequest, opts ...grpc.CallOption) (*GetValueResponse, error) {
	out := new(GetValueResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/GetValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) OnCreate(ctx context.Context, in *OnCreateRequest, opts ...grpc.CallOption) (*OnCreateResponse, error) {
	out := new(OnCreateResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/OnCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) OnUpdate(ctx context.Context, in *OnUpdateRequest, opts ...grpc.CallOption) (*OnUpdateResponse, error) {
	out := new(OnUpdateResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/OnUpdate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) OnDelete(ctx context.Context, in *OnDeleteRequest, opts ...grpc.CallOption) (*OnDeleteResponse, error) {
	out := new(OnDeleteResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/OnDelete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) GetLockedBy(ctx context.Context, in *GetLockedByRequest, opts ...grpc.CallOption) (*GetLockedByResponse, error) {
	out := new(GetLockedByResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/GetLockedBy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) OnLock(ctx context.Context, in *OnLockRequest, opts ...grpc.CallOption) (*OnLockResponse, error) {
	out := new(OnLockResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/OnLock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageBackendClient) OnUnlock(ctx context.Context, in *OnUnlockRequest, opts ...grpc.CallOption) (*OnUnlockResponse, error) {
	out := new(OnUnlockResponse)
	err := c.cc.Invoke(ctx, "/storage_backend.StorageBackend/OnUnlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StorageBackendServer is the server API for StorageBackend service.
// All implementations must embed UnimplementedStorageBackendServer
// for forward compatibility
type StorageBackendServer interface {
	// value
	GetValue(context.Context, *GetValueRequest) (*GetValueResponse, error)
	OnCreate(context.Context, *OnCreateRequest) (*OnCreateResponse, error)
	OnUpdate(context.Context, *OnUpdateRequest) (*OnUpdateResponse, error)
	OnDelete(context.Context, *OnDeleteRequest) (*OnDeleteResponse, error)
	// lock
	GetLockedBy(context.Context, *GetLockedByRequest) (*GetLockedByResponse, error)
	OnLock(context.Context, *OnLockRequest) (*OnLockResponse, error)
	OnUnlock(context.Context, *OnUnlockRequest) (*OnUnlockResponse, error)
	mustEmbedUnimplementedStorageBackendServer()
}

// UnimplementedStorageBackendServer must be embedded to have forward compatible implementations.
type UnimplementedStorageBackendServer struct {
}

func (UnimplementedStorageBackendServer) GetValue(context.Context, *GetValueRequest) (*GetValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}
func (UnimplementedStorageBackendServer) OnCreate(context.Context, *OnCreateRequest) (*OnCreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnCreate not implemented")
}
func (UnimplementedStorageBackendServer) OnUpdate(context.Context, *OnUpdateRequest) (*OnUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnUpdate not implemented")
}
func (UnimplementedStorageBackendServer) OnDelete(context.Context, *OnDeleteRequest) (*OnDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnDelete not implemented")
}
func (UnimplementedStorageBackendServer) GetLockedBy(context.Context, *GetLockedByRequest) (*GetLockedByResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLockedBy not implemented")
}
func (UnimplementedStorageBackendServer) OnLock(context.Context, *OnLockRequest) (*OnLockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnLock not implemented")
}
func (UnimplementedStorageBackendServer) OnUnlock(context.Context, *OnUnlockRequest) (*OnUnlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnUnlock not implemented")
}
func (UnimplementedStorageBackendServer) mustEmbedUnimplementedStorageBackendServer() {}

// UnsafeStorageBackendServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StorageBackendServer will
// result in compilation errors.
type UnsafeStorageBackendServer interface {
	mustEmbedUnimplementedStorageBackendServer()
}

func RegisterStorageBackendServer(s grpc.ServiceRegistrar, srv StorageBackendServer) {
	s.RegisterService(&StorageBackend_ServiceDesc, srv)
}

func _StorageBackend_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/GetValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).GetValue(ctx, req.(*GetValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_OnCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).OnCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/OnCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).OnCreate(ctx, req.(*OnCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_OnUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).OnUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/OnUpdate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).OnUpdate(ctx, req.(*OnUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_OnDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).OnDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/OnDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).OnDelete(ctx, req.(*OnDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_GetLockedBy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLockedByRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).GetLockedBy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/GetLockedBy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).GetLockedBy(ctx, req.(*GetLockedByRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_OnLock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnLockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).OnLock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/OnLock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).OnLock(ctx, req.(*OnLockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageBackend_OnUnlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnUnlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageBackendServer).OnUnlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/storage_backend.StorageBackend/OnUnlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageBackendServer).OnUnlock(ctx, req.(*OnUnlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// StorageBackend_ServiceDesc is the grpc.ServiceDesc for StorageBackend service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StorageBackend_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "storage_backend.StorageBackend",
	HandlerType: (*StorageBackendServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetValue",
			Handler:    _StorageBackend_GetValue_Handler,
		},
		{
			MethodName: "OnCreate",
			Handler:    _StorageBackend_OnCreate_Handler,
		},
		{
			MethodName: "OnUpdate",
			Handler:    _StorageBackend_OnUpdate_Handler,
		},
		{
			MethodName: "OnDelete",
			Handler:    _StorageBackend_OnDelete_Handler,
		},
		{
			MethodName: "GetLockedBy",
			Handler:    _StorageBackend_GetLockedBy_Handler,
		},
		{
			MethodName: "OnLock",
			Handler:    _StorageBackend_OnLock_Handler,
		},
		{
			MethodName: "OnUnlock",
			Handler:    _StorageBackend_OnUnlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "storage_backend.proto",
}
