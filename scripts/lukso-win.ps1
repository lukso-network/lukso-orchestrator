param (
    [Parameter(Position = 0, Mandatory)][String]$command,
    [Parameter(Position = 1)][String]$argument,

    [String]$orchestrator = "",
    [String]$orchestrator_verbosity = "",
    [String]$pandora = "",
    [String]$vanguard = "",
    [String]$validator = "",
    [String]$deposit = "",
    [String]$eth2stats = "",
    [String]$network = "l15",
    [String]$lukso_home = "$HOME\.lukso",
    [String]$datadir = "$lukso_home\$network\datadir",
    [String]$logsdir = "$lukso_home\$network\logs",
    [String]$keys_dir = "",
    [String]$keys_password_file = "",
    [String]$wallet_dir = "",
    [String]$wallet_password_file = "",
    [Switch]$l15,
    [Switch]$l15_staging,
    [Switch]$l15_dev,
    [String]$config = "",
    [String]$coinbase = "0x616e6f6e796d6f75730000000000000000000000",
    [String]$node_name = "",
    [Switch]$validate,
    [String]$pandora_bootnodes = "",
    [String]$pandora_http_port = "8545",
    [Switch]$pandora_metrics,
    [String]$pandora_nodekey = "",
    [String]$pandora_external_ip = "",
    [String]$pandora_verbosity = "",
    [String]$vanguard_bootnodes = "",
    [String]$vanguard_p2p_priv_key = "",
    [String]$vanguard_external_ip = "",
    [String]$vanguard_p2p_host_dns = "",
    [String]$vanguard_verbosity = "",
    [String]$validator_verbosity = "",
    [String]$external_ip = "",
    [Switch]$allow_respin,
    [Switch]$force
)

$platform = "Windows"
$architecture = "x86_64"
$InstallDir = $Env:APPDATA + "\LUKSO";

$runDate = Get-Date -Format "yyyy-m-dd__HH-mm-ss"

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
        orchestrator {
            $repo = "lukso-orchestrator"
        }

        pandora {
            $repo = "pandora-execution-engine"
        }

        vanguard {
            $repo = "vanguard-consensus-engine"
        }

        validator {
            $repo = "vanguard-consensus-engine"
        }
    }

    $TARGET = "$InstallDir\binaries\$CLIENT\$TAG"
    New-Item -ItemType Directory -Force -Path $TARGET
    download "https://github.com/lukso-network/$repo/releases/download/$tag/$client-$platform-$architecture.exe" "$TARGET\$CLIENT-$PLATFORM-$ARCHITECTURE"

}

Function download_network_config($network)
{
    $CDN = "https://storage.googleapis.com/l15-cdn/networks/" + $network
    $TARGET = $InstallDir + "\networks\" + $network + "\config"
    New-Item -ItemType Directory -Force -Path $TARGET
    download $CDN"\network-config.yaml?ignoreCache=1" $TARGET"\network-config.yaml"
}

Function bind_binary($client, $tag)
{
    if (!(Test-Path "$InstallDir/binaries/$client/$tag/$client-$platform-$architecture"))
    {
        download_binary $client $tag
    }
    rm "$InstallDir\globalPath\$client"
    cmd /c mklink "$InstallDir\globalPath\$client" "$InstallDir\binaries\$client\$tag\$client-$platform-$architecture"
}

Function bind_binaries()
{
    Write-Output binding
}

Function pick_network($picked_network)
{
    $network = $picked_network
    if (!(Test-Path "$InstallDir\networks\$network"))
    {
        download_network_config $network
    }
}

function check_validator_requirements()
{

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

    Start-Process -FilePath "lukso-orchestrator" `
    -ArgumentList $arguments `
    -NoNewWindow `
    -RedirectStandardOutput "orchestrator_$runDate.out" `
    -RedirectStandardError "orchestrator_$runDate..err"


}

function start_pandora()
{
    switch ($pandora_verbosity)
    {
        silent {
            PANDORA_VERBOSITY = 0
        }
        error {
            PANDORA_VERBOSITY = 1
        }
        warn {
            PANDORA_VERBOSITY = 2
        }
        info {
            PANDORA_VERBOSITY = 3
        }
        debug {
            PANDORA_VERBOSITY = 4
        }
        detail {
            PANDORA_VERBOSITY= 5
        }
        trace {
            PANDORA_VERBOSITY= 5
        }
    }

    New-Item -ItemType Directory -Force -Path $logsdir/pandora


    if (!(Test-Path $datadir/pandora)) {
        New-Item -ItemType Directory -Force -Path $datadir/pandora
    }

    pandora --datadir $DATADIR/pandora init $InstallDir\networks\$NETWORK\config\pandora-genesis.json
    Copy-Item $InstallDir\networks\$NETWORK\config\pandora-nodes.json -Destination $datadir\pandora\geth

    $Arguments = New-Object System.Collections.Generic.List[System.Object]

    $Arguments.Add("--datadir=$DATADIR/pandora")
    $Arguments.Add("--networkid=$NETWORK_ID")
    $Arguments.Add("--ethstats=${NODE_NAME}:$ETH1_STATS_APIKEY@$ETH1_STATS_URL")
    $Arguments.Add("--port=30405")
    $Arguments.Add("--rpc")
    $Arguments.Add("--rpcaddr=0.0.0.0")
    $Arguments.Add("--rpcapi=admin,net,eth,debug,miner,personal,txpool,web3")
    $Arguments.Add("--bootnodes=$PANDORA_BOOTNODES")
    $Arguments.Add("--http.corsdomain=*")
    $Arguments.Add("--ws")
    $Arguments.Add("--ws.api=admin,net,eth,debug,miner,personal,txpool,web3")
    $Arguments.Add("--ws.port=8546")
    $Arguments.Add("--ws.origins=*")
    $Arguments.Add("--mine")
    $Arguments.Add("--miner.notify=ws://127.0.0.1:7878,http://127.0.0.1:7877")
    $Arguments.Add("--miner.etherbase=$COINBASE")
    $Arguments.Add("--syncmode=full")
    $Arguments.Add("--allow-insecure-unlock")
    $Arguments.Add("--verbosity=$pandora_verbosity")
    $Arguments.Add("--nat=extip:$pandora_external_ip")

    if ($pandora_metrics) {
        $Arguments.Add("--metrics")
        $Arguments.Add("--metrics.expensive")
        $Arguments.Add("--pprof")
        $Arguments.Add("--pprof.addr=0.0.0.0")
    }

    if ($pandora_nodekey) {
        $Arguments.Add("--nodekey=$pandora_nodekey")
    }

    Start-Process -FilePath "pandora" `
    -ArgumentList $arguments `
    -NoNewWindow `
    -RedirectStandardOutput "$logsdir\pandora_$runDate.out" `
    -RedirectStandardError "$logsdir\pandora_$runDate..err"
}

function start_vanguard() {}

function start_validator() {}

function start_eth2stats() {}

function start_all() {}

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

        Default {
            Write-Output none
        }
    }
}

function stop_orchestrator() {
    Stop-Process -ProcessName "lukso-orchestrator"
}

function stop_pandora() {}

function stop_vanguard() {}

function stop_validator() {}

function stop_eth2stats() {}

function stop_all() {}

function stop() {}

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
    bind_binary orchestrator $orchestrator
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
    bind_binary validator $validator
}

switch ($command)
{
    start {
        _start $argument
        $KeepShell = $true
    }

    stop {
        _stop $argument
    }

    restart {
        _restart $argument
    }

    help {
        _help
    }

    logs {
        logs $argument
    }

    bind-binaries {
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

