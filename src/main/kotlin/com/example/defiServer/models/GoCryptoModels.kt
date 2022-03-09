package com.example.defiServer.models

import com.fasterxml.jackson.annotation.JsonAlias
import com.fasterxml.jackson.annotation.JsonProperty
import kotlinx.serialization.*
import kotlinx.serialization.json.*


@Serializable
data class GoCryptoConfig(
    val name: String,
    val dataDirectory: String,
    val encryptConfig: Int,
    val globalHTTPTimeout: Long,
    val database : GoCryptoDatabase,
    val logging : GoCryptoLogging,
    val connectionMonitor : GoCryptoMonitor,
    val dataHistoryManager : GoCryptoDHManager,
    val currencyStateManager : GoCryptoCSManager,
    val profiler : GoCryptoProfiler,
    val ntpclient : GoCryptoNTPClient,
    val gctscript : GoCryptoGCTScript,
    val currencyConfig : GoCryptoCurrency,
    val communications : GoCryptoCommunications,
    val remoteControl : GoCryptoRemoteControl,
    val portfolioAddresses : GoCryptoPortfolioAddresses,
    var exchanges : Array<GoCryptoExchanges>,
    val bankAccounts : Array<GoCryptoBankAccounts>
)


@Serializable
data class GoCryptoDatabase(
    val enabled : Boolean,
    val verbose : Boolean,
    val driver : String,
    val connectionDetails : GoCryptoConnectionDetails
)

@Serializable
data class GoCryptoLogging(
    val enabled : Boolean,
    val level : String,
    val output : String,
    val fileSettings : GoCryptoFileSettings,
    val advancedSettings : GoCryptoAdvanceSettings

)

@Serializable
data class GoCryptoMonitor(
    val preferredDNSList : Array<String>,
    val preferredDomainList : Array<String>,
    val checkInterval : Long
)

@Serializable
data class GoCryptoDHManager(
    val enabled : Boolean,
    val checkInterval : Long,
    val maxJobsPerCycle : Int,
    val maxResultInsertions : Int,
    val verbose : Boolean
)

@Serializable
data class GoCryptoCSManager(
    val enabled : Boolean,
    val delay : Long
)

@Serializable
data class GoCryptoProfiler(
    val enabled : Boolean,
    val mutex_profile_fraction : Int
)

@Serializable
data class GoCryptoNTPClient(
    val enabled : Int,
    val pool : Array<String>,
    val allowedDifference : Long,
    val allowedNegativeDifference : Long
)

@Serializable
data class GoCryptoGCTScript(
    val enabled : Boolean,
    val timeout : Long,
    val max_virtual_machines : Int,
    val allow_imports : Boolean,
    val auto_load : Array<String>,
    val verbose : Boolean
)

@Serializable
data class GoCryptoCurrency(
    val forexProviders : Array<GoCryptoForexProvider>,
    val cryptocurrencyProvider : GoCryptoCurrencyProvider,
    val cryptocurrencies : String,
    val currencyPairFormat : GoCryptoPairFormat,
    val fiatDisplayCurrency : String,
    val currencyFileUpdateDuration : Int,
    val foreignExchangeUpdateDuration : Int
)

@Serializable
data class GoCryptoCommunications(
    val slack : GoCryptoSlack,
    val smsGlobal : GoCryptoSMSGlobal,
    val smtp : GoCryptoSMTP,
    val telegram : GoCryptoTelegram
)
@Serializable
data class GoCryptoConnectionDetails(
    val host : String,
    val port : Int,
    val username : String,
    val password : String,
    val database : String,
    val sslmode : String

)

@Serializable
data class GoCryptoFileSettings(
    val filename : String,
    val rotate : Boolean,
    val maxsize : Int
)

@Serializable
data class GoCryptoAdvanceSettings(
    val showLogSystemName : Boolean,
    val spacer : String,
    val timeStampFormat : String,
    val headers : GoCryptoAdvanceSettingsHeaders
)

@Serializable
data class GoCryptoAdvanceSettingsHeaders(
    val info : String,
    val warn : String,
    val debug : String,
    val error : String
)

