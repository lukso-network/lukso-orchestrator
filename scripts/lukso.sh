#### Lukso Deployment script for linux and mac ####
# PLEASE change the git tag acording to the latest release #

export GIT_PANDORA="v0.0.0-alpha.fix"
export GIT_VANGUARD="v0.0.0-alpha.fix"
export GIT_ORCH="v0.0.0-alpha.1"

export OS_NAME=$(uname -s)

#set up the external IP first
export EXTERNAL_IP=$(curl ident.me)


function get_necessary_dependencies {
	sudo apt update &&
	yes | sudo apt install wget
}

################################# BOOT NODE MANAGEMENT ############################################################

function download_and_run_pandora_bootnode {
	# download bootnode
	mkdir -p ./bin
	mkdir -p ./bootnode_pandora

	wget https://storage.googleapis.com/lukso-assets/bootnode/pandora_bootnode/bootnode -O ./bin/bootnode
	chmod +x ./bin/bootnode

	#run bootnode
	echo "####  Strating pandora bootnode ####"
    nohup ./bin/bootnode \
    -addr=":45451" \
    -nodekeyhex=d1b8543a818563d1ecec4ef6e25268fe4bf210777807e63235160ee0ffc3ca78 \
    -nat=extip:$EXTERNAL_IP \
    -verbosity=4 > ./bootnode_pandora/bootnode.txt 2>&1 &
    disown
}

function download_and_run_vanguard_bootnode {
	# download bootnode
	mkdir -p ./bin
	mkdir -p ./bootnode_vanguard

	wget https://storage.googleapis.com/lukso-assets/bootnode/bootnode -O ./bin/vbootnode
	chmod +x ./bin/vbootnode

	echo "####  Strating vanguard bootnode ####"
    nohup ./bin/vbootnode \
    --discv5-port=4841 \
    --private=1083c5f7f50aefc0c2ff0d7020316d243639ecc266ef3c03b75fcf126abb811e \
    --external-ip=$EXTERNAL_IP \
    --fork-version=83a55317 \
    --log-file=./bootnode_vanguard/bootnode.txt \
    --debug > ./bootnode_vanguard/bootnode.txt 2>&1 &
    disown
}

function stop_bootnodes {
	echo "### Stopping pandora bootnode ####"
    sudo kill -9 $(sudo lsof -t -i:30301) &> /dev/null
    sleep 1


    echo "### Stopping vanguard bootnode ###"
    sudo kill -9 $(sudo lsof -t -i:4001) &> /dev/null
    sleep 1
}

########################################################################################################################

########################################## ORCHESTRATOR MANAGEMENT #####################################################


function download_orchestrator_binary {

	mkdir -p ./bin

	GIT_TAG=$1
	if [[ -z $1 ]]; then
		#statements
		GIT_TAG=$GIT_ORCH
	fi

	arch=$(uname -m)

	wget https://github.com/lukso-network/lukso-orchestrator/releases/download/"$GIT_TAG"/lukso-orchestrator_"$GIT_TAG"_"$OS_NAME"_"$arch" -O ./bin/orchestrator
	chmod +x ./bin/orchestrator
}


function run_orchestrator {
	# download create directory if not exists
	mkdir -p ./orchestrator

	# download orchestrator
	if [[ ! -f ./bin/orchestrator ]]; then
    	download_orchestrator_binary $1
	fi


	nohup ./bin/orchestrator --datadir=./orchestrator/datadir \
		--vanguard-grpc-endpoint=127.0.0.1:4000 \
		--http \
		--http.addr=0.0.0.0 \
		--http.port=7877 \
		--ws \
		--ws.addr=0.0.0.0 \
		--ws.port=7878 \
		--pandora-rpc-endpoint=ws://127.0.0.1:8546 \
		--verbosity=trace > ./orchestrator/orchestrator.log  2>&1 &
	disown
}


function stop_orchestrator {
	echo "###### Stopping orchestrator client ######"
	sudo kill $(sudo lsof -t -i:7877) &> /dev/null
	sleep 1
}

function clear_orchestrator {
	rm -rf ./orchestrator/datadir
}

