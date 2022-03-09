package com.example.defiServer.connector.cryptogalaxy

import io.netty.bootstrap.Bootstrap
import io.netty.buffer.Unpooled
import io.netty.channel.*
import io.netty.channel.nio.NioEventLoopGroup
import io.netty.channel.socket.SocketChannel
import io.netty.channel.socket.nio.NioSocketChannel
import io.netty.handler.codec.http.DefaultHttpHeaders
import io.netty.handler.codec.http.FullHttpResponse
import io.netty.handler.codec.http.HttpClientCodec
import io.netty.handler.codec.http.HttpObjectAggregator
import io.netty.handler.codec.http.websocketx.*
import io.netty.handler.codec.http.websocketx.extensions.compression.WebSocketClientCompressionHandler
import io.netty.handler.ssl.SslContext
import io.netty.handler.ssl.SslContextBuilder
import io.netty.handler.ssl.util.InsecureTrustManagerFactory
import io.netty.util.CharsetUtil
import kotlinx.serialization.InternalSerializationApi
import org.springframework.messaging.simp.stomp.*
import org.springframework.stereotype.Service
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.URI
import java.util.*


class Ticker(
    val Exchange: String,
    val MktID : String,
    val MktCommitName : String,
    val Price : Float,
    val Timestamp : String,
    val InfluxVal : Int
    )



@InternalSerializationApi
@Service
class CryptogalaxyWebsocketService {
    val URL = System.getProperty("url", "ws://127.0.0.1:8080/ws")


    fun set() {
        val uri = URI(URL)
        val scheme = if (uri.scheme == null) "ws" else uri.scheme
        val host = if (uri.host == null) "127.0.0.1" else uri.host
        val port: Int
        port = if (uri.port == -1) {
            if ("ws".equals(scheme, ignoreCase = true)) {
                80
            } else if ("wss".equals(scheme, ignoreCase = true)) {
                443
            } else {
                -1
            }
        } else {
            uri.port
        }
        if (!"ws".equals(scheme, ignoreCase = true) && !"wss".equals(scheme, ignoreCase = true)) {
            System.err.println("Only WS(S) is supported.")
            return
        }
        val ssl = "wss".equals(scheme, ignoreCase = true)
        val sslCtx: SslContext?
        sslCtx = if (ssl) {
            SslContextBuilder.forClient()
                .trustManager(InsecureTrustManagerFactory.INSTANCE).build()
        } else {
            null
        }
        val group: EventLoopGroup = NioEventLoopGroup()
        try {
            // Connect with V13 (RFC 6455 aka HyBi-17). You can change it to V08 or V00.
            // If you change it to V00, ping is not supported and remember to change
            // HttpResponseDecoder to WebSocketHttpResponseDecoder in the pipeline.
            val handler = WebSocketClientHandler(
                WebSocketClientHandshakerFactory.newHandshaker(
                    uri, WebSocketVersion.V13, null, true, DefaultHttpHeaders()
                )
            )
            val b = Bootstrap()
            b.group(group)
                .channel(NioSocketChannel::class.java)
                .handler(object : ChannelInitializer<SocketChannel>() {
                    override fun initChannel(ch: SocketChannel) {
                        val p: ChannelPipeline = ch.pipeline()
                        if (sslCtx != null) {
                            p.addLast(sslCtx.newHandler(ch.alloc(), host, port))
                        }
                        p.addLast(
                            HttpClientCodec(),
                            HttpObjectAggregator(8192),
                            WebSocketClientCompressionHandler.INSTANCE,
                            handler
                        )
                    }
                })
            val ch: Channel = b.connect(uri.host, port).sync().channel()
            handler.handshakeFuture()!!.sync()
            val console = BufferedReader(InputStreamReader(System.`in`))
            while (true) {
                val msg = console.readLine()
                if (msg == null) {
                    break
                } else if ("bye" == msg.toLowerCase()) {
                    ch.writeAndFlush(CloseWebSocketFrame())
                    ch.closeFuture().sync()
                    break
                } else if ("ping" == msg.toLowerCase()) {
                    val frame: WebSocketFrame = PingWebSocketFrame(Unpooled.wrappedBuffer(byteArrayOf(8, 1, 8, 1)))
                    ch.writeAndFlush(frame)
                } else {
                    val frame: WebSocketFrame = TextWebSocketFrame(msg)
                    ch.writeAndFlush(frame)
                }
            }
        } finally {
            group.shutdownGracefully()
        }
    }
}

class WebSocketClientHandler(private val handshaker: WebSocketClientHandshaker) :
    SimpleChannelInboundHandler<Any?>() {
    private var handshakeFuture: ChannelPromise? = null
    fun handshakeFuture(): ChannelFuture? {
        return handshakeFuture
    }

    override fun handlerAdded(ctx: ChannelHandlerContext) {
        handshakeFuture = ctx.newPromise()
    }

    override fun channelActive(ctx: ChannelHandlerContext) {
        handshaker.handshake(ctx.channel())
    }

    override fun channelInactive(ctx: ChannelHandlerContext) {
        println("WebSocket Client disconnected!")
    }

    @Throws(Exception::class)
    public override fun channelRead0(ctx: ChannelHandlerContext, msg: Any?) {
        val ch: Channel = ctx.channel()
        if (!handshaker.isHandshakeComplete) {
            handshaker.finishHandshake(ch, msg as FullHttpResponse?)
            println("WebSocket Client connected!")
            handshakeFuture!!.setSuccess()
            return
        }
        if (msg is FullHttpResponse) {
            val response = msg
            throw IllegalStateException(
                "Unexpected FullHttpResponse (getStatus=" + response.status() +
                        ", content=" + response.content().toString(CharsetUtil.UTF_8) + ')'
            )
        }
        val frame = msg as WebSocketFrame?
        if (frame is TextWebSocketFrame) {
            println("WebSocket Client received message: " + frame.text())
        } else if (frame is PongWebSocketFrame) {
            println("WebSocket Client received pong")
        } else if (frame is CloseWebSocketFrame) {
            println("WebSocket Client received closing")
            ch.close()
        }
    }

    override fun exceptionCaught(ctx: ChannelHandlerContext, cause: Throwable) {
        cause.printStackTrace()
        if (!handshakeFuture!!.isDone) {
            handshakeFuture!!.setFailure(cause)
        }
        ctx.close()
    }
}

