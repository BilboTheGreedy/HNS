function Test-HostnameAvailability {
    <#
    .SYNOPSIS
        Tests if a hostname is available across multiple systems
    .DESCRIPTION
        Checks hostname availability against DNS, ICMP, Active Directory, and ServiceNow CMDB
    .PARAMETER Hostname
        The hostname to test
    .PARAMETER CheckDNS
        Check DNS resolution
    .PARAMETER CheckICMP
        Check ICMP ping response
    .PARAMETER CheckActiveDirectory
        Check Active Directory computer objects
    .PARAMETER CheckServiceNow
        Check ServiceNow CMDB
    .PARAMETER ServiceNowConfig
        ServiceNow connection configuration
    .EXAMPLE
        Test-HostnameAvailability -Hostname "server001" -CheckDNS -CheckICMP -CheckActiveDirectory
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Hostname,
        
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
    
    $results = @{
        Hostname = $Hostname
        Available = $true
        Checks = @{
            DNS = @{ Enabled = $CheckDNS; Available = $true; Details = "" }
            ICMP = @{ Enabled = $CheckICMP; Available = $true; Details = "" }
            ActiveDirectory = @{ Enabled = $CheckActiveDirectory; Available = $true; Details = "" }
            ServiceNow = @{ Enabled = $CheckServiceNow; Available = $true; Details = "" }
        }
    }
    
    # DNS Check
    if ($CheckDNS) {
        try {
            Write-Verbose "Checking DNS for $Hostname"
            $dnsResult = Resolve-DnsName -Name $Hostname -ErrorAction Stop
            $results.Checks.DNS.Available = $false
            $results.Checks.DNS.Details = "DNS record exists: $($dnsResult.IPAddress -join ', ')"
            $results.Available = $false
        }
        catch {
            $results.Checks.DNS.Available = $true
            $results.Checks.DNS.Details = "No DNS record found"
        }
    }
    
    # ICMP Check
    if ($CheckICMP) {
        try {
            Write-Verbose "Checking ICMP for $Hostname"
            $pingResult = Test-Connection -ComputerName $Hostname -Count 1 -Quiet -TimeoutSeconds 2
            if ($pingResult) {
                $results.Checks.ICMP.Available = $false
                $results.Checks.ICMP.Details = "Host responds to ping"
                $results.Available = $false
            } else {
                $results.Checks.ICMP.Available = $true
                $results.Checks.ICMP.Details = "No ping response"
            }
        }
        catch {
            $results.Checks.ICMP.Available = $true
            $results.Checks.ICMP.Details = "No ping response (error: $($_.Exception.Message))"
        }
    }
    
    # Active Directory Check
    if ($CheckActiveDirectory) {
        try {
            Write-Verbose "Checking Active Directory for $Hostname"
            Import-Module ActiveDirectory -ErrorAction Stop
            $adComputer = Get-ADComputer -Filter "Name -eq '$Hostname'" -ErrorAction Stop
            if ($adComputer) {
                $results.Checks.ActiveDirectory.Available = $false
                $results.Checks.ActiveDirectory.Details = "Computer object exists in AD: $($adComputer.DistinguishedName)"
                $results.Available = $false
            } else {
                $results.Checks.ActiveDirectory.Available = $true
                $results.Checks.ActiveDirectory.Details = "No computer object found in AD"
            }
        }
        catch [Microsoft.ActiveDirectory.Management.ADIdentityNotFoundException] {
            $results.Checks.ActiveDirectory.Available = $true
            $results.Checks.ActiveDirectory.Details = "No computer object found in AD"
        }
        catch {
            Write-Warning "Active Directory check failed: $($_.Exception.Message)"
            $results.Checks.ActiveDirectory.Details = "AD check failed: $($_.Exception.Message)"
        }
    }
    
    # ServiceNow Check
    if ($CheckServiceNow -and $ServiceNowConfig.Count -gt 0) {
        try {
            Write-Verbose "Checking ServiceNow CMDB for $Hostname"
            $snowResult = Invoke-ServiceNowCMDBCheck -Hostname $Hostname -Config $ServiceNowConfig
            if ($snowResult.Found) {
                $results.Checks.ServiceNow.Available = $false
                $results.Checks.ServiceNow.Details = "Found in ServiceNow CMDB: $($snowResult.SysId)"
                $results.Available = $false
            } else {
                $results.Checks.ServiceNow.Available = $true
                $results.Checks.ServiceNow.Details = "Not found in ServiceNow CMDB"
            }
        }
        catch {
            Write-Warning "ServiceNow check failed: $($_.Exception.Message)"
            $results.Checks.ServiceNow.Details = "ServiceNow check failed: $($_.Exception.Message)"
        }
    }
    
    return $results
}

function Invoke-ServiceNowCMDBCheck {
    <#
    .SYNOPSIS
        Checks ServiceNow CMDB for hostname existence
    .PARAMETER Hostname
        The hostname to search for
    .PARAMETER Config
        ServiceNow configuration (Instance, Username, Password/Token)
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Hostname,
        
        [Parameter(Mandatory = $true)]
        [hashtable]$Config
    )
    
    if (-not $Config.Instance -or -not $Config.Username -or (-not $Config.Password -and -not $Config.Token)) {
        throw "ServiceNow configuration incomplete. Need Instance, Username, and Password/Token"
    }
    
    $uri = "https://$($Config.Instance).service-now.com/api/now/table/cmdb_ci_computer"
    $query = "name=$Hostname"
    
    # Create headers
    $headers = @{
        'Accept' = 'application/json'
        'Content-Type' = 'application/json'
    }
    
    # Authentication
    if ($Config.Token) {
        $headers['Authorization'] = "Bearer $($Config.Token)"
    } else {
        $credentials = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($Config.Username):$($Config.Password)"))
        $headers['Authorization'] = "Basic $credentials"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$uri?sysparm_query=$query" -Headers $headers -Method Get
        
        return @{
            Found = $response.result.Count -gt 0
            SysId = if ($response.result.Count -gt 0) { $response.result[0].sys_id } else { $null }
            Details = $response.result
        }
    }
    catch {
        Write-Error "ServiceNow API call failed: $($_.Exception.Message)"
        throw
    }
}