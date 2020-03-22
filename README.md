# k8s-mutate-webhook

A Webhook that connects to SecretManager and mutates secrets.

## Setup

### Prerequisites for GCP Project

- enable KMS API
- enable Secret Manager API
- create Service Account for SecMan
- give Service Account Secret Manager Secret Accessor Role
- make sure your Kubernetes Cluster has Workload Identity enabled
- create IAM Binding for GCP Service Account and K8S Service Account
```
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:YOUR_PROJECT.svc.id.goog[YOUR_NAMESPACE/YOUR_K8S_SA]" \
  YOUR_GCP_SA@YOUR_PROJECT.iam.gserviceaccount.com
```
- in Case you use a private GKE Cluster edit the firewall for the GKE-Master rule and add port 8443 (next to 443 and 10250)

### Running Kubernetes Cluster
```
k label namespace default mutateme=enabled --overwrite=true
```