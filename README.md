# fabric-go-client

[![Build Status](https://travis-ci.org/TommyStarK/fabric-go-client.svg?branch=master)](https://travis-ci.org/TommyStarK/fabric-go-client)
[![codecov](https://codecov.io/gh/TommyStarK/fabric-go-client/branch/master/graph/badge.svg)](https://codecov.io/gh/TommyStarK/fabric-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/TommyStarK/fabric-go-client)](https://goreportcard.com/report/github.com/TommyStarK/fabric-go-client)
[![GoDoc](https://godoc.org/github.com/TommyStarK/fabric-go-client?status.svg)](https://pkg.go.dev/github.com/TommyStarK/fabric-go-client@v1.0.0-hlf-2.3?tab=doc)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

The aim of this client is to facilitate the development of solutions that interact with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-2.3/) thanks to the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go).

It is a wrapper around the fabric-sdk-go. The client has been designed for being able to manage multiple channels and interact with multiple chaincodes.

The client uses the new chaincode lifecycle as well as the gateway programming model, it is meant to be compliant with [Hyperledger Fabric v2.3](https://hyperledger-fabric.readthedocs.io/en/release-2.3/) .

If you wish to use the legacy chaincode lifecyle and run the client against Hyperledger Fabric v1.4, please take a look at [this version](https://github.com/TommyStarK/fabric-go-client/tree/v1.4) of the client.

:warning: For the moment, the client is only able to manage Go chaincodes, meaning you cannot install neither Node.js nor Java chaincodes with it.

## Contribution

Each Contribution is welcomed and encouraged. I do not claim to cover each use cases neither completely master the Go nor Hyperledger Fabric. If you encounter a non sense or any trouble, you can open an issue and I will be happy to discuss about it :smile:

## Usage

```bash
❯ go get github.com/TommyStarK/fabric-go-client@v1.0.0-hlf-2.3
```

You will find an example of how to instantiate and use the client [here](https://github.com/TommyStarK/fabric-go-client/blob/master/example_test.go). An example of how to configure the client is also available [here](https://github.com/TommyStarK/fabric-go-client/blob/master/testdata/organizations/org1/client-config.yaml).

## Test

```bash
❯ ./hack/run-integration-tests.sh
```
