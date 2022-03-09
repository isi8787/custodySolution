package com.example.defiServer.models

import com.fasterxml.jackson.annotation.JsonAlias
import com.fasterxml.jackson.annotation.JsonProperty
import kotlinx.serialization.*
import kotlinx.serialization.json.*

@Serializable
data class SimpleInfuraRequest(
    val id: Int,
    val jsonrpc: String,
    val method: String,
    val params: ArrayList<String>
)

@Serializable
data class InfuraCountResponse(
    val jsonrpc: String,
    val id: Int,
    val result: String
)

@Serializable
data class InfuraResponse(
    val jsonrpc: String,
    val id: Int,
    val result: TxResult
)

@Serializable
data class TxResult(
    val accessList: ArrayList<String>,
    val blockHash: String,
    val blockNumber: String,
    val chainId: String,
    val from: String,
    val gas: String,
    val gasPrice: String,
    val hash: String,
    val input: String,
    val maxFeePerGas: String,
    val maxPriorityFeePerGas: String,
    val nonce: String,
    val r: String,
    val s: String,
    val to: String,
    val transactionIndex: String,
    val type: String,
    val v: String,
    val value: String
)

@Serializable
data class InternalTx(
    val from: String,
    val toUser: String,
    val toOrgId: String,
    val tokenId: String,
    val quantity: Float,
    val timestamp : String
)


@Serializable
data class FabricTx(
    var signerID: String,
    val messageHash: String,
    val message: String,
    val transactionSubmittedTimeStamp : String,
    val approvalLevel: Int,
    val tokenId : String
)

@Serializable
data class FabricExternalTx(
    var identifier: String,
    val status: String,
    val accountOwner: FabricTx,
    val orgAdmin : FabricTx?,
    val orgSignatory : FabricTx?
)


@Serializable
data class GoogleTokenResponse(
    var access_token: String,
    val expires_in: String,
    val scope: String,
    val token_type: String,
    val id_token: String
)

@Serializable
data class GoogleProfile(

    var id	: String,
    var email: String?,
    var verified_email : Boolean?,
    var name: String,
    var given_name	: String,
    var family_name: String,
    var picture	: String,
    var locale: String
)

@Serializable
data class BasicProfile(

    var id	: String,
    var email: String?,
    var verified_email: Boolean?,
    var given_name	: String,
    var family_name: String,
    var phone : String?,
    var orgId : String?
)

@Serializable
data class TokenList(
    @JsonProperty("Key") var key : String,
    @JsonProperty("Record") var  record : BasicToken,
)

@Serializable
data class BasicToken(
    @JsonProperty("AssetType") var assetType : String,
    @JsonProperty("Behavior") var  behavior : ArrayList<String>,
    @JsonProperty("Currency_name") var currency_name : String,
    @JsonProperty("Divisible") var divisible : Divisible,
    @JsonProperty("Mintable") var mintable : Mintable,
    @JsonProperty("Roles") var roles : Roles,
    @JsonProperty("Token_desc") var token_desc : String,
    @JsonProperty("Token_id") var token_id : String,
    @JsonProperty("Token_name") var token_name : String,
    @JsonProperty("Token_to_currency_ratio") var Token_to_currency_ratio : Int,
    @JsonProperty("Token_type") var token_type : String,
    @JsonProperty("Token_symbol") var token_symbol : String,
)

@Serializable
data class Divisible(
    @JsonProperty("Decimal") var decimal : Int,
)

@Serializable
data class Mintable(
    @JsonProperty("Max_mint_quantity") var max_mint_quantity : Long,
)

@Serializable
data class Roles(
    @JsonProperty("minter_role_name") var minter_role_name : String,
)

@Serializable
data class FabricSigningTx(
    var signerID : String,
    var status : String,
    var message: String,
    var messageHash: String,
    var tokenId: String,
    var keyAlias : String,
    var signedMessage : String,
    var transactionDate : String,
    var transactionSubmittedTimeStamp : String,
    var ECParameters : ECParameters
)

@Serializable
data class ExternalFabricTx(
    @JsonProperty("identifier")
    @JsonAlias("identifier", "Identifier") var identifier : String,
    var status : String,
    var accountOwner : FabricSigningTx,
    var orgAdmin : FabricSigningTx,
    var orgSignatory: FabricSigningTx
)

