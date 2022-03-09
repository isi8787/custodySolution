package com.example.defiServer.service

import com.example.defiServer.models.FabricUser
import com.example.defiServer.models.GoCryptoConfig
import com.example.defiServer.models.GoCryptoExchanges
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.github.dockerjava.api.DockerClient
import com.github.dockerjava.api.command.CreateContainerResponse
import com.github.dockerjava.api.model.*
import com.github.dockerjava.core.DockerClientBuilder
import kotlinx.serialization.InternalSerializationApi
import org.springframework.stereotype.Service
import java.io.File
import java.util.*
import kotlin.io.path.ExperimentalPathApi
import kotlin.io.path.createTempDirectory


@InternalSerializationApi
@Service
class DockerService {

    fun listContainers() : List<Container>{
        val dockerClient: DockerClient = DockerClientBuilder.getInstance("unix:///var/run/docker.sock").build()

        val containers: List<Container> = dockerClient.listContainersCmd()
            .withShowSize(true)
            .withShowAll(true)
            .withStatusFilter(listOf("exited")).exec()

        return containers
    }

    @OptIn(ExperimentalPathApi::class)
    fun createContainers(name : String) : CreateContainerResponse{
        val mapper = jacksonObjectMapper()
        val jsonString: String = File("./config.json").readText(Charsets.UTF_8)
        val templateCrypto: GoCryptoConfig = mapper.readValue<GoCryptoConfig>(jsonString)

        val userExchanges = listOf("Kraken", "Binance")
        val usergocryptoexchange : MutableList<GoCryptoExchanges> = mutableListOf<GoCryptoExchanges>()
        for (exchange in templateCrypto.exchanges){
            if(userExchanges.contains(exchange.name)){
                usergocryptoexchange.add(exchange)
            }
        }
        var finalConfig = templateCrypto
        println(usergocryptoexchange)
        finalConfig.exchanges = usergocryptoexchange.toTypedArray()

        val finalJSON = mapper.writeValueAsString(finalConfig)
        val tmpDir = createTempDirectory(name+"tmp")
        File(tmpDir.toString()+"/config.json").writeText(finalJSON)

        val dockerClient: DockerClient = DockerClientBuilder.getInstance("unix:///var/run/docker.sock").build()
        var containerPath = tmpDir.toString().split("/")

        val containerResp: CreateContainerResponse = dockerClient.createContainerCmd("gocryptotrader:latest")
            .withPublishAllPorts(true)
            .withTty(true)
            .withName(name)
            .withEnv(listOf("VIRTUAL_PORT=9053", "VIRTUAL_HOST=gocrypto." + name, "TMPCONFIG="+tmpDir.toString()))
            .exec();


        val copyResy = dockerClient.copyArchiveToContainerCmd(containerResp.id)
            .withRemotePath("/root/.gocryptotrader")
            .withHostResource(tmpDir.toString()+"/config.json")
            .exec()
        println(copyResy)


        val startContainer= dockerClient.startContainerCmd(containerResp.id).exec()


        println(startContainer)
        return containerResp
    }

    @OptIn(ExperimentalPathApi::class)
    fun startGoCrypto(userInfo : FabricUser) : CreateContainerResponse{
        val mapper = jacksonObjectMapper()
        val jsonString: String = File("./config.json").readText(Charsets.UTF_8)
        val templateCrypto: GoCryptoConfig = mapper.readValue<GoCryptoConfig>(jsonString)


        val usergocryptoexchange : MutableList<GoCryptoExchanges> = mutableListOf<GoCryptoExchanges>()
        for (exchange in templateCrypto.exchanges){
            if(userInfo.connectors!!.containsKey(exchange.name)){
                println("FOUND: " + exchange.name)
                val connector = userInfo.connectors[exchange.name]
                var updatedExchange = exchange
                updatedExchange.api.authenticatedSupport = true
                updatedExchange.api.authenticatedWebsocketApiSupport = true
                updatedExchange.api.credentials.key = connector!![0].apiKey
                updatedExchange.api.credentials.secret = connector!![0].apiSecret
                if (connector!![0].clientId != null){
                    updatedExchange.api.credentials.clientID = connector!![0].clientId
                }
                usergocryptoexchange.add(updatedExchange)
            }
        }

        if (usergocryptoexchange.size < 1){
            return CreateContainerResponse()
        }
        var finalConfig = templateCrypto
        finalConfig.exchanges = usergocryptoexchange.toTypedArray()

        val finalJSON = mapper.writeValueAsString(finalConfig)
        val tmpDir = createTempDirectory(userInfo.id+"tmp")
        File(tmpDir.toString()+"/config.json").writeText(finalJSON)

        val dockerClient: DockerClient = DockerClientBuilder.getInstance("unix:///var/run/docker.sock").build()
        var containerPath = tmpDir.toString().split("/")

        val activeContainers: MutableList<Container> = dockerClient.listContainersCmd().exec()
        println(mapper.writeValueAsString(activeContainers))
        for (container in activeContainers){
            if (container.names.contains("/"+userInfo.id)){
                dockerClient.stopContainerCmd(container.id).exec()
                dockerClient.removeContainerCmd(container.id).exec()
                break
            }
        }

        val containerResp: CreateContainerResponse = dockerClient.createContainerCmd("gocryptotrader:latest")
            .withTty(true)
            .withName(userInfo.id)
            .withEnv(listOf("VIRTUAL_PORT=9053", "VIRTUAL_HOST=gocrypto." + userInfo.id, "TMPCONFIG="+tmpDir.toString()))
            .exec();

        val copyConfigFile = dockerClient.copyArchiveToContainerCmd(containerResp.id)
            .withRemotePath("/root/.gocryptotrader")
            .withHostResource(tmpDir.toString()+"/config.json")
            .exec()

        val startContainer= dockerClient.startContainerCmd(containerResp.id).exec()

        return containerResp
    }
}