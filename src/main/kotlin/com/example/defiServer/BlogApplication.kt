package com.example.defiServer

import org.springframework.boot.SpringApplication
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.autoconfigure.domain.EntityScan
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.boot.runApplication
import java.util.*
import javax.ws.rs.core.Application

@SpringBootApplication
@EntityScan("com.example.defiServer.models")
@EnableConfigurationProperties(BlogProperties::class)
class BlogApplication

fun main(args: Array<String>) {
	//runApplication<BlogApplication>(*args)

	val app = SpringApplication(BlogApplication::class.java)
	var port: String = "7080"

	app.setDefaultProperties(
		Collections
			.singletonMap("server.port", port) as Map<String, Any>?
	)
	app.run(*args)

}
