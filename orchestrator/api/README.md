### api

This package can contain all the protobuf api's definition for communicate with beaconchain, validator, and catalyst like `beacon_chain.proto`, `node.proto` etc. After getting any request from beaconchain, validator, or catalyst, this package will receive the request and dispatch the request to the proper handler.