package com.example.defiServer.service

import com.example.defiServer.impl.GatewayBuilder
import com.example.defiServer.models.*
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import kotlinx.serialization.InternalSerializationApi
import org.hyperledger.fabric.gateway.*
import org.hyperledger.fabric.gateway.impl.FileCheckpointer
import org.hyperledger.fabric.gateway.spi.Checkpointer
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import java.nio.file.Path
import java.nio.file.Paths
import java.util.function.Consumer


@InternalSerializationApi
@Service
class EventService (
    @Autowired private val hfSettings: HyperledgerSettings,
    @Autowired private val commonService: CommonService
    ) {
    lateinit var builder: Gateway.Builder
    lateinit var wallet: Wallet

    init{
        commonService.registerAdmin("admin")
        println("hello world, I have just started up")
        println("EVENT HUB STARTED")
        // Load an existing wallet holding identities used to access the network.
        val walletDirectory = Paths.get(hfSettings.keystore)
        GatewayBuilder.wallet = Wallets.newFileSystemWallet(walletDirectory)

        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)
        val signer : String
        signer = hfSettings.defaultSigner

        // Configure the gateway connection used to access the network.
        GatewayBuilder.builder = Gateway.createBuilder()
            .identity(GatewayBuilder.wallet, signer)
            .networkConfig(connectionProfile).discovery(true)


        GatewayBuilder.builder.connect().use { gateway ->
            val network: Network = gateway.getNetwork("mychannel")
            val contract: Contract = network.getContract(hfSettings.chaincode)
            contract.addContractListener(0, eventListener())

        }

        println("Done with event init")
    }


    fun eventListener(): Consumer<ContractEvent?>? {
        println("inside event listener")
        var mapper = jacksonObjectMapper()

        return Consumer { contractEvent : ContractEvent? ->
            println("FoundEVENT: $contractEvent")
            val payload = contractEvent?.payload?.get()
            val event : SigningNotification = mapper.readValue(payload!!)
            when (contractEvent?.name) {
                "TestEvent" -> {
                    println("found event")
                    println(mapper.writeValueAsString(event))
                }
                "PostThresholdTransaction" -> {
                    val signedTxJSON = commonService.getSignOrgThreshold(event.assetKey)
                    val signedTx =  mapper.readValue<ThresholdSignedTx>(signedTxJSON)
                    Thread.sleep((500 * hfSettings.orgNum * hfSettings.orgNum).toLong());
                    val postSignature = commonService.postSignShareThreshold(event.assetKey, signedTxJSON)
                }
                "CompletePostThresholdTransaction" -> {
                    val fabricTx = mapper.readValue<FabricThresholdSigningTx>(event.asset!!)
                    println("HELPME")
                    var found = false
                    if (fabricTx.signedShares != null){
                        for (org in fabricTx.signedShares!!){
                            if (org.shareIdentifier == hfSettings.orgNum){
                                found = true
                            }
                        }
                    }

                    if (found == false){
                        val signedTxJSON = commonService.getSignOrgThreshold(event.assetKey)
                        val signedTx =  mapper.readValue<ThresholdSignedTx>(signedTxJSON)
                        Thread.sleep((500 * hfSettings.orgNum * hfSettings.orgNum).toLong());
                        val postSignature = commonService.postSignShareThreshold(event.assetKey, signedTxJSON)
                    }
                }
                "PostInitialThresholdTransaction" -> {
                    val fabricTx = mapper.readValue<FabricThresholdSigningTx>(event.asset!!)
                    var found = false
                    if (fabricTx.shareInformation != null){
                        if (fabricTx.shareInformation!!.containsKey(hfSettings.orgMSP)){
                            found = true
                        }
                    }

                    if (found == false){
                        val shareNonceTxJSON = commonService.postNonceShareSingleWallet(fabricTx)
                        println("PostInitialThresholdTransaction")
                        println(shareNonceTxJSON)
                    }

                }
                "PostTransactionNonceShare" -> {
                    val fabricTx = mapper.readValue<FabricThresholdSigningTx>(event.asset!!)
                    var found = false
                    if (fabricTx.shareInformation != null){
                        if (fabricTx.shareInformation!!.containsKey(hfSettings.orgMSP)){
                            found = true
                        }
                    }
                    if (found == false){
                        val shareNonceTxJSON = commonService.postNonceShareSingleWallet(fabricTx)
                        println("PostTransactionNonceShare")
                        println(shareNonceTxJSON)
                    }
                }
                "RequestPartialTransactionSignature" -> {
                    val fabricTx = mapper.readValue<FabricThresholdSigningTx>(event.asset!!)
                    var found = false
                    if (fabricTx.shareInformation != null){
                        if (fabricTx.shareInformation!!.containsKey(hfSettings.orgMSP)){
                            if(fabricTx.shareInformation!![hfSettings.orgMSP]!!.signedShare!!.sig != null){
                                found = true
                            }
                        }
                    }
                    if (found == false && fabricTx.approvalLevel < hfSettings.orgNum){
                        val shareSigTxJSON = commonService.postPartialSignatureSingleWallet(fabricTx.messageHash)
                        println("RequestPartialTransactionSignature")
                        println(shareSigTxJSON)
                    }
                }
                "CompletePostThresholdTransactionSharedWallet" -> {
                    val fabricTx = mapper.readValue<FabricThresholdSigningTx>(event.asset!!)
                    var found = false
                    if (fabricTx.shareInformation != null){
                        if (fabricTx.shareInformation!!.containsKey(hfSettings.orgMSP)){
                            if(fabricTx.shareInformation!![hfSettings.orgMSP]!!.signedShare!!.sig != null){
                                found = true
                            }
                        }
                    }
                    if (found == false && fabricTx.approvalLevel < hfSettings.orgNum){
                        val shareSigTxJSON = commonService.postPartialSignatureSingleWallet(fabricTx.messageHash)
                        println("Complete Post THreshold Transaction Shared Wallet")
                        println(shareSigTxJSON)
                    }
                }
                "ThresholdSigningComplete" -> {
                    println("Signing Complete")
                    // Complete logic for transaciton
                }
                "ThresholdSigningFailed" -> {
                    println("Signing Failed")
                    // Logic to retry or handle
                }
                "PostInitialECDSAThresholdTransaction" -> {
                    val fabricTx = mapper.readValue<FabricECDSAThresholdSigningTx>(event.asset!!)
                    val approvalResponse = commonService.prepareApproveECDSATx("default", event.assetKey, fabricTx.signerID, fabricTx.tokenId)
                    println("Approval Response for " + event.assetKey + ": " + approvalResponse)
                }
                "ApproveECDSATx" -> {
                    if (hfSettings.orgMSP == event.org){
                        val signaturePostResponse = commonService.performECDSARounds("default", event.assetKey)
                        println("Tx signature response: " + signaturePostResponse)
                    }
                }
                "PostECDSASignature" -> {
                    if (hfSettings.orgMSP == event.org){
                        val fabricTx = mapper.readValue<FabricECDSAThresholdSigningTx>(event.asset!!)
                        if (fabricTx.tokenId == "ETH"){
                            val txPostResponse = commonService.performETHTx("default", fabricTx.messageHash)
                            println("Tx submmitted: " + txPostResponse)
                        }
                    }
                }
                else -> println("some other event")
            }
        }
    }


}

