param (
    [Parameter(Position = 0, Mandatory)][string]$command,
    [Parameter(Position = 1)][string]$argument,

    [string]$orchestrator = "",
    [string]$pandora = "",
    [string]$vanguard = "",
    [string]$validator = "",
    [string]$deposit = "",
    [string]$eth2stats = "",
    [string]$network = "l15",
    [string]$lukso_home = "$HOME\.lukso",
    [string]$datadir = "$lukso_home\datadir",
    [string]$keys_dir = "",
    [string]$keys_password_file = "",
    [string]$wallet_dir = "",
    [string]$wallet_password_file = "",
    [string]$logsdir = "",
    [string]$l15 = "",
    [string]$l15_staging = "",
    [string]$l15_dev = "",
    [string]$config = "",
    [string]$coinbase = "",
    [string]$node_name = "",
    [string]$validate = "",
    [string]$pandora_bootnodes = "",
    [string]$pandora_http_port = "",
    [string]$pandora_metrics = "",
    [string]$pandora_nodekey = "",
    [string]$pandora_external_ip = "",
    [string]$vanguard_bootnodes = "",
    [string]$vanguard_p2p_priv_key = "",
    [string]$vanguard_external_ip = "",
    [string]$vanguard_p2p_host_dns = "",
    [string]$external_ip = "",
    [string]$allow_respin = "",
    [Switch]$force = $false
)

#function start_orchestrator() {
#mkdir -p "$LOGSDIR"/orchestrator
#echo -n $RUN_DATE >|"$LOGSDIR"/orchestrator/current.tmp
#if [[ ! -d "$DATADIR/orchestrator" ]]; then
#mkdir -p $DATADIR/orchestrator
#fi
#
#Start-Process -FilePath lukso-orchestrator -ArgumentList "--datadir=$DATADIR/orchestrator", `
#"--vanguard-grpc-endpoint=127.0.0.1:4000", `
#"--http", `
#"--http.addr=0.0.0.0",  `
#"--http.port=7877" `
#"--ws" `
#"--ws.addr=0.0.0.0" `
#"--ws.port=7878" `
#"--pandora-rpc-endpoint=ws://127.0.0.1:8546" `
#--verbosity=trace | Out-File -FilePath $LOGSDIR/orchestrator/orchestrator_"$RUN_DATE".log
#
#}

function _start() {
  Write-Output $argument
}

switch ($command) {
start {
    _start $argument
}
Default {
    Write-Output "Unknown command"
    exit
}
}

Write-Output $command
Write-Output $server
Write-Output $force