@Serializable
data class ECParameters(
    var curveName : String,
    var curveType : String,
    var pX : String,
    var pY : String
)

@Serializable
data class PostSignature(
    var txHash : String,
    var signedMsg : String,
    var userId :String,
    var keyAlias : String,
    var signedDate : String,
    var tokenId : String,
    var role : String
)

@Serializable
data class UserAccount(
    @JsonProperty("AccountId") var accountId : String,
    @JsonProperty("Balance") var balance : Float,
    @JsonProperty("BalanceOnHold") var balanceOnHold : Float,
    @JsonProperty("OrgId") var orgId : String,
    @JsonProperty("TokenId") var tokenId : String,
    @JsonProperty("TokenName") var tokenName : String,
    @JsonProperty("TokenSymbol") var tokenSymbol : String,
    @JsonProperty("USDBalance") var usdBalance : Float?,
    @JsonProperty("USDBalanceHold") var usdBalanceHold : Float?,
    @JsonProperty("percent_change_24h") var percent_change_24h : Float?,
    @JsonProperty("volume_24h") var volume_24h : Float?
)

@Serializable
data class BasicQuote(
    var data : HashMap<String, QuoteData>,
    var status : CMCError
)

@Serializable
data class QuoteData(
    var symbol : String,
    var id : String,
    var name : String,
    var amount : Float,
    var  last_updated  : String,
    var quote : HashMap<String, QuotePrice>
)

@Serializable
data class QuotePrice(
    var price : Float,
    var  last_updated  : String,
    var percent_change_24h : Float,
    var volume_24h : Float
)


@Serializable
data class CMCError(
    var timestamp : String?,
    var error_code : Int?,
    var error_message : String?,
    var elapsed : Int?,
    var credit_count : Int?

)


@Serializable
data class ConnectorDetails(
    @JsonProperty("APIName") var apiName : String,
    @JsonProperty("apiKey") var apiKey : String,
    @JsonProperty("apiSecret") var apiSecret : String,
    @JsonProperty("Passphrase") var passphrase : String?,
    @JsonProperty("ClientId") var clientId : String?
)

@Serializable
data class ExchangeCheck(
    var id : String,
    var name : String,
    var status : String
)

@Serializable
data class FabricUser(
    var id : String,
    var email: String,
    val verified_email : Boolean,
    val given_name : String,
    val family_name : String,
    val phone : String,
    val orgId : String,
    val connectors :  HashMap<String, Array<ConnectorDetails>>?
)


@Serializable
data class FullAccountData(
    var custodyAccount : UserAccount,
    var exchanges : HashMap<String, GoCryptoCurrenciesResult>
)


@Serializable
data class SigningNotification(
    var org: String,
    var assetKey: String,
    var asset: ByteArray?
)

@Serializable
data class  ThresholdSignedTx(
    @JsonProperty("ShareIdentifier") var shareIdentifier : Int,
    @JsonProperty("Sig") var sig : String?
)

@Serializable
data class FabricThresholdSigningTx(
    var signerID : String,
    var status : String,
    var message: String,
    var messageHash: String,
    var tokenId: String,
    var signedShares : Array<ThresholdSignedTx>?,
    var transactionDate : String,
    var transactionSubmittedTimeStamp : String,
    var noncePub : String?,
    var pubKey : String?,
    var shareInformation : HashMap<String, ShareInformation>?,
    var approvalLevel : Int,
    var txType : String
)

@Serializable
data class ShareInformation(
    var signedShare : ThresholdSignedTx?,
    var nonceShare : Array<String>,
    var noncePubShare : String
)

@Serializable
data class FabricECDSAThresholdSigningTx(
    var signerID : String,
    var status : String,
    var message: String,
    var messageHash: String,
    var tokenId: String,
    var transactionDate : String,
    var signature : String?,
    var pubKey : String?,
    var approvalLevel : Int,
    var txType : String,
    var txReceipt : String?
)

@Serializable
data class ApprovalTECDSA(
    var approval : Boolean,
    var orgId : String?,
    var txHash: String,
    var participant : String?,
    var sk : String?,
    var userId : String,
    var tokenId: String
)

@Serializable
data class ETHTX(
    var hash: String,
    var tx : String
)