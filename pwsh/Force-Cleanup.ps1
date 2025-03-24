# Force-Cleanup.ps1
# Script to forcibly clean up database issues by handling all hostnames before template deletion

# Import the module
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$modulePath = Join-Path -Path $scriptDir -ChildPath "HNS-API.psm1"
if (Test-Path $modulePath) {
    Import-Module -Name $modulePath -Force -Verbose
} else {
    Write-Error "Module file not found at: $modulePath"
    exit 1
}

# Set variables
$HnsUrl = "http://localhost:8080"  # Change to your HNS server URL
$Username = "test"                  # Change to your username
$Password = ConvertTo-SecureString "Logon123!" -AsPlainText -Force  # Change to your password

# Template IDs to force clean - add the problematic ID
$TemplateIdsToClean = @(27)  # Add more IDs if needed

function Write-ColorOutput {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message,
        
        [Parameter(Mandatory = $false)]
        [System.ConsoleColor]$ForegroundColor = [System.ConsoleColor]::White
    )
    
    $originalColor = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $host.UI.RawUI.ForegroundColor = $originalColor
}

function Force-CleanTemplate {
    param(
        [Parameter(Mandatory = $true)]
        [int]$TemplateId,
        
        [Parameter(Mandatory = $false)]
        [switch]$SkipConfirmation
    )
    
    Write-ColorOutput "Cleaning template ID $TemplateId..." -ForegroundColor Cyan
    
    try {
        # Check if the template exists
        $template = Get-HnsTemplate -Id $TemplateId -ErrorAction Stop
        Write-ColorOutput "Found template: $($template.name) (ID: $TemplateId)" -ForegroundColor Green
        
        # Get all hostnames for this template
        Write-ColorOutput "Fetching all hostnames for template..." -ForegroundColor Yellow
        $hostnames = Get-HnsHostname -TemplateId $TemplateId -Limit 1000 -ErrorAction Stop
        
        if ($hostnames -and $hostnames.Count -gt 0) {
            Write-ColorOutput "Found $($hostnames.Count) hostnames for template ID $TemplateId" -ForegroundColor Yellow
            
            if (-not $SkipConfirmation) {
                $confirmation = Read-Host "Do you want to process and release all these hostnames? (y/n)"
                if ($confirmation -ne 'y') {
                    Write-ColorOutput "Operation cancelled." -ForegroundColor Yellow
                    return $false
                }
            }
            
            # First, commit all reserved hostnames
            $reservedHostnames = $hostnames | Where-Object { $_.status -eq "reserved" }
            foreach ($hostname in $reservedHostnames) {
                try {
                    Write-ColorOutput "Committing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                    Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false
                    
                    # Add delay between API calls to prevent overwhelming the server
                    Start-Sleep -Milliseconds 300
                }
                catch {
                    Write-ColorOutput "Failed to commit hostname $($hostname.id): $_" -ForegroundColor Red
                }
            }
            
            # Now, release all committed hostnames (including the ones we just committed)
            # We need to fetch the list again to get updated statuses
            $hostnames = Get-HnsHostname -TemplateId $TemplateId -Limit 1000 -ErrorAction Stop
            $committedHostnames = $hostnames | Where-Object { $_.status -eq "committed" }
            
            foreach ($hostname in $committedHostnames) {
                try {
                    Write-ColorOutput "Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                    Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                    
                    # Add delay between API calls
                    Start-Sleep -Milliseconds 300
                }
                catch {
                    Write-ColorOutput "Failed to release hostname $($hostname.id): $_" -ForegroundColor Red
                }
            }
            
            # Add a pause to let any database transactions complete
            Write-ColorOutput "Waiting for database to catch up..." -ForegroundColor Yellow
            Start-Sleep -Seconds 3
        }
        else {
            Write-ColorOutput "No hostnames found for template ID $TemplateId" -ForegroundColor Green
        }
        
        # Now try to delete the template
        Write-ColorOutput "Attempting to delete template ID $TemplateId..." -ForegroundColor Cyan
        Remove-HnsTemplate -Id $TemplateId -Confirm:$false
        Write-ColorOutput "Template ID $TemplateId deleted successfully!" -ForegroundColor Green
        return $true
    }
    catch {
        Write-ColorOutput "Failed to clean template ID $TemplateId: $_" -ForegroundColor Red
        return $false
    }
}

# Main execution
try {
    # Initialize connection to the HNS server
    Write-ColorOutput "Initializing connection to HNS server at $HnsUrl..." -ForegroundColor Cyan
    Initialize-HnsConnection -BaseUrl $HnsUrl
    
    # Authenticate
    Write-ColorOutput "Authenticating with username $Username..." -ForegroundColor Cyan
    Connect-HnsService -Username $Username -Password $Password
    
    # Test the connection
    $connected = Test-HnsConnection
    if (-not $connected) {
        throw "Failed to connect to HNS server."
    }
    
    # Process each template to clean
    foreach ($templateId in $TemplateIdsToClean) {
        $success = Force-CleanTemplate -TemplateId $templateId
        
        if ($success) {
            Write-ColorOutput "Successfully cleaned and deleted template ID $templateId" -ForegroundColor Green
        } else {
            Write-ColorOutput "Failed to complete cleanup for template ID $templateId" -ForegroundColor Red
        }
    }
    
    Write-ColorOutput "Force cleanup process complete!" -ForegroundColor Green
}
catch {
    Write-ColorOutput "An error occurred during cleanup: $_" -ForegroundColor Red
    Write-ColorOutput $_.ScriptStackTrace -ForegroundColor Red
}