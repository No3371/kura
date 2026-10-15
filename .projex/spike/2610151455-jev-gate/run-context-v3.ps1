# Buyer-only intents (q3): rich2+q3 and rich3+q3, then compares. 96 calls.
# Report output goes to context-v3.log, per-call progress to
# context-v3.progress.log. Neither file contains the API key: the spike only
# reads TYPESAFE_API_KEY from the environment and never prints it.
param([string]$Base = 'https://api.typesafe.ai') # override only to dry-run against a fake API
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$spike = $PSScriptRoot
$kura = (Resolve-Path (Join-Path $spike '..\..\..')).Path
$log = Join-Path $spike 'context-v3.log'
$progress = Join-Path $spike 'context-v3.progress.log'
$exe = Join-Path $spike 'jevspike.exe'
Set-Content -Path $log -Value "context-v3 run $(Get-Date -Format s)" -Encoding UTF8
Set-Content -Path $progress -Value '' -Encoding UTF8

Push-Location $kura
try {
    & go build -o $exe ./.projex/spike/2610151455-jev-gate
    if ($LASTEXITCODE -ne 0) { Add-Content $log 'BUILD FAILED' -Encoding UTF8; exit 1 }
} finally { Pop-Location }

function Step([string]$label, [string[]]$spikeArgs, [string]$produces = '') {
    Add-Content $log "`n===== $label =====" -Encoding UTF8
    if ($produces -and (Test-Path (Join-Path $spike $produces))) {
        Add-Content $log "skipped: $produces already exists (delete it to re-run)" -Encoding UTF8
        return
    }
    $errFile = Join-Path $spike 'context-v3.stderr.tmp'
    # Windows PowerShell 5.1 turns redirected native stderr into error records;
    # under 'Stop' the first progress line would abort the run.
    $ErrorActionPreference = 'Continue'
    $out = & $exe -dir $spike -base $Base @spikeArgs 2> $errFile
    $code = $LASTEXITCODE
    $ErrorActionPreference = 'Stop'
    Add-Content $log ($out -join "`n") -Encoding UTF8
    $err = Get-Content $errFile -ErrorAction SilentlyContinue | ForEach-Object { "$_" -replace '^.*NativeCommandError.*$', '' } | Where-Object { $_ }
    Add-Content $progress "===== $label (exit $code)" -Encoding UTF8
    if ($err) { Add-Content $progress ($err -join "`n") -Encoding UTF8 }
    Remove-Item $errFile -ErrorAction SilentlyContinue
    if ($code -ne 0) {
        Add-Content $log "STEP FAILED (exit $code): $(($err | Select-Object -Last 3) -join ' | ')" -Encoding UTF8
        throw "step '$label' failed"
    }
}

try {
    Step 'run rich2 q3' @('-context', 'rich2', '-questions', 'q3') 'results-rich2-q3.jsonl'
    Step 'run rich3 q3' @('-context', 'rich3', '-questions', 'q3') 'results-rich3-q3.jsonl'
    Step 'compare questions: rich2-q2 vs rich2-q3 (original 42 rows)' @('-compare', 'results-rich2-q2.jsonl,results-rich2-q3.jsonl')
    Step 'compare context: rich2-q3 vs rich3-q3 (all 48 rows)' @('-compare', 'results-rich2-q3.jsonl,results-rich3-q3.jsonl')
    Add-Content $log "`nDONE" -Encoding UTF8
} catch {
    Add-Content $log "`nABORTED: $_" -Encoding UTF8
} finally {
    Remove-Item $exe -ErrorAction SilentlyContinue
}
