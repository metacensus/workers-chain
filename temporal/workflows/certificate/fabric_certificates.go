package certificate

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const FabricCertificatesWorkflowName = "certificate.FabricCertificatesWorkflow"

// FabricCertificatesRequest represents a request to generate all Fabric certificates
type FabricCertificatesRequest struct {
	Namespace    string `json:"namespace"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	Province     string `json:"province"`
	Locality     string `json:"locality"`
	ValidDays    int    `json:"valid_days"`
}

// FabricCertificatesResult represents the result of generating all Fabric certificates
type FabricCertificatesResult struct {
	OrdererResults []*CertificateResult `json:"orderer_results"`
	PeerResults    []*CertificateResult `json:"peer_results"`
	CAResults      []*CertificateResult `json:"ca_results"`
	Status         string               `json:"status"`
	Message        string               `json:"message"`
}

// FabricCertificatesWorkflow generates all certificates needed for Hyperledger Fabric
func FabricCertificatesWorkflow(ctx workflow.Context, req FabricCertificatesRequest) (*FabricCertificatesResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Fabric certificates workflow started", "namespace", req.Namespace)

	// Set default values
	if req.ValidDays == 0 {
		req.ValidDays = 365
	}
	if req.Country == "" {
		req.Country = "US"
	}
	if req.Province == "" {
		req.Province = "CA"
	}
	if req.Locality == "" {
		req.Locality = "San Francisco"
	}

	// Generate orderer certificates
	ordererResults := make([]*CertificateResult, 0)
	for i := 0; i < 3; i++ { // Generate 3 orderer certificates
		ordererReq := CertificateRequest{
			Name:         fmt.Sprintf("orderer%d", i),
			Namespace:    req.Namespace,
			Type:         "orderer",
			SecretName:   fmt.Sprintf("orderer%d-cert", i),
			CommonName:   fmt.Sprintf("orderer%d.%s", i, req.Namespace),
			Organization: req.Organization,
			Country:      req.Country,
			Province:     req.Province,
			Locality:     req.Locality,
			ValidDays:    req.ValidDays,
		}

		var result CertificateResult
		err := workflow.ExecuteChildWorkflow(ctx, GenerateCertificateWorkflow, ordererReq).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to generate orderer certificate", "orderer", ordererReq.Name, "error", err)
			return nil, err
		}
		ordererResults = append(ordererResults, &result)
	}

	// Generate peer certificates
	peerResults := make([]*CertificateResult, 0)
	for i := 0; i < 3; i++ { // Generate 3 peer certificates
		peerReq := CertificateRequest{
			Name:         fmt.Sprintf("peer%d", i),
			Namespace:    req.Namespace,
			Type:         "peer",
			SecretName:   fmt.Sprintf("peer%d-cert", i),
			CommonName:   fmt.Sprintf("peer%d.%s", i, req.Namespace),
			Organization: req.Organization,
			Country:      req.Country,
			Province:     req.Province,
			Locality:     req.Locality,
			ValidDays:    req.ValidDays,
		}

		var result CertificateResult
		err := workflow.ExecuteChildWorkflow(ctx, GenerateCertificateWorkflow, peerReq).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to generate peer certificate", "peer", peerReq.Name, "error", err)
			return nil, err
		}
		peerResults = append(peerResults, &result)
	}

	// Generate CA certificates
	caResults := make([]*CertificateResult, 0)
	caReq := CertificateRequest{
		Name:         "ca0",
		Namespace:    req.Namespace,
		Type:         "ca",
		SecretName:   "ca0-cert",
		CommonName:   fmt.Sprintf("ca0.%s", req.Namespace),
		Organization: req.Organization,
		Country:      req.Country,
		Province:     req.Province,
		Locality:     req.Locality,
		ValidDays:    req.ValidDays,
	}

	var caResult CertificateResult
	err := workflow.ExecuteChildWorkflow(ctx, GenerateCertificateWorkflow, caReq).Get(ctx, &caResult)
	if err != nil {
		logger.Error("Failed to generate CA certificate", "ca", caReq.Name, "error", err)
		return nil, err
	}
	caResults = append(caResults, &caResult)

	logger.Info("Fabric certificates workflow completed successfully",
		"orderers", len(ordererResults),
		"peers", len(peerResults),
		"cas", len(caResults))

	return &FabricCertificatesResult{
		OrdererResults: ordererResults,
		PeerResults:    peerResults,
		CAResults:      caResults,
		Status:         "success",
		Message:        "All Fabric certificates generated successfully",
	}, nil
}
