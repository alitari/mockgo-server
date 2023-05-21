# mockgo serverless with knative

## prerequisites

- [knative](https://knative.dev/) cluster
- [kn](https://knative.dev/docs/client/install-kn/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [hey](https://github.com/rakyll/hey)

We install a auto-scaling mockgo-serverless service with knative. We are using redis to store the mockgo-server state.


## install redis

Installing just a master without replicas and without persistence, see [redis bitznami charts](https://github.com/bitnami/charts/tree/main/bitnami/redis) for all configuration options.

```bash
helm upgrade redis oci://registry-1.docker.io/bitnamicharts/redis --install --namespace redis --create-namespace --set master.persistence.enabled=false,replica.replicaCount=0
export REDIS_PASSWORD=$(kubectl get secret --namespace redis redis -o jsonpath="{.data.redis-password}" | base64 --decode)
```


## install the mockgo config using the [people-mock example](./Examples.md#people-mock)

```bash
kubectl ns default
kubectl create configmap people-mock-config  --from-file=test/main/people-mock.yaml
# ksvc with autoscaling and short scale window for demo purposes
kn service create mockgo --image=alitari/mockgo-redis:latest --mount /mockdir=cm:people-mock-config --env MOCK_DIR=/mockdir --env MOCK_PORT=8080 --env REDIS_ADDRESS=redis-master.redis.svc.cluster.local:6379 --env REDIS_PASSWORD=$REDIS_PASSWORD  --scale "0..5" --scale-window "10s" --force
export MOCKGO_URL=$(kn service describe mockgo -o url)
# check
curl -v $MOCKGO_URL/__/health
```

## create a croud of people

```bash
# create 10 people
for i in {1..10}
do
  age=$(( ( RANDOM % 100 )  + 1 ))
  curl -v -X PUT $MOCKGO_URL/storePeople -d "{ \"name\": \"Alex-$i\", \"age\": $age }"
done

# works also if service is scaled to zero
curl -v $MOCKGO_URL/getPeople/adults
curl -v $MOCKGO_URL/getPeople/children


# scaling up with a request storm
hey -n 10000 -c 100 -q 100 $MOCKGO_URL/getPeople/adults
```
