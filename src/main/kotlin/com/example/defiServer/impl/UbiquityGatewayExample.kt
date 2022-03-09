package com.example.defiServer.impl

import com.gitlab.blockdaemon.ubiquity.UbiquityClientBuilder
import com.gitlab.blockdaemon.ubiquity.model.MultiTransferOperation
import com.gitlab.blockdaemon.ubiquity.model.Network
import com.gitlab.blockdaemon.ubiquity.model.Order
import com.gitlab.blockdaemon.ubiquity.model.Platform
import org.web3j.ens.Contracts.ROPSTEN
import org.web3j.tx.ChainId.ROPSTEN


object UbiquityClientExample {
    //bd1bc0meoYrfp1sANZqKzB44QeWoB3LrEYKoiJfrt3N2jjH
    private const val ETH_ACCOUNT = "0x4f635AB7B9f488E6a43fDE8475C2dD54a2400686"
    private const val ETH_TX_ID = "0x6821b32162ad40f979ad8e999ffbe358e5df0f54e1894d1b3fc3e01fce6a134b"
    private const val ALGO_ACCOUNT = "HG2JL36OPPITBA7RNIPW4GUQS74AF3SEBO6DAJSLJC33C34I2DQ42F5MU4"
    @Throws(Exception::class)
    @JvmStatic
    fun example(args: Array<String>) {
        println("HELPME666")

        val authBearerToken = System.getenv("AUTH_TOKEN")
        val client = UbiquityClientBuilder().authBearerToken(authBearerToken).build()
        val details = client.platform().getPlatform(Platform.BITCOIN, Network.MAIN_NET)
        println(details)
        val height = client.sync().currentBlockNumber(Platform.BITCOIN, Network.MAIN_NET)
        println(height)
        val id = client.sync().currentBlockID(Platform.BITCOIN, Network.MAIN_NET)
        println(id)

        // New protocols or networks may not be present in the client but can be
        // accessed via the name method
        val block = client.block().getBlock(Platform.name("bitcoin"), Network.name("mainnet"), 685700)
        println(block)

        // Operation is a wrapper for various different types in this case a MultiTransferOperation
        val operation = block.txs!![0].operations!!["script"]
        val multiTransferOperation = operation!!.multiTransferOperation
        println(multiTransferOperation)

        // If you are unsure of which type to expect the getActualInstance returns the instance as an object which can then be type checked
        val unkownOperation = operation.actualInstance
        if (unkownOperation is MultiTransferOperation) {
            println("MultiTransferOperation")
        }
        val currentBlock = client.block().getBlock(Platform.ETHEREUM, Network.name("ropsten"), "current")
        println(currentBlock)
        val balances = client.accounts().getBalancesByAddress(
            Platform.ALGORAND,
            Network.name("mainnet"), ALGO_ACCOUNT
        )
        println(balances)
        val tx = client.transactions().getTx(Platform.name("ethereum"), Network.name("mainnet"), ETH_TX_ID)
        println(tx)
        val accountTxPage = client.accounts().getTxsByAddress(Platform.ETHEREUM, Network.MAIN_NET, ETH_ACCOUNT)
        println(accountTxPage)

        // Initial request to paged API's should not include a continuation.
        // If no limit is supplied the default of 25 will be applied
        // A filter can also be applied to select the returned assets
        val txPage1 = client.transactions().getTxs(Platform.ETHEREUM, Network.MAIN_NET, Order.DESC, "", null, null)
        println(txPage1)

        // To continue through the pages of transactions the continuation from the
        // previous page must be supplied to the next request
        val previous = txPage1.continuation
        val txPage2 = client.transactions().getTxs(Platform.ETHEREUM, Network.MAIN_NET, Order.DESC, previous)
        println(txPage2)
    }
}
