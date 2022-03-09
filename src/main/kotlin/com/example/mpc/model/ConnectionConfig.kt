package com.example.mpc.model

import com.fasterxml.jackson.annotation.JsonProperty

data class ConnectionConfig(
    val name: String,
    val version: String,
    val client: Client,
    val channels: Channels,
    val organizations: Organizations,
    val orderers: Orderers,
    val peers: Peers,
    val certificateAuthorities: CertificateAuthorities
)

data class CertificateAuthorities(
    @JsonProperty("ca")
    val ca: CA
)

data class CA(
    var url: String,
    val caName: String
)

data class Channels(
    val mychannel: Mychannel
)

data class Mychannel(
    val orderers: List<String>,
    val peers: MychannelPeers? = null
)

data class MychannelPeers(
    @JsonProperty("peer")
    val peerOptions: PeerOptions? = null
)

data class PeerOptions(
    val endorsingPeer: Boolean,
    val chaincodeQuery: Boolean,
    val ledgerQuery: Boolean,
    val eventSource: Boolean
)

data class Client(
    val organization: String,
    val connection: Connection
)

data class Connection(
    val timeout: Timeout
)

data class Timeout(
    val peer: Peer,
    val orderer: String
)

data class Peer(
    val endorser: String
)

data class Orderers(
    @JsonProperty("orderer")
    val ordererName: Entity
)

data class Entity(
    var url: String
)

data class Organizations(
    @JsonProperty("Org")
    val org: Org
)

data class Org(
    val mspid: String,
    val peers: List<String>,
    val certificateAuthorities: List<String>
)

//TODO: Read the property from the org.
/*
fun getPeerName(prefix, org : Organization) {
  return ("hlf-org-peer--" + org.orgName)
}
*/
data class Peers(
    @JsonProperty("peer")
    val peer: Entity
)
