<#
Install WinDivert driver and copy DLL for testing.

Place this script in the repository (support-files) and run it from an
Administrator PowerShell prompt:

  Start-Process powershell -Verb runAs -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File "./support-files/install-windivert.ps1"'

This script will:
 - locate WinDivert.sys and WinDivert.dll under support-files
 - copy WinDivert.dll to the project's target directory (./target)
 - register the driver service 'WinDivert' using sc.exe
 - start the driver and print its status
#>

Write-Host "Installing WinDivert from support-files..."

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Definition


# Find driver sys (allow WinDivert.sys or WinDivert64.sys etc.)
$sys = Get-ChildItem -Path $scriptRoot -Recurse -Include 'WinDivert*.sys' -ErrorAction SilentlyContinue | Select-Object -First 1
if ($null -ne $sys) { $sysPath = $sys.FullName } else { $sysPath = $null }

# Find DLL
$dll = Get-ChildItem -Path $scriptRoot -Recurse -Include 'WinDivert.dll' -ErrorAction SilentlyContinue | Select-Object -First 1
if ($null -ne $dll) { $dllPath = $dll.FullName } else { $dllPath = $null }

if (-not $sysPath) {
    Write-Error "WinDivert driver (.sys) not found under $scriptRoot. Place the driver file (WinDivert.sys or WinDivert64.sys) in support-files and retry."
    exit 2
}
if (-not $dllPath) {
    Write-Warning "WinDivert.dll not found under $scriptRoot. You may still be able to install the driver, but DLL will not be copied to target."
}

$targetDir = Resolve-Path (Join-Path $scriptRoot '..\target') -ErrorAction SilentlyContinue
if (-not $targetDir) {
    $targetDir = New-Item -ItemType Directory -Path (Join-Path $scriptRoot '..\target') -Force
    $targetDir = Resolve-Path $targetDir
}
$targetDir = $targetDir.Path

if ($dllPath) {
    Write-Host "Copying WinDivert.dll to $targetDir"
    Copy-Item -Path $dllPath -Destination $targetDir -Force
}

Write-Host "Registering WinDivert driver: $sysPath"

# If service exists, stop and delete it first
$serviceQuery = & sc.exe query WinDivert 2>&1
if ($serviceQuery -match 'FAILED') {
    # service does not exist
} else {
    if ($serviceQuery -match 'STATE') {
        Write-Host "Existing WinDivert service found. Attempting to stop and delete..."
        sc.exe stop WinDivert | Out-Null
        Start-Sleep -Seconds 1
        sc.exe delete WinDivert | Out-Null
        Start-Sleep -Seconds 1
    }
}

# Build argument list for sc create with properly quoted binPath
$createArgs = @(
    'create',
    'WinDivert',
    "binPath=`"$sysPath`"",
    'type= kernel',
    'start= demand'
)
Write-Host "sc.exe" ($createArgs -join ' ')
Start-Process -FilePath sc.exe -ArgumentList $createArgs -NoNewWindow -Wait

Write-Host "Starting WinDivert driver..."
Start-Process -FilePath sc.exe -ArgumentList @('start','WinDivert') -NoNewWindow -Wait

Write-Host "Driver status:"
Start-Process -FilePath sc.exe -ArgumentList @('query','WinDivert') -NoNewWindow -Wait

Write-Host "Installation complete. Now run the CLI in an Administrator PowerShell to test:"
Write-Host "  .\target\chaoblade-win.exe create net delay 10 --filter \"outbound and tcp\""
