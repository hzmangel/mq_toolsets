# Message Queue Toolset

This tool is designed to measure the latency of messages sent through a message queue system.
It can be used to test the performance and reliability of various message queue implementations.

## Prerequisites

### Prepare the message queue

Currently only rabbitmq topic exchange and redis pub/sub is supported by this tool.
And the only test is getting average latency of message delivery.

Before starting the test, the supported service should be setup correctly,
and the URI should be formatted in  `schema://user:pass@host:port` format.

### Configuration file

`config.yaml` is the dev config file used while developing.
Which contains `uri`, `publishers` and `subscribers` section.

The `uri` field is connection string of service, and the format is mentioned above.

`publishers` section configure how many publishers will be used.
Each configuration may generate one or more publisher goroutines, which is controlled by `concurrency` field.

`subscribers` defines subscriber of message, only one subscriber will be generated for each item. 


### Launch the app

```sh
$ go run main.go

$ go run main.go -config foo.yaml
```
