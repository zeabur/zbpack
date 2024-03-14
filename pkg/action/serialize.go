package action

import (
	"encoding/base64"
	"fmt"

	zbaction "github.com/zeabur/action"
	zbactionpb "github.com/zeabur/action/proto"
	"google.golang.org/protobuf/proto"
)

// SerializeAction serializes the given action to bytes that can transfer to another process.
func SerializeAction(action zbaction.Action) (string, error) {
	actionPb, err := zbaction.ActionToProto(action)
	if err != nil {
		return "", fmt.Errorf("convert action: %w", err)
	}

	actionByte, err := proto.Marshal(actionPb)
	if err != nil {
		return "", fmt.Errorf("marshal action: %w", err)
	}

	actionByteB64 := base64.URLEncoding.EncodeToString(actionByte)
	return actionByteB64, nil
}

// DeserializeAction deserializes the base64-encoded data to an action.
func DeserializeAction(actionByteB64 string) (zbaction.Action, error) {
	actionByte, err := base64.URLEncoding.DecodeString(actionByteB64)

	actionPb := new(zbactionpb.Action)
	if err := proto.Unmarshal(actionByte, actionPb); err != nil {
		return zbaction.Action{}, fmt.Errorf("unmarshal action: %w", err)
	}

	action, err := zbaction.ActionFromProto(actionPb)
	if err != nil {
		return zbaction.Action{}, fmt.Errorf("convert action: %w", err)
	}

	return action, nil
}
