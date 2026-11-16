# E2E smoke test

Validates automatic VIP allocation on a k3s cluster in k3d:

1. installs the chart with the local `l2keeper:e2e` image;
2. creates a `LoadBalancer` Service without a `VIPBinding`;
3. waits for `auto-<service>` to allocate an address from `VIPPool`;
4. checks that the address is published in the Service status;
5. deletes the Service and verifies that the binding is garbage-collected.

Prerequisites: `docker`, `k3d`, `kubectl`, `helm`.

```bash
docker build -t l2keeper:e2e .

k3d cluster create l2keeper-e2e --agents 1 \
  --k3s-arg "--disable=servicelb@server:*"
k3d image import -c l2keeper-e2e l2keeper:e2e
kubectl wait --for=condition=Ready nodes --all --timeout=180s

./test/e2e/run.sh

k3d cluster delete l2keeper-e2e
```

GitHub Actions runs the same test in `.github/workflows/e2e.yaml`.
