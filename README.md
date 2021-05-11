# k8s_config_merge

Merge multiples kubernetes config files.

## How to use?

```shell
$ k8s_config_merge [-d <configfile>] -s <config file> [-s <config file>]
```

where -d is the path to the file where you will merge the new files. The default value is $HOME/.kube/config.
and -s if the file you want to include.


