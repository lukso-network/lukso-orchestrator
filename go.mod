module github.com/lukso-network/lukso-orchestrator

go 1.16

require (
	github.com/boltdb/bolt v1.3.1
	github.com/d4l3k/messagediff v1.2.1
	github.com/dgraph-io/ristretto v0.0.4-0.20210318174700-74754f61e018
	github.com/ethereum/go-ethereum v1.10.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/joonix/log v0.0.0-20200409080653-9c1d2ceb5f1d
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/pkg/errors v0.9.1
	github.com/prysmaticlabs/eth2-types v0.0.0-20210303084904-c9735a06829d
	github.com/prysmaticlabs/prysm v1.4.4
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/wercker/journalhook v0.0.0-20180428041537-5d0a5ae867b3
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/prysmaticlabs/prysm => github.com/lukso-network/vanguard-consensus-engine v0.5.1-develop
