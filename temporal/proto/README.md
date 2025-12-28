# Table of Contents

- [aws.s3.v1](#aws-s3-v1)
  - Services
    - [aws.s3.v1.S3Provider](#aws-s3-v1-s3provider)
      - [Activities](#aws-s3-v1-s3provider-activities)
        - [aws.s3.v1.GetObject](#aws-s3-v1-getobject-activity)
        - [aws.s3.v1.PutObject](#aws-s3-v1-putobject-activity)
        - [aws.s3.v1.DeleteObject](#aws-s3-v1-deleteobject-activity)
  - Messages
    - [aws.s3.v1.S3DeleteOutput](#aws-s3-v1-s3deleteoutput)
    - [aws.s3.v1.S3Object](#aws-s3-v1-s3object)
    - [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)
- [cell.v1](#cell-v1)
  - Services
    - [cell.v1.Cell](#cell-v1-cell)
      - [Workflows](#cell-v1-cell-workflows)
        - [cell.v1.InitFabricShard](#cell-v1-initfabricshard-workflow)
  - Messages
    - [cell.v1.InitFabricShardInput](#cell-v1-initfabricshardinput)
    - [cell.v1.InitFabricShardOutput](#cell-v1-initfabricshardoutput)
- [cluster.fabric.v1](#cluster-fabric-v1)
  - Services
    - [cluster.fabric.v1.Cluster](#cluster-fabric-v1-cluster)
      - [Workflows](#cluster-fabric-v1-cluster-workflows)
        - [cluster.fabric.v1.CreateNamespace](#cluster-fabric-v1-createnamespace-workflow)
        - [cluster.fabric.v1.CreateServiceAccount](#cluster-fabric-v1-createserviceaccount-workflow)
        - [cluster.fabric.v1.CreateDeployment](#cluster-fabric-v1-createdeployment-workflow)
  - Messages
    - [cluster.fabric.v1.CreateDeploymentInput](#cluster-fabric-v1-createdeploymentinput)
    - [cluster.fabric.v1.CreateDeploymentOutput](#cluster-fabric-v1-createdeploymentoutput)
    - [cluster.fabric.v1.CreateNamespaceInput](#cluster-fabric-v1-createnamespaceinput)
    - [cluster.fabric.v1.CreateNamespaceOutput](#cluster-fabric-v1-createnamespaceoutput)
    - [cluster.fabric.v1.CreateServiceAccountInput](#cluster-fabric-v1-createserviceaccountinput)
    - [cluster.fabric.v1.CreateServiceAccountOutput](#cluster-fabric-v1-createserviceaccountoutput)
- [fabric.v1](#fabric-v1)
  - Services
    - [fabric.v1.Fabric](#fabric-v1-fabric)
      - [Workflows](#fabric-v1-fabric-workflows)
        - [fabric.v1.BootstrapShard](#fabric-v1-bootstrapshard-workflow)
        - [fabric.v1.BootstrapCertificateAuthorities](#fabric-v1-bootstrapcertificateauthorities-workflow)
        - [fabric.v1.DeployCertificateStore](#fabric-v1-deploycertificatestore-workflow)
        - [fabric.v1.DeployRootCertificateAuthority](#fabric-v1-deployrootcertificateauthority-workflow)
        - [fabric.v1.DeployCertificateAuthority](#fabric-v1-deploycertificateauthority-workflow)
  - Messages
    - [fabric.v1.BootstrapCertificateAuthoritiesInput](#fabric-v1-bootstrapcertificateauthoritiesinput)
    - [fabric.v1.BootstrapCertificateAuthoritiesOutput](#fabric-v1-bootstrapcertificateauthoritiesoutput)
    - [fabric.v1.BootstrapShardInput](#fabric-v1-bootstrapshardinput)
    - [fabric.v1.BootstrapShardOutput](#fabric-v1-bootstrapshardoutput)
    - [fabric.v1.DeployCertificateAuthorityInput](#fabric-v1-deploycertificateauthorityinput)
    - [fabric.v1.DeployCertificateAuthorityOutput](#fabric-v1-deploycertificateauthorityoutput)
    - [fabric.v1.DeployCertificateStoreInput](#fabric-v1-deploycertificatestoreinput)
    - [fabric.v1.DeployCertificateStoreOutput](#fabric-v1-deploycertificatestoreoutput)
    - [fabric.v1.DeployRootCertificateAuthorityInput](#fabric-v1-deployrootcertificateauthorityinput)
    - [fabric.v1.DeployRootCertificateAuthorityOutput](#fabric-v1-deployrootcertificateauthorityoutput)
- [helm.v1](#helm-v1)
  - Services
    - [helm.v1.Helm](#helm-v1-helm)
      - [Activities](#helm-v1-helm-activities)
        - [helm.v1.RenderChart](#helm-v1-renderchart-activity)
        - [helm.v1.RenderChartWithTemplate](#helm-v1-renderchartwithtemplate-activity)
  - Messages
    - [helm.v1.RenderChartInput](#helm-v1-renderchartinput)
    - [helm.v1.RenderChartOutput](#helm-v1-renderchartoutput)
    - [helm.v1.RenderChartOutput.FailuresEntry](#helm-v1-renderchartoutput-failuresentry)
    - [helm.v1.RenderChartOutput.ResourcesEntry](#helm-v1-renderchartoutput-resourcesentry)
    - [helm.v1.RenderChartWithTemplateInput](#helm-v1-renderchartwithtemplateinput)
    - [helm.v1.Resource](#helm-v1-resource)
- [kube.v1](#kube-v1)
  - Services
    - [kube.v1.Kubernetes](#kube-v1-kubernetes)
      - [Activities](#kube-v1-kubernetes-activities)
        - [kube.v1.Create](#kube-v1-create-activity)
        - [kube.v1.Update](#kube-v1-update-activity)
        - [kube.v1.Apply](#kube-v1-apply-activity)
        - [kube.v1.Delete](#kube-v1-delete-activity)
        - [kube.v1.GetStatus](#kube-v1-getstatus-activity)
  - Messages
    - [kube.v1.ApplyInput](#kube-v1-applyinput)
    - [kube.v1.ApplyInput.ResourcesEntry](#kube-v1-applyinput-resourcesentry)
    - [kube.v1.ApplyOutput](#kube-v1-applyoutput)
    - [kube.v1.ApplyOutput.ResultsEntry](#kube-v1-applyoutput-resultsentry)
    - [kube.v1.ApplyResult](#kube-v1-applyresult)
    - [kube.v1.ApplySummary](#kube-v1-applysummary)
    - [kube.v1.CreateInput](#kube-v1-createinput)
    - [kube.v1.CreateInput.ResourcesEntry](#kube-v1-createinput-resourcesentry)
    - [kube.v1.CreateOutput](#kube-v1-createoutput)
    - [kube.v1.CreateOutput.ResultsEntry](#kube-v1-createoutput-resultsentry)
    - [kube.v1.CreateResult](#kube-v1-createresult)
    - [kube.v1.CreateSummary](#kube-v1-createsummary)
    - [kube.v1.DeleteInput](#kube-v1-deleteinput)
    - [kube.v1.DeleteInput.ResourcesEntry](#kube-v1-deleteinput-resourcesentry)
    - [kube.v1.DeleteOutput](#kube-v1-deleteoutput)
    - [kube.v1.DeleteOutput.ResultsEntry](#kube-v1-deleteoutput-resultsentry)
    - [kube.v1.DeleteResult](#kube-v1-deleteresult)
    - [kube.v1.DeleteSummary](#kube-v1-deletesummary)
    - [kube.v1.GetStatusInput](#kube-v1-getstatusinput)
    - [kube.v1.GetStatusInput.ResourcesEntry](#kube-v1-getstatusinput-resourcesentry)
    - [kube.v1.GetStatusOutput](#kube-v1-getstatusoutput)
    - [kube.v1.GetStatusOutput.ResultsEntry](#kube-v1-getstatusoutput-resultsentry)
    - [kube.v1.Resource](#kube-v1-resource)
    - [kube.v1.ResourceStatus](#kube-v1-resourcestatus)
    - [kube.v1.UpdateInput](#kube-v1-updateinput)
    - [kube.v1.UpdateInput.ResourcesEntry](#kube-v1-updateinput-resourcesentry)
    - [kube.v1.UpdateOutput](#kube-v1-updateoutput)
    - [kube.v1.UpdateOutput.ResultsEntry](#kube-v1-updateoutput-resultsentry)
    - [kube.v1.UpdateResult](#kube-v1-updateresult)
    - [kube.v1.UpdateSummary](#kube-v1-updatesummary)
- [temporal.v1](#temporal-v1)
  - Services
    - [temporal.v1.Admin](#temporal-v1-admin)
      - [Workflows](#temporal-v1-admin-workflows)
        - [temporal.v1.CreateNamespace](#temporal-v1-createnamespace-workflow)
      - [Activities](#temporal-v1-admin-activities)
        - [temporal.v1.RegisterNamespace](#temporal-v1-registernamespace-activity)
        - [temporal.v1.DescribeNamespace](#temporal-v1-describenamespace-activity)
        - [temporal.v1.CreateNexusEndpoint](#temporal-v1-createnexusendpoint-activity)
        - [temporal.v1.DescribeNexusEndpoint](#temporal-v1-describenexusendpoint-activity)
        - [temporal.v1.UpdateNexusEndpoint](#temporal-v1-updatenexusendpoint-activity)
  - Messages
    - [temporal.v1.CreateNamespaceInput](#temporal-v1-createnamespaceinput)
    - [temporal.v1.CreateNamespaceOutput](#temporal-v1-createnamespaceoutput)
    - [temporal.v1.CreateNexusEndpointInput](#temporal-v1-createnexusendpointinput)
    - [temporal.v1.CreateNexusEndpointOutput](#temporal-v1-createnexusendpointoutput)
    - [temporal.v1.DescribeNamespaceInput](#temporal-v1-describenamespaceinput)
    - [temporal.v1.DescribeNamespaceOutput](#temporal-v1-describenamespaceoutput)
    - [temporal.v1.DescribeNexusEndpointInput](#temporal-v1-describenexusendpointinput)
    - [temporal.v1.DescribeNexusEndpointOutput](#temporal-v1-describenexusendpointoutput)
    - [temporal.v1.Endpoint](#temporal-v1-endpoint)
    - [temporal.v1.NamespaceState](#temporal-v1-namespacestate)
    - [temporal.v1.RegisterNamespaceInput](#temporal-v1-registernamespaceinput)
    - [temporal.v1.RegisterNamespaceOutput](#temporal-v1-registernamespaceoutput)
    - [temporal.v1.UpdateNexusEndpointInput](#temporal-v1-updatenexusendpointinput)
    - [temporal.v1.UpdateNexusEndpointOutput](#temporal-v1-updatenexusendpointoutput)
- [vault.v1](#vault-v1)
  - Services
    - [vault.v1.Vault](#vault-v1-vault)
      - [Workflows](#vault-v1-vault-workflows)
        - [vault.v1.CreateAccess](#vault-v1-createaccess-workflow)
      - [Activities](#vault-v1-vault-activities)
        - [vault.v1.CreateAuthRole](#vault-v1-createauthrole-activity)
        - [vault.v1.DeleteAuthRole](#vault-v1-deleteauthrole-activity)
        - [vault.v1.CreatePolicy](#vault-v1-createpolicy-activity)
        - [vault.v1.DeletePolicy](#vault-v1-deletepolicy-activity)
        - [vault.v1.StoreSecret](#vault-v1-storesecret-activity)
        - [vault.v1.FetchSecret](#vault-v1-fetchsecret-activity)
        - [vault.v1.DeleteSecret](#vault-v1-deletesecret-activity)
  - Messages
    - [vault.v1.Capability](#vault-v1-capability)
    - [vault.v1.CreateAccessInput](#vault-v1-createaccessinput)
    - [vault.v1.CreateAccessOutput](#vault-v1-createaccessoutput)
    - [vault.v1.CreateAuthRoleInput](#vault-v1-createauthroleinput)
    - [vault.v1.CreateAuthRoleOutput](#vault-v1-createauthroleoutput)
    - [vault.v1.CreatePolicyInput](#vault-v1-createpolicyinput)
    - [vault.v1.CreatePolicyOutput](#vault-v1-createpolicyoutput)
    - [vault.v1.DeleteAuthRoleInput](#vault-v1-deleteauthroleinput)
    - [vault.v1.DeleteAuthRoleOutput](#vault-v1-deleteauthroleoutput)
    - [vault.v1.DeletePolicyInput](#vault-v1-deletepolicyinput)
    - [vault.v1.DeletePolicyOutput](#vault-v1-deletepolicyoutput)
    - [vault.v1.DeleteSecretInput](#vault-v1-deletesecretinput)
    - [vault.v1.DeleteSecretOutput](#vault-v1-deletesecretoutput)
    - [vault.v1.FetchSecretInput](#vault-v1-fetchsecretinput)
    - [vault.v1.FetchSecretOutput](#vault-v1-fetchsecretoutput)
    - [vault.v1.FetchSecretOutput.DataEntry](#vault-v1-fetchsecretoutput-dataentry)
    - [vault.v1.PolicyDocument](#vault-v1-policydocument)
    - [vault.v1.PolicyRule](#vault-v1-policyrule)
    - [vault.v1.SecretRef](#vault-v1-secretref)
    - [vault.v1.StoreSecretInput](#vault-v1-storesecretinput)
    - [vault.v1.StoreSecretInput.DataEntry](#vault-v1-storesecretinput-dataentry)
    - [vault.v1.StoreSecretOutput](#vault-v1-storesecretoutput)
- [google.protobuf](#google-protobuf)
  - Messages
    - [google.protobuf.Duration](#google-protobuf-duration)
    - [google.protobuf.ListValue](#google-protobuf-listvalue)
    - [google.protobuf.NullValue](#google-protobuf-nullvalue)
    - [google.protobuf.Struct](#google-protobuf-struct)
    - [google.protobuf.Struct.FieldsEntry](#google-protobuf-struct-fieldsentry)
    - [google.protobuf.Value](#google-protobuf-value)

<a name="aws-s3-v1"></a>
# aws.s3.v1

<a name="aws-s3-v1-services"></a>
## Services

<a name="aws-s3-v1-s3provider"></a>
## aws.s3.v1.S3Provider

<pre>
S3Provider service defines activities for interacting with AWS S3
</pre>   

<a name="aws-s3-v1-s3provider-activities"></a>
### Activities

---
<a name="aws-s3-v1-getobject-activity"></a>
### aws.s3.v1.GetObject

<pre>
GetObject retrieves an object from S3
</pre>

**Input:** [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table> 

---
<a name="aws-s3-v1-putobject-activity"></a>
### aws.s3.v1.PutObject

<pre>
PutObject writes an object to S3
</pre>

**Input:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3Object](#aws-s3-v1-s3object)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table> 

---
<a name="aws-s3-v1-deleteobject-activity"></a>
### aws.s3.v1.DeleteObject

<pre>
DeleteObject deletes an object from S3
</pre>

**Input:** [aws.s3.v1.S3ObjectRef](#aws-s3-v1-s3objectref)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>

**Output:** [aws.s3.v1.S3DeleteOutput](#aws-s3-v1-s3deleteoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>delete_marker</td>
<td>bool</td>
<td><pre>
delete_marker indicates if a delete marker was created (for versioned buckets)<br>

json_name: deleteMarker
go_name: DeleteMarker</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>   

<a name="aws-s3-v1-messages"></a>
## Messages

<a name="aws-s3-v1-s3deleteoutput"></a>
### aws.s3.v1.S3DeleteOutput

<pre>
S3DeleteOutput represents the result of a delete operation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>delete_marker</td>
<td>bool</td>
<td><pre>
delete_marker indicates if a delete marker was created (for versioned buckets)<br>

json_name: deleteMarker
go_name: DeleteMarker</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>



<a name="aws-s3-v1-s3object"></a>
### aws.s3.v1.S3Object

<pre>
S3Object represents a complete S3 object with metadata
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: contentType
go_name: ContentType</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: etag
go_name: Etag</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: versionId
go_name: VersionId</pre></td>
</tr>
</table>



<a name="aws-s3-v1-s3objectref"></a>
### aws.s3.v1.S3ObjectRef

<pre>
S3ObjectRef references an S3 object without its data
Used for operations that only need bucket and key (read, delete)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bucket</td>
<td>string</td>
<td><pre>
bucket is the S3 bucket name<br>

json_name: bucket
go_name: Bucket</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: key
go_name: Key</pre></td>
</tr>
</table>



<a name="cell-v1"></a>
# cell.v1

<a name="cell-v1-services"></a>
## Services

<a name="cell-v1-cell"></a>
## cell.v1.Cell

<pre>
Cell service defines the root workflow for initializing a Fabric shard (cluster/cell)
This is the entry point for multi-tenant infrastructure provisioning
</pre>

<a name="cell-v1-cell-workflows"></a>
### Workflows

---
<a name="cell-v1-initfabricshard-workflow"></a>
### cell.v1.InitFabricShard

<pre>
InitFabricShard orchestrates the complete initialization of a Fabric shard
This root workflow executes in the central "infra" namespace and coordinates:
1. Creation of a Temporal namespace for the shard (temporal.v1.CreateNamespace)
2. Creation of a Kubernetes namespace for the shard (cluster.v1.CreateNamespace)
3. Creation of a service account for the fabric worker (cluster.v1.CreateServiceAccount)
4. Creation of Vault access policies for the shard (vault.v1.CreateAccess)
5. Deployment of the fabric worker pod (cluster.v1.CreateDeployment)
6. Registration of the Nexus endpoint for fabric.v1.BootstrapShard service
7. Cross-namespace Nexus call to BootstrapShard workflow in fabric namespace

This workflow is the "hub" that orchestrates the multi-phase bootstrap process.
Phase 1 (central): Sets up infrastructure prerequisites
Phase 2 (fabric worker): Initiated via Nexus, deploys Fabric components (CAs, peers, orderers)
</pre>

**Input:** [cell.v1.InitFabricShardInput](#cell-v1-initfabricshardinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1", "acme-corp")
Used to derive namespace names: Temporal namespace "fabric-{org}", K8s namespace "fabric-{org}"<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>

**Output:** [cell.v1.InitFabricShardOutput](#cell-v1-initfabricshardoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>     

<a name="cell-v1-messages"></a>
## Messages

<a name="cell-v1-initfabricshardinput"></a>
### cell.v1.InitFabricShardInput

<pre>
InitFabricShardInput specifies the parameters for initializing a Fabric shard
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1", "acme-corp")
Used to derive namespace names: Temporal namespace "fabric-{org}", K8s namespace "fabric-{org}"<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>



<a name="cell-v1-initfabricshardoutput"></a>
### cell.v1.InitFabricShardOutput

<pre>
InitFabricShardOutput indicates the result of shard initialization
</pre>



<a name="cluster-fabric-v1"></a>
# cluster.fabric.v1

<a name="cluster-fabric-v1-services"></a>
## Services

<a name="cluster-fabric-v1-cluster"></a>
## cluster.fabric.v1.Cluster

<pre>
Cluster defines workflows for infrastructure provisioning on Kubernetes
These workflows orchestrate the creation and management of cluster resources
by composing Helm (for templating) and Kubernetes (for deployment) activities
</pre>

<a name="cluster-fabric-v1-cluster-workflows"></a>
### Workflows

---
<a name="cluster-fabric-v1-createnamespace-workflow"></a>
### cluster.fabric.v1.CreateNamespace

<pre>
CreateNamespace orchestrates the creation of a Kubernetes namespace
This workflow combines Helm chart rendering with Kubernetes resource application
</pre>

**Input:** [cluster.fabric.v1.CreateNamespaceInput](#cluster-fabric-v1-createnamespaceinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
The namespace will be named "fabric-{org}"<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>

**Output:** [cluster.fabric.v1.CreateNamespaceOutput](#cluster-fabric-v1-createnamespaceoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace_name</td>
<td>string</td>
<td><pre>
namespace_name is the created namespace (e.g., "fabric-org1")<br>

json_name: namespaceName
go_name: NamespaceName</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="cluster-fabric-v1-createserviceaccount-workflow"></a>
### cluster.fabric.v1.CreateServiceAccount

<pre>
CreateServiceAccount orchestrates the creation of a Kubernetes service account
This workflow renders a Helm chart template for service account configuration,
then applies the resulting resources to the cluster
</pre>

**Input:** [cluster.fabric.v1.CreateServiceAccountInput](#cluster-fabric-v1-createserviceaccountinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
json_name: org
go_name: Org</pre></td>
</tr>
</table>

**Output:** [cluster.fabric.v1.CreateServiceAccountOutput](#cluster-fabric-v1-createserviceaccountoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="cluster-fabric-v1-createdeployment-workflow"></a>
### cluster.fabric.v1.CreateDeployment

<pre>
CreateDeployment orchestrates the deployment of a single resource to Kubernetes
This workflow renders a Helm chart with provided values and applies it to the cluster
</pre>

**Input:** [cluster.fabric.v1.CreateDeploymentInput](#cluster-fabric-v1-createdeploymentinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>component</td>
<td>string</td>
<td><pre>
component specifies which type of deployment to create
Currently supported: "worker", "certificate-store"
Future components: "peer", "orderer", "ca", "couchdb"<br>

json_name: component
go_name: Component</pre></td>
</tr><tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
Used to derive namespace names, service account names, etc.<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>

**Output:** [cluster.fabric.v1.CreateDeploymentOutput](#cluster-fabric-v1-createdeploymentoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>deployment_name</td>
<td>string</td>
<td><pre>
deployment_name is the name of the created deployment<br>

json_name: deploymentName
go_name: DeploymentName</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>     

<a name="cluster-fabric-v1-messages"></a>
## Messages

<a name="cluster-fabric-v1-createdeploymentinput"></a>
### cluster.fabric.v1.CreateDeploymentInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>component</td>
<td>string</td>
<td><pre>
component specifies which type of deployment to create
Currently supported: "worker", "certificate-store"
Future components: "peer", "orderer", "ca", "couchdb"<br>

json_name: component
go_name: Component</pre></td>
</tr><tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
Used to derive namespace names, service account names, etc.<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>



<a name="cluster-fabric-v1-createdeploymentoutput"></a>
### cluster.fabric.v1.CreateDeploymentOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>deployment_name</td>
<td>string</td>
<td><pre>
deployment_name is the name of the created deployment<br>

json_name: deploymentName
go_name: DeploymentName</pre></td>
</tr>
</table>



<a name="cluster-fabric-v1-createnamespaceinput"></a>
### cluster.fabric.v1.CreateNamespaceInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
The namespace will be named "fabric-{org}"<br>

json_name: org
go_name: Org</pre></td>
</tr>
</table>



<a name="cluster-fabric-v1-createnamespaceoutput"></a>
### cluster.fabric.v1.CreateNamespaceOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace_name</td>
<td>string</td>
<td><pre>
namespace_name is the created namespace (e.g., "fabric-org1")<br>

json_name: namespaceName
go_name: NamespaceName</pre></td>
</tr>
</table>



<a name="cluster-fabric-v1-createserviceaccountinput"></a>
### cluster.fabric.v1.CreateServiceAccountInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>org</td>
<td>string</td>
<td><pre>
json_name: org
go_name: Org</pre></td>
</tr>
</table>



<a name="cluster-fabric-v1-createserviceaccountoutput"></a>
### cluster.fabric.v1.CreateServiceAccountOutput



<a name="fabric-v1"></a>
# fabric.v1

<a name="fabric-v1-services"></a>
## Services

<a name="fabric-v1-fabric"></a>
## fabric.v1.Fabric

<pre>
Fabric service defines workflows for manipulating a Fabric shard
These workflows run on the fabric worker in the org-specific Temporal namespace
</pre>

<a name="fabric-v1-fabric-workflows"></a>
### Workflows

---
<a name="fabric-v1-bootstrapshard-workflow"></a>
### fabric.v1.BootstrapShard

<pre>
BootstrapShard is the entry point for Phase 2 of shard initialization
This workflow runs on the fabric worker in the org's Temporal namespace (e.g., "fabric-org1")
and deploys all Fabric components (CAs, peers, orderers) for the organization.
Called via Nexus from the central worker after Phase 1 completes
</pre>

**Input:** [fabric.v1.BootstrapShardInput](#fabric-v1-bootstrapshardinput)



**Output:** [fabric.v1.BootstrapShardOutput](#fabric-v1-bootstrapshardoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/bootstrap-shard</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-bootstrapcertificateauthorities-workflow"></a>
### fabric.v1.BootstrapCertificateAuthorities

<pre>
BootstrapCertificateAuthorities orchestrates the deployment of all certificate authorities
for a Fabric organization. This includes the certificate store, root CA, and intermediate CAs.
</pre>

**Input:** [fabric.v1.BootstrapCertificateAuthoritiesInput](#fabric-v1-bootstrapcertificateauthoritiesinput)



**Output:** [fabric.v1.BootstrapCertificateAuthoritiesOutput](#fabric-v1-bootstrapcertificateauthoritiesoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/bootstrap-certificate-authorities</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploycertificatestore-workflow"></a>
### fabric.v1.DeployCertificateStore

<pre>
DeployCertificateStore deploys the shared certificate store (PostgreSQL)
used by all certificate authorities in the organization.
</pre>

**Input:** [fabric.v1.DeployCertificateStoreInput](#fabric-v1-deploycertificatestoreinput)



**Output:** [fabric.v1.DeployCertificateStoreOutput](#fabric-v1-deploycertificatestoreoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-certificate-store</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deployrootcertificateauthority-workflow"></a>
### fabric.v1.DeployRootCertificateAuthority

<pre>
DeployRootCertificateAuthority deploys the root certificate authority (root CA)
for the Fabric organization. The root CA is the top-level CA in the PKI hierarchy
and is used to issue intermediate CAs for org and TLS identities.
</pre>

**Input:** [fabric.v1.DeployRootCertificateAuthorityInput](#fabric-v1-deployrootcertificateauthorityinput)



**Output:** [fabric.v1.DeployRootCertificateAuthorityOutput](#fabric-v1-deployrootcertificateauthorityoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-root-certificate-authority</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploycertificateauthority-workflow"></a>
### fabric.v1.DeployCertificateAuthority

<pre>
DeployCertificateAuthority deploys an intermediate certificate authority (intermediate CA)
for the Fabric organization. Intermediate CAs are deployed for both "org" and "tls" purposes.
</pre>

**Input:** [fabric.v1.DeployCertificateAuthorityInput](#fabric-v1-deploycertificateauthorityinput)



**Output:** [fabric.v1.DeployCertificateAuthorityOutput](#fabric-v1-deploycertificateauthorityoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-certificate-authority</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>     

<a name="fabric-v1-messages"></a>
## Messages

<a name="fabric-v1-bootstrapcertificateauthoritiesinput"></a>
### fabric.v1.BootstrapCertificateAuthoritiesInput

<pre>
BootstrapCertificateAuthoritiesInput specifies the parameters for bootstrapping certificate authorities
Empty for now - org is implicit from the worker's namespace
</pre>



<a name="fabric-v1-bootstrapcertificateauthoritiesoutput"></a>
### fabric.v1.BootstrapCertificateAuthoritiesOutput

<pre>
BootstrapCertificateAuthoritiesOutput indicates the result of CA bootstrapping
Empty for now - will be populated as implementation progresses
</pre>



<a name="fabric-v1-bootstrapshardinput"></a>
### fabric.v1.BootstrapShardInput

<pre>
BootstrapShardInput specifies the parameters for bootstrapping a Fabric shard
Empty for now - will be populated as implementation progresses
</pre>



<a name="fabric-v1-bootstrapshardoutput"></a>
### fabric.v1.BootstrapShardOutput

<pre>
BootstrapShardOutput indicates the result of shard bootstrapping
Empty for now - will be populated as implementation progresses
</pre>



<a name="fabric-v1-deploycertificateauthorityinput"></a>
### fabric.v1.DeployCertificateAuthorityInput

<pre>
DeployCertificateAuthorityInput specifies the parameters for deploying an intermediate CA
</pre>



<a name="fabric-v1-deploycertificateauthorityoutput"></a>
### fabric.v1.DeployCertificateAuthorityOutput

<pre>
DeployCertificateAuthorityOutput indicates the result of intermediate CA deployment
Empty for now - will be populated as implementation progresses
</pre>



<a name="fabric-v1-deploycertificatestoreinput"></a>
### fabric.v1.DeployCertificateStoreInput

<pre>
DeployCertificateStoreInput specifies the parameters for deploying the certificate store
Empty for now - org is implicit from the worker's namespace
</pre>



<a name="fabric-v1-deploycertificatestoreoutput"></a>
### fabric.v1.DeployCertificateStoreOutput

<pre>
DeployCertificateStoreOutput indicates the result of certificate store deployment
Empty for now - will be populated as implementation progresses
</pre>



<a name="fabric-v1-deployrootcertificateauthorityinput"></a>
### fabric.v1.DeployRootCertificateAuthorityInput

<pre>
DeployRootCertificateAuthorityInput specifies the parameters for deploying the root CA
</pre>



<a name="fabric-v1-deployrootcertificateauthorityoutput"></a>
### fabric.v1.DeployRootCertificateAuthorityOutput

<pre>
DeployRootCertificateAuthorityOutput indicates the result of root CA deployment
Empty for now - will be populated as implementation progresses
</pre>



<a name="helm-v1"></a>
# helm.v1

<a name="helm-v1-services"></a>
## Services

<a name="helm-v1-helm"></a>
## helm.v1.Helm

<pre>
Helm service defines activities for rendering Helm charts into Kubernetes manifests
Activities parse the rendered YAML and return individual resources
</pre>   

<a name="helm-v1-helm-activities"></a>
### Activities

---
<a name="helm-v1-renderchart-activity"></a>
### helm.v1.RenderChart

<pre>
RenderChart renders a Helm chart with given values and returns parsed resources
</pre>

**Input:** [helm.v1.RenderChartInput](#helm-v1-renderchartinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chart_name</td>
<td>string</td>
<td><pre>
chart_name is the chart name to render<br>

json_name: chartName
go_name: ChartName</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>values_yaml</td>
<td>string</td>
<td><pre>
values_yaml is Helm values as a YAML string (merged with chart defaults)<br>

json_name: valuesYaml
go_name: ValuesYaml</pre></td>
</tr>
</table>

**Output:** [helm.v1.RenderChartOutput](#helm-v1-renderchartoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>failures</td>
<td><a href="#helm-v1-renderchartoutput-failuresentry">helm.v1.RenderChartOutput.FailuresEntry</a></td>
<td><pre>
failures is a map of resource_id → error message for resources that failed to parse<br>

json_name: failures
go_name: Failures</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table> 

---
<a name="helm-v1-renderchartwithtemplate-activity"></a>
### helm.v1.RenderChartWithTemplate

<pre>
RenderChartWithTemplate renders a Helm chart after populating a values template
</pre>

**Input:** [helm.v1.RenderChartWithTemplateInput](#helm-v1-renderchartwithtemplateinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chart_name</td>
<td>string</td>
<td><pre>
chart_name is the chart name to render<br>

json_name: chartName
go_name: ChartName</pre></td>
</tr><tr>
<td>component</td>
<td>string</td>
<td><pre>
component is the component type whose values template to load and render<br>

json_name: component
go_name: Component</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>values</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
values is a structured map of template variables for rendering<br>

json_name: values
go_name: Values</pre></td>
</tr>
</table>

**Output:** [helm.v1.RenderChartOutput](#helm-v1-renderchartoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>failures</td>
<td><a href="#helm-v1-renderchartoutput-failuresentry">helm.v1.RenderChartOutput.FailuresEntry</a></td>
<td><pre>
failures is a map of resource_id → error message for resources that failed to parse<br>

json_name: failures
go_name: Failures</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>   

<a name="helm-v1-messages"></a>
## Messages

<a name="helm-v1-renderchartinput"></a>
### helm.v1.RenderChartInput

<pre>
RenderChartInput specifies which chart to render and with what values
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chart_name</td>
<td>string</td>
<td><pre>
chart_name is the chart name to render<br>

json_name: chartName
go_name: ChartName</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>values_yaml</td>
<td>string</td>
<td><pre>
values_yaml is Helm values as a YAML string (merged with chart defaults)<br>

json_name: valuesYaml
go_name: ValuesYaml</pre></td>
</tr>
</table>



<a name="helm-v1-renderchartoutput"></a>
### helm.v1.RenderChartOutput

<pre>
RenderChartOutput represents the rendered resources from a chart
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>failures</td>
<td><a href="#helm-v1-renderchartoutput-failuresentry">helm.v1.RenderChartOutput.FailuresEntry</a></td>
<td><pre>
failures is a map of resource_id → error message for resources that failed to parse<br>

json_name: failures
go_name: Failures</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="helm-v1-renderchartoutput-failuresentry"></a>
### helm.v1.RenderChartOutput.FailuresEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td>string</td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="helm-v1-renderchartoutput-resourcesentry"></a>
### helm.v1.RenderChartOutput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#helm-v1-resource">helm.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="helm-v1-renderchartwithtemplateinput"></a>
### helm.v1.RenderChartWithTemplateInput

<pre>
RenderChartWithTemplateInput specifies a chart to render after populating a values template
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chart_name</td>
<td>string</td>
<td><pre>
chart_name is the chart name to render<br>

json_name: chartName
go_name: ChartName</pre></td>
</tr><tr>
<td>component</td>
<td>string</td>
<td><pre>
component is the component type whose values template to load and render<br>

json_name: component
go_name: Component</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>values</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
values is a structured map of template variables for rendering<br>

json_name: values
go_name: Values</pre></td>
</tr>
</table>



<a name="helm-v1-resource"></a>
### helm.v1.Resource

<pre>
Resource represents a single rendered Kubernetes resource as YAML
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>manifest_yaml</td>
<td>bytes</td>
<td><pre>
manifest_yaml is the complete YAML definition of the resource<br>

json_name: manifestYaml
go_name: ManifestYaml</pre></td>
</tr>
</table>



<a name="kube-v1"></a>
# kube.v1

<a name="kube-v1-services"></a>
## Services

<a name="kube-v1-kubernetes"></a>
## kube.v1.Kubernetes

<pre>
Kubernetes defines activities for deploying resources to Kubernetes
Activities work with pre-rendered YAML manifests (from Helm, Kustomize, or plain YAML)
Caller is responsible for templating/rendering before calling these activities
</pre>   

<a name="kube-v1-kubernetes-activities"></a>
### Activities

---
<a name="kube-v1-create-activity"></a>
### kube.v1.Create

<pre>
Create creates new resources in the cluster (fails if already exists)
</pre>

**Input:** [kube.v1.CreateInput](#kube-v1-createinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-createinput-resourcesentry">kube.v1.CreateInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>

**Output:** [kube.v1.CreateOutput](#kube-v1-createoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-createoutput-resultsentry">kube.v1.CreateOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from CreateInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-createsummary">kube.v1.CreateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table> 

---
<a name="kube-v1-update-activity"></a>
### kube.v1.Update

<pre>
Update updates existing resources in the cluster (fails if not found)
</pre>

**Input:** [kube.v1.UpdateInput](#kube-v1-updateinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-updateinput-resourcesentry">kube.v1.UpdateInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>

**Output:** [kube.v1.UpdateOutput](#kube-v1-updateoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-updateoutput-resultsentry">kube.v1.UpdateOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from UpdateInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-updatesummary">kube.v1.UpdateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table> 

---
<a name="kube-v1-apply-activity"></a>
### kube.v1.Apply

<pre>
Apply deploys resources to the cluster (create or update)
</pre>

**Input:** [kube.v1.ApplyInput](#kube-v1-applyinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-applyinput-resourcesentry">kube.v1.ApplyInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>

**Output:** [kube.v1.ApplyOutput](#kube-v1-applyoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-applyoutput-resultsentry">kube.v1.ApplyOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from ApplyInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-applysummary">kube.v1.ApplySummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table> 

---
<a name="kube-v1-delete-activity"></a>
### kube.v1.Delete

<pre>
Delete removes resources from the cluster
</pre>

**Input:** [kube.v1.DeleteInput](#kube-v1-deleteinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-deleteinput-resourcesentry">kube.v1.DeleteInput.ResourcesEntry</a></td>
<td><pre>
resources to delete
Keys are resource_id (same as ApplyInput)
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>

**Output:** [kube.v1.DeleteOutput](#kube-v1-deleteoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-deleteoutput-resultsentry">kube.v1.DeleteOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-deletesummary">kube.v1.DeleteSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table> 

---
<a name="kube-v1-getstatus-activity"></a>
### kube.v1.GetStatus

<pre>
GetStatus retrieves the current status of deployed resources
</pre>

**Input:** [kube.v1.GetStatusInput](#kube-v1-getstatusinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-getstatusinput-resourcesentry">kube.v1.GetStatusInput.ResourcesEntry</a></td>
<td><pre>
resources to check status for
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>

**Output:** [kube.v1.GetStatusOutput](#kube-v1-getstatusoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>overall_status</td>
<td>string</td>
<td><pre>
overall_status is a summary: "ready", "pending", "degraded", "error", "unknown"<br>

json_name: overallStatus
go_name: OverallStatus</pre></td>
</tr><tr>
<td>results</td>
<td><a href="#kube-v1-getstatusoutput-resultsentry">kube.v1.GetStatusOutput.ResultsEntry</a></td>
<td><pre>
results contains status for each resource
Keys match resource_id from GetStatusInput<br>

json_name: results
go_name: Results</pre></td>
</tr>
</table>   

<a name="kube-v1-messages"></a>
## Messages

<a name="kube-v1-applyinput"></a>
### kube.v1.ApplyInput

<pre>
ApplyInput defines inputs for the Apply activity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-applyinput-resourcesentry">kube.v1.ApplyInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="kube-v1-applyinput-resourcesentry"></a>
### kube.v1.ApplyInput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resource">kube.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-applyoutput"></a>
### kube.v1.ApplyOutput

<pre>
ApplyOutput represents the result of applying resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-applyoutput-resultsentry">kube.v1.ApplyOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from ApplyInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-applysummary">kube.v1.ApplySummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table>



<a name="kube-v1-applyoutput-resultsentry"></a>
### kube.v1.ApplyOutput.ResultsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-applyresult">kube.v1.ApplyResult</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-applyresult"></a>
### kube.v1.ApplyResult

<pre>
ApplyResult represents the outcome of applying a single resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>error_message</td>
<td>string</td>
<td><pre>
error_message contains error details (if failed)<br>

json_name: errorMessage
go_name: ErrorMessage</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: resourceId
go_name: ResourceId</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if applied successfully)<br>

json_name: resourceStatus
go_name: ResourceStatus</pre></td>
</tr>
</table>



<a name="kube-v1-applysummary"></a>
### kube.v1.ApplySummary

<pre>
ApplySummary provides aggregate statistics
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>applied</td>
<td>int32</td>
<td><pre>
json_name: applied
go_name: Applied</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: failed
go_name: Failed</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: total
go_name: Total</pre></td>
</tr>
</table>



<a name="kube-v1-createinput"></a>
### kube.v1.CreateInput

<pre>
CreateInput defines inputs for the Create activity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-createinput-resourcesentry">kube.v1.CreateInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="kube-v1-createinput-resourcesentry"></a>
### kube.v1.CreateInput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resource">kube.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-createoutput"></a>
### kube.v1.CreateOutput

<pre>
CreateOutput represents the result of creating resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-createoutput-resultsentry">kube.v1.CreateOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from CreateInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-createsummary">kube.v1.CreateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table>



<a name="kube-v1-createoutput-resultsentry"></a>
### kube.v1.CreateOutput.ResultsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-createresult">kube.v1.CreateResult</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-createresult"></a>
### kube.v1.CreateResult

<pre>
CreateResult represents the outcome of creating a single resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>error_message</td>
<td>string</td>
<td><pre>
error_message contains error details (if failed)<br>

json_name: errorMessage
go_name: ErrorMessage</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: resourceId
go_name: ResourceId</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if created successfully)<br>

json_name: resourceStatus
go_name: ResourceStatus</pre></td>
</tr>
</table>



<a name="kube-v1-createsummary"></a>
### kube.v1.CreateSummary

<pre>
CreateSummary provides aggregate statistics
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>created</td>
<td>int32</td>
<td><pre>
json_name: created
go_name: Created</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: failed
go_name: Failed</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: total
go_name: Total</pre></td>
</tr>
</table>



<a name="kube-v1-deleteinput"></a>
### kube.v1.DeleteInput

<pre>
DeleteInput defines inputs for the Delete activity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-deleteinput-resourcesentry">kube.v1.DeleteInput.ResourcesEntry</a></td>
<td><pre>
resources to delete
Keys are resource_id (same as ApplyInput)
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="kube-v1-deleteinput-resourcesentry"></a>
### kube.v1.DeleteInput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resource">kube.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-deleteoutput"></a>
### kube.v1.DeleteOutput

<pre>
DeleteOutput represents the result of deleting resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-deleteoutput-resultsentry">kube.v1.DeleteOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-deletesummary">kube.v1.DeleteSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table>



<a name="kube-v1-deleteoutput-resultsentry"></a>
### kube.v1.DeleteOutput.ResultsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-deleteresult">kube.v1.DeleteResult</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-deleteresult"></a>
### kube.v1.DeleteResult

<pre>
DeleteResult represents the outcome of deleting a single resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>error_message</td>
<td>string</td>
<td><pre>
error_message contains error details (if failed)<br>

json_name: errorMessage
go_name: ErrorMessage</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: resourceId
go_name: ResourceId</pre></td>
</tr>
</table>



<a name="kube-v1-deletesummary"></a>
### kube.v1.DeleteSummary

<pre>
DeleteSummary provides aggregate statistics
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>deleted</td>
<td>int32</td>
<td><pre>
json_name: deleted
go_name: Deleted</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: failed
go_name: Failed</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: total
go_name: Total</pre></td>
</tr>
</table>



<a name="kube-v1-getstatusinput"></a>
### kube.v1.GetStatusInput

<pre>
GetStatusInput queries the status of deployed resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-getstatusinput-resourcesentry">kube.v1.GetStatusInput.ResourcesEntry</a></td>
<td><pre>
resources to check status for
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="kube-v1-getstatusinput-resourcesentry"></a>
### kube.v1.GetStatusInput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resource">kube.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-getstatusoutput"></a>
### kube.v1.GetStatusOutput

<pre>
GetStatusOutput represents the current status of resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>overall_status</td>
<td>string</td>
<td><pre>
overall_status is a summary: "ready", "pending", "degraded", "error", "unknown"<br>

json_name: overallStatus
go_name: OverallStatus</pre></td>
</tr><tr>
<td>results</td>
<td><a href="#kube-v1-getstatusoutput-resultsentry">kube.v1.GetStatusOutput.ResultsEntry</a></td>
<td><pre>
results contains status for each resource
Keys match resource_id from GetStatusInput<br>

json_name: results
go_name: Results</pre></td>
</tr>
</table>



<a name="kube-v1-getstatusoutput-resultsentry"></a>
### kube.v1.GetStatusOutput.ResultsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-resource"></a>
### kube.v1.Resource

<pre>
Resource represents a single Kubernetes resource as YAML
Activities parse the YAML to extract metadata before applying
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>manifest_yaml</td>
<td>bytes</td>
<td><pre>
manifest_yaml is the complete YAML definition of the resource
Activities parse this to extract kind, name, namespace, apiVersion<br>

json_name: manifestYaml
go_name: ManifestYaml</pre></td>
</tr>
</table>



<a name="kube-v1-resourcestatus"></a>
### kube.v1.ResourceStatus

<pre>
ResourceStatus represents the status of a single Kubernetes resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>desired_replicas</td>
<td>int32</td>
<td><pre>
desired_replicas is the desired number of replicas
Only populated for workload resources<br>

json_name: desiredReplicas
go_name: DesiredReplicas</pre></td>
</tr><tr>
<td>kind</td>
<td>string</td>
<td><pre>
kind is the Kubernetes resource kind (e.g., "Deployment", "Pod", "Service")<br>

json_name: kind
go_name: Kind</pre></td>
</tr><tr>
<td>message</td>
<td>string</td>
<td><pre>
message provides additional context (error messages, warnings, etc.)<br>

json_name: message
go_name: Message</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the resource name<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the resource namespace<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>ready_replicas</td>
<td>int32</td>
<td><pre>
ready_replicas is the number of ready replicas (for Deployments, StatefulSets, etc.)
Only populated for workload resources<br>

json_name: readyReplicas
go_name: ReadyReplicas</pre></td>
</tr><tr>
<td>status</td>
<td>string</td>
<td><pre>
status is the current status (e.g., "ready", "pending", "error", "unknown")<br>

json_name: status
go_name: Status</pre></td>
</tr>
</table>



<a name="kube-v1-updateinput"></a>
### kube.v1.UpdateInput

<pre>
UpdateInput defines inputs for the Update activity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources</td>
<td><a href="#kube-v1-updateinput-resourcesentry">kube.v1.UpdateInput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → resource
resource_id is caller-defined for tracking (e.g., "deployment-api", "service-api")
Each resource's YAML must include namespace (or be cluster-scoped)<br>

json_name: resources
go_name: Resources</pre></td>
</tr>
</table>



<a name="kube-v1-updateinput-resourcesentry"></a>
### kube.v1.UpdateInput.ResourcesEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-resource">kube.v1.Resource</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-updateoutput"></a>
### kube.v1.UpdateOutput

<pre>
UpdateOutput represents the result of updating resources
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>results</td>
<td><a href="#kube-v1-updateoutput-resultsentry">kube.v1.UpdateOutput.ResultsEntry</a></td>
<td><pre>
results contains the outcome for each resource
Keys match the resource_id from UpdateInput<br>

json_name: results
go_name: Results</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-updatesummary">kube.v1.UpdateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: summary
go_name: Summary</pre></td>
</tr>
</table>



<a name="kube-v1-updateoutput-resultsentry"></a>
### kube.v1.UpdateOutput.ResultsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#kube-v1-updateresult">kube.v1.UpdateResult</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-updateresult"></a>
### kube.v1.UpdateResult

<pre>
UpdateResult represents the outcome of updating a single resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>error_message</td>
<td>string</td>
<td><pre>
error_message contains error details (if failed)<br>

json_name: errorMessage
go_name: ErrorMessage</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: resourceId
go_name: ResourceId</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if updated successfully)<br>

json_name: resourceStatus
go_name: ResourceStatus</pre></td>
</tr>
</table>



<a name="kube-v1-updatesummary"></a>
### kube.v1.UpdateSummary

<pre>
UpdateSummary provides aggregate statistics
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: failed
go_name: Failed</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: total
go_name: Total</pre></td>
</tr><tr>
<td>updated</td>
<td>int32</td>
<td><pre>
json_name: updated
go_name: Updated</pre></td>
</tr>
</table>



<a name="temporal-v1"></a>
# temporal.v1

<a name="temporal-v1-services"></a>
## Services

<a name="temporal-v1-admin"></a>
## temporal.v1.Admin

<pre>
Admin service defines workflows and activities for Temporal namespace management
</pre>

<a name="temporal-v1-admin-workflows"></a>
### Workflows

---
<a name="temporal-v1-createnamespace-workflow"></a>
### temporal.v1.CreateNamespace

<pre>
CreateNamespace orchestrates the creation of a new Temporal namespace
This is a workflow (not an activity) because namespace creation in Temporal is async
and requires polling to verify the namespace is active before returning.
</pre>

**Input:** [temporal.v1.CreateNamespaceInput](#temporal-v1-createnamespaceinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to create (e.g., "fabric-org1", "tenant-acme")
This will be used as the workflow ID to ensure idempotency<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: retention
go_name: Retention</pre></td>
</tr>
</table>

**Output:** [temporal.v1.CreateNamespaceOutput](#temporal-v1-createnamespaceoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name echoes back the created namespace name<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>    

<a name="temporal-v1-admin-activities"></a>
### Activities

---
<a name="temporal-v1-registernamespace-activity"></a>
### temporal.v1.RegisterNamespace

<pre>
RegisterNamespace creates a new Temporal namespace with specified retention policy
Idempotent: returns success if namespace already exists
</pre>

**Input:** [temporal.v1.RegisterNamespaceInput](#temporal-v1-registernamespaceinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to create<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: retention
go_name: Retention</pre></td>
</tr>
</table>

**Output:** [temporal.v1.RegisterNamespaceOutput](#temporal-v1-registernamespaceoutput)

 

---
<a name="temporal-v1-describenamespace-activity"></a>
### temporal.v1.DescribeNamespace

<pre>
DescribeNamespace retrieves the current state of a Temporal namespace
Used by the workflow to poll until namespace is active
</pre>

**Input:** [temporal.v1.DescribeNamespaceInput](#temporal-v1-describenamespaceinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to describe<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>

**Output:** [temporal.v1.DescribeNamespaceOutput](#temporal-v1-describenamespaceoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the described namespace<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>state</td>
<td><a href="#temporal-v1-namespacestate">temporal.v1.NamespaceState</a></td>
<td><pre>
state is the current lifecycle state of the namespace<br>

json_name: state
go_name: State</pre></td>
</tr>
</table> 

---
<a name="temporal-v1-createnexusendpoint-activity"></a>
### temporal.v1.CreateNexusEndpoint

<pre>
CreateNexusEndpoint creates a global Nexus endpoint for cross-namespace Workflow calls
This endpoint acts as a reverse proxy, routing Nexus requests from caller namespaces
to a target namespace and task queue (typically for CreateScopedIdentityAccess service).
Idempotent: returns success if endpoint already exists with same configuration.
</pre>

**Input:** [temporal.v1.CreateNexusEndpointInput](#temporal-v1-createnexusendpointinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name: endpoint name<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route to<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: taskQueue
go_name: TaskQueue</pre></td>
</tr>
</table>

**Output:** [temporal.v1.CreateNexusEndpointOutput](#temporal-v1-createnexusendpointoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>id</td>
<td>string</td>
<td><pre>
id: server-generated unique endpoint identifier<br>

json_name: id
go_name: Id</pre></td>
</tr>
</table> 

---
<a name="temporal-v1-describenexusendpoint-activity"></a>
### temporal.v1.DescribeNexusEndpoint

<pre>
DescribeNexusEndpoint retrieves the current configuration of a Nexus endpoint
Used to fetch existing ACL before updating (for ACL update workflows)
</pre>

**Input:** [temporal.v1.DescribeNexusEndpointInput](#temporal-v1-describenexusendpointinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>id</td>
<td>string</td>
<td><pre>
id: endpoint identifier to describe<br>

json_name: id
go_name: Id</pre></td>
</tr>
</table>

**Output:** [temporal.v1.DescribeNexusEndpointOutput](#temporal-v1-describenexusendpointoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>endpoint</td>
<td><a href="#temporal-v1-endpoint">temporal.v1.Endpoint</a></td>
<td><pre>
json_name: endpoint
go_name: Endpoint</pre></td>
</tr>
</table> 

---
<a name="temporal-v1-updatenexusendpoint-activity"></a>
### temporal.v1.UpdateNexusEndpoint

<pre>
UpdateNexusEndpoint updates an existing Nexus endpoint configuration
NOTE: This REPLACES the entire endpoint configuration.
Caller must provide the complete desired configuration.
</pre>

**Input:** [temporal.v1.UpdateNexusEndpointInput](#temporal-v1-updatenexusendpointinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>endpoint</td>
<td><a href="#temporal-v1-endpoint">temporal.v1.Endpoint</a></td>
<td><pre>
endpoint: complete endpoint configuration for update
Must include id, namespace, and task_queue<br>

json_name: endpoint
go_name: Endpoint</pre></td>
</tr>
</table>

**Output:** [temporal.v1.UpdateNexusEndpointOutput](#temporal-v1-updatenexusendpointoutput)

   

<a name="temporal-v1-messages"></a>
## Messages

<a name="temporal-v1-createnamespaceinput"></a>
### temporal.v1.CreateNamespaceInput

<pre>
CreateNamespaceInput specifies parameters for creating a Temporal namespace
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to create (e.g., "fabric-org1", "tenant-acme")
This will be used as the workflow ID to ensure idempotency<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: retention
go_name: Retention</pre></td>
</tr>
</table>



<a name="temporal-v1-createnamespaceoutput"></a>
### temporal.v1.CreateNamespaceOutput

<pre>
CreateNamespaceOutput indicates successful namespace creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name echoes back the created namespace name<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="temporal-v1-createnexusendpointinput"></a>
### temporal.v1.CreateNexusEndpointInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name: endpoint name<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route to<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: taskQueue
go_name: TaskQueue</pre></td>
</tr>
</table>



<a name="temporal-v1-createnexusendpointoutput"></a>
### temporal.v1.CreateNexusEndpointOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>id</td>
<td>string</td>
<td><pre>
id: server-generated unique endpoint identifier<br>

json_name: id
go_name: Id</pre></td>
</tr>
</table>



<a name="temporal-v1-describenamespaceinput"></a>
### temporal.v1.DescribeNamespaceInput

<pre>
DescribeNamespaceInput specifies which namespace to describe
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to describe<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="temporal-v1-describenamespaceoutput"></a>
### temporal.v1.DescribeNamespaceOutput

<pre>
DescribeNamespaceOutput returns the current state of a namespace
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the described namespace<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>state</td>
<td><a href="#temporal-v1-namespacestate">temporal.v1.NamespaceState</a></td>
<td><pre>
state is the current lifecycle state of the namespace<br>

json_name: state
go_name: State</pre></td>
</tr>
</table>



<a name="temporal-v1-describenexusendpointinput"></a>
### temporal.v1.DescribeNexusEndpointInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>id</td>
<td>string</td>
<td><pre>
id: endpoint identifier to describe<br>

json_name: id
go_name: Id</pre></td>
</tr>
</table>



<a name="temporal-v1-describenexusendpointoutput"></a>
### temporal.v1.DescribeNexusEndpointOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>endpoint</td>
<td><a href="#temporal-v1-endpoint">temporal.v1.Endpoint</a></td>
<td><pre>
json_name: endpoint
go_name: Endpoint</pre></td>
</tr>
</table>



<a name="temporal-v1-endpoint"></a>
### temporal.v1.Endpoint

<pre>
Endpoint defines a Nexus endpoint configuration
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>id</td>
<td>string</td>
<td><pre>
id: server-generated unique endpoint identifier<br>

json_name: id
go_name: Id</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name: endpoint name (e.g., "CreateScopedIdentityAccess")<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route requests to<br>

json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: taskQueue
go_name: TaskQueue</pre></td>
</tr>
</table>



<a name="temporal-v1-namespacestate"></a>
### temporal.v1.NamespaceState

<pre>
NamespaceState represents the lifecycle state of a Temporal namespace
</pre>

<table>
<tr><th>Value</th><th>Description</th></tr>
<tr>
<td>UNSPECIFIED</td>
<td></td>
</tr><tr>
<td>REGISTERED</td>
<td></td>
</tr><tr>
<td>DEPRECATED</td>
<td></td>
</tr>
</table>

<a name="temporal-v1-registernamespaceinput"></a>
### temporal.v1.RegisterNamespaceInput

<pre>
RegisterNamespaceInput specifies parameters for the RegisterNamespace activity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the namespace to create<br>

json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: retention
go_name: Retention</pre></td>
</tr>
</table>



<a name="temporal-v1-registernamespaceoutput"></a>
### temporal.v1.RegisterNamespaceOutput

<pre>
RegisterNamespaceOutput indicates successful namespace registration
</pre>



<a name="temporal-v1-updatenexusendpointinput"></a>
### temporal.v1.UpdateNexusEndpointInput

<pre>
⚠️ REPLACES entire endpoint configuration (caller provides complete desired state)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>endpoint</td>
<td><a href="#temporal-v1-endpoint">temporal.v1.Endpoint</a></td>
<td><pre>
endpoint: complete endpoint configuration for update
Must include id, namespace, and task_queue<br>

json_name: endpoint
go_name: Endpoint</pre></td>
</tr>
</table>



<a name="temporal-v1-updatenexusendpointoutput"></a>
### temporal.v1.UpdateNexusEndpointOutput



<a name="vault-v1"></a>
# vault.v1

<a name="vault-v1-services"></a>
## Services

<a name="vault-v1-vault"></a>
## vault.v1.Vault

<pre>
Vault service defines activities related to secret and auth management
</pre>

<a name="vault-v1-vault-workflows"></a>
### Workflows

---
<a name="vault-v1-createaccess-workflow"></a>
### vault.v1.CreateAccess

<pre>
Access Management
</pre>

**Input:** [vault.v1.CreateAccessInput](#vault-v1-createaccessinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>auth_role_name</td>
<td>string</td>
<td><pre>
Auth role configuration
auth_role_name is the name of the Kubernetes auth role to create<br>

json_name: authRoleName
go_name: AuthRoleName</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
policy_document is the structured policy content<br>

json_name: policyDocument
go_name: PolicyDocument</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
Policy configuration
policy_name is the name of the policy to create<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>

**Output:** [vault.v1.CreateAccessOutput](#vault-v1-createaccessoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>auth_role_name</td>
<td>string</td>
<td><pre>
auth_role_name is the name of the created Kubernetes auth role<br>

json_name: authRoleName
go_name: AuthRoleName</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
policy_name is the name of the created policy<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>    

<a name="vault-v1-vault-activities"></a>
### Activities

---
<a name="vault-v1-createauthrole-activity"></a>
### vault.v1.CreateAuthRole

<pre>
Auth Role Management
</pre>

**Input:** [vault.v1.CreateAuthRoleInput](#vault-v1-createauthroleinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
policy_names is a list of Vault policy names to bind to this role<br>

json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>role_name</td>
<td>string</td>
<td><pre>
role_name is the name of the Kubernetes auth role to create<br>

json_name: roleName
go_name: RoleName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>

**Output:** [vault.v1.CreateAuthRoleOutput](#vault-v1-createauthroleoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the created Kubernetes auth role<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table> 

---
<a name="vault-v1-deleteauthrole-activity"></a>
### vault.v1.DeleteAuthRole



**Input:** [vault.v1.DeleteAuthRoleInput](#vault-v1-deleteauthroleinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the Kubernetes auth role to delete<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>

**Output:** [vault.v1.DeleteAuthRoleOutput](#vault-v1-deleteauthroleoutput)

 

---
<a name="vault-v1-createpolicy-activity"></a>
### vault.v1.CreatePolicy

<pre>
Policy Management
</pre>

**Input:** [vault.v1.CreatePolicyInput](#vault-v1-createpolicyinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
document contains the structured policy rules<br>

json_name: document
go_name: Document</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the policy<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>

**Output:** [vault.v1.CreatePolicyOutput](#vault-v1-createpolicyoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the created policy<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table> 

---
<a name="vault-v1-deletepolicy-activity"></a>
### vault.v1.DeletePolicy



**Input:** [vault.v1.DeletePolicyInput](#vault-v1-deletepolicyinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the policy to delete<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>

**Output:** [vault.v1.DeletePolicyOutput](#vault-v1-deletepolicyoutput)

 

---
<a name="vault-v1-storesecret-activity"></a>
### vault.v1.StoreSecret

<pre>
Secret Management
</pre>

**Input:** [vault.v1.StoreSecretInput](#vault-v1-storesecretinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-storesecretinput-dataentry">vault.v1.StoreSecretInput.DataEntry</a></td>
<td><pre>
data is the secret key-value pairs to store<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path to store the secret at<br>

json_name: path
go_name: Path</pre></td>
</tr>
</table>

**Output:** [vault.v1.StoreSecretOutput](#vault-v1-storesecretoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: secretRef
go_name: SecretRef</pre></td>
</tr>
</table> 

---
<a name="vault-v1-fetchsecret-activity"></a>
### vault.v1.FetchSecret



**Input:** [vault.v1.FetchSecretInput](#vault-v1-fetchsecretinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path to fetch the secret from<br>

json_name: path
go_name: Path</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
version is the specific version to fetch (0 or omit for latest)<br>

json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Output:** [vault.v1.FetchSecretOutput](#vault-v1-fetchsecretoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-fetchsecretoutput-dataentry">vault.v1.FetchSecretOutput.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>secret_ref</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: secretRef
go_name: SecretRef</pre></td>
</tr>
</table> 

---
<a name="vault-v1-deletesecret-activity"></a>
### vault.v1.DeleteSecret



**Input:** [vault.v1.DeleteSecretInput](#vault-v1-deletesecretinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path of the secret to delete<br>

json_name: path
go_name: Path</pre></td>
</tr>
</table>

**Output:** [vault.v1.DeleteSecretOutput](#vault-v1-deletesecretoutput)

   

<a name="vault-v1-messages"></a>
## Messages

<a name="vault-v1-capability"></a>
### vault.v1.Capability

<pre>
Capability enum for Vault policy permissions
</pre>

<table>
<tr><th>Value</th><th>Description</th></tr>
<tr>
<td>CAPABILITY_UNSPECIFIED</td>
<td></td>
</tr><tr>
<td>CAPABILITY_CREATE</td>
<td></td>
</tr><tr>
<td>CAPABILITY_READ</td>
<td></td>
</tr><tr>
<td>CAPABILITY_UPDATE</td>
<td></td>
</tr><tr>
<td>CAPABILITY_PATCH</td>
<td></td>
</tr><tr>
<td>CAPABILITY_DELETE</td>
<td></td>
</tr><tr>
<td>CAPABILITY_LIST</td>
<td></td>
</tr><tr>
<td>CAPABILITY_SUDO</td>
<td></td>
</tr><tr>
<td>CAPABILITY_DENY</td>
<td></td>
</tr><tr>
<td>CAPABILITY_SUBSCRIBE</td>
<td></td>
</tr><tr>
<td>CAPABILITY_RECOVER</td>
<td></td>
</tr>
</table>

<a name="vault-v1-createaccessinput"></a>
### vault.v1.CreateAccessInput

<pre>
CreateAccessInput specifies parameters for creating both a policy and auth role
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>auth_role_name</td>
<td>string</td>
<td><pre>
Auth role configuration
auth_role_name is the name of the Kubernetes auth role to create<br>

json_name: authRoleName
go_name: AuthRoleName</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
policy_document is the structured policy content<br>

json_name: policyDocument
go_name: PolicyDocument</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
Policy configuration
policy_name is the name of the policy to create<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>



<a name="vault-v1-createaccessoutput"></a>
### vault.v1.CreateAccessOutput

<pre>
CreateAccessOutput returns the created policy and auth role names
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>auth_role_name</td>
<td>string</td>
<td><pre>
auth_role_name is the name of the created Kubernetes auth role<br>

json_name: authRoleName
go_name: AuthRoleName</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
policy_name is the name of the created policy<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr>
</table>



<a name="vault-v1-createauthroleinput"></a>
### vault.v1.CreateAuthRoleInput

<pre>
CreateAuthRoleInput specifies parameters for creating an auth role in Vault
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
policy_names is a list of Vault policy names to bind to this role<br>

json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>role_name</td>
<td>string</td>
<td><pre>
role_name is the name of the Kubernetes auth role to create<br>

json_name: roleName
go_name: RoleName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>



<a name="vault-v1-createauthroleoutput"></a>
### vault.v1.CreateAuthRoleOutput

<pre>
CreateAuthRoleOutput returns the created role name
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the created Kubernetes auth role<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="vault-v1-createpolicyinput"></a>
### vault.v1.CreatePolicyInput

<pre>
CreatePolicyInput specifies parameters for creating a Vault policy
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
document contains the structured policy rules<br>

json_name: document
go_name: Document</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the policy<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="vault-v1-createpolicyoutput"></a>
### vault.v1.CreatePolicyOutput

<pre>
CreatePolicyOutput returns the created policy name
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the created policy<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="vault-v1-deleteauthroleinput"></a>
### vault.v1.DeleteAuthRoleInput

<pre>
DeleteAuthRoleInput specifies which auth role to delete
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the Kubernetes auth role to delete<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="vault-v1-deleteauthroleoutput"></a>
### vault.v1.DeleteAuthRoleOutput

<pre>
DeleteAuthRoleOutput confirms successful auth role deletion
</pre>



<a name="vault-v1-deletepolicyinput"></a>
### vault.v1.DeletePolicyInput

<pre>
DeletePolicyInput specifies which policy to delete
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the name of the policy to delete<br>

json_name: name
go_name: Name</pre></td>
</tr>
</table>



<a name="vault-v1-deletepolicyoutput"></a>
### vault.v1.DeletePolicyOutput

<pre>
DeletePolicyOutput confirms successful policy deletion
</pre>



<a name="vault-v1-deletesecretinput"></a>
### vault.v1.DeleteSecretInput

<pre>
DeleteSecretInput specifies which secret to delete
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path of the secret to delete<br>

json_name: path
go_name: Path</pre></td>
</tr>
</table>



<a name="vault-v1-deletesecretoutput"></a>
### vault.v1.DeleteSecretOutput

<pre>
DeleteSecretOutput confirms successful secret deletion
</pre>



<a name="vault-v1-fetchsecretinput"></a>
### vault.v1.FetchSecretInput

<pre>
FetchSecretInput specifies which secret to retrieve
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path to fetch the secret from<br>

json_name: path
go_name: Path</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
version is the specific version to fetch (0 or omit for latest)<br>

json_name: version
go_name: Version</pre></td>
</tr>
</table>



<a name="vault-v1-fetchsecretoutput"></a>
### vault.v1.FetchSecretOutput

<pre>
FetchSecretOutput returns the fetched secret with all its data
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-fetchsecretoutput-dataentry">vault.v1.FetchSecretOutput.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>secret_ref</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: secretRef
go_name: SecretRef</pre></td>
</tr>
</table>



<a name="vault-v1-fetchsecretoutput-dataentry"></a>
### vault.v1.FetchSecretOutput.DataEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td>bytes</td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="vault-v1-policydocument"></a>
### vault.v1.PolicyDocument

<pre>
PolicyDocument contains structured policy rules
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>rules</td>
<td><a href="#vault-v1-policyrule">vault.v1.PolicyRule</a></td>
<td><pre>
rules is a list of path-capability mappings<br>

json_name: rules
go_name: Rules</pre></td>
</tr>
</table>



<a name="vault-v1-policyrule"></a>
### vault.v1.PolicyRule

<pre>
PolicyRule defines a single path and its allowed capabilities
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>capabilities</td>
<td><a href="#vault-v1-capability">vault.v1.Capability</a></td>
<td><pre>
capabilities are the allowed capabilities for this path<br>

json_name: capabilities
go_name: Capabilities</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path pattern (e.g., "secret/data/myapp/*")<br>

json_name: path
go_name: Path</pre></td>
</tr>
</table>



<a name="vault-v1-secretref"></a>
### vault.v1.SecretRef

<pre>
SecretRef is a reference to a secret in Vault
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>path</td>
<td>string</td>
<td><pre>
json_name: path
go_name: Path</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
version is the specific version referenced. Set to 0 or omit for latest version<br>

json_name: version
go_name: Version</pre></td>
</tr>
</table>



<a name="vault-v1-storesecretinput"></a>
### vault.v1.StoreSecretInput

<pre>
StoreSecretInput specifies parameters for storing a secret in Vault
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-storesecretinput-dataentry">vault.v1.StoreSecretInput.DataEntry</a></td>
<td><pre>
data is the secret key-value pairs to store<br>

json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
path is the Vault path to store the secret at<br>

json_name: path
go_name: Path</pre></td>
</tr>
</table>



<a name="vault-v1-storesecretinput-dataentry"></a>
### vault.v1.StoreSecretInput.DataEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td>bytes</td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="vault-v1-storesecretoutput"></a>
### vault.v1.StoreSecretOutput

<pre>
StoreSecretOutput returns the secret reference with version information
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: secretRef
go_name: SecretRef</pre></td>
</tr>
</table>




<a name="google-protobuf"></a>
# google.protobuf

<a name="google-protobuf-messages"></a>
## Messages

<a name="google-protobuf-duration"></a>
### google.protobuf.Duration

<pre>
A Duration represents a signed, fixed-length span of time represented
as a count of seconds and fractions of seconds at nanosecond
resolution. It is independent of any calendar and concepts like "day"
or "month". It is related to Timestamp in that the difference between
two Timestamp values is a Duration and it can be added or subtracted
from a Timestamp. Range is approximately +-10,000 years.

# Examples

Example 1: Compute Duration from two Timestamps in pseudo code.

    Timestamp start = ...;
    Timestamp end = ...;
    Duration duration = ...;

    duration.seconds = end.seconds - start.seconds;
    duration.nanos = end.nanos - start.nanos;

    if (duration.seconds < 0 && duration.nanos > 0) {
      duration.seconds += 1;
      duration.nanos -= 1000000000;
    } else if (duration.seconds > 0 && duration.nanos < 0) {
      duration.seconds -= 1;
      duration.nanos += 1000000000;
    }

Example 2: Compute Timestamp from Timestamp + Duration in pseudo code.

    Timestamp start = ...;
    Duration duration = ...;
    Timestamp end = ...;

    end.seconds = start.seconds + duration.seconds;
    end.nanos = start.nanos + duration.nanos;

    if (end.nanos < 0) {
      end.seconds -= 1;
      end.nanos += 1000000000;
    } else if (end.nanos >= 1000000000) {
      end.seconds += 1;
      end.nanos -= 1000000000;
    }

Example 3: Compute Duration from datetime.timedelta in Python.

    td = datetime.timedelta(days=3, minutes=10)
    duration = Duration()
    duration.FromTimedelta(td)

# JSON Mapping

In JSON format, the Duration type is encoded as a string rather than an
object, where the string ends in the suffix "s" (indicating seconds) and
is preceded by the number of seconds, with nanoseconds expressed as
fractional seconds. For example, 3 seconds with 0 nanoseconds should be
encoded in JSON format as "3s", while 3 seconds and 1 nanosecond should
be expressed in JSON format as "3.000000001s", and 3 seconds and 1
microsecond should be expressed in JSON format as "3.000001s".
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>nanos</td>
<td>int32</td>
<td><pre>
Signed fractions of a second at nanosecond resolution of the span
of time. Durations less than one second are represented with a 0
`seconds` field and a positive or negative `nanos` field. For durations
of one second or more, a non-zero value for the `nanos` field must be
of the same sign as the `seconds` field. Must be from -999,999,999
to +999,999,999 inclusive.<br>

json_name: nanos
go_name: Nanos</pre></td>
</tr><tr>
<td>seconds</td>
<td>int64</td>
<td><pre>
Signed seconds of the span of time. Must be from -315,576,000,000
to +315,576,000,000 inclusive. Note: these bounds are computed from:
60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years<br>

json_name: seconds
go_name: Seconds</pre></td>
</tr>
</table>



<a name="google-protobuf-listvalue"></a>
### google.protobuf.ListValue

<pre>
`ListValue` is a wrapper around a repeated field of values.

The JSON representation for `ListValue` is JSON array.
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>values</td>
<td><a href="#google-protobuf-value">google.protobuf.Value</a></td>
<td><pre>
Repeated field of dynamically typed values.<br>

json_name: values
go_name: Values</pre></td>
</tr>
</table>



<a name="google-protobuf-nullvalue"></a>
### google.protobuf.NullValue

<pre>
`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

The JSON representation for `NullValue` is JSON `null`.
</pre>

<table>
<tr><th>Value</th><th>Description</th></tr>
<tr>
<td>NULL_VALUE</td>
<td><pre>
Null value.
</pre></td>
</tr>
</table>

<a name="google-protobuf-struct"></a>
### google.protobuf.Struct

<pre>
`Struct` represents a structured data value, consisting of fields
which map to dynamically typed values. In some languages, `Struct`
might be supported by a native representation. For example, in
scripting languages like JS a struct is represented as an
object. The details of that representation are described together
with the proto support for the language.

The JSON representation for `Struct` is JSON object.
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>fields</td>
<td><a href="#google-protobuf-struct-fieldsentry">google.protobuf.Struct.FieldsEntry</a></td>
<td><pre>
Unordered map of dynamically typed values.<br>

json_name: fields
go_name: Fields</pre></td>
</tr>
</table>



<a name="google-protobuf-struct-fieldsentry"></a>
### google.protobuf.Struct.FieldsEntry

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>key</td>
<td>string</td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>value</td>
<td><a href="#google-protobuf-value">google.protobuf.Value</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="google-protobuf-value"></a>
### google.protobuf.Value

<pre>
`Value` represents a dynamically typed value which can be either
null, a number, a string, a boolean, a recursive struct value, or a
list of values. A producer of value is expected to set one of these
variants. Absence of any variant indicates an error.

The JSON representation for `Value` is JSON value.
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bool_value</td>
<td>bool</td>
<td><pre>
Represents a boolean value.<br>

json_name: boolValue
go_name: BoolValue</pre></td>
</tr><tr>
<td>list_value</td>
<td><a href="#google-protobuf-listvalue">google.protobuf.ListValue</a></td>
<td><pre>
Represents a repeated `Value`.<br>

json_name: listValue
go_name: ListValue</pre></td>
</tr><tr>
<td>null_value</td>
<td><a href="#google-protobuf-nullvalue">google.protobuf.NullValue</a></td>
<td><pre>
Represents a null value.<br>

json_name: nullValue
go_name: NullValue</pre></td>
</tr><tr>
<td>number_value</td>
<td>double</td>
<td><pre>
Represents a double value.<br>

json_name: numberValue
go_name: NumberValue</pre></td>
</tr><tr>
<td>string_value</td>
<td>string</td>
<td><pre>
Represents a string value.<br>

json_name: stringValue
go_name: StringValue</pre></td>
</tr><tr>
<td>struct_value</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
Represents a structured value.<br>

json_name: structValue
go_name: StructValue</pre></td>
</tr>
</table>

