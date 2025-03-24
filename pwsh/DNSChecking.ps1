# Demo-DNSChecking.ps1
# Demonstrates the DNS checking and sequence management features
# Then prompts to run the deletion script

# Initialize connection
$HnsUrl = "http://localhost:8080"
$Username = "admin"
$Password = "admin123"

# Helper function for colorful output
function Write-Color {
    param([string]$Text, [string]$Color = "White", [switch]$NoNewline)
    if ($NoNewline) {
        Write-Host $Text -ForegroundColor $Color -NoNewline
    } else {
        Write-Host $Text -ForegroundColor $Color
    }
}

# 1. Initialize connection and authenticate
Write-Color "Connecting to HNS API at $HnsUrl..." "Cyan"
$Headers = @{
    "Content-Type" = "application/json"
    "Accept" = "application/json"
}

$loginBody = @{
    username = $Username
    password = $Password
} | ConvertTo-Json

try {
    $auth = Invoke-RestMethod -Uri "$HnsUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $Headers["Authorization"] = "Bearer $($auth.token)"
    Write-Color "Successfully authenticated as $Username" "Green"
}
catch {
    Write-Color "Authentication failed: $_" "Red"
    exit
}

# 2. List templates to select one
Write-Color "`nListing available templates..." "Cyan"

try {
    $response = Invoke-RestMethod -Uri "$HnsUrl/api/templates?limit=5" -Method GET -Headers $Headers
    
    if ($response.templates.Count -eq 0) {
        Write-Color "No templates found. Please create a template first." "Yellow"
        exit
    }
    
    Write-Color "Available templates:" "Green"
    foreach ($template in $response.templates) {
        Write-Color "  - ID: $($template.id), Name: $($template.name)" "White"
    }
    
    # Use the first template
    $templateId = $response.templates[0].id
    Write-Color "Using template ID: $templateId" "Cyan"
}
catch {
    Write-Color "Failed to get templates: $_" "Red"
    exit
}

# 3. Generate a hostname with DNS checking
Write-Color "`nGenerating a hostname with DNS checking..." "Cyan"

$params = @{
    # Dynamically create parameters based on the template groups
    # This is a simplified example - in a real script, you would get the template details 
    # and create parameters based on its groups
    "location" = "NYC"
    "type" = "WS"
    "environment" = "DEV"
}

$generateBody = @{
    template_id = $templateId
    params = $params
    dns_check = $true  # Enable DNS checking
} | ConvertTo-Json

try {
    $generatedResult = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/generate" -Method POST -Body $generateBody -Headers $Headers
    
    Write-Color "Generated hostname: $($generatedResult.hostname)" "Green"
    
    if ($generatedResult.dns_check) {
        if ($generatedResult.dns_check.exists) {
            Write-Color "WARNING: This hostname already exists in DNS with IP: $($generatedResult.dns_check.ip_address)" "Yellow"
        } else {
            Write-Color "Hostname is available in DNS" "Green"
        }
    }
}
catch {
    Write-Color "Failed to generate hostname: $_" "Red"
}

# 4. Reserve a hostname with automatic sequence management and DNS checking
Write-Color "`nReserving a hostname with automatic sequence management..." "Cyan"

$reserveBody = @{
    template_id = $templateId
    params = $params
    requested_by = "PowerShell-Demo"
    dns_check = $true
    auto_increment = $true  # Enable automatic sequence increment if DNS conflict found
    max_attempts = 5  # Try up to 5 different sequence numbers
} | ConvertTo-Json

try {
    $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/reserve" -Method POST -Body $reserveBody -Headers $Headers
    
    Write-Color "Reserved hostname successfully:" "Green"
    Write-Color "  - Name: $($hostname.name)" "White"
    Write-Color "  - ID: $($hostname.id)" "White"
    Write-Color "  - Sequence: $($hostname.sequence_num)" "White"
    Write-Color "  - Status: $($hostname.status)" "White"
    
    if ($hostname.dns_verified) {
        Write-Color "  - DNS Verified: Yes" "Green"
    } else {
        Write-Color "  - DNS Verified: No" "Yellow"
    }
    
    # Store hostname ID for later use
    $hostnameId = $hostname.id
    $hostnameNameVal = $hostname.name
}
catch {
    Write-Color "Failed to reserve hostname: $_" "Red"
    exit
}

# 5. Get the next sequence number instead of usage statistics
Write-Color "`nGetting next sequence number for template $templateId..." "Cyan"

try {
    $nextSeq = Invoke-RestMethod -Uri "$HnsUrl/api/sequences/next/$templateId" -Method GET -Headers $Headers
    
    Write-Color "Next sequence number: $($nextSeq.sequence_num)" "Green"
}
catch {
    Write-Color "Failed to get next sequence number: $_" "Red"
}

# 6. First commit the hostname before releasing it
Write-Color "`nCommitting hostname..." "Cyan"

$commitBody = @{
    hostname_id = $hostnameId
    committed_by = "PowerShell-Demo"
} | ConvertTo-Json

try {
    $committedHostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/commit" -Method POST -Body $commitBody -Headers $Headers
    Write-Color "Hostname committed successfully: $($committedHostname.name)" "Green"
}
catch {
    Write-Color "Failed to commit hostname: $_" "Red"
    exit
}

# 7. Now release the committed hostname
Write-Color "`nReleasing hostname..." "Cyan"

$releaseBody = @{
    hostname_id = $hostnameId
    released_by = "PowerShell-Demo"
} | ConvertTo-Json

try {
    $releasedHostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/release" -Method POST -Body $releaseBody -Headers $Headers
    Write-Color "Hostname released successfully: $($releasedHostname.name)" "Green"
}
catch {
    Write-Color "Failed to release hostname: $_" "Red"
}

Write-Color "`nDNS checking demo completed successfully!" "Magenta"

# 8. Check if Delete-Hostname.ps1 exists in the current directory
$deleteScriptPath = Join-Path -Path $PSScriptRoot -ChildPath "Delete-Hostname.ps1"
$deleteScriptExists = Test-Path -Path $deleteScriptPath

# 9. Prompt to run deletion script
Write-Color "`nWould you like to delete the hostname $hostnameNameVal (ID: $hostnameId)? [y/N] " "Yellow" -NoNewline
$response = Read-Host

if ($response.ToLower() -eq "y") {
    if ($deleteScriptExists) {
        Write-Color "Running deletion script..." "Cyan"
        
        # Invoke the deletion script with the hostname ID and skip confirmation since we just confirmed
        & $deleteScriptPath -HostnameId $hostnameId -SkipConfirmation -HnsUrl $HnsUrl -Username $Username -Password $Password
    }
    else {
        Write-Color "Delete-Hostname.ps1 script not found in the current directory." "Red"
        Write-Color "To delete the hostname manually, run:" "Yellow"
        Write-Color "  Invoke-RestMethod -Uri `"$HnsUrl/api/hostnames/$hostnameId`" -Method DELETE -Headers `$Headers" "Gray"
        
        # Ask if user wants to delete directly
        Write-Color "Would you like to delete the hostname directly? [y/N] " "Yellow" -NoNewline
        $directDelete = Read-Host
        
        if ($directDelete.ToLower() -eq "y") {
            try {
                $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$hostnameId" -Method DELETE -Headers $Headers
                Write-Color "Hostname deleted successfully!" "Green"
            }
            catch {
                Write-Color "Failed to delete hostname: $_" "Red"
            }
        }
    }
}
else {
    Write-Color "Deletion skipped. Hostname $hostnameNameVal (ID: $hostnameId) was not deleted." "Yellow"
}

Write-Color "`nScript completed!" "Magenta"