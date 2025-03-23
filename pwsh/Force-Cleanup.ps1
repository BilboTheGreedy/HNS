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
        [int]$BatchSize = 20
    )
    
    Write-ColorOutput "Force cleaning template ID $TemplateId..." -ForegroundColor Magenta
    
    # Process in batches to handle large numbers of hostnames
    $offset = 0
    $continueProcessing = $true
    $processedCount = 0
    
    while ($continueProcessing) {
        # Get hostnames in batches
        try {
            $hostnameParams = @{
                TemplateId = $TemplateId
                Limit = $BatchSize
                Offset = $offset
            }
            
            Write-ColorOutput "Fetching hostnames batch (Offset: $offset, Limit: $BatchSize)..." -ForegroundColor DarkGray
            $hostnames = Get-HnsHostname @hostnameParams
            
            if (-not $hostnames -or $hostnames.Count -eq 0) {
                Write-ColorOutput "No more hostnames found." -ForegroundColor Green
                $continueProcessing = $false
                continue
            }
            
            Write-ColorOutput "Processing batch of $($hostnames.Count) hostnames..." -ForegroundColor Cyan
            
            foreach ($hostname in $hostnames) {
                Write-ColorOutput "Processing hostname: $($hostname.name) (ID: $($hostname.id), Status: $($hostname.status))" -ForegroundColor DarkGray
                
                try {
                    # Process based on status
                    switch ($hostname.status) {
                        "reserved" {
                            # Commit then release
                            Write-ColorOutput "  - Committing reserved hostname..." -ForegroundColor Yellow
                            Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false -ErrorAction Continue
                            Start-Sleep -Milliseconds 300
                            
                            Write-ColorOutput "  - Releasing hostname..." -ForegroundColor Yellow
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false -ErrorAction Continue
                        }
                        "committed" {
                            # Just release
                            Write-ColorOutput "  - Releasing committed hostname..." -ForegroundColor Yellow
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false -ErrorAction Continue
                        }
                        "available" {
                            # For available hostnames, try to reserve, commit, then release
                            Write-ColorOutput "  - Attempting to reserve, commit, then release 'available' hostname..." -ForegroundColor Yellow
                            try {
                                # Get template first to know what params we need
                                $template = Get-HnsTemplate -Id $TemplateId
                                
                                # Build minimum params required (could be enhanced)
                                $params = @{}
                                foreach ($group in $template.groups) {
                                    if ($group.is_required -and $group.validation_type -eq "list") {
                                        $values = $group.validation_value.Split(",")
                                        if ($values.Length -gt 0) {
                                            $params[$group.name] = $values[0].Trim()
                                        }
                                    }
                                }
                                
                                # Reserve with the params
                                $reserved = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params -RequestedBy "cleanup_script"
                                Start-Sleep -Milliseconds 300
                                
                                # Commit it
                                Set-HnsHostnameCommit -HostnameId $reserved.id -Confirm:$false
                                Start-Sleep -Milliseconds 300
                                
                                # Release it
                                Set-HnsHostnameRelease -HostnameId $reserved.id -Confirm:$false
                            }
                            catch {
                                Write-ColorOutput "  - Could not process 'available' hostname: $_" -ForegroundColor Red
                            }
                        }
                        default {
                            Write-ColorOutput "  - Unrecognized status '$($hostname.status)'. Skipping..." -ForegroundColor Red
                        }
                    }
                    
                    $processedCount++
                }
                catch {
                    Write-ColorOutput "  - Failed to process hostname $($hostname.id): $_" -ForegroundColor Red
                }
                
                # Small delay to prevent API overload
                Start-Sleep -Milliseconds 100
            }
            
            # Move to next batch
            $offset += $BatchSize
            
            # Give the API a break
            Write-ColorOutput "Batch complete. Processed $processedCount hostnames so far..." -ForegroundColor Green
            Start-Sleep -Seconds 1
        }
        catch {
            Write-ColorOutput "Error processing batch: $_" -ForegroundColor Red
            # Try next batch - might be a temporary issue
            $offset += $BatchSize
            Start-Sleep -Seconds 2
        }
    }
    
    Write-ColorOutput "Finished processing hostnames for template ID $TemplateId. Total processed: $processedCount" -ForegroundColor Green
    Write-ColorOutput "Waiting for database to catch up before deleting template..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5
    
    # Now try to delete the template
    try {
        Write-ColorOutput "Attempting to delete template ID $TemplateId..." -ForegroundColor Cyan
        Remove-HnsTemplate -Id $TemplateId -Confirm:$false
        Write-ColorOutput "Template ID $TemplateId deleted successfully!" -ForegroundColor Green
        return $true
    }
    catch {
        Write-ColorOutput "Failed to delete template ID $TemplateId: $_" -ForegroundColor Red
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