#############################################################################################################################

############################################# PANDORA MANAGEMENT ############################################################



function download_pandora_binary {

	GIT_TAG=$1
	if [[ -z $1 ]]; then
		#statements
		GIT_TAG=$GIT_PANDORA
	fi

	arch=$(uname -m)

	mkdir -p ./bin

	wget https://github.com/lukso-network/pandora-execution-engine/releases/download/"$GIT_TAG"/pandora-"$OS_NAME"-"$arch" -O ./bin/pandora
	chmod +x ./bin/pandora
}

function download_pandora_genesis {
	mkdir -p ./config
	wget https://storage.googleapis.com/lukso-assets/config/pandora_genesis_staging.json -O ./config/pandora-genesis.json
}

function download_static_node_info {
	mkdir -p ./pandora/datadir/geth
	wget https://storage.googleapis.com/lukso-assets/config/static-nodes.json -O ./pandora/datadir/geth/static-nodes.json
}


function run_pandora {

	mkdir -p ./pandora

	if [[ ! -f ./config/pandora-genesis.json ]]; then
	  download_pandora_genesis
  	fi

  	if [[ ! -f ./bin/pandora ]]; then
	  download_pandora_binary $1
  	fi

  	if [[ ! -f ./pandora/datadir/geth/static-nodes.json ]]; then
	  download_static_node_info
  	fi

	./bin/pandora --datadir ./pandora/datadir init ./config/pandora-genesis.json &> /dev/null
	nohup ./bin/pandora --datadir=./pandora/datadir \
	 --networkid=808080 \
	 --ethstats=l15-$(hostname):VIyf7EjWlR48@stats.pandora.l15.lukso.network \
	 --port=30405 \
	 --rpc \
	 --rpcaddr=0.0.0.0 \
	 --rpcapi=admin,net,eth,debug,ethash,miner,personal,txpool,web3 \
	 --bootnodes=enode://4a6ab64c08eca2fd3a96d285b5a2db918f26220eb6b18842ce40b49354198f24eb0961b16c4d552c7318050ce2c7bcd30ef5ca2f9826811fb24b37e3bb07121f@35.234.122.88:45451,enode://4a6ab64c08eca2fd3a96d285b5a2db918f26220eb6b18842ce40b49354198f24eb0961b16c4d552c7318050ce2c7bcd30ef5ca2f9826811fb24b37e3bb07121f@34.141.112.243:45451,enode://4a6ab64c08eca2fd3a96d285b5a2db918f26220eb6b18842ce40b49354198f24eb0961b16c4d552c7318050ce2c7bcd30ef5ca2f9826811fb24b37e3bb07121f@35.198.170.46:45451,enode://4a6ab64c08eca2fd3a96d285b5a2db918f26220eb6b18842ce40b49354198f24eb0961b16c4d552c7318050ce2c7bcd30ef5ca2f9826811fb24b37e3bb07121f@34.141.46.202:45451 \
	 --rpcport=8545 \
	 --http.corsdomain="*" \
	 --ws \
	 --ws.addr=0.0.0.0 \
	 --ws.api=admin,net,eth,debug,ethash,miner,personal,txpool,web3 \
	 --ws.port=8546 \
	 --ws.origins="*" \
	 --mine \
	 --miner.notify=ws://127.0.0.1:7878,http://127.0.0.1:7877 \
	 --miner.etherbase=91b382af07767Bdab2569665AC30125E978a0688 \
	 --nat=extip:$EXTERNAL_IP \
	 --syncmode="full" \
   --allow-insecure-unlock \
   -nat=extip:$EXTERNAL_IP \
	 --verbosity=4 > ./pandora/pandora.log  2>&1 &
	 disown
}

function stop_pandora {
	echo "Stopping pandora node..."

	sudo kill $(sudo lsof -t -i:30405) &> /dev/null
	sleep 1
}

function clear_pandora {
	# remove everything but static-nodes.json. it holds static node info
	find ./pandora/datadir ! -name 'static-nodes.json' -type f -exec rm -f {} +
}

######################################################################################################################################



