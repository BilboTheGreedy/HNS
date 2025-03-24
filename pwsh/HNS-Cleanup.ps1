# HNS-Cleanup.ps1
# Enhanced cleanup script to remove resources created by HNS-Demo.ps1

param(
    [Parameter(Mandatory = $false)]
    [string]$NamePattern = "",
    
    [Parameter(Mandatory = $false)]
    [int]$TemplateId = 0,

    [Parameter(Mandatory = $false)]
    [switch]$Force = $false,
    
    [Parameter(Mandatory = $false)]
    [switch]$WhatIf = $false,
    
    [Parameter(Mandatory = $false)]
    [ValidateSet("All", "Templates", "Hostnames", "ApiKeys")]
    [string]$CleanupType = "All"
)

# Get the script's directory and import the module
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$modulePath = Join-Path -Path $scriptDir -ChildPath "HNS-API.psm1"
if (Test-Path $modulePath) {
    Import-Module -Name $modulePath -Force
} else {
    Write-Error "Module file not found at: $modulePath"
    exit 1
}

# Set variables
$HnsUrl = "http://localhost:8080"  # Change to your HNS server URL
$Username = "admin"                # Change to your username
$Password = ConvertTo-SecureString "admin123" -AsPlainText -Force  # Change to your password

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
        [string]$NamePattern = "",
        [Parameter(Mandatory = $false)]
        [int]$TemplateId = 0
    )
    # Determine filter criteria
    if ($TemplateId -gt 0) {
        Write-ColorOutput "Looking for template with ID: $TemplateId" -ForegroundColor Cyan
        $templates = @(Get-HnsTemplate -Id $TemplateId -ErrorAction SilentlyContinue)
    } elseif ($NamePattern) {
        Write-ColorOutput "Looking for templates with name pattern: $NamePattern" -ForegroundColor Cyan
        $allTemplates = Get-HnsTemplate -Limit 100
        $templates = @($allTemplates | Where-Object { $_.name -like "*$NamePattern*" })
    } else {
        Write-ColorOutput "No template filter specified. Use -TemplateId or -NamePattern to filter templates." -ForegroundColor Yellow
        return
    }
    
    if ($templates.Count -eq 0) {
        Write-ColorOutput "No matching templates found." -ForegroundColor Yellow
        return
    }
    
    Write-ColorOutput "Found $($templates.Count) matching templates:" -ForegroundColor Green
    foreach ($template in $templates) {
        Write-Output "  - ID: $($template.id), Name: $($template.name)"
    }
    
    # Ask for confirmation unless Force is used
    if (-not $Force) {
        $confirmation = Read-Host "Do you want to clean up these templates and their associated hostnames? (y/n)"
        if ($confirmation -ne 'y') {
            Write-ColorOutput "Template cleanup cancelled." -ForegroundColor Yellow
            return
        }
    }
    
    # Process each template
    foreach ($template in $templates) {
        Write-ColorOutput "`nProcessing template: $($template.name) (ID: $($template.id))" -ForegroundColor Cyan
        
        # Check for associated hostnames
        Write-ColorOutput "Checking for associated hostnames..." -ForegroundColor Yellow
        $hostnames = Get-HnsHostname -TemplateId $template.id
        
        if ($hostnames.Count -gt 0) {
            Write-ColorOutput "Found $($hostnames.total) associated hostnames" -ForegroundColor Yellow
            
            # Group hostnames by status
            $reservedHostnames = @($hostnames | Where-Object { $_.status -eq "reserved" })
            $committedHostnames = @($hostnames | Where-Object { $_.status -eq "committed" })
            
            Write-ColorOutput "Status summary:" -ForegroundColor Yellow
            Write-ColorOutput "  - Reserved: $($reservedHostnames.Count)" -ForegroundColor Yellow
            Write-ColorOutput "  - Committed: $($committedHostnames.Count)" -ForegroundColor Yellow
            
            # Process committed hostnames – release them
            if ($committedHostnames.Count -gt 0) {
                Write-ColorOutput "Releasing $($committedHostnames.Count) committed hostnames..." -ForegroundColor Yellow
                foreach ($hostname in $committedHostnames) {
                    if ($WhatIf) {
                        Write-ColorOutput "WhatIf: Would release hostname $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Gray
                    } else {
                        try {
                            Write-ColorOutput "  Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                            Write-ColorOutput "  ✓ Released" -ForegroundColor Green
                        } catch {
                            Write-ColorOutput "  ✗ Failed: $_" -ForegroundColor Red
                        }
                        Start-Sleep -Milliseconds 100
                    }
                }
            }
            
            # Process reserved hostnames – commit then release
            if ($reservedHostnames.Count -gt 0) {
                Write-ColorOutput "Processing $($reservedHostnames.Count) reserved hostnames..." -ForegroundColor Yellow
                foreach ($hostname in $reservedHostnames) {
                    if ($WhatIf) {
                        Write-ColorOutput "WhatIf: Would commit and release hostname $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Gray
                    } else {
                        try {
                            Write-ColorOutput "  Committing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false
                            Write-ColorOutput "  ✓ Committed" -ForegroundColor Green
                            
                            Write-ColorOutput "  Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                            Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                            Write-ColorOutput "  ✓ Released" -ForegroundColor Green
                        } catch {
                            Write-ColorOutput "  ✗ Failed: $_" -ForegroundColor Red
                        }
                        Start-Sleep -Milliseconds 100
                    }
                }
            }
            
            # Verify that all hostnames have been processed
            if (-not $WhatIf) {
                Write-ColorOutput "Verifying all hostnames are processed..." -ForegroundColor Yellow
                $remainingHostnames = Get-HnsHostname -TemplateId $template.id
                if ($remainingHostnames.Count -gt 0) {
                    Write-ColorOutput "There are still $($remainingHostnames.Count) hostnames associated with this template." -ForegroundColor Yellow
                    Write-ColorOutput "These might be in 'released' status and require manual cleanup." -ForegroundColor Yellow
                } else {
                    Write-ColorOutput "All hostnames successfully processed." -ForegroundColor Green
                }
            }
        } else {
            Write-ColorOutput "No associated hostnames found." -ForegroundColor Green
        }
        
        # Delete the template
        if ($WhatIf) {
            Write-ColorOutput "WhatIf: Would delete template $($template.name) (ID: $($template.id))" -ForegroundColor Gray
        } else {
            Write-ColorOutput "Deleting template $($template.name) (ID: $($template.id))..." -ForegroundColor Yellow
            try {
                Remove-HnsTemplate -Id $template.id -Confirm:$false
                Write-ColorOutput "✓ Template deleted successfully" -ForegroundColor Green
            } catch {
                Write-ColorOutput "✗ Failed to delete template: $_" -ForegroundColor Red
                
                if ($_.Exception.Message -like "*foreign key constraint*") {
                    Write-ColorOutput "This appears to be a dependency issue. Some hostnames may still be linked to this template." -ForegroundColor Yellow
                    Write-ColorOutput "Try running this script again with -Force to attempt a more aggressive cleanup." -ForegroundColor Yellow
                }
            }
        }
    }
}  # End of Cleanup-Templates

