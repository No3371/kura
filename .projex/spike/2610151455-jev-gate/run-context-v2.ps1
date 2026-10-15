# Context v2 rerun: rich2+q2 twice (effect + noise), thin+q2, then the three
# compares. Report output goes to context-v2.log, per-call progress to
# context-v2.progress.log. Neither file contains the API key: the spike only
# reads TYPESAFE_API_KEY from the environment and never prints it.
param([string]$Base = 'https://api.typesafe.ai') # override only to dry-run against a fake API
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$spike = $PSScriptRoot
$kura = (Resolve-Path (Join-Path $spike '..\..\..')).Path
$log = Join-Path $spike 'context-v2.log'
$progress = Join-Path $spike 'context-v2.progress.log'
$exe = Join-Path $spike 'jevspike.exe'
Set-Content -Path $log -Value "context-v2 run $(Get-Date -Format s)" -Encoding UTF8
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
    $errFile = Join-Path $spike 'context-v2.stderr.tmp'
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
    Step 'run rich2 q2' @('-context', 'rich2', '-questions', 'q2') 'results-rich2-q2.jsonl'
    Step 'run rich2 q2 (repeat)' @('-context', 'rich2', '-questions', 'q2', '-out', 'results-rich2-q2-b.jsonl') 'results-rich2-q2-b.jsonl'
    Step 'run thin q2' @('-context', 'thin', '-questions', 'q2') 'results-thin-q2.jsonl'
    Step 'compare noise: rich2-q2 vs rich2-q2-b' @('-compare', 'results-rich2-q2.jsonl,results-rich2-q2-b.jsonl')
    Step 'compare questions: thin-q1 vs thin-q2' @('-compare', 'results-thin.jsonl,results-thin-q2.jsonl')
    Step 'compare context: thin-q2 vs rich2-q2' @('-compare', 'results-thin-q2.jsonl,results-rich2-q2.jsonl')
    Add-Content $log "`nDONE" -Encoding UTF8
} catch {
    Add-Content $log "`nABORTED: $_" -Encoding UTF8
} finally {
    Remove-Item $exe -ErrorAction SilentlyContinue
}
