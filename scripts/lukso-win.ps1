param (
    [Parameter(Position = 0, Mandatory)][String]$command,
    [Parameter(Position = 1)][String]$argument,

    [String]$orchestrator = "",
    [String]$pandora = "",
    [String]$vanguard = "",
    [String]$validator = "",
    [String]$deposit = "",
    [String]$eth2stats = "",
    [String]$network = "l15",
    [String]$lukso_home = "$HOME\.lukso",
    [String]$datadir = "$lukso_home\datadir",
    [String]$logsdir = "$lukso_home\logs",
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
    [String]$vanguard_bootnodes = "",
    [String]$vanguard_p2p_priv_key = "",
    [String]$vanguard_external_ip = "",
    [String]$vanguard_p2p_host_dns = "",
    [String]$external_ip = "",
    [Switch]$allow_respin,
    [Switch]$force
)

$runDate = Get-Date -Format "yyyy-m-dd__HH-mm-ss"

function start_orchestrator() {

    New-Item -ItemType Directory -Force -Path $logsdir\orchestrator
    echo -n $runDate | Out-File -FilePath "$logsdir\orchestrator\current.tmp"

    if ( ! (Test-Path "$datadir\orchestrator") ) {
       New-Item -ItemType Directory -Force -Path $datadir\orchestrator
    }

Start-Process -FilePath "lukso-orchestrator" -ArgumentList "--datadir=$datadir\orchestrator", `
"--vanguard-grpc-endpoint=127.0.0.1:4000", `
"--http", `
"--http.addr=0.0.0.0",  `
"--http.port=7877", `
"--ws", `
"--ws.addr=0.0.0.0", `
"--ws.port=7878", `
"--pandora-rpc-endpoint=ws://127.0.0.1:8546", `
"--verbosity=trace" -NoNewWindow -RedirectStandardOutput "orchestrator_$runDate.out" -RedirectStandardError "orchestrator_$runDate..err"
#    | Out-File -FilePath $logsdir/orchestrator/orchestrator_"$RUN_DATE".log



}

# "start" is a reserved keyword in PowerShell
function _start($client) {
    switch ($client) {
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
            echo none
        }
    }
}

switch ($command) {
start {
    _start $argument
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

}

Default {
    Write-Output "Unknown command"
    exit
}
}

