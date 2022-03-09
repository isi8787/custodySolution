package com.example.defiServer.impl

import com.example.defiServer.models.HyperledgerSettings
import org.bouncycastle.asn1.cmc.RevokeRequest
import org.hyperledger.fabric.gateway.*
import org.hyperledger.fabric.sdk.Enrollment
import org.hyperledger.fabric.sdk.User
import org.hyperledger.fabric.sdk.security.CryptoSuite
import org.hyperledger.fabric.sdk.security.CryptoSuiteFactory
import org.hyperledger.fabric_ca.sdk.EnrollmentRequest
import org.hyperledger.fabric_ca.sdk.HFCAClient
import org.hyperledger.fabric_ca.sdk.RegistrationRequest
import org.hyperledger.fabric_ca.sdk.exception.RegistrationException
import java.nio.file.Path
import java.nio.file.Paths
import java.security.PrivateKey
import java.util.*

class FabricRegisterUser {
}

/*
SPDX-License-Identifier: Apache-2.0
*/

object EnrollAdmin {
    @Throws(java.lang.Exception::class)
    @JvmStatic
    fun enroll(adminId : String, hfSettings : HyperledgerSettings) : String {

        // Create a CA client for interacting with the CA.
        val props = Properties()
        props["pemFile"] = hfSettings.pemFile
        props["allowAllHostNames"] = "true"
        val caClient = HFCAClient.createNewInstance(hfSettings.ca, props)
        val cryptoSuite: CryptoSuite = CryptoSuiteFactory.getDefault().cryptoSuite
        caClient.cryptoSuite = cryptoSuite

        // Create a wallet for managing identities
        val wallet: Wallet = Wallets.newFileSystemWallet(Paths.get(hfSettings.keystore))

        // Check to see if we've already enrolled the admin user.
        if (wallet.get(adminId) != null) {
            return ("An identity for the admin user " + adminId + " already exists in the wallet")
        }

        // Enroll the admin user, and import the new identity into the wallet.
        val enrollmentRequestTLS = EnrollmentRequest()
        enrollmentRequestTLS.addHost("localhost")
        enrollmentRequestTLS.setProfile("tls")
        val enrollment = caClient.enroll(adminId, "adminpw", enrollmentRequestTLS)
        val user: Identity = Identities.newX509Identity(hfSettings.orgMSP, enrollment)
        wallet.put(adminId, user)
        return ("Successfully enrolled user \"admin\" and imported it into the wallet")
    }
}


object RegisterUser {
    @Throws(Exception::class)
    @JvmStatic
    fun register(userId : String, hfSettings: HyperledgerSettings) : String {
        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)

        // Create a CA client for interacting with the CA.
        val props = Properties()
        props["pemFile"] = hfSettings.pemFile
        props["allowAllHostNames"] = "true"
        val caClient = HFCAClient.createNewInstance(hfSettings.ca, props)
        val cryptoSuite = CryptoSuiteFactory.getDefault().cryptoSuite
        caClient.cryptoSuite = cryptoSuite

        // Create a wallet for managing identities
        val walletDirectory = Paths.get(hfSettings.keystore)
        val wallet = Wallets.newFileSystemWallet(walletDirectory)

        // Check to see if we've already enrolled the user.
        if (wallet[userId] != null) {
            println("An identity for the user " + userId + "already exists in the wallet")
            return "An identity for the user " + userId + "already exists in the wallet"
        }
        val adminIdentity = wallet["admin"] as X509Identity
        if (adminIdentity == null) {
            println("\"admin\" needs to be enrolled and added to the wallet first")
            return "\"admin\" needs to be enrolled and added to the wallet first"
        }
        val admin: User = object : User {
            override fun getName(): String {
                return "admin"
            }

            override fun getRoles(): Set<String> {
                return setOf("")
            }

            override fun getAccount(): String {
                return ""
            }

            override fun getAffiliation(): String {
                return hfSettings.affiliation
            }

            override fun getEnrollment(): Enrollment {
                return object : Enrollment {
                    override fun getKey(): PrivateKey {
                        return adminIdentity.privateKey
                    }

                    override fun getCert(): String {
                        return Identities.toPemString(adminIdentity.certificate)
                    }
                }
            }

            override fun getMspId(): String {
                return hfSettings.orgMSP
            }
        }

