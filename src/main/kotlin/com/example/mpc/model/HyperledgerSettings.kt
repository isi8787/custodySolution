package com.example.mpc.model


import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class HyperledgerSettings(

    @Value("\${hyperledger-fabric.k8sEnabled}")
    var k8sEnabled: Boolean,

    @Value("\${hyperledger-fabric.ink8sCluster}")
    val inK8sCluster: Boolean,

    @Value("\${hyperledger-fabric.roleName}")
    var roleName: String,

    @Value("\${hyperledger-fabric.userType}")
    var userType: String,

    @Value("\${hyperledger-fabric.keystore}")
    var keystore: String,

    @Value("\${hyperledger-fabric.connectionFile}")
    var connectionFile: String,

    @Value("\${hyperledger-fabric.defaultSigner}")
    var defaultSigner: String,

    @Value("\${hyperledger-fabric.channel}")
    var channel: String,

    @Value("\${hyperledger-fabric.chaincode}")
    var chaincode: String,

    @Value("\${hyperledger-fabric.eventCheckPointer}")
    var eventCheckPointer: String,

    @Value(value = "\${hyperledger-fabric.kubeConfigDirectory}")
    val kubeConfigDirectory: String,

    val apiVersion: String = "v1",

    @Value(value = "\${hyperledger-fabric.orgId")
    val orgId: String,

    @Value(value = "\${hyperledger-fabric.ordererServiceLabel}")
    val ordererServiceLabel: String,

    @Value(value = "\${hyperledger-fabric.peerServiceLabel}")
    val peerServiceLabel: String,

    @Value(value = "\${hyperledger-fabric.caServiceLabel}")
    val caServiceLabel: String
)

@Component
data class DFSConfig(

    @Value("\${dfs-server.dfsurl}")
    var dfsurl: String,

    @Value("\${dfs-server.defaultUser}")
    var dfsUser: String
)

@Component
data class IMSConfig(

    @Value("\${ims-server.imsurl}")
    var imsurl: String
)

@Component
data class DocusignConfig(

    @Value("\${docusign-endpoints.developer}")
    var developerurl: String,

    @Value("\${docusign-endpoints.production}")
    var productionurl: String
)

@Component
data class AVSConfig(

    @Value("\${avs-endpoints.avsqa}")
    var avsqa: String
)


@Component
data class ECFSConfig(

    @Value("\${ecfsURI.buyerURI}")
    var buyerURI: String,

    @Value("\${ecfsURI.funderURI}")
    var funderURI: String,

    @Value("\${ecfsURI.supplierURI}")
    var supplierURI: String
)

