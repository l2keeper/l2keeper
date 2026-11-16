# Миграция с MetalLB

L2Keeper закрывает **L2-режим** MetalLB и основной набор IPAM/аннотаций,
которые используют небольшие bare-metal кластеры. BGP, BFD, ECMP и продвинутый
traffic engineering MetalLB не реализованы — если они нужны, оставляй MetalLB.

## Матрица совместимости

| MetalLB | L2Keeper | Примечание |
|---|---|---|
| `IPAddressPool` | `VIPPool` | `addresses`, `autoAssign`, `avoidBuggyIPs` |
| `IPAddressPool.spec.serviceAllocation` | `VIPPool.spec.serviceAllocation` | селекторы namespace/service |
| `L2Advertisement` | `VIPPool.spec.mode: l2` | отдельного объекта advertisement пока нет |
| `metallb.io/address-pool` | поддерживается | legacy `metallb.universe.tf/address-pool` тоже |
| `metallb.io/loadBalancerIPs` | поддерживается | legacy-префикс и `spec.loadBalancerIP` тоже |
| `metallb.io/allow-shared-ip` | поддерживается | `VIPBinding.spec.sharingKey`, порты не должны пересекаться |
| `metallb.io/ip-allocated-from-pool` | не пишется | смотри `VIPBinding.status.allocatedIPs` |
| BGP/BFD/ECMP | не поддерживается | вне зоны L2Keeper |
| Per-pool node/interface advertisement | пока не реализовано | speaker использует глобальные `speaker.interface`/`nodeIPType` |

## Шаги миграции

1. **Выбери окно переключения.** Один VIP — ровно один владелец. Не запускай
   MetalLB и L2Keeper speaker на один и тот же адрес одновременно.
2. **Отключи ServiceLB k3s/RKE2**, если заменяешь и klipper:

   ```yaml
   # /etc/rancher/k3s/config.yaml
   disable:
     - servicelb
   ```

   ```bash
   sudo systemctl restart k3s
   ```

3. **Поставь L2Keeper** (CRD и чарт):

   ```bash
   helm upgrade --install l2keeper oci://ghcr.io/l2keeper/charts/l2keeper \
     --version <chart-version> \
     --namespace l2keeper-system \
     --create-namespace \
     -f values-l2keeper.yaml
   ```

4. **Преобразуй каждый `IPAddressPool` в `VIPPool`:**

   ```yaml
   # MetalLB
   apiVersion: metallb.io/v1beta1
   kind: IPAddressPool
   metadata:
     name: default
     namespace: metallb-system
   spec:
     addresses:
       - 192.168.56.240-192.168.56.249
     autoAssign: true
     avoidBuggyIPs: true
   ---
   # L2Keeper
   apiVersion: l2keeper.github.io/v1alpha1
   kind: VIPPool
   metadata:
     name: default
   spec:
     addresses:
       - 192.168.56.240-192.168.56.249
     autoAssign: true
     avoidBuggyIPs: true
     mode: l2
     serviceAllocation:
       namespaceSelector:
         matchLabels:
           kubernetes.io/metadata.name: production
   ```

5. **Перенеси смысл `L2Advertisement`.** L2Keeper анонсирует `mode: l2` пулы с
   подходящих speaker-нод. Отдельного `L2Advertisement` пока нет; выбор
   нод/интерфейсов на уровне пула не реализован, используются глобальные
   `speaker.interface` и `speaker.nodeIPType`.

6. **Останови MetalLB до того, как L2Keeper начнёт владеть адресами:**

   ```bash
   helm uninstall metallb -n metallb-system
   ```

   CRD MetalLB удаляй только после удаления его custom resources. Не оставляй
   оба controller'а работать одновременно.

7. **Проверь:**

   ```bash
   kubectl get vippool
   kubectl get vipbinding -A -o wide
   kubectl get svc -A --field-selector spec.type=LoadBalancer
   kubectl -n l2keeper-system get lease
   kubectl -n l2keeper-system logs ds/l2keeper-speaker --tail=100
   ```

   Service без ручного binding'а должен получить `auto-<service>` и
   `EXTERNAL-IP` из подходящего пула. `spec.loadBalancerIP`,
   `metallb.io/loadBalancerIPs`, `metallb.io/address-pool` и
   `metallb.io/allow-shared-ip` продолжают работать.

8. **Проверь трафик** через TCP, а не ICMP:

   ```bash
   arping -c 3 <VIP>          # IPv4 l2
   ip -6 neigh show <VIP>     # IPv6 l2
   nc -vz <VIP> 443
   curl -k --resolve example.com:443:<VIP> https://example.com/
   ```

## Откат

1. Удали L2Keeper и его `VIPPool`/`VIPBinding`.
2. Верни MetalLB и создай `IPAddressPool`/`L2Advertisement` заново.
3. Пересоздай Service или дай MetalLB заново выделить адреса.
4. Очисти старые адреса в `Service.status.loadBalancer.ingress`, если Service
   переиспользуется.

## Известные пробелы

- нет BGP/BFD/ECMP;
- нет per-pool селекторов нод/интерфейсов (`L2Advertisement`);
- нет `loadBalancerSourceRanges`;
- интеграция с Gateway API опциональна и не связана с миграцией MetalLB;
- IPv6 VIP анонсируются через NDP, но перед продакшеном прогони e2e на своём
  интерфейсе и ядре.
