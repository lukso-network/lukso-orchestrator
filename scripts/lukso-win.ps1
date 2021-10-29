# Note: Don't put default values here.
# The hierachy must be Flag THEN ConfigFile THEN Default value
# If you set it here, they will overwrite config file

# Flags declaration
param (
    [Parameter(Position = 0, Mandatory)][String]$command,
    [Parameter(Position = 1)][String]$argument,
    [String]$deposit,
    [String]$eth2stats,
    [String]$network,
    [String]${lukso-home},
    [String]$datadir,
    [String]$logsdir,
    [String]${keys-dir},
    [String]${keys-password-file},
    [String]${wallet-dir},
    [String]${wallet-password-file},
    [Switch]${l15-prod},
    [Switch]${l15-staging},
    [Switch]${l15-dev},
    [String]$config,
    [String]$coinbase,
    [String]${node-name},
    [Switch]$validate,
    [String]$orchestrator,
    [String]${orchestrator-verbosity},
    [String]$pandora,
    [String]${pandora-bootnodes},
    [String]${pandora-http-port},
    [Switch]${pandora-metrics},
    [String]${pandora-nodekey},
    [String]${pandora-rpcvhosts},
    [String]${pandora-external-ip},
    [Switch]${pandora-universal-profile-expose},
    [Switch]${pandora-unsafe-expose},
    [String]${pandora-verbosity},
    [String]$vanguard,
    [String]${vanguard-bootnodes},
    [String]${vanguard-p2p-priv-key},
    [String]${vanguard-external-ip},
    [String]${vanguard-p2p-host-dns},
    [String]${vanguard-rpc-host},
    [String]${vanguard-monitoring-host},
    [String]${vanguard-verbosity},
    [String]$validator,
    [String]${validator-verbosity},
    [String]${cors-domain},
    [String]${external-ip},
    [Switch]${allow-respin},
    [Switch]$force
)

$platform = "Windows"
$architecture = "x86_64"
$InstallDir = $Env:APPDATA + "\LUKSO";
$RunDate = Get-Date -Format "yyyy-m-dd__HH-mm-ss"

# Parsing config File and setting defaults
if ($config)
{
    $ConfigFile = ConvertFrom-Yaml $(Get-Content -Raw $config )
}

Function pick_network($picked_network)
{
    $network = $picked_network
    if (!(Test-Path "$InstallDir\networks\$network"))
    {
        download_network_config $network
    }
    $NetworkConfigFile = "$InstallDir\networks\$network\config\network-config.yaml"
    $NetworkConfig = ConvertFrom-Yaml $(Get-Content -Raw $NetworkConfigFile)
}

$network = If ($network) {$network} ElseIf ($ConfigFile.NETWORK) {$ConfigFile.NETWORK} Else {"l15-prod"}

${l15-prod} = If (${l15-prod}) {${l15-prod}} ElseIf ($ConfigFile.L15_PROD) {$ConfigFile.L15_PROD} Else {$false}
${l15-staging} = If (${l15-staging}) {${l15-staging}} ElseIf ($ConfigFile.L15_STAGING) {$ConfigFile.L15_STAGING} Else {$false}
${l15-dev} = If (${l15-dev}) {${l15-dev}} ElseIf ($ConfigFile.L15_DEV) {$ConfigFile.L15_DEV} Else {$false}

If (${l15-prod}) {$network = "l15-prod"}
If (${l15-staging}) {$network = "l15-staging"}
If (${l15-dev}) {$network = "l15-dev"}

If ( (!${l15-prod}) -or (!${l15-staging}) -or (!${l15-dev}) ) {$network = "l15-prod"}

echo "Connecting to: $network"

If ($network) {
    $NetworkConfigFile = "$InstallDir\networks\$network\config\network-config.yaml"
    $NetworkConfig = ConvertFrom-Yaml $(Get-Content -Raw $NetworkConfigFile)
}

