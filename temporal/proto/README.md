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
        - [cell.v1.BootstrapInfra](#cell-v1-bootstrapinfra-workflow)
        - [cell.v1.InitFabricShard](#cell-v1-initfabricshard-workflow)
  - Messages
    - [cell.v1.BootstrapInfraInput](#cell-v1-bootstrapinfrainput)
    - [cell.v1.BootstrapInfraOutput](#cell-v1-bootstrapinfraoutput)
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
        - [fabric.v1.DeployCertificateAuthority](#fabric-v1-deploycertificateauthority-workflow)
        - [fabric.v1.CreateCertificateDatabase](#fabric-v1-createcertificatedatabase-workflow)
        - [fabric.v1.CreateCertificate](#fabric-v1-createcertificate-workflow)
        - [fabric.v1.CreateBootstrapAdmin](#fabric-v1-createbootstrapadmin-workflow)
        - [fabric.v1.CreateAdmin](#fabric-v1-createadmin-workflow)
        - [fabric.v1.CreateIdentity](#fabric-v1-createidentity-workflow)
        - [fabric.v1.DeployCouchDb](#fabric-v1-deploycouchdb-workflow)
        - [fabric.v1.DeployPeer](#fabric-v1-deploypeer-workflow)
        - [fabric.v1.DeployOrderer](#fabric-v1-deployorderer-workflow)
        - [fabric.v1.DeployChainCode](#fabric-v1-deploychaincode-workflow)
        - [fabric.v1.CreateChannel](#fabric-v1-createchannel-workflow)
        - [fabric.v1.AddOrdererToChannel](#fabric-v1-addorderertochannel-workflow)
        - [fabric.v1.AddPeerToChannel](#fabric-v1-addpeertochannel-workflow)
        - [fabric.v1.JoinOrdererToChannel](#fabric-v1-joinorderertochannel-workflow)
        - [fabric.v1.JoinPeerToChannel](#fabric-v1-joinpeertochannel-workflow)
        - [fabric.v1.SetAnchorPeer](#fabric-v1-setanchorpeer-workflow)
        - [fabric.v1.AddOrdererToConsenterSet](#fabric-v1-addorderertoconsenterset-workflow)
        - [fabric.v1.AddOrgToChannel](#fabric-v1-addorgtochannel-workflow)
        - [fabric.v1.UpdateChannelConfig](#fabric-v1-updatechannelconfig-workflow)
    - [fabric.v1.FabricActivities](#fabric-v1-fabricactivities)
      - [Activities](#fabric-v1-fabricactivities-activities)
        - [fabric.v1.RegisterIdentity](#fabric-v1-registeridentity-activity)
        - [fabric.v1.EnrollIdentity](#fabric-v1-enrollidentity-activity)
        - [fabric.v1.RevokeIdentity](#fabric-v1-revokeidentity-activity)
        - [fabric.v1.BootstrapEnrollIdentity](#fabric-v1-bootstrapenrollidentity-activity)
  - Messages
    - [fabric.v1.AddOrdererToChannelInput](#fabric-v1-addorderertochannelinput)
    - [fabric.v1.AddOrdererToChannelOutput](#fabric-v1-addorderertochanneloutput)
    - [fabric.v1.AddOrdererToConsenterSetInput](#fabric-v1-addorderertoconsentersetinput)
    - [fabric.v1.AddOrdererToConsenterSetOutput](#fabric-v1-addorderertoconsentersetoutput)
    - [fabric.v1.AddOrgToChannelInput](#fabric-v1-addorgtochannelinput)
    - [fabric.v1.AddOrgToChannelOutput](#fabric-v1-addorgtochanneloutput)
    - [fabric.v1.AddPeerToChannelInput](#fabric-v1-addpeertochannelinput)
    - [fabric.v1.AddPeerToChannelOutput](#fabric-v1-addpeertochanneloutput)
    - [fabric.v1.BootstrapCertificateAuthoritiesInput](#fabric-v1-bootstrapcertificateauthoritiesinput)
    - [fabric.v1.BootstrapCertificateAuthoritiesOutput](#fabric-v1-bootstrapcertificateauthoritiesoutput)
    - [fabric.v1.BootstrapEnrollIdentityInput](#fabric-v1-bootstrapenrollidentityinput)
    - [fabric.v1.BootstrapEnrollIdentityOutput](#fabric-v1-bootstrapenrollidentityoutput)
    - [fabric.v1.BootstrapShardInput](#fabric-v1-bootstrapshardinput)
    - [fabric.v1.BootstrapShardOutput](#fabric-v1-bootstrapshardoutput)
    - [fabric.v1.CertificateAuthorityRef](#fabric-v1-certificateauthorityref)
    - [fabric.v1.CertificateSubject](#fabric-v1-certificatesubject)
    - [fabric.v1.CreateAdminInput](#fabric-v1-createadmininput)
    - [fabric.v1.CreateAdminOutput](#fabric-v1-createadminoutput)
    - [fabric.v1.CreateBootstrapAdminInput](#fabric-v1-createbootstrapadmininput)
    - [fabric.v1.CreateBootstrapAdminOutput](#fabric-v1-createbootstrapadminoutput)
    - [fabric.v1.CreateCertificateDatabaseInput](#fabric-v1-createcertificatedatabaseinput)
    - [fabric.v1.CreateCertificateDatabaseOutput](#fabric-v1-createcertificatedatabaseoutput)
    - [fabric.v1.CreateCertificateInput](#fabric-v1-createcertificateinput)
    - [fabric.v1.CreateCertificateInput.AttributesEntry](#fabric-v1-createcertificateinput-attributesentry)
    - [fabric.v1.CreateCertificateOutput](#fabric-v1-createcertificateoutput)
    - [fabric.v1.CreateChannelInput](#fabric-v1-createchannelinput)
    - [fabric.v1.CreateChannelOutput](#fabric-v1-createchanneloutput)
    - [fabric.v1.CreateIdentityInput](#fabric-v1-createidentityinput)
    - [fabric.v1.CreateIdentityInput.AttributesEntry](#fabric-v1-createidentityinput-attributesentry)
    - [fabric.v1.CreateIdentityOutput](#fabric-v1-createidentityoutput)
    - [fabric.v1.DeployCertificateAuthorityInput](#fabric-v1-deploycertificateauthorityinput)
    - [fabric.v1.DeployCertificateAuthorityOutput](#fabric-v1-deploycertificateauthorityoutput)
    - [fabric.v1.DeployCertificateStoreInput](#fabric-v1-deploycertificatestoreinput)
    - [fabric.v1.DeployCertificateStoreOutput](#fabric-v1-deploycertificatestoreoutput)
    - [fabric.v1.DeployChainCodeInput](#fabric-v1-deploychaincodeinput)
    - [fabric.v1.DeployChainCodeOutput](#fabric-v1-deploychaincodeoutput)
    - [fabric.v1.DeployCouchDbInput](#fabric-v1-deploycouchdbinput)
    - [fabric.v1.DeployCouchDbOutput](#fabric-v1-deploycouchdboutput)
    - [fabric.v1.DeployOrdererInput](#fabric-v1-deployordererinput)
    - [fabric.v1.DeployOrdererOutput](#fabric-v1-deployordereroutput)
    - [fabric.v1.DeployPeerInput](#fabric-v1-deploypeerinput)
    - [fabric.v1.DeployPeerOutput](#fabric-v1-deploypeeroutput)
    - [fabric.v1.EnrollIdentityInput](#fabric-v1-enrollidentityinput)
    - [fabric.v1.EnrollIdentityOutput](#fabric-v1-enrollidentityoutput)
    - [fabric.v1.JoinOrdererToChannelInput](#fabric-v1-joinorderertochannelinput)
    - [fabric.v1.JoinOrdererToChannelOutput](#fabric-v1-joinorderertochanneloutput)
    - [fabric.v1.JoinPeerToChannelInput](#fabric-v1-joinpeertochannelinput)
    - [fabric.v1.JoinPeerToChannelOutput](#fabric-v1-joinpeertochanneloutput)
    - [fabric.v1.RegisterIdentityInput](#fabric-v1-registeridentityinput)
    - [fabric.v1.RegisterIdentityInput.AttributesEntry](#fabric-v1-registeridentityinput-attributesentry)
    - [fabric.v1.RegisterIdentityOutput](#fabric-v1-registeridentityoutput)
    - [fabric.v1.RevokeIdentityInput](#fabric-v1-revokeidentityinput)
    - [fabric.v1.RevokeIdentityOutput](#fabric-v1-revokeidentityoutput)
    - [fabric.v1.SetAnchorPeerInput](#fabric-v1-setanchorpeerinput)
    - [fabric.v1.SetAnchorPeerOutput](#fabric-v1-setanchorpeeroutput)
    - [fabric.v1.UpdateChannelConfigInput](#fabric-v1-updatechannelconfiginput)
    - [fabric.v1.UpdateChannelConfigOutput](#fabric-v1-updatechannelconfigoutput)
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
        - [kube.v1.ExecPod](#kube-v1-execpod-activity)
  - Messages
    - [kube.v1.ApplyInput](#kube-v1-applyinput)
    - [kube.v1.ApplyInput.ResourcesEntry](#kube-v1-applyinput-resourcesentry)
    - [kube.v1.ApplyOutput](#kube-v1-applyoutput)
    - [kube.v1.ApplyOutput.ResultsEntry](#kube-v1-applyoutput-resultsentry)
    - [kube.v1.ApplyResult](#kube-v1-applyresult)
    - [kube.v1.ApplySummary](#kube-v1-applysummary)
    - [kube.v1.ContainerIdentifier](#kube-v1-containeridentifier)
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
    - [kube.v1.ExecPodInput](#kube-v1-execpodinput)
    - [kube.v1.ExecPodOutput](#kube-v1-execpodoutput)
    - [kube.v1.GetStatusInput](#kube-v1-getstatusinput)
    - [kube.v1.GetStatusInput.ResourcesEntry](#kube-v1-getstatusinput-resourcesentry)
    - [kube.v1.GetStatusOutput](#kube-v1-getstatusoutput)
    - [kube.v1.Resource](#kube-v1-resource)
    - [kube.v1.ResourceIdentifier](#kube-v1-resourceidentifier)
    - [kube.v1.ResourceStatus](#kube-v1-resourcestatus)
    - [kube.v1.UpdateInput](#kube-v1-updateinput)
    - [kube.v1.UpdateInput.ResourcesEntry](#kube-v1-updateinput-resourcesentry)
    - [kube.v1.UpdateOutput](#kube-v1-updateoutput)
    - [kube.v1.UpdateOutput.ResultsEntry](#kube-v1-updateoutput-resultsentry)
    - [kube.v1.UpdateResult](#kube-v1-updateresult)
    - [kube.v1.UpdateSummary](#kube-v1-updatesummary)
- [postgres.v1](#postgres-v1)
  - Services
    - [postgres.v1.Postgres](#postgres-v1-postgres)
      - [Activities](#postgres-v1-postgres-activities)
        - [postgres.v1.CreateDatabase](#postgres-v1-createdatabase-activity)
        - [postgres.v1.CreateUser](#postgres-v1-createuser-activity)
        - [postgres.v1.GrantPrivileges](#postgres-v1-grantprivileges-activity)
  - Messages
    - [postgres.v1.CreateDatabaseInput](#postgres-v1-createdatabaseinput)
    - [postgres.v1.CreateDatabaseOutput](#postgres-v1-createdatabaseoutput)
    - [postgres.v1.CreateUserInput](#postgres-v1-createuserinput)
    - [postgres.v1.CreateUserOutput](#postgres-v1-createuseroutput)
    - [postgres.v1.GrantPrivilegesInput](#postgres-v1-grantprivilegesinput)
    - [postgres.v1.GrantPrivilegesOutput](#postgres-v1-grantprivilegesoutput)
    - [postgres.v1.Privilege](#postgres-v1-privilege)
- [secrets.v1](#secrets-v1)
  - Services
    - [secrets.v1.Secrets](#secrets-v1-secrets)
      - [Workflows](#secrets-v1-secrets-workflows)
        - [secrets.v1.CreateAccess](#secrets-v1-createaccess-workflow)
      - [Activities](#secrets-v1-secrets-activities)
        - [secrets.v1.CreatePassword](#secrets-v1-createpassword-activity)
        - [secrets.v1.CreateKeyPair](#secrets-v1-createkeypair-activity)
        - [secrets.v1.CreateSelfSignedCertificate](#secrets-v1-createselfsignedcertificate-activity)
        - [secrets.v1.StoreCertificate](#secrets-v1-storecertificate-activity)
  - Messages
    - [secrets.v1.CreateAccessInput](#secrets-v1-createaccessinput)
    - [secrets.v1.CreateAccessOutput](#secrets-v1-createaccessoutput)
    - [secrets.v1.CreateKeyPairInput](#secrets-v1-createkeypairinput)
    - [secrets.v1.CreateKeyPairOutput](#secrets-v1-createkeypairoutput)
    - [secrets.v1.CreatePasswordInput](#secrets-v1-createpasswordinput)
    - [secrets.v1.CreatePasswordOutput](#secrets-v1-createpasswordoutput)
    - [secrets.v1.CreateSelfSignedCertificateInput](#secrets-v1-createselfsignedcertificateinput)
    - [secrets.v1.CreateSelfSignedCertificateOutput](#secrets-v1-createselfsignedcertificateoutput)
    - [secrets.v1.SecretRef](#secrets-v1-secretref)
    - [secrets.v1.StoreCertificateInput](#secrets-v1-storecertificateinput)
    - [secrets.v1.StoreCertificateOutput](#secrets-v1-storecertificateoutput)
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
      - [Activities](#vault-v1-vault-activities)
        - [vault.v1.CreateAuthRole](#vault-v1-createauthrole-activity)
        - [vault.v1.DeleteAuthRole](#vault-v1-deleteauthrole-activity)
        - [vault.v1.CreatePolicy](#vault-v1-createpolicy-activity)
        - [vault.v1.DeletePolicy](#vault-v1-deletepolicy-activity)
  - Messages
    - [vault.v1.Capability](#vault-v1-capability)
    - [vault.v1.CreateAuthRoleInput](#vault-v1-createauthroleinput)
    - [vault.v1.CreateAuthRoleOutput](#vault-v1-createauthroleoutput)
    - [vault.v1.CreatePolicyInput](#vault-v1-createpolicyinput)
    - [vault.v1.CreatePolicyOutput](#vault-v1-createpolicyoutput)
    - [vault.v1.DeleteAuthRoleInput](#vault-v1-deleteauthroleinput)
    - [vault.v1.DeleteAuthRoleOutput](#vault-v1-deleteauthroleoutput)
    - [vault.v1.DeletePolicyInput](#vault-v1-deletepolicyinput)
    - [vault.v1.DeletePolicyOutput](#vault-v1-deletepolicyoutput)
    - [vault.v1.PolicyDocument](#vault-v1-policydocument)
    - [vault.v1.PolicyRule](#vault-v1-policyrule)
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: ContentType
go_name: ContentType
go_tags: json:"ContentType,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: Data
go_name: Data
go_tags: json:"Data,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: Etag
go_name: Etag
go_tags: json:"Etag,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: ContentType
go_name: ContentType
go_tags: json:"ContentType,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: Data
go_name: Data
go_tags: json:"Data,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: Etag
go_name: Etag
go_tags: json:"Etag,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: ContentType
go_name: ContentType
go_tags: json:"ContentType,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: Data
go_name: Data
go_tags: json:"Data,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: Etag
go_name: Etag
go_tags: json:"Etag,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
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

json_name: DeleteMarker
go_name: DeleteMarker
go_tags: json:"DeleteMarker" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: DeleteMarker
go_name: DeleteMarker
go_tags: json:"DeleteMarker" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID involved in the deletion<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>content_type</td>
<td>string</td>
<td><pre>
content_type is the MIME type of the object (optional for reads, used for writes)<br>

json_name: ContentType
go_name: ContentType
go_tags: json:"ContentType,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>data</td>
<td>bytes</td>
<td><pre>
data is the object contents (for reads, populated from S3; for writes, content to upload)<br>

json_name: Data
go_name: Data
go_tags: json:"Data,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>etag</td>
<td>string</td>
<td><pre>
etag is the entity tag (version identifier) of the object (populated by S3 on read/write)<br>

json_name: Etag
go_name: Etag
go_tags: json:"Etag,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
</tr><tr>
<td>version_id</td>
<td>string</td>
<td><pre>
version_id is the version ID of the object (if versioning is enabled, populated by S3)<br>

json_name: VersionId
go_name: VersionId
go_tags: json:"VersionId,omitempty" validate:"omitempty"</pre></td>
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

json_name: Bucket
go_name: Bucket
go_tags: json:"Bucket" validate:"required"</pre></td>
</tr><tr>
<td>key</td>
<td>string</td>
<td><pre>
key is the S3 object key (path)<br>

json_name: Key
go_name: Key
go_tags: json:"Key" validate:"required"</pre></td>
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
<a name="cell-v1-bootstrapinfra-workflow"></a>
### cell.v1.BootstrapInfra

<pre>
BootstrapInfra orchestrates the initialization of the infrastructure platform
This root workflow executes in the central "infra" namespace and coordinates:
1. Registration of Nexus endpoints for cross-namespace service calls
2. Establishment of core platform services that enable other workflows

This workflow runs once at platform bootstrap time and sets up the Nexus
communication layer needed for workflows in different namespaces to call services.
</pre>

**Input:** [cell.v1.BootstrapInfraInput](#cell-v1-bootstrapinfrainput)



**Output:** [cell.v1.BootstrapInfraOutput](#cell-v1-bootstrapinfraoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>cell/bootstrap-infra</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

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

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
</tr>
</table>

**Output:** [cell.v1.InitFabricShardOutput](#cell-v1-initfabricshardoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>cell/init-fabric-shard/${! Org }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>     

<a name="cell-v1-messages"></a>
## Messages

<a name="cell-v1-bootstrapinfrainput"></a>
### cell.v1.BootstrapInfraInput

<pre>
BootstrapInfraInput specifies the parameters for bootstrapping infrastructure
</pre>



<a name="cell-v1-bootstrapinfraoutput"></a>
### cell.v1.BootstrapInfraOutput

<pre>
BootstrapInfraOutput indicates the result of infrastructure bootstrap
</pre>



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

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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

json_name: NamespaceName
go_name: NamespaceName
go_tags: json:"NamespaceName" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>cluster/fabric/create-namespace/${! Org }</code></pre></td></tr>
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
json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
</tr>
</table>

**Output:** [cluster.fabric.v1.CreateServiceAccountOutput](#cluster-fabric-v1-createserviceaccountoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>cluster/fabric/create-service-account/${! Org }</code></pre></td></tr>
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
Currently supported:
  - "worker": Fabric worker deployment
  - "certificate-store": PostgreSQL certificate store
  - "certificate-authority-{certType}-{caID}": Certificate Authority
      where certType is "tls" or "org", and caID is "root", "v1", "v2", etc.
      Examples: "certificate-authority-tls-root", "certificate-authority-org-v1"
Future components: "peer", "orderer", "couchdb"<br>

json_name: Component
go_name: Component
go_tags: json:"Component" validate:"required"</pre></td>
</tr><tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
Used to derive namespace names, service account names, etc.<br>

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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

json_name: DeploymentName
go_name: DeploymentName
go_tags: json:"DeploymentName" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>cluster/fabric/create-deployment/${! Org }/${! Component }</code></pre></td></tr>
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
Currently supported:
  - "worker": Fabric worker deployment
  - "certificate-store": PostgreSQL certificate store
  - "certificate-authority-{certType}-{caID}": Certificate Authority
      where certType is "tls" or "org", and caID is "root", "v1", "v2", etc.
      Examples: "certificate-authority-tls-root", "certificate-authority-org-v1"
Future components: "peer", "orderer", "couchdb"<br>

json_name: Component
go_name: Component
go_tags: json:"Component" validate:"required"</pre></td>
</tr><tr>
<td>org</td>
<td>string</td>
<td><pre>
org is the organization identifier (e.g., "org1")
Used to derive namespace names, service account names, etc.<br>

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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

json_name: DeploymentName
go_name: DeploymentName
go_tags: json:"DeploymentName" validate:"required"</pre></td>
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

json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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

json_name: NamespaceName
go_name: NamespaceName
go_tags: json:"NamespaceName" validate:"required"</pre></td>
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
json_name: Org
go_name: Org
go_tags: json:"Org" validate:"required"</pre></td>
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
Deploys TLS chain first (required for Org CA TLS certs), then Org chain.
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
<a name="fabric-v1-deploycertificateauthority-workflow"></a>
### fabric.v1.DeployCertificateAuthority

<pre>
DeployCertificateAuthority deploys a certificate authority (root or intermediate)
This unified workflow handles both root and intermediate CAs:
- Root CAs: self-signed signing cert, TLS from self (TLS Root) or TLS Int (Org Root)
- Int CAs: parent-signed signing cert, TLS from parent (TLS) or TLS Int (Org)
</pre>

**Input:** [fabric.v1.DeployCertificateAuthorityInput](#fabric-v1-deploycertificateauthorityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to deploy<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployCertificateAuthorityOutput](#fabric-v1-deploycertificateauthorityoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-certificate-authority/${! CaRef.CertType }/${! CaRef.Id }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createcertificatedatabase-workflow"></a>
### fabric.v1.CreateCertificateDatabase

<pre>
CreateCertificateDatabase creates a dedicated database and user for a certificate authority
in the shared certificate store (PostgreSQL). Each CA (root and intermediate) gets its own
isolated database to store issued certificates and revocation information.
</pre>

**Input:** [fabric.v1.CreateCertificateDatabaseInput](#fabric-v1-createcertificatedatabaseinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA this database is for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateCertificateDatabaseOutput](#fabric-v1-createcertificatedatabaseoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref is the Vault reference to retrieve the username and password<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-certificate-database/${! CaRef.CertType }/${! CaRef.Id }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createcertificate-workflow"></a>
### fabric.v1.CreateCertificate

<pre>
CreateCertificate creates a certificate issued by a CA (or self-signed if no issuer)
This is the atomic unit of certificate issuance - generates keypair, registers, enrolls.
Used for CA signing certs, TLS certs, and any other certificate needs.
</pre>

**Input:** [fabric.v1.CreateCertificateInput](#fabric-v1-createcertificateinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref is the issuer's admin credentials (for registration)
Required unless self-signed<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef"</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-v1-createcertificateinput-attributesentry">fabric.v1.CreateCertificateInput.AttributesEntry</a></td>
<td><pre>
attributes are additional CA attributes for the identity (e.g., hf.IntermediateCA)<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the Fabric CA identity type for registration.
This determines the OU field in the issued certificate.
Valid values: "peer", "orderer", "admin", "client" (default: "client")
The OU is critical for NodeOUs classification in the MSP.<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType,omitempty"</pre></td>
</tr><tr>
<td>issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
issuer_ref identifies which CA signs this certificate.
If nil/empty, creates a self-signed certificate.<br>

json_name: IssuerRef
go_name: IssuerRef
go_tags: json:"IssuerRef"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile determines key usages ("ca", "tls", or "" for default identity)<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile"</pre></td>
</tr><tr>
<td>purpose</td>
<td>string</td>
<td><pre>
purpose is a human-readable label for the workflow ID (e.g., "tls-root-signing")<br>

json_name: Purpose
go_name: Purpose
go_tags: json:"Purpose" validate:"required"</pre></td>
</tr><tr>
<td>storage_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_ref is where to store keypair + certificate<br>

json_name: StorageRef
go_name: StorageRef
go_tags: json:"StorageRef" validate:"required"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateCertificateOutput](#fabric-v1-createcertificateoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref points to the stored keypair + certificate<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-certificate/${! Purpose }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createbootstrapadmin-workflow"></a>
### fabric.v1.CreateBootstrapAdmin

<pre>
CreateBootstrapAdmin creates the bootstrap admin identity for a CA
This is called after CA deployment to enroll using bootstrap credentials.
Creates both signing and TLS client certs for the bootstrap admin.
</pre>

**Input:** [fabric.v1.CreateBootstrapAdminInput](#fabric-v1-createbootstrapadmininput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to create the bootstrap admin for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateBootstrapAdminOutput](#fabric-v1-createbootstrapadminoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the bootstrap admin's signing identity<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the bootstrap admin's TLS client certificate<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-bootstrap-admin/${! CaRef.CertType }/${! CaRef.Id }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createadmin-workflow"></a>
### fabric.v1.CreateAdmin

<pre>
CreateAdmin creates the permanent admin identity for a CA
Uses the bootstrap admin to register/enroll a new permanent admin.
Creates both signing and TLS client certs for the permanent admin.
</pre>

**Input:** [fabric.v1.CreateAdminInput](#fabric-v1-createadmininput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to create the admin for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateAdminOutput](#fabric-v1-createadminoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the admin's signing identity<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the admin's TLS client certificate<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-admin/${! CaRef.CertType }/${! CaRef.Id }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createidentity-workflow"></a>
### fabric.v1.CreateIdentity

<pre>
CreateIdentity creates a full Fabric identity with both org and TLS certificates
Used for creating peer, orderer, admin, and client identities.
This is the production workflow for identity creation (not bootstrap).
</pre>

**Input:** [fabric.v1.CreateIdentityInput](#fabric-v1-createidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>attributes</td>
<td><a href="#fabric-v1-createidentityinput-attributesentry">fabric.v1.CreateIdentityInput.AttributesEntry</a></td>
<td><pre>
attributes are additional CA attributes for the identity<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the unique name for this identity (e.g., "peer0", "orderer0", "admin")<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the Fabric CA identity type for registration.
This determines the OU field in the issued certificate.
Valid values: "peer", "orderer", "admin", "client" (default: "client")
The OU is critical for NodeOUs classification in the MSP.<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType,omitempty"</pre></td>
</tr><tr>
<td>org_issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
org_issuer_ref identifies which Org CA issues the identity certificate
Typically the Org Intermediate CA (org/v1)<br>

json_name: OrgIssuerRef
go_name: OrgIssuerRef
go_tags: json:"OrgIssuerRef" validate:"required"</pre></td>
</tr><tr>
<td>storage_signing_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_signing_ref is where to store the org identity (keypair + cert)<br>

json_name: StorageSigningRef
go_name: StorageSigningRef
go_tags: json:"StorageSigningRef" validate:"required"</pre></td>
</tr><tr>
<td>storage_tls_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_tls_ref is where to store the TLS identity (keypair + cert)<br>

json_name: StorageTlsRef
go_name: StorageTlsRef
go_tags: json:"StorageTlsRef" validate:"required"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr><tr>
<td>tls_issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
tls_issuer_ref identifies which TLS CA issues the TLS certificate
Typically the TLS Intermediate CA (tls/v1)<br>

json_name: TlsIssuerRef
go_name: TlsIssuerRef
go_tags: json:"TlsIssuerRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_profile</td>
<td>string</td>
<td><pre>
tls_profile specifies the TLS signing profile ("tls" for server/client, "" for client-only)<br>

json_name: TlsProfile
go_name: TlsProfile
go_tags: json:"TlsProfile"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateIdentityOutput](#fabric-v1-createidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the org identity credentials in Vault<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the TLS identity credentials in Vault<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-identity/${! IdentityName }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploycouchdb-workflow"></a>
### fabric.v1.DeployCouchDb

<pre>
DeployCouchDb deploys a CouchDB instance for a peer
CouchDB provides rich query support for Fabric state database
</pre>

**Input:** [fabric.v1.DeployCouchDbInput](#fabric-v1-deploycouchdbinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
peer_id identifies which peer this CouchDB is for (e.g., "peer0", "peer1")<br>

json_name: PeerId
go_name: PeerId
go_tags: json:"PeerId" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployCouchDbOutput](#fabric-v1-deploycouchdboutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-couchdb/${! PeerId }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploypeer-workflow"></a>
### fabric.v1.DeployPeer

<pre>
DeployPeer deploys a Fabric peer node
Peers endorse transactions and maintain ledger state
</pre>

**Input:** [fabric.v1.DeployPeerInput](#fabric-v1-deploypeerinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
peer_id identifies this peer (e.g., "peer0", "peer1")<br>

json_name: PeerId
go_name: PeerId
go_tags: json:"PeerId" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployPeerOutput](#fabric-v1-deploypeeroutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-peer/${! PeerId }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deployorderer-workflow"></a>
### fabric.v1.DeployOrderer

<pre>
DeployOrderer deploys a Fabric orderer node
Orderers provide consensus and block ordering services
</pre>

**Input:** [fabric.v1.DeployOrdererInput](#fabric-v1-deployordererinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
orderer_id identifies this orderer (e.g., "orderer0", "orderer1")<br>

json_name: OrdererId
go_name: OrdererId
go_tags: json:"OrdererId" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployOrdererOutput](#fabric-v1-deployordereroutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-orderer/${! OrdererId }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploychaincode-workflow"></a>
### fabric.v1.DeployChainCode

<pre>
DeployChainCode deploys a chaincode (smart contract) to the Fabric network
Chaincode runs in isolated containers and processes transactions
</pre>

**Input:** [fabric.v1.DeployChainCodeInput](#fabric-v1-deploychaincodeinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chain_code_id</td>
<td>string</td>
<td><pre>
chain_code_id identifies this chaincode (e.g., "asset-transfer", "fabcar")<br>

json_name: ChainCodeId
go_name: ChainCodeId
go_tags: json:"ChainCodeId" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployChainCodeOutput](#fabric-v1-deploychaincodeoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/deploy-chaincode/${! ChainCodeId }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createchannel-workflow"></a>
### fabric.v1.CreateChannel

<pre>
CreateChannel orchestrates creation of a new Fabric channel
Generates genesis block, joins orderers and peers, sets anchor peers
</pre>

**Input:** [fabric.v1.CreateChannelInput](#fabric-v1-createchannelinput)



**Output:** [fabric.v1.CreateChannelOutput](#fabric-v1-createchanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/create-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-addorderertochannel-workflow"></a>
### fabric.v1.AddOrdererToChannel

<pre>
AddOrdererToChannel adds a new orderer to an existing channel
Fetches latest config block, joins as follower, then promotes to consenter
</pre>

**Input:** [fabric.v1.AddOrdererToChannelInput](#fabric-v1-addorderertochannelinput)



**Output:** [fabric.v1.AddOrdererToChannelOutput](#fabric-v1-addorderertochanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/add-orderer-to-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-addpeertochannel-workflow"></a>
### fabric.v1.AddPeerToChannel

<pre>
AddPeerToChannel adds a new peer to an existing channel
Optionally adds the org to the channel first if not already a member
</pre>

**Input:** [fabric.v1.AddPeerToChannelInput](#fabric-v1-addpeertochannelinput)



**Output:** [fabric.v1.AddPeerToChannelOutput](#fabric-v1-addpeertochanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/add-peer-to-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-joinorderertochannel-workflow"></a>
### fabric.v1.JoinOrdererToChannel

<pre>
JoinOrdererToChannel joins an orderer node to a channel via osnadmin
Handles both genesis-based (consenter) and config-block-based (follower) joins
</pre>

**Input:** [fabric.v1.JoinOrdererToChannelInput](#fabric-v1-joinorderertochannelinput)



**Output:** [fabric.v1.JoinOrdererToChannelOutput](#fabric-v1-joinorderertochanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/join-orderer-to-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-joinpeertochannel-workflow"></a>
### fabric.v1.JoinPeerToChannel

<pre>
JoinPeerToChannel joins a peer node to a channel
Fetches the oldest block from an orderer and joins the peer
</pre>

**Input:** [fabric.v1.JoinPeerToChannelInput](#fabric-v1-joinpeertochannelinput)



**Output:** [fabric.v1.JoinPeerToChannelOutput](#fabric-v1-joinpeertochanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/join-peer-to-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-setanchorpeer-workflow"></a>
### fabric.v1.SetAnchorPeer

<pre>
SetAnchorPeer configures the anchor peer for an org on a channel
Uses UpdateChannelConfig to modify the channel's Application group
</pre>

**Input:** [fabric.v1.SetAnchorPeerInput](#fabric-v1-setanchorpeerinput)



**Output:** [fabric.v1.SetAnchorPeerOutput](#fabric-v1-setanchorpeeroutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/set-anchor-peer</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-addorderertoconsenterset-workflow"></a>
### fabric.v1.AddOrdererToConsenterSet

<pre>
AddOrdererToConsenterSet promotes an orderer from follower to consenter
via a channel config update to the ConsensusType metadata
</pre>

**Input:** [fabric.v1.AddOrdererToConsenterSetInput](#fabric-v1-addorderertoconsentersetinput)



**Output:** [fabric.v1.AddOrdererToConsenterSetOutput](#fabric-v1-addorderertoconsentersetoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/add-orderer-to-consenter-set</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-addorgtochannel-workflow"></a>
### fabric.v1.AddOrgToChannel

<pre>
AddOrgToChannel adds a new organization to an existing channel
Requires signatures from a majority of existing channel admins
</pre>

**Input:** [fabric.v1.AddOrgToChannelInput](#fabric-v1-addorgtochannelinput)



**Output:** [fabric.v1.AddOrgToChannelOutput](#fabric-v1-addorgtochanneloutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/add-org-to-channel</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>

---
<a name="fabric-v1-updatechannelconfig-workflow"></a>
### fabric.v1.UpdateChannelConfig

<pre>
UpdateChannelConfig is the common pattern for all channel configuration changes
Fetches current config, computes delta, collects signatures, and submits update
</pre>

**Input:** [fabric.v1.UpdateChannelConfigInput](#fabric-v1-updatechannelconfiginput)



**Output:** [fabric.v1.UpdateChannelConfigOutput](#fabric-v1-updatechannelconfigoutput)



**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>fabric/update-channel-config</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>     

<a name="fabric-v1-fabricactivities"></a>
## fabric.v1.FabricActivities

<pre>
FabricActivities service defines activities for Fabric CA operations
These activities interact with Fabric CA for identity management
Activities are NOT exposed via Nexus (only workflows can be called cross-namespace)
</pre>   

<a name="fabric-v1-fabricactivities-activities"></a>
### Activities

---
<a name="fabric-v1-registeridentity-activity"></a>
### fabric.v1.RegisterIdentity

<pre>
RegisterIdentity registers a new identity with a Fabric CA
This is the first step in identity creation - registration returns an enrollment secret
</pre>

**Input:** [fabric.v1.RegisterIdentityInput](#fabric-v1-registeridentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref specifies which admin credentials to use for registration<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
affiliation is the affiliation for the identity in the CA<br>

json_name: Affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-v1-registeridentityinput-attributesentry">fabric.v1.RegisterIdentityInput.AttributesEntry</a></td>
<td><pre>
attributes are additional attributes for the identity<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to register with<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the unique name for this identity<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the type of identity (e.g., "admin", "peer", "orderer", "client")<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.RegisterIdentityOutput](#fabric-v1-registeridentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>enrollment_secret</td>
<td>string</td>
<td><pre>
enrollment_secret is the secret used to enroll the identity<br>

json_name: EnrollmentSecret
go_name: EnrollmentSecret
go_tags: json:"EnrollmentSecret" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="fabric-v1-enrollidentity-activity"></a>
### fabric.v1.EnrollIdentity

<pre>
EnrollIdentity enrolls a registered identity and retrieves certificates
This is the second step in identity creation - enrollment uses the secret from registration
</pre>

**Input:** [fabric.v1.EnrollIdentityInput](#fabric-v1-enrollidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to enroll with<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_tls_cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
ca_tls_cert_ref points to the CA's TLS certificate in Vault<br>

json_name: CaTlsCertRef
go_name: CaTlsCertRef
go_tags: json:"CaTlsCertRef" validate:"required"</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.
The chain contains parent CA certs needed for x509 chain verification.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>enrollment_secret</td>
<td>string</td>
<td><pre>
enrollment_secret is the secret obtained from registration<br>

json_name: EnrollmentSecret
go_name: EnrollmentSecret
go_tags: json:"EnrollmentSecret" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the name of the identity to enroll<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref points to the private key to use for CSR generation<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile specifies which CA signing profile to use for the certificate.
Use "ca" for intermediate CA certificates (grants cert sign, crl sign key usages).
Use "tls" for TLS certificates (grants server auth, client auth key usages).
Leave empty to use the default profile (digital signature only).<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile,omitempty"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.EnrollIdentityOutput](#fabric-v1-enrollidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref points to where the certificate was stored<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="fabric-v1-revokeidentity-activity"></a>
### fabric.v1.RevokeIdentity

<pre>
RevokeIdentity revokes an identity from a Fabric CA
Used to revoke bootstrap admin credentials after permanent admin is created
</pre>

**Input:** [fabric.v1.RevokeIdentityInput](#fabric-v1-revokeidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref specifies which admin credentials to use for revocation<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to revoke from<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the name of the identity to revoke<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>reason</td>
<td>string</td>
<td><pre>
reason is the reason for revocation (optional)<br>

json_name: Reason
go_name: Reason</pre></td>
</tr>
</table>

**Output:** [fabric.v1.RevokeIdentityOutput](#fabric-v1-revokeidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>revoked</td>
<td>bool</td>
<td><pre>
revoked indicates if the identity was successfully revoked<br>

json_name: Revoked
go_name: Revoked</pre></td>
</tr>
</table> 

---
<a name="fabric-v1-bootstrapenrollidentity-activity"></a>
### fabric.v1.BootstrapEnrollIdentity

<pre>
BootstrapEnrollIdentity enrolls an identity using bootstrap credentials (username/password)
This is used ONLY during CA initialization to obtain the bootstrap admin's certificate
The activity generates a CSR internally and submits it to the CA
</pre>

**Input:** [fabric.v1.BootstrapEnrollIdentityInput](#fabric-v1-bootstrapenrollidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bootstrap_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
bootstrap_credentials_ref points to the bootstrap admin username/password in Vault<br>

json_name: BootstrapCredentialsRef
go_name: BootstrapCredentialsRef
go_tags: json:"BootstrapCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_tls_cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
ca_tls_cert_ref points to the CA's TLS certificate in Vault
This certificate is used to verify the CA server's TLS connection<br>

json_name: CaTlsCertRef
go_name: CaTlsCertRef
go_tags: json:"CaTlsCertRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_url</td>
<td>string</td>
<td><pre>
ca_url is the CA server URL (e.g., "https://ca-tls-root.fabric-org1.svc.cluster.local:7054")<br>

json_name: CaUrl
go_name: CaUrl
go_tags: json:"CaUrl" validate:"required,url"</pre></td>
</tr><tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the enrolled certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.
The chain contains parent CA certs needed for x509 chain verification.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref points to the private key in Vault to use for CSR generation<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile specifies which CA signing profile to use<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile,omitempty"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>

**Output:** [fabric.v1.BootstrapEnrollIdentityOutput](#fabric-v1-bootstrapenrollidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where the certificate was stored<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table>   

<a name="fabric-v1-messages"></a>
## Messages

<a name="fabric-v1-addorderertochannelinput"></a>
### fabric.v1.AddOrdererToChannelInput



<a name="fabric-v1-addorderertochanneloutput"></a>
### fabric.v1.AddOrdererToChannelOutput



<a name="fabric-v1-addorderertoconsentersetinput"></a>
### fabric.v1.AddOrdererToConsenterSetInput



<a name="fabric-v1-addorderertoconsentersetoutput"></a>
### fabric.v1.AddOrdererToConsenterSetOutput



<a name="fabric-v1-addorgtochannelinput"></a>
### fabric.v1.AddOrgToChannelInput



<a name="fabric-v1-addorgtochanneloutput"></a>
### fabric.v1.AddOrgToChannelOutput



<a name="fabric-v1-addpeertochannelinput"></a>
### fabric.v1.AddPeerToChannelInput



<a name="fabric-v1-addpeertochanneloutput"></a>
### fabric.v1.AddPeerToChannelOutput



<a name="fabric-v1-bootstrapcertificateauthoritiesinput"></a>
### fabric.v1.BootstrapCertificateAuthoritiesInput

<pre>
BootstrapCertificateAuthoritiesInput specifies the parameters for bootstrapping certificate authorities
</pre>



<a name="fabric-v1-bootstrapcertificateauthoritiesoutput"></a>
### fabric.v1.BootstrapCertificateAuthoritiesOutput

<pre>
BootstrapCertificateAuthoritiesOutput indicates the result of CA bootstrapping
</pre>



<a name="fabric-v1-bootstrapenrollidentityinput"></a>
### fabric.v1.BootstrapEnrollIdentityInput

<pre>
BootstrapEnrollIdentityInput specifies the parameters for bootstrap enrollment
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>bootstrap_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
bootstrap_credentials_ref points to the bootstrap admin username/password in Vault<br>

json_name: BootstrapCredentialsRef
go_name: BootstrapCredentialsRef
go_tags: json:"BootstrapCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_tls_cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
ca_tls_cert_ref points to the CA's TLS certificate in Vault
This certificate is used to verify the CA server's TLS connection<br>

json_name: CaTlsCertRef
go_name: CaTlsCertRef
go_tags: json:"CaTlsCertRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_url</td>
<td>string</td>
<td><pre>
ca_url is the CA server URL (e.g., "https://ca-tls-root.fabric-org1.svc.cluster.local:7054")<br>

json_name: CaUrl
go_name: CaUrl
go_tags: json:"CaUrl" validate:"required,url"</pre></td>
</tr><tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the enrolled certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.
The chain contains parent CA certs needed for x509 chain verification.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref points to the private key in Vault to use for CSR generation<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile specifies which CA signing profile to use<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile,omitempty"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-bootstrapenrollidentityoutput"></a>
### fabric.v1.BootstrapEnrollIdentityOutput

<pre>
BootstrapEnrollIdentityOutput confirms the certificate was enrolled and stored
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where the certificate was stored<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-bootstrapshardinput"></a>
### fabric.v1.BootstrapShardInput

<pre>
BootstrapShardInput specifies the parameters for bootstrapping a Fabric shard
</pre>



<a name="fabric-v1-bootstrapshardoutput"></a>
### fabric.v1.BootstrapShardOutput

<pre>
BootstrapShardOutput indicates the result of shard bootstrapping
</pre>



<a name="fabric-v1-certificateauthorityref"></a>
### fabric.v1.CertificateAuthorityRef

<pre>
CertificateAuthorityRef identifies a specific Certificate Authority
Used to specify issuers and targets for certificate operations
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_type</td>
<td>string</td>
<td><pre>
cert_type specifies the CA chain ("tls" or "org")<br>

json_name: CertType
go_name: CertType
go_tags: json:"CertType" validate:"required,oneof=tls org"</pre></td>
</tr><tr>
<td>id</td>
<td>string</td>
<td><pre>
id identifies the CA within its chain ("root" or version like "v1")<br>

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-certificatesubject"></a>
### fabric.v1.CertificateSubject

<pre>
CertificateSubject defines the subject fields for a certificate
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>common_name</td>
<td>string</td>
<td><pre>
common_name is the CN for the certificate subject<br>

json_name: CommonName
go_name: CommonName
go_tags: json:"CommonName" validate:"required"</pre></td>
</tr><tr>
<td>dns_names</td>
<td>string</td>
<td><pre>
dns_names are Subject Alternative Names for TLS certificates<br>

json_name: DnsNames
go_name: DnsNames
go_tags: json:"DnsNames,omitempty"</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
organization is the O for the certificate subject<br>

json_name: Organization
go_name: Organization
go_tags: json:"Organization"</pre></td>
</tr><tr>
<td>organizational_unit</td>
<td>string</td>
<td><pre>
organizational_unit is the OU for the certificate subject<br>

json_name: OrganizationalUnit
go_name: OrganizationalUnit
go_tags: json:"OrganizationalUnit"</pre></td>
</tr>
</table>



<a name="fabric-v1-createadmininput"></a>
### fabric.v1.CreateAdminInput

<pre>
CreateAdminInput specifies the parameters for creating a permanent admin
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to create the admin for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createadminoutput"></a>
### fabric.v1.CreateAdminOutput

<pre>
CreateAdminOutput indicates the result of admin creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the admin's signing identity<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the admin's TLS client certificate<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createbootstrapadmininput"></a>
### fabric.v1.CreateBootstrapAdminInput

<pre>
CreateBootstrapAdminInput specifies the parameters for creating a bootstrap admin
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to create the bootstrap admin for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createbootstrapadminoutput"></a>
### fabric.v1.CreateBootstrapAdminOutput

<pre>
CreateBootstrapAdminOutput indicates the result of bootstrap admin creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the bootstrap admin's signing identity<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the bootstrap admin's TLS client certificate<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createcertificatedatabaseinput"></a>
### fabric.v1.CreateCertificateDatabaseInput

<pre>
CreateCertificateDatabaseInput specifies the parameters for creating a certificate database
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA this database is for<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createcertificatedatabaseoutput"></a>
### fabric.v1.CreateCertificateDatabaseOutput

<pre>
CreateCertificateDatabaseOutput indicates the result of certificate database creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref is the Vault reference to retrieve the username and password<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createcertificateinput"></a>
### fabric.v1.CreateCertificateInput

<pre>
CreateCertificateInput specifies the parameters for creating a certificate
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref is the issuer's admin credentials (for registration)
Required unless self-signed<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef"</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-v1-createcertificateinput-attributesentry">fabric.v1.CreateCertificateInput.AttributesEntry</a></td>
<td><pre>
attributes are additional CA attributes for the identity (e.g., hf.IntermediateCA)<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the Fabric CA identity type for registration.
This determines the OU field in the issued certificate.
Valid values: "peer", "orderer", "admin", "client" (default: "client")
The OU is critical for NodeOUs classification in the MSP.<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType,omitempty"</pre></td>
</tr><tr>
<td>issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
issuer_ref identifies which CA signs this certificate.
If nil/empty, creates a self-signed certificate.<br>

json_name: IssuerRef
go_name: IssuerRef
go_tags: json:"IssuerRef"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile determines key usages ("ca", "tls", or "" for default identity)<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile"</pre></td>
</tr><tr>
<td>purpose</td>
<td>string</td>
<td><pre>
purpose is a human-readable label for the workflow ID (e.g., "tls-root-signing")<br>

json_name: Purpose
go_name: Purpose
go_tags: json:"Purpose" validate:"required"</pre></td>
</tr><tr>
<td>storage_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_ref is where to store keypair + certificate<br>

json_name: StorageRef
go_name: StorageRef
go_tags: json:"StorageRef" validate:"required"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createcertificateinput-attributesentry"></a>
### fabric.v1.CreateCertificateInput.AttributesEntry

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



<a name="fabric-v1-createcertificateoutput"></a>
### fabric.v1.CreateCertificateOutput

<pre>
CreateCertificateOutput indicates the result of certificate creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref points to the stored keypair + certificate<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-createchannelinput"></a>
### fabric.v1.CreateChannelInput



<a name="fabric-v1-createchanneloutput"></a>
### fabric.v1.CreateChannelOutput



<a name="fabric-v1-createidentityinput"></a>
### fabric.v1.CreateIdentityInput

<pre>
CreateIdentityInput specifies the parameters for creating a Fabric identity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>attributes</td>
<td><a href="#fabric-v1-createidentityinput-attributesentry">fabric.v1.CreateIdentityInput.AttributesEntry</a></td>
<td><pre>
attributes are additional CA attributes for the identity<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the unique name for this identity (e.g., "peer0", "orderer0", "admin")<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the Fabric CA identity type for registration.
This determines the OU field in the issued certificate.
Valid values: "peer", "orderer", "admin", "client" (default: "client")
The OU is critical for NodeOUs classification in the MSP.<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType,omitempty"</pre></td>
</tr><tr>
<td>org_issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
org_issuer_ref identifies which Org CA issues the identity certificate
Typically the Org Intermediate CA (org/v1)<br>

json_name: OrgIssuerRef
go_name: OrgIssuerRef
go_tags: json:"OrgIssuerRef" validate:"required"</pre></td>
</tr><tr>
<td>storage_signing_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_signing_ref is where to store the org identity (keypair + cert)<br>

json_name: StorageSigningRef
go_name: StorageSigningRef
go_tags: json:"StorageSigningRef" validate:"required"</pre></td>
</tr><tr>
<td>storage_tls_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
storage_tls_ref is where to store the TLS identity (keypair + cert)<br>

json_name: StorageTlsRef
go_name: StorageTlsRef
go_tags: json:"StorageTlsRef" validate:"required"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr><tr>
<td>tls_issuer_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
tls_issuer_ref identifies which TLS CA issues the TLS certificate
Typically the TLS Intermediate CA (tls/v1)<br>

json_name: TlsIssuerRef
go_name: TlsIssuerRef
go_tags: json:"TlsIssuerRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_profile</td>
<td>string</td>
<td><pre>
tls_profile specifies the TLS signing profile ("tls" for server/client, "" for client-only)<br>

json_name: TlsProfile
go_name: TlsProfile
go_tags: json:"TlsProfile"</pre></td>
</tr>
</table>



<a name="fabric-v1-createidentityinput-attributesentry"></a>
### fabric.v1.CreateIdentityInput.AttributesEntry

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



<a name="fabric-v1-createidentityoutput"></a>
### fabric.v1.CreateIdentityOutput

<pre>
CreateIdentityOutput indicates the result of identity creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>signing_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
signing_credentials_ref points to the org identity credentials in Vault<br>

json_name: SigningCredentialsRef
go_name: SigningCredentialsRef
go_tags: json:"SigningCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>tls_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
tls_credentials_ref points to the TLS identity credentials in Vault<br>

json_name: TlsCredentialsRef
go_name: TlsCredentialsRef
go_tags: json:"TlsCredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycertificateauthorityinput"></a>
### fabric.v1.DeployCertificateAuthorityInput

<pre>
DeployCertificateAuthorityInput specifies the parameters for deploying a CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to deploy<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycertificateauthorityoutput"></a>
### fabric.v1.DeployCertificateAuthorityOutput

<pre>
DeployCertificateAuthorityOutput indicates the result of CA deployment
</pre>



<a name="fabric-v1-deploycertificatestoreinput"></a>
### fabric.v1.DeployCertificateStoreInput

<pre>
DeployCertificateStoreInput specifies the parameters for deploying the certificate store
</pre>



<a name="fabric-v1-deploycertificatestoreoutput"></a>
### fabric.v1.DeployCertificateStoreOutput

<pre>
DeployCertificateStoreOutput indicates the result of certificate store deployment
</pre>



<a name="fabric-v1-deploychaincodeinput"></a>
### fabric.v1.DeployChainCodeInput

<pre>
DeployChainCodeInput specifies the parameters for deploying a chaincode
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>chain_code_id</td>
<td>string</td>
<td><pre>
chain_code_id identifies this chaincode (e.g., "asset-transfer", "fabcar")<br>

json_name: ChainCodeId
go_name: ChainCodeId
go_tags: json:"ChainCodeId" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deploychaincodeoutput"></a>
### fabric.v1.DeployChainCodeOutput

<pre>
DeployChainCodeOutput indicates the result of chaincode deployment
</pre>



<a name="fabric-v1-deploycouchdbinput"></a>
### fabric.v1.DeployCouchDbInput

<pre>
DeployCouchDbInput specifies the parameters for deploying a CouchDB instance
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
peer_id identifies which peer this CouchDB is for (e.g., "peer0", "peer1")<br>

json_name: PeerId
go_name: PeerId
go_tags: json:"PeerId" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycouchdboutput"></a>
### fabric.v1.DeployCouchDbOutput

<pre>
DeployCouchDbOutput indicates the result of CouchDB deployment
</pre>



<a name="fabric-v1-deployordererinput"></a>
### fabric.v1.DeployOrdererInput

<pre>
DeployOrdererInput specifies the parameters for deploying a Fabric orderer
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
orderer_id identifies this orderer (e.g., "orderer0", "orderer1")<br>

json_name: OrdererId
go_name: OrdererId
go_tags: json:"OrdererId" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deployordereroutput"></a>
### fabric.v1.DeployOrdererOutput

<pre>
DeployOrdererOutput indicates the result of orderer deployment
</pre>



<a name="fabric-v1-deploypeerinput"></a>
### fabric.v1.DeployPeerInput

<pre>
DeployPeerInput specifies the parameters for deploying a Fabric peer
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
peer_id identifies this peer (e.g., "peer0", "peer1")<br>

json_name: PeerId
go_name: PeerId
go_tags: json:"PeerId" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-deploypeeroutput"></a>
### fabric.v1.DeployPeerOutput

<pre>
DeployPeerOutput indicates the result of peer deployment
</pre>



<a name="fabric-v1-enrollidentityinput"></a>
### fabric.v1.EnrollIdentityInput

<pre>
EnrollIdentityInput specifies the parameters for enrolling an identity with a CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to enroll with<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_tls_cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
ca_tls_cert_ref points to the CA's TLS certificate in Vault<br>

json_name: CaTlsCertRef
go_name: CaTlsCertRef
go_tags: json:"CaTlsCertRef" validate:"required"</pre></td>
</tr><tr>
<td>chain_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
chain_ref is an optional location to store the CA chain separately.
Required for intermediate CA signing certs (profile="ca") so Fabric CA
can verify client certificates issued by the intermediate CA.
The chain contains parent CA certs needed for x509 chain verification.<br>

json_name: ChainRef
go_name: ChainRef
go_tags: json:"ChainRef,omitempty"</pre></td>
</tr><tr>
<td>enrollment_secret</td>
<td>string</td>
<td><pre>
enrollment_secret is the secret obtained from registration<br>

json_name: EnrollmentSecret
go_name: EnrollmentSecret
go_tags: json:"EnrollmentSecret" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the name of the identity to enroll<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref points to the private key to use for CSR generation<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
profile specifies which CA signing profile to use for the certificate.
Use "ca" for intermediate CA certificates (grants cert sign, crl sign key usages).
Use "tls" for TLS certificates (grants server auth, client auth key usages).
Leave empty to use the default profile (digital signature only).<br>

json_name: Profile
go_name: Profile
go_tags: json:"Profile,omitempty"</pre></td>
</tr><tr>
<td>subject</td>
<td><a href="#fabric-v1-certificatesubject">fabric.v1.CertificateSubject</a></td>
<td><pre>
subject specifies the certificate subject fields<br>

json_name: Subject
go_name: Subject
go_tags: json:"Subject" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-enrollidentityoutput"></a>
### fabric.v1.EnrollIdentityOutput

<pre>
EnrollIdentityOutput contains the certificates returned by the CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref points to where the certificate was stored<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-joinorderertochannelinput"></a>
### fabric.v1.JoinOrdererToChannelInput



<a name="fabric-v1-joinorderertochanneloutput"></a>
### fabric.v1.JoinOrdererToChannelOutput



<a name="fabric-v1-joinpeertochannelinput"></a>
### fabric.v1.JoinPeerToChannelInput



<a name="fabric-v1-joinpeertochanneloutput"></a>
### fabric.v1.JoinPeerToChannelOutput



<a name="fabric-v1-registeridentityinput"></a>
### fabric.v1.RegisterIdentityInput

<pre>
RegisterIdentityInput specifies the parameters for registering an identity with a CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref specifies which admin credentials to use for registration<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
affiliation is the affiliation for the identity in the CA<br>

json_name: Affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-v1-registeridentityinput-attributesentry">fabric.v1.RegisterIdentityInput.AttributesEntry</a></td>
<td><pre>
attributes are additional attributes for the identity<br>

json_name: Attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to register with<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the unique name for this identity<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>identity_type</td>
<td>string</td>
<td><pre>
identity_type specifies the type of identity (e.g., "admin", "peer", "orderer", "client")<br>

json_name: IdentityType
go_name: IdentityType
go_tags: json:"IdentityType"</pre></td>
</tr>
</table>



<a name="fabric-v1-registeridentityinput-attributesentry"></a>
### fabric.v1.RegisterIdentityInput.AttributesEntry

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



<a name="fabric-v1-registeridentityoutput"></a>
### fabric.v1.RegisterIdentityOutput

<pre>
RegisterIdentityOutput contains the enrollment secret returned by the CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>enrollment_secret</td>
<td>string</td>
<td><pre>
enrollment_secret is the secret used to enroll the identity<br>

json_name: EnrollmentSecret
go_name: EnrollmentSecret
go_tags: json:"EnrollmentSecret" validate:"required"</pre></td>
</tr>
</table>



<a name="fabric-v1-revokeidentityinput"></a>
### fabric.v1.RevokeIdentityInput

<pre>
RevokeIdentityInput specifies the parameters for revoking an identity from a CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
admin_credentials_ref specifies which admin credentials to use for revocation<br>

json_name: AdminCredentialsRef
go_name: AdminCredentialsRef
go_tags: json:"AdminCredentialsRef" validate:"required"</pre></td>
</tr><tr>
<td>ca_ref</td>
<td><a href="#fabric-v1-certificateauthorityref">fabric.v1.CertificateAuthorityRef</a></td>
<td><pre>
ca_ref identifies which CA to revoke from<br>

json_name: CaRef
go_name: CaRef
go_tags: json:"CaRef" validate:"required"</pre></td>
</tr><tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
identity_name is the name of the identity to revoke<br>

json_name: IdentityName
go_name: IdentityName
go_tags: json:"IdentityName" validate:"required"</pre></td>
</tr><tr>
<td>reason</td>
<td>string</td>
<td><pre>
reason is the reason for revocation (optional)<br>

json_name: Reason
go_name: Reason</pre></td>
</tr>
</table>



<a name="fabric-v1-revokeidentityoutput"></a>
### fabric.v1.RevokeIdentityOutput

<pre>
RevokeIdentityOutput indicates the result of identity revocation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>revoked</td>
<td>bool</td>
<td><pre>
revoked indicates if the identity was successfully revoked<br>

json_name: Revoked
go_name: Revoked</pre></td>
</tr>
</table>



<a name="fabric-v1-setanchorpeerinput"></a>
### fabric.v1.SetAnchorPeerInput



<a name="fabric-v1-setanchorpeeroutput"></a>
### fabric.v1.SetAnchorPeerOutput



<a name="fabric-v1-updatechannelconfiginput"></a>
### fabric.v1.UpdateChannelConfigInput



<a name="fabric-v1-updatechannelconfigoutput"></a>
### fabric.v1.UpdateChannelConfigOutput



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

json_name: ChartName
go_name: ChartName
go_tags: json:"ChartName" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>values_yaml</td>
<td>string</td>
<td><pre>
values_yaml is Helm values as a YAML string (merged with chart defaults)<br>

json_name: ValuesYaml
go_name: ValuesYaml
go_tags: json:"ValuesYaml" validate:"required"</pre></td>
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

json_name: Failures
go_name: Failures
go_tags: json:"Failures" validate:"required"</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: ChartName
go_name: ChartName
go_tags: json:"ChartName" validate:"required"</pre></td>
</tr><tr>
<td>component</td>
<td>string</td>
<td><pre>
component is the component type whose values template to load and render<br>

json_name: Component
go_name: Component
go_tags: json:"Component" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>values</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
values is a structured map of template variables for rendering<br>

json_name: Values
go_name: Values
go_tags: json:"Values" validate:"required"</pre></td>
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

json_name: Failures
go_name: Failures
go_tags: json:"Failures" validate:"required"</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: ChartName
go_name: ChartName
go_tags: json:"ChartName" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>values_yaml</td>
<td>string</td>
<td><pre>
values_yaml is Helm values as a YAML string (merged with chart defaults)<br>

json_name: ValuesYaml
go_name: ValuesYaml
go_tags: json:"ValuesYaml" validate:"required"</pre></td>
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

json_name: Failures
go_name: Failures
go_tags: json:"Failures" validate:"required"</pre></td>
</tr><tr>
<td>resources</td>
<td><a href="#helm-v1-renderchartoutput-resourcesentry">helm.v1.RenderChartOutput.ResourcesEntry</a></td>
<td><pre>
resources is a map of resource_id → rendered resource
Key is caller-defined (e.g., "deployment-api", "service-api")<br>

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: ChartName
go_name: ChartName
go_tags: json:"ChartName" validate:"required"</pre></td>
</tr><tr>
<td>component</td>
<td>string</td>
<td><pre>
component is the component type whose values template to load and render<br>

json_name: Component
go_name: Component
go_tags: json:"Component" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the Kubernetes namespace for template context (optional)<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>values</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
values is a structured map of template variables for rendering<br>

json_name: Values
go_name: Values
go_tags: json:"Values" validate:"required"</pre></td>
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

json_name: ManifestYaml
go_name: ManifestYaml
go_tags: json:"ManifestYaml" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-createsummary">kube.v1.CreateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-updatesummary">kube.v1.UpdateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-applysummary">kube.v1.ApplySummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-deletesummary">kube.v1.DeleteSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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
resources to check status for (by resource identifier, not full YAML)<br>

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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
<td>resources_ready</td>
<td>string</td>
<td><pre>
resources_ready lists resource IDs that are ready<br>

json_name: ResourcesReady
go_name: ResourcesReady
go_tags: json:"ResourcesReady" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="kube-v1-execpod-activity"></a>
### kube.v1.ExecPod

<pre>
ExecPod executes a command in a pod container
</pre>

**Input:** [kube.v1.ExecPodInput](#kube-v1-execpodinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>command</td>
<td>string</td>
<td><pre>
command is the command and arguments to execute
Example: ["sh", "-c", "echo hello"]<br>

json_name: Command
go_name: Command
go_tags: json:"Command" validate:"required"</pre></td>
</tr><tr>
<td>container</td>
<td>string</td>
<td><pre>
container is the name of the container in the pod
Must be explicitly specified<br>

json_name: Container
go_name: Container
go_tags: json:"Container" validate:"required"</pre></td>
</tr><tr>
<td>container_id</td>
<td><a href="#kube-v1-containeridentifier">kube.v1.ContainerIdentifier</a></td>
<td><pre>
container_id identifies the pod to execute the command in<br>

json_name: ContainerId
go_name: ContainerId
go_tags: json:"ContainerId" validate:"required"</pre></td>
</tr><tr>
<td>stdin</td>
<td>bytes</td>
<td><pre>
stdin is optional input to send to the command<br>

json_name: Stdin
go_name: Stdin
go_tags: json:"Stdin,omitempty" validate:"omitempty"</pre></td>
</tr>
</table>

**Output:** [kube.v1.ExecPodOutput](#kube-v1-execpodoutput)

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
error_message contains error details if execution failed<br>

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>exit_code</td>
<td>int32</td>
<td><pre>
exit_code is the exit code of the command
0 indicates success, non-zero indicates failure<br>

json_name: ExitCode
go_name: ExitCode
go_tags: json:"ExitCode" validate:"required"</pre></td>
</tr><tr>
<td>stderr</td>
<td>bytes</td>
<td><pre>
stderr is the standard error output from the command<br>

json_name: Stderr
go_name: Stderr
go_tags: json:"Stderr,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>stdout</td>
<td>bytes</td>
<td><pre>
stdout is the standard output from the command<br>

json_name: Stdout
go_name: Stdout
go_tags: json:"Stdout,omitempty" validate:"omitempty"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-applysummary">kube.v1.ApplySummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: ResourceId
go_name: ResourceId
go_tags: json:"ResourceId" validate:"required"</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if applied successfully)<br>

json_name: ResourceStatus
go_name: ResourceStatus
go_tags: json:"ResourceStatus,omitempty" validate:"omitempty"</pre></td>
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
json_name: Applied
go_name: Applied
go_tags: json:"Applied" validate:"required"</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: Failed
go_name: Failed
go_tags: json:"Failed" validate:"required"</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: Total
go_name: Total
go_tags: json:"Total" validate:"required"</pre></td>
</tr>
</table>



<a name="kube-v1-containeridentifier"></a>
### kube.v1.ContainerIdentifier

<pre>
ContainerIdentifier uniquely identifies a pod (namespace + pod name)
Used by ExecPod and future pod-specific operations (e.g. logs, copy)
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
name is the pod name<br>

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the namespace containing the pod<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-createsummary">kube.v1.CreateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: ResourceId
go_name: ResourceId
go_tags: json:"ResourceId" validate:"required"</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if created successfully)<br>

json_name: ResourceStatus
go_name: ResourceStatus
go_tags: json:"ResourceStatus,omitempty" validate:"omitempty"</pre></td>
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
json_name: Created
go_name: Created
go_tags: json:"Created" validate:"required"</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: Failed
go_name: Failed
go_tags: json:"Failed" validate:"required"</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: Total
go_name: Total
go_tags: json:"Total" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-deletesummary">kube.v1.DeleteSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: ResourceId
go_name: ResourceId
go_tags: json:"ResourceId" validate:"required"</pre></td>
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
json_name: Deleted
go_name: Deleted
go_tags: json:"Deleted" validate:"required"</pre></td>
</tr><tr>
<td>failed</td>
<td>int32</td>
<td><pre>
json_name: Failed
go_name: Failed
go_tags: json:"Failed" validate:"required"</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: Total
go_name: Total
go_tags: json:"Total" validate:"required"</pre></td>
</tr>
</table>



<a name="kube-v1-execpodinput"></a>
### kube.v1.ExecPodInput

<pre>
ExecPodInput defines inputs for executing a command in a pod
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>command</td>
<td>string</td>
<td><pre>
command is the command and arguments to execute
Example: ["sh", "-c", "echo hello"]<br>

json_name: Command
go_name: Command
go_tags: json:"Command" validate:"required"</pre></td>
</tr><tr>
<td>container</td>
<td>string</td>
<td><pre>
container is the name of the container in the pod
Must be explicitly specified<br>

json_name: Container
go_name: Container
go_tags: json:"Container" validate:"required"</pre></td>
</tr><tr>
<td>container_id</td>
<td><a href="#kube-v1-containeridentifier">kube.v1.ContainerIdentifier</a></td>
<td><pre>
container_id identifies the pod to execute the command in<br>

json_name: ContainerId
go_name: ContainerId
go_tags: json:"ContainerId" validate:"required"</pre></td>
</tr><tr>
<td>stdin</td>
<td>bytes</td>
<td><pre>
stdin is optional input to send to the command<br>

json_name: Stdin
go_name: Stdin
go_tags: json:"Stdin,omitempty" validate:"omitempty"</pre></td>
</tr>
</table>



<a name="kube-v1-execpodoutput"></a>
### kube.v1.ExecPodOutput

<pre>
ExecPodOutput represents the result of executing a command in a pod
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
error_message contains error details if execution failed<br>

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>exit_code</td>
<td>int32</td>
<td><pre>
exit_code is the exit code of the command
0 indicates success, non-zero indicates failure<br>

json_name: ExitCode
go_name: ExitCode
go_tags: json:"ExitCode" validate:"required"</pre></td>
</tr><tr>
<td>stderr</td>
<td>bytes</td>
<td><pre>
stderr is the standard error output from the command<br>

json_name: Stderr
go_name: Stderr
go_tags: json:"Stderr,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>stdout</td>
<td>bytes</td>
<td><pre>
stdout is the standard output from the command<br>

json_name: Stdout
go_name: Stdout
go_tags: json:"Stdout,omitempty" validate:"omitempty"</pre></td>
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
resources to check status for (by resource identifier, not full YAML)<br>

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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
<td><a href="#kube-v1-resourceidentifier">kube.v1.ResourceIdentifier</a></td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="kube-v1-getstatusoutput"></a>
### kube.v1.GetStatusOutput

<pre>
GetStatusOutput represents the current status of resources
Activity fails with error if any resources are not ready
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>resources_ready</td>
<td>string</td>
<td><pre>
resources_ready lists resource IDs that are ready<br>

json_name: ResourcesReady
go_name: ResourcesReady
go_tags: json:"ResourcesReady" validate:"required"</pre></td>
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

json_name: ManifestYaml
go_name: ManifestYaml
go_tags: json:"ManifestYaml" validate:"required"</pre></td>
</tr>
</table>



<a name="kube-v1-resourceidentifier"></a>
### kube.v1.ResourceIdentifier

<pre>
ResourceIdentifier uniquely identifies a Kubernetes resource
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>api_version</td>
<td>string</td>
<td><pre>
api_version is the Kubernetes API version (e.g., "apps/v1", "v1")<br>

json_name: ApiVersion
go_name: ApiVersion
go_tags: json:"ApiVersion" validate:"required"</pre></td>
</tr><tr>
<td>kind</td>
<td>string</td>
<td><pre>
kind is the Kubernetes resource kind (e.g., "Deployment", "Service", "Pod")<br>

json_name: Kind
go_name: Kind
go_tags: json:"Kind" validate:"required"</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the resource name<br>

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the resource namespace
Empty for cluster-scoped resources<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace,omitempty" validate:"omitempty"</pre></td>
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

json_name: DesiredReplicas
go_name: DesiredReplicas
go_tags: json:"DesiredReplicas,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>kind</td>
<td>string</td>
<td><pre>
kind is the Kubernetes resource kind (e.g., "Deployment", "Pod", "Service")<br>

json_name: Kind
go_name: Kind
go_tags: json:"Kind" validate:"required"</pre></td>
</tr><tr>
<td>message</td>
<td>string</td>
<td><pre>
message provides additional context (error messages, warnings, etc.)<br>

json_name: Message
go_name: Message
go_tags: json:"Message,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name is the resource name<br>

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace is the resource namespace<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace" validate:"required"</pre></td>
</tr><tr>
<td>ready_replicas</td>
<td>int32</td>
<td><pre>
ready_replicas is the number of ready replicas (for Deployments, StatefulSets, etc.)
Only populated for workload resources<br>

json_name: ReadyReplicas
go_name: ReadyReplicas
go_tags: json:"ReadyReplicas,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>status</td>
<td>string</td>
<td><pre>
status is the current status (e.g., "ready", "pending", "error", "unknown")<br>

json_name: Status
go_name: Status
go_tags: json:"Status" validate:"required"</pre></td>
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

json_name: Resources
go_name: Resources
go_tags: json:"Resources" validate:"required"</pre></td>
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

json_name: Results
go_name: Results
go_tags: json:"Results" validate:"required"</pre></td>
</tr><tr>
<td>summary</td>
<td><a href="#kube-v1-updatesummary">kube.v1.UpdateSummary</a></td>
<td><pre>
summary provides high-level statistics<br>

json_name: Summary
go_name: Summary
go_tags: json:"Summary" validate:"required"</pre></td>
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

json_name: ErrorMessage
go_name: ErrorMessage
go_tags: json:"ErrorMessage,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>resource_id</td>
<td>string</td>
<td><pre>
resource_id identifies which resource this result is for<br>

json_name: ResourceId
go_name: ResourceId
go_tags: json:"ResourceId" validate:"required"</pre></td>
</tr><tr>
<td>resource_status</td>
<td><a href="#kube-v1-resourcestatus">kube.v1.ResourceStatus</a></td>
<td><pre>
resource_status contains current state (if updated successfully)<br>

json_name: ResourceStatus
go_name: ResourceStatus
go_tags: json:"ResourceStatus,omitempty" validate:"omitempty"</pre></td>
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
json_name: Failed
go_name: Failed
go_tags: json:"Failed" validate:"required"</pre></td>
</tr><tr>
<td>total</td>
<td>int32</td>
<td><pre>
json_name: Total
go_name: Total
go_tags: json:"Total" validate:"required"</pre></td>
</tr><tr>
<td>updated</td>
<td>int32</td>
<td><pre>
json_name: Updated
go_name: Updated
go_tags: json:"Updated" validate:"required"</pre></td>
</tr>
</table>



<a name="postgres-v1"></a>
# postgres.v1

<a name="postgres-v1-services"></a>
## Services

<a name="postgres-v1-postgres"></a>
## postgres.v1.Postgres

<pre>
Postgres service defines activities for PostgreSQL database management.
</pre>   

<a name="postgres-v1-postgres-activities"></a>
### Activities

---
<a name="postgres-v1-createdatabase-activity"></a>
### postgres.v1.CreateDatabase

<pre>
CreateDatabase creates a new PostgreSQL database.
Idempotent: returns success if database already exists.
</pre>

**Input:** [postgres.v1.CreateDatabaseInput](#postgres-v1-createdatabaseinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>database_name</td>
<td>string</td>
<td><pre>
json_name: DatabaseName
go_name: DatabaseName
go_tags: json:"DatabaseName" validate:"required"</pre></td>
</tr>
</table>

**Output:** [postgres.v1.CreateDatabaseOutput](#postgres-v1-createdatabaseoutput)

 

---
<a name="postgres-v1-createuser-activity"></a>
### postgres.v1.CreateUser

<pre>
CreateUser creates a new PostgreSQL user with credentials from a SecretRef.
Idempotent: returns success if user already exists.
</pre>

**Input:** [postgres.v1.CreateUserInput](#postgres-v1-createuserinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref points to secret containing username and password<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [postgres.v1.CreateUserOutput](#postgres-v1-createuseroutput)

 

---
<a name="postgres-v1-grantprivileges-activity"></a>
### postgres.v1.GrantPrivileges

<pre>
GrantPrivileges grants privileges to a user on a database.
Idempotent: can be called multiple times safely.
</pre>

**Input:** [postgres.v1.GrantPrivilegesInput](#postgres-v1-grantprivilegesinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>database_name</td>
<td>string</td>
<td><pre>
json_name: DatabaseName
go_name: DatabaseName
go_tags: json:"DatabaseName" validate:"required"</pre></td>
</tr><tr>
<td>privileges</td>
<td><a href="#postgres-v1-privilege">postgres.v1.Privilege</a></td>
<td><pre>
json_name: Privileges
go_name: Privileges
go_tags: json:"Privileges,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>username</td>
<td>string</td>
<td><pre>
json_name: Username
go_name: Username
go_tags: json:"Username" validate:"required"</pre></td>
</tr>
</table>

**Output:** [postgres.v1.GrantPrivilegesOutput](#postgres-v1-grantprivilegesoutput)

   

<a name="postgres-v1-messages"></a>
## Messages

<a name="postgres-v1-createdatabaseinput"></a>
### postgres.v1.CreateDatabaseInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>database_name</td>
<td>string</td>
<td><pre>
json_name: DatabaseName
go_name: DatabaseName
go_tags: json:"DatabaseName" validate:"required"</pre></td>
</tr>
</table>



<a name="postgres-v1-createdatabaseoutput"></a>
### postgres.v1.CreateDatabaseOutput



<a name="postgres-v1-createuserinput"></a>
### postgres.v1.CreateUserInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
credentials_ref points to secret containing username and password<br>

json_name: CredentialsRef
go_name: CredentialsRef
go_tags: json:"CredentialsRef" validate:"required"</pre></td>
</tr>
</table>



<a name="postgres-v1-createuseroutput"></a>
### postgres.v1.CreateUserOutput



<a name="postgres-v1-grantprivilegesinput"></a>
### postgres.v1.GrantPrivilegesInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>database_name</td>
<td>string</td>
<td><pre>
json_name: DatabaseName
go_name: DatabaseName
go_tags: json:"DatabaseName" validate:"required"</pre></td>
</tr><tr>
<td>privileges</td>
<td><a href="#postgres-v1-privilege">postgres.v1.Privilege</a></td>
<td><pre>
json_name: Privileges
go_name: Privileges
go_tags: json:"Privileges,omitempty" validate:"omitempty"</pre></td>
</tr><tr>
<td>username</td>
<td>string</td>
<td><pre>
json_name: Username
go_name: Username
go_tags: json:"Username" validate:"required"</pre></td>
</tr>
</table>



<a name="postgres-v1-grantprivilegesoutput"></a>
### postgres.v1.GrantPrivilegesOutput



<a name="postgres-v1-privilege"></a>
### postgres.v1.Privilege

<table>
<tr><th>Value</th><th>Description</th></tr>
<tr>
<td>PRIVILEGE_UNSPECIFIED</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_ALL</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_SELECT</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_INSERT</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_UPDATE</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_DELETE</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_TRUNCATE</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_REFERENCES</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_TRIGGER</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_CREATE</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_CONNECT</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_TEMPORARY</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_EXECUTE</td>
<td></td>
</tr><tr>
<td>PRIVILEGE_USAGE</td>
<td></td>
</tr>
</table>

<a name="secrets-v1"></a>
# secrets.v1

<a name="secrets-v1-services"></a>
## Services

<a name="secrets-v1-secrets"></a>
## secrets.v1.Secrets

<pre>
Secrets service provides unified secret management operations.
Includes secret creation/storage, key generation, and access management.
Nexus-enabled for cross-namespace calls (fabric workers calling CreateAccess on infra worker).
</pre>

<a name="secrets-v1-secrets-workflows"></a>
### Workflows

---
<a name="secrets-v1-createaccess-workflow"></a>
### secrets.v1.CreateAccess

<pre>
CreateAccess creates a Vault policy and Kubernetes auth role for secret access.
Orchestrates CreatePolicy and CreateAuthRole vault activities.
</pre>

**Input:** [secrets.v1.CreateAccessInput](#secrets-v1-createaccessinput)

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
auth_role_name is the name of the auth role to create<br>

json_name: AuthRoleName
go_name: AuthRoleName
go_tags: json:"AuthRoleName" validate:"required"</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is where the service account exists<br>

json_name: KubernetesNamespace
go_name: KubernetesNamespace
go_tags: json:"KubernetesNamespace" validate:"required"</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
policy_document contains the structured policy rules<br>

json_name: PolicyDocument
go_name: PolicyDocument
go_tags: json:"PolicyDocument" validate:"required"</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
policy_name is the name of the policy to create<br>

json_name: PolicyName
go_name: PolicyName
go_tags: json:"PolicyName" validate:"required"</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: ServiceAccountName
go_name: ServiceAccountName
go_tags: json:"ServiceAccountName" validate:"required"</pre></td>
</tr>
</table>

**Output:** [secrets.v1.CreateAccessOutput](#secrets-v1-createaccessoutput)

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
json_name: AuthRoleName
go_name: AuthRoleName
go_tags: json:"AuthRoleName" validate:"required"</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
json_name: PolicyName
go_name: PolicyName
go_tags: json:"PolicyName" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>secrets/create-access/${! PolicyName.or("default") }</code></pre></td></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
</table>    

<a name="secrets-v1-secrets-activities"></a>
### Activities

---
<a name="secrets-v1-createpassword-activity"></a>
### secrets.v1.CreatePassword

<pre>
CreatePassword generates a secure password and stores it with the given username.
</pre>

**Input:** [secrets.v1.CreatePasswordInput](#secrets-v1-createpasswordinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
secret_ref is where credentials should be stored<br>

json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr><tr>
<td>username</td>
<td>string</td>
<td><pre>
username to store alongside the generated password<br>

json_name: Username
go_name: Username
go_tags: json:"Username" validate:"required"</pre></td>
</tr>
</table>

**Output:** [secrets.v1.CreatePasswordOutput](#secrets-v1-createpasswordoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="secrets-v1-createkeypair-activity"></a>
### secrets.v1.CreateKeyPair

<pre>
CreateKeyPair generates an ECDSA key pair (P-256) and stores it.
</pre>

**Input:** [secrets.v1.CreateKeyPairInput](#secrets-v1-createkeypairinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
secret_ref is where the key pair should be stored<br>

json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table>

**Output:** [secrets.v1.CreateKeyPairOutput](#secrets-v1-createkeypairoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="secrets-v1-createselfsignedcertificate-activity"></a>
### secrets.v1.CreateSelfSignedCertificate

<pre>
CreateSelfSignedCertificate fetches a key pair, generates a self-signed certificate, and stores it.
Used for root CA certificates only. Implemented as an activity to keep secrets out of workflow history.
</pre>

**Input:** [secrets.v1.CreateSelfSignedCertificateInput](#secrets-v1-createselfsignedcertificateinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the generated certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>common_name</td>
<td>string</td>
<td><pre>
common_name is the CN field in the certificate subject<br>

json_name: CommonName
go_name: CommonName
go_tags: json:"CommonName" validate:"required"</pre></td>
</tr><tr>
<td>dns_names</td>
<td>string</td>
<td><pre>
dns_names are Subject Alternative Names (SANs) for TLS server certificates
Required for TLS certificate validation - modern clients require SANs even if CN matches<br>

json_name: DnsNames
go_name: DnsNames
go_tags: json:"DnsNames"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref is where to fetch the private key from<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>max_path_length</td>
<td>int32</td>
<td><pre>
max_path_length sets the pathLenConstraint in BasicConstraints<br>

json_name: MaxPathLength
go_name: MaxPathLength
go_tags: json:"MaxPathLength" validate:"min=0"</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
organization is the O field in the certificate subject<br>

json_name: Organization
go_name: Organization
go_tags: json:"Organization" validate:"required"</pre></td>
</tr><tr>
<td>organizational_unit</td>
<td>string</td>
<td><pre>
organizational_unit is the OU field in the certificate subject<br>

json_name: OrganizationalUnit
go_name: OrganizationalUnit
go_tags: json:"OrganizationalUnit"</pre></td>
</tr><tr>
<td>validity_duration</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
validity_duration is how long the certificate is valid<br>

json_name: ValidityDuration
go_name: ValidityDuration
go_tags: json:"ValidityDuration" validate:"required"</pre></td>
</tr>
</table>

**Output:** [secrets.v1.CreateSelfSignedCertificateOutput](#secrets-v1-createselfsignedcertificateoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="secrets-v1-storecertificate-activity"></a>
### secrets.v1.StoreCertificate

<pre>
StoreCertificate stores certificate data at the specified location.
Implemented as an activity to keep certificate data out of workflow history.
</pre>

**Input:** [secrets.v1.StoreCertificateInput](#secrets-v1-storecertificateinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_chain_pem</td>
<td>bytes</td>
<td><pre>
ca_chain_pem is the optional PEM-encoded CA certificate chain<br>

json_name: CaChainPem
go_name: CaChainPem
go_tags: json:"CaChainPem"</pre></td>
</tr><tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>certificate_pem</td>
<td>bytes</td>
<td><pre>
certificate_pem is the PEM-encoded certificate to store<br>

json_name: CertificatePem
go_name: CertificatePem
go_tags: json:"CertificatePem" validate:"required"</pre></td>
</tr>
</table>

**Output:** [secrets.v1.StoreCertificateOutput](#secrets-v1-storecertificateoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table>   

<a name="secrets-v1-messages"></a>
## Messages

<a name="secrets-v1-createaccessinput"></a>
### secrets.v1.CreateAccessInput

<pre>
CreateAccessInput specifies parameters for creating secret access (policy + auth role)
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
auth_role_name is the name of the auth role to create<br>

json_name: AuthRoleName
go_name: AuthRoleName
go_tags: json:"AuthRoleName" validate:"required"</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is where the service account exists<br>

json_name: KubernetesNamespace
go_name: KubernetesNamespace
go_tags: json:"KubernetesNamespace" validate:"required"</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
policy_document contains the structured policy rules<br>

json_name: PolicyDocument
go_name: PolicyDocument
go_tags: json:"PolicyDocument" validate:"required"</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
policy_name is the name of the policy to create<br>

json_name: PolicyName
go_name: PolicyName
go_tags: json:"PolicyName" validate:"required"</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate<br>

json_name: ServiceAccountName
go_name: ServiceAccountName
go_tags: json:"ServiceAccountName" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createaccessoutput"></a>
### secrets.v1.CreateAccessOutput

<pre>
CreateAccessOutput confirms access was created
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
json_name: AuthRoleName
go_name: AuthRoleName
go_tags: json:"AuthRoleName" validate:"required"</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
json_name: PolicyName
go_name: PolicyName
go_tags: json:"PolicyName" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createkeypairinput"></a>
### secrets.v1.CreateKeyPairInput

<pre>
CreateKeyPairInput specifies where to store the generated key pair
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
secret_ref is where the key pair should be stored<br>

json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createkeypairoutput"></a>
### secrets.v1.CreateKeyPairOutput

<pre>
CreateKeyPairOutput confirms key pair was stored
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createpasswordinput"></a>
### secrets.v1.CreatePasswordInput

<pre>
CreatePasswordInput specifies where to store generated credentials
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
secret_ref is where credentials should be stored<br>

json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr><tr>
<td>username</td>
<td>string</td>
<td><pre>
username to store alongside the generated password<br>

json_name: Username
go_name: Username
go_tags: json:"Username" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createpasswordoutput"></a>
### secrets.v1.CreatePasswordOutput

<pre>
CreatePasswordOutput confirms credentials were stored
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: SecretRef
go_name: SecretRef
go_tags: json:"SecretRef" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createselfsignedcertificateinput"></a>
### secrets.v1.CreateSelfSignedCertificateInput

<pre>
CreateSelfSignedCertificateInput specifies key location, cert destination, and certificate parameters
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the generated certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>common_name</td>
<td>string</td>
<td><pre>
common_name is the CN field in the certificate subject<br>

json_name: CommonName
go_name: CommonName
go_tags: json:"CommonName" validate:"required"</pre></td>
</tr><tr>
<td>dns_names</td>
<td>string</td>
<td><pre>
dns_names are Subject Alternative Names (SANs) for TLS server certificates
Required for TLS certificate validation - modern clients require SANs even if CN matches<br>

json_name: DnsNames
go_name: DnsNames
go_tags: json:"DnsNames"</pre></td>
</tr><tr>
<td>key_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
key_ref is where to fetch the private key from<br>

json_name: KeyRef
go_name: KeyRef
go_tags: json:"KeyRef" validate:"required"</pre></td>
</tr><tr>
<td>max_path_length</td>
<td>int32</td>
<td><pre>
max_path_length sets the pathLenConstraint in BasicConstraints<br>

json_name: MaxPathLength
go_name: MaxPathLength
go_tags: json:"MaxPathLength" validate:"min=0"</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
organization is the O field in the certificate subject<br>

json_name: Organization
go_name: Organization
go_tags: json:"Organization" validate:"required"</pre></td>
</tr><tr>
<td>organizational_unit</td>
<td>string</td>
<td><pre>
organizational_unit is the OU field in the certificate subject<br>

json_name: OrganizationalUnit
go_name: OrganizationalUnit
go_tags: json:"OrganizationalUnit"</pre></td>
</tr><tr>
<td>validity_duration</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
validity_duration is how long the certificate is valid<br>

json_name: ValidityDuration
go_name: ValidityDuration
go_tags: json:"ValidityDuration" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-createselfsignedcertificateoutput"></a>
### secrets.v1.CreateSelfSignedCertificateOutput

<pre>
CreateSelfSignedCertificateOutput confirms certificate was stored
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-secretref"></a>
### secrets.v1.SecretRef

<pre>
SecretRef is a backend-agnostic reference to a secret.
The ref field is a structured object interpreted based on type.
Examples:
  - type: "vault", ref: {path: "fabric/org1/ca/tls/root/admin", version: 1}
  - type: "k8s", ref: {namespace: "default", name: "my-secret", key: "password"}
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ref</td>
<td><a href="#google-protobuf-struct">google.protobuf.Struct</a></td>
<td><pre>
ref is backend-specific structured reference data<br>

json_name: Ref
go_name: Ref
go_tags: json:"Ref" validate:"required"</pre></td>
</tr><tr>
<td>type</td>
<td>string</td>
<td><pre>
type identifies the secret backend (e.g., "vault", "k8s")<br>

json_name: Type
go_name: Type
go_tags: json:"Type" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-storecertificateinput"></a>
### secrets.v1.StoreCertificateInput

<pre>
StoreCertificateInput specifies certificate data to store
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_chain_pem</td>
<td>bytes</td>
<td><pre>
ca_chain_pem is the optional PEM-encoded CA certificate chain<br>

json_name: CaChainPem
go_name: CaChainPem
go_tags: json:"CaChainPem"</pre></td>
</tr><tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
cert_ref is where to store the certificate<br>

json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
</tr><tr>
<td>certificate_pem</td>
<td>bytes</td>
<td><pre>
certificate_pem is the PEM-encoded certificate to store<br>

json_name: CertificatePem
go_name: CertificatePem
go_tags: json:"CertificatePem" validate:"required"</pre></td>
</tr>
</table>



<a name="secrets-v1-storecertificateoutput"></a>
### secrets.v1.StoreCertificateOutput

<pre>
StoreCertificateOutput confirms certificate was stored
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cert_ref</td>
<td><a href="#secrets-v1-secretref">secrets.v1.SecretRef</a></td>
<td><pre>
json_name: CertRef
go_name: CertRef
go_tags: json:"CertRef" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: Retention
go_name: Retention
go_tags: json:"Retention,omitempty" validate:"omitempty"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id</td><td><pre><code>temporal/create-namespace/${! Name }</code></pre></td></tr>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: Retention
go_name: Retention
go_tags: json:"Retention,omitempty" validate:"omitempty"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>state</td>
<td><a href="#temporal-v1-namespacestate">temporal.v1.NamespaceState</a></td>
<td><pre>
state is the current lifecycle state of the namespace<br>

json_name: State
go_name: State
go_tags: json:"State" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route to<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace" validate:"required"</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: TaskQueue
go_name: TaskQueue
go_tags: json:"TaskQueue" validate:"required"</pre></td>
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

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
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

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
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
json_name: Endpoint
go_name: Endpoint
go_tags: json:"Endpoint" validate:"required"</pre></td>
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

json_name: Endpoint
go_name: Endpoint
go_tags: json:"Endpoint" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: Retention
go_name: Retention
go_tags: json:"Retention,omitempty" validate:"omitempty"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route to<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace" validate:"required"</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: TaskQueue
go_name: TaskQueue
go_tags: json:"TaskQueue" validate:"required"</pre></td>
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

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>state</td>
<td><a href="#temporal-v1-namespacestate">temporal.v1.NamespaceState</a></td>
<td><pre>
state is the current lifecycle state of the namespace<br>

json_name: State
go_name: State
go_tags: json:"State" validate:"required"</pre></td>
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

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
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
json_name: Endpoint
go_name: Endpoint
go_tags: json:"Endpoint" validate:"required"</pre></td>
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

json_name: Id
go_name: Id
go_tags: json:"Id" validate:"required"</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
name: endpoint name (e.g., "CreateScopedIdentityAccess")<br>

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
namespace: Temporal namespace to route requests to<br>

json_name: Namespace
go_name: Namespace
go_tags: json:"Namespace" validate:"required"</pre></td>
</tr><tr>
<td>task_queue</td>
<td>string</td>
<td><pre>
task_queue: task queue within target namespace<br>

json_name: TaskQueue
go_name: TaskQueue
go_tags: json:"TaskQueue" validate:"required"</pre></td>
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

json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr><tr>
<td>retention</td>
<td><a href="#google-protobuf-duration">google.protobuf.Duration</a></td>
<td><pre>
retention is the duration to retain workflow history
If not set, defaults to 30 days (2592000 seconds)<br>

json_name: Retention
go_name: Retention
go_tags: json:"Retention,omitempty" validate:"omitempty"</pre></td>
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

json_name: Endpoint
go_name: Endpoint
go_tags: json:"Endpoint" validate:"required"</pre></td>
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
Vault service defines low-level Vault operations for auth and policy management.
These activities are used internally by the secrets service.
</pre>   

<a name="vault-v1-vault-activities"></a>
### Activities

---
<a name="vault-v1-createauthrole-activity"></a>
### vault.v1.CreateAuthRole

<pre>
CreateAuthRole creates a Kubernetes auth role binding a service account to policies
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
json_name: KubernetesNamespace
go_name: KubernetesNamespace
go_tags: json:"KubernetesNamespace" validate:"required"</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
json_name: PolicyNames
go_name: PolicyNames
go_tags: json:"PolicyNames" validate:"required"</pre></td>
</tr><tr>
<td>role_name</td>
<td>string</td>
<td><pre>
json_name: RoleName
go_name: RoleName
go_tags: json:"RoleName" validate:"required"</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: ServiceAccountName
go_name: ServiceAccountName
go_tags: json:"ServiceAccountName" validate:"required"</pre></td>
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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="vault-v1-deleteauthrole-activity"></a>
### vault.v1.DeleteAuthRole

<pre>
DeleteAuthRole deletes an existing Kubernetes auth role
</pre>

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>

**Output:** [vault.v1.DeleteAuthRoleOutput](#vault-v1-deleteauthroleoutput)

 

---
<a name="vault-v1-createpolicy-activity"></a>
### vault.v1.CreatePolicy

<pre>
CreatePolicy creates a new Vault policy with the specified document
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
json_name: Document
go_name: Document
go_tags: json:"Document" validate:"required"</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table> 

---
<a name="vault-v1-deletepolicy-activity"></a>
### vault.v1.DeletePolicy

<pre>
DeletePolicy deletes an existing Vault policy
</pre>

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>

**Output:** [vault.v1.DeletePolicyOutput](#vault-v1-deletepolicyoutput)

   

<a name="vault-v1-messages"></a>
## Messages

<a name="vault-v1-capability"></a>
### vault.v1.Capability

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

<a name="vault-v1-createauthroleinput"></a>
### vault.v1.CreateAuthRoleInput

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
json_name: KubernetesNamespace
go_name: KubernetesNamespace
go_tags: json:"KubernetesNamespace" validate:"required"</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
json_name: PolicyNames
go_name: PolicyNames
go_tags: json:"PolicyNames" validate:"required"</pre></td>
</tr><tr>
<td>role_name</td>
<td>string</td>
<td><pre>
json_name: RoleName
go_name: RoleName
go_tags: json:"RoleName" validate:"required"</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: ServiceAccountName
go_name: ServiceAccountName
go_tags: json:"ServiceAccountName" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-createauthroleoutput"></a>
### vault.v1.CreateAuthRoleOutput

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-createpolicyinput"></a>
### vault.v1.CreatePolicyInput

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
json_name: Document
go_name: Document
go_tags: json:"Document" validate:"required"</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-createpolicyoutput"></a>
### vault.v1.CreatePolicyOutput

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-deleteauthroleinput"></a>
### vault.v1.DeleteAuthRoleInput

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-deleteauthroleoutput"></a>
### vault.v1.DeleteAuthRoleOutput



<a name="vault-v1-deletepolicyinput"></a>
### vault.v1.DeletePolicyInput

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
json_name: Name
go_name: Name
go_tags: json:"Name" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-deletepolicyoutput"></a>
### vault.v1.DeletePolicyOutput



<a name="vault-v1-policydocument"></a>
### vault.v1.PolicyDocument

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
json_name: Rules
go_name: Rules
go_tags: json:"Rules" validate:"required"</pre></td>
</tr>
</table>



<a name="vault-v1-policyrule"></a>
### vault.v1.PolicyRule

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
json_name: Capabilities
go_name: Capabilities
go_tags: json:"Capabilities" validate:"required"</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
json_name: Path
go_name: Path
go_tags: json:"Path" validate:"required"</pre></td>
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

