module github.com/lukso-network/lukso-orchestrator

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/d4l3k/messagediff v1.2.1
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/dgraph-io/ristretto v0.0.4-0.20210318174700-74754f61e018
	github.com/ethereum/go-ethereum v1.10.2
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.5.0
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/joonix/log v0.0.0-20200409080653-9c1d2ceb5f1d
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/prysmaticlabs/eth2-types v0.0.0-20210303084904-c9735a06829d
	github.com/prysmaticlabs/ethereumapis v0.0.0-20210323030846-26f696aa0cbc
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210202205921-7fcea7c45dc8 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/wercker/journalhook v0.0.0-20180428041537-5d0a5ae867b3
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20201026171402-d4b8fe4fd877 // indirect
	google.golang.org/grpc v1.37.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/prysmaticlabs/ethereumapis => github.com/lukso-network/vanguard-apis v0.0.0-20210511094517-7ef8f431075a
