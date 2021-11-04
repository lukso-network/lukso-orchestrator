$Network = "l15";
$InstallDir = $Env:APPDATA+"\LUKSO";

Function download ($url, $dst) {
    $client = New-Object System.Net.WebClient
    $client.DownloadFile($url, $dst)
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


New-Item -ItemType Directory -Force -Path $InstallDir
New-Item -ItemType Directory -Force -Path $InstallDir/binaries
New-Item -ItemType Directory -Force -Path $InstallDir/networks
New-Item -ItemType Directory -Force -Path $InstallDir/globalPath

download_network_config("l15-prod")
download_network_config("l15-staging")
download_network_config("l15-dev")

Invoke-WebRequest -Uri https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/l15-setup/scripts/lukso-win.ps1 -OutFile $InstallDir\lukso.ps1
if (Test-Path "$InstallDir\globalPath\lukso") {
    rm "$InstallDir\globalPath\lukso"
}

Write-Output "powershell.exe -File $InstallDir\lukso.ps1 %*" | Out-File -Encoding ASCII -FilePath "$InstallDir\globalPath\lukso.bat"

lukso bind-binaries `
-orchestrator v0.2.0-rc.1 `
-pandora v0.2.0-rc.1 `
-vanguard v0.2.0-rc.1 `
-validator v0.2.0-rc.1 `
-eth2stats v0.1.0-rc.1 `
-deposit v1.2.6-LUKSO
