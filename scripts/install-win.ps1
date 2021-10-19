$Network = "l15";
$InstallDir = $Env:APPDATA+"\LUKSO";

Function download ($url, $dst) {
    echo $url
    echo $dst
    $client = New-Object System.Net.WebClient
    $client.DownloadFile($url, $dst)
}

Function download_network_config ($network) {
    $CDN = "https://storage.googleapis.com/l15-cdn/networks/"+$network
    $TARGET = $InstallDir+"\networks\"+$network+"\config"
    New-Item -ItemType Directory -Force -Path $TARGET
    download $CDN"\network-config.yaml?ignoreCache=1" $TARGET"\network-config.yaml"
}


New-Item -ItemType Directory -Force -Path $InstallDir
New-Item -ItemType Directory -Force -Path $InstallDir/tmp
New-Item -ItemType Directory -Force -Path $InstallDir/binaries
New-Item -ItemType Directory -Force -Path $InstallDir/networks
New-Item -ItemType Directory -Force -Path $InstallDir/globalPath

download_network_config("l15-dev")

