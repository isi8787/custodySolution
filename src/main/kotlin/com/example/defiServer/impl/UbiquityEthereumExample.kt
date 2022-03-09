package com.example.defiServer.impl

import com.gitlab.blockdaemon.ubiquity.UbiquityClientBuilder
import com.gitlab.blockdaemon.ubiquity.model.MultiTransferOperation
import com.gitlab.blockdaemon.ubiquity.model.Network
import com.gitlab.blockdaemon.ubiquity.model.Order
import com.gitlab.blockdaemon.ubiquity.model.Platform
import org.web3j.ens.Contracts.ROPSTEN
import org.web3j.tx.ChainId.ROPSTEN


object UbiquityClientExample2 {
    //bd1bc0meoYrfp1sANZqKzB44QeWoB3LrEYKoiJfrt3N2jjH
    private const val ETH_ACCOUNT = "0x4f635AB7B9f488E6a43fDE8475C2dD54a2400686"
    private const val ETH_TX_ID = "0x4d84c4776bce861fc0841921244f845e7a4d8a3b4f2251989b286d16c19893d2"
    private const val ALGO_ACCOUNT = "HG2JL36OPPITBA7RNIPW4GUQS74AF3SEBO6DAJSLJC33C34I2DQ42F5MU4"
    @Throws(Exception::class)
    @JvmStatic
    fun example(args: Array<String>) {
        println("HELPME666")

        val authBearerToken = System.getenv("AUTH_TOKEN")
        val client = UbiquityClientBuilder().authBearerToken(authBearerToken).build()

        val details = client.platform().getPlatform(Platform.ETHEREUM, Network.name("ropsten"))
        println("HELP1")

        println(details)
        val height = client.sync().currentBlockNumber(Platform.ETHEREUM, Network.name("ropsten"))
        println("HELP2")

        println(height)
        val id = client.sync().currentBlockID(Platform.ETHEREUM, Network.name("ropsten"))

        println("HELP3")
        println(id)

        // New protocols or networks may not be present in the client but can be
        // accessed via the name method
        val block = client.block().getBlock(Platform.ETHEREUM, Network.name("ropsten"), "current")
        println("HELP4")
        println(block)


        val balances = client.accounts().getBalancesByAddress(Platform.ETHEREUM, Network.name("ropsten"), ETH_ACCOUNT)
        println("HELP5")

        println("HELP5" + balances)
    }
}
