# env-echo

A docker container that echos back environment variables.

The primary use case is to print back configuration to a frontend application that does not support runtime configuration
via env (looking at you next.js). The process takes a prefix for env and prints back all env variables that start with it.

The response is served according to Accept header, plaintext and json are the only types supported.

### Important

The default prefix is `CONFIG_`, don't deploy this container with empty prefix because it will print the entire environment
including any secrets that are present.

### Example

Within k8s environment, the `envFrom` configuration field supports an optional `prefix` which can be used to automatically
prefix all variables coming from a particular config map, example below:

```yaml

# Config map, note no need for prefix since we load the envFrom with a prefix
apiVersion: v1
kind: ConfigMap
metadata:
  name: ui-config
data:
  GRAPHQL_URL: "example.com/graphql"
---
# Define deployment, load config map
apiVersion: apps/v1
kind: Deployment
metadata:
  name: &name ui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: *name
  template:
    metadata:
      labels:
        app: *name
    spec:
      containers:
      - name: *name
        image: docker.io/your-org/your-image:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 3000
        resources:
          requests:
            cpu: 0m
            memory: 0Mi
          limits:
            cpu: 1000m
            memory: 3Gi
      - name: ui-env
        image: quay.io/road/env-echo:latest
        imagePullPolicy: Always
        envFrom:
          - configMapRef:
                name: ui-config
            prefix: CONFIG_
        ports:
          - containerPort: 8080
        resources:
          requests:
            cpu: 0m
            memory: 0Mi
          limits:
            cpu: 200m
            memory: 500Mi
---
# Define service for both
apiVersion: v1
kind: Service
metadata:
  labels:
    name: ui
  name: ui
spec:
  selector:
    app: ui
  ports:
    - protocol: TCP
      port: 3000
      name: ui
    - protocol: TCP
      port: 8080
      name: ui-env
---
# Define ingress with path mapping, example with Traefik
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: ui
spec:
  entryPoints:
    - web
  routes:
    - kind: Rule
      match: Host(`example.com`) && PathPrefix(`/.env`)
      services:
        - name: ui
          port: 8080
    - kind: Rule
      match: Host(`example.com`)
      services:
        - name: ui
          port: 3000
```

