# fabric-go-client

[![Build Status](https://travis-ci.org/TommyStarK/fabric-go-client.svg?branch=v1.4)](https://travis-ci.org/TommyStarK/fabric-go-client)
[![codecov](https://codecov.io/gh/TommyStarK/fabric-go-client/branch/v1.4/graph/badge.svg)](https://codecov.io/gh/TommyStarK/fabric-go-client/branch/v1.4)
[![Go Report Card](https://goreportcard.com/badge/github.com/TommyStarK/fabric-go-client)](https://goreportcard.com/report/github.com/TommyStarK/fabric-go-client)
[![GoDoc](https://godoc.org/github.com/TommyStarK/fabric-go-client?status.svg)](https://pkg.go.dev/github.com/TommyStarK/fabric-go-client)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

This client enables Go developers to build solutions that interact with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-1.4/whatsnew.html) thanks to the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go).

It is a wrapper around the [fabric-sdk-go](https://github.com/hyperledger/fabric-sdk-go) enabling managing resources in Fabric network and access to a channel on a Fabric network.
The client has been designed for being able to manage multiple channels and interact with multiple chaincodes.

This version is built to be compliant with [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/release-1.4/whatsnew.html) version 1.4. It uses the legacy chaincode lifecycle.

If you wish to use the new chaincode lifecyle as well as the gateway programming model, please take a look at this [version](https://github.com/TommyStarK/fabric-go-client) of the client.


## Usage

```bash
$ go get github.com/TommyStarK/fabric-go-client@v1.0.1-hlf-1.4
```

## Test

```bash
$ ./hack/run-integration-tests.sh
```
