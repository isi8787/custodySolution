package com.example.defiServer.models

import com.fasterxml.jackson.annotation.JsonProperty
import kotlinx.serialization.Serializable


/**
 *
 * @param id
 * @param type
 * @param body
 * @param message
 * @param time
 * @param role
 * @param org
 * @param isRead
 */

@Serializable
data class Notification(
    @JsonProperty("id") var id : String,
    @JsonProperty("type") var type : String,
    @JsonProperty("body") var body : String?,
    @JsonProperty("message") var message : String,
    @JsonProperty("time") var time : Long,
    @JsonProperty("org") var org : String?,
    @JsonProperty("isRead") var isRead : Boolean,
)


/**
 *
 * @param id
 * @param type
 * @param body
 * @param message
 * @param time
 * @param role
 * @param org
 * @param isRead
 */
@Serializable
data class NotificationResponse(
    @JsonProperty("id") var id : String,
    @JsonProperty("type") var type : String,
    @JsonProperty("body") var body : String?,
    @JsonProperty("message") var message : String,
    @JsonProperty("time") var time : String,
    @JsonProperty("org") var org : String?,
    @JsonProperty("isRead") var isRead : Boolean,
)
