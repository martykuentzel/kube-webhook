# kube-secHook

A Webhook that connects to Secret Manager and mutates secrets.

## Setup

#### Prerequisites for GCP Project

- enable KMS API
- enable Secret Manager API
- create Service Account for secHook
- give Service Account Secret Manager Secret Accessor Role
- make sure your Kubernetes Cluster has Workload Identity enabled
- create IAM Binding for GCP Service Account and K8S Service Account
```
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:YOUR_PROJECT.svc.id.goog[YOUR_NAMESPACE/YOUR_K8S_SA]" \
  YOUR_GCP_SA@YOUR_PROJECT.iam.gserviceaccount.com
```
- **Attention:** In Case you use a **private GKE** Cluster. Edit the firewall for the GKE-Master rule and add port 8443 (next to 443 and 10250)

#### Activation in Kubernetes
To enable the Webhook to listen for changes in your Namespace, you have to label the namespace
```
kubectl label namespace YOUR_NAMESPACE kube-secHook=enabled
```

## How it works

Create a secret with prefix "secHook:" and the address of the secret in Secret Manager
`/project/PROJECT_ID/secrets/SECRETNAME/versions/VERSIONNUMBER`

Example:
```
echo -n 'secHook:projects/dummy-playground/secrets/tester/versions/latest' | base64
```
Use the base64 Output in the secret Yaml. By omitting the 'secHook:' prefix you can also combine your secret with regular secrets.

```
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  mutatedPassword: c2VjSG9vazpwcm9qZWN0cy9kdW1teS1wbGF5Z3JvdW5kL3NlY3JldHMvdGVzdGVyL3ZlcnNpb25zL2xhdGVzdA==
  nonMutatedPassword: cGFzc3dvcmQ= # regular password
```

>Note: In case the secret cannot be found in Secret Manager (because it does not exsists or permission issues) the affected key:value of that secret will not be mutated and remains unchanged. In the future a validation Hook is planned, that runs a validation