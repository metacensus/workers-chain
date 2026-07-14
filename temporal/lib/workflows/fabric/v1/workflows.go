package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	certauthpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for fabric workflows
var (
	ErrInvalidInput               = errors.New("invalid input")
	ErrNotOwned                   = errors.New("certificate authority not owned by this worker")
	ErrDeployCouchDB              = errors.New("deploy couchdb")
	ErrDeployPeer                 = errors.New("deploy peer")
	ErrDeployOrderer              = errors.New("deploy orderer")
	ErrDeployChainCode            = errors.New("deploy chaincode")
	ErrDeployCertificateStore     = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase  = errors.New("create certificate database")
	ErrDeployCertificateAuthority = errors.New("deploy certificate authority")
	ErrCreateCertificate          = errors.New("create certificate")
	ErrCreateBootstrapAdmin       = errors.New("create bootstrap admin")
	ErrCreateAdmin                = errors.New("create admin")
	ErrCreateIdentity             = errors.New("create identity")
	ErrCreateAccess               = errors.New("create access")
	ErrInvalidCertType            = errors.New("invalid cert type")
)

// Workflows implements the FabricWorkflows interface.
//
// Fabric workflows act AS their worker's owner: the owner is injected at
// worker startup (operator for the shared fabric worker, org for shard
// workers) and never carried in workflow inputs. All Vault paths and
// Kubernetes namespaces are derived from the injected owner, so a workflow
// physically cannot operate on another owner's resources by mistake.
//
// The owner is constant for the lifetime of a Temporal namespace, so reading
// it inside workflow code is deterministic and replay-safe.
type Workflows struct {
	owner config.OwnerConfig
}

// NewWorkflows creates a new Workflows instance acting as the given owner.
func NewWorkflows(owner config.OwnerConfig) *Workflows {
	return &Workflows{owner: owner}
}

func (w *Workflows) BootstrapFabric(ctx workflow.Context, input *pb.BootstrapFabricWorkflowInput) (pb.BootstrapFabricWorkflow, error) {
	return newBootstrapFabricWorkflow(input, w.owner)
}

type bootstrapFabricWorkflow struct {
	input *pb.BootstrapFabricWorkflowInput
	owner config.OwnerConfig
}

// newBootstrapFabricWorkflow creates and validates a new bootstrapFabricWorkflow
func newBootstrapFabricWorkflow(input *pb.BootstrapFabricWorkflowInput, owner config.OwnerConfig) (*bootstrapFabricWorkflow, error) {
	w := &bootstrapFabricWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapFabricWorkflow) validate() error {
	if len(w.input.Req.GetCertificateAuthorities()) == 0 {
		return fmt.Errorf("%w: at least one certificate authority spec is required", ErrInvalidInput)
	}
	return nil
}

// Execute bootstraps the shared (operator-owned) Fabric infrastructure:
// the certificate store and the operator-owned certificate authorities,
// deployed in spec order. Runs on the fabric worker in the "fabric" Temporal
// namespace; called via Nexus from the central infra worker.
func (w *bootstrapFabricWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapFabricOutput, error) {
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	for _, spec := range w.input.Req.GetCertificateAuthorities() {
		if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
			Spec: spec,
		}); err != nil {
			return nil, fmt.Errorf("%w (%s-%s): %w", ErrDeployCertificateAuthority,
				spec.GetSelf().GetCertType(), spec.GetSelf().GetId(), err)
		}
	}

	return &pb.BootstrapFabricOutput{}, nil
}

func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return newBootstrapShardWorkflow(input, w.owner)
}

type bootstrapShardWorkflow struct {
	input *pb.BootstrapShardWorkflowInput
	owner config.OwnerConfig
}

// newBootstrapShardWorkflow creates and validates a new bootstrapShardWorkflow
func newBootstrapShardWorkflow(input *pb.BootstrapShardWorkflowInput, owner config.OwnerConfig) (*bootstrapShardWorkflow, error) {
	w := &bootstrapShardWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapShardWorkflow) validate() error {
	deployment := w.input.Req.GetDeployment()
	if deployment == nil {
		return fmt.Errorf("%w: deployment is required", ErrInvalidInput)
	}
	if len(deployment.GetCertificateAuthorities()) == 0 {
		return fmt.Errorf("%w: at least one certificate authority is required", ErrInvalidInput)
	}
	if len(deployment.GetPeers()) == 0 {
		return fmt.Errorf("%w: at least one peer is required", ErrInvalidInput)
	}
	if len(deployment.GetOrderers()) == 0 {
		return fmt.Errorf("%w: at least one orderer is required", ErrInvalidInput)
	}
	return nil
}

