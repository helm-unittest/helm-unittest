
$PROJECT_NAME = "helm-unittest"
$PROJECT_GH = "helm-unittest/$PROJECT_NAME"
$PROJECT_CHECKSUM_FILE = "$PROJECT_NAME-checksum.sha"
$HELM_PLUGIN_PATH = $env:HELM_PLUGIN_DIR

if ($env:SKIP_BIN_INSTALL -eq "1") {
    Write-Host "Skipping binary install"
    exit 0
}

if ($env:SKIP_BIN_DOWNLOAD -eq "1") {
    Write-Host "Preparing to install into $HELM_PLUGIN_PATH"
    Copy-Item -Path "plugin.yaml" -Destination "$HELM_PLUGIN_PATH\plugin.yaml" -Force
    Copy-Item -Path "untt.exe" -Destination "$HELM_PLUGIN_PATH\untt.exe" -Force
    Write-Host "$PROJECT_NAME installed into $HELM_PLUGIN_PATH"
    exit 0
}

# initArch discovers the architecture for this system.
function Initialize-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "x86" { return "386" }
        "ARM64" { return "arm64" }
        default { return "amd64" }
    }
}

# initOS discovers the operating system for this system.
function Initialize-OS {
    $os = $($env:OS).ToLower()
    return $os
}

# verifySupported checks that the os/arch combination is supported for binary builds.
function Test-SupportedPlatform {
    param($OS, $ARCH)
    
    $supported = @(
        "windows-amd64", "windows_nt-amd64"
    )
    
    $platform = "$OS-$ARCH"
    if ($platform -notin $supported) {
        Write-Error "No prebuild binary for $OS-$ARCH."
        exit 1
    }
}

# getDownloadURL checks the latest available version.
function Get-DownloadURL {
    param($OS, $ARCH)
    
    Push-Location $HELM_PLUGIN_PATH

    # Try to get version from git
    $version = $null
    try {
        $gitOutput = & git describe --tags --abbrev=0 2>$null
        if ($LASTEXITCODE -eq 0) {
            $version = $gitOutput
        }
    }
    catch {
        # Git command failed, ignore
    }
    
    # If no version found, try to fetch from plugin.yaml
    if (-not $version) {
        Write-Host "No version found from git"
        try {
            $pluginContent = Get-Content "plugin.yaml" -Raw
            if ($pluginContent -match 'version:\s*["`'']?([^"`'']*)["`'']?') {
                $version = "v$($matches[1])"
            }
        }
        catch {
            Write-Error "Could not determine version from plugin.yaml"
            exit 1
        }
    }

    Pop-Location

    # Setup Download URL
    $versionNumber = $version -replace '^v', ''
    $downloadUrl = "https://github.com/$PROJECT_GH/releases/download/$version/$PROJECT_NAME-$OS-$ARCH-$versionNumber.tgz"
    
    # Setup Checksum URL
    $checksumUrl = "https://github.com/$PROJECT_GH/releases/download/$version/$PROJECT_CHECKSUM_FILE"
    
    return @{
        DownloadUrl = $downloadUrl
        ChecksumUrl = $checksumUrl
    }
}

# downloadFile downloads the latest binary package and also the checksum
function Get-DownloadFile {
    param($DownloadUrl)
    
    $tempFolder = "$env:TEMP\_dist"
    if (Test-Path $tempFolder) {
        Remove-Item $tempFolder -Recurse -Force
    }
    New-Item -ItemType Directory -Path $tempFolder -Force | Out-Null
    
    Write-Host "Downloading $DownloadUrl to location $tempFolder"
    
    $fileName = Split-Path $DownloadUrl -Leaf
    $filePath = Join-Path $tempFolder $fileName
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $filePath -UseBasicParsing -MaximumRedirection 0
        return $filePath
    }
    catch {
        Write-Error "Failed to download file: $_"
        exit 1
    }
}

# installFile verifies the SHA256 for the file, then unpacks and installs it.
function Install-File {
    param($FilePath, $ChecksumUrl)
    
    $fileName = Split-Path $FilePath -Leaf
    $folderName = Split-Path $FilePath -Parent

    # Validate checksum if URL is provided
    if ($ChecksumUrl) {
        Write-Host "Validating Checksum."
        try {
            $checksumContent = Invoke-WebRequest -Uri $ChecksumUrl -UseBasicParsing -MaximumRedirection 0 | Select-Object -ExpandProperty RawContent
            $checksumLine = $checksumContent -split "`n" | Where-Object { $_ -match [regex]::Escape($fileName) }
            
            Write-Host "Checksum Line: $checksumLine"
            if ($checksumLine) {
                $expectedHash = ($checksumLine -split '\s+')[0]
                $actualHash = Get-FileHash -Path $FilePath -Algorithm SHA256 | Select-Object -ExpandProperty Hash
                
                if ($expectedHash.ToUpper() -ne $actualHash.ToUpper()) {
                    Write-Error "Checksum validation failed"
                    exit 1
                }
                Write-Host "Checksum validation successful"
            }
            else {
                Write-Host "No checksum found for file"
            }
        }
        catch {
            Write-Warning "Checksum validation failed: $_"
        }
    }
    else {
        Write-Host "No Checksum validated."
    }
    
    Write-Host "Preparing to install into $HELM_PLUGIN_PATH"

    Push-Location $folderName
    & tar -xzf $fileName -C .

    Copy-Item -Path *.* -Exclude $fileName -Destination $HELM_PLUGIN_PATH -Force
    Pop-Location

    Remove-Item "$env:TEMP\_dist" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "$PROJECT_NAME installed into $HELM_PLUGIN_PATH"    
}

# testVersion tests the installed client to make sure it is working.
function Test-Version {
    try {
        $unttPath = Join-Path $HELM_PLUGIN_PATH "untt.exe"
        if (Test-Path $unttPath) {
            & $unttPath -h
        }
        else {
            Write-Warning "untt.exe not found at expected location"
        }
    }
    catch {
        Write-Warning "Failed to test version: $_"
    }
}

# Main execution
try {
    $ARCH = Initialize-Architecture
    $OS = Initialize-OS
    Test-SupportedPlatform -OS $OS -ARCH $ARCH
    
    $urls = Get-DownloadURL -OS $OS -ARCH $ARCH
    $downloadedFile = Get-DownloadFile -DownloadUrl $urls.DownloadUrl
    Install-File -FilePath $downloadedFile -ChecksumUrl $urls.ChecksumUrl
    Test-Version
    
    Write-Host "Installation completed successfully"
}
catch {
    Write-Error "Failed to install $PROJECT_NAME"
    Write-Host "For support, go to https://github.com/helm-unittest/helm-unittest/blob/main/FAQ.md"
    exit 1
}