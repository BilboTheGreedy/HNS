# Simple-HNS-Test.ps1
# A simplified script to test HNS API commands individually

# Variables - edit these to match your environment
$HnsUrl = "http://localhost:8080"
$Username = "admin"
$Password = "admin123"
$ApiKey = "" # Leave empty if using username/password

# Helper function for colorful output
function Write-Color {
    param([string]$Text, [string]$Color = "White")
    Write-Host $Text -ForegroundColor $Color
}

# Initialize connection
function Initialize-Connection {
    Write-Color "Connecting to HNS API at $HnsUrl..." "Cyan"
    
    # Set up default headers
    $script:Headers = @{
        "Content-Type" = "application/json"
        "Accept" = "application/json"
    }
    
    # If API key is provided, use it
    if ($ApiKey) {
        $script:Headers["X-API-Key"] = $ApiKey
        Write-Color "Using API key authentication" "Green"
        return
    }
    
    # Otherwise use JWT authentication
    $loginBody = @{
        username = $Username
        password = $Password
    } | ConvertTo-Json
    
    try {
        $auth = Invoke-RestMethod -Uri "$HnsUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        $script:Headers["Authorization"] = "Bearer $($auth.token)"
        Write-Color "Successfully authenticated as $Username" "Green"
    }
    catch {
        Write-Color "Authentication failed: $_" "Red"
        exit
    }
}

# Test health check endpoint
function Test-Connection {
    Write-Color "Testing connection..." "Cyan"
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/health" -Method GET -Headers $script:Headers
        Write-Color "Connection successful! Status: $($response.status)" "Green"
    }
    catch {
        Write-Color "Connection test failed: $_" "Red"
    }
}

# List templates
function Get-Templates {
    Write-Color "Getting templates..." "Cyan"
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/templates?limit=10&offset=0" -Method GET -Headers $script:Headers
        Write-Color "Found $($response.templates.Count) templates:" "Green"
        foreach ($template in $response.templates) {
            Write-Color "  - ID: $($template.id), Name: $($template.name)" "White"
        }
        return $response.templates
    }
    catch {
        Write-Color "Failed to get templates: $_" "Red"
        return $null
    }
}

# Get a single template by ID
function Get-Template {
    param([int]$Id)
    Write-Color "Getting template with ID $Id..." "Cyan"
    try {
        $template = Invoke-RestMethod -Uri "$HnsUrl/api/templates/$Id" -Method GET -Headers $script:Headers
        Write-Color "Template details:" "Green"
        Write-Color "  Name: $($template.name)" "White"
        Write-Color "  Description: $($template.description)" "White"
        Write-Color "  Groups: $($template.groups.Count)" "White"
        foreach ($group in $template.groups) {
            Write-Color "    - $($group.name): Type=$($group.validation_type)" "White"
        }
        return $template
    }
    catch {
        Write-Color "Failed to get template: $_" "Red"
        return $null
    }
}

# Create a new template
function New-Template {
    Write-Color "Creating a new template..." "Cyan"
    
    $groups = @(
        @{
            name = "dc"
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "NY,LA,SF,DC"
        },
        @{
            name = "env"
            length = 3
            is_required = $true
            validation_type = "list"
            validation_value = "DEV,TST,PRD"
        },
        @{
            name = "app"
            length = 3
            is_required = $true
            validation_type = "list"
            validation_value = "WEB,APP,DB"
        },
        @{
            name = "seq"
            length = 3
            is_required = $true
            validation_type = "sequence"
            validation_value = ""
        }
    )
    
    $body = @{
        name = "Test-Template-" + (Get-Random -Minimum 1000 -Maximum 9999)
        description = "Test template created by PowerShell"
        max_length = 15
        sequence_start = 1
        sequence_length = 3
        sequence_padding = $true
        sequence_increment = 1
        created_by = "PowerShell"
        groups = $groups
    } | ConvertTo-Json -Depth 5
    
    try {
        $template = Invoke-RestMethod -Uri "$HnsUrl/api/templates" -Method POST -Body $body -Headers $script:Headers
        Write-Color "Template created successfully with ID: $($template.id)" "Green"
        return $template
    }
    catch {
        Write-Color "Failed to create template: $_" "Red"
        return $null
    }
}

