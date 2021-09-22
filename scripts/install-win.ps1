$Network = "l15";
$Repository = "https://storage.googleapis.com/l15-cdn";
$InstallDir = $Env:APPDATA;

$client = New-Object System.Net.WebClient

mkdir $InstallDir `
$InstallDir/tmp `
$InstallDir/binaries `
$InstallDir/networks `
$InstallDir/networks/"$NETWORK" `
$InstallDir/networks/"$NETWORK"/config;


$url = $Repository+"/config.zip";
$path = $InstallDir+"/tmp/config.zip";
$client.DownloadFile($url, $path)

