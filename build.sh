docker stop $(docker ps | grep java | cut -c1-8)
docker rm $(docker ps | grep java | cut -c1-8)

cp src/main/resources/application1.yaml src/main/resources/application.yaml
gradle build
docker build --no-cache -t finblockserver:latest -f ./Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --mount source=fiatwallet13,target=/app/wallet --restart always -d --network="host" -p 7080:7080 -t finblockserver:latest

cp src/main/resources/application2.yaml src/main/resources/application.yaml
gradle build
docker build --no-cache -t finblockserver:latest -f ./Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --mount source=fiatwallet13_2,target=/app/wallet --restart always -d --network="host" -p 7081:7081 -t finblockserver:latest

cp src/main/resources/application3.yaml src/main/resources/application.yaml
gradle build
docker build --no-cache -t finblockserver:latest -f ./Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --mount source=fiatwallet13_3,target=/app/wallet --restart always -d --network="host" -p 7082:7082 -t finblockserver:latest

cp src/main/resources/application4.yaml src/main/resources/application.yaml
gradle build
docker build --no-cache -t finblockserver:latest -f ./Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --mount source=fiatwallet13_4,target=/app/wallet --restart always -d --network="host" -p 7083:7083 -t finblockserver:latest

cp src/main/resources/application5.yaml src/main/resources/application.yaml
gradle build
docker build --no-cache -t finblockserver:latest -f ./Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --mount source=fiatwallet13_5,target=/app/wallet --restart always -d --network="host" -p 7084:7084 -t finblockserver:latest

