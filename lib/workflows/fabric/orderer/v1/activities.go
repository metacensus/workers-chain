package orderer

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/orderer/v1"
)

// ErrNotImplemented is returned by every activity in this package until the real
// logic lands. Stubs fail loudly rather than silently reporting success.
var ErrNotImplemented = errors.New("not implemented")

// Activities implements the orderer channel-participation activities
// (pb.OrdererActivities). These act on an orderer node the executing worker owns.
type Activities struct{}

// NewActivities creates a new Activities instance.
func NewActivities() *Activities {
	return &Activities{}
}

// ChannelJoin joins an orderer node to a channel via the channel participation API.
func (a *Activities) ChannelJoin(_ context.Context, _ *pb.ChannelJoinInput) (*pb.ChannelJoinOutput, error) {
	return nil, fmt.Errorf("ChannelJoin: %w", ErrNotImplemented)
}

// WaitForActive polls osnadmin channel list until the orderer reaches the desired status.
func (a *Activities) WaitForActive(_ context.Context, _ *pb.WaitForActiveInput) (*pb.WaitForActiveOutput, error) {
	return nil, fmt.Errorf("WaitForActive: %w", ErrNotImplemented)
}
