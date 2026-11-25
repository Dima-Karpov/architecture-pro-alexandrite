# Task 3.1 — OpenTelemetry + Jaeger (Minikube)

Все команды — из каталога **`Task3/`**:

```bash
cd Task3
```

## 1. Docker + Minikube

Запусти Docker Desktop (Ready), затем:

```bash
minikube start --addons=ingress
```

## 2. cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
kubectl -n cert-manager wait --for=condition=Available deployment --all --timeout=180s
```

## 3. Jaeger Operator

```bash
kubectl create namespace observability
kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.51.0/jaeger-operator.yaml -n observability

# образ kube-rbac-proxy в v1.51 с gcr.io часто 404 — сразу подменяем
kubectl -n observability set image deployment/jaeger-operator \
  kube-rbac-proxy=quay.io/brancz/kube-rbac-proxy:v0.14.2

kubectl -n observability rollout status deployment/jaeger-operator --timeout=180s
kubectl -n observability wait --for=condition=Ready certificate/jaeger-operator-serving-cert --timeout=180s
```

Ожидание: pod operator `2/2 Running`.

## 4. Инстанс Jaeger

```bash
kubectl apply -f k8s/jaeger-instance.yaml
kubectl get pods -l app.kubernetes.io/instance=simplest
```

Ожидание: pod `simplest` = `Running`.

## 5. Сборка и деплой сервисов

```bash
minikube image build -t service-a:latest services/service-a/
minikube image build -t service-b:latest services/service-b/

kubectl apply -f k8s/services.yaml
kubectl rollout status deployment/service-a --timeout=120s
kubectl rollout status deployment/service-b --timeout=120s
```

Ожидание: оба pod `1/1 Running`.

## 6. Вызов (трейс A → B)

```bash
kubectl exec -it "$(kubectl get pods -l app=service-a -o jsonpath='{.items[0].metadata.name}')" -- wget -qO- http://service-a:8080
```

Ожидание: ответ вида `service-a -> service-b` + тело от service-b.

## 7. Jaeger UI

```bash
kubectl port-forward svc/simplest-query 16686:16686
```

1. http://localhost:16686  
2. Service: `service-a` → Find Traces  
3. Открыть трейс с `service-a` и `service-b`  
4. Скрин → `jaeger-trace.png` в этой папке (для сдачи 3.1)

