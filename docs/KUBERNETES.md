# Kubernetes Deployment

Cloud DR Orchestrator can run as a CronJob in Kubernetes for automated backups.

## Quick Start

### 1. Build Image

```bash
./build.sh
```

### 2. Deploy to K8s

See the [home_infra integration](https://github.com/Kobeep/home_infra/tree/main/k8s/backup) for complete K8s manifests.

Or deploy standalone:

```bash
# Create namespace
kubectl create namespace backup

# Create secret with Oracle Cloud credentials
kubectl create secret generic backup-secrets -n backup \
  --from-literal=oci-user-ocid="ocid1.user..." \
  --from-literal=oci-tenancy-ocid="ocid1.tenancy..." \
  --from-literal=oci-fingerprint="aa:bb:cc..." \
  --from-file=oci-private-key=~/.oci/oci_api_key.pem \
  --from-literal=encryption-key="your-32-char-key-here"

# Create ConfigMap
kubectl create configmap backup-config -n backup \
  --from-file=config.yaml=configs/config.yaml

# Deploy CronJob
kubectl apply -f k8s/cronjob.yaml
```

## Features in K8s

- ✅ Scheduled backups via CronJob
- ✅ Persistent storage for local cache
- ✅ Prometheus metrics endpoint
- ✅ RBAC for security
- ✅ ConfigMap-based configuration
- ✅ Secret management for credentials

## Manual Backup Trigger

```bash
kubectl create job --from=cronjob/backup-daily manual-backup-$(date +%s) -n backup
```

## Monitoring

```bash
# View logs
kubectl logs -n backup -l job-name=backup-daily

# Check job status
kubectl get jobs -n backup

# Get metrics
kubectl port-forward -n backup svc/backup-metrics 9090:9090
curl localhost:9090/metrics
```

## Integration with home_infra

For full home lab integration, see [home_infra repository](https://github.com/Kobeep/home_infra).

This includes:
- Automated deployment with Ansible
- Pre-configured manifests
- Integration with other services
- Prometheus monitoring setup
- Complete documentation
