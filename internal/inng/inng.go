package inng

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol"
	"go.nanomsg.org/mangos/v3/protocol/rep"
	"go.nanomsg.org/mangos/v3/protocol/req"

	_ "go.nanomsg.org/mangos/v3/transport/tcp"
)

type Server struct {
	address string
	wg      *sync.WaitGroup
	nsock   mangos.Socket
	closed  bool
}

func NewServer(address string) *Server {

	s := &Server{
		address: address,
		wg:      &sync.WaitGroup{},
	}

	return s
}

func (s *Server) StartListenAndServe() error {
	var err error

	if s.nsock, err = rep.NewSocket(); err != nil {
		return err
	}

	s.nsock.SetOption(protocol.OptionSendDeadline, time.Second)

	log.Printf("NNG server start on %s", s.address)
	if err = s.nsock.Listen(s.address); err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		var err error
		var msg []byte
		defer s.wg.Done()
		for {
			if msg, err = s.nsock.Recv(); err != nil {
				break
			}

			if msg, err = s.callImageProcessing(msg); err != nil {
				s.nsock.Send([]byte{0})
				continue
			}

			if err = s.nsock.Send(msg); err != nil {
				log.Printf("Failed on nsocket send: %v", err)
				break
			}
		}
	}()

	return nil
}

func (s *Server) Close() {
	if s.closed {
		return
	}
	s.nsock.Close()
	s.wg.Wait()
	log.Printf("NNG server closed")
	s.closed = true
}

func (s *Server) callImageProcessing(msg []byte) ([]byte, error) {
	req := &RequestImageProcessing{}
	payload, err := req.UnmarshalCborWithPayload(msg)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return nil, err
	}

	res := ResponseImageProcessing{
		ReqMessageId: req.MessageId,
		Success:      true,
		SrcWidth:     1280,
		SrcHeight:    720,
		DstWidth:     640,
		DstHeight:    480,
	}

	msg, err = res.MarshalCborWithPayload(payload)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return nil, err
	}

	return msg, nil
}

// ---------------------------------------------------------

type Client struct {
	nsock   mangos.Socket
	address string
	closed  bool
}

func NewClient(address string) (*Client, error) {
	var nsock mangos.Socket
	var err error

	if nsock, err = req.NewSocket(); err != nil {
		err = fmt.Errorf("Can't get new req socket: %v", err)
		return nil, err
	}

	nsock.SetOption(protocol.OptionRecvDeadline, time.Second)

	if err := nsock.Dial(address); err != nil {
		err = fmt.Errorf("Can't dial on socket %s: %v", address, err)
		return nil, err
	}

	c := &Client{
		nsock:   nsock,
		address: address,
	}

	return c, nil
}

func (c *Client) Close() {
	if c.closed {
		return
	}
	c.nsock.Close()
	c.closed = true
}

func (c *Client) CallImageProcessing(
	req *RequestImageProcessing, payload []byte) (*ResponseImageProcessing, error) {
	var err error

	sctx, err := c.nsock.OpenContext()
	if err != nil {
		return nil, err
	}
	defer sctx.Close()

	msg, err := req.MarshalCborWithPayload(payload)
	if err != nil {
		return nil, err
	}

	if err := sctx.Send(msg); err != nil {
		return nil, err
	}

	msg, err = sctx.Recv()
	if err != nil {
		return nil, err
	}

	res := &ResponseImageProcessing{}
	payload, err = res.UnmarshalCborWithPayload(msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}
