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
```







