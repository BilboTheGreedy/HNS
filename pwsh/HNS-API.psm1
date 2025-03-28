# HNS-API.psm1
# PowerShell module for interacting with the Hostname Naming System API

# Global variables
$script:BaseUrl = $null
$script:ApiKey = $null
$script:JwtToken = $null
$script:DefaultHeaders = @{}

#region Configuration Functions

function Initialize-HnsConnection {
    <#
    .SYNOPSIS
        Initializes the connection to the HNS API service.
    
    .DESCRIPTION
        Sets up the base URL and authentication headers for all subsequent API calls.
    
    .PARAMETER BaseUrl
        The base URL of the HNS API (e.g., "http://localhost:8080").
    
    .PARAMETER ApiKey
        Optional API key for authentication.
    
    .EXAMPLE
        Initialize-HnsConnection -BaseUrl "http://hostname-service.example.com:8080"
    
    .EXAMPLE
        Initialize-HnsConnection -BaseUrl "http://hostname-service.example.com:8080" -ApiKey "abc123"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$BaseUrl,
        
        [Parameter(Mandatory = $false)]
        [string]$ApiKey
    )

    # Set global base URL
    $script:BaseUrl = $BaseUrl

    # Set default headers
    $script:DefaultHeaders = @{
        "Content-Type" = "application/json"
        "Accept" = "application/json"
    }

    # Add API key if provided
    if ($ApiKey) {
        $script:ApiKey = $ApiKey
        $script:DefaultHeaders["X-API-Key"] = $ApiKey
    }

    Write-Host "HNS API connection initialized with base URL: $BaseUrl"
}

function Connect-HnsService {
    <#
    .SYNOPSIS
        Connects to the HNS service using username/password authentication.
    
    .DESCRIPTION
        Authenticates using the provided username and password and stores the JWT token.
    
    .PARAMETER Username
        The username for authentication.
    
    .PARAMETER Password
        The password for authentication.
    
    .EXAMPLE
        Connect-HnsService -Username "admin" -Password "password123"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Username,
        
        [Parameter(Mandatory = $true)]
        [ValidateNotNull()]
        [SecureString]$Password
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    # Convert secure string to plain text
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($Password)
    $plainPassword = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)

    $body = @{
        username = $Username
        password = $plainPassword
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/auth/login" -Method Post -Body $body -ContentType "application/json"
        $script:JwtToken = $response.token
        $script:DefaultHeaders["Authorization"] = "Bearer $($response.token)"
        
        Write-Host "Successfully authenticated as $Username"
        return $response
    }
    catch {
        $errorMessage = "Authentication failed`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
    finally {
        # Clear the plain text password from memory
        $plainPassword = $null
    }
}

function Test-HnsConnection {
    <#
    .SYNOPSIS
        Tests the connection to the HNS API service.
    
    .DESCRIPTION
        Checks if the API service is accessible by calling the health endpoint.
    
    .EXAMPLE
        Test-HnsConnection
    #>
    [CmdletBinding()]
    param()

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/health" -Method Get -Headers $script:DefaultHeaders
        Write-Host "Connection successful. Service status: $($response.status)"
        return $true
    }
    catch {
        $errorMessage = "Connection test failed`: $_"
        Write-Error $errorMessage
        return $false
    }
}

#endregion

#region Sequence Management Functions

