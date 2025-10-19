# Table of Contents

- [fabric.identity.v1](#fabric-identity-v1)
  - Services
    - [fabric.identity.v1.FabricIdentity](#fabric-identity-v1-fabricidentity)
      - [Workflows](#fabric-identity-v1-fabricidentity-workflows)
        - [fabric.identity.v1.RegisterIdentity](#fabric-identity-v1-registeridentity-workflow)
        - [fabric.identity.v1.EnrollIdentity](#fabric-identity-v1-enrollidentity-workflow)
        - [fabric.identity.v1.AssembleMsp](#fabric-identity-v1-assemblemsp-workflow)
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
- [vault.v1](#vault-v1)
  - Services
    - [vault.v1.Vault](#vault-v1-vault)
      - [Workflows](#vault-v1-vault-workflows)
        - [vault.v1.StoreSecret](#vault-v1-storesecret-workflow)
        - [vault.v1.FetchSecret](#vault-v1-fetchsecret-workflow)
      - [Activities](#vault-v1-vault-activities)
        - [vault.v1.StoreSecret](#vault-v1-storesecret-activity)
        - [vault.v1.FetchSecret](#vault-v1-fetchsecret-activity)
  - Messages
    - [vault.v1.Secret](#vault-v1-secret)
    - [vault.v1.Secret.DataEntry](#vault-v1-secret-dataentry)
    - [vault.v1.Secret.LabelsEntry](#vault-v1-secret-labelsentry)
    - [vault.v1.SecretRef](#vault-v1-secretref)

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
<a name="fabric-identity-v1-registeridentity-workflow"></a>
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

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-enrollidentity-workflow"></a>
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

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="fabric-identity-v1-assemblemsp-workflow"></a>
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

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

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



<a name="vault-v1"></a>
# vault.v1

<a name="vault-v1-services"></a>
## Services

<a name="vault-v1-vault"></a>
## vault.v1.Vault

<pre>
Vault service defines workflows and activities related to secret management
</pre>

<a name="vault-v1-vault-workflows"></a>
### Workflows

---
<a name="vault-v1-storesecret-workflow"></a>
### vault.v1.StoreSecret

<pre>
StoreSecret stores (or updates) a secret in Vault
</pre>

**Input:** [vault.v1.Secret](#vault-v1-secret)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-secret-dataentry">vault.v1.Secret.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>labels</td>
<td><a href="#vault-v1-secret-labelsentry">vault.v1.Secret.LabelsEntry</a></td>
<td><pre>
json_name: labels
go_name: Labels</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Output:** [vault.v1.SecretRef](#vault-v1-secretref)

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
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Defaults:**

<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>id_reuse_policy</td><td><pre><code>WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE</code></pre></td></tr>
<tr><td>parent_close_policy</td><td><pre><code>PARENT_CLOSE_POLICY_REQUEST_CANCEL</code></pre></td></tr>
</table>

---
<a name="vault-v1-fetchsecret-workflow"></a>
### vault.v1.FetchSecret

<pre>
FetchSecret retrieves a secret from Vault
</pre>

**Input:** [vault.v1.SecretRef](#vault-v1-secretref)

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
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Output:** [vault.v1.Secret](#vault-v1-secret)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-secret-dataentry">vault.v1.Secret.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>labels</td>
<td><a href="#vault-v1-secret-labelsentry">vault.v1.Secret.LabelsEntry</a></td>
<td><pre>
json_name: labels
go_name: Labels</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
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
<a name="vault-v1-storesecret-activity"></a>
### vault.v1.StoreSecret

<pre>
StoreSecret stores (or updates) a secret in Vault
</pre>

**Input:** [vault.v1.Secret](#vault-v1-secret)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-secret-dataentry">vault.v1.Secret.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>labels</td>
<td><a href="#vault-v1-secret-labelsentry">vault.v1.Secret.LabelsEntry</a></td>
<td><pre>
json_name: labels
go_name: Labels</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Output:** [vault.v1.SecretRef](#vault-v1-secretref)

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
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table> 

---
<a name="vault-v1-fetchsecret-activity"></a>
### vault.v1.FetchSecret

<pre>
FetchSecret retrieves a secret from Vault
</pre>

**Input:** [vault.v1.SecretRef](#vault-v1-secretref)

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
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

**Output:** [vault.v1.Secret](#vault-v1-secret)

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-secret-dataentry">vault.v1.Secret.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>labels</td>
<td><a href="#vault-v1-secret-labelsentry">vault.v1.Secret.LabelsEntry</a></td>
<td><pre>
json_name: labels
go_name: Labels</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>   

<a name="vault-v1-messages"></a>
## Messages

<a name="vault-v1-secret"></a>
### vault.v1.Secret

<pre>
Secret describes a secret entity in vault
</pre>

<table>
<tr>
<th>Attribute</th>
<th>Type</th>
<th>Description</th>
</tr>
<tr>
<td>data</td>
<td><a href="#vault-v1-secret-dataentry">vault.v1.Secret.DataEntry</a></td>
<td><pre>
json_name: data
go_name: Data</pre></td>
</tr><tr>
<td>labels</td>
<td><a href="#vault-v1-secret-labelsentry">vault.v1.Secret.LabelsEntry</a></td>
<td><pre>
json_name: labels
go_name: Labels</pre></td>
</tr><tr>
<td>name</td>
<td>string</td>
<td><pre>
json_name: name
go_name: Name</pre></td>
</tr><tr>
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>



<a name="vault-v1-secret-dataentry"></a>
### vault.v1.Secret.DataEntry

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



<a name="vault-v1-secret-labelsentry"></a>
### vault.v1.Secret.LabelsEntry

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



<a name="vault-v1-secretref"></a>
### vault.v1.SecretRef

<pre>
SecretRef references an existing secret (or where to store one)
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
<td>namespace</td>
<td>string</td>
<td><pre>
json_name: namespace
go_name: Namespace</pre></td>
</tr><tr>
<td>version</td>
<td>int64</td>
<td><pre>
json_name: version
go_name: Version</pre></td>
</tr>
</table>

