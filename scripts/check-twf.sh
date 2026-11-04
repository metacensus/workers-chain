#!/usr/bin/env bash
# twf/ checks clean with the library's module in the run, and each worker in
# twf/deploy hosts exactly what proto and lib/workers/<worker>/module.go say it does.
#
# twf binds an import only to a package of a module whose directory is among the inputs, so the
# library's module root (its twf.mod) is passed beside twf/: the module cache, at the version
# go.mod pins.
set -euo pipefail

library=$(GOWORK=off go mod download -json github.com/metacensus/library-temporal-kube | jq -r .Dir)

# A warning is an error here: an import that names no package in the run turns
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
        definitions activity "$library"/twf/helm/v1/*.twf | sed 's/^/activity helm./'
        definitions activity "$library"/twf/kube/v1/*.twf | sed 's/^/activity kube./'
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
plain=$(awk '/pseudo-twf\(plain-function\)/ { on=1 } on && /^workflow / { sub(/\(.*/, "", $2); print "workflow certauth." $2; on=0 }' twf/fabric/certauth/v1/certauth.twf)

# A worker registers one type per bare name, and the design has no wire names, so these two are
# named apart in the design from the RPC they mirror.
renamed() {
  sed -e 's/^workflow temporal\.CreateNamespace$/workflow temporal.CreateTemporalNamespace/' \
    -e 's/^activity peer\.ChannelJoin$/activity peer.PeerChannelJoin/'
}

status=0
for worker in infra fabric shard; do
  hosted=$(awk -v w="worker $worker:" '$0 == w { on=1; next } /^[^[:space:]]/ { on=0 } on && NF && !/^[ \t]*#/ { sub(/^[ \t]*/, ""); print }' twf/deploy/v1/deploy.twf \
    | grep -vxF "$plain" | sort)
  diff -u --label "module.go ($worker)" --label "twf/deploy ($worker)" <(expected "$worker" | renamed | sort) <(echo "$hosted") || status=1
done
exit $status
