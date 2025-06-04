# Code Signing Guide for HNS-Standalone Module

## Overview
Code signing helps ensure the integrity and authenticity of PowerShell scripts and modules. This guide explains how to sign the HNS-Standalone module.

## Quick Start

### Option 1: Self-Signed Certificate (Testing)
```powershell
# Navigate to the module directory
cd C:\Users\eitxdr\Desktop\projects\HNS-Standalone

# Run the signing script with self-signed option
.\Sign-Module.ps1 -CreateSelfSigned
```

### Option 2: Existing Certificate
```powershell
# List available certificates
Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert

# Sign with specific certificate
.\Sign-Module.ps1 -CertificateThumbprint "YOUR_CERT_THUMBPRINT"
```

## Certificate Types

### 1. Self-Signed Certificates
- **Use Case**: Development and testing
- **Validity**: Created script sets 5 years
- **Trust**: Only trusted on local machine
- **Cost**: Free

### 2. Commercial Code Signing Certificates
- **Use Case**: Production distribution
- **Providers**: DigiCert, Sectigo, GlobalSign
- **Trust**: Trusted by Windows by default
- **Cost**: $200-500/year

### 3. Internal CA Certificates
- **Use Case**: Enterprise environments
- **Trust**: Within organization's domain
- **Cost**: Free (if CA exists)

## Manual Signing Process

### 1. Get a Certificate
```powershell
# View existing certificates
Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert

# Or create self-signed
$cert = New-SelfSignedCertificate `
    -Subject "CN=Your Name" `
    -Type CodeSigningCert `
    -CertStoreLocation Cert:\CurrentUser\My
```

### 2. Sign Individual Files
```powershell
# Sign a single file
Set-AuthenticodeSignature `
    -FilePath .\HNS-Standalone.psm1 `
    -Certificate $cert `
    -TimestampServer "http://timestamp.digicert.com"
```

### 3. Verify Signatures
```powershell
# Check signature status
Get-AuthenticodeSignature .\HNS-Standalone.psm1
```

## Execution Policy Considerations

After signing, you can use more restrictive policies:
```powershell
# AllSigned - Only signed scripts allowed
Set-ExecutionPolicy AllSigned -Scope CurrentUser

# RemoteSigned - Local scripts can be unsigned
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Troubleshooting

### Certificate Not Trusted
```powershell
# Add certificate to trusted stores (self-signed only)
$cert = Get-ChildItem Cert:\CurrentUser\My | Where {$_.Subject -match "HNS"}
$store = New-Object System.Security.Cryptography.X509Certificates.X509Store "TrustedPublisher", "CurrentUser"
$store.Open("ReadWrite")
$store.Add($cert)
$store.Close()
```

### Signature Invalid After Edit
- Any change to a signed file invalidates the signature
- Must re-sign after modifications
- Consider signing as final step before distribution

### Timestamp Server Issues
Alternative timestamp servers:
- http://timestamp.comodoca.com
- http://timestamp.sectigo.com
- http://time.certum.pl

## Best Practices

1. **Always use timestamps** - Ensures signatures remain valid after certificate expiry
2. **Sign all module files** - Include .psm1, .psd1, and all .ps1 files
3. **Document certificate details** - Keep thumbprint and expiry date recorded
4. **Test after signing** - Verify module loads correctly
5. **Backup certificates** - Export certificates with private keys for safekeeping

## Enterprise Deployment

For organization-wide deployment:
1. Use enterprise CA or commercial certificate
2. Deploy certificate to all machines via GPO
3. Set execution policy via GPO
4. Consider using catalog signing for better performance