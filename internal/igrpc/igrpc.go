package igrpc

import (
	context "context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteServiceServer struct {
	UnimplementedRemoteServiceServer
}

func (s *remoteServiceServer) CallImageProcessing(
	ctx context.Context, req *RequestImageProcessing) (*ResponseImageProcessing, error) {
	res := &ResponseImageProcessing{
		ReqMessageId: req.MessageId,
		Success:      true,
		SrcWidth:     1280,
		SrcHeight:    720,
		DstWidth:     640,
		DstHeight:    480,
		Payload:      req.Payload,
	}
	return res, nil
}

// ---------------------------------------------------------

type Server struct {
	port     string
	wg       *sync.WaitGroup
	gsrv     *grpc.Server
	listener net.Listener
	closed   bool
}

func NewServer(port string) (*Server, error) {
	gsrv := grpc.NewServer()
	RegisterRemoteServiceServer(gsrv, &remoteServiceServer{})

	s := &Server{
		port: port,
		gsrv: gsrv,
		wg:   &sync.WaitGroup{},
	}

	return s, nil
}

func (s *Server) StartListenAndServe() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("gRPC server start on port %s", s.port)
		if err := s.gsrv.Serve(s.listener); err != nil {
			log.Printf("Failed to serve gRPC server: %v", err)
		}
	}()

	return nil
}

func (s *Server) Close() {
	if s.closed {
		return
	}
	s.gsrv.GracefulStop()
	s.wg.Wait()
	log.Printf("gRPC server stopped")
	s.closed = true
}

// ---------------------------------------------------------

type Client struct {
	address string
	gconn   *grpc.ClientConn
	gclient RemoteServiceClient
	closed  bool
}

func NewClient(address string) (*Client, error) {

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	gclient := NewRemoteServiceClient(conn)

	c := &Client{
		address: address,
		gclient: gclient,
		gconn:   conn,
	}

	return c, nil
}

func (c *Client) Close() {
	if c.closed {
		return
	}
	c.gconn.Close()
	c.closed = true
}

func (c *Client) CallImageProcessing(
	req *RequestImageProcessing) (*ResponseImageProcessing, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.gclient.CallImageProcessing(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