##################################################### Vanguard Management ############################################################

function download_vanguard_binary {
	GIT_TAG=$1
	if [[ -z $1 ]]; then
		#statements
		GIT_TAG=$GIT_VANGUARD
	fi

	arch=$(uname -m)

	mkdir -p ./bin
	wget https://github.com/lukso-network/vanguard-consensus-engine/releases/download/"$GIT_TAG"/beacon-chain-"$OS_NAME"-"$arch" -O ./bin/beacon-chain
	chmod +x ./bin/beacon-chain
}


function download_vanguard_genesis {
	mkdir -p ./config
	wget https://storage.googleapis.com/lukso-assets/config/vanguard_genesis_staging.ssz -O ./config/vanguard_genesis.ssz
}

function download_vanguard_config {
	mkdir -p ./config
	wget https://storage.googleapis.com/lukso-assets/config/vanguard_config_staging.yml -O ./config/vanguard-config.yml
}

function download_network_priv_key {
	wget https://storage.googleapis.com/lukso-assets/config/"$1" -O ./network-keys
}


function run_vanguard {

	mkdir -p ./vanguard

	if [[ ! -f ./config/vanguard_genesis.ssz ]]; then
		download_vanguard_genesis
  	fi

  	if [[ ! -f ./config/vanguard-config.yml ]]; then
  		download_vanguard_config
  	fi
	
	if [[ ! -f ./bin/beacon-chain ]]; then
		download_vanguard_binary $1
  	fi

	nohup ./bin/beacon-chain \
	      --accept-terms-of-use \
	      --chain-id=808080 \
	      --network-id=808080 \
	      --genesis-state=./config/vanguard_genesis.ssz \
	      --datadir=./vanguard/datadir \
	      --chain-config-file=./config/vanguard-config.yml \
	      --bootstrap-node="enr:-Ku4QEL0I7H3EawRwc2ZUevmj-_T0R6JZGMhfp_2KHBlwAt5bwA19c8LSYZzy63EvpsYbifKye6qnE-_vsNimWOz8scBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCPqeliJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk" \
	      --bootstrap-node="enr:-Ku4QAmYtwrQBZ-WJwTPL4xMpTO6BlZcU6IuXljtd_SgC51nGRs98WvxCX0-ZJBs0G9m9tcFPsktbdSr7EliMhrZnfEBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCKNcPOJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk" \
	      --bootstrap-node="enr:-Ku4QEXRrSXB7od-xNeoLuq6GicTHpuuCNRPPR9tM48Iai0-FoHL4JsntmpnwUrC-di-lT6gkbxV7Jikg9s6ImsAT1oBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCPGqi6Jc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk" \
	      --bootstrap-node="enr:-Ku4QBuS5wqvF6SHaPpuu4r4ZlRRVC1Ojp1zDOAVC1X0PB3gRujAhWZdk2m0kn3FwoPuHft_Sku0tWHSfBVlHoER160Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCKNLsqJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk" \
	      --http-web3provider=http://127.0.0.1:8545 \
	      --deposit-contract=0x000000000000000000000000000000000000cafe \
	      --contract-deployment-block=0 \
	      --rpc-host=0.0.0.0 \
	      --monitoring-host=0.0.0.0 \
	      --verbosity=debug \
	      --min-sync-peers=2 \
	      --p2p-max-peers=50 \
	      --orc-http-provider=http://127.0.0.1:7877 \
	      --p2p-host-ip=$EXTERNAL_IP \
	      --rpc-port=4000 \
	      --p2p-udp-port=12000 \
	      --p2p-tcp-port=13000 \
	      --grpc-gateway-port=3500 \
	      --update-head-timely \
	      --log-file=./vanguard/vanguard.log \
	      --lukso-network > ./vanguard/vanguard.log  2>&1 &
	    disown
}

function stop_vanguard {
	echo "###### Stopping vanguard #######"

	sudo kill $(sudo lsof -t -i:4000) &> /dev/null
	sleep 1
	sudo kill $(sudo lsof -t -i:12000) &> /dev/null
	sleep 1
	sudo kill $(sudo lsof -t -i:13000) &> /dev/null
	sleep 1
}

