#Requires -Version 5.1
<#
.SYNOPSIS
    Signs the HNS-Standalone PowerShell module and all its component files
.DESCRIPTION
    This script helps sign the HNS-Standalone module with a code signing certificate.
    It can use an existing certificate or help create a self-signed one for testing.
.PARAMETER CertificateThumbprint
    Thumbprint of an existing code signing certificate
.PARAMETER CreateSelfSigned
    Create a self-signed certificate for testing purposes
.PARAMETER ModulePath
    Path to the HNS-Standalone module (defaults to current directory)
.EXAMPLE
    .\Sign-Module.ps1 -CreateSelfSigned
    Creates a self-signed certificate and signs the module
.EXAMPLE
    .\Sign-Module.ps1 -CertificateThumbprint "1234567890ABCDEF"
    Signs the module with an existing certificate
#>

[CmdletBinding()]
param(
    [Parameter(ParameterSetName='ExistingCert')]
    [string]$CertificateThumbprint,
    
    [Parameter(ParameterSetName='SelfSigned')]
    [switch]$CreateSelfSigned,
    
    [Parameter()]
    [string]$ModulePath = $PSScriptRoot
)

function New-SelfSignedCodeSigningCert {
    Write-Host "Creating self-signed code signing certificate..." -ForegroundColor Yellow
    
    $cert = New-SelfSignedCertificate `
        -Subject "CN=HNS-Standalone Code Signing" `
        -Type CodeSigningCert `
        -KeySpec Signature `
        -KeyUsage DigitalSignature `
        -FriendlyName "HNS-Standalone Code Signing Certificate" `
        -CertStoreLocation "Cert:\CurrentUser\My" `
        -HashAlgorithm SHA256 `
        -KeyExportPolicy Exportable `
        -NotAfter (Get-Date).AddYears(5)
    
    # Add to Trusted Publishers and Root
    $store = New-Object System.Security.Cryptography.X509Certificates.X509Store "TrustedPublisher", "CurrentUser"
    $store.Open("ReadWrite")
    $store.Add($cert)
    $store.Close()
    
    $store = New-Object System.Security.Cryptography.X509Certificates.X509Store "Root", "CurrentUser"
    $store.Open("ReadWrite")
    $store.Add($cert)
    $store.Close()
    
    Write-Host "Certificate created successfully!" -ForegroundColor Green
    Write-Host "Thumbprint: $($cert.Thumbprint)" -ForegroundColor Cyan
    
    return $cert
}

function Get-CodeSigningCertificate {
    param([string]$Thumbprint)
    
    if ($Thumbprint) {
        $cert = Get-ChildItem -Path Cert:\CurrentUser\My -CodeSigningCert | 
            Where-Object { $_.Thumbprint -eq $Thumbprint }
    } else {
        Write-Host "Available code signing certificates:" -ForegroundColor Yellow
        $certs = Get-ChildItem -Path Cert:\CurrentUser\My -CodeSigningCert
        
        if ($certs.Count -eq 0) {
            Write-Host "No code signing certificates found!" -ForegroundColor Red
            return $null
        }
        
        $certs | ForEach-Object {
            Write-Host "`nThumbprint: $($_.Thumbprint)" -ForegroundColor Cyan
            Write-Host "Subject: $($_.Subject)"
            Write-Host "Expires: $($_.NotAfter)"
        }
        
        $selectedThumb = Read-Host "`nEnter the thumbprint of the certificate to use"
        $cert = $certs | Where-Object { $_.Thumbprint -eq $selectedThumb }
    }
    
    if (-not $cert) {
        Write-Host "Certificate not found!" -ForegroundColor Red
        return $null
    }
    
    return $cert
}

function Sign-PowerShellFiles {
    param(
        [System.Security.Cryptography.X509Certificates.X509Certificate2]$Certificate,
        [string]$Path
    )
    
    $filesToSign = @()
    
    # Get all PowerShell files
    $filesToSign += Get-ChildItem -Path $Path -Include "*.ps1", "*.psm1", "*.psd1" -Recurse
    
    Write-Host "`nFiles to sign:" -ForegroundColor Yellow
    $filesToSign | ForEach-Object { Write-Host "  - $($_.FullName)" }
    
    $signed = 0
    $failed = 0
    
    foreach ($file in $filesToSign) {
        try {
            Write-Host "`nSigning: $($file.Name)..." -NoNewline
            Set-AuthenticodeSignature -FilePath $file.FullName -Certificate $Certificate -TimestampServer "http://timestamp.digicert.com" | Out-Null
            Write-Host " Done!" -ForegroundColor Green
            $signed++
        }
        catch {
            Write-Host " Failed!" -ForegroundColor Red
            Write-Host "  Error: $_" -ForegroundColor Red
            $failed++
        }
    }
    
    Write-Host "`nSigning Summary:" -ForegroundColor Cyan
    Write-Host "  Signed: $signed files" -ForegroundColor Green
    if ($failed -gt 0) {
        Write-Host "  Failed: $failed files" -ForegroundColor Red
    }
}

# Main execution
Write-Host "HNS-Standalone Module Signing Tool" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# Get or create certificate
if ($CreateSelfSigned) {
    $cert = New-SelfSignedCodeSigningCert
} else {
    $cert = Get-CodeSigningCertificate -Thumbprint $CertificateThumbprint
}

if (-not $cert) {
    Write-Host "`nNo certificate selected. Exiting." -ForegroundColor Red
    exit 1
}

Write-Host "`nUsing certificate:" -ForegroundColor Green
Write-Host "  Subject: $($cert.Subject)"
Write-Host "  Thumbprint: $($cert.Thumbprint)"
Write-Host "  Expires: $($cert.NotAfter)"

# Sign files
$confirm = Read-Host "`nProceed with signing? (Y/N)"
if ($confirm -eq 'Y') {
    Sign-PowerShellFiles -Certificate $cert -Path $ModulePath
    
    Write-Host "`nModule signing complete!" -ForegroundColor Green
    Write-Host "You can now import the module without execution policy issues." -ForegroundColor Cyan
} else {
    Write-Host "Signing cancelled." -ForegroundColor Yellow
}