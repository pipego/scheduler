package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"

	mock "github.com/pipego/scheduler/server/mock"
	pb "github.com/pipego/scheduler/server/proto"

	"github.com/pipego/scheduler/external/grpctest"
)

type rpcMsg struct {
	msg proto.Message
}

type rpcTest struct {
	grpctest.Tester
}

func TestServer(t *testing.T) {
	grpctest.RunSubTests(t, rpcTest{})
}

func (r *rpcMsg) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}

	return proto.Equal(m, r.msg)
}

func (r *rpcMsg) String() string {
	return fmt.Sprintf("msg: %s", r.msg)
}

func (rpcTest) TestSendServer(t *testing.T) {
	helper := func(t *testing.T, client pb.ServerProtoClient) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := client.SendServer(ctx, &pb.ServerRequest{ApiVersion: "v1"})
		if err != nil || r.GetError() != "" || r.GetName() != "node" {
			t.Errorf("mocking failed")
		}

		t.Log("reply: ", r.GetName())
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ServerRequest{ApiVersion: "v1"}

	client := mock.NewMockServerProtoClient(ctrl)
	client.EXPECT().SendServer(
		gomock.Any(),
		&rpcMsg{msg: req},
	).Return(&pb.ServerReply{Name: "node", Error: ""}, nil)

	helper(t, client)
}