function clear_vanguard {
	rm -rf ./vanguard/datadir
}

#######################################################################################################################################


################################################## Validator Management ###############################################################

function download_validator_binary {
	GIT_TAG=$1
	if [[ -z $1 ]]; then
		#statements
		GIT_TAG=$GIT_VANGUARD
	fi

	arch=$(uname -m)
	mkdir -p ./bin

	wget https://github.com/lukso-network/vanguard-consensus-engine/releases/download/"$GIT_TAG"/validator-"$OS_NAME"-"$arch" -O ./bin/validator
	chmod +x ./bin/validator
}


function download_validator_password {
	echo "downloading default validator password ..."
	mkdir -p ./validator
	wget https://storage.googleapis.com/lukso-assets/config/password.txt -O ./validator/password.txt	
}

function create_validator_password {
	echo "importing validator password ..."
	mkdir -p ./validator
	echo "$1" > ./validator/password.txt	
}

function run_validator {

	mkdir -p ./validator
	
	if [[ ! -f ./bin/validator ]]; then
		download_validator_binary $1
	fi

	if [[ ! -f ./validator/password.txt ]]; then
		download_validator_password
	fi

	if [[ ! -f ./config/vanguard-config.yml ]]; then
		download_vanguard_config
	fi	

	echo "##### Starting validator node #####"
	nohup ./bin/validator \
	  --datadir=./validator/datadir \
	  --accept-terms-of-use \
	  --beacon-rpc-provider=localhost:4000 \
	  --chain-config-file=./config/vanguard-config.yml \
	  --verbosity=debug \
	  --pandora-http-provider=http://127.0.0.1:8545 \
	  --wallet-dir=./validator/wallet \
	  --wallet-password-file=./validator/password.txt \
	  --rpc \
	  --log-file=./validator/validator.log \
	  --lukso-network > ./validator/validator.log  2>&1 &
	disown

}

function stop_validator {
	echo "###### Stopping validator client ######"
	sudo kill $(sudo lsof -t -i:7000) &> /dev/null
	sleep 1
}

function clear_validator {
	rm -rf ./validator/datadir
}



#######################################################################################################################################

function find_and_kill {
	process_id=` /bin/ps -fu $USER| grep "$1" | grep -v "grep" | awk '{print $2}' `
	for process in $process_id
	do
		kill -INT $process
	done

}

function download_validator_keys {
	mkdir -p ./validator
	wget  https://storage.googleapis.com/lukso-assets/key_files/"$1" -O ./validator/"$1" &&
	tar -xvf ./validator/"$1" -C ./validator/
}

function import_account {

	mkdir -p ./validator

	if [[ ! -f ./bin/validator ]]; then
		download_validator_binary
	fi

	if [[ ! -f ./validator/"$2" ]]; then
		download_validator_keys $1
	fi

	if [[ ! -f ./validator/password.txt ]]; then
		download_validator_password
	fi

	echo "### Importing accounts"
	./bin/validator accounts import --keys-dir=./validator/keys \
	--wallet-dir=./validator/wallet \
	--wallet-password-file=./validator/password.txt
}


function download_all_binaries {
	download_orchestrator_binary $1
	download_validator_binary $1
	download_vanguard_binary $1
	download_pandora_binary $1
}

function stop_all {
	find_and_kill "pandora"
	find_and_kill "validator"
	find_and_kill "beacon-chain"
	find_and_kill "orchestrator"
	rm -f ./vanguard/datadir/network-keys
}

function reset_all {
	stop_all
	clear_pandora
	clear_validator
	clear_vanguard
	clear_orchestrator
}

function run_full_set {
	echo "################# Running orchestrator #######################"
	run_orchestrator $1
	sleep 2
	echo "################# Running pandora #######################"
	run_pandora $1
	sleep 2
	echo "################# Running vanguard #######################"
	run_vanguard $1
	sleep 2
	echo "################# Running validator #######################"
	run_validator $1
	sleep 2
}