# Generate a hostname without reserving it
function Test-GenerateHostname {
    param(
        [int]$TemplateId,
        [hashtable]$Params
    )
    
    Write-Color "Generating a hostname using template $TemplateId..." "Cyan"
    
    $body = @{
        template_id = $TemplateId
        params = $Params
    } | ConvertTo-Json
    
    try {
        $result = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/generate" -Method POST -Body $body -Headers $script:Headers
        Write-Color "Generated hostname: $($result.hostname)" "Green"
        return $result
    }
    catch {
        Write-Color "Failed to generate hostname: $_" "Red"
        return $null
    }
}

# Reserve a hostname
function Reserve-Hostname {
    param(
        [int]$TemplateId,
        [hashtable]$Params
    )
    
    Write-Color "Reserving a hostname using template $TemplateId..." "Cyan"
    
    $body = @{
        template_id = $TemplateId
        requested_by = "PowerShell-Test"
        params = $Params
    } | ConvertTo-Json
    
    try {
        $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/reserve" -Method POST -Body $body -Headers $script:Headers
        Write-Color "Reserved hostname: $($hostname.name) with ID: $($hostname.id)" "Green"
        return $hostname
    }
    catch {
        Write-Color "Failed to reserve hostname: $_" "Red"
        return $null
    }
}

# List hostnames
function Get-Hostnames {
    param(
        [string]$Status = "",
        [int]$TemplateId = 0
    )
    
    Write-Color "Listing hostnames..." "Cyan"
    
    $uri = "$HnsUrl/api/hostnames?limit=10"
    if ($Status) { $uri += "&status=$Status" }
    if ($TemplateId -gt 0) { $uri += "&template_id=$TemplateId" }
    
    try {
        $response = Invoke-RestMethod -Uri $uri -Method GET -Headers $script:Headers
        Write-Color "Found $($response.total) hostnames:" "Green"
        foreach ($hostname in $response.hostnames) {
            Write-Color "  - $($hostname.name) (ID: $($hostname.id), Status: $($hostname.status))" "White"
        }
        return $response.hostnames
    }
    catch {
        Write-Color "Failed to list hostnames: $_" "Red"
        return $null
    }
}

# Commit a hostname
function Commit-Hostname {
    param([int]$Id)
    
    Write-Color "Committing hostname with ID $Id..." "Cyan"
    
    $body = @{
        hostname_id = $Id
        committed_by = "PowerShell-Test"
    } | ConvertTo-Json
    
    try {
        $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/commit" -Method POST -Body $body -Headers $script:Headers
        Write-Color "Hostname committed successfully: $($hostname.name)" "Green"
        return $hostname
    }
    catch {
        Write-Color "Failed to commit hostname: $_" "Red"
        return $null
    }
}

# Release a hostname
function Release-Hostname {
    param([int]$Id)
    
    Write-Color "Releasing hostname with ID $Id..." "Cyan"
    
    $body = @{
        hostname_id = $Id
        released_by = "PowerShell-Test"
    } | ConvertTo-Json
    
    try {
        $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/release" -Method POST -Body $body -Headers $script:Headers
        Write-Color "Hostname released successfully: $($hostname.name)" "Green"
        return $hostname
    }
    catch {
        Write-Color "Failed to release hostname: $_" "Red"
        return $null
    }
}

# Check if a hostname exists in DNS
function Check-Dns {
    param([string]$Hostname)
    
    Write-Color "Checking DNS for hostname: $Hostname..." "Cyan"
    
    try {
        $result = Invoke-RestMethod -Uri "$HnsUrl/api/dns/check/$Hostname" -Method GET -Headers $script:Headers
        if ($result.exists) {
            Write-Color "Hostname exists in DNS with IP: $($result.ip_address)" "Green"
        } else {
            Write-Color "Hostname does not exist in DNS" "Yellow"
        }
        return $result
    }
    catch {
        Write-Color "Failed to check DNS: $_" "Red"
        return $null
    }
}

