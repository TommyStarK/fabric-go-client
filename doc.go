// Package fabclient aims to facilitate the development of solutions that interact with Hyperledger Fabric thanks
// to the fabric-sdk-go.
//
// Package fabclient enables the creation and update of a channel, for peers to join channels. It allows administrators
// to perform chaincode related operations on a peer. It uses the legacy chaincode lifecycle enabling to install, instantiate
// and upgrade a chaincode.
// Furthermore, package fabclient provides access to a channel on a Fabric network, allowing users to query/invoke chaincodes,
// register/unregister for chaincode events on specific channel and perform ledger queries.
//
// It is a wrapper around the fabric-sdk-go. The client has been designed for being able to manage multiple channels
// and interact with multiple chaincodes.
package fabclient
