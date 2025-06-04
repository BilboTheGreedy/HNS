function Get-AvailableHostname {
    <#
    .SYNOPSIS
        Generates the next available hostname following Epiroc VM Standard
    
    .DESCRIPTION
        Generates hostnames following the Epiroc 15-character standard and checks availability
        across multiple systems (DNS, ICMP, Active Directory, ServiceNow CMDB).
        Auto-increments sequence number until an available hostname is found.
    
    .PARAMETER UnitCode
        3-character business unit code (e.g., 'SFD', 'USS')
    
    .PARAMETER Provider
        Cloud/infrastructure provider: E (Epiroc), M (Azure), A (AWS), G (Google)
    
    .PARAMETER Region
        4-character region code (e.g., 'WEEU', 'SCUS', 'EUSE')
    
    .PARAMETER Environment
        Environment: P (Production), T (Test), D (Development)
    
    .PARAMETER Function
        2-character function code (e.g., 'AS', 'DB', 'WS', 'BU')
    
    .PARAMETER StartSequence
        Starting sequence number (default: 1)
    
    .PARAMETER MaxSequence
        Maximum sequence number to try (default: 999)
    
    .PARAMETER CheckDNS
        Check DNS resolution (default: $true)
    
    .PARAMETER CheckICMP
        Check ICMP ping (default: $true)
    
    .PARAMETER CheckActiveDirectory
        Check Active Directory computer objects (default: $true)
    
    .PARAMETER CheckServiceNow
        Check ServiceNow CMDB (default: $false, requires configuration)
    
    .PARAMETER ServiceNowInstance
        ServiceNow instance URL (required if CheckServiceNow is true)
    
    .PARAMETER ServiceNowCredential
        ServiceNow credentials (required if CheckServiceNow is true)
    
    .EXAMPLE
        Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS'
        Returns: SFDSMWEEUPAS001 (if available) or next available sequence
    
    .EXAMPLE
        Get-AvailableHostname -UnitCode 'USS' -Provider 'M' -Region 'SCUS' -Environment 'T' -Function 'DB' -StartSequence 5
        Returns: USSSMSCUSTDB005 or higher available sequence
    
    .EXAMPLE
        Get-AvailableHostname -UnitCode 'SFD' -Provider 'E' -Region 'EUSE' -Environment 'P' -Function 'BU' -CheckServiceNow -ServiceNowInstance 'https://company.servicenow.com'
        Returns: Available hostname after checking all systems including ServiceNow
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateSet('SFD','SFS','DAL','COP','JAC','RDT','AVT','RIG','FRD','EST','FAG','CER','ALL','HPI','NJB','OUL','ARE','CHR','CLK','FRB','GAR','GHA','KAL','LAN','NAC','OEB','SKE','UDD','ZED','GDE','JHA','NAN','ROS','ITS','LEG','CHI','BRA','USA','PER','MEX','CAN','AUS','IND','AFR','RUS','MON','KAZ','EUR','GLO','MAL','THA','KOR','JAP','NIG','GEO','FIN')]
        [string]$UnitCode,
        
        [Parameter(Mandatory = $true)]
        [ValidateSet('E','M','A','G')]
        [string]$Provider,
        
        [Parameter(Mandatory = $true)]
        [ValidateSet('OCAU','ASCN','ASIN','ASSG','EUSE','AFZA','NAUS','ACL1','ACL2','AUEA','SEAU','INCE','SHA1','BJB1','EAAS','JPEA','JPWE','KRCE','KRSO','INSO','SEAS','INWE','FRCE','DEWC','ISCE','ITNO','NOEU','NWEA','PLCE','QACE','SANO','SAWE','SWCE','SZNO','UANO','UKSO','UKWE','WEEU','BRSO','CCAN','ECAN','CEUS','CLCE','EUS1','EUS2','NCUS','SCUS','WCUS','WUS1','WUS2','WUS3')]
        [string]$Region,
        
        [Parameter(Mandatory = $true)]
        [ValidateSet('P','T','D')]
        [string]$Environment,
        
        [Parameter(Mandatory = $true)]
        [ValidateSet('AS','BU','DB','DC','FS','IS','MG','PO','WS','PV','CC','UC','OT')]
        [string]$Function,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, 999)]
        [int]$StartSequence = 1,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, 999)]
        [int]$MaxSequence = 999,
        
        [Parameter(Mandatory = $false)]
        [bool]$CheckDNS = $true,
        
        [Parameter(Mandatory = $false)]
        [bool]$CheckICMP = $true,
        
        [Parameter(Mandatory = $false)]
        [bool]$CheckActiveDirectory = $true,
        
        [Parameter(Mandatory = $false)]
        [bool]$CheckServiceNow = $false,
        
        [Parameter(Mandatory = $false)]
        [string]$ServiceNowInstance,
        
        [Parameter(Mandatory = $false)]
        [System.Management.Automation.PSCredential]$ServiceNowCredential
    )
    
    begin {
        Write-Verbose "Starting hostname availability check for Epiroc VM Standard"
        Write-Verbose "Parameters: Unit=$UnitCode, Provider=$Provider, Region=$Region, Env=$Environment, Function=$Function"
        
        # Validate ServiceNow parameters
        if ($CheckServiceNow) {
            if (-not $ServiceNowInstance -or -not $ServiceNowCredential) {
                throw "ServiceNow instance URL and credentials are required when CheckServiceNow is enabled"
            }
        }
        
        # Initialize results tracking
        $checkResults = @{
            DNS = @()
            ICMP = @()
            ActiveDirectory = @()
            ServiceNow = @()
            Checked = 0
            Found = $null
        }
    }
    
    process {
        for ($sequence = $StartSequence; $sequence -le $MaxSequence; $sequence++) {
            # Generate hostname following Epiroc standard
            $sequenceStr = $sequence.ToString("000")
            $hostname = "$UnitCode" + "S" + "$Provider" + "$Region" + "$Environment" + "$Function" + "$sequenceStr"
            
            Write-Verbose "Testing hostname: $hostname (sequence $sequence)"
            $checkResults.Checked++
            
            # Check availability across all requested systems
            $isAvailable = $true
            $conflicts = @()
            
            # DNS Check
            if ($CheckDNS -and $isAvailable) {
                Write-Verbose "  Checking DNS for $hostname"
                try {
                    $dnsResult = Resolve-DnsName -Name $hostname -ErrorAction SilentlyContinue
                    if ($dnsResult) {
                        $conflicts += "DNS"
                        $isAvailable = $false
                        $checkResults.DNS += @{Hostname = $hostname; Status = "Conflict"; Details = $dnsResult.IPAddress -join ","}
                    } else {
                        $checkResults.DNS += @{Hostname = $hostname; Status = "Available"}
                    }
                } catch {
                    # DNS resolution failed = hostname is available in DNS
                    $checkResults.DNS += @{Hostname = $hostname; Status = "Available"}
                }
            }
            
            # ICMP Check
            if ($CheckICMP -and $isAvailable) {
                Write-Verbose "  Checking ICMP ping for $hostname"
                try {
                    $pingResult = Test-Connection -ComputerName $hostname -Count 1 -Quiet -ErrorAction SilentlyContinue
                    if ($pingResult) {
                        $conflicts += "ICMP"
                        $isAvailable = $false
                        $checkResults.ICMP += @{Hostname = $hostname; Status = "Conflict"; Details = "Responds to ping"}
                    } else {
                        $checkResults.ICMP += @{Hostname = $hostname; Status = "Available"}
                    }
                } catch {
                    # Ping failed = hostname is available
                    $checkResults.ICMP += @{Hostname = $hostname; Status = "Available"}
                }
            }
            
            # Active Directory Check
            if ($CheckActiveDirectory -and $isAvailable) {
                Write-Verbose "  Checking Active Directory for $hostname"
                try {
                    $adComputer = Get-ADComputer -Identity $hostname -ErrorAction SilentlyContinue
                    if ($adComputer) {
                        $conflicts += "ActiveDirectory"
                        $isAvailable = $false
                        $checkResults.ActiveDirectory += @{Hostname = $hostname; Status = "Conflict"; Details = "Computer object exists"}
                    } else {
                        $checkResults.ActiveDirectory += @{Hostname = $hostname; Status = "Available"}
                    }
                } catch {
                    # Computer not found = available
                    $checkResults.ActiveDirectory += @{Hostname = $hostname; Status = "Available"}
                }
            }
            
            # ServiceNow CMDB Check
            if ($CheckServiceNow -and $isAvailable) {
                Write-Verbose "  Checking ServiceNow CMDB for $hostname"
                try {
                    $servicenowResult = Test-ServiceNowCMDB -Hostname $hostname -Instance $ServiceNowInstance -Credential $ServiceNowCredential
                    if ($servicenowResult.Exists) {
                        $conflicts += "ServiceNow"
                        $isAvailable = $false
                        $checkResults.ServiceNow += @{Hostname = $hostname; Status = "Conflict"; Details = $servicenowResult.Details}
                    } else {
                        $checkResults.ServiceNow += @{Hostname = $hostname; Status = "Available"}
                    }
                } catch {
                    Write-Warning "ServiceNow check failed for $hostname : $_"
                    # Continue assuming available if ServiceNow check fails
                    $checkResults.ServiceNow += @{Hostname = $hostname; Status = "Available"; Details = "Check failed"}
                }
            }
            
            # If hostname is available across all systems, return it
            if ($isAvailable) {
                Write-Verbose "✓ Found available hostname: $hostname"
                $checkResults.Found = $hostname
                
                $result = [PSCustomObject]@{
                    Hostname = $hostname
                    UnitCode = $UnitCode
                    Type = "S"
                    Provider = $Provider
                    Region = $Region
                    Environment = $Environment
                    Function = $Function
                    Sequence = $sequence
                    Length = $hostname.Length
                    IsValid = ($hostname.Length -eq 15)
                    ChecksPerformed = @{
                        DNS = $CheckDNS
                        ICMP = $CheckICMP
                        ActiveDirectory = $CheckActiveDirectory
                        ServiceNow = $CheckServiceNow
                    }
                    SequencesTested = $sequence - $StartSequence + 1
                    Summary = $checkResults
                }
                
                return $result
            } else {
                Write-Verbose "✗ Hostname $hostname is not available (conflicts: $($conflicts -join ', '))"
            }
        }
        
        # If we get here, no available hostname was found
        throw "No available hostname found after checking sequences $StartSequence to $MaxSequence. All hostnames are in use."
    }
}

