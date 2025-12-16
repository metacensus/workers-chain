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
- [fabric.identity.v1](#fabric-identity-v1)
  - Services
    - [fabric.identity.v1.FabricIdentity](#fabric-identity-v1-fabricidentity)
      - [Workflows](#fabric-identity-v1-fabricidentity-workflows)
        - [fabric.identity.v1.ProvisionPeer](#fabric-identity-v1-provisionpeer-workflow)
        - [fabric.identity.v1.ProvisionOrderer](#fabric-identity-v1-provisionorderer-workflow)
        - [fabric.identity.v1.ProvisionClient](#fabric-identity-v1-provisionclient-workflow)
        - [fabric.identity.v1.BootstrapAdmin](#fabric-identity-v1-bootstrapadmin-workflow)
        - [fabric.identity.v1.ProvisionOrg](#fabric-identity-v1-provisionorg-workflow)
      - [Activities](#fabric-identity-v1-fabricidentity-activities)
        - [fabric.identity.v1.RegisterIdentity](#fabric-identity-v1-registeridentity-activity)
        - [fabric.identity.v1.EnrollIdentity](#fabric-identity-v1-enrollidentity-activity)
        - [fabric.identity.v1.AssembleMsp](#fabric-identity-v1-assemblemsp-activity)
        - [fabric.identity.v1.GenerateSecureSecret](#fabric-identity-v1-generatesecuresecret-activity)
        - [fabric.identity.v1.CreateAdminIdentity](#fabric-identity-v1-createadminidentity-activity)
        - [fabric.identity.v1.RevokeDefaultAdmin](#fabric-identity-v1-revokedefaultadmin-activity)
        - [fabric.identity.v1.CreateAffiliations](#fabric-identity-v1-createaffiliations-activity)
  - Messages
    - [fabric.identity.v1.AdminIdentity](#fabric-identity-v1-adminidentity)
    - [fabric.identity.v1.AssembleMspInput](#fabric-identity-v1-assemblemspinput)
    - [fabric.identity.v1.AssembleMspOutput](#fabric-identity-v1-assemblemspoutput)
    - [fabric.identity.v1.Attribute](#fabric-identity-v1-attribute)
    - [fabric.identity.v1.AttributeRequest](#fabric-identity-v1-attributerequest)
    - [fabric.identity.v1.BootstrapAdminInput](#fabric-identity-v1-bootstrapadmininput)
    - [fabric.identity.v1.BootstrapAdminOutput](#fabric-identity-v1-bootstrapadminoutput)
    - [fabric.identity.v1.CSRInfo](#fabric-identity-v1-csrinfo)
    - [fabric.identity.v1.CSRName](#fabric-identity-v1-csrname)
    - [fabric.identity.v1.CreateAdminIdentityInput](#fabric-identity-v1-createadminidentityinput)
    - [fabric.identity.v1.CreateAdminIdentityOutput](#fabric-identity-v1-createadminidentityoutput)
    - [fabric.identity.v1.CreateAffiliationsInput](#fabric-identity-v1-createaffiliationsinput)
    - [fabric.identity.v1.CreateAffiliationsOutput](#fabric-identity-v1-createaffiliationsoutput)
    - [fabric.identity.v1.EnrollIdentityInput](#fabric-identity-v1-enrollidentityinput)
    - [fabric.identity.v1.EnrollIdentityOutput](#fabric-identity-v1-enrollidentityoutput)
    - [fabric.identity.v1.GenerateSecureSecretInput](#fabric-identity-v1-generatesecuresecretinput)
    - [fabric.identity.v1.GenerateSecureSecretOutput](#fabric-identity-v1-generatesecuresecretoutput)
    - [fabric.identity.v1.KeyRequest](#fabric-identity-v1-keyrequest)
    - [fabric.identity.v1.NodeOUConfig](#fabric-identity-v1-nodeouconfig)
    - [fabric.identity.v1.OUIdentifier](#fabric-identity-v1-ouidentifier)
    - [fabric.identity.v1.ProvisionClientInput](#fabric-identity-v1-provisionclientinput)
    - [fabric.identity.v1.ProvisionClientOutput](#fabric-identity-v1-provisionclientoutput)
    - [fabric.identity.v1.ProvisionOrdererInput](#fabric-identity-v1-provisionordererinput)
    - [fabric.identity.v1.ProvisionOrdererOutput](#fabric-identity-v1-provisionordereroutput)
    - [fabric.identity.v1.ProvisionOrgInput](#fabric-identity-v1-provisionorginput)
    - [fabric.identity.v1.ProvisionOrgOutput](#fabric-identity-v1-provisionorgoutput)
    - [fabric.identity.v1.ProvisionPeerInput](#fabric-identity-v1-provisionpeerinput)
    - [fabric.identity.v1.ProvisionPeerOutput](#fabric-identity-v1-provisionpeeroutput)
    - [fabric.identity.v1.RegisterIdentityInput](#fabric-identity-v1-registeridentityinput)
    - [fabric.identity.v1.RegisterIdentityOutput](#fabric-identity-v1-registeridentityoutput)
    - [fabric.identity.v1.RevokeDefaultAdminInput](#fabric-identity-v1-revokedefaultadmininput)
    - [fabric.identity.v1.RevokeDefaultAdminOutput](#fabric-identity-v1-revokedefaultadminoutput)
- [fabric.v1](#fabric-v1)
  - Services
    - [fabric.v1.FabricDeploy](#fabric-v1-fabricdeploy)
      - [Workflows](#fabric-v1-fabricdeploy-workflows)
        - [fabric.v1.CreateNamespace](#fabric-v1-createnamespace-workflow)
        - [fabric.v1.CreateServiceAccount](#fabric-v1-createserviceaccount-workflow)
        - [fabric.v1.CreateVaultRole](#fabric-v1-createvaultrole-workflow)
        - [fabric.v1.DeployCA](#fabric-v1-deployca-workflow)
        - [fabric.v1.DeployCouchDB](#fabric-v1-deploycouchdb-workflow)
        - [fabric.v1.DeployPeer](#fabric-v1-deploypeer-workflow)
        - [fabric.v1.DeployOrderer](#fabric-v1-deployorderer-workflow)
        - [fabric.v1.DeployOrg](#fabric-v1-deployorg-workflow)
  - Messages
    - [fabric.v1.CreateNamespaceInput](#fabric-v1-createnamespaceinput)
    - [fabric.v1.CreateNamespaceOutput](#fabric-v1-createnamespaceoutput)
    - [fabric.v1.CreateServiceAccountInput](#fabric-v1-createserviceaccountinput)
    - [fabric.v1.CreateServiceAccountOutput](#fabric-v1-createserviceaccountoutput)
    - [fabric.v1.CreateVaultRoleInput](#fabric-v1-createvaultroleinput)
    - [fabric.v1.CreateVaultRoleOutput](#fabric-v1-createvaultroleoutput)
    - [fabric.v1.DeployCaInput](#fabric-v1-deploycainput)
    - [fabric.v1.DeployCaOutput](#fabric-v1-deploycaoutput)
    - [fabric.v1.DeployCouchDbInput](#fabric-v1-deploycouchdbinput)
    - [fabric.v1.DeployCouchDbOutput](#fabric-v1-deploycouchdboutput)
    - [fabric.v1.DeployOrdererInput](#fabric-v1-deployordererinput)
    - [fabric.v1.DeployOrdererOutput](#fabric-v1-deployordereroutput)
    - [fabric.v1.DeployOrgInput](#fabric-v1-deployorginput)
    - [fabric.v1.DeployOrgOutput](#fabric-v1-deployorgoutput)
    - [fabric.v1.DeployPeerInput](#fabric-v1-deploypeerinput)
    - [fabric.v1.DeployPeerOutput](#fabric-v1-deploypeeroutput)
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



<a name="fabric-identity-v1"></a>
# fabric.identity.v1

<a name="fabric-identity-v1-services"></a>
## Services

<a name="fabric-identity-v1-fabricidentity"></a>
## fabric.identity.v1.FabricIdentity

<pre>
FabricIdentity service defines workflows and activities for Hyperledger Fabric identity and certificate management
</pre>

<a name="fabric-identity-v1-fabricidentity-workflows"></a>
### Workflows

---
<a name="fabric-identity-v1-provisionpeer-workflow"></a>
### fabric.identity.v1.ProvisionPeer

<pre>
ProvisionPeer provisions a complete peer with all necessary certificates
</pre>

**Input:** [fabric.identity.v1.ProvisionPeerInput](#fabric-identity-v1-provisionpeerinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.ProvisionPeerOutput](#fabric-identity-v1-provisionpeeroutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-provisionorderer-workflow"></a>
### fabric.identity.v1.ProvisionOrderer

<pre>
ProvisionOrderer provisions a complete orderer with all necessary certificates
</pre>

**Input:** [fabric.identity.v1.ProvisionOrdererInput](#fabric-identity-v1-provisionordererinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
json_name: ordererId
go_name: OrdererId</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.ProvisionOrdererOutput](#fabric-identity-v1-provisionordereroutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-provisionclient-workflow"></a>
### fabric.identity.v1.ProvisionClient

<pre>
ProvisionClient provisions a client identity (simpler flow)
</pre>

**Input:** [fabric.identity.v1.ProvisionClientInput](#fabric-identity-v1-provisionclientinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>client_id</td>
<td>string</td>
<td><pre>
json_name: clientId
go_name: ClientId</pre></td>
</tr><tr>
<td>custom_attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: customAttributes
go_name: CustomAttributes</pre></td>
</tr><tr>
<td>enable_tls</td>
<td>bool</td>
<td><pre>
json_name: enableTls
go_name: EnableTls</pre></td>
</tr><tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.ProvisionClientOutput](#fabric-identity-v1-provisionclientoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-bootstrapadmin-workflow"></a>
### fabric.identity.v1.BootstrapAdmin

<pre>
BootstrapAdmin creates a new admin identity and revokes the default one
</pre>

**Input:** [fabric.identity.v1.BootstrapAdminInput](#fabric-identity-v1-bootstrapadmininput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_id</td>
<td>string</td>
<td><pre>
json_name: adminId
go_name: AdminId</pre></td>
</tr><tr>
<td>admin_secret</td>
<td>string</td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr><tr>
<td>vault_namespace</td>
<td>string</td>
<td><pre>
json_name: vaultNamespace
go_name: VaultNamespace</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.BootstrapAdminOutput](#fabric-identity-v1-bootstrapadminoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr><tr>
<td>message</td>
<td>string</td>
<td><pre>
json_name: message
go_name: Message</pre></td>
</tr><tr>
<td>success</td>
<td>bool</td>
<td><pre>
json_name: success
go_name: Success</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-provisionorg-workflow"></a>
### fabric.identity.v1.ProvisionOrg

<pre>
ProvisionOrg provisions a complete organization with admin identity, peers, and orderers
</pre>

**Input:** [fabric.identity.v1.ProvisionOrgInput](#fabric-identity-v1-provisionorginput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>additional_hosts</td>
<td>string</td>
<td><pre>
Additional SAN hosts for TLS certificates (optional)<br>

json_name: additionalHosts
go_name: AdditionalHosts</pre></td>
</tr><tr>
<td>admin</td>
<td><a href="#fabric-identity-v1-bootstrapadmininput">fabric.identity.v1.BootstrapAdminInput</a></td>
<td><pre>
Admin configuration<br>

json_name: admin
go_name: Admin</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
Affiliations to create (e.g., ["org.dept1", "org.dept2"])<br>

json_name: affiliations
go_name: Affiliations</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_ids</td>
<td>string</td>
<td><pre>
Orderer identities to provision<br>

json_name: ordererIds
go_name: OrdererIds</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_ids</td>
<td>string</td>
<td><pre>
Peer identities to provision<br>

json_name: peerIds
go_name: PeerIds</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.ProvisionOrgOutput](#fabric-identity-v1-provisionorgoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>ca_certificates</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: caCertificates
go_name: CaCertificates</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>orderers</td>
<td><a href="#fabric-identity-v1-provisionordereroutput">fabric.identity.v1.ProvisionOrdererOutput</a></td>
<td><pre>
json_name: orderers
go_name: Orderers</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peers</td>
<td><a href="#fabric-identity-v1-provisionpeeroutput">fabric.identity.v1.ProvisionPeerOutput</a></td>
<td><pre>
json_name: peers
go_name: Peers</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>    

<a name="fabric-identity-v1-fabricidentity-activities"></a>
### Activities

---
<a name="fabric-identity-v1-registeridentity-activity"></a>
### fabric.identity.v1.RegisterIdentity

<pre>
RegisterIdentity registers a new identity with Fabric CA
</pre>

**Input:** [fabric.identity.v1.RegisterIdentityInput](#fabric-identity-v1-registeridentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
json_name: affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>max_enrollments</td>
<td>int32</td>
<td><pre>
json_name: maxEnrollments
go_name: MaxEnrollments</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr><tr>
<td>type</td>
<td>string</td>
<td><pre>
json_name: type
go_name: Type</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.RegisterIdentityOutput](#fabric-identity-v1-registeridentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
json_name: affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr><tr>
<td>type</td>
<td>string</td>
<td><pre>
json_name: type
go_name: Type</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-enrollidentity-activity"></a>
### fabric.identity.v1.EnrollIdentity

<pre>
EnrollIdentity enrolls an identity to obtain certificates
</pre>

**Input:** [fabric.identity.v1.EnrollIdentityInput](#fabric-identity-v1-enrollidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>attr_reqs</td>
<td><a href="#fabric-identity-v1-attributerequest">fabric.identity.v1.AttributeRequest</a></td>
<td><pre>
json_name: attrReqs
go_name: AttrReqs</pre></td>
</tr><tr>
<td>csr</td>
<td><a href="#fabric-identity-v1-csrinfo">fabric.identity.v1.CSRInfo</a></td>
<td><pre>
json_name: csr
go_name: Csr</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
json_name: profile
go_name: Profile</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.EnrollIdentityOutput](#fabric-identity-v1-enrollidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_cert</td>
<td>bytes</td>
<td><pre>
json_name: caCert
go_name: CaCert</pre></td>
</tr><tr>
<td>certificate</td>
<td>bytes</td>
<td><pre>
json_name: certificate
go_name: Certificate</pre></td>
</tr><tr>
<td>private_key</td>
<td>bytes</td>
<td><pre>
json_name: privateKey
go_name: PrivateKey</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
json_name: profile
go_name: Profile</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-assemblemsp-activity"></a>
### fabric.identity.v1.AssembleMsp

<pre>
AssembleMsp assembles a complete MSP structure from enrollments
</pre>

**Input:** [fabric.identity.v1.AssembleMspInput](#fabric-identity-v1-assemblemspinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_certs</td>
<td>bytes</td>
<td><pre>
json_name: caCerts
go_name: CaCerts</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>signing_enrollment</td>
<td><a href="#fabric-identity-v1-enrollidentityoutput">fabric.identity.v1.EnrollIdentityOutput</a></td>
<td><pre>
json_name: signingEnrollment
go_name: SigningEnrollment</pre></td>
</tr><tr>
<td>tls_enrollment</td>
<td><a href="#fabric-identity-v1-enrollidentityoutput">fabric.identity.v1.EnrollIdentityOutput</a></td>
<td><pre>
json_name: tlsEnrollment
go_name: TlsEnrollment</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.AssembleMspOutput](#fabric-identity-v1-assemblemspoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_certs</td>
<td>bytes</td>
<td><pre>
json_name: caCerts
go_name: CaCerts</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>signing_cert</td>
<td>bytes</td>
<td><pre>
json_name: signingCert
go_name: SigningCert</pre></td>
</tr><tr>
<td>signing_key</td>
<td>bytes</td>
<td><pre>
json_name: signingKey
go_name: SigningKey</pre></td>
</tr><tr>
<td>tls_cert</td>
<td>bytes</td>
<td><pre>
json_name: tlsCert
go_name: TlsCert</pre></td>
</tr><tr>
<td>tls_key</td>
<td>bytes</td>
<td><pre>
json_name: tlsKey
go_name: TlsKey</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-generatesecuresecret-activity"></a>
### fabric.identity.v1.GenerateSecureSecret

<pre>
GenerateSecureSecret generates a secure random secret
</pre>

**Input:** [fabric.identity.v1.GenerateSecureSecretInput](#fabric-identity-v1-generatesecuresecretinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>length</td>
<td>int32</td>
<td><pre>
json_name: length
go_name: Length</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.GenerateSecureSecretOutput](#fabric-identity-v1-generatesecuresecretoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-createadminidentity-activity"></a>
### fabric.identity.v1.CreateAdminIdentity

<pre>
CreateAdminIdentity creates a new admin identity using the default admin
</pre>

**Input:** [fabric.identity.v1.CreateAdminIdentityInput](#fabric-identity-v1-createadminidentityinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_id</td>
<td>string</td>
<td><pre>
json_name: adminId
go_name: AdminId</pre></td>
</tr><tr>
<td>admin_secret</td>
<td>string</td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.CreateAdminIdentityOutput](#fabric-identity-v1-createadminidentityoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-revokedefaultadmin-activity"></a>
### fabric.identity.v1.RevokeDefaultAdmin

<pre>
RevokeDefaultAdmin revokes the default admin identity
</pre>

**Input:** [fabric.identity.v1.RevokeDefaultAdminInput](#fabric-identity-v1-revokedefaultadmininput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.RevokeDefaultAdminOutput](#fabric-identity-v1-revokedefaultadminoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>message</td>
<td>string</td>
<td><pre>
json_name: message
go_name: Message</pre></td>
</tr><tr>
<td>success</td>
<td>bool</td>
<td><pre>
json_name: success
go_name: Success</pre></td>
</tr>
</table> 

---
<a name="fabric-identity-v1-createaffiliations-activity"></a>
### fabric.identity.v1.CreateAffiliations

<pre>
CreateAffiliations creates affiliations in the Fabric CA
</pre>

**Input:** [fabric.identity.v1.CreateAffiliationsInput](#fabric-identity-v1-createaffiliationsinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
json_name: affiliations
go_name: Affiliations</pre></td>
</tr>
</table>

**Output:** [fabric.identity.v1.CreateAffiliationsOutput](#fabric-identity-v1-createaffiliationsoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>created_affiliations</td>
<td>string</td>
<td><pre>
json_name: createdAffiliations
go_name: CreatedAffiliations</pre></td>
</tr><tr>
<td>existing_affiliations</td>
<td>string</td>
<td><pre>
json_name: existingAffiliations
go_name: ExistingAffiliations</pre></td>
</tr>
</table>   

<a name="fabric-identity-v1-messages"></a>
## Messages

<a name="fabric-identity-v1-adminidentity"></a>
### fabric.identity.v1.AdminIdentity

<pre>
AdminIdentity represents an admin identity for Fabric CA
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>ca_cert</td>
<td>bytes</td>
<td><pre>
json_name: caCert
go_name: CaCert</pre></td>
</tr><tr>
<td>certificate</td>
<td>bytes</td>
<td><pre>
json_name: certificate
go_name: Certificate</pre></td>
</tr><tr>
<td>id</td>
<td>string</td>
<td><pre>
json_name: id
go_name: Id</pre></td>
</tr><tr>
<td>private_key</td>
<td>bytes</td>
<td><pre>
json_name: privateKey
go_name: PrivateKey</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-assemblemspinput"></a>
### fabric.identity.v1.AssembleMspInput

<pre>
AssembleMspInput contains parameters for MSP assembly
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_certs</td>
<td>bytes</td>
<td><pre>
json_name: caCerts
go_name: CaCerts</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>signing_enrollment</td>
<td><a href="#fabric-identity-v1-enrollidentityoutput">fabric.identity.v1.EnrollIdentityOutput</a></td>
<td><pre>
json_name: signingEnrollment
go_name: SigningEnrollment</pre></td>
</tr><tr>
<td>tls_enrollment</td>
<td><a href="#fabric-identity-v1-enrollidentityoutput">fabric.identity.v1.EnrollIdentityOutput</a></td>
<td><pre>
json_name: tlsEnrollment
go_name: TlsEnrollment</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-assemblemspoutput"></a>
### fabric.identity.v1.AssembleMspOutput

<pre>
AssembleMspOutput represents a complete MSP structure
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_certs</td>
<td>bytes</td>
<td><pre>
json_name: caCerts
go_name: CaCerts</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>signing_cert</td>
<td>bytes</td>
<td><pre>
json_name: signingCert
go_name: SigningCert</pre></td>
</tr><tr>
<td>signing_key</td>
<td>bytes</td>
<td><pre>
json_name: signingKey
go_name: SigningKey</pre></td>
</tr><tr>
<td>tls_cert</td>
<td>bytes</td>
<td><pre>
json_name: tlsCert
go_name: TlsCert</pre></td>
</tr><tr>
<td>tls_key</td>
<td>bytes</td>
<td><pre>
json_name: tlsKey
go_name: TlsKey</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-attribute"></a>
### fabric.identity.v1.Attribute

<pre>
Attribute represents an identity attribute (aligned with Fabric CA api.Attribute)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ecert</td>
<td>bool</td>
<td><pre>
json_name: ecert
go_name: Ecert</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>value</td>
<td>string</td>
<td><pre>
json_name: value
go_name: Value</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-attributerequest"></a>
### fabric.identity.v1.AttributeRequest

<pre>
AttributeRequest requests an attribute to be added to certificate (aligned with Fabric CA api.AttributeRequest)
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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>optional</td>
<td>bool</td>
<td><pre>
json_name: optional
go_name: Optional</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-bootstrapadmininput"></a>
### fabric.identity.v1.BootstrapAdminInput

<pre>
BootstrapAdminInput contains parameters for bootstrap admin workflow
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_id</td>
<td>string</td>
<td><pre>
json_name: adminId
go_name: AdminId</pre></td>
</tr><tr>
<td>admin_secret</td>
<td>string</td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr><tr>
<td>vault_namespace</td>
<td>string</td>
<td><pre>
json_name: vaultNamespace
go_name: VaultNamespace</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-bootstrapadminoutput"></a>
### fabric.identity.v1.BootstrapAdminOutput

<pre>
BootstrapAdminOutput contains the result of bootstrap admin workflow
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr><tr>
<td>message</td>
<td>string</td>
<td><pre>
json_name: message
go_name: Message</pre></td>
</tr><tr>
<td>success</td>
<td>bool</td>
<td><pre>
json_name: success
go_name: Success</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-csrinfo"></a>
### fabric.identity.v1.CSRInfo

<pre>
CSRInfo represents Certificate Signing Request information (aligned with Fabric CA api.CSRInfo)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>cn</td>
<td>string</td>
<td><pre>
json_name: cn
go_name: Cn</pre></td>
</tr><tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>key</td>
<td><a href="#fabric-identity-v1-keyrequest">fabric.identity.v1.KeyRequest</a></td>
<td><pre>
json_name: key
go_name: Key</pre></td>
</tr><tr>
<td>names</td>
<td><a href="#fabric-identity-v1-csrname">fabric.identity.v1.CSRName</a></td>
<td><pre>
json_name: names
go_name: Names</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-csrname"></a>
### fabric.identity.v1.CSRName

<pre>
CSRName represents subject name fields (aligned with cfssl csr.Name)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>c</td>
<td>string</td>
<td><pre>
json_name: c
go_name: C</pre></td>
</tr><tr>
<td>l</td>
<td>string</td>
<td><pre>
json_name: l
go_name: L</pre></td>
</tr><tr>
<td>o</td>
<td>string</td>
<td><pre>
json_name: o
go_name: O</pre></td>
</tr><tr>
<td>ou</td>
<td>string</td>
<td><pre>
json_name: ou
go_name: Ou</pre></td>
</tr><tr>
<td>st</td>
<td>string</td>
<td><pre>
json_name: st
go_name: St</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-createadminidentityinput"></a>
### fabric.identity.v1.CreateAdminIdentityInput

<pre>
CreateAdminIdentityInput contains parameters for creating admin identity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_id</td>
<td>string</td>
<td><pre>
json_name: adminId
go_name: AdminId</pre></td>
</tr><tr>
<td>admin_secret</td>
<td>string</td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-createadminidentityoutput"></a>
### fabric.identity.v1.CreateAdminIdentityOutput

<pre>
CreateAdminIdentityOutput contains the created admin identity
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-createaffiliationsinput"></a>
### fabric.identity.v1.CreateAffiliationsInput

<pre>
CreateAffiliationsInput contains parameters for creating affiliations
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_identity</td>
<td><a href="#fabric-identity-v1-adminidentity">fabric.identity.v1.AdminIdentity</a></td>
<td><pre>
json_name: adminIdentity
go_name: AdminIdentity</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
json_name: affiliations
go_name: Affiliations</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-createaffiliationsoutput"></a>
### fabric.identity.v1.CreateAffiliationsOutput

<pre>
CreateAffiliationsOutput contains result of affiliation creation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>created_affiliations</td>
<td>string</td>
<td><pre>
json_name: createdAffiliations
go_name: CreatedAffiliations</pre></td>
</tr><tr>
<td>existing_affiliations</td>
<td>string</td>
<td><pre>
json_name: existingAffiliations
go_name: ExistingAffiliations</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-enrollidentityinput"></a>
### fabric.identity.v1.EnrollIdentityInput

<pre>
EnrollIdentityInput represents input for enrolling an identity (aligned with Fabric CA api.EnrollmentRequest)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>attr_reqs</td>
<td><a href="#fabric-identity-v1-attributerequest">fabric.identity.v1.AttributeRequest</a></td>
<td><pre>
json_name: attrReqs
go_name: AttrReqs</pre></td>
</tr><tr>
<td>csr</td>
<td><a href="#fabric-identity-v1-csrinfo">fabric.identity.v1.CSRInfo</a></td>
<td><pre>
json_name: csr
go_name: Csr</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
json_name: profile
go_name: Profile</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-enrollidentityoutput"></a>
### fabric.identity.v1.EnrollIdentityOutput

<pre>
EnrollIdentityOutput represents the result of enrollment
Note: private_key is generated locally (via GenCSR), not returned by CA api.EnrollmentResponse
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>ca_cert</td>
<td>bytes</td>
<td><pre>
json_name: caCert
go_name: CaCert</pre></td>
</tr><tr>
<td>certificate</td>
<td>bytes</td>
<td><pre>
json_name: certificate
go_name: Certificate</pre></td>
</tr><tr>
<td>private_key</td>
<td>bytes</td>
<td><pre>
json_name: privateKey
go_name: PrivateKey</pre></td>
</tr><tr>
<td>profile</td>
<td>string</td>
<td><pre>
json_name: profile
go_name: Profile</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-generatesecuresecretinput"></a>
### fabric.identity.v1.GenerateSecureSecretInput

<pre>
GenerateSecureSecretInput contains parameters for secret generation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>length</td>
<td>int32</td>
<td><pre>
json_name: length
go_name: Length</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-generatesecuresecretoutput"></a>
### fabric.identity.v1.GenerateSecureSecretOutput

<pre>
GenerateSecureSecretOutput contains the generated secret
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-keyrequest"></a>
### fabric.identity.v1.KeyRequest

<pre>
KeyRequest specifies key algorithm and size (aligned with Fabric CA api.KeyRequest)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>algo</td>
<td>string</td>
<td><pre>
json_name: algo
go_name: Algo</pre></td>
</tr><tr>
<td>size</td>
<td>int32</td>
<td><pre>
json_name: size
go_name: Size</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-nodeouconfig"></a>
### fabric.identity.v1.NodeOUConfig

<pre>
NodeOUConfig represents Node Organizational Unit configuration (Fabric MSP config)
Enables role-based certificate classification
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin</td>
<td><a href="#fabric-identity-v1-ouidentifier">fabric.identity.v1.OUIdentifier</a></td>
<td><pre>
json_name: admin
go_name: Admin</pre></td>
</tr><tr>
<td>client</td>
<td><a href="#fabric-identity-v1-ouidentifier">fabric.identity.v1.OUIdentifier</a></td>
<td><pre>
json_name: client
go_name: Client</pre></td>
</tr><tr>
<td>enable</td>
<td>bool</td>
<td><pre>
json_name: enable
go_name: Enable</pre></td>
</tr><tr>
<td>orderer</td>
<td><a href="#fabric-identity-v1-ouidentifier">fabric.identity.v1.OUIdentifier</a></td>
<td><pre>
json_name: orderer
go_name: Orderer</pre></td>
</tr><tr>
<td>peer</td>
<td><a href="#fabric-identity-v1-ouidentifier">fabric.identity.v1.OUIdentifier</a></td>
<td><pre>
json_name: peer
go_name: Peer</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-ouidentifier"></a>
### fabric.identity.v1.OUIdentifier

<pre>
OUIdentifier identifies an organizational unit in the certificate
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>certificate</td>
<td>string</td>
<td><pre>
json_name: certificate
go_name: Certificate</pre></td>
</tr><tr>
<td>organizational_unit_identifier</td>
<td>string</td>
<td><pre>
json_name: organizationalUnitIdentifier
go_name: OrganizationalUnitIdentifier</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionclientinput"></a>
### fabric.identity.v1.ProvisionClientInput

<pre>
ProvisionClientInput contains parameters for client provisioning workflow
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>client_id</td>
<td>string</td>
<td><pre>
json_name: clientId
go_name: ClientId</pre></td>
</tr><tr>
<td>custom_attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: customAttributes
go_name: CustomAttributes</pre></td>
</tr><tr>
<td>enable_tls</td>
<td>bool</td>
<td><pre>
json_name: enableTls
go_name: EnableTls</pre></td>
</tr><tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionclientoutput"></a>
### fabric.identity.v1.ProvisionClientOutput

<pre>
ProvisionClientOutput contains references to stored secrets for provisioned client
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionordererinput"></a>
### fabric.identity.v1.ProvisionOrdererInput

<pre>
ProvisionOrdererInput contains parameters for orderer provisioning workflow
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
json_name: ordererId
go_name: OrdererId</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionordereroutput"></a>
### fabric.identity.v1.ProvisionOrdererOutput

<pre>
ProvisionOrdererOutput contains references to stored secrets for provisioned orderer
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionorginput"></a>
### fabric.identity.v1.ProvisionOrgInput

<pre>
ProvisionOrgInput contains parameters for organization provisioning workflow
Note: CA endpoint removed - client configured at init time
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>additional_hosts</td>
<td>string</td>
<td><pre>
Additional SAN hosts for TLS certificates (optional)<br>

json_name: additionalHosts
go_name: AdditionalHosts</pre></td>
</tr><tr>
<td>admin</td>
<td><a href="#fabric-identity-v1-bootstrapadmininput">fabric.identity.v1.BootstrapAdminInput</a></td>
<td><pre>
Admin configuration<br>

json_name: admin
go_name: Admin</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
Affiliations to create (e.g., ["org.dept1", "org.dept2"])<br>

json_name: affiliations
go_name: Affiliations</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_ids</td>
<td>string</td>
<td><pre>
Orderer identities to provision<br>

json_name: ordererIds
go_name: OrdererIds</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_ids</td>
<td>string</td>
<td><pre>
Peer identities to provision<br>

json_name: peerIds
go_name: PeerIds</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionorgoutput"></a>
### fabric.identity.v1.ProvisionOrgOutput

<pre>
ProvisionOrgOutput contains references to all provisioned identities and org-wide config
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>ca_certificates</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: caCertificates
go_name: CaCertificates</pre></td>
</tr><tr>
<td>node_ou_config</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: nodeOuConfig
go_name: NodeOuConfig</pre></td>
</tr><tr>
<td>orderers</td>
<td><a href="#fabric-identity-v1-provisionordereroutput">fabric.identity.v1.ProvisionOrdererOutput</a></td>
<td><pre>
json_name: orderers
go_name: Orderers</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peers</td>
<td><a href="#fabric-identity-v1-provisionpeeroutput">fabric.identity.v1.ProvisionPeerOutput</a></td>
<td><pre>
json_name: peers
go_name: Peers</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionpeerinput"></a>
### fabric.identity.v1.ProvisionPeerInput

<pre>
ProvisionPeerInput contains parameters for peer provisioning workflow
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-provisionpeeroutput"></a>
### fabric.identity.v1.ProvisionPeerOutput

<pre>
ProvisionPeerOutput contains references to stored secrets for provisioned peer
Note: To assemble complete MSP, fetch identity secrets + org-wide secrets
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>identity_name</td>
<td>string</td>
<td><pre>
json_name: identityName
go_name: IdentityName</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-registeridentityinput"></a>
### fabric.identity.v1.RegisterIdentityInput

<pre>
RegisterIdentityInput represents input for registering a new identity (aligned with Fabric CA api.RegistrationRequest)
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
json_name: affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>max_enrollments</td>
<td>int32</td>
<td><pre>
json_name: maxEnrollments
go_name: MaxEnrollments</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr><tr>
<td>type</td>
<td>string</td>
<td><pre>
json_name: type
go_name: Type</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-registeridentityoutput"></a>
### fabric.identity.v1.RegisterIdentityOutput

<pre>
RegisterIdentityOutput represents the result of identity registration
Note: CA api.RegistrationResponse only returns Secret, but we preserve request params for workflow state
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>affiliation</td>
<td>string</td>
<td><pre>
json_name: affiliation
go_name: Affiliation</pre></td>
</tr><tr>
<td>attributes</td>
<td><a href="#fabric-identity-v1-attribute">fabric.identity.v1.Attribute</a></td>
<td><pre>
json_name: attributes
go_name: Attributes</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>secret</td>
<td>string</td>
<td><pre>
json_name: secret
go_name: Secret</pre></td>
</tr><tr>
<td>type</td>
<td>string</td>
<td><pre>
json_name: type
go_name: Type</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-revokedefaultadmininput"></a>
### fabric.identity.v1.RevokeDefaultAdminInput

<pre>
RevokeDefaultAdminInput contains parameters for revoking default admin
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>default_admin_id</td>
<td>string</td>
<td><pre>
json_name: defaultAdminId
go_name: DefaultAdminId</pre></td>
</tr><tr>
<td>default_secret</td>
<td>string</td>
<td><pre>
json_name: defaultSecret
go_name: DefaultSecret</pre></td>
</tr>
</table>



<a name="fabric-identity-v1-revokedefaultadminoutput"></a>
### fabric.identity.v1.RevokeDefaultAdminOutput

<pre>
RevokeDefaultAdminOutput contains result of revocation
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>message</td>
<td>string</td>
<td><pre>
json_name: message
go_name: Message</pre></td>
</tr><tr>
<td>success</td>
<td>bool</td>
<td><pre>
json_name: success
go_name: Success</pre></td>
</tr>
</table>



<a name="fabric-v1"></a>
# fabric.v1

<a name="fabric-v1-services"></a>
## Services

<a name="fabric-v1-fabricdeploy"></a>
## fabric.v1.FabricDeploy

<pre>
FabricDeploy service defines workflows for Kubernetes infrastructure provisioning
These workflows orchestrate the deployment of Fabric components (CA, peers, orderers, CouchDB)
to a Kubernetes cluster with CSI-based secret mounting from Vault.
</pre>

<a name="fabric-v1-fabricdeploy-workflows"></a>
### Workflows

---
<a name="fabric-v1-createnamespace-workflow"></a>
### fabric.v1.CreateNamespace

<pre>
CreateNamespace creates a new Kubernetes namespace for an organization
</pre>

**Input:** [fabric.v1.CreateNamespaceInput](#fabric-v1-createnamespaceinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateNamespaceOutput](#fabric-v1-createnamespaceoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createserviceaccount-workflow"></a>
### fabric.v1.CreateServiceAccount

<pre>
CreateServiceAccount creates a Kubernetes service account for Vault authentication
</pre>

**Input:** [fabric.v1.CreateServiceAccountInput](#fabric-v1-createserviceaccountinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateServiceAccountOutput](#fabric-v1-createserviceaccountoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-createvaultrole-workflow"></a>
### fabric.v1.CreateVaultRole

<pre>
CreateVaultRole creates a Vault Kubernetes auth role for the service account
</pre>

**Input:** [fabric.v1.CreateVaultRoleInput](#fabric-v1-createvaultroleinput)

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
json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>

**Output:** [fabric.v1.CreateVaultRoleOutput](#fabric-v1-createvaultroleoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>role_name</td>
<td>string</td>
<td><pre>
json_name: roleName
go_name: RoleName</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deployca-workflow"></a>
### fabric.v1.DeployCA

<pre>
DeployCA deploys the Fabric CA pod to the Kubernetes cluster
</pre>

**Input:** [fabric.v1.DeployCaInput](#fabric-v1-deploycainput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployCaOutput](#fabric-v1-deploycaoutput)

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploycouchdb-workflow"></a>
### fabric.v1.DeployCouchDB

<pre>
DeployCouchDB deploys a CouchDB instance for peer state storage
Each peer gets its own CouchDB instance
</pre>

**Input:** [fabric.v1.DeployCouchDbInput](#fabric-v1-deploycouchdbinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployCouchDbOutput](#fabric-v1-deploycouchdboutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: credentials
go_name: Credentials</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deploypeer-workflow"></a>
### fabric.v1.DeployPeer

<pre>
DeployPeer deploys a peer node with certificates and CSI volume mounts
Calls ProvisionPeerWorkflow to generate certs, then deploys the peer pod
</pre>

**Input:** [fabric.v1.DeployPeerInput](#fabric-v1-deploypeerinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>couchdb_service</td>
<td>string</td>
<td><pre>
json_name: couchdbService
go_name: CouchdbService</pre></td>
</tr><tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployPeerOutput](#fabric-v1-deploypeeroutput)

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deployorderer-workflow"></a>
### fabric.v1.DeployOrderer

<pre>
DeployOrderer deploys an orderer node with certificates and CSI volume mounts
Calls ProvisionOrdererWorkflow to generate certs, then deploys the orderer pod
</pre>

**Input:** [fabric.v1.DeployOrdererInput](#fabric-v1-deployordererinput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
json_name: ordererId
go_name: OrdererId</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployOrdererOutput](#fabric-v1-deployordereroutput)

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-v1-deployorg-workflow"></a>
### fabric.v1.DeployOrg

<pre>
DeployOrg is the master workflow that orchestrates complete organization deployment
Execution flow:
1. CreateNamespace
2. CreateServiceAccount
3. CreateVaultRole
4. DeployCA
5. CreateAffiliations (activity - for CA identity organization)
6. BootstrapAdmin (workflow - create initial admin identity)
7. For each peer_id (sequential):
   - DeployCouchDB (for this peer)
   - DeployPeer (this peer)
8. For each orderer_id (sequential):
   - DeployOrderer (this orderer)
</pre>

**Input:** [fabric.v1.DeployOrgInput](#fabric-v1-deployorginput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>additional_hosts</td>
<td>string</td>
<td><pre>
Additional SAN hostnames for TLS certificates (optional)
Useful for external DNS names or load balancer endpoints<br>

json_name: additionalHosts
go_name: AdditionalHosts</pre></td>
</tr><tr>
<td>admin</td>
<td><a href="#fabric-identity-v1-bootstrapadmininput">fabric.identity.v1.BootstrapAdminInput</a></td>
<td><pre>
Admin identity configuration (to be created via BootstrapAdmin)<br>

json_name: admin
go_name: Admin</pre></td>
</tr><tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
CA configuration (bootstrap credentials for Fabric CA)<br>

json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
Affiliations to create in the Fabric CA
Examples: ["metacensus", "metacensus.engineering", "metacensus.operations"]<br>

json_name: affiliations
go_name: Affiliations</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_ids</td>
<td>string</td>
<td><pre>
Orderer identities to provision and deploy
Examples: ["orderer0", "orderer1", "orderer2"]<br>

json_name: ordererIds
go_name: OrdererIds</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
Node OU configuration (stored org-wide in Vault)<br>

json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_ids</td>
<td>string</td>
<td><pre>
Peer identities to provision and deploy
Examples: ["peer0", "peer1"]<br>

json_name: peerIds
go_name: PeerIds</pre></td>
</tr>
</table>

**Output:** [fabric.v1.DeployOrgOutput](#fabric-v1-deployorgoutput)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>ca_certificates</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: caCertificates
go_name: CaCertificates</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderers</td>
<td><a href="#fabric-v1-deployordereroutput">fabric.v1.DeployOrdererOutput</a></td>
<td><pre>
json_name: orderers
go_name: Orderers</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peers</td>
<td><a href="#fabric-v1-deploypeeroutput">fabric.v1.DeployPeerOutput</a></td>
<td><pre>
json_name: peers
go_name: Peers</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>     

<a name="fabric-v1-messages"></a>
## Messages

<a name="fabric-v1-createnamespaceinput"></a>
### fabric.v1.CreateNamespaceInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>



<a name="fabric-v1-createnamespaceoutput"></a>
### fabric.v1.CreateNamespaceOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr>
</table>



<a name="fabric-v1-createserviceaccountinput"></a>
### fabric.v1.CreateServiceAccountInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>



<a name="fabric-v1-createserviceaccountoutput"></a>
### fabric.v1.CreateServiceAccountOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>



<a name="fabric-v1-createvaultroleinput"></a>
### fabric.v1.CreateVaultRoleInput

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
json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
json_name: serviceAccountName
go_name: ServiceAccountName</pre></td>
</tr>
</table>



<a name="fabric-v1-createvaultroleoutput"></a>
### fabric.v1.CreateVaultRoleOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>role_name</td>
<td>string</td>
<td><pre>
json_name: roleName
go_name: RoleName</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycainput"></a>
### fabric.v1.DeployCaInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycaoutput"></a>
### fabric.v1.DeployCaOutput

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycouchdbinput"></a>
### fabric.v1.DeployCouchDbInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>



<a name="fabric-v1-deploycouchdboutput"></a>
### fabric.v1.DeployCouchDbOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>credentials</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: credentials
go_name: Credentials</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr>
</table>



<a name="fabric-v1-deployordererinput"></a>
### fabric.v1.DeployOrdererInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_id</td>
<td>string</td>
<td><pre>
json_name: ordererId
go_name: OrdererId</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr>
</table>



<a name="fabric-v1-deployordereroutput"></a>
### fabric.v1.DeployOrdererOutput

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>



<a name="fabric-v1-deployorginput"></a>
### fabric.v1.DeployOrgInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>additional_hosts</td>
<td>string</td>
<td><pre>
Additional SAN hostnames for TLS certificates (optional)
Useful for external DNS names or load balancer endpoints<br>

json_name: additionalHosts
go_name: AdditionalHosts</pre></td>
</tr><tr>
<td>admin</td>
<td><a href="#fabric-identity-v1-bootstrapadmininput">fabric.identity.v1.BootstrapAdminInput</a></td>
<td><pre>
Admin identity configuration (to be created via BootstrapAdmin)<br>

json_name: admin
go_name: Admin</pre></td>
</tr><tr>
<td>admin_password</td>
<td>string</td>
<td><pre>
json_name: adminPassword
go_name: AdminPassword</pre></td>
</tr><tr>
<td>admin_username</td>
<td>string</td>
<td><pre>
CA configuration (bootstrap credentials for Fabric CA)<br>

json_name: adminUsername
go_name: AdminUsername</pre></td>
</tr><tr>
<td>affiliations</td>
<td>string</td>
<td><pre>
Affiliations to create in the Fabric CA
Examples: ["metacensus", "metacensus.engineering", "metacensus.operations"]<br>

json_name: affiliations
go_name: Affiliations</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderer_ids</td>
<td>string</td>
<td><pre>
Orderer identities to provision and deploy
Examples: ["orderer0", "orderer1", "orderer2"]<br>

json_name: ordererIds
go_name: OrdererIds</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
Node OU configuration (stored org-wide in Vault)<br>

json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_ids</td>
<td>string</td>
<td><pre>
Peer identities to provision and deploy
Examples: ["peer0", "peer1"]<br>

json_name: peerIds
go_name: PeerIds</pre></td>
</tr>
</table>



<a name="fabric-v1-deployorgoutput"></a>
### fabric.v1.DeployOrgOutput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>admin_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: adminSecret
go_name: AdminSecret</pre></td>
</tr><tr>
<td>ca_certificates</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: caCertificates
go_name: CaCertificates</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>orderers</td>
<td><a href="#fabric-v1-deployordereroutput">fabric.v1.DeployOrdererOutput</a></td>
<td><pre>
json_name: orderers
go_name: Orderers</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peers</td>
<td><a href="#fabric-v1-deploypeeroutput">fabric.v1.DeployPeerOutput</a></td>
<td><pre>
json_name: peers
go_name: Peers</pre></td>
</tr>
</table>



<a name="fabric-v1-deploypeerinput"></a>
### fabric.v1.DeployPeerInput

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>couchdb_service</td>
<td>string</td>
<td><pre>
json_name: couchdbService
go_name: CouchdbService</pre></td>
</tr><tr>
<td>hosts</td>
<td>string</td>
<td><pre>
json_name: hosts
go_name: Hosts</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>org_node_ous</td>
<td><a href="#fabric-identity-v1-nodeouconfig">fabric.identity.v1.NodeOUConfig</a></td>
<td><pre>
json_name: orgNodeOus
go_name: OrgNodeOus</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>peer_id</td>
<td>string</td>
<td><pre>
json_name: peerId
go_name: PeerId</pre></td>
</tr>
</table>



<a name="fabric-v1-deploypeeroutput"></a>
### fabric.v1.DeployPeerOutput

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
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>service</td>
<td>string</td>
<td><pre>
json_name: service
go_name: Service</pre></td>
</tr><tr>
<td>signing_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: signingSecret
go_name: SigningSecret</pre></td>
</tr><tr>
<td>tls_secret</td>
<td><a href="#vault-v1-secretref">vault.v1.SecretRef</a></td>
<td><pre>
json_name: tlsSecret
go_name: TlsSecret</pre></td>
</tr>
</table>



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
<td>additional_policy_names</td>
<td>string</td>
<td><pre>
additional_policy_names is an optional list of additional Vault policy names to bind
The policy created by this workflow (policy_name) will automatically be included<br>

json_name: additionalPolicyNames
go_name: AdditionalPolicyNames</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
Auth role configuration
organization is the organization name (e.g., "metacensus", "acme-corp")<br>

json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
document is the structured policy content<br>

json_name: policyDocument
go_name: PolicyDocument</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
Policy configuration
name is the name of the policy to create (e.g., "fabric-foundational")<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate
(e.g., "fabric-org-sa")<br>

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
<td>organization</td>
<td>string</td>
<td><pre>
organization is the organization name (e.g., "metacensus", "acme-corp")<br>

json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
policy_names is a list of Vault policy names to bind to this role
(e.g., ["fabric-foundational", "fabric-org-specific"])<br>

json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate
(e.g., "fabric-org-sa")<br>

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
name is the name of the policy (e.g., "fabric-foundational")<br>

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
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
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
json_name: path
go_name: Path</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
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
<td>additional_policy_names</td>
<td>string</td>
<td><pre>
additional_policy_names is an optional list of additional Vault policy names to bind
The policy created by this workflow (policy_name) will automatically be included<br>

json_name: additionalPolicyNames
go_name: AdditionalPolicyNames</pre></td>
</tr><tr>
<td>kubernetes_namespace</td>
<td>string</td>
<td><pre>
kubernetes_namespace is the Kubernetes namespace where the service account exists<br>

json_name: kubernetesNamespace
go_name: KubernetesNamespace</pre></td>
</tr><tr>
<td>organization</td>
<td>string</td>
<td><pre>
Auth role configuration
organization is the organization name (e.g., "metacensus", "acme-corp")<br>

json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_document</td>
<td><a href="#vault-v1-policydocument">vault.v1.PolicyDocument</a></td>
<td><pre>
document is the structured policy content<br>

json_name: policyDocument
go_name: PolicyDocument</pre></td>
</tr><tr>
<td>policy_name</td>
<td>string</td>
<td><pre>
Policy configuration
name is the name of the policy to create (e.g., "fabric-foundational")<br>

json_name: policyName
go_name: PolicyName</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate
(e.g., "fabric-org-sa")<br>

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
<td>organization</td>
<td>string</td>
<td><pre>
organization is the organization name (e.g., "metacensus", "acme-corp")<br>

json_name: organization
go_name: Organization</pre></td>
</tr><tr>
<td>policy_names</td>
<td>string</td>
<td><pre>
policy_names is a list of Vault policy names to bind to this role
(e.g., ["fabric-foundational", "fabric-org-specific"])<br>

json_name: policyNames
go_name: PolicyNames</pre></td>
</tr><tr>
<td>service_account_name</td>
<td>string</td>
<td><pre>
service_account_name is the Kubernetes service account that can authenticate
(e.g., "fabric-org-sa")<br>

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
name is the name of the policy (e.g., "fabric-foundational")<br>

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
json_name: path
go_name: Path</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
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
path is the Vault path pattern (e.g., "secret/data/fabric/*")<br>

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
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>path</td>
<td>string</td>
<td><pre>
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

