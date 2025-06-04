function Get-NextAvailableHostname {
    <#
    .SYNOPSIS
        Generates the next available hostname by incrementing sequence number
    .DESCRIPTION
        Takes a template and generates hostnames, incrementing the sequence number 
        until an available hostname is found across all configured systems
    .PARAMETER Template
        The hostname template to use
    .PARAMETER StartingSequence
        Starting sequence number (default: 1)
    .PARAMETER MaxAttempts
        Maximum number of attempts (default: 1000)
    .PARAMETER AvailabilityChecks
        Hashtable of availability check settings
    .EXAMPLE
        Get-NextAvailableHostname -Template $template -StartingSequence 1
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [Template]$Template,
        
        [Parameter()]
        [int]$StartingSequence = 1,
        
        [Parameter()]
        [int]$MaxAttempts = 1000,
        
        [Parameter()]
        [hashtable]$AvailabilityChecks = @{
            CheckDNS = $true
            CheckICMP = $true
            CheckActiveDirectory = $true
            CheckServiceNow = $false
            ServiceNowConfig = @{}
        }
    )
    
    Write-Verbose "Finding next available hostname for template: $($Template.Name)"
    Write-Verbose "Starting sequence: $StartingSequence, Max attempts: $MaxAttempts"
    
    $currentSequence = $StartingSequence
    $attempts = 0
    
    while ($attempts -lt $MaxAttempts) {
        $attempts++
        
        # Generate hostname with current sequence
        try {
            $hostname = $Template.GenerateHostname($currentSequence)
            Write-Verbose "Attempt $attempts`: Testing hostname: $hostname"
            
            # Test availability
            $availabilityResult = Test-HostnameAvailability -Hostname $hostname @AvailabilityChecks
            
            if ($availabilityResult.Available) {
                Write-Host "Found available hostname: $hostname (sequence: $currentSequence)" -ForegroundColor Green
                
                return @{
                    Hostname = $hostname
                    Sequence = $currentSequence
                    Template = $Template
                    AvailabilityChecks = $availabilityResult.Checks
                    Attempts = $attempts
                }
            } else {
                Write-Verbose "Hostname $hostname is not available:"
                foreach ($check in $availabilityResult.Checks.GetEnumerator()) {
                    if ($check.Value.Enabled -and -not $check.Value.Available) {
                        Write-Verbose "  $($check.Key): $($check.Value.Details)"
                    }
                }
                
                $currentSequence++
            }
        }
        catch {
            Write-Warning "Error generating/testing hostname at sequence $currentSequence`: $($_.Exception.Message)"
            $currentSequence++
        }
    }
    
    throw "Could not find available hostname after $MaxAttempts attempts. Last sequence tried: $($currentSequence - 1)"
}

function Get-HighestUsedSequence {
    <#
    .SYNOPSIS
        Finds the highest used sequence number for a template by scanning existing hostnames
    .PARAMETER Template
        The template to check
    .PARAMETER ScanRange
        Number of sequence numbers to scan (default: 100)
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [Template]$Template,
        
        [Parameter()]
        [int]$ScanRange = 100
    )
    
    Write-Verbose "Scanning for highest used sequence for template: $($Template.Name)"
    
    $highestFound = 0
    
    for ($seq = 1; $seq -le $ScanRange; $seq++) {
        try {
            $hostname = $Template.GenerateHostname($seq)
            
            # Quick check - just DNS and ping
            $available = Test-HostnameAvailability -Hostname $hostname -CheckDNS -CheckICMP -CheckActiveDirectory:$false -CheckServiceNow:$false
            
            if (-not $available.Available) {
                $highestFound = $seq
                Write-Verbose "Found used hostname: $hostname (sequence: $seq)"
            }
        }
        catch {
            Write-Verbose "Error checking sequence $seq`: $($_.Exception.Message)"
        }
    }
    
    Write-Verbose "Highest used sequence found: $highestFound"
    return $highestFound
}