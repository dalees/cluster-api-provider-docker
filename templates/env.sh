# Usage:
# source env.sh
# clusterctl generate cluster kubecontest --from cluster-template.yaml > output.yaml

#export KUBERNETES_VERSION=v1.23.0
export KUBERNETES_VERSION=v1.24.12
#export KUBERNETES_VERSION=v1.25.8
#export KUBERNETES_VERSION=v1.26.3
#export KUBERNETES_VERSION=v1.27.1
export CLUSTER_NAME=kubecontest
export CONTROL_PLANE_MACHINE_COUNT=3
export WORKER_MACHINE_COUNT=3
