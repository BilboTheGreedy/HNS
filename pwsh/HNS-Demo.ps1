# HNS-Demo.ps1
# Demonstration script for the HNS PowerShell module

# Import the module - modify path as needed
Import-Module -Name '.\HNS-API.psm1' -Force

# Set variables
$HnsUrl = "http://localhost:8080"  # Change to your HNS server URL
$Username = "test"                # Change to your username
$Password = ConvertTo-SecureString "Logon123!" -AsPlainText -Force  # Change to your password

# Helper function for console output without creating return objects
function Write-Console {
    param([string]$Message, [ConsoleColor]$ForegroundColor = [ConsoleColor]::White)
    [Console]::ForegroundColor = $ForegroundColor
    [Console]::WriteLine($Message)
    [Console]::ResetColor()
}

function Show-Header {
    param([string] $Title)
    
    Write-Console "`n========================================================" -ForegroundColor Cyan
    Write-Console " $Title" -ForegroundColor Cyan
    Write-Console "========================================================" -ForegroundColor Cyan
}

function Pause-ForReview {
    Write-Console "`nPress any key to continue..." -ForegroundColor Yellow
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

function Demo-TemplateOperations {
    [OutputType([int])]
    [CmdletBinding()]
    param()
    
    Show-Header "Template Operations"
    
    # List existing templates
    Write-Console "Listing templates..."
    $templates = @(Get-HnsTemplates -Limit 5)
    if ($templates.Count -eq 0) {
        Write-Console "No templates found. Let's create one!" -ForegroundColor Yellow
    } else {
        Write-Console "Found $($templates.Count) templates:" -ForegroundColor Green
        foreach ($template in $templates) {
            Write-Console "  ID: $($template.id), Name: $($template.name), Max Length: $($template.max_length)"
        }
    }
    
    # Create a new template
    Write-Console "`nCreating a new template..." -ForegroundColor Cyan
    $groups = @(
        @{
            name = "location"
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "NY,LA,SF,DC"
        },
        @{
            name = "type"
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "WS,DB,AP"
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
    
    $newTemplate = New-HnsTemplate -Name "ServerTemplate" -Description "Server naming template for PowerShell demo" `
        -MaxLength 15 -SequenceStart 1 -SequenceLength 3 -SequencePadding $true -SequenceIncrement 1 -Groups $groups `
        -CreatedBy "PowerShell_Demo_User"
    
    # Force to integer 
    [int]$resultId = $newTemplate.id
    
    Write-Console "New template created with ID: $resultId" -ForegroundColor Green
    
    # Get template details
    Write-Console "`nGetting template details for ID $resultId..." -ForegroundColor Cyan
    $templateDetails = Get-HnsTemplate -Id $resultId
    
    Write-Console "Template details:" -ForegroundColor Green
    Write-Console "  Name: $($templateDetails.name)"
    Write-Console "  Description: $($templateDetails.description)"
    Write-Console "  Max Length: $($templateDetails.max_length)"
    Write-Console "  Sequence Start: $($templateDetails.sequence_start)"
    
    Write-Console "`nTemplate Groups:" -ForegroundColor Cyan
    foreach ($group in $templateDetails.groups) {
        Write-Console "  $($group.name): Length=$($group.length), Position=$($group.position), Type=$($group.validation_type)"
    }
    
    # Return ONLY the integer - no additional output
    return $resultId
}

function Demo-HostnameOperations {
    [CmdletBinding()]
    param([int] $TemplateId)
    
    Show-Header "Hostname Operations"
    
    # Generate a hostname without reserving it
    Write-Console "Generating a hostname using template $TemplateId..." -ForegroundColor Cyan
    $params = @{
        location = "NY"
        type = "WS"
        environment = "DEV"
    }
    
    $generatedHostname = New-HnsHostnameGeneration -TemplateId $TemplateId -Params $params
    Write-Console "Generated hostname: $($generatedHostname.hostname)" -ForegroundColor Green
    
    # Get the next sequence number
    Write-Console "`nGetting next sequence number for template $TemplateId..." -ForegroundColor Cyan
    $nextSeq = Get-HnsNextSequenceNumber -TemplateId $TemplateId
    Write-Console "Next sequence number: $($nextSeq.sequence_num)" -ForegroundColor Green
    
    # Reserve a hostname
    Write-Console "`nReserving a hostname using template $TemplateId..." -ForegroundColor Cyan
    $params = @{
        location = "LA"
        type = "DB"
        environment = "TST"
    }
    
    $reservedHostname = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params -RequestedBy "Demo_User"
    Write-Console "Reserved hostname: $($reservedHostname.name) with ID: $($reservedHostname.id)" -ForegroundColor Green
    
    # List reserved hostnames
    Write-Console "`nListing reserved hostnames..." -ForegroundColor Cyan
    $reservedHostnames = @(Get-HnsHostnames -Status "reserved" -Limit 5)
    if ($reservedHostnames.Count -gt 0) {
        Write-Console "Found $($reservedHostnames.Count) reserved hostnames:" -ForegroundColor Green
        foreach ($hostname in $reservedHostnames) {
            Write-Console "  ID: $($hostname.id), Name: $($hostname.name), Status: $($hostname.status)"
        }
    } else {
        Write-Console "No reserved hostnames found." -ForegroundColor Yellow
    }
    
    # Get details of a specific hostname
    Write-Console "`nGetting details for hostname ID $($reservedHostname.id)..." -ForegroundColor Cyan
    $hostnameDetails = Get-HnsHostname -Id $reservedHostname.id
    Write-Console "Hostname details:" -ForegroundColor Green
    Write-Console "  Name: $($hostnameDetails.name)"
    Write-Console "  Status: $($hostnameDetails.status)"
    Write-Console "  Sequence: $($hostnameDetails.sequence_num)"
    Write-Console "  Reserved by: $($hostnameDetails.reserved_by)"
    
    # Commit the hostname
    Write-Console "`nCommitting hostname $($reservedHostname.name)..." -ForegroundColor Cyan
    $committedHostname = Set-HnsHostnameCommit -HostnameId $reservedHostname.id
    Write-Console "Hostname committed successfully. New status: $($committedHostname.status)" -ForegroundColor Green
    
    # List committed hostnames
    Write-Console "`nListing committed hostnames..." -ForegroundColor Cyan
    $committedHostnames = @(Get-HnsHostnames -Status "committed" -Limit 5)
    if ($committedHostnames.Count -gt 0) {
        Write-Console "Found $($committedHostnames.Count) committed hostnames:" -ForegroundColor Green
        foreach ($hostname in $committedHostnames) {
            Write-Console "  ID: $($hostname.id), Name: $($hostname.name), Status: $($hostname.status)"
        }
    } else {
        Write-Console "No committed hostnames found." -ForegroundColor Yellow
    }
    
    # Release the hostname
    Write-Console "`nReleasing hostname $($reservedHostname.name)..." -ForegroundColor Cyan
    $releasedHostname = Set-HnsHostnameRelease -HostnameId $reservedHostname.id
    Write-Console "Hostname released successfully. New status: $($releasedHostname.status)" -ForegroundColor Green
    
    # Reserve multiple hostnames with different parameters
    Write-Console "`nReserving multiple hostnames with different parameters..." -ForegroundColor Cyan
    $locations = @("NY", "LA", "SF", "DC")
    $types = @("WS", "DB", "AP")
    $environments = @("DEV", "TST")
    
    $reservedIds = @()
    
    for ($i = 0; $i -lt 3; $i++) {
        $location = $locations[(Get-Random -Minimum 0 -Maximum $locations.Length)]
        $type = $types[(Get-Random -Minimum 0 -Maximum $types.Length)]
        $environment = $environments[(Get-Random -Minimum 0 -Maximum $environments.Length)]
        
        $params = @{
            location = $location
            type = $type
            environment = $environment
        }
        
        Write-Console "Reserving hostname with params: location=$location, type=$type, environment=$environment" -ForegroundColor DarkGray
        $hostname = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params -RequestedBy "Demo_User"
        $reservedIds += $hostname.id
    }
    
    # Return the reserved IDs for later use
    return $reservedIds
}

function Demo-DnsOperations {
    [CmdletBinding()]
    param([int] $TemplateId)
    
    Show-Header "DNS Operations"
    
    # Check if a hostname exists in DNS
    Write-Console "Checking if a hostname exists in DNS..." -ForegroundColor Cyan
    $params = @{
        location = "SF"
        type = "WS"
        environment = "PRD"
    }
    
    # First generate the hostname
    $generatedHostname = New-HnsHostnameGeneration -TemplateId $TemplateId -Params $params
    $hostname = $generatedHostname.hostname
    
    # Then check DNS
    Write-Console "Checking DNS for hostname: $hostname" -ForegroundColor Cyan
    $dnsResult = Test-HnsDnsHostname -Hostname $hostname
    
    if ($dnsResult.exists) {
        Write-Console "Hostname $hostname exists in DNS with IP: $($dnsResult.ip_address)" -ForegroundColor Green
    } else {
        Write-Console "Hostname $hostname does not exist in DNS" -ForegroundColor Yellow
    }
    
    # Scan a range of hostnames in DNS
    Write-Console "`nScanning a range of hostnames in DNS..." -ForegroundColor Cyan
    $params = @{
        location = "NY"
        type = "WS"
        environment = "DEV"
    }
    
    Write-Console "Scanning sequence range 1-10 for template $TemplateId..." -ForegroundColor Cyan
    $scanResult = Start-HnsDnsScan -TemplateId $TemplateId -StartSeq 1 -EndSeq 10 -Params $params -MaxConcurrent 5
    
    Write-Console "Scan completed:" -ForegroundColor Green
    Write-Console "  Total hostnames scanned: $($scanResult.total_hostnames)" -ForegroundColor Green
    Write-Console "  Existing hostnames found: $($scanResult.existing_hostnames)" -ForegroundColor Green
    Write-Console "  Scan duration: $($scanResult.scan_duration)" -ForegroundColor Green
    
    if ($scanResult.existing_hostnames -gt 0) {
        Write-Console "`nExisting hostnames:" -ForegroundColor Cyan
        $existingHosts = $scanResult.results | Where-Object { $_.exists -eq $true }
        foreach ($host in $existingHosts) {
            Write-Console "  $($host.hostname): $($host.ip_address)"
        }
    }
}

function Demo-ApiKeyOperations {
    Show-Header "API Key Operations"
    
    # List existing API keys
    Write-Console "Listing existing API keys..." -ForegroundColor Cyan
    $existingKeys = @(Get-HnsApiKeys)
    
    if ($existingKeys.Count -gt 0) {
        Write-Console "Found $($existingKeys.Count) API keys:" -ForegroundColor Green
        foreach ($key in $existingKeys) {
            Write-Console "  ID: $($key.id), Name: $($key.name), Scope: $($key.scope)"
        }
    } else {
        Write-Console "No API keys found." -ForegroundColor Yellow
    }
    
    # Create a new API key
    Write-Console "`nCreating a new API key..." -ForegroundColor Cyan
    $keyName = "PowerShell-Demo-Key-" + (Get-Random -Minimum 1000 -Maximum 9999)
    $scope = "read,reserve"
    
    $newKey = New-HnsApiKey -Name $keyName -Scope $scope
    Write-Console "New API key created:" -ForegroundColor Green
    Write-Console "  Name: $($newKey.name)" -ForegroundColor Green
    Write-Console "  Key: $($newKey.key)" -ForegroundColor Green
    Write-Console "  Scope: $($newKey.scope)" -ForegroundColor Green
    Write-Console "  Expires: $($newKey.expires_at)" -ForegroundColor Green
    
    # Store the key for using it later
    $apiKeyValue = $newKey.key
    
    # List API keys again to see the new one
    Write-Console "`nListing API keys after creation..." -ForegroundColor Cyan
    $updatedKeys = @(Get-HnsApiKeys)
    foreach ($key in $updatedKeys) {
        Write-Console "  ID: $($key.id), Name: $($key.name), Scope: $($key.scope)"
    }
    
    # Test the API key
    Write-Console "`nTesting the new API key..." -ForegroundColor Cyan
    
    # Save the current connection info
    $currentHeaders = $script:DefaultHeaders.Clone()
    
    # Initialize a new connection with the API key
    Initialize-HnsConnection -BaseUrl $HnsUrl -ApiKey $apiKeyValue
    
    # Test connection
    $testResult = Test-HnsConnection
    if ($testResult) {
        Write-Console "Successfully connected using the new API key!" -ForegroundColor Green
        
        # Try to list templates (should work with read scope)
        $templates = @(Get-HnsTemplates -Limit 3)
        Write-Console "Successfully retrieved $($templates.Count) templates using API key" -ForegroundColor Green
    }
    
    # Restore the original connection
    $script:DefaultHeaders = $currentHeaders
    
    # Delete the API key
    Write-Console "`nDeleting the API key..." -ForegroundColor Cyan
    Remove-HnsApiKey -Id $newKey.id
    
    # List API keys one more time to confirm deletion
    Write-Console "`nListing API keys after deletion..." -ForegroundColor Cyan
    $finalKeys = @(Get-HnsApiKeys)
    if ($finalKeys.Count -lt $updatedKeys.Count) {
        Write-Console "API key successfully deleted!" -ForegroundColor Green
    }
    foreach ($key in $finalKeys) {
        Write-Console "  ID: $($key.id), Name: $($key.name), Scope: $($key.scope)"
    }
}

# Main demo script
try {
    # Initialize connection to the HNS server
    Write-Console "Initializing connection to HNS server at $HnsUrl..." -ForegroundColor Cyan
    Initialize-HnsConnection -BaseUrl $HnsUrl
    
    # Authenticate
    Write-Console "Authenticating with username $Username..." -ForegroundColor Cyan
    Connect-HnsService -Username $Username -Password $Password
    
    # Test the connection
    $connected = Test-HnsConnection
    if (-not $connected) {
        throw "Failed to connect to HNS server."
    }
    
    # CRITICAL: Get template ID as an integer and store it properly
    [int]$templateId = Demo-TemplateOperations
    Write-Console "Template ID returned: $templateId (Type: $($templateId.GetType().FullName))" -ForegroundColor DarkCyan
    Pause-ForReview
    
    # Run hostname operations with the template
    $reservedIds = Demo-HostnameOperations -TemplateId $templateId
    Pause-ForReview
    
    # Run DNS operations - THIS IS LINE 332 WHERE THE ERROR HAPPENS
    # Make sure to pass an integer, not an array
    Demo-DnsOperations -TemplateId ([int]$templateId)
    Pause-ForReview
    
    # Run API key operations
    Demo-ApiKeyOperations
    
    Write-Console "`nDemo completed successfully!" -ForegroundColor Green
}
catch {
    Write-Console "An error occurred: $_" -ForegroundColor Red
    Write-Console $_.ScriptStackTrace -ForegroundColor Red
}
finally {
    Write-Console "`nDemo script finished." -ForegroundColor Cyan
}