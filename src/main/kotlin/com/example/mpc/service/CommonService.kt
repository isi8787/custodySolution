package com.example.mpc.service


import cbps.ecs.impl.GatewayBuilder
import com.example.mpc.model.HyperledgerSettings
import com.example.mpc.model.TestName

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import kotlinx.serialization.InternalSerializationApi
import org.hyperledger.fabric.gateway.ContractException
import org.json.JSONObject
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Service
import java.io.*
import java.lang.Integer.parseInt
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets
import java.nio.file.Path
import java.nio.file.Paths
import java.security.SecureRandom
import java.security.cert.X509Certificate
import java.text.SimpleDateFormat
import java.time.LocalDateTime
import java.util.*
import java.util.concurrent.TimeoutException
import javax.net.ssl.*

import khttp.post
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.decodeFromJsonElement
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity


@InternalSerializationApi
@Service
class CommonService(

    @Autowired private val hfSettings: HyperledgerSettings,
) {

    fun getSupplier(id: String): TestName {

        var supplierAsJSON: TestName

        GatewayBuilder.builder.connect().use { gateway ->
            // Obtain a smart contract deployed on the network.
            val network = gateway.getNetwork(hfSettings.channel)
            val contract = network.getContract(hfSettings.chaincode)
            // Evaluate transactions that query state from the ledger.
            val supplierAsByte = contract.evaluateTransaction("readSupplier", id)
            val stringiFiedSupplier = String(supplierAsByte, StandardCharsets.UTF_8)


            val format = Json {
                isLenient = true
                ignoreUnknownKeys = true
            }
            supplierAsJSON = format.decodeFromString<TestName>(stringiFiedSupplier)


            return supplierAsJSON
        }

    }


    fun updateSupplier(
        supplier: TestName): TestName {
        var supplierAsJSON: TestName

        GatewayBuilder.builder.connect().use { gateway ->
            // Obtain a smart contract deployed on the network.
            val network = gateway.getNetwork(hfSettings.channel)
            val contract = network.getContract(hfSettings.chaincode)
            // Evaluate transactions that query state from the ledger.
            val mapper = jacksonObjectMapper()
            var supplierAsString = mapper.writeValueAsString(supplier)
            var supplierAsBytes = contract.createTransaction("updateSupplier")
                .submit(supplierAsString)

            supplierAsString = String(supplierAsBytes, StandardCharsets.UTF_8)

            val format = Json {
                isLenient = true
                ignoreUnknownKeys = true
            }
            supplierAsJSON = format.decodeFromString<TestName>(supplierAsString)


            return supplierAsJSON
        }

    }

}






