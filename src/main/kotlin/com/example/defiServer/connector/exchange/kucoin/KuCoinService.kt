package com.example.defiServer.connector.exchange.kucoin

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.kucoin.sdk.KucoinClientBuilder
import com.kucoin.sdk.KucoinPrivateWSClient
import com.kucoin.sdk.KucoinPublicWSClient
import com.kucoin.sdk.KucoinRestClient
import com.kucoin.sdk.rest.response.TickerResponse
import com.kucoin.sdk.websocket.event.KucoinEvent
import com.kucoin.sdk.websocket.event.Level2ChangeEvent
import com.kucoin.sdk.websocket.event.TickerChangeEvent
import kotlinx.serialization.InternalSerializationApi
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import java.io.IOException


@InternalSerializationApi
@Service
class KuCoinService(
    @Autowired private val kucoinSetting : KuCoinSettings
) {
    fun check() : String{
        val builder: KucoinClientBuilder = KucoinClientBuilder().withBaseUrl("https://openapi-v2.kucoin.com")
        val kucoinRestClient: KucoinRestClient = builder.buildRestClient()
        val result: TickerResponse? = kucoinRestClient.symbolAPI().getTicker("BTC")
        if (result != null){
            return "Active"
        }
        return "Inactive"
    }
    fun set() {

        val builder: KucoinClientBuilder = KucoinClientBuilder().withBaseUrl("https://openapi-v2.kucoin.com")
            .withApiKey(kucoinSetting.apiKey, kucoinSetting.apiSecret ,kucoinSetting.passphrase)
        val kucoinPrivateWSClient: KucoinPrivateWSClient = builder.buildPrivateWSClient()
        val kucoinPublicWSClient: KucoinPublicWSClient = builder.buildPublicWSClient()



        /*kucoinPublicWSClient.onTicker({ response: KucoinEvent<TickerChangeEvent?>? ->
            kucoinPublicWSClient.onTicker({ response2: KucoinEvent<TickerChangeEvent?>? ->
                println(
                    "PRICE DIFFERENCE: " + (response!!.data!!.price - response2!!.data!!.price)
                )
                var mapper = jacksonObjectMapper()
                println(mapper.writeValueAsString(response!!.data))
                println(mapper.writeValueAsString(response2!!.data))
            }, "KCS-BTC")
        }, "ETH-BTC")*/

        /*

        binanceWSClient.onDepthEvent("ethbtc"){ response : DepthEvent? ->
                val mapper = jacksonObjectMapper()
                println(
                    "BINANCE: " + mapper.writeValueAsString((response))
                )
        }*/



        kucoinPublicWSClient.onTicker({ response: KucoinEvent<TickerChangeEvent?>? ->
            val mapper = jacksonObjectMapper()
            println(
                "KUCOIN: " + mapper.writeValueAsString((response))
            )
        }, "ETH-BTC")


        /*kucoinPublicWSClient.onMatchExecutionData({ response: KucoinEvent<MatchExcutionChangeEvent?>? ->
            println(
                response
            )
        })
        kucoinPublicWSClient.onLevel3Data({ response: KucoinEvent<Level3ChangeEvent?>? ->
            println(
                response
            )
        }, "ETH-BTC", "KCS-BTC")*/

    }

}