package com.example.defiServer.impl

import com.example.defiServer.models.InfuraResponse
import com.example.defiServer.models.SimpleInfuraRequest
import com.example.defiServer.service.CommonService
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.gitlab.blockdaemon.ubiquity.UbiquityClientBuilder
import com.gitlab.blockdaemon.ubiquity.model.*
import com.gitlab.blockdaemon.ubiquity.ws.Notification.*
import com.gitlab.blockdaemon.ubiquity.ws.Subscription
import com.gitlab.blockdaemon.ubiquity.ws.WebsocketClient
import kotlinx.serialization.InternalSerializationApi
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
import org.hyperledger.fabric.gateway.Contract
import org.hyperledger.fabric.gateway.ContractException
import java.io.*
import java.net.HttpURLConnection
import java.net.URL
import java.nio.ByteBuffer
import java.nio.CharBuffer
import java.nio.charset.Charset
import java.nio.charset.StandardCharsets
import java.util.*
import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeoutException
import java.util.function.BiConsumer


object WebsocketExample {

    private const val ETH_ACCOUNT = "0xA97845d0f342218de104DC3E6712bd3D6a60AFf4"

    // Consumer used to handle the message recieves
    private val txConsumer =
        BiConsumer { client: WebsocketClient,  notification : TxNotification ->
            println("Got a TX notification:: " + notification.content.id)
        }

    private val blockConsumer =
        BiConsumer { client: WebsocketClient, notification: BlockNotification ->
            println("Got a block notification:: " + notification.content.id)
            // Unsubscribe from this subscription
        }

    private val blockIdentifierConsumer =
        BiConsumer { client: WebsocketClient?, notification: BlockIdentifierNotification ->
            println(
                "Got a block identifier notification:: " + notification.content.id
            )
            val blockInfo  = UbiquityClientGetTx.clientTx(notification.content.id!!)
            for (tx in blockInfo.txs!!) {
                for (account in tx.addresses!!){
                    if (account ==  ETH_ACCOUNT){
                        println("Found a TX involving our account:")
                        if(tx.operations!!.containsKey("native")){
                            val nativeOp: Operation? = tx.operations!!["native"]
                            if(nativeOp!!.transferOperation.detail.to == ETH_ACCOUNT){
                                println("This need to go to our ledger")
                                val fullTx = tx.id?.let { getTxInfura(it) }
                                val data = fullTx!!.result.input
                                println("This is the HEX Data:" + data)

                                val hex = data.drop(2)

                                val buff: ByteBuffer = ByteBuffer.allocate(data.length / 2)
                                var i = 0
                                while (i < hex.length) {
                                    buff.put(hex.substring(i, i + 2).toInt(16).toByte())
                                    i += 2
                                }
                                buff.rewind()
                                val cs: Charset = Charset.forName("UTF-8") // BBB

                                val cb: CharBuffer = cs.decode(buff) // BBB
                                println("String DAta:" + cb.toString())

                                val mintTokentoFab = mintToken("test1", 1)
                                println("For now Minting Manually 1 but it should be: " + fullTx.result.value)

                            }

                        }
                    }
                }
            }
        }

    @Throws(InterruptedException::class, ExecutionException::class)
    @JvmStatic
    fun set(args: Array<String>) {
        val authBearerToken = System.getenv("AUTH_TOKEN")
        val client = UbiquityClientBuilder()
            .authBearerToken(authBearerToken)
            .build()
        client.ws(Platform.ETHEREUM, Network.name("ropsten")).use { wsClient ->

            // Create a subscription to a particular channel.
            val txSubscription =
                Subscription.tx(
                    Collections.singletonMap<String, Any>(
                        "addresses",
                        listOf("0xa97845d0f342218de104dc3e6712bd3d6a60aff4")
                    ), txConsumer
                )
            wsClient.subscribe(txSubscription).get()
            val blocksubscription =
                Subscription.blocks(blockConsumer)
            wsClient.subscribe(blocksubscription).get()
            val blockIdentifierSubscription =
                Subscription.blockIdentifiers(blockIdentifierConsumer)
            wsClient.subscribe(blockIdentifierSubscription).get()

            // Wait for the server to send some new tx to the specified consumer.
            Thread.sleep(140000L)

        }
        // closing the socket gracefully may take a few seconds
        // during the graceful shutdown active subscription handlers may receive notification
    }

    fun getTxInfura(txHash : String): InfuraResponse {
        val infuraRequest : SimpleInfuraRequest = SimpleInfuraRequest(1337, "2.0", "eth_getTransactionByHash", arrayListOf(txHash))
        val mapper = jacksonObjectMapper()
        var requestAsString = mapper.writeValueAsString(infuraRequest)
        val txDetails = executePost("https://ropsten.infura.io/v3/234372f56a5d46f5820dc9780ad2f0b0", requestAsString)
        println(txDetails)
        val txFull = Json.decodeFromString<InfuraResponse>(txDetails)

        return txFull
    }

    @OptIn(InternalSerializationApi::class)
    fun mintToken(tokenId: String, quantity: Number): String {
        try {
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: org.hyperledger.fabric.gateway.Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract("fiat")
                // Evaluate transactions that query state from the ledger.
                //val tokenRes = contract.evaluateTransaction("IssueTokens", tokenId, quantity.toString())
                var tokenRes = contract.submitTransaction("IssueTokens",tokenId, quantity.toString())
                val stringifiedToken = String(tokenRes, StandardCharsets.UTF_8)
                return stringifiedToken
            }
        } catch (e: ContractException) {
            val ex = e.proposalResponses
            println(ex.first().proposalResponse.toString())
            println(ex.last().proposalResponse.toString())
            return ex.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun executePost(targetURL: String?, requestJSON: String): String {
        var connection: HttpURLConnection? = null
        var `is`: InputStream? = null
        return try {
            val url = URL(targetURL)
            connection = url.openConnection() as HttpURLConnection
            connection.requestMethod = "POST"
            connection.setRequestProperty("Content-Type", "application/json")

            connection.setRequestProperty("Content-Length", Integer.toString(requestJSON.toByteArray().size))
            connection.setRequestProperty("Content-Language", "en-US")
            connection.useCaches = false
            connection.doOutput = true


            val wr = DataOutputStream(
                connection.outputStream
            )
            wr.writeBytes(requestJSON)
            wr.close()

            //Get Response
            try {
                `is` = connection.inputStream
            } catch (ioe: IOException) {
                val httpConn = connection
                val statusCode = httpConn.responseCode
                if (statusCode != 200) {
                    `is` = httpConn.errorStream
                }
            }
            val rd = BufferedReader(InputStreamReader(`is`))
            val response = StringBuilder() // or StringBuffer if Java version 5+
            var line: String?
            while (rd.readLine().also { line = it } != null) {
                response.append(line)
                response.append('\r')
            }
            rd.close()
            response.toString()
        } catch (e: Exception) {
            return e.toString()
        } finally {
            connection?.disconnect()
        }

    }

}

object UbiquityClientGetTx {
    @Throws(Exception::class)
    @JvmStatic
    fun clientTx(blockHash : String) : Block {
        val authBearerToken = System.getenv("AUTH_TOKEN")
        val client = UbiquityClientBuilder().authBearerToken(authBearerToken).build()
        val block = client.block().getBlock(Platform.ETHEREUM, Network.name("ropsten"), blockHash)
        return block
    }
}