function Test-ServiceNowCMDB {
    <#
    .SYNOPSIS
        Check ServiceNow CMDB for hostname existence
    #>
    [CmdletBinding()]
    param(
        [string]$Hostname,
        [string]$Instance,
        [System.Management.Automation.PSCredential]$Credential
    )
    
    try {
        # Build ServiceNow API URL for CMDB CI lookup
        $uri = "$Instance/api/now/table/cmdb_ci_computer"
        $query = "?sysparm_query=name=$Hostname"
        $fullUri = $uri + $query
        
        # Prepare headers
        $headers = @{
            'Accept' = 'application/json'
            'Content-Type' = 'application/json'
        }
        
        # Make API call
        $response = Invoke-RestMethod -Uri $fullUri -Method Get -Credential $Credential -Headers $headers
        
        if ($response.result -and $response.result.Count -gt 0) {
            return @{
                Exists = $true
                Details = "Found $($response.result.Count) CMDB record(s)"
                Records = $response.result
            }
        } else {
            return @{
                Exists = $false
                Details = "No CMDB records found"
                Records = @()
            }
        }
    } catch {
        Write-Warning "ServiceNow API call failed: $_"
        return @{
            Exists = $false
            Details = "API call failed: $($_.Exception.Message)"
            Records = @()
        }
    }
}