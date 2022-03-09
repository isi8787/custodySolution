package com.example.defiServer.connector.exchange.binance

import com.binance.api.client.BinanceApiClientFactory
import com.binance.api.client.domain.event.AggTradeEvent
import com.binance.api.client.domain.event.CandlestickEvent
import com.binance.api.client.domain.event.DepthEvent
import com.binance.api.client.domain.event.TickerEvent
import com.binance.api.client.domain.general.ExchangeInfo
import com.binance.api.client.domain.market.CandlestickInterval
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import kotlinx.serialization.InternalSerializationApi
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import java.io.IOException


@InternalSerializationApi
@Service
class BinanceService(
    @Autowired private val kucoinSetting : BinanceSettings
) {

    fun check() : String {

        val binanceRESTClient = BinanceApiClientFactory.newInstance().newRestClient()

        val response: ExchangeInfo? = binanceRESTClient.exchangeInfo
        if(response != null){
            return "Active"
        }
        return "Inactive"
    }

    fun set() {

        val binanceWSClient = BinanceApiClientFactory.newInstance().newWebSocketClient()


        binanceWSClient.onDepthEvent("ethbtc") { response: DepthEvent? ->
            val mapper = jacksonObjectMapper()
            println(
                "BINANCE: " + mapper.writeValueAsString((response))
            )
        }


        binanceWSClient.onTickerEvent("ethbtc") { response: TickerEvent? ->
            val mapper = jacksonObjectMapper()
            println(
                "BINANCE: " + mapper.writeValueAsString((response))
            )
        }



    }

}