function Cleanup-Hostnames {
    param(
        [Parameter(Mandatory = $false)]
        [string]$NamePattern = "",
        [Parameter(Mandatory = $false)]
        [int]$TemplateId = 0
    )
    $filterMsg = "Looking for hostnames"
    $hostnameParams = @{}
    
    if ($NamePattern) {
        $filterMsg += " with name pattern: $NamePattern"
        $hostnameParams["Name"] = $NamePattern
    }
    
    if ($TemplateId -gt 0) {
        $filterMsg += " with template ID: $TemplateId"
        $hostnameParams["TemplateId"] = $TemplateId
    }
    
    Write-ColorOutput $filterMsg -ForegroundColor Cyan
    
    try {
        $hostnames = Get-HnsHostname @hostnameParams
        
        if ($hostnames.Count -eq 0) {
            Write-ColorOutput "No matching hostnames found." -ForegroundColor Yellow
            return
        }
        
        Write-ColorOutput "Found $($hostnames.total) matching hostnames." -ForegroundColor Green
        
        $committedCount = @($hostnames | Where-Object { $_.status -eq "committed" }).Count
        $reservedCount = @($hostnames | Where-Object { $_.status -eq "reserved" }).Count
        $releasedCount = @($hostnames | Where-Object { $_.status -eq "released" }).Count
        
        Write-ColorOutput "Status summary:" -ForegroundColor Yellow
        Write-ColorOutput "  - Reserved: $reservedCount" -ForegroundColor Yellow
        Write-ColorOutput "  - Committed: $committedCount" -ForegroundColor Yellow
        Write-ColorOutput "  - Released: $releasedCount" -ForegroundColor Yellow
        
        Write-ColorOutput "`nSample of hostnames:" -ForegroundColor Cyan
        $sampleSize = [Math]::Min(5, $hostnames.Count)
        for ($i = 0; $i -lt $sampleSize; $i++) {
            Write-ColorOutput "  - $($hostnames[$i].name) (ID: $($hostnames[$i].id), Status: $($hostnames[$i].status))" -ForegroundColor Gray
        }
        
        if (-not $Force) {
            $confirmation = Read-Host "Do you want to process these hostnames? (y/n)"
            if ($confirmation -ne 'y') {
                Write-ColorOutput "Hostname cleanup cancelled." -ForegroundColor Yellow
                return
            }
        }
        
        $processed = 0
        $failed = 0
        
        $committedHostnames = @($hostnames | Where-Object { $_.status -eq "committed" })
        if ($committedHostnames.Count -gt 0) {
            Write-ColorOutput "`nProcessing $($committedHostnames.Count) committed hostnames..." -ForegroundColor Yellow
            foreach ($hostname in $committedHostnames) {
                if ($WhatIf) {
                    Write-ColorOutput "WhatIf: Would release hostname $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Gray
                } else {
                    try {
                        Write-ColorOutput "Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                        Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                        Write-ColorOutput "✓ Released" -ForegroundColor Green
                        $processed++
                    } catch {
                        Write-ColorOutput "✗ Failed: $_" -ForegroundColor Red
                        $failed++
                    }
                    Start-Sleep -Milliseconds 100
                }
            }
        }
        
        $reservedHostnames = @($hostnames | Where-Object { $_.status -eq "reserved" })
        if ($reservedHostnames.Count -gt 0) {
            Write-ColorOutput "`nProcessing $($reservedHostnames.Count) reserved hostnames..." -ForegroundColor Yellow
            foreach ($hostname in $reservedHostnames) {
                if ($WhatIf) {
                    Write-ColorOutput "WhatIf: Would commit and release hostname $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Gray
                } else {
                    try {
                        Write-ColorOutput "Committing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                        Set-HnsHostnameCommit -HostnameId $hostname.id -Confirm:$false
                        
                        Write-ColorOutput "Releasing hostname $($hostname.name) (ID: $($hostname.id))..." -ForegroundColor DarkGray
                        Set-HnsHostnameRelease -HostnameId $hostname.id -Confirm:$false
                        Write-ColorOutput "✓ Committed and released" -ForegroundColor Green
                        $processed++
                    } catch {
                        Write-ColorOutput "✗ Failed: $_" -ForegroundColor Red
                        $failed++
                    }
                    Start-Sleep -Milliseconds 100
                }
            }
        }
        
        if (-not $WhatIf) {
            Write-ColorOutput "`nHostname processing complete:" -ForegroundColor Cyan
            Write-ColorOutput "  - Successfully processed: $processed" -ForegroundColor Green
            if ($failed -gt 0) {
                Write-ColorOutput "  - Failed: $failed" -ForegroundColor Red
            }
        }
    } catch {
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
        $apiKeys = Get-HnsApiKeys
        $matchingKeys = @($apiKeys | Where-Object { $_.name -like "*$NamePattern*" })
        
        if ($matchingKeys.Count -eq 0) {
            Write-ColorOutput "No matching API keys found." -ForegroundColor Yellow
            return
        }
        
        Write-ColorOutput "Found $($matchingKeys.Count) matching API keys:" -ForegroundColor Green
        foreach ($key in $matchingKeys) {
            Write-Output "  - ID: $($key.id), Name: $($key.name), Scope: $($key.scope)"
        }
        
        if (-not $Force) {
            $confirmation = Read-Host "Do you want to delete these API keys? (y/n)"
            if ($confirmation -ne 'y') {
                Write-ColorOutput "API key deletion cancelled." -ForegroundColor Yellow
                return
            }
        }
        
        $processed = 0
        $failed = 0
        
        foreach ($key in $matchingKeys) {
            if ($WhatIf) {
                Write-ColorOutput "WhatIf: Would delete API key $($key.name) (ID: $($key.id))" -ForegroundColor Gray
            } else {
                try {
                    Write-ColorOutput "Deleting API key $($key.name) (ID: $($key.id))..." -ForegroundColor Yellow
                    Remove-HnsApiKey -Id $key.id -Confirm:$false
                    Write-ColorOutput "✓ API key deleted successfully" -ForegroundColor Green
                    $processed++
                } catch {
                    Write-ColorOutput "✗ Failed to delete API key: $_" -ForegroundColor Red
                    $failed++
                }
            }
        }
        
        if (-not $WhatIf) {
            Write-ColorOutput "`nAPI key cleanup complete:" -ForegroundColor Cyan
            Write-ColorOutput "  - Successfully deleted: $processed" -ForegroundColor Green
            if ($failed -gt 0) {
                Write-ColorOutput "  - Failed: $failed" -ForegroundColor Red
            }
        }
    } catch {
        Write-ColorOutput "Error getting API keys: $_" -ForegroundColor Red
    }
}

