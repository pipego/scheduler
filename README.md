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
usage: scheduler --config-file=CONFIG-FILE --listen-url=LISTEN-URL [<flags>]

pipego scheduler

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --version                  Show application version.
  --config-file=CONFIG-FILE  Config file (.yml)
  --listen-url=LISTEN-URL    Listen URL (host:port)
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
      - name: LocalHost
        path: ./plugin/fetch-localhost
      - name: MetalFlow
        path: ./plugin/fetch-metalflow
  filter:
    enabled:
      - name: NodeName
        path: ./plugin/filter-nodename
        weight: 4
      - name: NodeAffinity
        path: ./plugin/filter-nodeaffinity
        weight: 3
      - name: NodeResourcesFit
        path: ./plugin/filter-noderesourcesfit
        weight: 2
      - name: NodeUnschedulable
        path: ./plugin/filter-nodeunschedulable
        weight: 1
  score:
    enabled:
      - name: NodeResourcesFit
        path: ./plugin/score-noderesourcesfit
        weight: 2
      - name: NodeResourcesBalancedAllocation
        path: ./plugin/score-noderesourcesbalancedallocation
        weight: 1
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
      "nodeSelector": {
        "diskType": [
          "ssd"
        ]
      },
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
        "label": {
          "diskType": "ssd"
        },
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



## plugins

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
- [drone-pipeline](https://docs.drone.io/pipeline/overview/)
- [grpctest](https://github.com/grpc/grpc-go/tree/master/internal/grpctest)
- [kube-schduler](https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/schedule_one.go)
- [kube-scheduling](https://cloud.tencent.com/developer/article/1644857)
- [kube-scheduling](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/kube-scheduler/)
- [kube-scheduling](https://kubernetes.io/zh/docs/reference/scheduling/config/)
- [machinery](https://github.com/RichardKnop/machinery/blob/master/v2/example/go-redis/main.go)
- [termui](https://github.com/gizak/termui)
- [websocket-command](https://github.com/gorilla/websocket/tree/master/examples/command)
- [wiki-dag](https://en.wikipedia.org/wiki/Directed_acyclic_graph)
