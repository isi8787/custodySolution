package com.example.defiServer.service


import com.example.defiServer.impl.*
import com.example.defiServer.models.*
import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.google.common.hash.Hashing
import kotlinx.serialization.InternalSerializationApi
import org.hyperledger.fabric.gateway.Contract
import org.hyperledger.fabric.gateway.ContractException
import org.hyperledger.fabric.gateway.Network
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.context.event.ApplicationReadyEvent
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.stereotype.Service
import java.io.*
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets
import java.security.MessageDigest
import java.time.Instant
import java.time.format.DateTimeFormatter
import java.util.*
import java.util.concurrent.TimeoutException


@InternalSerializationApi
@Service
class CommonService(
    @Autowired private val hfSettings: HyperledgerSettings,
    @Autowired private val googleSettings: GoogleOAuth2,
    ) {


    fun registerUser(userId: String, userProfile : BasicProfile?, isRegistered : Boolean): String {
        var registerUser = "Successfully enrolled"
        if (!isRegistered) {
            registerUser = RegisterUser.register(userId, hfSettings)
        }
        try {
            if (registerUser.contains("Successfully enrolled") && userProfile != null) {
                userProfile.orgId = hfSettings.orgMSP
                val mapper = jacksonObjectMapper()
                val profileJSOn = mapper.writeValueAsString(userProfile)
                GatewayBuilder.set(hfSettings, "default")
                GatewayBuilder.builder.connect().use { gateway ->
                    val network: Network = gateway.getNetwork("mychannel")
                    val contract: Contract = network.getContract(hfSettings.chaincode)
                    var registerRes = contract.submitTransaction("RegisterUser", profileJSOn)
                    val stringifiedUser = String(registerRes, StandardCharsets.UTF_8)
                    return stringifiedUser
                }
            } else {
                return registerUser
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun updateUser(userId: String, userProfile : String): String {

        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                var registerRes = contract.submitTransaction("UpdateUser", userProfile)
                val stringifiedUser = String(registerRes, StandardCharsets.UTF_8)
                return stringifiedUser
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun updateUserConector(userId: String, exchangeName : String, conectorJson : String): String {

        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                var registerRes = contract.submitTransaction("AddConnector", userId, exchangeName, conectorJson)
                val stringifiedUser = String(registerRes, StandardCharsets.UTF_8)
                return stringifiedUser
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun removeConnectorAPI (userId: String, exchangeName : String, apiName : String): String {

        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                var registerRes = contract.submitTransaction("RemoveConnectorAPI", userId, exchangeName, apiName)
                val stringifiedUser = String(registerRes, StandardCharsets.UTF_8)
                return stringifiedUser
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun registerAdmin(userId:String): String {

        try {
            val registerUser = EnrollAdmin.enroll(userId, hfSettings)
            return registerUser
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun revokeUser(userId:String): String {
        val registerUser = RegisterUser.revoke(userId, hfSettings)
        return registerUser
    }

    fun reenrollUser(userId:String): String {
        return RegisterUser.reenroll(userId, hfSettings)
    }

    fun checkUser(user:GoogleProfile): Boolean {
        val checkUser = RegisterUser.checkUser(user.id, hfSettings)
        return checkUser
    }

    fun getAuthToken(token:String): String {

        try {
            val urlParameters = "code=${token}&client_id=${googleSettings.clientId}&client_secret=${googleSettings.clientSecret}&redirect_uri=http://localhost:7080/v1/login/callback&grant_type=authorization_code"
            return executeFormPost("https://oauth2.googleapis.com/token", urlParameters)
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getUserInfo(accessToken : String) : String {
        try {
            return executeGet("https://www.googleapis.com/oauth2/v1/userinfo", accessToken)
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getAccountBalance(tokenId : String, userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val inoviceRes = contract.evaluateTransaction("GetAccountBalance", tokenId, hfSettings.orgMSP, userId)
                val stringifiedInvoice = String(inoviceRes, StandardCharsets.UTF_8)
                return stringifiedInvoice
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }
    fun getUser(userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val userRes = contract.evaluateTransaction("GetUser", userId)
                val userData = String(userRes, StandardCharsets.UTF_8)
                return userData
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getUserByPhone(phone : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val userRes = contract.evaluateTransaction("GetUserByPhone", phone)
                val userJSON = String(userRes, StandardCharsets.UTF_8)
                return userJSON
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getUserByEmail(email : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val userRes = contract.evaluateTransaction("GetUserByEmail", email)
                val userJSON = String(userRes, StandardCharsets.UTF_8)
                return userJSON
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getUserAccounts(userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val inoviceRes = contract.evaluateTransaction("GetAccountsByUser", hfSettings.orgMSP, userId)
                val stringifiedInvoice = String(inoviceRes, StandardCharsets.UTF_8)
                return stringifiedInvoice
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getTokenList(userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val tokenListBytes = contract.evaluateTransaction("GetTokenList")
                val mapper = jacksonObjectMapper()
                val tokenList = mapper.readValue<Array<BasicToken>>(tokenListBytes)
                return mapper.writeValueAsString(tokenList)
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }


    fun getETHExample(): String {
        try {
            var args = arrayOf("strings")
            var ubiquityExample = UbiquityClientExample.example(args)
            return "ubiquityExample"
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getETHExample2(): String {
        try {
            var args = arrayOf("strings")
            var ubiquityExample = UbiquityClientExample2.example(args)
            return "ubiquityExample"
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun wsETHExample3(): String {
        try {
            var args = arrayOf("strings")
            var ubiquityExample = WebsocketExample.set(args)
            return "ubiquityExample"
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun mintToken(tokenId: String, quantity: Number, userId: String): String {
        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
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

    fun transferedTokenAdmin(userId : String, signedTx : FabricSigningTx) : String {
        val originalMsgBytes : ByteArray = Base64.getDecoder().decode(signedTx.message)
        val stringifiedpostfabrictx = String(originalMsgBytes, StandardCharsets.UTF_8)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.readValue<InternalTx>(stringifiedpostfabrictx)
        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                //val inoviceRes = contract.evaluateTransaction("TransferTokens", tokenId, orgId, userId, quantity.toString())
                var tokenRes = contract.submitTransaction("TransferTokens", messageAsString.tokenId, hfSettings.orgMSP, "default", messageAsString.quantity.toString())
                val stringifiedToken = String(tokenRes, StandardCharsets.UTF_8)
                return stringifiedToken
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun transferedIssueToken(user : String, tokenId: String, orgId: String, userId:String, quantity: Float): String {

        try {
            GatewayBuilder.set(hfSettings, user)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                //val inoviceRes = contract.evaluateTransaction("TransferTokens", tokenId, orgId, userId, quantity.toString())
                var tokenRes = contract.submitTransaction("TransferTokens", tokenId, orgId, userId, quantity.toString())
                val stringifiedToken = String(tokenRes, StandardCharsets.UTF_8)
                return stringifiedToken

            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postSignedTransaction(userId : String , txHash : String, signedMsg : String, signedDate : String, keyAlias : String, user_id : String, tokenId: String) : String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postsignedtx = contract.submitTransaction("PostSignature", tokenId, hfSettings.orgMSP, user_id, txHash, signedMsg, signedDate, keyAlias)
                val stringifiedpostsigntx = String(postsignedtx, StandardCharsets.UTF_8)
                return stringifiedpostsigntx

            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postSignedExternalTransaction(userId : String , txHash : String, signedMsg : String, signedDate : String, keyAlias : String, user_id : String, tokenId: String) : String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postsignedtx = contract.submitTransaction("PostExternalTransactionSignature", tokenId, hfSettings.orgMSP, user_id, txHash, signedMsg, signedDate, keyAlias)
                val stringifiedpostsigntx = String(postsignedtx, StandardCharsets.UTF_8)
                return stringifiedpostsigntx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun executeSignedTx(userId: String, signedTx : FabricSigningTx) : String {
        val originalMsgBytes : ByteArray = Base64.getDecoder().decode(signedTx.message)
        val stringifiedpostfabrictx = String(originalMsgBytes, StandardCharsets.UTF_8)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.readValue<InternalTx>(stringifiedpostfabrictx)
        try {
            GatewayBuilder.set(hfSettings, userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                //val inoviceRes = contract.evaluateTransaction("TransferTokens", tokenId, orgId, userId, quantity.toString())
                var tokenRes = contract.submitTransaction("TransferTokens", messageAsString.tokenId, messageAsString.toOrgId, messageAsString.toUser, messageAsString.quantity.toString())
                val stringifiedToken = String(tokenRes, StandardCharsets.UTF_8)
                return stringifiedToken
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }

    }

    fun executeExternalSignedTx(userId: String, signedTx : FabricSigningTx, ethService : EthTransactionService) : String {
        val originalMsgBytes : ByteArray = Base64.getDecoder().decode(signedTx.message)
        val stringifiedpostfabrictx = String(originalMsgBytes, StandardCharsets.UTF_8)
        val mapper = jacksonObjectMapper()
        var txInfo = mapper.readValue<InternalTx>(stringifiedpostfabrictx)

        val executedETH = ethService.ethereumsendtx(txInfo.toUser, txInfo.quantity.toString())
        if (!executedETH.contains("Error")) {
            try {
                GatewayBuilder.set(hfSettings,"default")
                GatewayBuilder.builder.connect().use { gateway ->
                    // Obtain a smart contract deployed on the network.
                    val network: Network = gateway.getNetwork("mychannel")
                    val contract: Contract = network.getContract(hfSettings.chaincode)
                    // Evaluate transactions that query state from the ledger.
                    var postfabrictx  = contract.submitTransaction("BurnTokens", hfSettings.orgMSP ,"default" , txInfo.tokenId, txInfo.quantity.toString() )
                    val stringifiedpostfabrictx = String(postfabrictx, StandardCharsets.UTF_8)
                    println(stringifiedpostfabrictx)
                    //return stringifiedpostfabrictx

                }
            } catch (e: ContractException) {
                return e.toString()
            } catch (e: TimeoutException) {
                return e.toString()
            } catch (e: InterruptedException) {
                return e.toString()
            }
            return executedETH
        } else {
            return "Failed to execute ETH Transaction"
        }


    }

    fun postTransaction(userHeader : String, tokenId: String, orgId: String, userId:String, quantity: Float): String {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val message : InternalTx = InternalTx(userHeader, userId, orgId, tokenId, quantity,timestamp)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.writeValueAsString(message)
        val b64encodedString: String = Base64.getEncoder().encodeToString(messageAsString.toByteArray())
        val sha256hex = Hashing.sha256()
            .hashString(b64encodedString, StandardCharsets.UTF_8)
            .asBytes()

        val md = MessageDigest.getInstance("SHA-256")
        val digest = md.digest(sha256hex)
        val hexHash : String = digest.fold("", { str, it -> str + "%02x".format(it) })

        val fabricTx = FabricTx(userHeader, hexHash, b64encodedString, timestamp, 0, tokenId)
        var fabricTxAsString = mapper.writeValueAsString(fabricTx)
        try {
            GatewayBuilder.set(hfSettings,userHeader)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postfabrictx  = contract.submitTransaction("PostTransaction", fabricTxAsString)
                val stringifiedpostfabrictx = String(postfabrictx, StandardCharsets.UTF_8)
                return stringifiedpostfabrictx

            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postThresholdTransaction(userHeader : String, tokenId: String, orgId: String, userId:String, quantity: Float): String {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val message : InternalTx = InternalTx(userHeader, userId, orgId, tokenId, quantity,timestamp)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.writeValueAsString(message)
        val b64encodedString: String = Base64.getEncoder().encodeToString(messageAsString.toByteArray())
        val sha256hex = Hashing.sha256()
            .hashString(b64encodedString, StandardCharsets.UTF_8)
            .asBytes()

        val md = MessageDigest.getInstance("SHA-256")
        val digest = md.digest(sha256hex)
        val hexHash : String = digest.fold("", { str, it -> str + "%02x".format(it) })

        val fabricTx = FabricTx(userHeader, hexHash, b64encodedString, timestamp,0, tokenId)
        var fabricTxAsString = mapper.writeValueAsString(fabricTx)
        try {
            GatewayBuilder.set(hfSettings,userHeader)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var generateKey = contract.evaluateTransaction("GenerateSharedKey", fabricTx.message)
                //val keyInfo = String(generateKey, StandardCharsets.UTF_8)
                val keyFragList = mapper.readValue<Array<String>>(generateKey)
                val keyInfoFragment = keyFragList.joinToString(",")
                var postfabrictx  = contract.submitTransaction("PostThresholdTransaction", fabricTxAsString, keyInfoFragment)
                val stringifiedpostfabrictx = String(postfabrictx, StandardCharsets.UTF_8)
                return stringifiedpostfabrictx

            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getSigningTransaction(txHash : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val txRes = contract.evaluateTransaction("QueryTransaction", txHash)
                val txJSON = String(txRes, StandardCharsets.UTF_8)
                return txJSON
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getSigningExternalTransaction(txHash : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val txRes = contract.evaluateTransaction("QueryExternalTransaction", txHash)
                val txJSON = String(txRes, StandardCharsets.UTF_8)
                return txJSON
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }


    fun createAccount(tokenId: String,  userId:String): String {
        GatewayBuilder.set(hfSettings,userId)
        try {
            val keyJson = GatewayBuilder.builder.connect().use { gateway ->
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val keyBytes = contract.evaluateTransaction("GetUserPubKey")
                String(keyBytes, StandardCharsets.UTF_8)
            }
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var accountBytes = contract.submitTransaction("CreateAccount", tokenId, hfSettings.orgMSP, userId, "fabric", keyJson)
                val accountJSON = String(accountBytes, StandardCharsets.UTF_8)
                return accountJSON
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun  externalTransfer(userId : String, tokenId : String, address : String, quantity : Float) : String {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val message : InternalTx = InternalTx(userId, address, "external", tokenId, quantity, timestamp)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.writeValueAsString(message)
        val b64encodedString: String = Base64.getEncoder().encodeToString(messageAsString.toByteArray())
        val sha256hex = Hashing.sha256()
            .hashString(b64encodedString, StandardCharsets.UTF_8)
            .asBytes()

        val md = MessageDigest.getInstance("SHA-256")
        val digest = md.digest(sha256hex)
        val hexHash : String = digest.fold("", { str, it -> str + "%02x".format(it) })


        val fabricTx = FabricTx(userId, hexHash, b64encodedString, timestamp,0, tokenId)
        val externalFabricTx = FabricExternalTx(hexHash, "Owner", fabricTx, null, null)
        var fabricTxAsString = mapper.writeValueAsString(externalFabricTx)
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postfabrictx  = contract.submitTransaction("PostExternalTransaction", fabricTxAsString)
                val stringifiedpostfabrictx = String(postfabrictx, StandardCharsets.UTF_8)
                return stringifiedpostfabrictx

            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }


    }

    fun executeFormPost(targetURL: String, urlParameters : String): String {
        var connection: HttpURLConnection? = null
        var `is`: InputStream? = null
        return try {
            val url = URL(targetURL)
            val postData: ByteArray = urlParameters.toByteArray()
            val postDataLength : Int = postData.size
            connection = url.openConnection() as HttpURLConnection
            connection.requestMethod = "POST"
            connection.setRequestProperty("Content-Type", "application/x-www-form-urlencoded")

            connection.setRequestProperty("Content-Length", Integer.toString(postDataLength))
            connection.setRequestProperty("Content-Language", "en-US")
            connection.useCaches = false
            connection.doOutput = true


            val wr = DataOutputStream(
                connection.outputStream
            )
            wr.writeBytes(urlParameters)
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

    fun executeGet(targetURL: String?, header: String?): String {
        var connection: HttpURLConnection? = null
        var `is`: InputStream? = null
        return try {
            val url = URL(targetURL)
            connection = url.openConnection() as HttpURLConnection
            connection.requestMethod = "GET"
            connection.setRequestProperty("Authorization", "Bearer "+ header)
            connection.useCaches = false
            connection.doOutput = true

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
            return "Error" // This needs the error status.
        } finally {
            connection?.disconnect()
        }
    }

    fun startConnector(userResponse : String, dockerService: DockerService) : String {
        val mapper = jacksonObjectMapper()
        var userFabricInfo = mapper.readValue<FabricUser>(userResponse)

        if (userFabricInfo.connectors != null){
            val startConnectors = dockerService.startGoCrypto(userFabricInfo)
            for ((key, value) in userFabricInfo.connectors!!) {
                for (api in value){
                    api.apiKey = "HIDE"
                    api.apiSecret = "HIDE"
                    api.passphrase = "HIDE"
                }

            }
        }

        val hideInfoJSON = mapper.writeValueAsString(userFabricInfo)
        return hideInfoJSON
    }

    @OptIn(ExperimentalStdlibApi::class)
    fun getExchangeAccounts(userId : String, userAccounts : ArrayList<UserAccount>, dataService: CryptoDataService) : MutableList<FullAccountData> {
        val convertedAccounts = dataService.priceConversion(userAccounts)

        val userInfo = getUser(userId)
        val mapper = jacksonObjectMapper()
        mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
        val fabricUser = mapper.readValue<FabricUser>(userInfo)

        var exchangeAccounts : MutableList<GoCryptoExchangeAccountResults> = mutableListOf<GoCryptoExchangeAccountResults>()
        if (fabricUser.connectors != null){
            for (connector in fabricUser.connectors){
                val url = "http://localhost/v1/getaccountinfo?assetType=spot&exchange=" + connector.key.lowercase()
                val exchangeAccountData = callGoCrypto("GET", url, userId, null)
                try{
                    val result = mapper.readValue<GoCryptoExchangeAccountResults>(exchangeAccountData)
                    if(result.accounts.size > 0){
                        exchangeAccounts.add(result)
                    }
                } catch (e: Exception) {
                    println("Exchange Error: " + exchangeAccountData)
                }

            }

        }


        var fullAccountDataList : MutableList<FullAccountData> = mutableListOf()

        for (tokenAccount in convertedAccounts){
            var fullAccountData : FullAccountData = FullAccountData(tokenAccount, hashMapOf())


            val rate = tokenAccount.usdBalance?.div(tokenAccount.balance)
            for (exchange in exchangeAccounts){
                for (currency in exchange.accounts[0].currencies){
                    if (currency.currency == tokenAccount.tokenId){
                        currency.usdBalanceHold = rate!! * currency.hold
                        currency.usdBalance = rate!! * currency.totalValue
                        fullAccountData.exchanges[exchange.exchange] = currency
                        break
                    }
                }
            }

            fullAccountDataList.add(fullAccountData)

        }

        return  fullAccountDataList

    }

    fun callGoCrypto(method: String, pathUrl : String , userId : String, bodyJSON: String?): String {
        var connection: HttpURLConnection? = null
        var `is`: InputStream? = null
        return try {
            val url = URL(pathUrl)
            connection = url.openConnection() as HttpURLConnection
            connection.requestMethod = method
            connection.setRequestProperty("Host", "gocrypto."+userId)
            connection.useCaches = false
            connection.doOutput = true
            
            if(bodyJSON != null){
                connection.setRequestProperty("Content-Type", "application/json")
                connection.setRequestProperty("Content-Language", "en-US")
                connection.setRequestProperty("Content-Length", Integer.toString(bodyJSON.toByteArray().size))
                val wr = DataOutputStream(
                    connection.outputStream
                )
                wr.writeBytes(bodyJSON)
                wr.close()
            }

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

    fun getSignOrgThreshold(valueKey : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val signedTx = contract.evaluateTransaction("TSignOrg", valueKey, hfSettings.orgMSP)
                val mapper = jacksonObjectMapper()

                val stringifiedResponse = mapper.writeValueAsString(mapper.readValue<ThresholdSignedTx>(signedTx))
                return stringifiedResponse
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postSignShareThreshold(valueKey : String, signedTx : String): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val inoviceRes = contract.submitTransaction("PostSignShare", valueKey, signedTx)
                val stringifiedResponse = String(inoviceRes, StandardCharsets.UTF_8)
                return stringifiedResponse
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun generateSingleWallet(): String {
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val walletInfo = contract.evaluateTransaction("GenerateWalletSharedKey", )
                val mapper = jacksonObjectMapper()
                val keyFragList = mapper.readValue<String>(walletInfo)
                val hashMutableMap : MutableMap<String, ByteArray> = mutableMapOf()
                hashMutableMap["asset_properties"] = keyFragList.toByteArray()
                val postfabrictx = contract.createTransaction("PostDealerKeyFragments").setTransient(hashMutableMap).submit()
                val stringifiedpostfabrictx = String(postfabrictx, StandardCharsets.UTF_8)
                return stringifiedpostfabrictx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }


    fun postThresholdTransactionSingleWallet(userHeader : String, tokenId: String, orgId: String, userId:String, quantity: Float, approvalLevel : Int): String {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val message : InternalTx = InternalTx(userHeader, userId, orgId, tokenId, quantity,timestamp)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.writeValueAsString(message)
        val b64encodedString: String = Base64.getEncoder().encodeToString(messageAsString.toByteArray())
        val sha256hex = Hashing.sha256()
            .hashString(b64encodedString, StandardCharsets.UTF_8)
            .asBytes()

        val md = MessageDigest.getInstance("SHA-256")
        val digest = md.digest(sha256hex)
        val hexHash : String = digest.fold("", { str, it -> str + "%02x".format(it) })

        val fabricTx = FabricTx(userHeader, hexHash, b64encodedString, timestamp, approvalLevel, tokenId)
        var fabricTxAsString = mapper.writeValueAsString(fabricTx)
        try {
            GatewayBuilder.set(hfSettings,userHeader)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postInitialThresholdTransaction = contract.submitTransaction("PostInitialThresholdTransaction", fabricTxAsString)
                var stringifiedpostfabrictx = String(postInitialThresholdTransaction, StandardCharsets.UTF_8)

                var generateKey = contract.evaluateTransaction("GenerateSharedNonce", fabricTx.message, hfSettings.orgMSP)
                //val keyInfo = String(generateKey, StandardCharsets.UTF_8)
                val keyFragList = mapper.readValue<Array<String>>(generateKey)
                //val keyInfoFragment = keyFragList.joinToString(",")

                var postfabrictx  = contract.submitTransaction("PostTransactionSharedNonce", fabricTx.messageHash, hfSettings.orgMSP, keyFragList[0], keyFragList[1])
                val tx = String(postfabrictx, StandardCharsets.UTF_8)
                return tx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postNonceShareSingleWallet(fabricTx: FabricThresholdSigningTx): String {
        var mapper = jacksonObjectMapper()
        var fabricTxAsString = mapper.writeValueAsString(fabricTx)
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                var generateKey = contract.evaluateTransaction("GenerateSharedNonce", fabricTx.message, hfSettings.orgMSP)
                //val keyInfo = String(generateKey, StandardCharsets.UTF_8)
                val keyFragList = mapper.readValue<Array<String>>(generateKey)
                var postfabrictx  = contract.submitTransaction("PostTransactionSharedNonce", fabricTx.messageHash, hfSettings.orgMSP, keyFragList[0], keyFragList[1])
                val tx = String(postfabrictx, StandardCharsets.UTF_8)
                return tx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postPartialSignatureSingleWallet(txHash: String): String {
        var mapper = jacksonObjectMapper()
        try {
            GatewayBuilder.set(hfSettings,"default")
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                var partialSignatureBytes = contract.evaluateTransaction("TSignOrgSharedWallet", txHash, hfSettings.orgMSP)
                var partialSignatureJSON = String(partialSignatureBytes, StandardCharsets.UTF_8)
                val partialSignature =  mapper.readValue<ThresholdSignedTx>(partialSignatureJSON)
                val partialSignedTx = contract.submitTransaction("PostSignShareWallet", txHash, partialSignatureJSON, hfSettings.orgMSP)
                val tx = String(partialSignedTx, StandardCharsets.UTF_8)
                return tx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getPendingThresholdTxs(userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val inoviceRes = contract.evaluateTransaction("GetAllPendingTxs", )
                val stringifiedInvoice = String(inoviceRes, StandardCharsets.UTF_8)
                return stringifiedInvoice
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getAllThresholdTxs(userId : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val inoviceRes = contract.submitTransaction("GetAllThresholdTxs", )
                val stringifiedInvoice = String(inoviceRes, StandardCharsets.UTF_8)
                return stringifiedInvoice
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun queryECDSAThresholdTransaction(userId : String, txHash : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val txRes = contract.submitTransaction("QueryECDSAThresholdTransaction", txHash )
                val stringifiedTx = String(txRes, StandardCharsets.UTF_8)
                return stringifiedTx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun generateECDSAWalletSharedKey(userId : String, tokenId: String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val walletFrags = contract.evaluateTransaction("GenerateECDSAWalletSharedKey", userId, tokenId)
                val keys = String(walletFrags, StandardCharsets.UTF_8)
                return keys
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }


    fun postECDSAWalletSharedKey(userId : String, fragmentInfo : String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                //val b64encodedString: ByteArray = Base64.getEncoder().encode(fragmentInfo.toByteArray())
                val hashMutableMap : MutableMap<String, ByteArray> = mutableMapOf()
                hashMutableMap["asset_properties"] = fragmentInfo.toByteArray()
                val transientResponse = contract.createTransaction("PostECDSAWalletSharedKey").setTransient(hashMutableMap).submit()
                val transientResponseString = String(transientResponse, StandardCharsets.UTF_8)
                return transientResponseString
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun postECDSAThresholdTransaction(userHeader : String, tokenId: String, orgId: String, userId:String, quantity: Float, approvalLevel : Int): String {
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        val message : InternalTx = InternalTx(userHeader, userId, orgId, tokenId, quantity,timestamp)
        val mapper = jacksonObjectMapper()
        var messageAsString = mapper.writeValueAsString(message)
        val b64encodedString: String = Base64.getEncoder().encodeToString(messageAsString.toByteArray())
        val sha256hex = Hashing.sha256()
            .hashString(b64encodedString, StandardCharsets.UTF_8)
            .asBytes()

        val md = MessageDigest.getInstance("SHA-256")
        val digest = md.digest(sha256hex)
        val hexHash : String = digest.fold("", { str, it -> str + "%02x".format(it) })

        val fabricTx = FabricTx(userHeader, hexHash, b64encodedString, timestamp, approvalLevel, tokenId)
        var fabricTxAsString = mapper.writeValueAsString(fabricTx)
        try {
            GatewayBuilder.set(hfSettings,userHeader)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var postInitialThresholdTransaction = contract.submitTransaction("PostInitialECDSAThresholdTransaction", fabricTxAsString)
                var tx = String(postInitialThresholdTransaction, StandardCharsets.UTF_8)
                return tx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }


    fun postETHECDSAThresholdTransaction(userHeader : String, tokenId: String, orgId: String, destination :String, quantity: Float, approvalLevel : Int): String {
        var mapper = jacksonObjectMapper()
        val timestamp = DateTimeFormatter.ISO_INSTANT.format(Instant.now())
        try {
            GatewayBuilder.set(hfSettings,userHeader)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                var ethTxData = contract.evaluateTransaction("PrepareEthTx", userHeader, tokenId, destination, quantity.toString())
                var ethTx = mapper.readValue<ETHTX>(ethTxData)
                var ethTxHash = mapper.readValue<String>(ethTx.hash)
                val fabricTx = FabricTx(userHeader, ethTxHash, ethTx.tx, timestamp, approvalLevel, tokenId)
                var fabricTxAsString = mapper.writeValueAsString(fabricTx)
                var postInitialThresholdTransaction = contract.submitTransaction("PostInitialECDSAThresholdTransaction", fabricTxAsString)
                var tx = String(postInitialThresholdTransaction, StandardCharsets.UTF_8)
                return tx
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun prepareApproveECDSATx(userId : String, txHash : String, targetUser : String, tokenId: String): String {
        var mapper = jacksonObjectMapper()
        val approvalECDSA = ApprovalTECDSA(true, null, txHash, null, null, targetUser, tokenId)
        var approvalECDSAAsString = mapper.writeValueAsString(approvalECDSA)
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                val hashMutableMap : MutableMap<String, ByteArray> = mutableMapOf()
                hashMutableMap["asset_properties"] = approvalECDSAAsString.toByteArray()
                val postfabrictx = contract.createTransaction("PrepareApproveECDSATx").setTransient(hashMutableMap).evaluate()
                val mapper = jacksonObjectMapper()
                val approvalResponse = mapper.readValue<String>(postfabrictx)
                val hashMutableMap2 : MutableMap<String, ByteArray> = mutableMapOf()
                hashMutableMap2["asset_properties"] = approvalResponse.toByteArray()
                val postfabrictx2 = contract.createTransaction("ApproveECDSATx").setTransient(hashMutableMap2).submit()
                val approvalResponse2 = String(postfabrictx, StandardCharsets.UTF_8)
                return approvalResponse2
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun performECDSARounds(userId : String, txHash: String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val signature = contract.evaluateTransaction("PerformECDSARounds", txHash)
                val mapper = jacksonObjectMapper()
                val signatureResponse = mapper.readValue<String>(signature)
                if (signatureResponse != null && signatureResponse != "Not Verified"){
                    val postsignature = contract.submitTransaction("PostECDSASignature", txHash, signatureResponse)
                    val postsignatureResponse = String(postsignature, StandardCharsets.UTF_8)
                    return postsignatureResponse
                }
                return signatureResponse
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun getWalletData(userId : String, tokenId: String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val walletFrags = contract.evaluateTransaction("GetWalletId", userId, tokenId)
                val keys = String(walletFrags, StandardCharsets.UTF_8)
                return keys
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

    fun performETHTx(userId : String, txHash: String): String {
        try {
            GatewayBuilder.set(hfSettings,userId)
            GatewayBuilder.builder.connect().use { gateway ->
                // Obtain a smart contract deployed on the network.
                val network: Network = gateway.getNetwork("mychannel")
                val contract: Contract = network.getContract(hfSettings.chaincode)
                // Evaluate transactions that query state from the ledger.
                val transactionReceipt = contract.evaluateTransaction("SubmitEthTx", txHash)
                val txReceipt = String(transactionReceipt, StandardCharsets.UTF_8)
                val fabrictransactionReceipt = contract.submitTransaction("UpdateTxReceipt", txHash, txReceipt)
                return  String(fabrictransactionReceipt, StandardCharsets.UTF_8)
            }
        } catch (e: ContractException) {
            return e.toString()
        } catch (e: TimeoutException) {
            return e.toString()
        } catch (e: InterruptedException) {
            return e.toString()
        }
    }

}






