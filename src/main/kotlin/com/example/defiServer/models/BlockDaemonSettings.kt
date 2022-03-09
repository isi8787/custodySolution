package com.example.defiServer.models

import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import org.springframework.stereotype.Service

@Component
@Service
data class BlockDaemonSettings(

    @Value("\${blockDaemon.apiKey}")
    var apiKey: String,

    @Value("\${blockDaemon.ethPK}")
    var ethPK: String,

    @Value("\${blockDaemon.ethPBK}")
    var ethPBK: String,
)