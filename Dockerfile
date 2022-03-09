FROM openjdk:11

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      ca-certificates \
      curl \
      gnupg \
      lsb-release \
      apt-transport-https \
      software-properties-common \
      docker-compose \
      docker.io

ENV APPLICATION_USER ktor
RUN adduser -disabled-password -gecos '' $APPLICATION_USER

RUN mkdir -p /app/wallet
RUN chown -R $APPLICATION_USER /app/wallet
RUN chown -R $APPLICATION_USER /app
COPY ./src/main/resources /app/src/main/resources
RUN chown -R $APPLICATION_USER /app/src/main/resources/
RUN mkdir -p /app/pemWallet
COPY ./pemWallet/ca.org1.example.com-cert.pem /app/pemWallet/
RUN chown -R $APPLICATION_USER /app/pemWallet/
COPY ./build /app/build 
COPY ./gocrypto_config.json /app/config.json

USER $APPLICATION_USER

COPY ./build/libs/blog-0.0.1-SNAPSHOT.jar /app/kotlin-server.jar

WORKDIR /app

CMD ["java", "-Dsun.net.http.allowRestrictedHeaders=true", "-Dcom.sun.jndi.ldap.object.disableEndpointIdentification=true", "-Djava.net.preferIPv4Stack=true", "-jar" , "kotlin-server.jar"]

