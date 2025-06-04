# HNS-Standalone v2.0 - Epiroc VM Standard Hostname Generator

> **Lightweight, database-free hostname generator that returns only available hostnames**

## 🎯 What it does

- **Generates** Epiroc VM Standard compliant hostnames (15 characters)
- **Checks availability** across DNS, ICMP, Active Directory, ServiceNow CMDB
- **Auto-increments** sequence numbers until finding available hostname
- **Returns immediately usable** hostnames with no reservations needed

## 🚀 Quick Start

```powershell
# Import module
Import-Module .\HNS-Standalone\HNS-Standalone.psm1

# Get next available hostname
$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS'

# Use immediately
Write-Host "Deploy to: $($hostname.Hostname)"  # e.g., SFDSMWEEUPAS001
```

## 📋 Parameters

| Parameter | Values | Description |
|-----------|---------|-------------|
| **UnitCode** | SFD, USS, DAL, etc. | 3-char business unit code |
| **Provider** | E, M, A, G | E=Epiroc, M=Azure, A=AWS, G=Google |
| **Region** | WEEU, SCUS, EUSE, etc. | 4-char region code |
| **Environment** | P, T, D | P=Prod, T=Test, D=Dev |
| **Function** | AS, DB, WS, BU, etc. | 2-char server function |

## 💡 Examples

### Basic Azure Production App Server
```powershell
$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS'
# Returns: SFDSMWEEUPAS001 (or next available)
```

### Test Database Starting from Sequence 10
```powershell
$hostname = Get-AvailableHostname -UnitCode 'USS' -Provider 'M' -Region 'SCUS' -Environment 'T' -Function 'DB' -StartSequence 10
# Returns: USSSMSCUSTDB010 (or next available from 10+)
```

### Epiroc Private Cloud Backup
```powershell
$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'E' -Region 'EUSE' -Environment 'P' -Function 'BU'
# Returns: SFDSEUSEPBU001 (or next available)
```

### Skip Active Directory Check
```powershell
$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'D' -Function 'WS' -CheckActiveDirectory $false
# Only checks DNS and ICMP
```

### With ServiceNow CMDB Check
```powershell
$cred = Get-Credential
$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS' `
            -CheckServiceNow -ServiceNowInstance 'https://company.servicenow.com' -ServiceNowCredential $cred
```

## 🔍 What Gets Checked

| System | Check Type | Purpose |
|--------|------------|---------|
| **DNS** | `Resolve-DnsName` | Hostname resolves to IP |
| **ICMP** | `Test-Connection` | Host responds to ping |
| **Active Directory** | `Get-ADComputer` | Computer object exists |
| **ServiceNow CMDB** | REST API | CI record exists |

## ✅ Return Object

```powershell
@{
    Hostname = "SFDSMWEEUPAS001"    # Ready-to-use hostname
    UnitCode = "SFD"                # Breakdown of components
    Provider = "M"
    Region = "WEEU"
    Environment = "P"
    Function = "AS"
    Sequence = 1
    Length = 15                     # Always 15 for Epiroc standard
    IsValid = $true
    SequencesTested = 1             # How many sequences checked
    ChecksPerformed = @{            # What was verified
        DNS = $true
        ICMP = $true
        ActiveDirectory = $true
        ServiceNow = $false
    }
}
```

## 🎛️ Advanced Options

```powershell
Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS' `
    -StartSequence 50 `           # Start checking from sequence 50
    -MaxSequence 100 `            # Stop checking at sequence 100
    -CheckDNS $true `             # Check DNS (default: true)
    -CheckICMP $true `            # Check ping (default: true)
    -CheckActiveDirectory $false ` # Skip AD check
    -CheckServiceNow $true `      # Enable ServiceNow check
    -ServiceNowInstance 'https://company.servicenow.com' `
    -ServiceNowCredential $cred
```

## 🏗️ Epiroc VM Standard Structure

```
Position:  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15
Format:    [Unit Code][T][P][   Region   ][E][Function][Seq]
Example:   S  F  D  S  M  W  E  E  U  P  A  S  0  0  1

Where:
- Unit Code (3): Business unit identifier
- Type (1): Always 'S' for Server  
- Provider (1): E/M/A/G for cloud provider
- Region (4): Geographic/datacenter location
- Environment (1): P/T/D for environment
- Function (2): Server role/purpose
- Sequence (3): Auto-incrementing 001-999
```

## 🛠️ No Database Required

Unlike the full HNS system, this module:
- ❌ No PostgreSQL database
- ❌ No reservations or lifecycle management  
- ❌ No commit/release workflows
- ✅ **Just returns available hostnames immediately**

Perfect for:
- **CI/CD pipelines** - Get hostname and deploy
- **Provisioning scripts** - Immediate hostname assignment
- **Infrastructure automation** - No state management needed

## 🔧 Installation

1. **Download** the module files
2. **Import** the module: `Import-Module .\HNS-Standalone\HNS-Standalone.psm1`
3. **Start using**: `Get-AvailableHostname` or aliases `Get-Hostname`, `New-Hostname`

## 📞 Quick Reference

```powershell
# Aliases available
Get-Hostname      # Same as Get-AvailableHostname
New-Hostname      # Same as Get-AvailableHostname

# Help
Get-Help Get-AvailableHostname -Examples
Get-Help Get-AvailableHostname -Full
```

---

**Ready to use!** Every call returns an immediately deployable hostname that's guaranteed available across all checked systems.