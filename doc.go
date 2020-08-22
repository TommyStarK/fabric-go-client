// Package fabclient enables Go developers to build solutions that interact with Hyperledger Fabric thanks to the fabric-sdk-go.
//
// It enables creation and update of resources on a Fabric network.
// It allows administrators to create and/or update channnels, and for peers to join channels.
// Administrators can also perform chaincode related operations on a peer, such as installing, instantiating, and upgrading chaincode.
//
// Package fabclient also enables access to a channel on a Fabric network. It  provides a handler to interact with peers on specified channel.
// Client can query chaincode, execute chaincode and register/unregister for chaincode events on specific channel.
// Finally the client enables ledger queries on specified channel on a Fabric network
package fabclient
