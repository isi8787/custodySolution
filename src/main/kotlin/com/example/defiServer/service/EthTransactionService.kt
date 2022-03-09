package com.example.defiServer.service

import com.example.defiServer.impl.WebsocketExample
import com.example.defiServer.models.*
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.gitlab.blockdaemon.ubiquity.UbiquityClientBuilder
import com.gitlab.blockdaemon.ubiquity.invoker.ApiException
import com.gitlab.blockdaemon.ubiquity.model.Network
import com.gitlab.blockdaemon.ubiquity.model.Platform
import com.gitlab.blockdaemon.ubiquity.model.TxReceipt
import com.gitlab.blockdaemon.ubiquity.tx.eth.EthTransactionBuilder
import kotlinx.serialization.InternalSerializationApi
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import org.web3j.crypto.Credentials
import org.web3j.crypto.RawTransaction
import org.web3j.utils.Convert
import java.math.BigInteger

@InternalSerializationApi
@Service
class EthTransactionService (
    @Autowired private val blockDaemonSetting: BlockDaemonSettings,
        ) {

    fun ethereumsendtx(toAccount : String, quantity:String) : String {
        //val ethACCOUNT = "0xA97845d0f342218de104DC3E6712bd3D6a60AFf4"

        val authBearerToken = blockDaemonSetting.apiKey //System.getenv("AUTH_TOKEN")  // bd1bc0meoYrfp1sANZqKzB44QeWoB3LrEYKoiJfrt3N2jjH


        val client = UbiquityClientBuilder().authBearerToken(authBearerToken).build()
        val fromPrivateKey = blockDaemonSetting.ethPK  //"8086a40ea5fcb7e445a4317fd20654543840374174fa5d8a0485a68351ed5ef4"
        val credentials = Credentials.create(fromPrivateKey)

        println("Address: " + credentials.address)

        //val toPublicKey = "0x9a60fffc6dc39dceebd401ca70e3022d017baaf5"
        val value = Convert.toWei(quantity, Convert.Unit.ETHER).toBigInteger()


        val resInfura = getTxInfura()
        val infuraNonce: Int = resInfura.result.drop(2).toInt(16)
        val nonce = infuraNonce.toBigInteger()
        //val accountTxPage = client.accounts().getTxsByAddress(Platform.ETHEREUM, Network.name("ropsten"), blockDaemonSetting.ethPBK)
        //println(accountTxPage)


        // An estimate for the fee can be retrieved from the Ubiquity API
        val suggestedFee = client.transactions().estimateFee(Platform.ETHEREUM, Network.name("ropsten"), 10)
        println("SF" + suggestedFee)



        val gasprice = suggestedFee.toBigInteger()
        val gaslimit = BigInteger.valueOf(21000)

        // To prevent replay attacks the chain id should also be included
        // https://chainid.network/
        val chainId = 3

        // Signed transactions can be created using the web3j library
        val rawTransaction = RawTransaction.createEtherTransaction(nonce, gasprice, gaslimit, toAccount, value)
        //val rawTransaction = RawTransaction.createEtherTransaction(nonce,gaslimit, toPublicKey, value, gasprice, gasprice)
        // A signed ubiquity tx can the be created through the EthTransactionBuilder
        // class
        val tx = EthTransactionBuilder()
            .chainId(chainId)
            .credentials(credentials)
            .rawTransaction(rawTransaction)
            .build()

        // The signed transaction can then be sent using the ubiquity API
        println("Signed" + tx)
        // The signed transaction can then be sent using the ubiquity API

        try {
            val receipt = client.transactions().txSend(Platform.ETHEREUM, Network.name("ropsten"), tx)
            println(receipt)
            return receipt.id
        } catch (e : Exception) {
            println(e.toString())
            return "Error Executing ETH Transaction"
        }

    }

    fun getTxInfura() : InfuraCountResponse {
        val fromPrivateKey = blockDaemonSetting.ethPK
        val credentials = Credentials.create(fromPrivateKey)
        val params = arrayListOf(credentials.address, "latest")
        println(params)
        val infuraRequest : SimpleInfuraRequest = SimpleInfuraRequest(1, "2.0", "eth_getTransactionCount", params)
        println(infuraRequest)
        val mapper = jacksonObjectMapper()
        var requestAsString = mapper.writeValueAsString(infuraRequest)
        println(requestAsString)
        val txDetails = WebsocketExample.executePost(
            "https://ropsten.infura.io/v3/234372f56a5d46f5820dc9780ad2f0b0",
            requestAsString
        )
        val txFull = mapper.readValue<InfuraCountResponse>(txDetails)
        return txFull
    }

    fun btcUTXO(address : String) : String {

        val authBearerToken = blockDaemonSetting.apiKey //System.getenv("AUTH_TOKEN")  // bd1bc0meoYrfp1sANZqKzB44QeWoB3LrEYKoiJfrt3N2jjH
        println(authBearerToken)
        val client = UbiquityClientBuilder().authBearerToken(authBearerToken).build()
        val details = client.platform().getPlatform(Platform.BITCOIN, Network.TEST_NET)
        println(details)

        val utxo = client.accounts().getTxsByAddress(Platform.BITCOIN, Network.MAIN_NET, address)
        var utxostring = jacksonObjectMapper().writeValueAsString(utxo)

        return utxostring

    }


}


