# k8s_config_merge

A command-line tool for merging multiple Kubernetes configuration files while handling naming conflicts interactively.

## Description

k8s_config_merge is a utility that allows you to combine multiple Kubernetes configuration files (`kubeconfig`) into a single configuration file. It automatically handles naming conflicts for clusters, users, and contexts by prompting for new names when conflicts are detected.

## Features

- Merge multiple kubeconfig files into a single configuration
- Interactive conflict resolution for duplicate names
- Support for reading configuration from stdin
- Optional custom naming for imported configurations
- Automatic backup of existing configuration (destination file)
- Preserves existing configurations while adding new ones

## Installation

```bash
go get github.com/yourusername/k8s_config_merge
```

## Usage

```bash
k8s_config_merge [-d destination_file] [-n new_config_name] -s source_file1 [-s source_file2 ...]
```

### Options

- `-d string`: Destination file where the merged config will be saved (default: "~/.kube/config")
- `-n string`: Optional name to use for the imported configuration (applies to cluster, context, and user names)
- `-s string`: Source configuration file(s) to merge. Use "-" to read from stdin. Can be specified multiple times.

### Examples

1. Merge a single config file:

```bash
k8s_config_merge -s new-cluster.yaml
```

2. Merge multiple config files:

```bash
k8s_config_merge -s cluster1.yaml -s cluster2.yaml
```

3. Read config from stdin:

```bash
cat cluster-config.yaml | k8s_config_merge -s -
```

4. Specify custom name for imported config:

```bash
k8s_config_merge -s new-cluster.yaml -n production-cluster
```

5. Specify custom destination file:

```bash
k8s_config_merge -d ./custom-config -s new-cluster.yaml
```

## Behavior

1. The tool reads the destination config file (defaults to ~/.kube/config)
2. For each source config file:
   - Loads and validates the configuration
   - Checks for naming conflicts with existing configurations
   - Prompts for new names when conflicts are found
   - Merges the configuration into the destination
3. Saves the merged configuration back to the destination file

### Conflict Resolution

When naming conflicts are detected, the tool will:

1. Prompt for a new name
2. Validate the input (must contain only alphanumeric characters)
3. Update all related references in the configuration
4. Continue with the merge process

## License

MIT License (Copyright (C) 2021 rmelo)

## Author

Ricardo Melo (rmelo@ludia.com)
