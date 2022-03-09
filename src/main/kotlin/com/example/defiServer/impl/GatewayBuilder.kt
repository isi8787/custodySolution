package com.example.defiServer.impl

import com.example.defiServer.models.HyperledgerSettings
import com.example.defiServer.models.Notification
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import kotlinx.serialization.InternalSerializationApi
import org.hyperledger.fabric.gateway.*
import org.hyperledger.fabric.sdk.BlockEvent
import org.springframework.stereotype.Service
import java.nio.file.Path
import java.nio.file.Paths
import java.time.Instant
import java.util.concurrent.BlockingQueue
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit
import java.util.function.Consumer


@InternalSerializationApi
@Service
object GatewayBuilder {
    lateinit var builder: Gateway.Builder
    lateinit var wallet: Wallet
    private val blockEvents: BlockingQueue<BlockEvent> = LinkedBlockingQueue()
    private var blockListener: Consumer<BlockEvent>? = null

    fun set(hfSettings: HyperledgerSettings, user : String) {
        // Load an existing wallet holding identities used to access the network.
        val walletDirectory = Paths.get(hfSettings.keystore)
        wallet = Wallets.newFileSystemWallet(walletDirectory)
        println(hfSettings.keystore)
        println(wallet)

        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)
        val signer : String
        if (user == "default") {
            signer = hfSettings.defaultSigner
        } else {
            signer = user
        }
        // Configure the gateway connection used to access the network.
        builder = Gateway.createBuilder()
            .identity(wallet, signer)
            .networkConfig(connectionProfile).discovery(false)
    }


}