function run_arch_node {
	echo "################# Running orchestrator #######################"
	run_orchestrator $1
	sleep 2
	echo "################# Running pandora #######################"
	run_pandora $1
	sleep 2
	echo "################# Running vanguard #######################"
	run_vanguard $1
	sleep 2
}

function reset_wallet {
	rm -rf ./validator/keys/
	rm -rf ./validator/wallet/
}

function log_accumulator {
	rm -f all_logs.tar.gz
	mkdir -p ./all_logs
	cp ./pandora/pandora.log ./all_logs
	cp ./orchestrator/orchestrator.log ./all_logs
	cp ./vanguard/vanguard.log ./all_logs
	cp ./validator/validator.log ./all_logs

	tar -czvf all_logs.tar.gz ./all_logs
	rm -rf ./all_logs
}

while test $# -gt 0; do
  case "$1" in
    -h|--help)
      echo "This installer will attempt to run pandora nodes in your setup"
      echo " "
      echo " [options] application [arguments]"
      echo ""
      echo "options:"
      echo "-h, --help          show brief help"

      echo ""
      echo "reset options:"
      echo "--reset		stops pandora, vanguard, validator, orchestrator and removes their databases"
      echo "--reset_wallet	removes key files and wallet directory"
      echo "--reset_pandora	stops pandora and removes the datadir"
      echo "--reset_vanguard	stops vanguard and removes the datadir"
      echo "--reset_validator	stops validator and removes the datadir"
      echo "--reset_orchestrator	stops orchestrator and removes the datadir"


      echo ""
      echo "new download options:"
      echo "--download_binaries=GIT_TAG	download freash binaries for validator, vanguard, pandora, orchestrator"
      echo "--download_pandora_binary=GIT_TAG	download pandora binary for a specific git tag"
      echo "--download_vanguard_binary=GIT_TAG	download vanguard binary for a specific git tag"
      echo "--download_validator_binary=GIT_TAG	download validator binary for a specific git tag"
      echo "--download_orchestrator_binary=GIT_TAG	download orchestrator binary for a specific git tag"
      echo "--download_net_key=KEY_FILE	download network private key file"
      echo "--download_static_node download static files to connect with pandora"


      echo ""
      echo "account import options:"
      echo "--create_password=PASSWORD 	create a password.txt file with PASSWORD for wallet usage"
      echo "--import_account=GIT_TAG	imports account with the validator GIT_TAG [must put --keyfile_name flag]"
      echo "--keyfile_name=KEY_FILE 	key file name from google cloud. contact with the owner if you don't know yours [will go with --import_account]"

      echo ""
      echo "run options"
      echo "--run_bootnode 	run pandora and vanguard bootnodes by downloading the bin"
      echo "--run 	run pandora with $GIT_PANDORA, vanguard with $GIT_VANGUARD, validator with $GIT_VANGUARD and orchestrator with $GIT_ORCH"
      echo "--run_arch run 	pandora with $GIT_PANDORA, vanguard with $GIT_VANGUARD and orchestrator with $GIT_ORCH" 
      echo "--l15=GIT_TAG 	run all l15 lukso nodes"
      echo "--arch=GIT_TAG 	run archieve node the GIT_TAG"
      echo "--run_orchestrator=GIT_TAG 	run orchestrator with the GIT_TAG"
      echo "--run_pandora=GIT_TAG 	run pandora with the GIT_TAG"
      echo "--run_vanguard=GIT_TAG 	run vanguard with the GIT_TAG"
      echo "--run_validator=GIT_TAG 	run validator with the GIT_TAG"


      echo ""
      echo "stop options"
      echo "--stop_bootnode 	stop running bootnodes"
      echo "--stop_pandora 	stop pandora node"
      echo "--stop_orchestrator 	stop pandora node"
      echo "--stop_vanguard	stop pandora node"
      echo "--stop_validator	stop pandora node"
      echo "--stop 	stop all node"
      echo "--log 	log everything"
      echo "--prep 	download depenencies"
      exit 0
      ;;
    --reset)
      reset_all
      exit 0
      ;;
    --reset_wallet)
	  reset_wallet
	  exit 0
      ;;
    --reset_pandora)
		stop_pandora
		clear_pandora
		shift
		;;
    --reset_vanguard)
		stop_vanguard
    	clear_vanguard
    	shift
    	;;
    --reset_validator)
		stop_validator
    	clear_validator
    	shift
    	;;

   	--reset_orchestrator)
		stop_orchestrator
    	clear_orchestrator
    	shift
    	;;

    --download_net_key*)
	  export keyfileName=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_network_priv_key $keyfileName
      shift
	  ;;
    --download_binaries*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_all_binaries $GIT_VERSION
      shift
	  ;;

	--download_pandora_binary*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_pandora_binary $GIT_VERSION
      shift
	  ;;

	--download_vanguard_binary*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_vanguard_binary $GIT_VERSION
      shift
	  ;;
	--download_validator_binary*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_validator_binary $GIT_VERSION
      shift
	  ;;

	--download_orchestrator_binary*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
      download_orchestrator_binary $GIT_VERSION
      shift
	  ;;
	--download_static_node)
		download_static_node_info
		shift
		;;  
    --create_password*)
	  export PASSWORD=`echo $1 | sed -e 's/^[^=]*=//g'`
      create_validator_password $PASSWORD
      exit 0
      ;;

    --import_account*)
		export KEY_FILE=`echo $1 | sed -e 's/^[^=]*=//g'`
		if [ -z ${KEY_FILE+x} ] 
		then
			echo "--keyfile_name flag is not present"
			exit 1
		else
      		import_account $KEY_FILE
		fi
      	shift
      ;;
    --keyfile_name*)
		export KEY_FILE=`echo $1 | sed -e 's/^[^=]*=//g'`
		echo "$KEY_FILE new key file"
		shift
		;;

    --run_bootnode)
      download_and_run_pandora_bootnode
      download_and_run_vanguard_bootnode
      shift
      ;;

    --run)
		echo "running pandora with $GIT_PANDORA, vanguard with $GIT_VANGUARD, validator with $GIT_VANGUARD and orchestrator with $GIT_ORCH"
		run_full_set
		shift
		;;
	--run_arch)
		echo "running pandora with $GIT_PANDORA, vanguard with $GIT_VANGUARD and orchestrator with $GIT_ORCH"
		run_arch_node
		shift
		;;

    --run_orchestrator*)
      export GIT_TAG=`echo $1 | sed -e 's/^[^=]*=//g'`
      run_orchestrator $GIT_TAG
      shift
      ;;
    --run_default_orchestrator)
      run_orchestrator
      shift
      ;;
	--run_pandora*)
	  export GIT_TAG=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_pandora $GIT_TAG
	  shift
	  ;;
	--run_default_pandora)
	  export GIT_TAG=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_pandora
	  shift
	  ;;
	--run_vanguard*)
	  export GIT_TAG=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_vanguard $GIT_TAG
	  shift
	  ;;
	--run_default_vanguard)
	  run_vanguard
	  shift
	  ;;
	--run_validator*)
	  export GIT_TAG=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_validator $GIT_TAG
	  shift
	  ;;
	--run_default_validator)
	  run_validator
	  shift
	  ;;
	--l15*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_full_set $GIT_VERSION
	  shift
	  ;;
	--arch*)
	  export GIT_VERSION=`echo $1 | sed -e 's/^[^=]*=//g'`
	  run_arch_node $GIT_VERSION
	  shift
	  ;;
	--stop_pandora)
	  stop_pandora
	  shift
	  ;;
	--stop_orchestrator)
	  stop_orchestrator
	  shift
	  ;;
	--stop_validator)
	  stop_validator
	  shift
	  ;;
	--stop_vanguard)
	  stop_vanguard
	  shift
	  ;;
	--stop_bootnode)
      stop_bootnodes
      shift
      ;;
	--stop)
	  stop_all
	  shift
	  ;;
	--log)
	  log_accumulator
	  shift
	  ;;
    --prep)
	  get_necessary_dependencies
	  shift
	  ;;
    *)
      break
      ;;
  esac
done
