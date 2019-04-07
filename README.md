# Art Galleries Blockchain

Art-galleries-blockchain is workshop for testing the power of Blockchain, using Hyperledger Fabric, Docker and Go.

# New Features!

  - Modify your own chaincode using Go

### Installation

Art-galleries-blockchain requires [Hyperledger Fabric(fabric-samples)](https://github.com/hyperledger/fabric-samples), [Docker and Docker-Compose](https://www.docker.com/) and [Golang](https://golang.org/doc/install)  to run.

Install the dependencies

```sh
$ git clone https://github.com/hyperledger/fabric-samples
$ cd fabric-samples
$ git checkout v1.1.0
$ curl -sSL https://goo.gl/6wtTN5 | bash -s 1.1.0
```


Let's start...

```sh
$ mkdir channel-artifacts
$ ../bin/cryptogen generate --config=./crypto-config.yaml
$ export FABRIC_CFG_PATH=$PWD
$ ../bin/configtxgen -profile ArtgalleriesOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
$ export CHANNEL_NAME=artgallerieschannel
$ ../bin/configtxgen -profile ArtgalleriesChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
$ ../bin/configtxgen -profile ArtgalleriesChannel -outputAnchorPeersUpdate ./channel-artifacts/LouvreMSPanchors.tx -channelID $CHANNEL_NAME -asOrg LouvreMSP
$ ../bin/configtxgen -profile ArtgalleriesChannel -outputAnchorPeersUpdate ./channel-artifacts/GuggenheimMSPanchors.tx -channelID $CHANNEL_NAME -asOrg GuggenheimMSP
```

Let's run the containers

```sh
$ export IMAGE_TAG=latest
$ docker-compose -f docker-compose-cli.yaml down --volumes --remove-orphans
$ docker-compose -f docker-compose-cli.yaml -f docker-compose-couch.yaml up -d
```

Now we make the connections with the cli container

```sh
$ docker exec -it cli bash
$ export CHANNEL_NAME=artgallerieschannel
$ peer channel create -o orderer.artgalleries.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/artgalleries.com/orderers/orderer.artgalleries.com/msp/tlscacerts/tlsca.artgalleries.com-cert.pem
$ peer channel join -b artgalleries.block
$CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/guggenheim.artgalleries.com/users/Admin@guggenheim.artgalleries.com/msp CORE_PEER_ADDRESS=peer0.guggenheim.artgalleries.com:7051 CORE_PEER_LOCALMSPID="GuggenheimMSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/guggenheim.artgalleries.com/peers/peer0.guggenheim.artgalleries.com/tls/ca.crt peer channel join -b artgallerieschannel.block
$ peer channel update -o orderer.artgalleries.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/LouvreMSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/artgalleries.com/orderers/orderer.artgalleries.com/msp/tlscacerts/tlsca.artgalleries.com-cert.pem
$ CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/guggenheim.artgalleries.com/users/Admin@guggenheim.artgalleries.com/msp CORE_PEER_ADDRESS=peer0.guggenheim.artgalleries.com:7051 CORE_PEER_LOCALMSPID="GuggenheimMSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/guggenheim.artgalleries.com/peers/peer0.guggenheim.artgalleries.com/tls/ca.crt peer channel update -o orderer.artgalleries.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/GuggenheimMSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/artgalleries.com/orderers/orderer.artgalleries.com/msp/tlscacerts/tlsca.artgalleries.com-cert.pem


```







