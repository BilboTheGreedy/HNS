# Demo-ManualHostname.ps1
# Demonstrates creating hostnames without a template

# Initialize connection
$HnsUrl = "http://localhost:8080"
$Username = "admin"
$Password = "admin123"

# Helper function for colorful output
function Write-Color {
    param([string]$Text, [string]$Color = "White")
    Write-Host $Text -ForegroundColor $Color
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

# 2. Create a manual hostname (new feature)
Write-Color "`nCreating a manual hostname..." "Cyan"

$Components = @{
    "datacenter" = "NYC"
    "role" = "DB"
    "environment" = "PROD"
    "number" = "001"
}

$manualHostnameBody = @{
    name = "NYC-DB-PROD-001"
    is_manual = $true
    status = "reserved"
    requested_by = "PowerShell-Demo"
    components = $Components
    dns_check = $true
} | ConvertTo-Json

try {
    $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/manual" -Method POST -Body $manualHostnameBody -Headers $Headers
    
    Write-Color "Manual hostname created successfully:" "Green"
    Write-Color "  - ID: $($hostname.id)" "White"
    Write-Color "  - Name: $($hostname.name)" "White"
    Write-Color "  - Status: $($hostname.status)" "White"
    Write-Color "  - Components:" "White"
    
    foreach ($component in $hostname.components.PSObject.Properties) {
        Write-Color "      $($component.Name): $($component.Value)" "Gray"
    }
    
    # Store hostname ID for later use
    $hostnameId = $hostname.id
}
catch {
    Write-Color "Failed to create manual hostname: $_" "Red"
    exit
}

# 3. Check DNS availability (can also be integrated during creation)
Write-Color "`nChecking if hostname is available in DNS..." "Cyan"

try {
    $dnsResult = Invoke-RestMethod -Uri "$HnsUrl/api/dns/check/$($hostname.name)" -Method GET -Headers $Headers
    
    if ($dnsResult.exists) {
        Write-Color "WARNING: Hostname already exists in DNS with IP: $($dnsResult.ip_address)" "Yellow"
    } else {
        Write-Color "Hostname is available in DNS" "Green"
    }
}
catch {
    Write-Color "Failed to check DNS: $_" "Red"
}

# 4. Commit the hostname
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
}

# 5. Release the hostname when done
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

Write-Color "`nDemo completed successfully!" "Magenta"