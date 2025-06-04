#Requires -Version 5.1
<#
.SYNOPSIS
    HNS-Standalone PowerShell Module - Epiroc VM Standard Hostname Generator
.DESCRIPTION
    Lightweight PowerShell module for generating available hostnames following the Epiroc VM Standard.
    Checks DNS, ICMP, Active Directory, and ServiceNow CMDB for hostname availability.
    No database required - returns only available hostnames.
.NOTES
    Version: 2.0.0
    Author: HNS Development Team
    Focus: Availability-first hostname generation
#>

# Get public function definitions
$Public = @(Get-ChildItem -Path $PSScriptRoot\Source\Public\Hostname\*.ps1 -ErrorAction SilentlyContinue)

# Dot source the public functions
foreach ($import in $Public) {
    if (Test-Path $import) {
        try {
            Write-Verbose "Importing $($import.BaseName)"
            . $import.FullName
        }
        catch {
            Write-Error "Failed to import $($import.FullName): $_"
            throw
        }
    }
}

# Create aliases for convenience
Set-Alias -Name Get-Hostname -Value Get-AvailableHostname
Set-Alias -Name New-Hostname -Value Get-AvailableHostname

# Export functions and aliases
Export-ModuleMember -Function Get-AvailableHostname -Alias Get-Hostname, New-Hostname

# Display module information on load
$moduleInfo = @"
HNS-Standalone Module v2.0.0 Loaded
===================================
Epiroc VM Standard Hostname Generator
• No database required
• Checks DNS, ICMP, AD, ServiceNow
• Returns only available hostnames
• Auto-increments sequences

Get started: Get-Help Get-AvailableHostname -Examples
"@

Write-Host $moduleInfo -ForegroundColor Green