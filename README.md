# Policy-Based Image Compliance with Kyverno

This is a demo to verify an image from a JSON payload using Kyverno 1.14's image verification policy type.

## Usage

Create a kind cluster:

```sh
kind create cluster --name=verify-images --image kindest/node:v1.32.0
```

Install `nirmata-image-compliance` in the namespace `nirmata`:

```sh
kubectl create ns nirmata
kubectl apply -f "https://raw.githubusercontent.com/nirmata/demo-image-compliance/refs/heads/main/config/install.yaml"
```

Run port forwarding to send requests to the service:

```sh
kubectl -n nirmata port-forward svc/nirmata-image-compliance-svc 9443:443
```

In a new shell, post a request with signed image:

```sh
curl -k https://localhost:9443/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:signed"}}'
```

Post a request with unsigned image

```sh
curl -k https://localhost:9443/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:unsigned"}}'
```

Update `POLICY_PATH` environment variable in deployment to block critical & high vulnerabilities: 

```sh
kubectl -n nirmata edit deploy nirmata-image-compliance
```

```
- name: POLICY_PATH
  value: oci://ghcr.io/nirmata/demo-image-compliance-policies:block-high-vulnerabilites
```

Post a request with signed image

```sh
curl -k https://localhost:9443/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:signed"}}'
```

This should fail, as it does not comply with the policy requirements.