@Serializable
data class GoCryptoForexProvider(
    val name : String,
    val enabled : Boolean,
    val verbose : Boolean,
    val restPollingDelay : Long,
    val apiKey : String,
    val apiKeyLvl : Int,
    val primaryProvider : Boolean
)

@Serializable
data class GoCryptoCurrencyProvider(
    val name : String,
    val enabled : Boolean,
    val verbose : Boolean,
    val apiKey: String,
    val accountPlan : String
)

@Serializable
data class GoCryptoPairFormat(
    val uppercase : Boolean,
    val delimiter : String?,
    val separator : String?,
    val index : String?
)

@Serializable
data class GoCryptoSlack(
    val name : String,
    val enabled : Boolean,
    val verbose : Boolean,
    val targetChannel : String,
    val verificationToken : String
)

@Serializable
data class GoCryptoSMSGlobal(
    val name : String,
    val from : String,
    val enabled :Boolean,
    val verbose : Boolean,
    val username : String,
    val password: String,
    val contacts : Array<SMSContact>,
)

@Serializable
data class SMSContact(
    val name : String,
    val number : String,
    val enabled : Boolean
)

@Serializable
data class GoCryptoSMTP(
    val name : String,
    val enabled :Boolean,
    val verbose : Boolean,
    val host : String,
    val port : String,
    val accountName : String,
    val accountPassword : String,
    val from :String,
    val recipientList : String
)

@Serializable
data class GoCryptoTelegram(
    val name : String,
    val enabled :Boolean,
    val verbose : Boolean,
    val verificationToken : String
)

@Serializable
data class GoCryptoRemoteControl(
    val username : String,
    val password : String,
    val gRPC : GoCryptoGRPC,
    val deprecatedRPC : GoCryptoDeprecatedGRPC,
    val websocketRPC : GoCryptoWebsocket
)

@Serializable
data class GoCryptoGRPC(
    val enabled : Boolean,
    val listenAddress : String,
    val grpcProxyEnabled : Boolean,
    val grpcProxyListenAddress : String,
    val timeInNanoSeconds : Boolean
)

@Serializable
data class  GoCryptoDeprecatedGRPC(
    val enabled : Boolean,
    val listenAddress : String
)

@Serializable
data class GoCryptoWebsocket(
    val enabled : Boolean,
    val listenAddress : String,
    val connectionLimit : Int,
    val maxAuthFailures : Int,
    val allowInsecureOrigin : Boolean
)

@Serializable
data class GoCryptoPortfolioAddresses(
    val addresses : Array<GoCryptoAddress>,
    val Verbose : Boolean
)

@Serializable
data class GoCryptoAddress(
    val Address : String,
    val AddressTag : String,
    val CoinType : String,
    val Balance : Float,
    val Description : String,
    val WhiteListed : Boolean,
    val ColdStorage : Boolean,
    val SupportedExchanges : String
)

@Serializable
data class GoCryptoExchanges(
    val name : String,
    val enabled :Boolean,
    val verbose : Boolean,
    val httpTimeout: Long,
    val websocketResponseCheckTimeout: Long,
    val websocketResponseMaxLimit: Long,
    val websocketTrafficTimeout: Long,
    val baseCurrencies: String,
    val currencyPairs : GoCryptoCurrencyPair,
    val api : GoCryptoExchangeAPI,
    val features : GoCryptoExchangeFeatures,
    val bankAccounts : Array<GoCryptoExchangeBankAccounts>,
    val orderbook :  GoCryptoExchangeOrderbook
)

@Serializable
data class GoCryptoCurrencyPair(
    val bypassConfigFormatUpgrades : Boolean,
    val pairs : HashMap<String,GoCryptoExchangePairFormat>,
    val requestFormat : GoCryptoPairFormat?,
    val configFormat: GoCryptoPairFormat?,
    val useGlobalFormat : Boolean?,
    val lastUpdated : Long
)

@Serializable
data class GoCryptoExchangePair(
    val futures : GoCryptoExchangePairFormat,
    val spot : GoCryptoExchangePairFormat
)

@Serializable
data class GoCryptoExchangePairFormat(
    val assetEnabled : Boolean,
    val enabled : String,
    val available : String,
    val requestFormat : GoCryptoPairFormat?,
    val configFormat : GoCryptoPairFormat?
)

