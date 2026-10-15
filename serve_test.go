package gateway

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/mlund01/squadron-gateway-sdk/proto"
)

type postOnlyGateway struct{ Gateway }

func (postOnlyGateway) PostMessage(context.Context, PostMessageRequest) (PostMessageResponse, error) {
	return PostMessageResponse{MessageID: "C1/1700000000.000001"}, nil
}

type editingGateway struct {
	postOnlyGateway
	edited EditMessageRequest
}

func (g *editingGateway) EditMessage(_ context.Context, req EditMessageRequest) error {
	g.edited = req
	return nil
}

func TestPostMessageReturnsTheGatewaysMessageID(t *testing.T) {
	s := &gatewayServer{impl: postOnlyGateway{}}
	resp, err := s.PostMessage(context.Background(), &pb.PostMessageRequest{Payload: `{"text":"hi"}`})
	if err != nil || resp.GetMessageId() != "C1/1700000000.000001" {
		t.Errorf("got %v, %v", resp, err)
	}
}

func TestEditMessageWithoutEditorIsUnimplemented(t *testing.T) {
	s := &gatewayServer{impl: postOnlyGateway{}}
	_, err := s.EditMessage(context.Background(), &pb.EditMessageRequest{MessageId: "C1/1"})
	if status.Code(err) != codes.Unimplemented {
		t.Errorf("want Unimplemented, got %v", err)
	}
}

func TestEditMessageReachesTheEditor(t *testing.T) {
	g := &editingGateway{}
	s := &gatewayServer{impl: g}
	if _, err := s.EditMessage(context.Background(), &pb.EditMessageRequest{MessageId: "C1/1", Payload: `{"text":"done"}`, MissionId: "run-1"}); err != nil {
		t.Fatal(err)
	}
	if want := (EditMessageRequest{MessageID: "C1/1", Payload: `{"text":"done"}`, MissionID: "run-1"}); g.edited != want {
		t.Errorf("got %+v, want %+v", g.edited, want)
	}
}
