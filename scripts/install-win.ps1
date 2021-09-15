$Network = "l15";
$Repository = "https://storage.googleapis.com/l16-common/l15-cdn";
$InstallDir = "/home/pk/tmp/dupa";

$client = New-Object System.Net.WebClient

mkdir $InstallDir `
$InstallDir/tmp `
$InstallDir/binaries `
$InstallDir/networks `
$InstallDir/networks/"$NETWORK" `
$InstallDir/networks/"$NETWORK"/config;


$url = $Repository+"/config.zip";
$client.DownloadFile($url, $path)

