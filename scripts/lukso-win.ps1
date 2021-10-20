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
    [String]$vanguard_bootnodes = "",
    [String]$vanguard_p2p_priv_key = "",
    [String]$vanguard_external_ip = "",
    [String]$vanguard_p2p_host_dns = "",
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
    echo $url
    echo $dst
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
    Write-Output
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
    -NoNewWindow -RedirectStandardOutput "orchestrator_$runDate.out" -RedirectStandardError "orchestrator_$runDate..err"


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

        Default {
            echo none
        }
    }
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
        Write-Output binding
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

