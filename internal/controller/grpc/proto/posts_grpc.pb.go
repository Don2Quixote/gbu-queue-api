// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

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

// PostsClient is the client API for Posts service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PostsClient interface {
	// Consumes events about new posts.
	Consume(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Posts_ConsumeClient, error)
}

type postsClient struct {
	cc grpc.ClientConnInterface
}

func NewPostsClient(cc grpc.ClientConnInterface) PostsClient {
	return &postsClient{cc}
}

func (c *postsClient) Consume(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Posts_ConsumeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Posts_ServiceDesc.Streams[0], "/posts.Posts/Consume", opts...)
	if err != nil {
		return nil, err
	}
	x := &postsConsumeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Posts_ConsumeClient interface {
	Recv() (*Post, error)
	grpc.ClientStream
}

type postsConsumeClient struct {
	grpc.ClientStream
}

func (x *postsConsumeClient) Recv() (*Post, error) {
	m := new(Post)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PostsServer is the server API for Posts service.
// All implementations must embed UnimplementedPostsServer
// for forward compatibility
type PostsServer interface {
	// Consumes events about new posts.
	Consume(*Empty, Posts_ConsumeServer) error
	mustEmbedUnimplementedPostsServer()
}

// UnimplementedPostsServer must be embedded to have forward compatible implementations.
type UnimplementedPostsServer struct {
}

func (UnimplementedPostsServer) Consume(*Empty, Posts_ConsumeServer) error {
	return status.Errorf(codes.Unimplemented, "method Consume not implemented")
}
func (UnimplementedPostsServer) mustEmbedUnimplementedPostsServer() {}

// UnsafePostsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PostsServer will
// result in compilation errors.
type UnsafePostsServer interface {
	mustEmbedUnimplementedPostsServer()
}

func RegisterPostsServer(s grpc.ServiceRegistrar, srv PostsServer) {
	s.RegisterService(&Posts_ServiceDesc, srv)
}

func _Posts_Consume_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PostsServer).Consume(m, &postsConsumeServer{stream})
}

type Posts_ConsumeServer interface {
	Send(*Post) error
	grpc.ServerStream
}

type postsConsumeServer struct {
	grpc.ServerStream
}

func (x *postsConsumeServer) Send(m *Post) error {
	return x.ServerStream.SendMsg(m)
}

// Posts_ServiceDesc is the grpc.ServiceDesc for Posts service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Posts_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "posts.Posts",
	HandlerType: (*PostsServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Consume",
			Handler:       _Posts_Consume_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "posts.proto",
}
