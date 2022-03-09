package com.example.defiServer.connector.exchange.binance

import com.example.defiServer.models.BasicToken
import com.example.defiServer.models.TxResult
import com.fasterxml.jackson.annotation.JsonProperty
import kotlinx.serialization.Serializable
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class BinanceSettings(

    @Value("\${binanceSetting.apiKey}")
    var apiKey: String,


    @Value("\${binanceSetting.apiSecret}")
    var apiSecret : String,
)


