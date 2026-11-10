package serverfx_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	commonpb "go.temporal.io/api/common/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	commonnexus "go.temporal.io/server/common/nexus"
	"go.temporal.io/server/common/nexus/nexusrpc"
	sqliteplugin "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"go.temporal.io/server/common/testing/freeport"
	"go.temporal.io/server/schema/sqlite"
	"go.temporal.io/server/temporal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/metacensus/workers-chain/temporal-server/authz"
	"github.com/metacensus/workers-chain/temporal-server/serverfx"
)

const (
	targetNamespace = "fabric"
	targetTaskQueue = "certauth"
	acmeEndpoint    = "fabric-shard-acme-CreateCertificate"
)

// The payload binding has to hold on the path a request really takes:
// nexusHandler.StartOperation authorizes before it reads the payload, so a
// check that only exists in a unit test of a hand-built request cannot show
// it runs where the payload is set. This starts the server from
// TopLevelModule and sends each case through the frontend's Nexus HTTP
// handler, the way the history service dispatches an in-workflow call.
func TestCertificateEndpointBindsRequesterThroughFrontend(t *testing.T) {
	srv := startServer(t)

	cases := []struct {
		name    string
		payload *commonpb.Payload
		allow   bool
	}{
		{name: "requester is the endpoint's org", payload: signCertificatePayload("acme"), allow: true},
		{name: "requester is another org", payload: signCertificatePayload("other")},
		{name: "requester is absent", payload: signCertificatePayload("")},
		{name: "payload is empty", payload: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, stopPolling := context.WithTimeout(context.Background(), 30*time.Second)
			polled := srv.pollOnce(ctx)
			_, err := srv.startSignCertificate(t, tc.payload)
			// Dispatch is a sync match: an allowed task was polled and
			// completed before StartOperation returned, and a denied one
			// never will be.
			stopPolling()
			task := <-polled

			if tc.allow {
				if err != nil {
					t.Fatalf("StartOperation() error = %v, want success", err)
				}
				if task == nil {
					t.Fatal("no Nexus task reached the worker")
				}
				return
			}
			var handlerErr *nexus.HandlerError
			if !errors.As(err, &handlerErr) || handlerErr.Type != nexus.HandlerErrorTypeUnauthorized {
				t.Fatalf("StartOperation() error = %v, want an UNAUTHORIZED handler error", err)
			}
			if task != nil {
				t.Fatal("a denied request reached the worker")
			}
		})
	}
}

func signCertificatePayload(org string) *commonpb.Payload {
	if org == "" {
		return jsonProtoPayload(`{}`)
	}
	return jsonProtoPayload(`{"requester":{"orgId":"` + org + `"}}`)
}

// jsonProtoPayload is what the SDK's default data converter produces for
// SignCertificateInput, so the frontend sees the content type a worker sends.
func jsonProtoPayload(data string) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			"encoding":    []byte("json/protobuf"),
			"messageType": []byte("fabric.certauth.v1.SignCertificateInput"),
		},
		Data: []byte(data),
	}
}

type testServer struct {
	httpAddr   string
	endpointID string
	workflow   workflowservice.WorkflowServiceClient
}

func startServer(t *testing.T) *testServer {
	t.Helper()
	cfg := liteConfig()
	sqlCfg := cfg.Persistence.DataStores[sqliteplugin.PluginName].SQL
	nsCfg, err := sqlite.NewNamespaceConfig(cfg.ClusterMetadata.CurrentClusterName, targetNamespace, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.CreateNamespaces(sqlCfg, nsCfg); err != nil {
		t.Fatal(err)
	}

	s, err := temporal.NewServerFx(
		serverfx.TopLevelModule,
		// The worker service runs system workflows only, and needs an SDK
		// client this configuration does not give it.
		temporal.ForServices([]string{"frontend", "history", "matching"}),
		temporal.WithConfig(cfg),
		temporal.WithLogger(log.NewNoopLogger()),
		temporal.WithAuthorizer(authz.NewAuthorizer(nil)),
		// Every call here carries the internode identity, as an in-workflow
		// Nexus dispatch does in production: a system role, so only the
		// payload binding can deny.
		temporal.WithClaimMapper(func(*config.Config) authorization.ClaimMapper { return internodeClaims{} }),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })

	conn, err := grpc.NewClient(cfg.PublicClient.HostPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	request := &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name: acmeEndpoint,
			Target: &nexuspb.EndpointTarget{Variant: &nexuspb.EndpointTarget_Worker_{
				Worker: &nexuspb.EndpointTarget_Worker{Namespace: targetNamespace, TaskQueue: targetTaskQueue},
			}},
		},
	}
	// The frontend answers Unavailable until it reports itself healthy.
	var created *operatorservice.CreateNexusEndpointResponse
	deadline := time.Now().Add(30 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		created, err = operatorservice.NewOperatorServiceClient(conn).CreateNexusEndpoint(ctx, request)
		cancel()
		if status.Code(err) != codes.Unavailable || time.Now().After(deadline) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("CreateNexusEndpoint() error = %v", err)
	}

	return &testServer{
		httpAddr:   fmt.Sprintf("127.0.0.1:%d", cfg.Services["frontend"].RPC.HTTPPort),
		endpointID: created.GetEndpoint().GetId(),
		workflow:   workflowservice.NewWorkflowServiceClient(conn),
	}
}

