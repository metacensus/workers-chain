#!/usr/bin/env bash
# twf/ checks clean with the library's packages in the tree, and each worker in
# twf/deploy hosts exactly what proto and lib/workers/<worker>/module.go say it does.
set -euo pipefail

library=$(GOWORK=off go mod download -json github.com/metacensus/library-temporal-kube | jq -r .Dir)/twf

# A warning is an error here: an import that names no package in the tree turns
# every reference through it into an unchecked external.
out=$(twf check twf "$library" 2>&1) || { echo "$out"; exit 1; }
if grep -q '^warning' <<<"$out"; then echo "$out"; exit 1; fi

rpcs() { # kind protofile...: the RPC names whose temporal option is that kind
  awk -v kind="temporal.v1.$1" '/^ *rpc /{ rpc=$2; sub(/\(.*/, "", rpc) } index($0, "(" kind ")"){ print rpc }' "${@:2}"
}
definitions() { # kind twffile...: the names defined in a .twf file
  sed -n "s/^$1 \([A-Za-z0-9]*\)(.*/\1/p" "${@:2}"
}

expected() { # worker: what module.go registers, as twf deploy entries
  grep -oE '\b[a-z]+\.Register[A-Za-z]+' "lib/workers/$1/module.go" | while IFS=. read -r pkg call; do
    case "$pkg.$call" in
      clusterfabric.RegisterWorkflows) rpcs workflow proto/cluster/fabric/v1/*.proto | sed 's/^/workflow cluster./' ;;
      certauth.RegisterWorkflows) rpcs workflow proto/fabric/certauth/v1/*.proto | sed 's/^/workflow certauth./' ;;
      certauth.RegisterActivities | channel.RegisterActivities | orderer.RegisterActivities | peer.RegisterActivities)
        rpcs activity proto/fabric/"$pkg"/v1/*.proto | sed "s/^/activity $pkg./" ;;
      fabric.RegisterWorkflows) rpcs workflow proto/fabric/v1/*.proto | sed 's/^/workflow fabric./' ;;
      workers.RegisterClusterActivities)
        definitions activity "$library"/helm/v1/*.twf | sed 's/^/activity helm./'
        definitions activity "$library"/kube/v1/*.twf | sed 's/^/activity kube./'
        echo activity temporalmtls.EnsureTemporalMTLSClientCert ;;
      secrets.RegisterNexusService) echo nexus service secrets.Secrets ;;
      certauth.RegisterSignCertificateNexusService) echo nexus service certauth.CertificateAuthority ;;
      fabric.RegisterBootstrapFabricNexusService) echo nexus service fabric.FabricBootstrap ;;
      fabric.RegisterBootstrapShardNexusService) echo nexus service fabric.ShardBootstrap ;;
      *.RegisterWorkflows) rpcs workflow proto/"$pkg"/v1/*.proto | sed "s/^/workflow $pkg./" ;;
      *.RegisterActivities) rpcs activity proto/"$pkg"/v1/*.proto | sed "s/^/activity $pkg./" ;;
      *) echo "check-twf: no mapping for $pkg.$call in lib/workers/$1/module.go" >&2; exit 1 ;;
    esac
  done
}

# A plain Go function is a workflow in the design only, so it has no RPC.
plain=$(grep -h -A2 'pseudo-twf(plain-function)' twf/fabric/certauth/v1/certauth.twf \
  | sed -n 's/^workflow \([A-Za-z]*\)(.*/workflow certauth.\1/p')

status=0
for worker in infra fabric shard; do
  hosted=$(awk -v w="worker $worker:" '$0 == w { on=1; next } /^[^ ]/ { on=0 } on && NF && !/^ *#/ { sub(/^ */, ""); print }' twf/deploy/v1/deploy.twf \
    | grep -vxF "$plain" | sort)
  diff -u --label "module.go ($worker)" --label "twf/deploy ($worker)" <(expected "$worker" | sort) <(echo "$hosted") || status=1
done
exit $status