// Execute bootstraps the org-owned Fabric shard: the certificate store, the
// org certificate authorities (in spec order), and the peers and orderers.
// Runs on the shard worker in the org's Temporal namespace; called via Nexus
// from the central infra worker.
func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	deployment := w.input.Req.GetDeployment()

	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	for _, spec := range deployment.GetCertificateAuthorities() {
		if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
			Spec: spec,
		}); err != nil {
			return nil, fmt.Errorf("%w (%s-%s): %w", ErrDeployCertificateAuthority,
				spec.GetSelf().GetCertType(), spec.GetSelf().GetId(), err)
		}
	}

	// Peer/orderer deployment is intentionally out of scope for this PR, which
	// covers the certificate-authority deployment and CSR issuance pipeline
	// only. Bringing up CouchDB/peers/orderers (and the Vault policy + network
	// wiring they need) is tracked as separate follow-up work; until then we
	// stop after the CAs are deployed. Flip deployWorkloads to true once peer
	// startup lands.
	deployWorkloads := false
	if !deployWorkloads {
		return &pb.BootstrapShardOutput{}, nil
	}

	for _, peerSpec := range deployment.GetPeers() {
		if _, err := pb.DeployCouchDbChild(ctx, &pb.DeployCouchDbInput{
			PeerId: peerSpec.GetId(),
		}); err != nil {
			return nil, fmt.Errorf("deploy couchdb for %s: %w", peerSpec.GetId(), err)
		}

		if _, err := pb.DeployPeerChild(ctx, &pb.DeployPeerInput{
			Spec: peerSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy peer %s: %w", peerSpec.GetId(), err)
		}
	}

	for _, ordererSpec := range deployment.GetOrderers() {
		if _, err := pb.DeployOrdererChild(ctx, &pb.DeployOrdererInput{
			Spec: ordererSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy orderer %s: %w", ordererSpec.GetId(), err)
		}
	}

	// TODO: Deploy chaincode, hook into channels

	return &pb.BootstrapShardOutput{}, nil
}

func (w *Workflows) DeployCouchDb(ctx workflow.Context, input *pb.DeployCouchDbWorkflowInput) (pb.DeployCouchDbWorkflow, error) {
	return newDeployCouchDbWorkflow(input, w.owner)
}

type deployCouchDbWorkflow struct {
	input *pb.DeployCouchDbWorkflowInput
	owner config.OwnerConfig
}

// newDeployCouchDbWorkflow creates and validates a new deployCouchDbWorkflow
func newDeployCouchDbWorkflow(input *pb.DeployCouchDbWorkflowInput, owner config.OwnerConfig) (*deployCouchDbWorkflow, error) {
	w := &deployCouchDbWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCouchDbWorkflow) validate() error {
	if w.input.Req.GetPeerId() == "" {
		return fmt.Errorf("%w: peer_id is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployCouchDbWorkflow) Execute(ctx workflow.Context) (*pb.DeployCouchDbOutput, error) {
	peerID := w.input.Req.GetPeerId()
	couchDbConfig := fabricconfig.NewCouchDbConfig(w.owner, peerID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	_, err := NexusSecretsCreateAccess(ctx, couchDbConfig)
	if err != nil {
		return nil, err
	}

	credentialsRef := couchDbConfig.VaultAdminRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.EnsurePasswordActivityName, &secretspb.EnsurePasswordInput{
		SecretRef: credentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate couchdb credentials: %w", ErrDeployCouchDB, err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     w.owner.OwnerProto(),
		Component: couchDbConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create couchdb deployment: %w", ErrDeployCouchDB, err)
	}

	// This ensures the peer can connect to CouchDB when it starts
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    0, // Infinite retry until timeout
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err = workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			couchDbConfig.ComponentName(): couchDbConfig.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: CouchDB did not become ready: %w", ErrDeployCouchDB, err)
	}

	return &pb.DeployCouchDbOutput{}, nil
}

func (w *Workflows) DeployPeer(ctx workflow.Context, input *pb.DeployPeerWorkflowInput) (pb.DeployPeerWorkflow, error) {
	return newDeployPeerWorkflow(input, w.owner)
}

type deployPeerWorkflow struct {
	input *pb.DeployPeerWorkflowInput
	owner config.OwnerConfig
}

// newDeployPeerWorkflow creates and validates a new deployPeerWorkflow
func newDeployPeerWorkflow(input *pb.DeployPeerWorkflowInput, owner config.OwnerConfig) (*deployPeerWorkflow, error) {
	w := &deployPeerWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployPeerWorkflow) validate() error {
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	if spec.GetOrgIssuer() == nil {
		return fmt.Errorf("%w: spec.org_issuer is required", ErrInvalidInput)
	}
	if spec.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: spec.tls_issuer is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployPeerWorkflow) Execute(ctx workflow.Context) (*pb.DeployPeerOutput, error) {
	spec := w.input.Req.GetSpec()
	peerID := spec.GetId()
	peerConfig := fabricconfig.NewPeerConfig(w.owner, peerID)

	_, err := NexusSecretsCreateAccess(ctx, peerConfig)
	if err != nil {
		return nil, err
	}

	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "peer" ensures the certificate gets OU=peer for NodeOUs classification
	_, err = certauthpb.CreateIdentityChild(ctx, &certauthpb.CreateIdentityInput{
		CommonName:   peerConfig.CommonName(),
		Hosts:        peerConfig.TLSDNSNames(),
		OrgIssuerRef: spec.GetOrgIssuer(),
		TlsIssuerRef: spec.GetTlsIssuer(),
		SigningStorage: &pb.IdentityRef{
			KeyRef:  peerConfig.VaultSigningKeyRef(),
			CertRef: peerConfig.VaultSigningCertRef(),
		},
		TlsStorage: &pb.IdentityRef{
			KeyRef:  peerConfig.VaultTLSKeyRef(),
			CertRef: peerConfig.VaultTLSCertRef(),
		},
		TlsProfile:   "tls",
		IdentityType: "peer", // Critical: sets OU=peer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer identity: %w", ErrDeployPeer, err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     w.owner.OwnerProto(),
		Component: peerConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer deployment: %w", ErrDeployPeer, err)
	}

	return &pb.DeployPeerOutput{}, nil
}

func (w *Workflows) DeployOrderer(ctx workflow.Context, input *pb.DeployOrdererWorkflowInput) (pb.DeployOrdererWorkflow, error) {
	return newDeployOrdererWorkflow(input, w.owner)
}

type deployOrdererWorkflow struct {
	input *pb.DeployOrdererWorkflowInput
	owner config.OwnerConfig
}

// newDeployOrdererWorkflow creates and validates a new deployOrdererWorkflow
func newDeployOrdererWorkflow(input *pb.DeployOrdererWorkflowInput, owner config.OwnerConfig) (*deployOrdererWorkflow, error) {
	w := &deployOrdererWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployOrdererWorkflow) validate() error {
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	if spec.GetOrgIssuer() == nil {
		return fmt.Errorf("%w: spec.org_issuer is required", ErrInvalidInput)
	}
	if spec.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: spec.tls_issuer is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrdererWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrdererOutput, error) {
	spec := w.input.Req.GetSpec()
	ordererID := spec.GetId()
	ordererConfig := fabricconfig.NewOrdererConfig(w.owner, ordererID)

	_, err := NexusSecretsCreateAccess(ctx, ordererConfig)
	if err != nil {
		return nil, err
	}

	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "orderer" ensures the certificate gets OU=orderer for NodeOUs classification
	_, err = certauthpb.CreateIdentityChild(ctx, &certauthpb.CreateIdentityInput{
		CommonName:   ordererConfig.CommonName(),
		Hosts:        ordererConfig.TLSDNSNames(),
		OrgIssuerRef: spec.GetOrgIssuer(),
		TlsIssuerRef: spec.GetTlsIssuer(),
		SigningStorage: &pb.IdentityRef{
			KeyRef:  ordererConfig.VaultSigningKeyRef(),
			CertRef: ordererConfig.VaultSigningCertRef(),
		},
		TlsStorage: &pb.IdentityRef{
			KeyRef:  ordererConfig.VaultTLSKeyRef(),
			CertRef: ordererConfig.VaultTLSCertRef(),
		},
		TlsProfile:   "tls",
		IdentityType: "orderer", // Critical: sets OU=orderer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer identity: %w", ErrDeployOrderer, err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     w.owner.OwnerProto(),
		Component: ordererConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer deployment: %w", ErrDeployOrderer, err)
	}

	return &pb.DeployOrdererOutput{}, nil
}

func (w *Workflows) DeployChainCode(ctx workflow.Context, input *pb.DeployChainCodeWorkflowInput) (pb.DeployChainCodeWorkflow, error) {
	return newDeployChainCodeWorkflow(input, w.owner)
}

type deployChainCodeWorkflow struct {
	input *pb.DeployChainCodeWorkflowInput
	owner config.OwnerConfig
}

// newDeployChainCodeWorkflow creates and validates a new deployChainCodeWorkflow
func newDeployChainCodeWorkflow(input *pb.DeployChainCodeWorkflowInput, owner config.OwnerConfig) (*deployChainCodeWorkflow, error) {
	w := &deployChainCodeWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployChainCodeWorkflow) validate() error {
	if w.input.Req.GetChainCodeId() == "" {
		return fmt.Errorf("%w: chain_code_id is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployChainCodeWorkflow) Execute(ctx workflow.Context) (*pb.DeployChainCodeOutput, error) {
	// chainCodeID := w.input.Req.GetChainCodeId()

	// TODO: Implement chaincode deployment
	// 1. Nexus Call: secrets.v1.CreateAccess
	// 2. for only ("org"):
	//    - Workflow: fabric.v1.CreateIdentity (with IdentityType: "client")
	// 3. Workflow: cluster.fabric.v1.CreateDeployment
	// NOTE: IdentityType should be "client" for chaincode identities
	return &pb.DeployChainCodeOutput{}, nil
}
