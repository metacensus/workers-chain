package channel

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/channel/v1"
)

// ErrNotImplemented is returned by every activity in this package until the real
// logic lands. Stubs fail loudly rather than silently reporting success.
var ErrNotImplemented = errors.New("not implemented")

// Activities implements the channel config/governance activities
// (pb.ChannelActivities). Their subject is the channel configuration, not any node.
type Activities struct{}

// NewActivities creates a new Activities instance.
func NewActivities() *Activities {
	return &Activities{}
}

// GenerateConfigTx generates a configtx.yaml from org MSP definitions.
func (a *Activities) GenerateConfigTx(_ context.Context, _ *pb.GenerateConfigTxInput) (*pb.GenerateConfigTxOutput, error) {
	return nil, fmt.Errorf("GenerateConfigTx: %w", ErrNotImplemented)
}

// GenerateGenesisBlock creates the channel genesis block using configtxgen.
func (a *Activities) GenerateGenesisBlock(_ context.Context, _ *pb.GenerateGenesisBlockInput) (*pb.GenerateGenesisBlockOutput, error) {
	return nil, fmt.Errorf("GenerateGenesisBlock: %w", ErrNotImplemented)
}

// FetchConfig fetches and decodes the current channel configuration block.
func (a *Activities) FetchConfig(_ context.Context, _ *pb.FetchConfigInput) (*pb.FetchConfigOutput, error) {
	return nil, fmt.Errorf("FetchConfig: %w", ErrNotImplemented)
}

// ComputeConfigUpdate computes the config update delta between original and modified configs.
func (a *Activities) ComputeConfigUpdate(_ context.Context, _ *pb.ComputeConfigUpdateInput) (*pb.ComputeConfigUpdateOutput, error) {
	return nil, fmt.Errorf("ComputeConfigUpdate: %w", ErrNotImplemented)
}

// SignConfigUpdate signs a channel config update with the org admin identity.
func (a *Activities) SignConfigUpdate(_ context.Context, _ *pb.SignConfigUpdateInput) (*pb.SignConfigUpdateOutput, error) {
	return nil, fmt.Errorf("SignConfigUpdate: %w", ErrNotImplemented)
}

// SubmitConfigUpdate wraps a config update in an envelope and submits it to the orderer.
func (a *Activities) SubmitConfigUpdate(_ context.Context, _ *pb.SubmitConfigUpdateInput) (*pb.SubmitConfigUpdateOutput, error) {
	return nil, fmt.Errorf("SubmitConfigUpdate: %w", ErrNotImplemented)
}

// GenerateOrgDefinition creates an org MSP definition for use in AddOrgToChannel.
func (a *Activities) GenerateOrgDefinition(_ context.Context, _ *pb.GenerateOrgDefinitionInput) (*pb.GenerateOrgDefinitionOutput, error) {
	return nil, fmt.Errorf("GenerateOrgDefinition: %w", ErrNotImplemented)
}

// CollectSignatures gathers required config update signatures from existing channel admins.
func (a *Activities) CollectSignatures(_ context.Context, _ *pb.CollectSignaturesInput) (*pb.CollectSignaturesOutput, error) {
	return nil, fmt.Errorf("CollectSignatures: %w", ErrNotImplemented)
}
