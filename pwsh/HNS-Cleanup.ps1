# HNS-Cleanup.ps1
# Cleanup script to remove resources created by HNS-Demo.ps1

# Get the script's directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Import the module with an absolute path to ensure it's found
$modulePath = Join-Path -Path $scriptDir -ChildPath "HNS-API.psm1"
if (Test-Path $modulePath) {
    Import-Module -Name $modulePath -Force -Verbose
} else {
    Write-Error "Module file not found at: $modulePath"
    exit 1
}

# Set variables
$HnsUrl = "http://localhost:8080"  # Change to your HNS server URL
$Username = "test"                # Change to your username
$Password = ConvertTo-SecureString "Logon123!" -AsPlainText -Force  # Change to your password

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

function Cleanup-Templates {
    param(
        [Parameter(Mandatory = $false)]
        [string]$NamePattern = "ServerTemplate"
    )
    
    Write-ColorOutput "Looking for templates with name pattern: $NamePattern" -ForegroundColor Cyan
    
    # Check if the function exists before trying to use it
    if (-not (Get-Command -Name Get-HnsTemplate -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Error: Get-HnsTemplate function not found. Is the module properly imported?" -ForegroundColor Red
        return
    }
    
    # Get all templates
    try {
        # Using the updated Get-HnsTemplate function
        $templates = Get-HnsTemplate -Limit 100
        
        # Filter templates by name pattern
        $matchingTemplates = $templates | Where-Object { $_.name -like "*$NamePattern*" }
        
        if (-not $matchingTemplates -or $matchingTemplates.Count -eq 0) {
            Write-ColorOutput "No matching templates found." -ForegroundColor Yellow
            return
        }
        
        Write-ColorOutput "Found $($matchingTemplates.Count) matching templates:" -ForegroundColor Green
        foreach ($template in $matchingTemplates) {
            Write-Output "  - ID: $($template.id), Name: $($template.name)"
        }
        
        # Ask for confirmation
        $confirmation = Read-Host "Do you want to delete these templates and their associated hostnames? (y/n)"
        if ($confirmation -ne 'y') {
            Write-ColorOutput "Template deletion cancelled." -ForegroundColor Yellow
            return
        }
        
        # First, find and release all hostnames associated with each template
        foreach ($template in $matchingTemplates) {
            Write-ColorOutput "Finding hostnames associated with template ID $($template.id)..." -ForegroundColor Yellow
            
            try {
                # Get all hostnames for this template
                $hostnames = Get-HnsHostname -TemplateId $template.id -Limit 1000
                
                if ($hostnames -and $hostnames.Count -gt 0) {
                    Write-ColorOutput "Found $($hostnames.Count) hostnames for template ID $($template.id)" -ForegroundColor Yellow
                    
                    # Process reserved hostnames - commit then release
                    $reservedHostnames = $hostnames | Where-Object { $_.status -eq "reserved" }
                    foreach ($hostname in $reservedHostnames) {
                        try {
                            Write-ColorOutput "Committing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false
                            
                            Start-Sleep -Milliseconds 500  # Small delay to prevent API overload
                            
                            Write-ColorOutput "Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                        }
                        catch {
                            Write-ColorOutput "Failed to process hostname $($hostname.id): $_" -ForegroundColor Red
                        }
                    }
                    
                    # Process committed hostnames - release
                    $committedHostnames = $hostnames | Where-Object { $_.status -eq "committed" }
                    foreach ($hostname in $committedHostnames) {
                        try {
                            Write-ColorOutput "Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                        }
                        catch {
                            Write-ColorOutput "Failed to release hostname $($hostname.id): $_" -ForegroundColor Red
                        }
                    }
                } else {
                    Write-ColorOutput "No hostnames found for template ID $($template.id)" -ForegroundColor Green
                }
            }
            catch {
                Write-ColorOutput "Failed to get hostnames for template $($template.id): $_" -ForegroundColor Red
            }
            
            # Add a small delay before trying to delete the template
            Start-Sleep -Seconds 1
        }
        
        # Now try to delete the templates
        foreach ($template in $matchingTemplates) {
            try {
                Write-ColorOutput "Deleting template: $($template.name) (ID: $($template.id))..." -ForegroundColor Cyan
                # Using the updated Remove-HnsTemplate function with ShouldProcess support
                Remove-HnsTemplate -Id $template.id -Confirm:$false
                Write-ColorOutput "Deleted template: $($template.name) (ID: $($template.id))" -ForegroundColor Green
            }
            catch {
                Write-ColorOutput "Failed to delete template $($template.id): $_" -ForegroundColor Red
            }
        }
    }
    catch {
        Write-ColorOutput "Error getting templates: $_" -ForegroundColor Red
    }
}

function Cleanup-Hostnames {
    param(
        [Parameter(Mandatory = $false)]
        [string]$NamePattern = "",
        
        [Parameter(Mandatory = $false)]
        [int]$TemplateId = 0
    )
    
    $filterMsg = "Looking for hostnames"
    if ($NamePattern) {
        $filterMsg += " with name pattern: $NamePattern"
    }
    if ($TemplateId -gt 0) {
        $filterMsg += " with template ID: $TemplateId"
    }
    
    Write-ColorOutput $filterMsg -ForegroundColor Cyan
    
    # Create filter params for the updated Get-HnsHostname function
    $hostnameParams = @{
        Limit = 100
        Offset = 0
    }
    
    if ($TemplateId -gt 0) {
        $hostnameParams["TemplateId"] = $TemplateId
    }
    
    # Get hostnames
    try {
        # Using the updated Get-HnsHostname function
        $hostnames = Get-HnsHostname @hostnameParams
        
        # Apply name filter if provided
        if ($NamePattern) {
            $hostnames = $hostnames | Where-Object { $_.name -like "*$NamePattern*" }
        }
        
        if (-not $hostnames -or $hostnames.Count -eq 0) {
            Write-ColorOutput "No matching hostnames found." -ForegroundColor Yellow
            return
        }
        
        Write-ColorOutput "Found $($hostnames.Count) matching hostnames:" -ForegroundColor Green
        foreach ($hostname in $hostnames) {
            Write-Output "  - ID: $($hostname.id), Name: $($hostname.name), Status: $($hostname.status)"
        }
        
        # Ask for confirmation
        $confirmation = Read-Host "Do you want to release these hostnames? (y/n)"
        if ($confirmation -ne 'y') {
            Write-ColorOutput "Hostname release cancelled." -ForegroundColor Yellow
            return
        }
        
        # Release hostnames (if committed)
        $committedHostnames = $hostnames | Where-Object { $_.status -eq "committed" }
        foreach ($hostname in $committedHostnames) {
            try {
                # Using the updated Set-HnsHostnameRelease function with ShouldProcess support
                Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                Write-ColorOutput "Released hostname: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Green
            }
            catch {
                Write-ColorOutput "Failed to release hostname $($hostname.id): $_" -ForegroundColor Red
            }
        }
        
        # For reserved hostnames, commit them first, then release them
        $reservedHostnames = $hostnames | Where-Object { $_.status -eq "reserved" }
        foreach ($hostname in $reservedHostnames) {
            try {
                # Using the updated Set-HnsHostnameCommit and Set-HnsHostnameRelease functions with ShouldProcess support
                Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false
                Write-ColorOutput "Committed hostname: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Yellow
                
                Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                Write-ColorOutput "Released hostname: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Green
            }
            catch {
                Write-ColorOutput "Failed to process hostname $($hostname.id): $_" -ForegroundColor Red
            }
        }
    }
    catch {
        Write-ColorOutput "Error getting hostnames: $_" -ForegroundColor Red
    }
}

function Cleanup-ApiKeys {
    param(
        [Parameter(Mandatory = $false)]
        [string]$NamePattern = "PowerShell-Demo"
    )
    
    Write-ColorOutput "Looking for API keys with name pattern: $NamePattern" -ForegroundColor Cyan
    
    try {
        # Get all API keys
        $apiKeys = Get-HnsApiKeys
        
        # Filter API keys by name pattern
        $matchingKeys = $apiKeys | Where-Object { $_.name -like "*$NamePattern*" }
        
        if (-not $matchingKeys -or $matchingKeys.Count -eq 0) {
            Write-ColorOutput "No matching API keys found." -ForegroundColor Yellow
            return
        }
        
        Write-ColorOutput "Found $($matchingKeys.Count) matching API keys:" -ForegroundColor Green
        foreach ($key in $matchingKeys) {
            Write-Output "  - ID: $($key.id), Name: $($key.name), Scope: $($key.scope)"
        }
        
        # Ask for confirmation
        $confirmation = Read-Host "Do you want to delete these API keys? (y/n)"
        if ($confirmation -ne 'y') {
            Write-ColorOutput "API key deletion cancelled." -ForegroundColor Yellow
            return
        }
        
        # Delete API keys
        foreach ($key in $matchingKeys) {
            try {
                # Using the updated Remove-HnsApiKey function with ShouldProcess support
                Remove-HnsApiKey -Id $key.id -Confirm:$false
                Write-ColorOutput "Deleted API key: $($key.name) (ID: $($key.id))" -ForegroundColor Green
            }
            catch {
                Write-ColorOutput "Failed to delete API key $($key.id): $_" -ForegroundColor Red
            }
        }
    }
    catch {
        Write-ColorOutput "Error getting API keys: $_" -ForegroundColor Red
    }
}

# Main cleanup script
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
    
    # Run cleanup operations
    Write-ColorOutput "`n===== Cleaning up resources created by HNS-Demo.ps1 =====" -ForegroundColor Magenta
    
    # 1. Clean up hostnames
    Write-ColorOutput "`n[1/3] Cleaning up hostnames..." -ForegroundColor Magenta
    Cleanup-Hostnames
    
    # 2. Clean up templates
    Write-ColorOutput "`n[2/3] Cleaning up templates..." -ForegroundColor Magenta
    Cleanup-Templates -NamePattern "ServerTemplate"
    
    # 3. Clean up API keys
    Write-ColorOutput "`n[3/3] Cleaning up API keys..." -ForegroundColor Magenta
    Cleanup-ApiKeys -NamePattern "PowerShell-Demo"
    
    Write-ColorOutput "`nCleanup completed successfully!" -ForegroundColor Green
}
catch {
    Write-ColorOutput "An error occurred during cleanup: $_" -ForegroundColor Red
    Write-ColorOutput $_.ScriptStackTrace -ForegroundColor Red
}