package com.example.mpc.model

import com.fasterxml.jackson.annotation.JsonProperty


//TODO Deprecated this.
data class TestName(

    @JsonProperty("name") val name: kotlin.String? = null,

    @JsonProperty("phone") val phone: kotlin.String? = null

)