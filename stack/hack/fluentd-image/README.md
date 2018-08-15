Based mainly on https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/fluentd-elasticsearch/fluentd-es-image with a few modifications from https://github.com/fluent/fluentd-kubernetes-daemonset/tree/master/docker-image/v1.2/debian-logzio.

Notable changes:
- Adding the logzio and rewrite-tag-filter plugins to the Gemfile
- Adding the parser_kubernetes and passthru plugins to the plugins folder