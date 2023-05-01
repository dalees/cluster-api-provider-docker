# Usage:
# source env.sh
# clusterctl generate cluster kubecontest --from cluster-template.yaml > output.yaml

export KUBERNETES_VERSION=v1.23.0
#export KUBERNETES_VERSION=v1.26.3
# Fails with; Failed to create sandbox for pod \\\"kube-apiserver-kubecontest-control-plane-gf5f4_kube-system(7690535ab745cd438a6b753796ac489f)\\\": rpc error: code = Unknown desc = failed to create containerd task: failed to create shim task: OCI runtime create failed: runc create failed: expected cgroupsPath to be of format \\\"slice:prefix:name\\\" for systemd cgroups, got \\\"/kubelet/kubepods/burstable/pod7690535ab745cd438a6b753796ac489f/e8f44413129ec5e5f296177ec10785b864e2fbaefe73d943e0621130784e9574\\\" instead: unknown
export CLUSTER_NAME=kubecontest
export CONTROL_PLANE_MACHINE_COUNT=3
export WORKER_MACHINE_COUNT=3