# Create an API key
function New-ApiKey {
    param(
        [string]$Name,
        [string]$Scope = "read,reserve"
    )
    
    Write-Color "Creating a new API key..." "Cyan"
    
    $body = @{
        name = $Name
        scope = $Scope
    } | ConvertTo-Json
    
    try {
        $key = Invoke-RestMethod -Uri "$HnsUrl/api/apikeys" -Method POST -Body $body -Headers $script:Headers
        Write-Color "API key created successfully!" "Green"
        Write-Color "Key: $($key.key)" "Yellow"
        Write-Color "IMPORTANT: Save this key value - it won't be shown again!" "Yellow"
        return $key
    }
    catch {
        Write-Color "Failed to create API key: $_" "Red"
        return $null
    }
}

function Remove-Template {
    param([int]$Id)
    
    Write-Color "Deleting template with ID $Id..." "Cyan"
    
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/templates/$Id" -Method DELETE -Headers $script:Headers
        Write-Color "Template deleted successfully!" "Green"
        return $true
    }
    catch {
        # Detailed error handling
        try {
            $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Color "Failed to delete template: $($errorDetails.error)" "Red"
            
            # If the error suggests hostname dependencies
            if ($errorDetails.error -like "*associated hostnames*") {
                Write-Color "The template has associated hostnames. You must:" "Yellow"
                Write-Color "1. Release all committed hostnames" "Gray"
                Write-Color "2. Remove all reserved/released hostnames first" "Gray"
                Write-Color "Use the following PowerShell functions to clean up:" "Gray"
                Write-Color "  - Get-Hostnames -TemplateId $Id -Status committed" "Gray"
                Write-Color "  - Get-Hostnames -TemplateId $Id -Status reserved" "Gray"
            }
        }
        catch {
            Write-Color "Failed to delete template: $_" "Red"
        }
        return $false
    }
}

# NEW: Delete a hostname
function Remove-Hostname {
    param(
        [Parameter(Mandatory = $true)]
        [int]$Id,
        
        [Parameter(Mandatory = $false)]
        [switch]$Force
    )
    
    Write-Color "Deleting hostname with ID $Id..." "Cyan"
    
    try {
        if ($Force) {
            # First, release the hostname if it's in a non-released state
            $hostname = Get-HnsHostname -Id $Id
            
            if ($hostname.status -eq "committed") {
                Write-Color "Releasing committed hostname first..." "Yellow"
                Release-Hostname -Id $Id
            }
            elseif ($hostname.status -eq "reserved") {
                Write-Color "Releasing and then committing reserved hostname..." "Yellow"
                Commit-Hostname -Id $Id
                Release-Hostname -Id $Id
            }
        }
        
        # Attempt to delete - note: actual deletion might not be supported by API
        Write-Color "WARNING: Direct hostname deletion may not be supported by the API." "Yellow"
        Write-Color "This operation might fail or be intentionally restricted." "Yellow"
        
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$Id" -Method DELETE -Headers $script:Headers
        Write-Color "Hostname deleted successfully!" "Green"
        return $true
    }
    catch {
        Write-Color "Failed to delete hostname: $_" "Red"
        return $false
    }
}

# NEW: Delete an API Key
function Remove-ApiKey {
    param([int]$Id)
    
    Write-Color "Deleting API key with ID $Id..." "Cyan"
    
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/apikeys/$Id" -Method DELETE -Headers $script:Headers
        Write-Color "API key deleted successfully!" "Green"
        return $true
    }
    catch {
        Write-Color "Failed to delete API key: $_" "Red"
        return $false
    }
}
# Updated Menu Function
function Show-Menu {
    Write-Color "`n===== HNS API Test Menu =====" "Magenta"
    Write-Color "1. Initialize Connection" "Cyan"
    Write-Color "2. Test Connection" "Cyan"
    Write-Color "3. List Templates" "Cyan"
    Write-Color "4. Get Template Details" "Cyan"
    Write-Color "5. Create New Template" "Cyan"
    Write-Color "6. Generate Hostname (without reserving)" "Cyan"
    Write-Color "7. Reserve a Hostname" "Cyan"
    Write-Color "8. List Hostnames" "Cyan"
    Write-Color "9. Commit a Hostname" "Cyan"
    Write-Color "10. Release a Hostname" "Cyan"
    Write-Color "11. Check DNS" "Cyan"
    Write-Color "12. Create API Key" "Cyan"
    # NEW: Deletion Options
    Write-Color "13. Delete a Template" "Red"
    Write-Color "14. Delete a Hostname" "Red"
    Write-Color "15. Delete an API Key" "Red"
    Write-Color "0. Exit" "Red"
    Write-Color "`nEnter your choice: " "Yellow" -NoNewline
    
    $choice = Read-Host
    return $choice
}
# Script execution
$templateId = 0
$hostnameId = 0
$apiKeyId = 0

