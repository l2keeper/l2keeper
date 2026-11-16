# Миграция с ServiceLB (klipper)

В k3s/RKE2 встроен ServiceLB (klipper). Он создаёт `svclb`-поды с `hostPort` и
DNAT+MASQUERADE: появляется лишний хоп, а client source IP может теряться.
L2Keeper заменяет этот control plane, не проксируя пакеты: kube-proxy ставит
правила по адресам, которые публикует L2Keeper.

Гайд покрывает оба режима:

- только IP нод — ближе всего к ServiceLB, без VIP;
- стабильные VIP — L2 speaker L2Keeper или внешний владелец VIP.

## 1. Отключи ServiceLB

На каждой server-ноде k3s/RKE2:

```yaml
# /etc/rancher/k3s/config.yaml
disable:
  - servicelb
```

```bash
sudo systemctl restart k3s
```

После рестарта существующие `svclb`-поды удаляются, а `LoadBalancer` Service
временно остаются без внешнего адреса. Ставь L2Keeper в это же окно, чтобы
сократить простой.

## 2. Установи L2Keeper

```bash
helm upgrade --install l2keeper oci://ghcr.io/l2keeper/charts/l2keeper \
  --version <chart-version> \
  --namespace l2keeper-system \
  --create-namespace \
  -f values-l2keeper.yaml
```

### Вариант A: только IP нод

Минимальный шаг от ServiceLB. L2Keeper публикует IP нод в статусе Service, а
kube-proxy делает остальное.

```yaml
controller:
  publishNodeIPs: true
  requireClass: false

speaker:
  enabled: false

pools: {}
bindings: {}
```

DNS продолжает указывать на IP нод. VIP и speaker не нужны.

### Вариант B: стабильные L2 VIP

```yaml
controller:
  publishNodeIPs: false   # true, если DNS всё ещё использует IP нод
  requireClass: false

speaker:
  enabled: true

pools:
  ingress-vips:
    addresses:
      - 192.168.56.240-192.168.56.249
    autoAssign: true
    mode: l2
    serviceAllocation:
      namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: ingress-nginx
      serviceSelector:
        matchLabels:
          app.kubernetes.io/name: ingress-nginx
```

После появления `EXTERNAL-IP` обнови DNS на выделенный VIP.

### Вариант C: оставить существующий VIP

Если VIP уже держит keepalived или аппаратный балансировщик, L2 speaker не
включай. Создай явный binding:

```yaml
speaker:
  enabled: false

bindings:
  ingress-keepalived:
    namespace: ingress-nginx
    serviceRef: my-release-ingress-nginx-controller
    mode: external
    vips:
      - 192.168.56.250
```

## 3. Убери оставшиеся `svclb`-объекты

```bash
kubectl -n kube-system get ds | grep svclb || true
kubectl -n kube-system get pods | grep svclb || true
```

Если после отключения ServiceLB что-то осталось, удали соответствующие
DaemonSet'ы/поды. После отключения k3s ServiceLB controller они больше не
управляются.

## 4. Проверь

```bash
kubectl -n l2keeper-system get deploy,ds,pods
kubectl get vippool
kubectl get vipbinding -A
kubectl get svc -A --field-selector spec.type=LoadBalancer
kubectl -n l2keeper-system get lease
kubectl -n l2keeper-system logs ds/l2keeper-speaker --tail=100
```

Ожидаемо:

- `EXTERNAL-IP` снова заполнен (IP нод или VIP);
- `svclb`-подов нет;
- для `mode: l2` работают `arping`/`ip -6 neigh` и TCP-проверки; ICMP echo не
  отвечает;
- в логах ingress виден IP клиента при `externalTrafficPolicy: Local` и
  локальном ingress endpoint.

## 5. Откат

1. Удали L2Keeper (и его `VIPPool`/`VIPBinding`).
2. Убери `disable: [servicelb]` из конфига k3s и перезапусти k3s.
3. ServiceLB заново создаст `svclb`-поды для существующих `LoadBalancer`
   Service.

## Примечания

- Один VIP — один владелец. Не запускай L2Keeper L2 и keepalived на одном VIP.
- Режим IP нод не требует пула, binding'ов и speaker'а.
- При переходе с IP нод на VIP обнови DNS и поставь `publishNodeIPs: false`,
  чтобы публиковались только VIP.
- Для MetalLB смотри [migration-metallb.md](migration-metallb.md).