        lateinit var enrollment : Enrollment
        try{
            // Register the user, enroll the user, and import the new identity into the wallet.
            val registrationRequest = RegistrationRequest(userId)
            registrationRequest.affiliation = hfSettings.affiliation
            registrationRequest.enrollmentID = userId
            val enrollmentSecret = caClient.register(registrationRequest, admin)
            enrollment = caClient.enroll(userId, enrollmentSecret)
        } catch (e : Exception) {
            val userIdentity = wallet[userId] as X509Identity
            val user: User = object : User {
                override fun getName(): String {
                    return userId
                }

                override fun getRoles(): Set<String> {
                    return setOf("")
                }

                override fun getAccount(): String {
                    return ""
                }

                override fun getAffiliation(): String {
                    return hfSettings.affiliation
                }

                override fun getEnrollment(): Enrollment {
                    return object : Enrollment {
                        override fun getKey(): PrivateKey {
                            return userIdentity.privateKey
                        }

                        override fun getCert(): String {
                            return Identities.toPemString(userIdentity.certificate)
                        }
                    }
                }

                override fun getMspId(): String {
                    return hfSettings.orgMSP
                }
            }
            enrollment = caClient.reenroll(user)
        }
        val user: Identity = Identities.newX509Identity(hfSettings.orgMSP, enrollment)
        wallet.put(userId, user)
        println("Successfully enrolled user" + userId + "  and imported it into the wallet")
        return "Successfully enrolled user" + userId + "  and imported it into the wallet"
    }

    fun checkUser(userId : String, hfSettings: HyperledgerSettings) : Boolean {
        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)

        // Create a wallet for managing identities
        val walletDirectory = Paths.get(hfSettings.keystore)
        val wallet = Wallets.newFileSystemWallet(walletDirectory)

        // Check to see if we've already enrolled the user.
        if (wallet[userId] != null) {
            return true
        }

        return false
    }

    fun revoke(userId : String, hfSettings: HyperledgerSettings) : String {
        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)

        // Create a CA client for interacting with the CA.
        val props = Properties()
        props["pemFile"] = hfSettings.pemFile
        props["allowAllHostNames"] = "true"
        val caClient = HFCAClient.createNewInstance(hfSettings.ca, props)
        val cryptoSuite = CryptoSuiteFactory.getDefault().cryptoSuite
        caClient.cryptoSuite = cryptoSuite

        // Create a wallet for managing identities
        val walletDirectory = Paths.get(hfSettings.keystore)
        val wallet = Wallets.newFileSystemWallet(walletDirectory)

        // Check to see if we've already enrolled the user.
        if (wallet[userId] == null) {
            return "User " + userId + "does not exist in wallet"
        }
        val adminIdentity = wallet["admin"] as X509Identity
        if (adminIdentity == null) {
            println("\"admin\" needs to be enrolled and added to the wallet first")
            return "\"admin\" needs to be enrolled and added to the wallet first"
        }
        val admin: User = object : User {
            override fun getName(): String {
                return "admin"
            }

            override fun getRoles(): Set<String> {
                return setOf("")
            }

            override fun getAccount(): String {
                return ""
            }

            override fun getAffiliation(): String {
                return hfSettings.affiliation
            }

            override fun getEnrollment(): Enrollment {
                return object : Enrollment {
                    override fun getKey(): PrivateKey {
                        return adminIdentity.privateKey
                    }

                    override fun getCert(): String {
                        return Identities.toPemString(adminIdentity.certificate)
                    }
                }
            }

            override fun getMspId(): String {
                return hfSettings.orgMSP
            }
        }

        // Revoke User
        val revocation = caClient.revoke(admin, userId, "removefromcrl", true)
        wallet.remove(userId)
        println("Successfully revoked and removed user" + userId + "  and imported it into the wallet")
        return "Successfully revoked and removed user" + userId + "  and imported it into the wallet"
    }

    fun reenroll(userId : String, hfSettings: HyperledgerSettings) : String {
        val connectionProfile: Path = Paths.get(hfSettings.connectionFile)

        // Create a CA client for interacting with the CA.
        val props = Properties()
        props["pemFile"] = hfSettings.pemFile
        props["allowAllHostNames"] = "true"
        val caClient = HFCAClient.createNewInstance(hfSettings.ca, props)
        val cryptoSuite = CryptoSuiteFactory.getDefault().cryptoSuite
        caClient.cryptoSuite = cryptoSuite

        // Create a wallet for managing identities
        val walletDirectory = Paths.get(hfSettings.keystore)
        val wallet = Wallets.newFileSystemWallet(walletDirectory)

        val userIdentity = wallet[userId] as X509Identity
        val userinfo : User = object : User {
            override fun getName(): String {
                return userId
            }

            override fun getRoles(): Set<String> {
                return setOf("")
            }

            override fun getAccount(): String {
                return ""
            }

            override fun getAffiliation(): String {
                return hfSettings.affiliation
            }

            override fun getEnrollment(): Enrollment {
                return object : Enrollment {
                    override fun getKey(): PrivateKey {
                        return userIdentity.privateKey
                    }

                    override fun getCert(): String {
                        return Identities.toPemString(userIdentity.certificate)
                    }
                }
            }

            override fun getMspId(): String {
                return hfSettings.orgMSP
            }
        }

        val enrollment = caClient.reenroll(userinfo)
        val user: Identity = Identities.newX509Identity(hfSettings.orgMSP, enrollment)
        wallet.put(userId, user)
        return "Successfully reenrolled user" + userId + "  and imported it into the wallet"
    }



}
