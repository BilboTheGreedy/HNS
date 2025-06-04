function New-HNSHostname {
    <#
    .SYNOPSIS
        Generates the next available hostname using the specified template
    .DESCRIPTION
        Creates a new hostname based on a template with automatic sequence numbering
        and availability checking across DNS, ICMP, Active Directory, and ServiceNow
    .PARAMETER TemplateID
        The ID of the template to use
    .PARAMETER TemplateName
        The name of the template to use (alternative to TemplateID)
    .PARAMETER Parameters
        Hashtable of parameter values for template groups that require input
    .PARAMETER StartingSequence
        Starting sequence number for availability search
    .PARAMETER ReservedBy
        Username who is reserving this hostname
    .PARAMETER Notes
        Optional notes for the hostname
    .PARAMETER MaxAttempts
        Maximum number of hostnames to try (default: 1000)
    .PARAMETER CheckDNS
        Check DNS resolution (default: true)
    .PARAMETER CheckICMP
        Check ICMP ping response (default: true)  
    .PARAMETER CheckActiveDirectory
        Check Active Directory computer objects (default: true)
    .PARAMETER CheckServiceNow
        Check ServiceNow CMDB (default: false)
    .PARAMETER ServiceNowConfig
        ServiceNow connection configuration
    .EXAMPLE
        New-HNSHostname -TemplateName "Epiroc VM Standard" -Parameters @{unit_code='DAL'; provider='A'; region='EAUS'; environment='D'; function='DB'} -ReservedBy "john.doe"
    .EXAMPLE
        New-HNSHostname -TemplateID 1 -StartingSequence 42 -ReservedBy "admin" -CheckServiceNow -ServiceNowConfig @{Instance='dev12345'; Username='user'; Password='pass'}
    #>
    [CmdletBinding(DefaultParameterSetName = 'ByID')]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByID')]
        [int]$TemplateID,
        
        [Parameter(Mandatory, ParameterSetName = 'ByName')]
        [string]$TemplateName,
        
        [Parameter()]
        [hashtable]$Parameters = @{},
        
        [Parameter()]
        [int]$StartingSequence,
        
        [Parameter(Mandatory = $true)]
        [string]$ReservedBy,
        
        [Parameter()]
        [string]$Notes = "",
        
        [Parameter()]
        [int]$MaxAttempts = 1000,
        
        [Parameter()]
        [switch]$CheckDNS = $true,
        
        [Parameter()]
        [switch]$CheckICMP = $true,
        
        [Parameter()]
        [switch]$CheckActiveDirectory = $true,
        
        [Parameter()]
        [switch]$CheckServiceNow = $false,
        
        [Parameter()]
        [hashtable]$ServiceNowConfig = @{}
    )
    
    begin {
        Test-HNSInitialized -Throw
        Write-Host "Generating next available hostname..." -ForegroundColor Cyan
    }
    
    process {
        try {
            # Get template
            $template = if ($PSCmdlet.ParameterSetName -eq 'ByID') {
                $foundTemplates = Get-HNSTemplate -ID $TemplateID
                $foundTemplates | Where-Object { $_ -is [Template] } | Select-Object -First 1
            } else {
                $foundTemplates = Get-HNSTemplate -Name $TemplateName
                $foundTemplates | Where-Object { $_ -is [Template] } | Select-Object -First 1
            }
            
            if (-not $template) {
                throw "Template not found"
            }
            
            if (-not $template.IsActive) {
                throw "Template '$($template.Name)' is not active"
            }
            
            Write-Host "Using template: $($template.Name)" -ForegroundColor Gray
            
            # Determine starting sequence
            if (-not $StartingSequence) {
                Write-Host "Scanning for highest used sequence..." -ForegroundColor Yellow
                $highestUsed = Get-HighestUsedSequence -Template $template
                $StartingSequence = $highestUsed + 1
                Write-Host "Starting sequence search at: $StartingSequence" -ForegroundColor Gray
            }
            
            # Prepare availability check parameters
            $availabilityChecks = @{
                CheckDNS = $CheckDNS
                CheckICMP = $CheckICMP  
                CheckActiveDirectory = $CheckActiveDirectory
                CheckServiceNow = $CheckServiceNow
                ServiceNowConfig = $ServiceNowConfig
            }
            
            Write-Host "Searching for available hostname (max $MaxAttempts attempts)..." -ForegroundColor Yellow
            
            $currentSequence = $StartingSequence
            $attempts = 0
            
            while ($attempts -lt $MaxAttempts) {
                $attempts++
                
                try {
                    # Build hostname parts
                    $hostnameParts = @{}
                    
                    foreach ($group in $template.Groups | Sort-Object Position) {
                        $value = ""
                        
                        switch ($group.ValidationType) {
                            ([ValidationType]::Fixed) {
                                $value = $group.ValidationValue
                            }
                            ([ValidationType]::Sequence) {
                                if ($group.Position -eq $template.SequencePosition) {
                                    $value = $currentSequence.ToString()
                                    if ($template.SequencePadding) {
                                        $value = $value.PadLeft($template.SequenceLength, '0')
                                    }
                                }
                            }
                            ([ValidationType]::List) {
                                if ($Parameters.ContainsKey($group.Name)) {
                                    $value = $Parameters[$group.Name]
                                } elseif ($group.IsRequired) {
                                    throw "Required parameter '$($group.Name)' not provided. Valid values: $($group.ValidationValue)"
                                }
                            }
                            ([ValidationType]::Regex) {
                                if ($Parameters.ContainsKey($group.Name)) {
                                    $value = $Parameters[$group.Name]
                                } elseif ($group.IsRequired) {
                                    throw "Required parameter '$($group.Name)' not provided. Must match pattern: $($group.ValidationValue)"
                                }
                            }
                        }
                        
                        # Validate value
                        if ($value -and -not $group.Validate($value)) {
                            throw "Value '$value' is not valid for group '$($group.Name)'. $($group.ToString())"
                        }
                        
                        $hostnameParts[$group.Position] = $value
                    }
                    
                    # Assemble hostname
                    $hostname = ""
                    for ($i = 1; $i -le $template.Groups.Count; $i++) {
                        if ($hostnameParts.ContainsKey($i)) {
                            $hostname += $hostnameParts[$i]
                        }
                    }
                    
                    Write-Verbose "Attempt $attempts`: Testing hostname: $hostname"
                    
                    # Test availability
                    $availabilityResult = Test-HostnameAvailability -Hostname $hostname @availabilityChecks
                    
                    if ($availabilityResult.Available) {
                        Write-Host "`nFound available hostname: $hostname (sequence: $currentSequence)" -ForegroundColor Green
                        
                        # Display availability check results
                        Write-Host "`nAvailability Check Results:" -ForegroundColor Cyan
                        foreach ($check in $availabilityResult.Checks.GetEnumerator()) {
                            if ($check.Value.Enabled) {
                                $status = if ($check.Value.Available) { "AVAILABLE" } else { "IN USE" }
                                $color = if ($check.Value.Available) { "Green" } else { "Red" }
                                Write-Host "  $($check.Key): $status - $($check.Value.Details)" -ForegroundColor $color
                            }
                        }
                        
                        # Create hostname object for tracking
                        $hostnameRecord = [Hostname]::new(@{
                            ID = if ($script:Configuration) { $script:Configuration.GetNextHostnameID() } else { 1 }
                            Name = $hostname
                            TemplateID = $template.ID
                            TemplateName = $template.Name
                            SequenceNum = $currentSequence
                            Notes = $Notes
                            Metadata = @{
                                Parameters = ($Parameters.GetEnumerator() | ForEach-Object { "$($_.Key)=$($_.Value)" }) -join '; '
                                GeneratedBy = $ReservedBy
                                AvailabilityChecks = $availabilityResult.Checks
                            }
                        })
                        
                        # Reserve it
                        $hostnameRecord.Reserve($ReservedBy)
                        
                        Write-Host "`nHostname '$hostname' generated and reserved successfully!" -ForegroundColor Green
                        Write-Host "Template: $($template.Name)" -ForegroundColor Gray
                        Write-Host "Sequence: $currentSequence" -ForegroundColor Gray
                        Write-Host "Reserved by: $ReservedBy" -ForegroundColor Gray
                        Write-Host "Attempts required: $attempts" -ForegroundColor Gray
                        
                        return $hostnameRecord
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
        catch {
            Write-Error "Failed to generate available hostname: $($_.Exception.Message)"
            throw
        }
    }
}