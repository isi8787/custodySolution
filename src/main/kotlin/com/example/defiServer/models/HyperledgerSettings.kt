package com.example.defiServer.models

import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class HyperledgerSettings(

    @Value("\${hyperledger-fabric.keystore}")
    var keystore: String,

    @Value("\${hyperledger-fabric.connectionFile}")
    var connectionFile: String,

    @Value("\${hyperledger-fabric.defaultSigner}")
    var defaultSigner: String,

    @Value("\${hyperledger-fabric.channel}")
    var channel: String,

    @Value("\${hyperledger-fabric.chaincode}")
    var chaincode: String,

    @Value("\${hyperledger-fabric.ca}")
    var ca: String,

    @Value("\${hyperledger-fabric.pemFile}")
    var pemFile : String,

    @Value("\${hyperledger-fabric.affiliation}")
    var affiliation : String,

    @Value("\${hyperledger-fabric.orgMSP}")
    var orgMSP : String,

    @Value("\${hyperledger-fabric.orgNum}")
    var orgNum : Int
)

@Component
@Service
data class GoogleOAuth2(
    @Value("\${google-settings.clientId}")
    var clientId : String,

    @Value("\${google-settings.clientSecret}")
    var clientSecret : String
)

@Component
@Service
data class ExchangeList(
    @Value("\${exchanges.exchangeList}")
    var exchangeList:  MutableList<String>
)