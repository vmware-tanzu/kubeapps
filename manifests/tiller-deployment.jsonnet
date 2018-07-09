// Based on `helm init -o json` from helm version v2.9.1
{
    "apiVersion": "extensions/v1beta1",
    "kind": "Deployment",
    "metadata": {
        "creationTimestamp": null,
        "labels": {
            "app": "helm",
            "name": "tiller"
        },
        "name": "tiller-deploy",
        "namespace": "kube-system"
    },
    "spec": {
        "replicas": 1,
        "strategy": {},
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app": "helm",
                    "name": "tiller"
                }
            },
            "spec": {
                "containers": [
                    {
                        "env": [
                            {
                                "name": "TILLER_NAMESPACE",
                                "value": "kube-system"
                            },
                            {
                                "name": "TILLER_HISTORY_MAX",
                                "value": "0"
                            }
                        ],
                        "image": "gcr.io/kubernetes-helm/tiller:v2.9.1",
                        "imagePullPolicy": "IfNotPresent",
                        "livenessProbe": {
                            "httpGet": {
                                "path": "/liveness",
                                "port": 44135
                            },
                            "initialDelaySeconds": 1,
                            "timeoutSeconds": 1
                        },
                        "name": "tiller",
                        "ports": [
                            {
                                "containerPort": 44134,
                                "name": "tiller"
                            },
                            {
                                "containerPort": 44135,
                                "name": "http"
                            }
                        ],
                        "readinessProbe": {
                            "httpGet": {
                                "path": "/readiness",
                                "port": 44135
                            },
                            "initialDelaySeconds": 1,
                            "timeoutSeconds": 1
                        },
                        "resources": {},
                        "args": [
                            "--storage=secret",
                        ],
                    }
                ]
            }
        }
    },
    "status": {}
}
