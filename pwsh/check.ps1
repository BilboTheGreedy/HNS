# Check-TemplateDependencies.ps1
# Script to check for template dependencies before deletion

param(
    [Parameter(Mandatory = $true)]
    [int]$TemplateId,
    
    [Parameter(Mandatory = $false)]
    [string]$HnsUrl = "http://localhost:8080",
    
    [Parameter(Mandatory = $false)]
    [string]$Username = "admin",
    
    [Parameter(Mandatory = $false)]
    [string]$Password = "admin123"
)

# Import the module
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$modulePath = Join-Path -Path $scriptDir -ChildPath "HNS-API.psm1"
Import-Module -Name $modulePath -Force

function Write-ColorOutput {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message,
        
        [Parameter(Mandatory = $false)]
        [ConsoleColor]$ForegroundColor = [ConsoleColor]::White
    )
    
    $originalColor = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $host.UI.RawUI.ForegroundColor = $originalColor
}

function Get-HostnameSummaryByStatus {
    param(
        [Parameter(Mandatory = $true)]
        [int]$TemplateId
    )
    
    $summaryByStatus = @{}
    
    # Get hostname counts by status
    foreach ($status in @("reserved", "committed", "released", "available")) {
        $hostnames = Get-HnsHostname -TemplateId $TemplateId -Status $status -Limit 1
        if ($hostnames -and $hostnames.Count -gt 0) {
            $summaryByStatus[$status] = $hostnames.total
        } else {
            $summaryByStatus[$status] = 0
        }
    }
    
    return $summaryByStatus
}

# Main script
try {
    # Initialize connection
    Write-ColorOutput "Connecting to HNS server at $HnsUrl..." -ForegroundColor Cyan
    Initialize-HnsConnection -BaseUrl $HnsUrl
    
    # Authenticate
    $secPassword = ConvertTo-SecureString $Password -AsPlainText -Force
    Connect-HnsService -Username $Username -Password $secPassword
    
    # Get template information
    Write-ColorOutput "Checking template ID $TemplateId..." -ForegroundColor Cyan
    $template = Get-HnsTemplate -Id $TemplateId
    
    if (-not $template) {
        Write-ColorOutput "Template ID $TemplateId not found!" -ForegroundColor Red
        exit 1
    }
    
    Write-ColorOutput "Found template: $($template.name)" -ForegroundColor Green
    Write-ColorOutput "Description: $($template.description)" -ForegroundColor Green
    
    # Check for associated hostnames
    Write-ColorOutput "`nChecking for associated hostnames..." -ForegroundColor Cyan
    $hostnames = Get-HnsHostname -TemplateId $TemplateId -Limit 10
    
    if (-not $hostnames -or $hostnames.Count -eq 0) {
        Write-ColorOutput "No hostnames found using this template. It can be safely deleted." -ForegroundColor Green
        exit 0
    }
    
    # Get a summary by status
    $summary = Get-HostnameSummaryByStatus -TemplateId $TemplateId
    
    Write-ColorOutput "Found $($hostnames.total) hostnames using this template!" -ForegroundColor Yellow
    Write-ColorOutput "Hostname status summary:" -ForegroundColor Yellow
    foreach ($status in $summary.Keys) {
        $count = $summary[$status]
        if ($count -gt 0) {
            Write-ColorOutput "  - $status: $count" -ForegroundColor Yellow
        }
    }
    
    # Show a sample of hostnames
    Write-ColorOutput "`nSample of hostnames:" -ForegroundColor Cyan
    foreach ($hostname in $hostnames) {
        Write-ColorOutput "  - $($hostname.name) (ID: $($hostname.id), Status: $($hostname.status))" -ForegroundColor Gray
    }
    
    # Provide guidance on how to clean up
    Write-ColorOutput "`nTo delete this template, you must first remove all associated hostnames:" -ForegroundColor Magenta
    Write-ColorOutput "1. For RESERVED hostnames:" -ForegroundColor White
    Write-ColorOutput "   - Commit them: Set-HnsHostnameCommit -HostnameId <id>" -ForegroundColor Gray
    Write-ColorOutput "   - Then release them: Set-HnsHostnameRelease -HostnameId <id>" -ForegroundColor Gray
    
    Write-ColorOutput "2. For COMMITTED hostnames:" -ForegroundColor White
    Write-ColorOutput "   - Release them: Set-HnsHostnameRelease -HostnameId <id>" -ForegroundColor Gray
    
    Write-ColorOutput "`nAutomated cleanup options:" -ForegroundColor Green
    Write-ColorOutput "1. Use the cleanup script: ./HNS-Cleanup.ps1" -ForegroundColor Green
    Write-ColorOutput "2. Use Remove-HnsTemplate with the -Force parameter (if implemented)" -ForegroundColor Green
    
    Write-ColorOutput "`nAfter cleaning up all hostnames, you can delete the template with:" -ForegroundColor Cyan
    Write-ColorOutput "Remove-HnsTemplate -Id $TemplateId" -ForegroundColor Cyan
}
catch {
    Write-ColorOutput "Error: $_" -ForegroundColor Red
    Write-ColorOutput $_.ScriptStackTrace -ForegroundColor Red
    exit 1
}