package fabric

import (
	"fmt"
	"time"

	secretspb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
	"github.com/metacensus/workers-chain/gen/workflows/secrets/v1/secretsv1nexustemporal"
	vaultv1 "github.com/metacensus/workers-chain/gen/workflows/vault/v1"
	cellv1 "github.com/metacensus/workers-chain/lib/workflows/cell/v1"
	"go.temporal.io/sdk/workflow"
)

type Accessor interface {
	PolicyName() string
	VaultPolicy() *vaultv1.PolicyDocument
	AuthRoleName() string
	ServiceAccountName() string
	KubernetesNamespace() string
}

func NexusSecretsCreateAccess(ctx workflow.Context, caConfig Accessor) (*secretspb.CreateAccessOutput, error) {
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(cellv1.SecretsCreateAccessNexusEndpoint)
	output, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          caConfig.PolicyName(),
		PolicyDocument:      caConfig.VaultPolicy(),
		AuthRoleName:        caConfig.AuthRoleName(),
		ServiceAccountName:  caConfig.ServiceAccountName(),
		KubernetesNamespace: caConfig.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccess, err)
	}
	return output, nil
}
