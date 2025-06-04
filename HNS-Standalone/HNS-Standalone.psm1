#Requires -Version 5.1
<#
.SYNOPSIS
    HNS-Standalone PowerShell Module
.DESCRIPTION
    Professional standalone Hostname Naming Service module for generating and managing
    standardized hostnames based on configurable templates.
.NOTES
    Version: 1.0.0
    Author: HNS Development Team
#>

# Get public and private function definitions
$Public  = @(Get-ChildItem -Path $PSScriptRoot\Source\Public\*.ps1 -Recurse -ErrorAction SilentlyContinue)
$Private = @(Get-ChildItem -Path $PSScriptRoot\Source\Private\*.ps1 -Recurse -ErrorAction SilentlyContinue)
$Enums   = @(Get-ChildItem -Path $PSScriptRoot\Source\Enums\*.ps1 -ErrorAction SilentlyContinue)

# Load classes in dependency order
$ClassOrder = @(
    "$PSScriptRoot\Source\Classes\TemplateGroup.ps1",
    "$PSScriptRoot\Source\Classes\Template.ps1", 
    "$PSScriptRoot\Source\Classes\Hostname.ps1",
    "$PSScriptRoot\Source\Classes\HNSConfiguration.ps1"
)

# Module-level variables
$script:Configuration = $null
$script:Templates = @()
$script:Hostnames = @()

# Dot source the files in order
foreach ($import in @($Enums + $ClassOrder + $Private + $Public)) {
    if (Test-Path $import) {
        try {
            Write-Verbose "Importing $import"
            . $import
        }
        catch {
            Write-Error "Failed to import $import`: $_"
            throw
        }
    }
}

# Create aliases
Set-Alias -Name ghns -Value Get-HNSHostname
Set-Alias -Name nhns -Value New-HNSHostname
Set-Alias -Name rhns -Value Reserve-HNSHostname
Set-Alias -Name ghnt -Value Get-HNSTemplate
Set-Alias -Name nhnt -Value New-HNSTemplate

# Export Public functions and specific functions
$exportedFunctions = $Public.BaseName + @('Get-HNSHostname', 'Test-HostnameAvailability')
Export-ModuleMember -Function $exportedFunctions -Alias *

# Display module information on load
$moduleInfo = @"
HNS-Standalone Module v1.0.0 Loaded
===================================
Initialize the module with: Initialize-HNSEnvironment
Get help with: Get-Help about_HNS-Standalone
"@

Write-Host $moduleInfo -ForegroundColor Cyan