$deposit = If ($deposit) {$deposit} ElseIf ($ConfigFile.DEPOSIT) {$ConfigFile.DEPOSIT} Else {""}
$eth2stats = If ($eth2stats) {$eth2stats} ElseIf ($ConfigFile.ETH2STATS) {$ConfigFile.ETH2STATS} Else {""}
${lukso-home} = If (${lukso-home}) {${lukso-home}} ElseIf ($ConfigFile.LUKSO_HOME) {$ConfigFile.LUKSO_HOME} Else {"$HOME\.lukso"}
$datadir = If ($datadir) {$datadir} ElseIf ($ConfigFile.DATADIR) {$ConfigFile.DATADIR} Else {"${lukso-home}\$network\datadir"}
$logsdir = If ($logsdir) {$logsdir} ElseIf ($ConfigFile.LOGSDIR) {$ConfigFile.LOGSDIR} Else {"${lukso-home}\$network\logs"}
${keys-dir} = If (${keys-dir}) {${keys-dir}} ElseIf ($ConfigFile.KEYS_DIR) {$ConfigFile.KEYS_DIR} Else {"${lukso-home}\validator_keys"}
${keys-password-file} = If (${keys-password-file}) {${keys-password-file}} ElseIf ($ConfigFile.KEYS_PASSWORD_FILE) {$ConfigFile.KEYS_PASSWORD_FILE} Else {""}
${wallet-dir} = If (${wallet-dir}) {${wallet-dir}} ElseIf ($ConfigFile.WALLET_DIR) {$ConfigFile.WALLET_DIR} Else {"${lukso-home}\wallet"}
${wallet-password-file} = If (${wallet-password-file}) {${wallet-password-file}} ElseIf ($ConfigFile.WALLET_PASSWORD_FILE) {$ConfigFile.WALLET_PASSWORD_FILE} Else {""}

