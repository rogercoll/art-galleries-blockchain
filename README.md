#Art Galleries Blockchain

[![N|Solid](https://cldup.com/dTxpPi9lDf.thumb.png)](https://nodesource.com/products/nsolid)

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
$ docker-compose -f docker-compose-cli.yaml -f docker-compose-couch.yaml up -d
```







