demo stuff

code is hacked together it may eat your computer or cat :P

kustomize version used: `{Version:4.1.2 GitCommit:$Format:%H$ BuildDate:2021-04-19T00:04:53Z GoOs:linux GoArch:amd64}`
try it:

```bash
# build plugin
go mod download
make all

# configure kustomize plugin path
export KUSTOMIZE_PLUGIN_HOME=$PWD/.cache/plugins

# render example kustomization into packagedeployment
kustomize build --enable-alpha-plugins config/example/kustomize

# render example kustomization (which inflates a helm chart) into packagedeployment
# needs helm v3 installed
kustomize build --enable-alpha-plugins config/example/kustomize-helm
```