function Get-HnsNextSequenceNumber {
    <#
    .SYNOPSIS
        Gets the next available sequence number for a template.
    
    .DESCRIPTION
        Retrieves the next available sequence number for hostname generation based on the template ID.
    
    .PARAMETER TemplateId
        The ID of the template.
    
    .EXAMPLE
        Get-HnsNextSequenceNumber -TemplateId 1
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$TemplateId
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/sequences/next/$TemplateId" -Method Get -Headers $script:DefaultHeaders
        return $response
    }
    catch {
        $errorMessage = "Failed to get next sequence number`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

#endregion

#region Template Management Functions

function Get-HnsTemplate {
    <#
    .SYNOPSIS
        Gets one or more hostname templates.
    
    .DESCRIPTION
        Retrieves either a specific template by ID or a list of templates with pagination.
    
    .PARAMETER Id
        The ID of a specific template to retrieve. If omitted, returns a list.
    
    .PARAMETER Limit
        Maximum number of templates to return when listing templates.
    
    .PARAMETER Offset
        Number of templates to skip when listing templates.
    
    .EXAMPLE
        Get-HnsTemplate -Id 1
    
    .EXAMPLE
        Get-HnsTemplate -Limit 10 -Offset 0
    #>
    [CmdletBinding(DefaultParameterSetName = 'List')]
    param(
        [Parameter(Mandatory = $true, ParameterSetName = 'Single')]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$Id,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateRange(1, 100)]
        [int]$Limit = 10,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateRange(0, [int]::MaxValue)]
        [int]$Offset = 0
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    try {
        if ($PSCmdlet.ParameterSetName -eq 'Single') {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/templates/$Id" -Method Get -Headers $script:DefaultHeaders
            return $response
        }
        else {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/templates?limit=$Limit&offset=$Offset" -Method Get -Headers $script:DefaultHeaders
            return $response.templates
        }
    }
    catch {
        if ($PSCmdlet.ParameterSetName -eq 'Single') {
            $errorContext = "Failed to get template $Id"
        } else {
            $errorContext = "Failed to get templates"
        }
        $errorMessage = "$errorContext`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function New-HnsTemplate {
    <#
    .SYNOPSIS
        Creates a new hostname template.
    
    .DESCRIPTION
        Creates a new template for hostname generation with specified groups and sequence parameters.
    
    .PARAMETER Name
        Name of the template.
    
    .PARAMETER Description
        Description of the template.
    
    .PARAMETER MaxLength
        Maximum length of hostnames generated with this template.
    
    .PARAMETER SequenceStart
        Starting number for the sequence.
    
    .PARAMETER SequenceLength
        Length of the sequence portion of hostnames.
    
    .PARAMETER SequencePadding
        Whether to pad sequence numbers with zeros.
    
    .PARAMETER SequenceIncrement
        Increment value for sequence numbers.
    
    .PARAMETER Groups
        Array of groups for the template, each with a name, length, validation type, validation value, and required flag.
    
    .PARAMETER CreatedBy
        Username of the creator. If not specified, will attempt to determine from authentication context.
    
    .EXAMPLE
        $groups = @(
            @{
                name = "location"
                length = 2
                is_required = $true
                validation_type = "list"
                validation_value = "NY,LA,SF,DC"
            },
            @{
                name = "environment"
                length = 3
                is_required = $true
                validation_type = "list"
                validation_value = "DEV,TST,STG,PRD"
            },
            @{
                name = "sequence"
                length = 3
                is_required = $true
                validation_type = "sequence"
                validation_value = ""
            }
        )

        New-HnsTemplate -Name "Server Template" -Description "For server hostnames" -MaxLength 15 -SequenceStart 1 -SequenceLength 3 -SequencePadding $true -SequenceIncrement 1 -Groups $groups
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,
        
        [Parameter(Mandatory = $false)]
        [string]$Description = "",
        
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, 64)]
        [int]$MaxLength,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$SequenceStart = 1,
        
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, 10)]
        [int]$SequenceLength,
        
        [Parameter(Mandatory = $false)]
        [bool]$SequencePadding = $true,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, 100)]
        [int]$SequenceIncrement = 1,
        
        [Parameter(Mandatory = $true)]
        [ValidateNotNull()]
        [array]$Groups,
        
        [Parameter(Mandatory = $false)]
        [string]$CreatedBy = ""
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }
    
    # If CreatedBy is not provided, try to get username from authentication context
    if ([string]::IsNullOrEmpty($CreatedBy)) {
        # Check if we have a username from JWT auth
        $username = $null
        foreach ($key in $script:DefaultHeaders.Keys) {
            if ($key -eq "Authorization" -and $script:DefaultHeaders[$key].StartsWith("Bearer ")) {
                $username = "jwt_user"
                break
            }
        }
        
        # If not, check if we're using API key
        if (-not $username -and $script:ApiKey) {
            $username = "api_user"
        }
        
        # If we still don't have a username, use a default
        if (-not $username) {
            $username = "powershell_user"
        }
        
        $CreatedBy = $username
    }

    $body = @{
        name = $Name
        description = $Description
        max_length = $MaxLength
        sequence_start = $SequenceStart
        sequence_length = $SequenceLength
        sequence_padding = $SequencePadding
        sequence_increment = $SequenceIncrement
        created_by = $CreatedBy
        groups = $Groups
    } | ConvertTo-Json -Depth 10

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/templates" -Method Post -Body $body -Headers $script:DefaultHeaders
        Write-Host "Template created successfully with ID: $($response.id)"
        return $response
    }
    catch {
        $errorMessage = "Failed to create template`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function Remove-HnsTemplate {
    <#
    .SYNOPSIS
        Deletes a hostname template.
    
    .DESCRIPTION
        Permanently removes a template from the HNS system. Checks for dependencies before attempting deletion.
    
    .PARAMETER Id
        The ID of the template to delete.
    
    .PARAMETER Force
        If specified, attempts to release associated hostnames before deleting the template.
        
    .PARAMETER CheckOnly
        If specified, only checks for dependencies without attempting to delete the template.
    
    .EXAMPLE
        Remove-HnsTemplate -Id 3
        
    .EXAMPLE
        Remove-HnsTemplate -Id 3 -Force
        
    .EXAMPLE
        Remove-HnsTemplate -Id 3 -CheckOnly
    #>
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$Id,
        
        [Parameter(Mandatory = $false)]
        [switch]$Force,
        
        [Parameter(Mandatory = $false)]
        [switch]$CheckOnly
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    # First, check for associated hostnames
    try {
        Write-Verbose "Checking template $Id for associated hostnames..."
        $template = Get-HnsTemplate -Id $Id
        
        if (-not $template) {
            Write-Error "Template with ID $Id not found."
            return $false
        }
        
        # Check for hostnames using this template
        $uri = "$script:BaseUrl/api/hostnames?template_id=$Id&limit=1"
        $hostnameCheck = Invoke-RestMethod -Uri $uri -Method Get -Headers $script:DefaultHeaders
        
        if ($hostnameCheck.total -gt 0) {
            $errorMsg = "Template '$($template.name)' (ID: $Id) has $($hostnameCheck.total) associated hostnames that must be handled first."
            
            if ($CheckOnly) {
                Write-Output $errorMsg
                return $false
            }
            
            if ($Force) {
                Write-Warning $errorMsg
                Write-Host "Force flag specified. Attempting to release and clean up hostnames..." -ForegroundColor Yellow
                
                # Implementation for Force parameter would go here
                # This could be quite complex and is not implemented in this example
                Write-Warning "Force cleanup not implemented in this version. Please use HNS-Cleanup.ps1 script instead."
                return $false
            }
            else {
                Write-Error $errorMsg
                Write-Host "To check all dependencies: Remove-HnsTemplate -Id $Id -CheckOnly" -ForegroundColor Cyan
                Write-Host "To delete all associated hostnames manually:" -ForegroundColor Cyan
                Write-Host "1. Get all hostnames: Get-HnsHostname -TemplateId $Id" -ForegroundColor Cyan
                Write-Host "2. Commit any reserved hostnames: Set-HnsHostnameCommit -HostnameId <id>" -ForegroundColor Cyan
                Write-Host "3. Release any committed hostnames: Set-HnsHostnameRelease -HostnameId <id>" -ForegroundColor Cyan
                Write-Host "4. Then retry template deletion" -ForegroundColor Cyan
                Write-Host "Or use the cleanup script: ./HNS-Cleanup.ps1 to handle this automatically" -ForegroundColor Green
                return $false
            }
        }
        
        if ($CheckOnly) {
            Write-Host "Template '$($template.name)' (ID: $Id) has no associated hostnames and can be safely deleted." -ForegroundColor Green
            return $true
        }
        
        # If we get here, there are no dependencies and we can delete if not in CheckOnly mode
        if ($PSCmdlet.ShouldProcess("Template with ID $Id", "Delete")) {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/templates/$Id" -Method Delete -Headers $script:DefaultHeaders
            Write-Host "Template '$($template.name)' (ID: $Id) deleted successfully." -ForegroundColor Green
            return $true
        }
    }
    catch {
        # Check if the error response contains a structured error
        try {
            $errorObj = $_.ErrorDetails.Message | ConvertFrom-Json
            if ($errorObj.error -and $errorObj.message) {
                Write-Error "$($errorObj.error): $($errorObj.message)"
                
                # Provide helpful suggestions based on error
                if ($errorObj.error -like "*dependencies*" -or $errorObj.error -like "*constraint*") {
                    Write-Host "Suggested actions:" -ForegroundColor Cyan
                    Write-Host "1. Use Remove-HnsTemplate -Id $Id -CheckOnly to see details" -ForegroundColor Cyan
                    Write-Host "2. Use the cleanup script: ./HNS-Cleanup.ps1 to handle this automatically" -ForegroundColor Cyan
                }
                return $false
            }
        }
        catch {
            # Could not parse as JSON, continue with standard error handling
        }
        
        $errorMessage = "Failed to delete template $Id`: $_"
        Write-Error $errorMessage
        
        # Try to provide helpful context based on error message
        if ($_.Exception.Message -like "*foreign key constraint*") {
            Write-Host "This error indicates the template has associated hostnames or groups." -ForegroundColor Yellow
            Write-Host "You need to remove all hostnames using this template before it can be deleted." -ForegroundColor Yellow
            Write-Host "Consider using the cleanup script: ./HNS-Cleanup.ps1" -ForegroundColor Green
        }
        
        return $false
    }
}
#endregion

