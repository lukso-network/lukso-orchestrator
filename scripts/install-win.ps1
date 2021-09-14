$Repository = "https://storage.googleapis.com/l16-common/l15-cdn";
$client = New-Object System.Net.WebClient

$url = $Repository+"/config.zip";
$client.DownloadFile($url, $path)

mkdir /opt/lukso0 `
/opt/lukso/tmp `
/opt/lukso/binaries `
/opt/lukso/networks `
/opt/lukso/networks/"$NETWORK" `
/opt/lukso/networks/"$NETWORK"/config;

