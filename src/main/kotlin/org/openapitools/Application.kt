package org.openapitools

import org.glassfish.tyrus.server.Server
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.beans.factory.annotation.Value
import org.springframework.boot.autoconfigure.EnableAutoConfiguration
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.autoconfigure.domain.EntityScan
import org.springframework.boot.runApplication
import org.springframework.context.annotation.ComponentScan
import org.springframework.context.annotation.Configuration
import org.springframework.data.jpa.repository.config.EnableJpaRepositories
import org.springframework.scheduling.annotation.EnableAsync
import org.springframework.scheduling.annotation.EnableScheduling
import org.springframework.stereotype.Component
import org.springframework.web.servlet.config.annotation.CorsRegistry
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer
import java.io.BufferedReader
import java.io.InputStreamReader



val LOGGER_TAG="Application"
@SpringBootApplication
@ComponentScan(basePackages = ["org.openapitools", "com.example.mpc.*"])
@EntityScan("com.example.mpc.model")
@EnableJpaRepositories("com.example.mpc.repository")
@EnableScheduling
@EnableAutoConfiguration
@EnableAsync
class Application {

    @Autowired
    fun setServerPorts(serverPorts: ServerPorts?) {
        Companion.serverPorts = serverPorts!!
    }

    companion object {
        @Autowired
        private lateinit var serverPorts: ServerPorts

        @JvmStatic
        fun main(args: Array<String>) {

            runApplication<Application>(*args)
            serverPorts.getWSport()?.let { runWSServer(it) }
        }
    }
}

@Component
class ServerPorts {
    @Value("\${server.wsPort}")
    private val wsPort: String? = null

    fun getWSport(): String? {
        return wsPort
    }
}

fun runWSServer(port: String) {
    val server = Server("localhost", port.toInt(), "/websockets")

    try {
        server.start()
        val reader = BufferedReader(InputStreamReader(System.`in`))
        print("Please press a key to stop the server.")
        reader.readLine()
    } catch (e: Exception) {
        throw RuntimeException(e)
    }
}

