package com.example.defiServer.models

import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class CoinMarketCapSettings(

    @Value("\${coinmarketcapSetting.apiKey}")
    var apiKey: String

)