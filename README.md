# fabric-go-client

[![Build Status](https://travis-ci.org/TommyStarK/fabric-go-client.svg?branch=master)](https://travis-ci.org/TommyStarK/fabric-go-client)
[![codecov](https://codecov.io/gh/TommyStarK/fabric-go-client/branch/master/graph/badge.svg)](https://codecov.io/gh/TommyStarK/fabric-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/TommyStarK/fabric-go-client)](https://goreportcard.com/report/github.com/TommyStarK/fabric-go-client)
[![GoDoc](https://godoc.org/github.com/TommyStarK/fabric-go-client?status.svg)](https://godoc.org/github.com/TommyStarK/fabric-go-client)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

:warning: Work in progress :warning:

The aim of this client is to facilitate the development of solutions that interact with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-2.2/) thanks to the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go).

It is a wrapper around the fabric-sdk-go. The client has been designed for being able to manage multiple channels and interact with multiple chaincodes.

The client uses the new chaincode lifecycle as well as the gateway programming model, it is meant to be compliant with [Hyperledger Fabric v2.2](https://hyperledger-fabric.readthedocs.io/en/release-2.2/) .

If you wish to use the legacy chaincode lifecyle and run the client against Hyperledger Fabric v1.4, please take a look at [this version](https://github.com/TommyStarK/fabric-go-client/tree/v1.4) of the client.

:warning: For the moment, the client is only able to manage Go chaincodes, meaning you cannot install neither Node.js nor Java chaincodes with it.

## Usage

```bash
❯ go get github.com/TommyStarK/fabric-go-client
```

You will find an example of how to instantiate and use the client [here](https://github.com/TommyStarK/fabric-go-client/blob/master/example_test.go). An example of how to configure the client is also available [here](https://github.com/TommyStarK/fabric-go-client/blob/master/testdata/client/client-config.yaml).

## Test

```bash
❯ ./hack/run-integration-tests.sh
```
