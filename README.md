# Image Verification Serivce

Verify any json payload against kyverno's image verification policies

## Usage

Create a kind cluster 

```sh
make kind-create
```

Install kyverno-image-verification-service

```sh
make kind-install
```

Start a netshoot pod

```sh
kubectl run netshoot --rm -i --tty --image nicolaka/netshoot
```

Post a request with signed image

```sh
curl -k https://kyverno-image-verification-service-svc.kyverno-image-verification-service/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:signed"}}'
```

Post a request with unsigned image

```sh
curl -k https://kyverno-image-verification-service-svc.kyverno-image-verification-service/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:unsigned"}}'
```

Update `POLICY_PATH` environment variable in deployment to: 

```sh
kubectl -n kyverno-image-verification-service edit deploy kyverno-image-verification-service
```

```
- name: POLICY_PATH
  value: oci://ghcr.io/vishal-chdhry/ivpol:high-crit-vuln
```

Post a request with signed image

```sh
curl -k https://kyverno-image-verification-service-svc.kyverno-image-verification-service/verifyimages -X POST -d '{"foo":{"bar": "ghcr.io/kyverno/test-verify-image:signed"}}'
```

This should fail now