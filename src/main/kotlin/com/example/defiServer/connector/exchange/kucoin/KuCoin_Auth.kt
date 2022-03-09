package com.example.defiServer.connector.exchange.kucoin

import com.example.defiServer.models.*
import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import org.apache.http.HttpHeaders
import org.apache.http.NameValuePair
import org.apache.http.client.methods.HttpGet
import org.apache.http.client.utils.URIBuilder
import org.apache.http.impl.client.HttpClients
import org.apache.http.util.EntityUtils
import org.springframework.beans.factory.annotation.Autowired
import java.time.Instant
import java.time.format.DateTimeFormatter
import java.util.*
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec
import kotlin.collections.HashMap


class KuCoin_Auth (
    @Autowired private val kucoinSetting: KuCoinSettings,
){
    fun third_party_header(timestamp : String) : HashMap<String, Any> {
        val partner_payload = timestamp + kucoinSetting.partner_id + kucoinSetting.apiKey
        val partner_signature = HMAC_SHA256_encode(kucoinSetting.partner_key, partner_payload)
        val thirdParty = HashMap<String, Any>()
        thirdParty["KC-API-PARTNER"] = kucoinSetting.partner_id
        thirdParty["KC-API-PARTNER-SIGN"]=  String(partner_signature, charset("UTF-8"))
        return thirdParty
    }

    fun HMAC_SHA256_encode(key: String, message: String): ByteArray {
        val keySpec = SecretKeySpec(
            key.toByteArray(),
            "HmacMD5"
        )
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(keySpec)
        val rawHmac = mac.doFinal(message.toByteArray())
        val base64encoded =  Base64.getEncoder().encode(rawHmac)
        return base64encoded
    }

    fun add_auth_to_params(uri: String?, parameters: HashMap<String, Any>?, method : String, partner_header: Boolean? ) : HashMap<String, Any> {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val response = HashMap<String, Any>()

        response["X-KC-API-KEY"] = kucoinSetting.apiKey
        response["KC-API-TIMESTAMP"] = timestamp
        response["KC-API-KEY-VERSION"] = "2"
        response["Content-Type"] = "application/json"
        var payload : String
        if (parameters != null){
            val mapper = jacksonObjectMapper()
            val query_string = mapper.writeValueAsString(parameters)
            payload = timestamp + method + uri + query_string
        } else {
            payload = timestamp + method + uri
        }
        val signature = HMAC_SHA256_encode(kucoinSetting.apiSecret, payload)
        response["KC-API-SIGN"] = String(signature, charset("UTF-8"))
        val passphrase = HMAC_SHA256_encode(kucoinSetting.apiSecret, kucoinSetting.passphrase)
        response["KC-API-PASSPHRASE"] = String(passphrase, charset("UTF-8"))
        if(partner_header == true){
            val headers = third_party_header(timestamp)
            response.putAll(headers)
        }
        return  response
    }
}