# Main cleanup script
try {
    Write-ColorOutput "Initializing connection to HNS server at $HnsUrl..." -ForegroundColor Cyan
    Initialize-HnsConnection -BaseUrl $HnsUrl
    
    Write-ColorOutput "Authenticating with username $Username..." -ForegroundColor Cyan
    Connect-HnsService -Username $Username -Password $Password
    
    $connected = Test-HnsConnection
    if (-not $connected) {
        throw "Failed to connect to HNS server."
    }
    
    Write-ColorOutput "`n===== HNS Cleanup Tool =====" -ForegroundColor Magenta
    
    if ($WhatIf) {
        Write-ColorOutput "Running in WhatIf mode - no changes will be made" -ForegroundColor Cyan
    }
    
    if ($Force) {
        Write-ColorOutput "Force mode enabled - no confirmations will be requested" -ForegroundColor Yellow
    }
    
    if ($CleanupType -eq "All" -or $CleanupType -eq "Hostnames") {
        Write-ColorOutput "`n[Hostname Cleanup]" -ForegroundColor Magenta
        Cleanup-Hostnames -NamePattern $NamePattern -TemplateId $TemplateId
    }
    
    if ($CleanupType -eq "All" -or $CleanupType -eq "Templates") {
        Write-ColorOutput "`n[Template Cleanup]" -ForegroundColor Magenta
        Cleanup-Templates -NamePattern $NamePattern -TemplateId $TemplateId
    }
    
    if ($CleanupType -eq "All" -or $CleanupType -eq "ApiKeys") {
        Write-ColorOutput "`n[API Key Cleanup]" -ForegroundColor Magenta
        Cleanup-ApiKeys -NamePattern $NamePattern
    }
    
    Write-ColorOutput "`nCleanup completed!" -ForegroundColor Green
} catch {
    Write-ColorOutput "An error occurred during cleanup: $_" -ForegroundColor Red
    Write-ColorOutput $_.ScriptStackTrace -ForegroundColor Red
}