#region Hostname Management Functions

function Get-HnsHostname {
    <#
    .SYNOPSIS
        Gets one or more hostnames.
    
    .DESCRIPTION
        Retrieves either a specific hostname by ID or a list of hostnames with optional filtering.
        By default, returns up to 100 results without requiring additional parameters.
    
    .PARAMETER Id
        The ID of a specific hostname to retrieve. If specified, returns just that hostname.
    
    .PARAMETER Limit
        Maximum number of hostnames to return when listing hostnames (default: 100).
    
    .PARAMETER Offset
        Number of hostnames to skip when listing hostnames (default: 0).
    
    .PARAMETER Status
        Filter by hostname status (available, reserved, committed, released).
    
    .PARAMETER TemplateId
        Filter by template ID.
    
    .PARAMETER Name
        Filter by hostname name (supports partial matches).
    
    .PARAMETER ReservedBy
        Filter by the username who reserved the hostname.
    
    .EXAMPLE
        Get-HnsHostname
        
        # Returns the first 100 hostnames
    
    .EXAMPLE
        Get-HnsHostname -Id 123
        
        # Returns the specific hostname with ID 123
    
    .EXAMPLE
        Get-HnsHostname -Status "reserved"
        
        # Returns up to 100 reserved hostnames
        
    .EXAMPLE
        Get-HnsHostname -TemplateId 5 -Status "committed"
        
        # Returns up to 100 committed hostnames from template ID 5
    #>
    [CmdletBinding(DefaultParameterSetName = 'List')]
    param(
        [Parameter(Mandatory = $true, ParameterSetName = 'Single')]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$Id,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateRange(1, 500)]
        [int]$Limit = 100,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateRange(0, [int]::MaxValue)]
        [int]$Offset = 0,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateSet("available", "reserved", "committed", "released")]
        [string]$Status,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$TemplateId,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [string]$Name,
        
        [Parameter(Mandatory = $false, ParameterSetName = 'List')]
        [string]$ReservedBy
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    try {
        if ($PSCmdlet.ParameterSetName -eq 'Single') {
            # Get a single hostname by ID
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/hostnames/$Id" -Method Get -Headers $script:DefaultHeaders
            return $response
        }
        else {
            # Build query string for filtering
            $queryParams = [System.Web.HttpUtility]::ParseQueryString([string]::Empty)
            $queryParams.Add("limit", $Limit)
            $queryParams.Add("offset", $Offset)

            # Add filters if specified
            if ($Status) { $queryParams.Add("status", $Status) }
            if ($TemplateId) { $queryParams.Add("template_id", $TemplateId) }
            if ($Name) { $queryParams.Add("name", $Name) }
            if ($ReservedBy) { $queryParams.Add("reserved_by", $ReservedBy) }

            # Construct the full URI
            $uri = "$script:BaseUrl/api/hostnames?$($queryParams.ToString())"
            Write-Verbose "Requesting: $uri"
            
            # Make the API call
            $response = Invoke-RestMethod -Uri $uri -Method Get -Headers $script:DefaultHeaders
            
            # Add metadata as custom properties to the result
            $result = $response.hostnames
            if ($result -and $result.Count -gt 0) {
                # Create a custom object to hold the result with metadata
                $outputObject = [PSCustomObject]@{
                    hostnames = $result
                    total = $response.total
                    limit = $response.limit
                    offset = $response.offset
                    count = $result.Count
                }
                
                # Add a custom property to support array and object behavior
                # This allows the command to be used both ways:
                # $hostnames = Get-HnsHostname
                # foreach ($hostname in $hostnames) { ... }
                # AND
                # $result = Get-HnsHostname
                # $result.total - to get total count
                Add-Member -InputObject $outputObject -MemberType ScriptMethod -Name "GetEnumerator" -Value {
                    return $this.hostnames.GetEnumerator()
                }
                
                # Add these methods for PowerShell array-like behavior
                Add-Member -InputObject $outputObject -MemberType ScriptProperty -Name "Count" -Value { 
                    return $this.hostnames.Count 
                }
                
                Add-Member -InputObject $outputObject -MemberType ScriptMethod -Name "get_Item" -Value {
                    param($index)
                    return $this.hostnames[$index]
                }
                
                # Return the enhanced result
                return $outputObject
            }
            
            # Return empty array if no results
            return @()
        }
    }
    catch {
        if ($PSCmdlet.ParameterSetName -eq 'Single') {
            $errorContext = "Failed to get hostname $Id"
        } else {
            $errorContext = "Failed to get hostnames"
        }
        $errorMessage = "$errorContext`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function New-HnsHostnameGeneration {
    <#
    .SYNOPSIS
        Generates a new hostname based on a template.
    
    .DESCRIPTION
        Generates a hostname using the specified template and parameters without reserving it.
    
    .PARAMETER TemplateId
        The ID of the template to use.
    
    .PARAMETER SequenceNum
        Optional specific sequence number to use.
    
    .PARAMETER Params
        Hashtable of parameters for the template groups (e.g., @{location = "NY"; environment = "DEV"}).
    
    .PARAMETER CheckDns
        Whether to check if the hostname exists in DNS.
    
    .EXAMPLE
        New-HnsHostnameGeneration -TemplateId 1 -Params @{location = "NY"; environment = "DEV"}
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$TemplateId,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(0, [int]::MaxValue)]
        [int]$SequenceNum = 0,
        
        [Parameter(Mandatory = $false)]
        [hashtable]$Params = @{},
        
        [Parameter(Mandatory = $false)]
        [switch]$CheckDns
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    $body = @{
        template_id = $TemplateId
        sequence_num = $SequenceNum
        params = $Params
    } | ConvertTo-Json

    $uri = "$script:BaseUrl/api/hostnames/generate"
    if ($CheckDns) {
        $uri += "?check_dns=true"
    }

    try {
        $response = Invoke-RestMethod -Uri $uri -Method Post -Body $body -Headers $script:DefaultHeaders
        return $response
    }
    catch {
        $errorMessage = "Failed to generate hostname`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function New-HnsHostnameReservation {
    <#
    .SYNOPSIS
        Reserves a new hostname based on a template.
    
    .DESCRIPTION
        Generates and reserves a hostname using the specified template and parameters.
    
    .PARAMETER TemplateId
        The ID of the template to use.
    
    .PARAMETER Params
        Hashtable of parameters for the template groups (e.g., @{location = "NY"; environment = "DEV"}).
    
    .PARAMETER RequestedBy
        The username or identifier of the person requesting the hostname reservation.
    
    .EXAMPLE
        New-HnsHostnameReservation -TemplateId 1 -Params @{location = "NY"; environment = "DEV"} -RequestedBy "admin"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$TemplateId,
        
        [Parameter(Mandatory = $false)]
        [hashtable]$Params = @{},
        
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$RequestedBy
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    # Create the request body
    $body = @{
        template_id = $TemplateId
        params = $Params
        requested_by = $RequestedBy
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/hostnames/reserve" -Method Post -Body $body -Headers $script:DefaultHeaders
        Write-Host "Hostname $($response.name) reserved successfully with ID: $($response.id)" -ForegroundColor Green
        return $response
    }
    catch {
        $errorMessage = "Failed to reserve hostname`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function Set-HnsHostnameCommit {
    <#
    .SYNOPSIS
        Commits a reserved hostname.
    
    .DESCRIPTION
        Changes the status of a hostname from reserved to committed.
    
    .PARAMETER HostnameId
        The ID of the hostname to commit.
    
    .EXAMPLE
        Set-HnsHostnameCommit -HostnameId 123
    #>
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$HostnameId
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    $body = @{
        hostname_id = $HostnameId
    } | ConvertTo-Json

    try {
        if ($PSCmdlet.ShouldProcess("Hostname ID $HostnameId", "Commit")) {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/hostnames/commit" -Method Post -Body $body -Headers $script:DefaultHeaders
            Write-Host "Hostname $($response.name) committed successfully"
            return $response
        }
    }
    catch {
        $errorMessage = "Failed to commit hostname`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function Set-HnsHostnameRelease {
    <#
    .SYNOPSIS
        Releases a committed hostname.
    
    .DESCRIPTION
        Changes the status of a hostname from committed to released.
    
    .PARAMETER HostnameId
        The ID of the hostname to release.
    
    .EXAMPLE
        Set-HnsHostnameRelease -HostnameId 123
    #>
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$HostnameId
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    $body = @{
        hostname_id = $HostnameId
    } | ConvertTo-Json

    try {
        if ($PSCmdlet.ShouldProcess("Hostname ID $HostnameId", "Release")) {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/hostnames/release" -Method Post -Body $body -Headers $script:DefaultHeaders
            Write-Host "Hostname $($response.name) released successfully"
            return $response
        }
    }
    catch {
        $errorMessage = "Failed to release hostname`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

#endregion

#region DNS Functions

function Test-HnsDnsHostname {
    <#
    .SYNOPSIS
        Checks if a hostname exists in DNS.
    
    .DESCRIPTION
        Verifies if a hostname exists in DNS by querying the API service.
    
    .PARAMETER Hostname
        The hostname to check.
    
    .EXAMPLE
        Test-HnsDnsHostname -Hostname "server1.example.com"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Hostname
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/dns/check/$Hostname" -Method Get -Headers $script:DefaultHeaders
        return $response
    }
    catch {
        $errorMessage = "Failed to check hostname in DNS`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function Start-HnsDnsScan {
    <#
    .SYNOPSIS
        Scans a range of hostnames in DNS.
    
    .DESCRIPTION
        Scans a range of hostnames generated from a template to see if they exist in DNS.
    
    .PARAMETER TemplateId
        The ID of the template to use.
    
    .PARAMETER StartSeq
        The starting sequence number.
    
    .PARAMETER EndSeq
        The ending sequence number.
    
    .PARAMETER Params
        Hashtable of parameters for the template groups.
    
    .PARAMETER MaxConcurrent
        Maximum number of concurrent DNS checks.
    
    .EXAMPLE
        Start-HnsDnsScan -TemplateId 1 -StartSeq 1 -EndSeq 100 -Params @{location = "NY"; environment = "DEV"} -MaxConcurrent 10
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$TemplateId,
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$StartSeq = 1,
        
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$EndSeq,
        
        [Parameter(Mandatory = $false)]
        [hashtable]$Params = @{},
        
        [Parameter(Mandatory = $false)]
        [ValidateRange(1, 50)]
        [int]$MaxConcurrent = 10
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    if ($EndSeq -lt $StartSeq) {
        throw "EndSeq must be greater than or equal to StartSeq."
    }

    $body = @{
        template_id = $TemplateId
        start_seq = $StartSeq
        end_seq = $EndSeq
        params = $Params
        max_concurrent = $MaxConcurrent
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/dns/scan" -Method Post -Body $body -Headers $script:DefaultHeaders
        return $response
    }
    catch {
        $errorMessage = "Failed to scan DNS`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}
#endregion

#region API Key Management Functions

function Get-HnsApiKeys {
    <#
    .SYNOPSIS
        Gets a list of API keys for the authenticated user.
    
    .DESCRIPTION
        Retrieves all API keys associated with the authenticated user.
    
    .EXAMPLE
        Get-HnsApiKeys
    #>
    [CmdletBinding()]
    param()

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    if (-not $script:JwtToken) {
        throw "Not authenticated with JWT token. Call Connect-HnsService first."
    }

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/apikeys" -Method Get -Headers $script:DefaultHeaders
        return $response.api_keys
    }
    catch {
        $errorMessage = "Failed to get API keys`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function New-HnsApiKey {
    <#
    .SYNOPSIS
        Creates a new API key.
    
    .DESCRIPTION
        Creates a new API key with the specified name and scope.
    
    .PARAMETER Name
        The name for the API key.
    
    .PARAMETER Scope
        The scope for the API key (comma-separated list of permissions).
    
    .EXAMPLE
        New-HnsApiKey -Name "Automation key" -Scope "read,reserve"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,
        
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Scope
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    if (-not $script:JwtToken) {
        throw "Not authenticated with JWT token. Call Connect-HnsService first."
    }

    $body = @{
        name = $Name
        scope = $Scope
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/apikeys" -Method Post -Body $body -Headers $script:DefaultHeaders
        Write-Host "API key created successfully. Make sure to save the key value as it won't be shown again."
        return $response
    }
    catch {
        $errorMessage = "Failed to create API key`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

function Remove-HnsApiKey {
    <#
    .SYNOPSIS
        Deletes an API key.
    
    .DESCRIPTION
        Deletes the API key with the specified ID.
    
    .PARAMETER Id
        The ID of the API key to delete.
    
    .EXAMPLE
        Remove-HnsApiKey -Id 5
    #>
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateRange(1, [int]::MaxValue)]
        [int]$Id
    )

    if (-not $script:BaseUrl) {
        throw "HNS connection not initialized. Call Initialize-HnsConnection first."
    }

    if (-not $script:JwtToken) {
        throw "Not authenticated with JWT token. Call Connect-HnsService first."
    }

    try {
        if ($PSCmdlet.ShouldProcess("API key with ID $Id", "Delete")) {
            $response = Invoke-RestMethod -Uri "$script:BaseUrl/api/apikeys/$Id" -Method Delete -Headers $script:DefaultHeaders
            Write-Host "API key deleted successfully"
            return $true
        }
    }
    catch {
        $errorMessage = "Failed to delete API key`: $_"
        Write-Error $errorMessage
        throw $errorMessage
    }
}

#endregion

#region