@Serializable
data class GoCryptoExchangeAPI(
    var authenticatedSupport : Boolean,
    var authenticatedWebsocketApiSupport : Boolean,
    val credentials : GoCryptoExchangeAPICredentials,
    val credentialsValidator : GoCryptoCredentialsValidator,
    val urlEndpoints : GoCryptoExchangeAPIEndpoint
)

@Serializable
data class GoCryptoExchangeAPICredentials(
    var key : String,
    var secret : String,
    var clientID : String?
)

@Serializable
data class GoCryptoCredentialsValidator(
    val requiresKey: Boolean,
    val requiresSecret: Boolean,
    val requiresBase64DecodeSecret: Boolean,
    val requiresClientID : Boolean?
)

@Serializable
data class GoCryptoExchangeAPIEndpoint(
    @JsonProperty("RestCoinMarginedFuturesURL") val restCoinMarginedFuturesURL : String?,
    @JsonProperty("RestFuturesURL") val restFuturesURL : String?,
    @JsonProperty("RestSpotURL") val restSpotURL : String?,
    @JsonProperty("WebsocketSpotURL") val websocketSpotURL : String?,
    @JsonProperty("EdgeCase1URL") val edgeCase1URL : String?,
    @JsonProperty("RestSpotSupplementaryURL") val restSpotSupplementaryURL : String?,
    @JsonProperty("RestUSDTMarginedFuturesURL") val restUSDTMarginedFuturesURL : String?,
    @JsonProperty("ChainAnalysisURL") val chainAnalysisURL : String?,
    @JsonProperty("WebsocketSpotSupplementaryURL") val websocketSpotSupplementaryURL : String?,
    @JsonProperty("RestSandboxURL") val restSandboxURL : String?

)

@Serializable
data class GoCryptoExchangeFeatures(
    val supports : GoCryptoExchangeFeaturesSupports,
    val enabled : GoCryptoExchangeFeaturesEnabled
)

@Serializable
data class GoCryptoExchangeFeaturesSupports(
    val restAPI : Boolean,
    val restCapabilities : GoCryptoRestCapabilities,
    val websocketAPI : Boolean,
    val websocketCapabilities : GoCryptoWebsocketCapabilities?
)

@Serializable
data class GoCryptoWebsocketCapabilities(
    val name : String?
)

@Serializable
data class GoCryptoRestCapabilities(
    val tickerBatching : Boolean,
    val autoPairUpdates : Boolean
)

@Serializable
data class GoCryptoExchangeFeaturesEnabled(
    val autoPairUpdates: Boolean,
    val websocketAPI: Boolean,
    val saveTradeData: Boolean,
    val tradeFeed: Boolean,
    val fillsFeed: Boolean
)

@Serializable
data class GoCryptoExchangeBankAccounts(
    val enabled: Boolean,
    val bankName: String,
    val bankAddress: String,
    val bankPostalCode: String,
    val bankPostalCity: String,
    val bankCountry: String,
    val accountName: String,
    val accountNumber: String,
    val swiftCode: String,
    val iban: String,
    val supportedCurrencies: String
)

@Serializable
data class GoCryptoExchangeOrderbook(
    val verificationBypass: Boolean,
    val websocketBufferLimit: Int,
    val websocketBufferEnabled: Boolean,
    val publishPeriod: Long
)

@Serializable
data class GoCryptoBankAccounts(
    val enabled: Boolean,
    val id : String,
    val bankName: String,
    val bankAddress: String,
    val bankPostalCode: String,
    val bankPostalCity: String,
    val bankCountry: String,
    val accountName: String,
    val accountNumber: String,
    val swiftCode: String,
    val iban: String,
    val supportedCurrencies: String,
    val supportedExchanges : String
)


@Serializable
data class GoCryptoExchangeAccountResults(
    val exchange : String,
    val accounts : Array<GoCryptoAccountsResults>
)

@Serializable
data class GoCryptoAccountsResults(
    val id : String,
    val currencies : Array<GoCryptoCurrenciesResult>
)

@Serializable
data class GoCryptoCurrenciesResult(
    val currency : String,
    val totalValue : Float,
    val hold : Float,
    var usdBalance : Float?,
    var usdBalanceHold : Float?,
)