$coinbase = If ($coinbase) {$coinbase} ElseIf ($ConfigFile.COINBASE) {$ConfigFile.COINBASE} Else {""}
${node-name} = If (${node-name}) {${node-name}} ElseIf ($ConfigFile.NODE_NAME) {$ConfigFile.NODE_NAME} Else {""}
$validate = If ($validate) {$validate} ElseIf ($ConfigFile.VALIDATE) {$ConfigFile.VALIDATE} Else {$false}
$orchestrator = If ($orchestrator) {$orchestrator} ElseIf ($ConfigFile.ORCHESTRATOR) {$ConfigFile.ORCHESTRATOR} Else {""}
${orchestrator-verbosity} = If (${orchestrator-verbosity}) {${orchestrator-verbosity}} ElseIf ($ConfigFile.ORCHESTRATOR_VERBOSITY) {$ConfigFile.ORCHESTRATOR_VERBOSITY} Else {"info"}
$pandora = If ($pandora) {$pandora} ElseIf ($ConfigFile.PANDORA) {$ConfigFile.PANDORA} Else {""}
${pandora-bootnodes} = If (${pandora-bootnodes}) {${pandora-bootnodes}} ElseIf ($ConfigFile.PANDORA_BOOTNODES) {$ConfigFile.PANDORA_BOOTNODES} Else {$NetworkConfig.PANDORA_BOOTNODES}
${pandora-http-port} = If (${pandora-http-port}) {${pandora-http-port}} ElseIf ($ConfigFile.PANDORA_HTTP_PORT) {$ConfigFile.PANDORA_HTTP_PORT} Else {"8545"}
${pandora-metrics} = If (${pandora-metrics}) {${pandora-metrics}} ElseIf ($ConfigFile.PANDORA_METRICS) {$ConfigFile.PANDORA_METRICS} Else {$false}
${pandora-nodekey} = If (${pandora-nodekey}) {${pandora-nodekey}} ElseIf ($ConfigFile.PANDORA_NODEKEY) {$ConfigFile.PANDORA_NODEKEY} Else {""}
${pandora-rpcvhosts} = If (${pandora-rpcvhosts}) {${pandora-rpcvhosts}} ElseIf ($ConfigFile.PANDORA_RPCVHOSTS) {$ConfigFile.PANDORA_RPCVHOSTS} Else {""}
${pandora-external-ip} = If (${pandora-external-ip}) {${pandora-external-ip}} ElseIf ($ConfigFile.PANDORA_EXTERNAL_IP) {$ConfigFile.PANDORA_EXTERNAL_IP} Else {""}
${pandora-universal-profile-expose} = If (${pandora-universal-profile-expose}) {${pandora-universal-profile-expose}} ElseIf ($ConfigFile.PANDORA_UNIVERSAL_PROFILE_EXPOSE) {$ConfigFile.PANDORA_UNIVERSAL_PROFILE_EXPOSE} Else {$false}
${pandora-unsafe-expose} = If (${pandora-unsafe-expose}) {${pandora-unsafe-expose}} ElseIf ($ConfigFile.PANDORA_UNSAFE_EXPOSE) {$ConfigFile.PANDORA_UNSAFE_EXPOSE} Else {$false}
${pandora-verbosity} = If (${pandora-verbosity}) {${pandora-verbosity}} ElseIf ($ConfigFile.PANDORA_VERBOSITY) {$ConfigFile.PANDORA_VERBOSITY} Else {"info"}
$vanguard = If ($vanguard) {$vanguard} ElseIf ($ConfigFile.VANGUARD) {$ConfigFile.VANGUARD} Else {""}
${vanguard-bootnodes} = If (${vanguard-bootnodes}) {${vanguard-bootnodes}} ElseIf ($ConfigFile.VANGUARD_BOOTNODES) {$ConfigFile.VANGUARD_BOOTNODES} Else {$NetworkConfig.VANGUARD_BOOTNODES}
${vanguard-p2p-priv-key} = If (${vanguard-p2p-priv-key}) {${vanguard-p2p-priv-key}} ElseIf ($ConfigFile.VANGUARD_P2P_PRIV_KEY) {$ConfigFile.VANGUARD_P2P_PRIV_KEY} Else {""}
${vanguard-external-ip} = If (${vanguard-external-ip}) {${vanguard-external-ip}} ElseIf ($ConfigFile.VANGUARD_EXTERNAL_IP) {$ConfigFile.VANGUARD_EXTERNAL_IP} Else {""}
${vanguard-p2p-host-dns} = If (${vanguard-p2p-host-dns}) {${vanguard-p2p-host-dns}} ElseIf ($ConfigFile.VANGUARD_P2P_HOST_DNS) {$ConfigFile.VANGUARD_P2P_HOST_DNS} Else {""}
${vanguard-rpc-host} = If (${vanguard-rpc-host}) {${vanguard-rpc-host}} ElseIf ($ConfigFile.VANGUARD_RPC_HOST) {$ConfigFile.VANGUARD_RPC_HOST} Else {""}
${vanguard-monitoring-host} = If (${vanguard-monitoring-host}) {${vanguard-monitoring-host}} ElseIf ($ConfigFile.VANGUARD_MONITORING_HOST) {$ConfigFile.VANGUARD_MONITORING_HOST} Else {""}
${vanguard-verbosity} = If (${vanguard-verbosity}) {${vanguard-verbosity}} ElseIf ($ConfigFile.VANGUARD_VERBOSITY) {$ConfigFile.VANGUARD_VERBOSITY} Else {"info"}
$validator = If ($validator) {$validator} ElseIf ($ConfigFile.VALIDATOR) {$ConfigFile.VALIDATOR} Else {}
${validator-verbosity} = If (${validator-verbosity}) {${validator-verbosity}} ElseIf ($ConfigFile.VALIDATOR_VERBOSITY) {$ConfigFile.VALIDATOR_VERBOSITY} Else {"info"}
${cors-domain} = If (${cors-domain}) {${cors-domain}} ElseIf ($ConfigFile.CORS_DOMAIN) {$ConfigFile.CORS_DOMAIN} Else {""}
${external-ip} = If (${external-ip}) {${external-ip}} ElseIf ($ConfigFile.EXTERNAL_IP) {$ConfigFile.EXTERNAL_IP} Else {""}
${allow-respin} = If (${allow-respin}) {${allow-respin}} ElseIf ($ConfigFile.ALLOW_RESPIN) {$ConfigFile.ALLOW_RESPIN} Else {$false}
$force = If ($force) {$force} ElseIf ($ConfigFile.FORCE) {$ConfigFile.FORCE} Else {$false}







Function download($url, $dst)
{
    Write-Output $url
    Write-Output $dst
    $client = New-Object System.Net.WebClient
    $client.DownloadFile($url, $dst)
}

