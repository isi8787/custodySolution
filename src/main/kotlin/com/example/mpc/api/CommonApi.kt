package com.example.mpc.api



import com.example.mpc.model.TestName
import com.example.mpc.service.CommonService
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import io.kubernetes.client.openapi.models.*
import kotlinx.serialization.InternalSerializationApi
import org.springframework.core.io.InputStreamResource
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.*
import java.io.ByteArrayInputStream
import java.util.*
import javax.validation.Valid
import javax.validation.constraints.NotNull
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
import kotlinx.serialization.KSerializer
import kotlinx.serialization.serializer


val LOGGER_tag="CommonApiController"

@InternalSerializationApi
@RestController
@Validated
@RequestMapping("\${api.base-path:/v1}")
class CommonApiController(val commonService: CommonService) {

    @InternalSerializationApi
    @RequestMapping(
        value = ["/suppliers/{supplierId}"],
        produces = ["application/json"],
        method = [RequestMethod.GET]
    )
    fun getSupplier(
        @PathVariable("supplierId") supplierId: String
    ): ResponseEntity<TestName> {

        val testName = commonService.getSupplier(supplierId)

        return ResponseEntity(testName, HttpHeaders(), HttpStatus.OK)


    }


}


