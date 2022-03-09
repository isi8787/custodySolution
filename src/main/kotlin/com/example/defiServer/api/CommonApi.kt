package com.example.defiServer.api


import com.example.defiServer.connector.cryptogalaxy.CryptogalaxyWebsocketService
import com.example.defiServer.connector.exchange.binance.BinanceService
import com.example.defiServer.connector.exchange.binance.BinanceSettings
import com.example.defiServer.connector.exchange.kucoin.KuCoinService
import com.example.defiServer.models.*
import com.example.defiServer.service.EthTransactionService
import com.example.defiServer.service.CommonService
import com.example.defiServer.service.CryptoDataService
import com.example.defiServer.service.DockerService
import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.SerializationFeature
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.github.dockerjava.api.command.CreateContainerResponse
import com.github.dockerjava.api.model.Container
import com.gitlab.blockdaemon.ubiquity.model.TxReceipt
import kotlinx.serialization.InternalSerializationApi
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.*
import java.io.File


@InternalSerializationApi
@RestController
@CrossOrigin
@Validated
@RequestMapping("\${api.base-path:/v1}")
class CommonApiController(
    val commonService: CommonService,
    val ethService : EthTransactionService,
    val dataService : CryptoDataService,
    val kucoinService : KuCoinService,
    val binanceService: BinanceService,
    val cryptogalaxyService : CryptogalaxyWebsocketService,
    val dockerService : DockerService,
    @Autowired private val exchangeList : ExchangeList
) {


    @InternalSerializationApi
    @RequestMapping(
        value = ["/login/callback"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun loginCallback(
        @RequestParam() code : String
    ): ResponseEntity<String> {
        val response = commonService.getAuthToken(code)
        val mapper = jacksonObjectMapper()
        val tokenInfo = mapper.readValue<GoogleTokenResponse>(response)
        println(tokenInfo)
        val userInfo = commonService.getUserInfo(tokenInfo.access_token)
        return ResponseEntity(userInfo, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/login/getUser"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun loginGetUser(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        // Check if user is registered already if not register new user
        if (checkUser) {
            var userResponse = commonService.getUser(user.id)
            if (userResponse.contains("does not exist")){
                val basicProfile = BasicProfile(user.id,user.email,user.verified_email, user.given_name, user.family_name, "", "")
                var userResponse = commonService.registerUser(user.id, basicProfile,true)
                val hideInfoJSON = commonService.startConnector(userResponse, dockerService)
                return ResponseEntity(hideInfoJSON, HttpHeaders(), HttpStatus.OK)
            }
            val hideInfoJSON = commonService.startConnector(userResponse, dockerService)
            return ResponseEntity(hideInfoJSON, HttpHeaders(), HttpStatus.OK)
        }
        val basicProfile = BasicProfile(user.id,user.email,user.verified_email, user.given_name, user.family_name, "", "")
        val userResponse = commonService.registerUser(user.id, basicProfile, false)
        val hideInfoJSON = commonService.startConnector(userResponse, dockerService)

        return ResponseEntity(hideInfoJSON, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/updateUser"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun updateUser(
        @RequestBody userProfile: BasicProfile,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        val userProfileJSON = mapper.writeValueAsString(userProfile)
        // Check if user is registered already if not register new user
        if (!checkUser) {
            return ResponseEntity("User is not registered", HttpHeaders(), HttpStatus.OK)
        } else {
            var userResponse = commonService.updateUser(user.id, userProfileJSON)
            return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
        }
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/updateUser/connector/{exchangeName}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun updateUserConnector(
        @PathVariable("exchangeName") exchangeName: String,
        @RequestBody connectorDetails : ConnectorDetails,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        val conectorJSON = mapper.writeValueAsString(connectorDetails)
        // Check if user is registered already if not register new user
        if (!checkUser) {
            return ResponseEntity("User is not registered", HttpHeaders(), HttpStatus.OK)
        } else {
            var userResponse = commonService.updateUserConector(user.id, exchangeName, conectorJSON)
            return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
        }
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/updateUser/connector/{exchangeName}"],
        produces = ["application/json"],
        method = [RequestMethod.DELETE]
    )
    fun removeConnectorAPI(
        @PathVariable("exchangeName") exchangeName: String,
        @RequestParam("apiName") apiName: String,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        // Check if user is registered already if not register new user
        if (!checkUser) {
            return ResponseEntity("User is not registered", HttpHeaders(), HttpStatus.OK)
        } else {
            var userResponse = commonService.removeConnectorAPI(user.id, exchangeName, apiName)
            return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
        }
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/registerUser/{userId}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun registerUser(
        @PathVariable("userId") userId: String
    ): ResponseEntity<String> {
        var userResponse = commonService.registerUser(userId,null, false)
        return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/checkUser"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun checkUser(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<Boolean> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        return ResponseEntity(checkUser, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/registerAdmin/{userId}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun registerAdmin(
        @PathVariable("userId") userId: String
    ): ResponseEntity<String> {
        var userResponse = commonService.registerAdmin(userId)
        return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/revokeUser/{userId}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun revokeUser(
        @PathVariable("userId") userId: String
    ): ResponseEntity<String> {
        var userResponse = commonService.revokeUser(userId)
        return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/reenrollUser/{userId}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun reenrollUser(
        @PathVariable("userId") userId: String
    ): ResponseEntity<String> {
        var userResponse = commonService.reenrollUser(userId)
        return ResponseEntity(userResponse, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getAccountBalance/{tokenId}"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getAccountBalance(
        @PathVariable("tokenId") tokenId: kotlin.String,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var accountBalance = commonService.getAccountBalance(tokenId, user.id)
        return ResponseEntity(accountBalance, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/testETH"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getETHExample(): ResponseEntity<String> {
        println("HELPME")
        var assetList = commonService.getETHExample()
        return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/testETH2"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getETHExample2(): ResponseEntity<String> {
        println("HELPME")
        var assetList = commonService.getETHExample2()
        return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/sendETH/{account}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun getETHExample3(
        @PathVariable("account") account : String,
        @PathVariable("quantity") quantity : String
    ): ResponseEntity<String> {
        println("HELPME")
        var sendResult = ethService.ethereumsendtx(account,quantity)
        return ResponseEntity(sendResult, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getBTCTxs/{address}"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getBTCTXS(
        @PathVariable("address") address : String
    ): ResponseEntity<String> {
        println("HELPME")
        var sendResult = ethService.btcUTXO(address)
        return ResponseEntity(sendResult, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/wsTest"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getWS(): ResponseEntity<String> {
        println("HELPME")
        var assetList = commonService.wsETHExample3()
        return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/kuCoinTest"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getKuCoinWS(): ResponseEntity<String> {
        println("HELPME")
        var assetList = kucoinService.set()
        return ResponseEntity("TEST", HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getListExchanges"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getListExchanges(): ResponseEntity<List<String>> {
        return ResponseEntity(exchangeList.exchangeList, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/cryptogalaxyTest"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getcryptogalaxyWS(): ResponseEntity<String> {
        println("HELPME")
        var assetList = cryptogalaxyService.set()
        return ResponseEntity("TEST", HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/mintToken/{tokenId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun mintToken(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        var assetList = commonService.mintToken(tokenId, quantity, user.id)
        return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/transferToken/{tokenId}/{orgId}/{userId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun mintToken(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("orgId") orgId : String,
        @PathVariable("userId") userId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        if (quantity.toFloat() < 100){
            var assetList = commonService.transferedIssueToken(user.id, tokenId, orgId, userId, quantity)
            return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
        } else {
            var postTransaction = commonService.postTransaction(user.id, tokenId, orgId, userId, quantity)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        }

    }

    @RequestMapping(
        value = ["/transferToken/{tokenId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun transferToken(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String,
        @RequestParam() phone : String?,
        @RequestParam() email : String?
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        lateinit var userProfile : String
        if(phone != null){
            userProfile = commonService.getUserByPhone(phone)
        } else if (email != null) {
            userProfile = commonService.getUserByEmail(email)
        }

        lateinit var targetUser : ArrayList<BasicProfile>
        if(userProfile != null) {
            val mapper = jacksonObjectMapper()
            targetUser = mapper.readValue<ArrayList<BasicProfile>>(userProfile)
        } else {
            return ResponseEntity("User Not Found", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        if (quantity.toFloat() < 100){
            var assetList = commonService.transferedIssueToken(user.id, tokenId, targetUser[0].orgId!! , targetUser[0].id, quantity)
            return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
        } else {
            var postTransaction = commonService.postTransaction(user.id, tokenId, targetUser[0].orgId!! , targetUser[0].id, quantity)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        }

    }

    @RequestMapping(
        value = ["/externalTransferToken/{tokenId}/{quantity}/{address}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun externalTransferToken(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("quantity") quantity : Float,
        @PathVariable("address") address : String,
        @RequestHeader("Authorization") token : String,
        @RequestParam() phone : String?,
        @RequestParam() email : String?
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        var externalTransfer = commonService.externalTransfer(user.id, tokenId, address, quantity)
        return ResponseEntity(externalTransfer, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/thresholdTransferToken/{tokenId}/{orgId}/{userId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun ThresholdTransfer(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("orgId") orgId : String,
        @PathVariable("userId") userId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        if (quantity.toFloat() < 100){
            var assetList = commonService.transferedIssueToken(user.id, tokenId, orgId, userId, quantity)
            return ResponseEntity(assetList, HttpHeaders(), HttpStatus.OK)
        } else {
            var postTransaction = commonService.postThresholdTransaction(user.id, tokenId, orgId, userId, quantity)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        }

    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/generateSingleWallet"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun SingleWallet(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        var postTransaction = commonService.generateSingleWallet()
        return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)

    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/thresholdTransferTokenSingleWallet/{tokenId}/{orgId}/{userId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun ThresholdTransferSingleWallet(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("orgId") orgId : String,
        @PathVariable("userId") userId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        if (quantity.toFloat() < 100){
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 0)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else if (quantity.toFloat() < 1000){
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 1)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else if (quantity.toFloat() < 10000){
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 2)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else if (quantity.toFloat() < 100000){
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 3)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else if (quantity.toFloat() < 1000000){
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 4)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else {
            var postTransaction = commonService.postThresholdTransactionSingleWallet(user.id, tokenId, orgId, userId, quantity, 5)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        }

    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/createUserAccount/{tokenId}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun createAccount(
        @PathVariable("tokenId") tokenId : String,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var account = commonService.createAccount(tokenId, user.id)
        var keys = commonService.generateECDSAWalletSharedKey(user.id, tokenId)
        var postedkeysresponse = commonService.postECDSAWalletSharedKey(user.id, keys)
        return ResponseEntity(account, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/createAdminAccount/{tokenId}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun createAdminAccount(
        @PathVariable("tokenId") tokenId : String
    ): ResponseEntity<String> {
        var account = commonService.createAccount(tokenId, "default")
        return ResponseEntity(account, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getAdminAccounts"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getAdminAccounts(
    ): ResponseEntity<String> {
        var accountBalance = commonService.getUserAccounts("default")
        return ResponseEntity(accountBalance, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getUserAccounts"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getUserAccounts(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var accountBalance = commonService.getUserAccounts(user.id)
        val mapper2 = jacksonObjectMapper()
        mapper2.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
        val listUserAccounts = mapper2.readValue<ArrayList<UserAccount>>(accountBalance)

        val fullExchangeAccounts = commonService.getExchangeAccounts(user.id,listUserAccounts, dataService)

        // Loop over connector list to get external accounts also
        val accountJSON = mapper2.writeValueAsString(fullExchangeAccounts)
        return ResponseEntity(accountJSON, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getSigningTx"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getSigningTx(
        @RequestHeader("Authorization") token : String,
        @RequestParam ("txHash") txHash : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tx = commonService.getSigningTransaction(txHash)
        return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getSigningExternalTx"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getSigningExternalTx(
        @RequestHeader("Authorization") token : String,
        @RequestParam ("txHash") txHash : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tx = commonService.getSigningExternalTransaction(txHash)
        return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/postSignedTx"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun postSignedTx(
        @RequestHeader("Authorization") token : String,
        @RequestBody signedTx: PostSignature
        ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tx = commonService.postSignedTransaction(user.id , signedTx.txHash, signedTx.signedMsg, signedTx.signedDate, signedTx.keyAlias, signedTx.userId, signedTx.tokenId)
        var fabrictx = mapper.readValue<FabricSigningTx>(tx)
        if (fabrictx.status == "Verified"){
            var submitTx = commonService.executeSignedTx(user.id, fabrictx)
            return ResponseEntity(submitTx, HttpHeaders(), HttpStatus.OK)

        }
        return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/postSignedExternalTx"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun postSignedExternalTx(
        @RequestHeader("Authorization") token : String,
        @RequestBody signedTx: PostSignature
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tx = commonService.postSignedExternalTransaction(user.id , signedTx.txHash, signedTx.signedMsg, signedTx.signedDate, signedTx.keyAlias, signedTx.userId, signedTx.tokenId)
        val externalTx = mapper.readValue<ExternalFabricTx>(tx)
        if (externalTx.status == "Verified"){
            val completeSendETH = commonService.executeExternalSignedTx(user.id, externalTx.accountOwner, ethService)
        } else if (externalTx.status == "Admin") {
            var assetList = commonService.transferedTokenAdmin(user.id, externalTx.accountOwner)
            println(assetList)
            return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
        }
        return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getTokenList"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getTokenList(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tokenList = commonService.getTokenList(user.id)
        return ResponseEntity(tokenList, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getLatestData"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getLatestList(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tokenListJSON = commonService.getTokenList(user.id)
        var tokenList = mapper.readValue<List<BasicToken>>(tokenListJSON)
        val symbolList : MutableList<String> = mutableListOf()
        for (token in tokenList){
            symbolList.add(token.token_symbol)
        }
        val arrayList = symbolList.toCollection(ArrayList())
        val marketData = dataService.latestQuote(arrayList)

        return ResponseEntity(marketData, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getContainerList"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getContainerList(
    ): ResponseEntity<List<Container>> {

        val response = dockerService.listContainers()
        return ResponseEntity(response, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/createContainer/{name}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun createContainer(
        @PathVariable("name") name : String
    ): ResponseEntity<CreateContainerResponse> {

        val response = dockerService.createContainers(name)
        return ResponseEntity(response, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/readGoCryptoConfig"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun readConfig(
    ): ResponseEntity<GoCryptoConfig> {
        val mapper = jacksonObjectMapper()
        val jsonString: String = File("./config.json").readText(Charsets.UTF_8)
        val jsonres: GoCryptoConfig = mapper.readValue<GoCryptoConfig>(jsonString)

        return ResponseEntity(jsonres, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getAllPendingTxs"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getAllPendingTxs(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var pendingTxs = commonService.getPendingThresholdTxs(user.id)
        return ResponseEntity(pendingTxs, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getAllThresholdTxs"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getAllThresholdTxs(
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var pendingTxs = commonService.getAllThresholdTxs(user.id)
        return ResponseEntity(pendingTxs, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/approvePendingTxapprovePendingTx/{txHash}"],
        produces = ["application/json"],
        method = [RequestMethod.POST]
    )
    fun approvePendingTx(
        @RequestHeader("Authorization") token : String,
        @PathVariable("txHash") txHash: String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var approveTxHash = commonService.postPartialSignatureSingleWallet(txHash)
        return ResponseEntity(approveTxHash, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/generateECDSAWalletSharedKey/{tokenId}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun generateECDSAWalletSharedKey(
        @PathVariable("tokenId") tokenId : String,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        var keys = commonService.generateECDSAWalletSharedKey(user.id, tokenId)
        var postedkeysresponse = commonService.postECDSAWalletSharedKey(user.id, keys)


        return ResponseEntity(postedkeysresponse, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/thresholdECDSATransferToken/{tokenId}/{orgId}/{userId}/{quantity}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun ThresholdECDSATransfer(
        @PathVariable("tokenId") tokenId : String,
        @PathVariable("orgId") orgId : String,
        @PathVariable("userId") userId : String,
        @PathVariable("quantity") quantity : Float,
        @RequestHeader("Authorization") token : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)

        if (!checkUser){
            return ResponseEntity("User is Not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }

        if (tokenId == "ETH"){
            var postTransaction = commonService.postETHECDSAThresholdTransaction(user.id, tokenId, orgId, userId, quantity, 0)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        } else {
            var postTransaction = commonService.postECDSAThresholdTransaction(user.id, tokenId, orgId, userId, quantity, 0)
            return ResponseEntity(postTransaction, HttpHeaders(), HttpStatus.OK)
        }
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/approveECDSATx/{txHash}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun approveECDSATx(
        @RequestHeader("Authorization") token : String,
        @PathVariable("txHash") txHash : String,
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var ecdsaTxJSON = commonService.queryECDSAThresholdTransaction(user.id, txHash)
        val ecdsaTx = mapper.readValue<FabricECDSAThresholdSigningTx>(ecdsaTxJSON)
        var approvedECDSAResponse = commonService.prepareApproveECDSATx(user.id, txHash, ecdsaTx.signerID, ecdsaTx.tokenId)
        return ResponseEntity(approvedECDSAResponse, HttpHeaders(), HttpStatus.OK)
    }


    @InternalSerializationApi
    @RequestMapping(
        value = ["/performECDSARounds/{txHash}"],
        produces = ["application/json"],
        method = [RequestMethod.PUT]
    )
    fun performECDSARounds(
        @RequestHeader("Authorization") token : String,
        @PathVariable("txHash") txHash : String,
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var signatureResponse = commonService.performECDSARounds(user.id, txHash)
        return ResponseEntity(signatureResponse, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getECDSASigningTx"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getECDSASigningTx(
        @RequestHeader("Authorization") token : String,
        @RequestParam ("txHash") txHash : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var tx = commonService.queryECDSAThresholdTransaction(user.id, txHash)
        return ResponseEntity(tx, HttpHeaders(), HttpStatus.OK)
    }

    @InternalSerializationApi
    @RequestMapping(
        value = ["/getWalletData"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getWalletData(
        @RequestHeader("Authorization") token : String,
        @RequestParam ("tokenId") tokenId : String
    ): ResponseEntity<String> {
        val userInfo = commonService.getUserInfo(token)
        val mapper = jacksonObjectMapper()
        val user = mapper.readValue<GoogleProfile>(userInfo)
        val checkUser = commonService.checkUser(user)
        if (!checkUser){
            return ResponseEntity("User not Registered", HttpHeaders(), HttpStatus.BAD_REQUEST)
        }
        var pkInfo = commonService.getWalletData(user.id, tokenId)
        return ResponseEntity(pkInfo, HttpHeaders(), HttpStatus.OK)
    }

}

