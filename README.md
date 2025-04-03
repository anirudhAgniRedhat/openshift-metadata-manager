# OpenShift Metadata Manager

A CLI tool for managing cloud infrastructure resource tags in OpenShift clusters with real-time progress feedback.

![CLI Demo](https://media.giphy.com/media/v1.Y2lkPTc5MGI3NjExdXl5ZzB1c3M0bXl5OTg1dGJ4MWpvbTRlZ3Z2aHl2Z3J6cW5xM3B5aCZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/3oKIPEqDGUULpEU0aQ/giphy.gif)

## Features

- ☁️ **Multi-Cloud Support**: AWS, Azure, GCP.
- 🔍 **Automatic Platform Detection**: Identifies cloud provider from cluster
- 🧪 **Dry-Run Mode**: Test changes without modifications
- 🛠 **Kubernetes Integration**: Works with OpenShift cluster configuration
- 🚦 **Clean Output**: Preserves terminal state on exit

## Installation

### Prerequisites
- Go 1.20+
- OpenShift cluster access
- Cloud provider credentials (AWS/Azure/GCP/IBM/OpenStack)

### Quick Install
```bash
git clone https://github.com/anirudhAgniRedhat/openshift-metadata-manager.git
cd openshift-metadata-manager
go build -o openshift-metadata-manager
```

### Commands
Help Command
```bash 
./openshift-metadata-manager --help
```

Add path to yuor kubeconfig file in KUBECONFIG env var. 
```bash
export KUBECONFIG=<PATH to KUBECONFIG>
```

List Resource Command for your openshift cluster.
```bash
./openshift-metadata-manager list
```


Sync Command: update Tags on the cluster resources.
```bash 
./openshift-metadata-manager sync --platform aws --tags CostCenter=1234 
```




