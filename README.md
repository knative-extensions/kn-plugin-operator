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

Check if plugin is loaded

```sh
kn -h
```

Run it

```sh
kn operator -h
```

You can use the built binary to run the commands. You can also use the bash scripts directly to run your commands.
All the bash scripts are available under the directory [scripts](scripts/).
