package com.example.defiServer.service

import com.example.defiServer.models.*
import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import kotlinx.serialization.InternalSerializationApi
import org.apache.http.HttpHeaders
import org.apache.http.NameValuePair
import org.apache.http.client.methods.HttpGet
import org.apache.http.client.utils.URIBuilder
import org.apache.http.impl.client.HttpClients
import org.apache.http.message.BasicNameValuePair
import org.apache.http.util.EntityUtils
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import java.io.IOException
import java.net.URISyntaxException



/**
 * This example uses the Apache HTTPComponents library.
 */

@InternalSerializationApi
@Service
class CryptoDataService(
    @Autowired private val coinmarketcapSettings: CoinMarketCapSettings,
){
    fun latestQuote(tokenList : ArrayList<String>) : String {
        val uri = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
        val parameters: MutableList<NameValuePair> = ArrayList()
        val symbolList = tokenList.joinToString(separator = ",")
        parameters.add(BasicNameValuePair("symbol", symbolList))
        try {
            val result = makeAPICall(uri, parameters)
            println(result)
            return result
        } catch (e: IOException) {
            println("Error: cannont access content - $e")
            return e.toString()
        } catch (e: URISyntaxException) {
            println("Error: Invalid URL $e")
            return e.toString()
        }
    }

    fun priceConversion(accountList : ArrayList<UserAccount>) : ArrayList<UserAccount> {
        for(account in accountList){
            if (account.balance > 0) {
                val uri = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
                val parameters: MutableList<NameValuePair> = ArrayList()
                parameters.add(BasicNameValuePair("symbol", account.tokenSymbol))
                parameters.add(BasicNameValuePair("convert", "USD"))

                try {
                    val result = makeAPICall(uri, parameters)
                    val mapper = jacksonObjectMapper()
                    mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
                    val quotedata = mapper.readValue<BasicQuote>(result)
                    val rate = quotedata.data[account.tokenSymbol]!!.quote["USD"]?.price!!
                    account.usdBalance = account.balance * rate
                    account.usdBalanceHold = account.balanceOnHold * rate
                    account.percent_change_24h = quotedata.data[account.tokenSymbol]!!.quote["USD"]?.percent_change_24h
                    account.volume_24h = quotedata.data[account.tokenSymbol]!!.quote["USD"]?.volume_24h
                } catch (e: IOException) {
                    println("Error: cannont access content - $e")
                } catch (e: URISyntaxException) {
                    println("Error: Invalid URL $e")
                }


            } else {
                account.usdBalance = (0).toFloat()
            }


        }

        return accountList

    }

    @Throws(URISyntaxException::class, IOException::class)
    fun makeAPICall(uri: String?, parameters: List<NameValuePair>?): String {
        var response_content = ""
        val query = URIBuilder(uri)
        query.addParameters(parameters)
        val client = HttpClients.createDefault()
        val request = HttpGet(query.build())
        request.setHeader(HttpHeaders.ACCEPT, "application/json")
        request.addHeader("X-CMC_PRO_API_KEY", coinmarketcapSettings.apiKey)
        val response = client.execute(request)
        try {
            println(response.statusLine)
            val entity = response.entity
            response_content = EntityUtils.toString(entity)
            EntityUtils.consume(entity)
        } finally {
            response.close()
        }
        return response_content
    }
}