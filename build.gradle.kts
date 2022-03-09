import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import org.jetbrains.kotlin.kapt3.base.Kapt.kapt

buildscript {
	repositories {
		jcenter()
		maven { url = uri("https://repo1.maven.org/maven2") }
		maven { url = uri("https://repo1.maven.org/maven2") }
		maven { url = uri("https://repo.spring.io/snapshot") }
		maven { url = uri("https://repo.spring.io/milestone") }
		maven { url = uri("https://repo1.maven.org/maven2") }
		maven{
			url = uri("https://gitlab.com/api/v4/projects/26939001/packages/maven")
			name = ("ubiquity-client-gitLab")
		}
	}
	dependencies {
		classpath("org.springframework.boot:spring-boot-gradle-plugin:2.2.0.M3")
		classpath("org.jetbrains.kotlin:kotlin-gradle-plugin:1.4.0")
	}
}



plugins {
	id("org.springframework.boot") version "2.4.4"
	id("io.spring.dependency-management") version "1.0.11.RELEASE"
	id("org.jetbrains.kotlin.plugin.jpa") version "1.3.30"


	id("io.freefair.lombok") version "6.0.0-m2"
	kotlin("jvm") version "1.4.32"
	kotlin("plugin.spring") version "1.4.32"
	kotlin("plugin.allopen") version "1.4.32"
	kotlin("kapt") version "1.4.32"
	kotlin("plugin.serialization") version "1.4.32"
}

group = "com.example"
version = "0.0.1-SNAPSHOT"
java.sourceCompatibility = JavaVersion.VERSION_11

repositories {
	mavenCentral(	)
	maven{
		url = uri("https://gitlab.com/api/v4/projects/26939001/packages/maven")
		name = ("ubiquity-client-gitLab")
	}

}


dependencies {
	implementation("org.web3j:core:4.8.4")
	implementation("org.springframework.boot:spring-boot-starter-mustache")
	implementation("org.springframework.boot:spring-boot-starter-web")
	implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
	implementation("org.apache.httpcomponents:httpclient")
	implementation("org.springframework.boot:spring-boot-starter-data-jpa")

	implementation("org.jetbrains.kotlin:kotlin-reflect")
	implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
	implementation(group= "com.gitlab.blockdaemon", name= "ubiquity-client", version= "0.0.2")
	implementation("org.hyperledger.fabric:fabric-gateway-java:2.1.0")
	runtimeOnly("com.h2database:h2")
	runtimeOnly("org.springframework.boot:spring-boot-devtools")
	kapt("org.springframework.boot:spring-boot-configuration-processor")
	testImplementation("org.springframework.boot:spring-boot-starter-test") {
		exclude(group = "org.junit.vintage", module = "junit-vintage-engine")
		exclude(module = "mockito-core")
	}
	testImplementation("org.junit.jupiter:junit-jupiter-api")
	testRuntimeOnly("org.junit.jupiter:junit-jupiter-engine")
	testImplementation("com.ninja-squad:springmockk:3.0.1")
	val kotlinxCoroutinesVersion = "1.2.1"
	implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:$kotlinxCoroutinesVersion")
	implementation("org.reactivestreams:reactive-streams")
	implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor:$kotlinxCoroutinesVersion")
	implementation("org.jetbrains.kotlinx:kotlinx-coroutines-rx2:$kotlinxCoroutinesVersion")
	implementation("org.jetbrains.kotlinx:kotlinx-serialization-core:1.0.1")
	implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.0.1")
	implementation("com.jakewharton.retrofit:retrofit2-kotlinx-serialization-converter:0.7.0")
	runtimeOnly("mysql:mysql-connector-java")
	implementation("com.squareup.okhttp3:okhttp:4.3.1")
	implementation("com.google.code.gson:gson:2.8.5")
	implementation("com.fasterxml.jackson.dataformat:jackson-dataformat-yaml:2.1.2")
	implementation("org.springframework.boot:spring-boot-starter-actuator:2.1.8.RELEASE")
	implementation("com.google.oauth-client:google-oauth-client-assembly:1.20.0")
	implementation( "com.gruelbox", "kucoin-java-sdk", "2.0.0")

	implementation("io.github.serg-maximchuk", "binance-api-client",  "1.0.4")
	implementation("org.springframework.boot:spring-boot-starter-websocket")
	implementation("io.ktor:ktor-client-websockets:1.4.2")
	implementation("io.ktor:ktor-client-cio:1.4.2")
	implementation("com.github.docker-java","docker-java", "3.2.12")


}

tasks.withType<KotlinCompile> {
	kotlinOptions {
		freeCompilerArgs = listOf("-Xjsr305=strict")
		jvmTarget = "1.8"
	}
}

allOpen {
	annotation("javax.persistence.Entity")
	annotation("javax.persistence.Embeddable")
	annotation("javax.persistence.MappedSuperclass")
}


tasks.withType<Test> {
	useJUnitPlatform()
}
