# kn-plugin-operator

`kn-plugin-operator` is a plugin of Knative Client, for managing Knative with Knative Operator from the
command line.

## Description

With this plugin, you can install/uninstall Knative Operator, install/uninstall Knative components,
and configure Knative.

## Build and Install

You must
[set up your development environment](https://github.com/knative/client/blob/master/docs/DEVELOPMENT.md#prerequisites)
before you build.

**Building:**

Once you've set up your development environment, let's build the plugin.

```sh
$ go build -o kn-operator ./cmd/kn-operator.go
```

You'll get an executable plugin binary namely `kn-operator` in your current dir.
You're ready to use `kn-operator` as a stand alone binary, check the available
commands `./kn-operator -h`.

**Installing:**

If you'd like to use the plugin with `kn` CLI, install the plugin by simply
copying the executable file under `kn` plugins directory as:

```sh
mkdir -p ~/.config/kn/plugins
cp kn-operator ~/.config/kn/plugins
```

> Note: The plugins directory defaults to `$base_dir/plugins` relative to your [kn config file](https://knative.dev/docs/client/configure-kn/) location.
>  
> On Windows, the default plugins directory is in `%APPDATA%\kn\plugins`

Check if plugin is loaded

```sh
kn -h
```

Run it

```sh
kn operator -h
```

## Remote Component Installs With ClusterProfile

You can install Serving or Eventing to a remote target cluster by creating the
hub `KnativeServing` or `KnativeEventing` CR with `spec.clusterProfileRef`:

```sh
kn operator install -c serving \
  --namespace knative-serving \
  --cluster-profile spoke \
  --cluster-profile-namespace fleet-system

kn operator install -c eventing \
  --namespace knative-eventing \
  --cluster-profile spoke \
  --cluster-profile-namespace fleet-system
```

For remote installs, `--kubeconfig` must point to the hub cluster. The Knative
Operator must already be installed on the hub and configured with
`--clusterprofile-provider-file`; this plugin does not create provider
configuration.

`spec.clusterProfileRef` is immutable. Moving a component to another
ClusterProfile requires deleting and recreating the component CR.

You can use the built binary to run the commands. You can also use the bash scripts directly to run your commands.
All the bash scripts are available under the directory [scripts](scripts/).
