# scheduler

[![Build Status](https://github.com/pipego/scheduler/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/pipego/scheduler/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/pipego/scheduler/branch/main/graph/badge.svg?token=y5anikgcTz)](https://codecov.io/gh/pipego/scheduler)
[![Go Report Card](https://goreportcard.com/badge/github.com/pipego/scheduler)](https://goreportcard.com/report/github.com/pipego/scheduler)
[![License](https://img.shields.io/github/license/pipego/scheduler.svg)](https://github.com/pipego/scheduler/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/pipego/scheduler.svg)](https://github.com/pipego/scheduler/tags)



## Introduction

*scheduler* is the scheduler of [pipego](https://github.com/pipego) written in Go.



## Prerequisites

- Go >= 1.18.0



## Run

```bash
version=latest make build
./bin/scheduler --config-file="$PWD"/config/config.yml --listen-url=:28082
```



## Docker

```bash
version=latest make docker
docker run -v "$PWD"/config:/tmp ghcr.io/pipego/scheduler:latest --config-file=/tmp/config.yml --listen-url=:28082
```



## Usage

```
pipego scheduler

Usage:
  scheduler [flags]

Flags:
  -c, --config-file string   config file (.yml)
  -h, --help                 help for scheduler
  -l, --listen-url string    listen url (host:port)
  -v, --version              version for scheduler
```



## Settings

*scheduler* parameters can be set in the directory [config](https://github.com/pipego/scheduler/blob/main/config).

An example of configuration in [config.yml](https://github.com/pipego/scheduler/blob/main/config/config.yml):

```yaml
apiVersion: v1
kind: scheduler
metadata:
  name: scheduler
spec:
  fetch:
    disabled:
      - name: MetalFlow
        path: ./fetch-metalflow
    enabled:
      - name: LocalHost
        path: ./fetch-localhost
  filter:
    enabled:
      - name: NodeName
        path: ./filter-nodename
        priority: 1
      - name: NodeAffinity
        path: ./filter-nodeaffinity
        priority: 2
      - name: NodeResourcesFit
        path: ./filter-noderesourcesfit
        priority: 3
      - name: NodeUnschedulable
        path: ./filter-nodeunschedulable
        priority: 4
  score:
    enabled:
      - name: NodeResourcesFit
        path: ./score-noderesourcesfit
        weight: 2
      - name: NodeResourcesBalancedAllocation
        path: ./score-noderesourcesbalancedallocation
        weight: 1
  logger:
    callerSkip: 2
    fileCompress: false
    fileName: scheduler.log
    logLevel: debug
    maxAge: 1
    maxBackups: 60
    maxSize: 100
```



## Protobuf

```json
{
  "apiVersion": "v1",
  "kind": "scheduler",
  "metadata": {
    "name": "scheduler"
  },
  "spec": {
    "task": {
      "name": "task1",
      "nodeName": "node1",
      "nodeSelectors": [
        "ssd"
      ],
      "requestedResource": {
        "milliCPU": 256,
        "memory": 512,
        "storage": 1024
      },
      "toleratesUnschedulable": true
    },
    "nodes": [
      {
        "name": "node1",
        "host": "127.0.0.1",
        "label": "ssd",
        "allocatableResource": {
          "milliCPU": 1024,
          "memory": 2048,
          "storage": 4096
        },
        "requestedResource": {
          "milliCPU": 512,
          "memory": 1024,
          "storage": 2048
        },
        "unschedulable": true
      }
    ]
  }
}
```



## Plugins

- [plugin-fetch](https://github.com/pipego/plugin-fetch)
- [plugin-filter](https://github.com/pipego/plugin-filter)
- [plugin-score](https://github.com/pipego/plugin-score)



## License

Project License can be found [here](LICENSE).



## Reference

- [asynq](https://github.com/hibiken/asynq)
- [asynqmon](https://github.com/hibiken/asynqmon)
- [bufio-example](https://golang.org/src/bufio/example_test.go)
- [cuelang](https://cuelang.org)
- [dagger](https://dagger.io/)
- [drone-dag](https://github.com/drone/dag)
- [drone-livelog](https://github.com/harness/drone/tree/master/livelog)
- [drone-pipeline](https://docs.drone.io/pipeline/overview/)
- [grpctest](https://github.com/grpc/grpc-go/tree/master/internal/grpctest)
- [kube-parallelize](https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/framework/parallelize/parallelism.go)
- [kube-schduler](https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/schedule_one.go)
- [kube-scheduling](https://cloud.tencent.com/developer/article/1644857)
- [kube-scheduling](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/kube-scheduler/)
- [kube-scheduling](https://kubernetes.io/zh/docs/reference/scheduling/config/)
- [kube-workqueue](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/client-go/util/workqueue)
- [machinery](https://github.com/RichardKnop/machinery/blob/master/v2/example/go-redis/main.go)
- [termui](https://github.com/gizak/termui)
- [websocket-command](https://github.com/gorilla/websocket/tree/master/examples/command)
- [wiki-dag](https://en.wikipedia.org/wiki/Directed_acyclic_graph)
