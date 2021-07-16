# LUKSO Orchestrator
Orchestrating the clients to dance to the drum


# cmd/celebrimbor
This binary will also support only some of the possible networks.
It will allow you to spin every client in one command.
Make a pull request to attach your network.
We are also very open to any improvements. Please make some issue or hackmd proposal to make it better.
Join our lukso discord https://discord.gg/E2rJPP4 to ask some questions

## command to join ephemeral proxima centauri network
Dont forget to change `--pandora-etherbase` !! Change it with one of your ECDSA metamask public keys
```shell
./celebrimbor --pandora-etherbase 0xE6D95f736b2e89B9b6062EF7c39ea740B4801D85 --vanguard-grpc-endpoint=127.0.0.1:4000 --http --http.addr=0.0.0.0 --http.port=7877 --ws --ws.addr=0.0.0.0 --ws.port=7878 --pandora-rpc-endpoint=./pandora/geth.ipc --verbosity=trace --accept-terms-of-use
```



## help
```shell
./celebrimbor --h
```

## Importing wallet keys via celebrimbor
There are two ways. Via flag, or via default.
Default:
copy your `validator_keys` to `./validator_keys` and create file `./password.txt` with your password filled in the file.


## Helpful
https://github.com/ethereum/eth2.0-deposit-cli -> deposit cli
