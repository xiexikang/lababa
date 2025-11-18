param()
$ErrorActionPreference = 'Stop'

Set-Location "$PSScriptRoot"

if (Test-Path ".env") {
  Get-Content ".env" | ForEach-Object {
    $line = $_.Trim()
    if ($line -and -not $line.StartsWith('#') -and $line.Contains('=')) {
      $kv = $line.Split('=',2)
      $name = $kv[0].Trim()
      $value = $kv[1].Trim()
      if ($name) { Set-Item -Path Env:$name -Value $value }
    }
  }
}

try {
  $net = netstat -ano | Select-String ":8081" | Select-String "LISTENING"
  if ($net) {
    $pid = ($net.ToString().Trim() -split "\s+")[-1]
    if ($pid -match '^[0-9]+$') { taskkill /PID $pid /F | Out-Null }
  }
} catch {}

$env:PORT = "8081"
$log = Join-Path $PSScriptRoot 'run-8081.log'

while ($true) {
  Write-Host "[run-8081] starting server-go.exe on :8081 at $(Get-Date)"
  "[run-8081] start at $(Get-Date)" | Out-File -FilePath $log -Append -Encoding utf8
  & "$PSScriptRoot/server-go.exe" *>> $log
  $code = $LASTEXITCODE
  Write-Host "[run-8081] server exited with code $code, restarting in 2s..."
  "[run-8081] exit code $code at $(Get-Date)" | Out-File -FilePath $log -Append -Encoding utf8
  Start-Sleep -Seconds 2
}
