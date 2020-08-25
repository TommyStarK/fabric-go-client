# fabric-go-client

[![Build Status](https://travis-ci.org/TommyStarK/fabric-go-client.svg?branch=master)](https://travis-ci.org/TommyStarK/fabric-go-client)
[![codecov](https://codecov.io/gh/TommyStarK/fabric-go-client/branch/master/graph/badge.svg)](https://codecov.io/gh/TommyStarK/fabric-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/TommyStarK/fabric-go-client)](https://goreportcard.com/report/github.com/TommyStarK/fabric-go-client)
[![GoDoc](https://godoc.org/github.com/TommyStarK/fabric-go-client?status.svg)](https://godoc.org/github.com/TommyStarK/fabric-go-client)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

:warning: Work in progress :warning:

This client enables Go developers to build solutions that interact with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-2.2/) thanks to the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go).

It is a wrapper around the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go) enabling managing resources in Fabric network and access to a channel on a Fabric network.

The client has been designed to use the new chaincode lifecycle as well as the gateway programming model, it is meant to be compliant with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-2.2/) version 2.2.

If you wish to use the legacy chaincode lifecyle and run the client against HLF 1.4, please take a look at this [version](https://github.com/TommyStarK/fabric-go-client/tree/v1.4) of the client.

## Usage

```bash
$ go get github.com/TommyStarK/fabric-go-client
```

## Test

```bash
$ ./hack/run-integration-tests.sh
```
