# Explicit LAN reachability test (Windows + WSL). No admin required.
$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
$Port = "18080"

$winIP = (Get-NetIPAddress -AddressFamily IPv4 |
  Where-Object { $_.InterfaceAlias -match 'Wi-Fi' -and $_.IPAddress -notlike '169.*' } |
  Select-Object -First 1 -ExpandProperty IPAddress)
if (-not $winIP) { throw "could not detect Windows Wi-Fi IP" }

$wslIP = (wsl -e bash -lc "hostname -I | awk '{print `$1}'").Trim().Split(" ")[0]
Write-Host "Windows Wi-Fi IP: $winIP  WSL eth: $wslIP  port: $Port"

wsl -e bash -lc "sed -i 's/\r`$//' /mnt/c/termshare/scripts/start-lan-test.sh"
$jsonLine = (wsl -e bash -lc "bash /mnt/c/termshare/scripts/start-lan-test.sh $winIP $Port").Trim()
Write-Host "JSON: $jsonLine"
if ($jsonLine -notmatch [regex]::Escape($winIP)) { throw "lanViewer missing Windows Wi-Fi IP" }
if ($jsonLine -match "10\.255\.255\.254") { throw "still advertising WSL loopback alias" }
if ($jsonLine -notmatch ":$Port/") { throw "JSON missing test port $Port" }

$j = $jsonLine | ConvertFrom-Json
$id = $j.id

# Verify server alive inside WSL first
$wslLocal = (wsl -e bash -lc "curl -s -o /dev/null -w '%{http_code}' --connect-timeout 2 http://127.0.0.1:$Port/s/$id").Trim()
Write-Host "WSL localhost -> $wslLocal"
if ($wslLocal -ne "200") { throw "termshare not serving inside WSL: $wslLocal" }

$wslEth = (wsl -e bash -lc "curl -s -o /dev/null -w '%{http_code}' --connect-timeout 2 http://$wslIP`:$Port/s/$id").Trim()
Write-Host "WSL eth -> $wslEth"

# Bridge Windows LAN -> WSL eth (not 127.0.0.1; localhost relay is unreliable here)
$bridgePath = Join-Path $PWD "scripts\lan-bridge-tmp.js"
@"
const net = require("net");
const lanIP = process.env.LAN_IP;
const port = Number(process.env.PORT);
const target = process.env.TARGET_HOST;
const server = net.createServer((client) => {
  const up = net.connect(port, target);
  client.pipe(up);
  up.pipe(client);
  const fail = () => { try { client.destroy(); } catch {} try { up.destroy(); } catch {} };
  client.on("error", fail);
  up.on("error", fail);
});
server.on("error", (e) => { console.error(String(e)); process.exit(1); });
server.listen(port, lanIP, () => console.log("bridge-ready " + lanIP + ":" + port + " -> " + target));
"@ | Set-Content -Path $bridgePath -Encoding ascii

$env:LAN_IP = $winIP
$env:PORT = $Port
$env:TARGET_HOST = $wslIP
$bridge = Start-Process -FilePath "node" -ArgumentList @($bridgePath) -PassThru -WindowStyle Hidden `
  -RedirectStandardOutput (Join-Path $PWD "bridge-out.txt") `
  -RedirectStandardError (Join-Path $PWD "bridge-err.txt")
Start-Sleep -Seconds 1
Write-Host ("bridge out: " + ((Get-Content (Join-Path $PWD "bridge-out.txt") -Raw -ErrorAction SilentlyContinue) + "").Trim())
$berr = ((Get-Content (Join-Path $PWD "bridge-err.txt") -Raw -ErrorAction SilentlyContinue) + "").Trim()
if ($berr) { Write-Host "bridge err: $berr" }

function Get-Status([string]$url) {
  try {
    $resp = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 5
    return [int]$resp.StatusCode
  } catch {
    if ($_.Exception.Response) { return [int]$_.Exception.Response.StatusCode }
    return 0
  }
}

$winToWsl = Get-Status "http://${wslIP}:${Port}/s/$id"
$lanCode = Get-Status "http://${winIP}:${Port}/s/$id"
Write-Host "Windows -> WSL eth:$Port -> $winToWsl"
Write-Host "Windows -> LAN ${winIP}:$Port -> $lanCode"

Stop-Process -Id $bridge.Id -Force -ErrorAction SilentlyContinue
wsl -e bash -lc "kill `$(cat /tmp/ts-lan.pid) 2>/dev/null; pkill -f /tmp/termshare-lan-test || true"
Remove-Item $bridgePath, (Join-Path $PWD "bridge-out.txt"), (Join-Path $PWD "bridge-err.txt") -ErrorAction SilentlyContinue

if ($winToWsl -ne 200) { throw "Windows cannot reach WSL eth IP: $winToWsl" }
if ($lanCode -ne 200) { throw "LAN session failed: $lanCode" }
Write-Host "PASS: LAN URL advertises $winIP and returns 200 via WSL-eth bridge"
