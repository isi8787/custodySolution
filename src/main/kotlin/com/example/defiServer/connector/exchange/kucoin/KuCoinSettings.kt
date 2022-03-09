package com.example.defiServer.connector.exchange.kucoin

import com.example.defiServer.models.BasicToken
import com.example.defiServer.models.TxResult
import com.fasterxml.jackson.annotation.JsonProperty
import kotlinx.serialization.Serializable
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class KuCoinSettings(

    @Value("\${kucoinSetting.apiKey}")
    var apiKey: String,

    @Value("\${kucoinSetting.passphrase}")
    var passphrase: String,

    @Value("\${kucoinSetting.apiSecret}")
    var apiSecret : String,

    @Value("\${kucoinSetting.partner_id}")
    var partner_id: String,

    @Value("\${kucoinSetting.partner_key}")
    var partner_key: String
)

@Serializable
data class KuCoinThirdPartyHeader(
    @JsonProperty("KC-API-PARTNER") var partnerAPIKey : String,
    @JsonProperty("KC-API-PARTNER-SIGN") var  partnerSign : String,
)

