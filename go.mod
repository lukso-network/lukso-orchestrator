module github.com/lukso-network/lukso-orchestrator

go 1.16

require (
	github.com/ethereum/go-ethereum v1.10.2
	github.com/gogo/protobuf v1.3.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/prysmaticlabs/eth2-types v0.0.0-20210303084904-c9735a06829d
	github.com/prysmaticlabs/ethereumapis v0.0.0-20210323030846-26f696aa0cbc
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli/v2 v2.3.0
	google.golang.org/grpc v1.37.0
)

replace github.com/prysmaticlabs/ethereumapis => github.com/lukso-network/vanguard-apis v0.0.0-20210331083856-a569864eb9aa
