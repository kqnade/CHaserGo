# CHaser Integration Test for Windows
Write-Host "=== CHaser Integration Test ===" -ForegroundColor Cyan

# テストマップのパスを確認
$mapFile = ""
if (Test-Path "testdata\RandMap_1.map") {
    $mapFile = "testdata\RandMap_1.map"
} elseif (Test-Path "testdata\sample_map.txt") {
    $mapFile = "testdata\sample_map.txt"
} else {
    Write-Host "Error: No test map found" -ForegroundColor Red
    exit 1
}

Write-Host "Using map: $mapFile" -ForegroundColor Green

# サーバーをバックグラウンドで起動
Write-Host "Starting server..." -ForegroundColor Yellow
$server = Start-Process -FilePath ".\bin\chaser-server.exe" -ArgumentList "-nd", $mapFile -PassThru -NoNewWindow

# サーバーの起動を待つ
Start-Sleep -Seconds 2

# サーバーが起動しているか確認
if ($server.HasExited) {
    Write-Host "Error: Server failed to start" -ForegroundColor Red
    exit 1
}

Write-Host "Server started (PID: $($server.Id))" -ForegroundColor Green

# test1をビルドして実行
Write-Host "Building test clients..." -ForegroundColor Yellow
Push-Location examples\test1
go build -o ..\..\bin\test1.exe .
Pop-Location

Push-Location examples\test2
go build -o ..\..\bin\test2.exe .
Pop-Location

# クライアントを起動
Write-Host "Starting test clients..." -ForegroundColor Yellow
$client1 = Start-Process -FilePath ".\bin\test1.exe" -PassThru -NoNewWindow
$client2 = Start-Process -FilePath ".\bin\test2.exe" -PassThru -NoNewWindow

# クライアントの終了を待つ（最大30秒）
Write-Host "Waiting for clients to finish..." -ForegroundColor Yellow
$timeout = 30
$elapsed = 0

while ((-not $client1.HasExited -or -not $client2.HasExited) -and $elapsed -lt $timeout) {
    Start-Sleep -Seconds 1
    $elapsed++
}

# タイムアウトした場合はクライアントを強制終了
if ($elapsed -ge $timeout) {
    Write-Host "Timeout reached, stopping clients..." -ForegroundColor Yellow
    if (-not $client1.HasExited) { Stop-Process -Id $client1.Id -Force }
    if (-not $client2.HasExited) { Stop-Process -Id $client2.Id -Force }
}

# サーバーの終了を待つ（最大5秒）
Write-Host "Waiting for server to finish..." -ForegroundColor Yellow
$server.WaitForExit(5000)

# サーバーがまだ動いていたら強制終了
if (-not $server.HasExited) {
    Write-Host "Stopping server..." -ForegroundColor Yellow
    Stop-Process -Id $server.Id -Force
}

Write-Host "Server exited with code: $($server.ExitCode)" -ForegroundColor Cyan
Write-Host "Client 1 exited with code: $($client1.ExitCode)" -ForegroundColor Cyan
Write-Host "Client 2 exited with code: $($client2.ExitCode)" -ForegroundColor Cyan

# 結果判定
if ($server.ExitCode -eq 0) {
    Write-Host "=== Integration Test PASSED ===" -ForegroundColor Green
    exit 0
} else {
    Write-Host "=== Integration Test FAILED ===" -ForegroundColor Red
    exit 1
}
