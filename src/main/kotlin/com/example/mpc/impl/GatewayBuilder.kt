package cbps.ecs.impl

import com.example.mpc.model.ConnectionConfig
import com.example.mpc.model.HyperledgerSettings
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory
import com.fasterxml.jackson.module.kotlin.registerKotlinModule
import kotlinx.serialization.InternalSerializationApi
import org.hyperledger.fabric.gateway.Gateway
import org.hyperledger.fabric.gateway.Wallet
import org.hyperledger.fabric.gateway.Wallets
import org.hyperledger.fabric.gateway.spi.WalletStore
import java.io.ByteArrayInputStream
import java.io.File
import java.nio.file.Paths

val LOGGER_TAG="GatewayBuilder"
// Initialized by common service
@InternalSerializationApi
object GatewayBuilder {

    lateinit var builder: Gateway.Builder
    lateinit var wallet: Wallet

    fun set(hfSettings: HyperledgerSettings) {

        val walletDirectory = Paths.get(hfSettings.keystore)
        wallet = Wallets.newFileSystemWallet(walletDirectory)

        // Read Connection details from the file
        val mapper = ObjectMapper(YAMLFactory())
        mapper.registerKotlinModule()

        var connectionConfig: ConnectionConfig =
            mapper.readValue(File(hfSettings.connectionFile), ConnectionConfig::class.java)


        // Update the config with new IP addressees
        connectionConfig.orderers.ordererName.url = "ordererIP"
        connectionConfig.peers.peer.url = "peerIP"
        connectionConfig.certificateAuthorities.ca.url = "caIP"

        // Convert object to bytes
        var connectionConfigAsString = mapper.writeValueAsBytes(connectionConfig)

        // Configure the gateway connection used to access the network.
        builder = Gateway.createBuilder()
            .identity(wallet, hfSettings.defaultSigner)
            .networkConfig(ByteArrayInputStream(connectionConfigAsString))
    }

}