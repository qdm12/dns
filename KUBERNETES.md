# Using `qmcgaw/dns` as the DNS Resolver in K8s

This guide provides an introduction on how to use `qmcgaw/dns` as the DNS resolver in your Kubernetes cluster. This brings the advantage of supporting DoT, DoH, DNSSEC, rebinding protection and more without needing to heavily adjust already present components in your cluster. Note that we are already using `v2` (which is still in beta). You may use `v1` as well, but make sure to change the environment variables' names and the image tag.

**Disclaimer** This content is community-supported. If you find a bug or error in this configuration, please open an issue and provide a PR to fix it.

## The Manifest Files

First of all, we want to provide the environment variables with a `ConfigMap`. These settings are just an example - you may tune them to your liking. The configuration is straight-forward.

``` YAML
---
kind: ConfigMap
apiVersion: v1

metadata:
  name: dns.environment

immutable: true

data:
  UPSTREAM_TYPE: DoT
  DOT_RESOLVERS: cloudflare,quad9,libredns,google
  DOH_RESOLVERS: cloudflare,quad9,libredns,google

  BLOCK_MALICIOUS: 'off'
  BLOCK_SURVEILLANCE: 'off'
  BLOCK_ADS: 'off'

  LOG_LEVEL: warning
  MIDDLEWARE_LOG_REQUESTS: 'off'
  MIDDLEWARE_LOG_RESPONSES: 'off'

  DOT_TIMEOUT: 3s
  DOH_TIMEOUT: 3s

  LISTENING_ADDRESS: ':8053'

  CACHE_TYPE: lru
  CACHE_LRU_MAX_ENTRIES: '10000'

  CHECK_DNS: 'off'
  UPDATE_PERIOD: '0' # actually 24h, but there is bug currently
  REBINDING_PROTECTION: 'on'
```

The next manifest file is the `Service` definition, which is straight-forward as well. Make sure to provide a `ClusterIP` that is in your service CIDR range, i.e. an IP in the range of all services.

``` YAML
---
kind: Service
apiVersion: v1

metadata:
  name: dns
  labels:
    app: dns

spec:
  type: ClusterIP
  clusterIP: <A CLUSTER IP IN YOUR SERVICE CIDR RANGE>

  selector:
    app: dns

  ports:
    - name: dns
      port: 53
      targetPort: dns
      protocol: UDP
```

Last but not least, deploying the workload with a `Deployment`. This is the most complex part. Since `qmcgaw/dns` does not require root privileges, we can set a strict `securityContext`.

``` YAML
---
apiVersion: apps/v1
kind: Deployment

metadata:
  name: dns

spec:
  replicas: 1
  selector:
    matchLabels:
      app: dns

  template:
    metadata:
      labels:
        app: dns

    spec:
      containers:
        - name: dns
          # use proper tag when beta is over and v2 released
          image: qmcgaw/dns:v2.0.0-beta
          imagePullPolicy: IfNotPresent

          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            runAsUser: 1000
            runAsGroup: 1000
            runAsNonRoot: true
            privileged: false

          resources:
            limits:
              memory: 75Mi
              cpu: 100m
            requests:
              memory: 25Mi
              cpu: 20m

          volumeMounts:
            - name: tmp-files
              mountPath: /unbound
              readOnly: false

          ports:
            - name: dns
              containerPort: 8053
              protocol: UDP

          envFrom:
            - configMapRef:
                name: dns.environment

      restartPolicy: Always

      volumes:
        - name: tmp-files
          emptyDir: {}
```

## Integrating `qmcgaw/dns` with coreDNS

It is encouraged to use `qmcgaw/dns` together with coreDNS is your cluster. You will just need to replace one line in coreDNS' configuration. If there is a `forward` chain in the `Corefile`, replace it with

``` INI
forward . dns://<A CLUSTER IP IN YOUR SERVICE CIDR RANGE> {
  prefer_udp
}
```

This will forward DNS traffic to your container. The whole coreDNS configuration may look like this afterwards:

``` INI
.:53 {
  errors
  health
  ready
  kubernetes cluster.local in-addr.arpa ip6.arpa {
    pods insecure
    fallthrough in-addr.arpa ip6.arpa
  }
  hosts /etc/coredns/NodeHosts {
    ttl 60
    reload 15s
    fallthrough
  }
  prometheus :9153
  forward . dns://10.100.0.11 {
    prefer_udp
  }
  cache 30
  loop
  reload
  loadbalance
}
```

Using coreDNS has the advantage of a working local, in-cluster DNS resolution for services.
