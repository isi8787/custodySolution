package com.example.defiServer.impl

import org.hyperledger.fabric.gateway.impl.GatewayUtils
import org.hyperledger.fabric.gateway.spi.Checkpointer
import java.io.IOException
import java.io.Reader
import java.io.Writer
import java.nio.channels.Channels
import java.nio.channels.FileChannel
import java.nio.channels.FileLock
import java.nio.channels.OverlappingFileLockException
import java.nio.charset.StandardCharsets
import java.nio.file.Files
import java.nio.file.OpenOption
import java.nio.file.Path
import java.nio.file.StandardOpenOption
import java.util.*
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicLong
import java.util.function.Consumer
import java.util.function.Function
import javax.json.Json
import javax.json.JsonObject
import javax.json.JsonString

class FileCheckpointer @Throws(IOException::class)
constructor(private val filePath: Path) : Checkpointer {
    private val fileChannel: FileChannel
    private val fileReader: Reader
    private val fileWriter: Writer
    private val blockNumber = AtomicLong(Checkpointer.UNSET_BLOCK_NUMBER)
    private val transactionIds = Collections.newSetFromMap(ConcurrentHashMap<String, Boolean>())

    init {
        val isFileAlreadyPresent = Files.exists(filePath)
        fileChannel = FileChannel.open(filePath, OPEN_OPTIONS)
        lockFile()

        val utf8Decoder = StandardCharsets.UTF_8.newDecoder()
        fileReader = Channels.newReader(fileChannel, utf8Decoder, -1)

        val utf8Encoder = StandardCharsets.UTF_8.newEncoder()
        fileWriter = Channels.newWriter(fileChannel, utf8Encoder, -1)

        if (isFileAlreadyPresent) {
            load()
        } else {
            save()
        }
    }

    @Throws(IOException::class)
    private fun lockFile() {
        val fileLock: FileLock?
        try {
            fileLock = fileChannel.tryLock()
        } catch (e: OverlappingFileLockException) {
            throw IOException("File is already locked: $filePath", e)
        }

        if (fileLock == null) {
            throw IOException("Another process holds an overlapping lock for file: $filePath")
        }
    }

    @Synchronized
    @Throws(IOException::class)
    private fun load() {
        fileChannel.position(0)
        val savedData = loadJson()
        val version = savedData.getInt(CONFIG_KEY_VERSION, 0)
        if (version == 1) {
            parseDataV1(savedData)
        } else {
            throw IOException("Unsupported checkpoint data version $version from file: $filePath")
        }
    }

    @Throws(IOException::class)
    private fun loadJson(): JsonObject {
        val jsonReader = Json.createReader(fileReader)
        try {
            return jsonReader.readObject()
        } catch (e: RuntimeException) {
            throw IOException("Failed to parse checkpoint data from file: $filePath", e)
        }

    }

    @Throws(IOException::class)
    private fun parseDataV1(json: JsonObject) {
        // Version 1 JSON format:
        // {
        //     version: 1,
        //     block: 1,
        //     transactions: ["a", "b"]
        // }
        try {
            blockNumber.set(json.getJsonNumber(CONFIG_KEY_BLOCK).longValue())
            transactionIds.clear()
            json.getJsonArray(CONFIG_KEY_TRANSACTIONS).getValuesAs(JsonString::class.java).stream()
                .map<String>(Function<JsonString, String> { it.string })
                .forEach(Consumer<String> { transactionIds.add(it) })
        } catch (e: RuntimeException) {
            throw IOException("Bad format of checkpoint data from file: $filePath", e)
        }

    }

    @Synchronized
    @Throws(IOException::class)
    private fun save() {
        val jsonData = buildJson()
        fileChannel.position(0)
        saveJson(jsonData)
        fileChannel.truncate(fileChannel.position())
    }

    private fun buildJson(): JsonObject {
        return Json.createObjectBuilder()
            .add(CONFIG_KEY_VERSION, VERSION)
            .add(CONFIG_KEY_BLOCK, blockNumber.get())
            .add(CONFIG_KEY_TRANSACTIONS, Json.createArrayBuilder(transactionIds))
            .build()
    }

    @Throws(IOException::class)
    private fun saveJson(json: JsonObject) {
        val jsonWriter = Json.createWriter(fileWriter)
        try {
            jsonWriter.writeObject(json)
        } catch (e: RuntimeException) {
            throw IOException("Failed to write checkpoint data to file: $filePath", e)
        }

        fileWriter.flush()
    }

    override fun getBlockNumber(): Long {
        return blockNumber.get()
    }

    @Synchronized
    @Throws(IOException::class)
    override fun setBlockNumber(blockNumber: Long) {
        this.blockNumber.set(blockNumber)
        transactionIds.clear()
        save()
    }

    override fun getTransactionIds(): Set<String> {
        return Collections.unmodifiableSet(transactionIds)
    }

    @Synchronized
    @Throws(IOException::class)
    override fun addTransactionId(transactionId: String) {
        transactionIds.add(transactionId)
        save()
    }

    @Throws(IOException::class)
    override fun close() {
        fileChannel.close() // Also releases lock
    }

    override fun toString(): String {
        return GatewayUtils.toString(
            this,
            "file=$filePath",
            "blockNumber=" + blockNumber.get(),
            "transactionIds=$transactionIds"
        )
    }

    companion object {
        private val OPEN_OPTIONS = Collections.unmodifiableSet<OpenOption>(
            EnumSet.of(
                StandardOpenOption.CREATE,
                StandardOpenOption.READ,
                StandardOpenOption.WRITE
            )
        )
        private val VERSION = 1
        private val CONFIG_KEY_VERSION = "version"
        private val CONFIG_KEY_BLOCK = "block"
        private val CONFIG_KEY_TRANSACTIONS = "transactions"
    }
}