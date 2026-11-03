param(
    [Parameter(Mandatory)][ValidatePattern('^\d+\.\d+\.\d+(?:-meta\.\d+)?$')][string]$Version,
    [Parameter(Mandatory)][string]$OutDir,
    [Parameter(Mandatory)][string]$NotesFile,
    [string]$ReleaseRepository = 'AlexbeatsZ/magpie'
)
$ErrorActionPreference = 'Stop'
$repoPath = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$outputPath = [IO.Path]::GetFullPath($OutDir)
$notesPath = [IO.Path]::GetFullPath($NotesFile)
if ($env:OS -ne 'Windows_NT') { throw 'Build the desktop release on Windows.' }
if (-not (Test-Path -LiteralPath $notesPath -PathType Leaf)) { throw 'Release notes are required.' }
if ((Test-Path -LiteralPath $outputPath) -and (Get-ChildItem -LiteralPath $outputPath -Force | Select-Object -First 1)) {
    throw "Output must be empty to prevent stale release files: $outputPath"
}
New-Item -ItemType Directory -Path $outputPath -Force | Out-Null
$resourcePath = Join-Path $repoPath 'rsrc_windows_amd64.syso'
$launcherPath = Join-Path $repoPath 'internal/update/launcher-amd64.com'
foreach ($generatedPath in @($resourcePath, $launcherPath)) {
    if (Test-Path -LiteralPath $generatedPath) { throw "Refusing to overwrite build input: $generatedPath" }
}
$previous = @{}
foreach ($name in @('GOOS','GOARCH','CGO_ENABLED')) { $previous[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
Push-Location -LiteralPath $repoPath
try {
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    $numericVersion = ($Version -replace '-meta\.', '.')
    if (($numericVersion -split '\.').Count -eq 3) { $numericVersion += '.0' }
    go run github.com/tc-hib/go-winres@v0.3.3 make --in build/windows/winres.json --arch amd64 --out rsrc --file-version $numericVersion --product-version $numericVersion
    if ($LASTEXITCODE -ne 0) { throw 'Resource generation failed.' }
    go build -trimpath -buildvcs=false -ldflags '-s -w' -o $launcherPath ./internal/update/launcher
    if ($LASTEXITCODE -ne 0) { throw 'Console launcher build failed.' }
    $exePath = Join-Path $outputPath 'magpie.exe'
    go build -tags production,comlauncher -trimpath -buildvcs=false -ldflags "-s -w -X main.version=$Version -H=windowsgui" -o $exePath .
    if ($LASTEXITCODE -ne 0) { throw 'Desktop build failed.' }
    Copy-Item -LiteralPath (Join-Path $repoPath 'LICENSE'), (Join-Path $repoPath 'THIRD_PARTY_NOTICES') -Destination $outputPath
    $zipPath = Join-Path $outputPath 'magpie-windows-amd64.zip'
    Compress-Archive -LiteralPath $exePath, (Join-Path $outputPath 'LICENSE'), (Join-Path $outputPath 'THIRD_PARTY_NOTICES') -DestinationPath $zipPath
    $assetPath = Join-Path $outputPath 'magpie-windows-amd64.exe'
    Move-Item -LiteralPath $exePath -Destination $assetPath
    $releaseURL = "https://github.com/$ReleaseRepository/releases/tag/v$Version"
    $assets = [ordered]@{}
    foreach ($file in @($assetPath, $zipPath)) {
        $item = Get-Item -LiteralPath $file
        $assets[$item.Name] = [ordered]@{
            url = "https://github.com/$ReleaseRepository/releases/download/v$Version/$($item.Name)"
            size = $item.Length
            sha256 = (Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant()
        }
    }
    $notes = Get-Content -LiteralPath $notesPath -Raw -Encoding utf8
    $utf8 = [Text.UTF8Encoding]::new($false)
    $feed = [ordered]@{ version = $Version; notes = $notes; url = $releaseURL; assets = $assets }
    [IO.File]::WriteAllText((Join-Path $outputPath 'update.json'), ($feed | ConvertTo-Json -Depth 8), $utf8)
    $history = @{ releases = @(@{ version = $Version; notes = $notes; url = $releaseURL }) }
    [IO.File]::WriteAllText((Join-Path $outputPath 'notes.json'), ($history | ConvertTo-Json -Depth 8), $utf8)
    Get-Item -LiteralPath $assetPath, $zipPath | Select-Object Name, Length
} finally {
    foreach ($generatedPath in @($resourcePath, $launcherPath)) {
        if (Test-Path -LiteralPath $generatedPath) { Remove-Item -LiteralPath $generatedPath -Force }
    }
    foreach ($name in $previous.Keys) { [Environment]::SetEnvironmentVariable($name, $previous[$name], 'Process') }
    Pop-Location
}