Function download_binary($client, $tag)
{

    switch ($client)
    {
        lukso-orchestrator {
            $repo = "lukso-orchestrator"
        }

        pandora {
            $repo = "pandora-execution-engine"
        }

        vanguard {
            $repo = "vanguard-consensus-engine"
        }

        lukso-validator {
            $repo = "vanguard-consensus-engine"
        }

        eth2stats {
            $repo = "network-vanguard-stats-client"
        }
    }

    $Target = "$InstallDir\binaries\$CLIENT\$TAG"
    New-Item -ItemType Directory -Force -Path $Target
    download "https://github.com/lukso-network/$repo/releases/download/$tag/$client-$platform-$architecture.exe" "$Target\$CLIENT-$PLATFORM-$ARCHITECTURE.exe"

}

Function download_network_config ($network) {
    $CDN = "https://storage.googleapis.com/l15-cdn/networks/"+$network
    $TARGET = $InstallDir+"\networks\"+$network+"\config"
    New-Item -ItemType Directory -Force -Path $TARGET
    download $CDN"/network-config.yaml?ignoreCache=1" $TARGET"\network-config.yaml"
    download $CDN"/pandora-genesis.json?ignoreCache=1" $TARGET"\pandora-genesis.json"
    download $CDN"/pandora-nodes.json?ignoreCache=1" $TARGET"\pandora-nodes.json"
    download $CDN"/vanguard-config.yaml?ignoreCache=1" $TARGET"\vanguard-config.yaml"
    download $CDN"/vanguard-genesis.ssz?ignoreCache=1" $TARGET"\vanguard-genesis.ssz"
}

Function bind_binary($client, $tag)
{
    if (!(Test-Path "$InstallDir/binaries/$client/$tag/$client-$platform-$architecture.exe"))
    {
        download_binary $client $tag
    }
    if (Test-Path "$InstallDir\globalPath\$client") {
        rm "$InstallDir\globalPath\$client"
    }

    cmd /c mklink "$InstallDir\globalPath\$client" "$InstallDir\binaries\$client\$tag\$client-$platform-$architecture.exe"
}

Function bind_binaries()
{
    Write-Output Binding
}



Function check_validator_requirements()
{

  if (${wallet-dir}) {
      echo ${wallet-dir}
      if (!(Test-Path ${wallet-dir})) {
          Write-Output "ERROR! Cannot Validate, wallet not found"
          exit
      }
  }

   if (${wallet-password-file}) {
      if (!(Test-Path ${wallet-password-file})) {
          Write-Output "ERROR! Cannot Validate, password file not found"
          exit
      }
   }



  if (!${wallet-password-file}) {
      $securedValue = Read-Host -AsSecureString -Prompt "Enter validator password"
      $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($securedValue)
      $value = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
      $value | Out-File $Env:APPDATA\LUKSO\temp_pass.txt
  }
}

