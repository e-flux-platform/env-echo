# env-echo

A docker container that echos back environment variables.

The primary use case is to print back configuration to a frontend application that does not support runtime configuration
via env (looking at you next.js). The process takes a prefix for env and prints back all env variables that start with it.

The response is served according to Accept header, plaintext and json are the only types supported.

Within k8s environment, the `envFrom` configuration field supports an optional `prefix` which can be used to automatically
prefix all variables coming from a particular config map, example below:

```yaml

```