// pollOnce polls the endpoint's task queue for one Nexus task until ctx ends,
// completes it, and sends it on the channel; nil means none arrived.
func (s *testServer) pollOnce(ctx context.Context) <-chan *workflowservice.PollNexusTaskQueueResponse {
	out := make(chan *workflowservice.PollNexusTaskQueueResponse, 1)
	go func() {
		res, err := s.workflow.PollNexusTaskQueue(ctx, &workflowservice.PollNexusTaskQueueRequest{
			Namespace: targetNamespace,
			Identity:  "test-worker",
			TaskQueue: &taskqueuepb.TaskQueue{Name: targetTaskQueue},
		})
		if err != nil || len(res.GetTaskToken()) == 0 {
			out <- nil
			return
		}
		_, _ = s.workflow.RespondNexusTaskCompleted(ctx, &workflowservice.RespondNexusTaskCompletedRequest{
			Namespace: targetNamespace,
			Identity:  "test-worker",
			TaskToken: res.GetTaskToken(),
			Response: &nexuspb.Response{Variant: &nexuspb.Response_StartOperation{
				StartOperation: &nexuspb.StartOperationResponse{Variant: &nexuspb.StartOperationResponse_SyncSuccess{
					SyncSuccess: &nexuspb.StartOperationResponse_Sync{Payload: jsonProtoPayload(`{}`)},
				}},
			}},
		})
		out <- res
	}()
	return out
}

// startSignCertificate retries while the frontend's endpoint registry has not
// yet loaded the endpoint, and returns the first other outcome.
func (s *testServer) startSignCertificate(t *testing.T, payload *commonpb.Payload) (*nexusrpc.ClientStartOperationResponse[*nexus.LazyValue], error) {
	t.Helper()
	client, err := nexusrpc.NewHTTPClient(nexusrpc.HTTPClientOptions{
		BaseURL:    "http://" + s.httpAddr + "/" + commonnexus.RouteDispatchNexusTaskByEndpoint.Path(s.endpointID),
		Service:    "fabric.certauth.v1.CertificateAuthority",
		Serializer: commonnexus.PayloadSerializer,
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(15 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		res, err := client.StartOperation(ctx, "SignCertificate", payload, nexus.StartOperationOptions{})
		cancel()
		var handlerErr *nexus.HandlerError
		if errors.As(err, &handlerErr) && handlerErr.Type == nexus.HandlerErrorTypeNotFound && time.Now().Before(deadline) {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return res, err
	}
}

type internodeClaims struct{}

func (internodeClaims) GetClaims(*authorization.AuthInfo) (*authorization.Claims, error) {
	return &authorization.Claims{Subject: "temporal-internode", System: authorization.RoleAdmin}, nil
}

// AuthInfoRequired lets the claims apply without a TLS connection.
func (internodeClaims) AuthInfoRequired() bool { return false }

// liteConfig is the single-process, in-memory SQLite configuration of
// go.temporal.io/server/temporaltest, which cannot take TopLevelModule, plus
// the frontend's HTTP port.
func liteConfig() *config.Config {
	frontendPort := freeport.MustGetFreePort()
	service := func(grpcPort int) config.Service {
		return config.Service{RPC: config.RPC{
			GRPCPort:        grpcPort,
			MembershipPort:  freeport.MustGetFreePort(),
			BindOnLocalHost: true,
		}}
	}
	frontend := service(frontendPort)
	frontend.RPC.HTTPPort = freeport.MustGetFreePort()

	return &config.Config{
		Global: config.Global{
			Membership: config.Membership{MaxJoinDuration: 30 * time.Second, BroadcastAddress: "127.0.0.1"},
		},
		Persistence: config.Persistence{
			DefaultStore:     sqliteplugin.PluginName,
			VisibilityStore:  sqliteplugin.PluginName,
			NumHistoryShards: 1,
			DataStores: map[string]config.DataStore{sqliteplugin.PluginName: {SQL: &config.SQL{
				PluginName:        sqliteplugin.PluginName,
				DatabaseName:      fmt.Sprintf("nexus-dispatch-%d", time.Now().UnixNano()),
				ConnectAttributes: map[string]string{"mode": "memory", "cache": "shared"},
			}}},
		},
		ClusterMetadata: &cluster.Config{
			FailoverVersionIncrement: 10,
			MasterClusterName:        "active",
			CurrentClusterName:       "active",
			ClusterInformation: map[string]cluster.ClusterInformation{"active": {
				Enabled:                true,
				InitialFailoverVersion: 1,
				RPCAddress:             fmt.Sprintf("127.0.0.1:%d", frontendPort),
			}},
		},
		DCRedirectionPolicy: config.DCRedirectionPolicy{Policy: "noop"},
		Services: map[string]config.Service{
			"frontend": frontend,
			"history":  service(freeport.MustGetFreePort()),
			"matching": service(freeport.MustGetFreePort()),
		},
		Archival: config.Archival{
			History:    config.HistoryArchival{State: "disabled"},
			Visibility: config.VisibilityArchival{State: "disabled"},
		},
		NamespaceDefaults: config.NamespaceDefaults{Archival: config.ArchivalNamespaceDefaults{
			History:    config.HistoryArchivalNamespaceDefaults{State: "disabled"},
			Visibility: config.VisibilityArchivalNamespaceDefaults{State: "disabled"},
		}},
		PublicClient: config.PublicClient{HostPort: fmt.Sprintf("127.0.0.1:%d", frontendPort)},
	}
}
