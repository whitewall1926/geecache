param(
    [string]$Target = "http://10.152.42.150:8001/_geecache/",
    [string]$Group = "scores",
    [string]$Mode = "mixed",
    [string]$HotKey = "Tom",
    [int]$HotRatio = 80,
    [string]$Keys = "Tom,Jack,Sam",
    [int]$Requests = 50000,
    [int[]]$Concurrencies = @(1, 10, 50, 100, 200, 500),
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

if ($OutFile -eq "") {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $OutFile = "sweep_$timestamp.tsv"
}

"concurrency`tmode`ttotal`tsuccess`tfailed`tqps`tp50`tp95`tp99" | Out-File -FilePath $OutFile -Encoding utf8

Write-Host "Start sweep"
Write-Host "target=$Target group=$Group mode=$Mode requests=$Requests"
Write-Host "concurrency list: $($Concurrencies -join ', ')"

$benchBin = Join-Path $scriptDir "geecache-bench.exe"
go build -o $benchBin .\client.go
if ($LASTEXITCODE -ne 0) {
    throw "go build failed"
}

foreach ($c in $Concurrencies) {
    Write-Host ""
    Write-Host "== Run concurrency: $c =="

    $output = & $benchBin `
        -target $Target `
        -group $Group `
        -mode $Mode `
        -hot-key $HotKey `
        -hot-ratio $HotRatio `
        -keys $Keys `
        -n $Requests `
        -c $c
    if ($LASTEXITCODE -ne 0) {
        throw "benchmark client failed at concurrency=$c"
    }

    $output | ForEach-Object { Write-Host $_ }

    function Read-Metric([string]$Name) {
        $line = $output | Where-Object { $_ -match "^$([regex]::Escape($Name))\s*:" } | Select-Object -First 1
        if ($null -eq $line) {
            return ""
        }
        return (($line -split ":\s*", 2)[1]).Trim()
    }

    $total = Read-Metric "Total Requests"
    $success = Read-Metric "Success"
    $failed = Read-Metric "Failed"
    $qps = Read-Metric "QPS"
    $p50 = Read-Metric "P50 Latency"
    $p95 = Read-Metric "P95 Latency"
    $p99 = Read-Metric "P99 Latency"

    "$c`t$Mode`t$total`t$success`t$failed`t$qps`t$p50`t$p95`t$p99" | Add-Content -Path $OutFile -Encoding utf8
}

Write-Host ""
Write-Host "Sweep finished. Summary saved to: $((Resolve-Path $OutFile).Path)"
Get-Content $OutFile