while ($true) {
    $choice = Show-Menu
    
    switch ($choice) {
        "0" { 
            Write-Color "Exiting..." "Red"
            exit 
        }
        "1" { Initialize-Connection }
        "2" { Test-Connection }
        "3" { Get-Templates }
        "4" { 
            if ($templateId -eq 0) {
                $templateId = Read-Host "Enter template ID"
            }
            Get-Template -Id $templateId 
        }
        "5" { 
            $template = New-Template
            if ($template) {
                $templateId = $template.id
            }
        }
        "6" { 
            if ($templateId -eq 0) {
                $templateId = Read-Host "Enter template ID"
            }
            $dc = Read-Host "Enter data center (e.g., NY)"
            $env = Read-Host "Enter environment (e.g., DEV)"
            $app = Read-Host "Enter application (e.g., WEB)"
            
            $params = @{
                dc = $dc
                env = $env
                app = $app
            }
            
            Test-GenerateHostname -TemplateId $templateId -Params $params
        }
        "7" { 
            if ($templateId -eq 0) {
                $templateId = Read-Host "Enter template ID"
            }
            $dc = Read-Host "Enter data center (e.g., NY)"
            $env = Read-Host "Enter environment (e.g., DEV)"
            $app = Read-Host "Enter application (e.g., WEB)"
            
            $params = @{
                dc = $dc
                env = $env
                app = $app
            }
            
            $hostname = Reserve-Hostname -TemplateId $templateId -Params $params
            if ($hostname) {
                $hostnameId = $hostname.id
            }
        }
        "8" { 
            $status = Read-Host "Enter status to filter by (leave empty for all)"
            $templateFilter = Read-Host "Enter template ID to filter by (leave empty for all)"
            
            $templateFilterId = 0
            if ($templateFilter) {
                $templateFilterId = [int]$templateFilter
            }
            
            Get-Hostnames -Status $status -TemplateId $templateFilterId
        }
        "9" { 
            if ($hostnameId -eq 0) {
                $hostnameId = Read-Host "Enter hostname ID to commit"
            }
            Commit-Hostname -Id $hostnameId
        }
        "10" { 
            if ($hostnameId -eq 0) {
                $hostnameId = Read-Host "Enter hostname ID to release"
            }
            Release-Hostname -Id $hostnameId
        }
        "11" { 
            $hostname = Read-Host "Enter hostname to check in DNS"
            Check-Dns -Hostname $hostname
        }
        "12" { 
            $name = Read-Host "Enter a name for the API key"
            $scope = Read-Host "Enter scope (comma-separated, default: read,reserve)"
            if (-not $scope) { $scope = "read,reserve" }
            
            New-ApiKey -Name $name -Scope $scope
        }
        "13" { 
            if ($templateId -eq 0) {
                $templateId = Read-Host "Enter template ID to delete"
            }
            Remove-Template -Id $templateId
        }
        "14" { 
            if ($hostnameId -eq 0) {
                $hostnameId = Read-Host "Enter hostname ID to delete"
            }
            $forceDelete = Read-Host "Force delete (attempt to release first)? (y/n)"
            if ($forceDelete -eq 'y') {
                Remove-Hostname -Id $hostnameId -Force
            } else {
                Remove-Hostname -Id $hostnameId
            }
        }
        "15" { 
            if ($apiKeyId -eq 0) {
                $apiKeyId = Read-Host "Enter API key ID to delete"
            }
            Remove-ApiKey -Id $apiKeyId
        }
        default { Write-Color "Invalid choice, try again" "Red" }
    }
    
    Write-Color "`nPress any key to continue..." "Yellow"
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}