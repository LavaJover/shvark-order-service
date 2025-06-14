// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: order.proto

package orderpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	OrderService_CreateOrder_FullMethodName         = "/order.OrderService/CreateOrder"
	OrderService_ApproveOrder_FullMethodName        = "/order.OrderService/ApproveOrder"
	OrderService_CancelOrder_FullMethodName         = "/order.OrderService/CancelOrder"
	OrderService_OpenOrderDispute_FullMethodName    = "/order.OrderService/OpenOrderDispute"
	OrderService_ResolveOrderDispute_FullMethodName = "/order.OrderService/ResolveOrderDispute"
	OrderService_GetOrderByID_FullMethodName        = "/order.OrderService/GetOrderByID"
	OrderService_GetOrdersByTraderID_FullMethodName = "/order.OrderService/GetOrdersByTraderID"
)

// OrderServiceClient is the client API for OrderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrderServiceClient interface {
	CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderResponse, error)
	ApproveOrder(ctx context.Context, in *ApproveOrderRequest, opts ...grpc.CallOption) (*ApproveOrderResponse, error)
	CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderResponse, error)
	OpenOrderDispute(ctx context.Context, in *OpenOrderDisputeRequest, opts ...grpc.CallOption) (*OpenOrderDisputeResponse, error)
	ResolveOrderDispute(ctx context.Context, in *ResolveOrderDisputeRequest, opts ...grpc.CallOption) (*ResolveOrderDisputeResponse, error)
	GetOrderByID(ctx context.Context, in *GetOrderByIDRequest, opts ...grpc.CallOption) (*GetOrderByIDResponse, error)
	GetOrdersByTraderID(ctx context.Context, in *GetOrdersByTraderIDRequest, opts ...grpc.CallOption) (*GetOrdersByTraderIDResponse, error)
}

type orderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrderServiceClient(cc grpc.ClientConnInterface) OrderServiceClient {
	return &orderServiceClient{cc}
}

func (c *orderServiceClient) CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateOrderResponse)
	err := c.cc.Invoke(ctx, OrderService_CreateOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) ApproveOrder(ctx context.Context, in *ApproveOrderRequest, opts ...grpc.CallOption) (*ApproveOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ApproveOrderResponse)
	err := c.cc.Invoke(ctx, OrderService_ApproveOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelOrderResponse)
	err := c.cc.Invoke(ctx, OrderService_CancelOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) OpenOrderDispute(ctx context.Context, in *OpenOrderDisputeRequest, opts ...grpc.CallOption) (*OpenOrderDisputeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OpenOrderDisputeResponse)
	err := c.cc.Invoke(ctx, OrderService_OpenOrderDispute_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) ResolveOrderDispute(ctx context.Context, in *ResolveOrderDisputeRequest, opts ...grpc.CallOption) (*ResolveOrderDisputeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResolveOrderDisputeResponse)
	err := c.cc.Invoke(ctx, OrderService_ResolveOrderDispute_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) GetOrderByID(ctx context.Context, in *GetOrderByIDRequest, opts ...grpc.CallOption) (*GetOrderByIDResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetOrderByIDResponse)
	err := c.cc.Invoke(ctx, OrderService_GetOrderByID_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) GetOrdersByTraderID(ctx context.Context, in *GetOrdersByTraderIDRequest, opts ...grpc.CallOption) (*GetOrdersByTraderIDResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetOrdersByTraderIDResponse)
	err := c.cc.Invoke(ctx, OrderService_GetOrdersByTraderID_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrderServiceServer is the server API for OrderService service.
// All implementations must embed UnimplementedOrderServiceServer
// for forward compatibility.
type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
	ApproveOrder(context.Context, *ApproveOrderRequest) (*ApproveOrderResponse, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error)
	OpenOrderDispute(context.Context, *OpenOrderDisputeRequest) (*OpenOrderDisputeResponse, error)
	ResolveOrderDispute(context.Context, *ResolveOrderDisputeRequest) (*ResolveOrderDisputeResponse, error)
	GetOrderByID(context.Context, *GetOrderByIDRequest) (*GetOrderByIDResponse, error)
	GetOrdersByTraderID(context.Context, *GetOrdersByTraderIDRequest) (*GetOrdersByTraderIDResponse, error)
	mustEmbedUnimplementedOrderServiceServer()
}

// UnimplementedOrderServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOrderServiceServer struct{}

func (UnimplementedOrderServiceServer) CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrder not implemented")
}
func (UnimplementedOrderServiceServer) ApproveOrder(context.Context, *ApproveOrderRequest) (*ApproveOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveOrder not implemented")
}
func (UnimplementedOrderServiceServer) CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelOrder not implemented")
}
func (UnimplementedOrderServiceServer) OpenOrderDispute(context.Context, *OpenOrderDisputeRequest) (*OpenOrderDisputeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OpenOrderDispute not implemented")
}
func (UnimplementedOrderServiceServer) ResolveOrderDispute(context.Context, *ResolveOrderDisputeRequest) (*ResolveOrderDisputeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResolveOrderDispute not implemented")
}
func (UnimplementedOrderServiceServer) GetOrderByID(context.Context, *GetOrderByIDRequest) (*GetOrderByIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderByID not implemented")
}
func (UnimplementedOrderServiceServer) GetOrdersByTraderID(context.Context, *GetOrdersByTraderIDRequest) (*GetOrdersByTraderIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrdersByTraderID not implemented")
}
func (UnimplementedOrderServiceServer) mustEmbedUnimplementedOrderServiceServer() {}
func (UnimplementedOrderServiceServer) testEmbeddedByValue()                      {}

// UnsafeOrderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrderServiceServer will
// result in compilation errors.
type UnsafeOrderServiceServer interface {
	mustEmbedUnimplementedOrderServiceServer()
}

func RegisterOrderServiceServer(s grpc.ServiceRegistrar, srv OrderServiceServer) {
	// If the following call pancis, it indicates UnimplementedOrderServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OrderService_ServiceDesc, srv)
}

func _OrderService_CreateOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).CreateOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_CreateOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).CreateOrder(ctx, req.(*CreateOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_ApproveOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApproveOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).ApproveOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_ApproveOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).ApproveOrder(ctx, req.(*ApproveOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_CancelOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).CancelOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_CancelOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).CancelOrder(ctx, req.(*CancelOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_OpenOrderDispute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenOrderDisputeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).OpenOrderDispute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_OpenOrderDispute_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).OpenOrderDispute(ctx, req.(*OpenOrderDisputeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_ResolveOrderDispute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResolveOrderDisputeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).ResolveOrderDispute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_ResolveOrderDispute_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).ResolveOrderDispute(ctx, req.(*ResolveOrderDisputeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_GetOrderByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).GetOrderByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_GetOrderByID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).GetOrderByID(ctx, req.(*GetOrderByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_GetOrdersByTraderID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrdersByTraderIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).GetOrdersByTraderID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrderService_GetOrdersByTraderID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).GetOrdersByTraderID(ctx, req.(*GetOrdersByTraderIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OrderService_ServiceDesc is the grpc.ServiceDesc for OrderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "order.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler:    _OrderService_CreateOrder_Handler,
		},
		{
			MethodName: "ApproveOrder",
			Handler:    _OrderService_ApproveOrder_Handler,
		},
		{
			MethodName: "CancelOrder",
			Handler:    _OrderService_CancelOrder_Handler,
		},
		{
			MethodName: "OpenOrderDispute",
			Handler:    _OrderService_OpenOrderDispute_Handler,
		},
		{
			MethodName: "ResolveOrderDispute",
			Handler:    _OrderService_ResolveOrderDispute_Handler,
		},
		{
			MethodName: "GetOrderByID",
			Handler:    _OrderService_GetOrderByID_Handler,
		},
		{
			MethodName: "GetOrdersByTraderID",
			Handler:    _OrderService_GetOrdersByTraderID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "order.proto",
}
