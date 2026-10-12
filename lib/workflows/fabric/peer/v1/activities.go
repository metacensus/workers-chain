package peer

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/peer/v1"
)

// ErrNotImplemented is returned by every activity in this package until the real
// logic lands. Stubs fail loudly rather than silently reporting success.
var ErrNotImplemented = errors.New("not implemented")

// Activities implements the peer channel-membership activities (pb.PeerActivities).
// These act on a peer node the executing worker owns.
type Activities struct{}

// NewActivities creates a new Activities instance.
func NewActivities() *Activities {
	return &Activities{}
}

// ChannelFetch fetches a block from an orderer for the peer to use when joining a channel.
func (a *Activities) ChannelFetch(_ context.Context, _ *pb.ChannelFetchInput) (*pb.ChannelFetchOutput, error) {
	return nil, fmt.Errorf("ChannelFetch: %w", ErrNotImplemented)
}

// ChannelJoin joins the peer node to a channel using a previously fetched block.
func (a *Activities) ChannelJoin(_ context.Context, _ *pb.ChannelJoinInput) (*pb.ChannelJoinOutput, error) {
	return nil, fmt.Errorf("ChannelJoin: %w", ErrNotImplemented)
}