Function start_orchestrator()
{

    if (!(Test-Path "$datadir\orchestrator"))
    {
        New-Item -ItemType Directory -Force -Path $datadir\orchestrator
    }

    Write-Output $runDate | Out-File -FilePath "$logsdir\orchestrator\current.tmp"

    $arguments = @(
    "--datadir=$datadir\orchestrator"
    "--vanguard-grpc-endpoint=127.0.0.1:4000"
    "--http"
    "--http.addr=0.0.0.0"
    "--http.port=7877"
    "--ws"
    "--ws.addr=0.0.0.0"
    "--ws.port=7878"
    "--pandora-rpc-endpoint=ws://127.0.0.1:8546"
    "--verbosity=trace"
    )

    Write-Output $arguments

    Start-Process -FilePath lukso-orchestrator `
    -ArgumentList $arguments `
    -NoNewWindow `
    -RedirectStandardOutput "orchestrator_$runDate.out" `
    -RedirectStandardError "orchestrator_$runDate.err" `
}

function start_pandora()
{
    switch (${pandora-verbosity})
    {
        silent {
            ${pandora-verbosity} = 0
        }
        error {
            ${pandora-verbosity} = 1
        }
        warn {
            ${pandora-verbosity} = 2
        }
        info {
            ${pandora-verbosity} = 3
        }
        debug {
            ${pandora-verbosity} = 4
        }
        detail {
            ${pandora-verbosity}= 5
        }
        trace {
            ${pandora-verbosity}= 5
        }
    }

    New-Item -ItemType Directory -Force -Path $logsdir/pandora


    if (!(Test-Path $datadir/pandora)) {
        New-Item -ItemType Directory -Force -Path $datadir/pandora
    }

    Write-Output $runDate | Out-File -FilePath "$logsdir\pandora\current.tmp"

#    pandora init $InstallDir\networks\$NETWORK\config\pandora-genesis.json --datadir $datadir\pandora
#    echo $InstallDir\networks\$NETWORK\config\pandora-genesis.json
#    echo $datadir\pandora

    $Arguments = New-Object System.Collections.Generic.List[System.Object]
    $Arguments.Add("init")
    $Arguments.Add("$InstallDir\networks\$NETWORK\config\pandora-genesis.json")
    $Arguments.Add("--datadir=$datadir\pandora")
    Start-Process -Wait -FilePath "pandora" `
    -ArgumentList $Arguments `
    -NoNewWindow `
    -RedirectStandardOutput "$logsdir\pandora\init_pandora_$runDate.out" `
    -RedirectStandardError "$logsdir\pandora\init_pandora_$runDate.err"
    echo $Arguments


#    Copy-Item $InstallDir\networks\$NETWORK\config\pandora-nodes.json -Destination $datadir\pandora\geth

    $Arguments = New-Object System.Collections.Generic.List[System.Object]
    echo $($NetworkConfig.NETWORK_ID)
    $Arguments.Add("--datadir=$datadir\pandora")
    $Arguments.Add("--networkid=$($NetworkConfig.NETWORK_ID)")
    $Arguments.Add("--ethstats=${node-name}:$($NetworkConfig.ETH1_STATS_APIKEY)@$($NetworkConfig.ETH1_STATS_URL)")
    $Arguments.Add("--port=30405")
    $Arguments.Add("--http")
    $Arguments.Add("--http.addr=0.0.0.0")
    $Arguments.Add("--http.port=${pandora-http-port}")
    $Arguments.Add("--http.api=admin,net,eth,debug,miner,personal,txpool,web3")
    $Arguments.Add("--http.corsdomain=*")
    $Arguments.Add("--bootnodes=${pandora-bootnodes}")
    $Arguments.Add("--ws")
    $Arguments.Add("--ws.port=8546")
    $Arguments.Add("--ws.api=admin,net,eth,debug,miner,personal,txpool,web3")
    $Arguments.Add("--ws.origins=*")
    $Arguments.Add("--mine")
    $Arguments.Add("--miner.notify=ws://127.0.0.1:7878,http://127.0.0.1:7877")
    $Arguments.Add("--miner.etherbase=$coinbase")
    $Arguments.Add("--syncmode=full")
    $Arguments.Add("--allow-insecure-unlock")
    $Arguments.Add("--verbosity=${pandora-verbosity}")
    if (${pandora-external-ip}) {
        $Arguments.Add("--nat=extip:${pandora-external-ip}")
    }

    if (${pandora-metrics}) {
        $Arguments.Add("--metrics")
        $Arguments.Add("--metrics.expensive")
        $Arguments.Add("--pprof")
        $Arguments.Add("--pprof.addr=0.0.0.0")
    }

    if (${pandora-nodekey}) {
        $Arguments.Add("--nodekey=${pandora-nodekey}")
    }

    Write-Output $Arguments

    Start-Process -FilePath "pandora" `
    -ArgumentList $Arguments `
    -NoNewWindow `
    -RedirectStandardOutput "$logsdir\pandora\pandora_$runDate.out" `
    -RedirectStandardError "$logsdir\pandora\pandora_$runDate.err"
}

function start_vanguard() {
    if (!(Test-Path $logsdir\vanguard))
    {
        New-Item -ItemType Directory -Force -Path $logsdir\vanguard
    }

    Write-Output $runDate | Out-File -FilePath "$logsdir\vanguard\current.tmp"
    $Arguments = New-Object System.Collections.Generic.List[System.Object]

    $BootnodesArray = ${vanguard-bootnodes}.Split(",")
    echo $BootnodesArray
    foreach ($Bootnode in $BootnodesArray) {
        echo $Bootnode
    }

    $Arguments.Add("--accept-terms-of-use")
    $Arguments.Add("--chain-id=$($NetworkConfig.CHAIN_ID)")
    $Arguments.Add("--network-id=$($NetworkConfig.NETWORK_ID)")
    $Arguments.Add("--genesis-state=$InstallDir\networks\$network\config\vanguard-genesis.ssz")
    $Arguments.Add("--datadir=$datadir\vanguard\")
    $Arguments.Add("--chain-config-file=$InstallDir\networks\$network\config\vanguard-config.yaml")
    foreach ($Bootnode in $BootnodesArray) {
        $Arguments.Add("--bootstrap-node=$Bootnode")
    }
    $Arguments.Add("--http-web3provider=http://127.0.0.1:${pandora-http-port}")
    $Arguments.Add("--deposit-contract=0x000000000000000000000000000000000000cafe")
    $Arguments.Add("--contract-deployment-block=0")
    $Arguments.Add("--rpc-host=0.0.0.0")
    $Arguments.Add("--monitoring-host=0.0.0.0")
    $Arguments.Add("--verbosity=${vanguard-verbosity}")
    $Arguments.Add("--min-sync-peers=2")
    $Arguments.Add("--p2p-max-peers=50")
    $Arguments.Add("--orc-http-provider=http://127.0.0.1:7877")
    $Arguments.Add("--rpc-port=4000")
    $Arguments.Add("--p2p-udp-port=12000")
    $Arguments.Add("--p2p-tcp-port=13000")
    $Arguments.Add("--grpc-gateway-port=3500")
    $Arguments.Add("--update-head-timely")
    $Arguments.Add("--lukso-network")

    if (${vanguard-p2p-priv-key}) {
        $Arguments.Add("--p2p-priv-key=${vanguard-p2p-priv-key}")
    }

    if ($vanguard_p2p_host_dns) {
        $Arguments.Add("--p2p-host-dns=$vanguard_p2p_host_dns")
    }

    elseif ($vanguard_external_ip) {
        $Arguments.Add("--p2p-host-ip=$vanguard_external_ip")
    }

    else {
        $Arguments.Add("--p2p-host-ip=$vanguard_external_ip")
    }
    echo $Arguments
    Start-Process -FilePath "vanguard" `
    -ArgumentList $Arguments `
    -NoNewWindow `
    -RedirectStandardOutput "$logsdir\vanguard\vanguard_$runDate.out" `
    -RedirectStandardError "$logsdir\vanguard\vanguard_$runDate.err"

}

function start_validator() {
    check_validator_requirements
    if (!(Test-Path $logsdir\validator))
    {
        New-Item -ItemType Directory -Force -Path $logsdir\validator
    }

    Write-Output $runDate | Out-File -FilePath "$logsdir\validator\current.tmp"

    $Arguments = New-Object System.Collections.Generic.List[System.Object]
    $Arguments.Add("--datadir=$datadir\validator")
    $Arguments.Add("--accept-terms-of-use")
    $Arguments.Add("--beacon-rpc-provider=localhost:4000")
    $Arguments.Add("--chain-config-file=$InstallDir\networks\$network\config\vanguard-config.yaml")
    $Arguments.Add("--verbosity=$(${validator-verbosity})")
    $Arguments.Add("--wallet-dir=$(${wallet-dir})")
    $Arguments.Add("--rpc")
    $Arguments.Add("--log-file=$logsdir\validator\validator_$runDate.log")
    $Arguments.Add("--lukso-network")

    if (${wallet-password-file}) {
      $Arguments.Add("--wallet-password-file=${wallet-password-file}")
    }

    if (!${wallet-password-file}) {
      $Arguments.Add("--wallet-password-file=$Env:APPDATA\Lukso\temp_pass.txt")
    }

    Start-Process -FilePath "lukso-validator" `
    -ArgumentList $arguments `
    -NoNewWindow `
    -RedirectStandardOutput "$logsdir\validator\validator_$runDate.out" `
    -RedirectStandardError "$logsdir\validator\validator_$runDate.err"

}

function start_eth2stats() {

}

function start_all() {
    if ($validate) {
        check_validator_requirements
    }
    start_orchestrator
    start_pandora
    start_vanguard
    start_validator
    start_eth2stats
}

# "start" is a reserved keyword in PowerShell
function _start($client)
{
    switch ($client)
    {
        orchestrator {
            start_orchestrator
        }

        pandora {
            start_pandora
        }

        vanguard {
            start_vanguard
        }

        validator {
            start_validator
        }

        all {
            start_all
        }

        Default {
            start_all
        }
    }
}

function stop_orchestrator() {
    Stop-Process -ProcessName "lukso-orchestrator"
}

function stop_pandora() {
    Stop-Process -ProcessName "pandora"
}

function stop_vanguard() {
    Stop-Process -ProcessName "vanguard"
}

function stop_validator() {
    Stop-Process -ProcessName "lukso-validator"
}

function stop_eth2stats() {
    Stop-Process -ProcessName "eth2stats"
}

function stop_all() {
    stop_orchestrator
    stop_pandora
    stop_vanguard
    stop_validator
    stop_eth2stats
}

function _stop() {}

function reset_orchestrator () {

}

function reset_pandora() {

}

function reset_vanguard() {

}

function reset_validator() {

}

function reset_eth2stats() {

}

function reset_all() {

}

function reset() {

}

function _help() {
    Write-Output "USAGE:"
    Write-Output "lukso <command> [argument] [--flags]"
    Write-Output "`n"
    Write-Output "Available commands with arguments:"
    Write-Output "start)         Starts up all or specific client(s)"
    Write-Output "               [orchestrator, pandora, vanguard, validator, eth2stats-client, all]"
    Write-Output "`n"
    Write-Output "stop)          Stops all or specific client(s)"
    Write-Output "               [orchestrator, pandora, vanguard, validator, eth2stats-client, all]"
    Write-Output "`n"
    Write-Output "reset)         Clears client(s) datadirs (this also removes chain-data) 	"
    Write-Output "               [orchestrator, pandora, vanguard, validator, all, none]"
    Write-Output "`n"
    Write-Output "config)        Interactive tool for creating config file"
    Write-Output "`n"
    Write-Output "keygen)        Runs lukso-deposit-cli"
    Write-Output "`n"
    Write-Output "wallet)        Imports lukso-deposit-cli keys"
    Write-Output "`n"
    Write-Output "logs)          Shows logs"
    Write-Output "               [orchestrator, pandora, vanguard, validator, eth2stats-client]"
    Write-Output "`n"
    Write-Output "attach)        Attaches to pandora console via IPC socket (use with --datadir if not default)"
    Write-Output "`n"
    Write-Output "bind-binaries) sets client(s) to desired version, use with flags for setting tag: --orchestrator v0.1.0, --pandora v0.1.0, --vanguard v0.1.0, --validator v0.1.0"
    Write-Output "`n"
    Write-Output "`n"
    Write-Output "Available flags:"
    Write-Output "--network              Picks config collection to be used (and downloads if it doesn't exist)"
    Write-Output "                       [l15, l15-staging, l15-dev]"
    Write-Output "`n"
    Write-Output "--l15                  Alias for --network l15"
    Write-Output "`n"
    Write-Output "--l15-staging          Alias for --network l15-staging"
    Write-Output "`n"
    Write-Output "--l15-dev              Alias for --network l15-dev"
    Write-Output "`n"
    Write-Output "--config               Path to config file"
    Write-Output "                       [config.yaml]"
    Write-Output "`n"
    Write-Output "--datadir              Sets datadir path"
    Write-Output "                       [Ex. /mnt/external/lukso-datadir]"
    Write-Output "`n"
    Write-Output "--logsdir              Sets logs path"
    Write-Output "                       [Ex. /mnt/external/lukso-logs]"
    Write-Output "`n"
    Write-Output "--home                 Sets path for datadir and logs in a single location (--datadir and --logs take priority)"
    Write-Output "                       [Ex. /var/lukso]"
    Write-Output "`n"

    Write-Output "--validate             Starts validator"
    Write-Output "`n"
    Write-Output "--coinbase             Sets pandora coinbase"
    Write-Output "                       [ETH1 address ex. 0x616e6f6e796d6f75730000000000000000000777]"
    Write-Output "`n"
    Write-Output "--node-name            Name of node that's shown on pandora stats and vanguard stats"
    Write-Output "                       [String ex. johnsmith123]"
    Write-Output "`n"
    Write-Output "--orchestrator         Sets orchestrator tag to be used"
    Write-Output "                       [Tag name ex. v0.1.0-develop]"
    Write-Output "`n"
    Write-Output "--orchestrator-verbosity Sets orchestrator logging depth"
    Write-Output "                       [silent, error, warn, info, debug, trace]"
    Write-Output "`n"
    Write-Output "--pandora              Sets pandora tag to be used"
    Write-Output "                       [Tag name ex. v0.1.0-develop]"
    Write-Output "`n"
    Write-Output "--pandora-verbosity    Sets pandora logging depth (note: pandora uses integers for that flag, script will convert those to proper values)"
    Write-Output "                       [silent, error, warn, info, debug, trace]"
    Write-Output "`n"
    Write-Output "--pandora-bootnodes    Sets pandora bootnodes"
    Write-Output "                       [Strings of bootnodes separated by commas: \"enode://72caa...,enode://b4a11a...\"]"
    Write-Output "`n"
    Write-Output "--pandora-http-port    Sets pandora RPC (over http) port"
    Write-Output "                       [Number between 1023-65535]"
    Write-Output "`n"
    Write-Output "--pandora-metrics      Enables pandora metrics server"
    Write-Output "`n"
    Write-Output "--pandora-nodekey      P2P node key file"
    Write-Output "                       [Path to file (relative or absolute)]"
    Write-Output "`n"
    Write-Output "--pandora-external-ip  Sets external IP for pandora (overrides --external-ip if present)"
    Write-Output "                       [72.122.32.234]"
    Write-Output "`n"
    Write-Output "--vanguard             Sets vanguard tag to be used"
    Write-Output "                       [Tag name ex. v0.1.0-develop]"
    Write-Output "`n"
    Write-Output "--vanguard-verbosity   Sets vanguard logging depth"
    Write-Output "                       [silent, error, warn, info, debug, trace]"
    Write-Output "`n"
    Write-Output "--vanguard-bootnodes   Sets vanguard bootnodes"
    Write-Output "                       [Strings of bootnodes separated by commas: \"enr:-Ku4QAmY...,enr:-M23QLmY...\"]"
    Write-Output "`n"
    Write-Output "--vanguard-p2p-priv-key The file containing the private key to use in communications with other peers."
    Write-Output "                       [Path to file (relative or absolute)]"
    Write-Output "`n"
    Write-Output "--vanguard-external-ip Sets external IP for vanguard (overrides --external-ip if present)"
    Write-Output "                       [72.122.32.234]"
    Write-Output "`n"
    Write-Output "--vanguard-p2p-host-dns Sets host DNS vanguard (overrides --external-ip AND --vanguard-external-ip if present)"
    Write-Output "                       [72.122.32.234]"
    Write-Output "`n"
    Write-Output "--validator            Sets validator tag to be used"
    Write-Output "                       [Tag name ex. v0.1.0-develop]"
    Write-Output "`n"
    Write-Output "--validator-verbosity  Sets validator logging depth"
    Write-Output "                       [silent, error, warn, info, debug, trace]"
    Write-Output "`n"
    Write-Output "--external-ip          Sets external IP for pandora and vanguard"
    Write-Output "                       [72.122.32.234]"
    Write-Output "`n"
    Write-Output "--allow-respin         Deletes all datadirs IF network config changed (based on genesis time)"
    Write-Output "`n"
    Write-Output "--force                Enables force mode for stopping"

    exit
}


##Flags action
if ($orchestrator)
{
    bind_binary lukso-orchestrator $orchestrator
}

if ($pandora)
{
    bind_binary pandora $pandora
}

if ($vanguard)
{
    bind_binary vanguard $vanguard
}

if ($validator)
{
    bind_binary lukso-validator $validator
}

switch ($command)
{
    "start" {
        _start $argument
        $KeepShell = $true
    }

    "stop" {
        _stop $argument
    }

    "restart" {
        _restart $argument
    }

    "help" {
        _help
    }

    "logs" {
        logs $argument
    }

    "bind-binaries" {
        bind_binaries
    }

    Default {
        Write-Output "Unknown command"
        exit
    }
}

if ($KeepShell)
{
    Write-Output "LUKSO clients are working do not close this shell"
}

while ($KeepShell)
{
    Read-Host
}

