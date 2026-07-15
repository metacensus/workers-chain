# RFC: CertAuth Issuance — Naming & Message Model

**Status:** ✅ Implemented (2026-07-12) — both commits landed (unstaged)
**Scope:** the `fabric.certauth.v1` issuance path — CN derivation (#4) and the
issuance message shapes (#5).
**Non-goals:** Nexus RBAC/auth (still deferred); peer/orderer startup; storage
path/naming (`storage_*_ref`) redesign; any change to the security boundary
(no key/registrar identity crosses Nexus — that invariant holds and stays).

This RFC covers two review items together because they share one domain — how an
issued certificate's *subject* is named and represented — but proposes landing
them as **two separate commits** with opposite risk profiles.

---

## 1. Background: two problems, one domain

### #4 — Cross-owner CN collision at the shared operator TLS CA

An org CA's own **TLS** cert is issued by the operator's shared TLS CA over Nexus
(intentional: operator owns the TLS chain, orgs own their org chains). The CN it
presents is:

```
caConfig.ComponentName()  ==  "certificate-authority-<certType>-<caID>"
```

(`workers/fabric/config/certificate_authority.go:201`), derived from
`certType`+`caID` **only — not the owner**. So org1's `org/int` CA and org2's
`org/int` CA both present CN `certificate-authority-org-int` to the *same*
operator TLS CA. At that shared server they are one registration identity, and
`SignCsr`'s register-or-`ModifyIdentity` (`activities.go:304-320`) mutates a
single shared record across two trust domains. Functional today
(`MaxEnrollments:-1`), but a cross-org identity clobber waiting to happen.

**Why the caller must fix the CN (not the server):** by design no owner/registrar
identity crosses Nexus, so the operator TLS CA *cannot* know which org is asking
except from the CSR itself. Namespacing therefore has to be in the CN the caller
presents. This is consistent with the boundary, not a hole in it.

`ComputeTLSCommonName` (`workflows_admin.go:19`) is a holdover from per-shard TLS
CAs: it only disambiguates *same-CA-issues-both-signing-and-TLS* (appends
`-tls`). It never contemplated two owners meeting at one shared CA.

### #5 — `CertificateSubject` conflates unrelated concerns

`CertificateSubject` (`shared.proto:11`) carries `common_name`, `organization`,
`organizational_unit`, `dns_names`. It's threaded through `GenerateCsrInput`,
`SignCertificateInput`, and `CreateIdentityInput`. But the consumers tell a
different story (evidence below): most of those fields are **write-only dead
weight**, and SANs are represented twice.

---

## 2. Current-state evidence (producer → consumer trace)

| Field | Set by producers? | Ever read into a CSR / registration? |
|---|---|---|
| `subject.common_name` | yes | **yes** — CSR CN (`activities.go:241`), registration `identity_name` |
| `subject.organization` | yes (`O = OwnerID`) | **no reader anywhere** |
| `subject.organizational_unit` | yes (`OU = "peer"/"orderer"/...`) | **no reader anywhere** — real OU is server-set from `identity_type` |
| `subject.dns_names` | yes | yes — but only as SANs, and duplicated by the separate `hosts` field |

Key facts confirmed in code:

- **`GenerateCsr`** builds `api.CSRInfo{CN: subject.CommonName}` + `Hosts` from
  `dns_names`. It never sets O/OU on the CSR (`activities.go:227-243`).
- **`SignCsrInput`** has already been flattened to `identity_name` + `hosts`
  (no `subject` at all) — this is the shape we like.
- **`SignCertificate`** consumer reads only `subject.CommonName` + `hosts`
  (`workflows_sign.go:79`); `organization`, `organizational_unit`, and
  `subject.dns_names` are dropped on the floor.
- **`IssueCertificate`/`signCsrLocal`** forward only `subject.CommonName` (→
  `identity_name`) and `subject.DnsNames` (→ `hosts`).
- Grep for `GetOrganization`/`GetOrganizationalUnit` readers: **none** outside
  producers.

**Conclusion:** the actual issuance contract is `common_name` + SANs +
(`identity_type`, `attributes`, `profile`). O/OU are noise; SANs are double-named;
and `identity_name` vs `common_name` is a must-be-equal coupling modeled twice.

---

## 3. Proposed direction

### 3a. Naming (#4)  — DECIDED (scope corrected 2026-07-12)

**The rule:** owner-qualify an identity's **enrollment CN where it enrolls at a
cross-owner (shared) CA**; leave it unqualified where the CA is single-tenant.
Qualified form appends the trust domain as the final `@`-segment:

```
<baseName>@<owner>            # certificate-authority-org-int@org1
                              # admin@certificate-authority-org-int@org1
                              # bootstrap-admin@certificate-authority-org-int@org1
```

**Scope correction (found in implementation):** the RFC first draft scoped this to
the CA's own TLS cert alone. Reading the code, `ComputeTLSCommonName` feeds **three**
foreign-TLS enrollment sites, and for an org CA *all three* are issued by the shared
operator TLS CA (`workflows_ca.go:211-217` sets `adminTLSIssuer` to the foreign
issuer), so all three collide across owners:

| Identity | site | CN today | after |
|---|---|---|---|
| CA TLS cert | `workflows_ca.go:350` | `certificate-authority-org-int` | `…@org1` |
| bootstrap-admin TLS cert | `workflows_admin.go:171` | `bootstrap-admin@certificate-authority-org-int` | `…@org1` |
| admin TLS cert | `workflows_admin.go:294` | `admin@certificate-authority-org-int` | `…@org1` |

**Guard-rail refinement:** "don't touch admin CNs" splits by cert purpose:
- **Signing CNs** (the MSP identity) stay **frozen** — untouched.
- **TLS CNs** are transport identity (not MSP membership; hostname check uses SANs),
  and they are *part of the collision*, so they get owner-qualified. Safe: verified
  nothing in the tree parses CNs on `@` (opaque enrollment strings), and CN ==
  registered enrollment ID is preserved (error-77 holds).

**Why CN and not the `O` field:** Fabric CA keys its identity registry by enrollment
ID (= CN here), not the full DN. The code already sets `Organization = owner` and the
collision persists regardless — proof `O` can't disambiguate. Qualifying the CN is
the idiomatic multi-tenant-shared-CA fix, working with Fabric CA's grain.

**Double-`@` is fine, not misleading:** `@` already means "within namespace" in this
codebase (`peer0@org1` = owner; `admin@ca` = CA scope). Nested — `admin@ca@org1` reads
"admin, in the org-int CA, in org1" — and the **final `@`-segment is always the trust
domain**, consistent with `peer0@org1`. A prefix scheme (`org1/…`) was rejected: it
would make owner a *suffix* for peers but a *prefix* for CAs — less uniform.

Local same-CA issuance keeps the historical `-tls` disambiguation. Replace
`ComputeTLSCommonName` with one helper encoding: *foreign → `base@owner`; local
same-CA → `base-tls`; else → `base`.*

### 3b. Message model (#5)  — DECIDED: retire `CertificateSubject`, go flat

Collapse `CertificateSubject` in favor of the flat shape `SignCsrInput` already
proves out. Everywhere issuance happens, the input becomes:

- `string common_name` (the one caller-controlled subject field; also the
  registration `identity_name` — model the equality once, in one field)
- `repeated string hosts` (single SANs source; retire `dns_names`)
- keep `identity_type`, `attributes`, `profile` as today

Then **retire `CertificateSubject`** and drop `organization` / `organizational_unit`
entirely (dead), `reserve`-ing their field numbers.

### 3c. Illustration — consequences of retiring vs slimming

**Before** (`GenerateCsrInput`, representative):

```proto
message GenerateCsrInput {
  secrets.v1.SecretRef key_ref = 1;
  CertificateSubject   subject = 2;   // {common_name, organization*, organizational_unit*, dns_names}
}                                      // * = never read
```

**After (option a — retire, RECOMMENDED):**

```proto
message GenerateCsrInput {
  secrets.v1.SecretRef key_ref = 1;
  string               common_name = 2;
  repeated string      hosts       = 3;
}
```

**After (option b — slim, keep the wrapper):**

```proto
message CertificateSubject {         // kept, but only 2 live fields
  string          common_name = 1;
  reserved 2, 3;                     // organization, organizational_unit
  repeated string hosts       = 4;   // renamed from dns_names
}
```

Consequences to weigh:

| | (a) retire | (b) slim |
|---|---|---|
| Kills dead O/OU | yes | yes (reserved) |
| Kills SANs double-naming (`dns_names` vs `hosts`) | yes — one `hosts` everywhere | partial — `CertificateSubject.hosts` still coexists with `SignCsrInput.hosts` unless you also flatten consumers |
| Forces resolving `identity_name` vs `subject.common_name` dup in `CreateIdentityInput` | **yes** — collapses to one `common_name` (removes a real divergence-bug surface: today `identity_name` is validated but the CN actually used is `subject.common_name`) | no — the two fields still coexist |
| Message shapes touched | 3 (`GenerateCsrInput`, `SignCertificateInput`, `CreateIdentityInput`) + delete `CertificateSubject` | 1 (`CertificateSubject`) |
| Producer call sites simplified | more (stop building a struct to set one field) | fewer |
| Blast radius / diff size | larger (but mechanical) | smaller |

**Replay/serialization:** these are Temporal **workflow-input** payloads persisted
in history, so a shape change could break replay of a workflow started under the
old shape. Per the project green-field stance (`/AGENTS.md` — no backwards-compat
owed yet), this is explicitly a non-concern: we take the clean shape and don't
build migration scaffolding. Noted only so the reasoning is on record.

The decisive factor for (a): it's the *only* option that eliminates the
`identity_name`/`common_name` divergence in `CreateIdentityInput` — a latent
"these two must be equal but nothing enforces it" bug. Slimming leaves that intact.

---

## 4. Decisions — RESOLVED (2026-07-11)

1. **CA-TLS CN format** → `<ComponentName()>@<owner>` (long form, e.g.
   `certificate-authority-org-int@org1`). No new short token; one name per CA. (§3a)
2. **Qualify scope** → minimal, by the rule "qualify at a cross-owner CA, leave
   single-tenant CAs unqualified." Confirmed a rule, not an inconsistency; full
   uniformity rejected (drags in MSP/admin CNs). (§3a)
3. **Message model depth** → (a) retire `CertificateSubject`, go flat. Consequences
   in §3c; decisive factor is collapsing the `identity_name`/`common_name`
   divergence in `CreateIdentityInput`. Replay/back-compat a non-concern per
   `/AGENTS.md`.
4. **Commit order** → naming fix first, message refactor second. (§5)

All decisions resolved — ready to implement Commit 1.

---

## 5. Implementation plan (two commits)

Neither item technically blocks the other; the naming fix needs **no** proto
change (CN flows through the existing `common_name`). Recommended order:

**Commit 1 — Naming fix (#4). ✅ DONE (2026-07-12).**
- Replaced `ComputeTLSCommonName` with `TLSEnrollmentCommonName(owner, base, signing, tls)`
  (`workflows_admin.go`): foreign → `base@owner`; local same-CA → `base-tls`; else `base`.
- Applied at all **three** foreign-TLS sites: CA (`workflows_ca.go:350`), bootstrap-admin
  (`workflows_admin.go:171`), admin (`workflows_admin.go:294`). Signing CNs untouched.
- Tests: `TestTLSEnrollmentCommonName` (8 cases incl. double-@ admin/bootstrap) +
  `TestTLSEnrollmentCommonName_DistinctOwners` (collision regression). Package green.

**Commit 2 — Message model (#5). ✅ DONE (2026-07-12). No behavior change.**
- Retired `CertificateSubject` (deleted `shared.proto` + stale `shared.pb.go`).
  Flattened `GenerateCsrInput`, `SignCertificateInput`, `CreateIdentityInput` to
  `common_name` (+ `hosts` where SANs apply). `CreateIdentityInput` lost its vestigial
  `identity_name` (only ever validated) — CN divergence surface gone.
- Dropped dead `organization`/`organizational_unit` (never consumed into a CSR) and
  the duplicate `subject.dns_names` (now single `hosts`).
- Fixed the two workflow-ID templates (`${! Subject.CommonName }` → `${! CommonName }`).
- `IssueCertificate`/`BootstrapIssueCertificate`/`ensureKeyAndCsr` now take
  `commonName string, hosts []string` instead of a `*CertificateSubject`.
- Regen (`make proto`) + producers/consumers in `temporal/lib/…` + 3 test files updated.
  No mocks referenced `CertificateSubject`. Full `temporal` module builds; all tests
  green (incl. `-race` on the two core packages).

Per `/AGENTS.md` green-field stance, field numbers were reused (no `reserved`-for-
migration scaffolding); replay/back-compat intentionally not preserved.

Keeping the correctness fix (Commit 1) out of this noisy refactor was the point: each
half is independently reviewable and verifiable.

---

## 6. Do NOT touch (guard rails)

- **Signing / admin / bootstrap-admin CNs** — MSP-adjacent identity. Only the
  CA→foreign-TLS CN moves.
- **Error-77 coupling** — a bootstrap-issued cert's CSR CN MUST equal the CA's
  bootstrap enrollment ID (`BootstrapAdminUsername`). Do not let naming changes
  touch that path.
- **Security boundary** — no key/registrar/owner identity crosses Nexus. The
  owner-qualified CN is caller-supplied public material; nothing new crosses.
- **Storage refs** (`signing_storage`/`tls_storage`) — out of scope here.
