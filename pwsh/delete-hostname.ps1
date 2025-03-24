# Delete-Hostname.ps1
# Parameterized script for deleting hostnames from the HNS system

param(
    # You can specify either a hostname or a hostname ID
    [Parameter(Mandatory = $false, ParameterSetName = "ByName")]
    [string]$Hostname,
    
    [Parameter(Mandatory = $false, ParameterSetName = "ById")]
    [int]$HostnameId,
    
    # Optional parameters
    [Parameter(Mandatory = $false)]
    [string]$HnsUrl = "http://localhost:8080",
    
    [Parameter(Mandatory = $false)]
    [string]$Username = "admin",
    
    [Parameter(Mandatory = $false)]
    [string]$Password = "admin123",
    
    [Parameter(Mandatory = $false)]
    [string]$ApiKey = "",
    
    [Parameter(Mandatory = $false)]
    [switch]$Force,
    
    [Parameter(Mandatory = $false)]
    [switch]$SkipConfirmation
)

# Helper function for colorful output
function Write-Color {
    param(
        [string]$Text,
        [string]$Color = "White",
        [switch]$NoNewline
    )
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

# If API key is provided, use it instead of username/password
if ($ApiKey) {
    $Headers["X-API-Key"] = $ApiKey
    Write-Color "Using API key authentication" "Green"
}
else {
    # Authenticate with username/password
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
        exit 1
    }
}

# 2. Determine the hostname ID if provided with a hostname
if ($Hostname -and -not $HostnameId) {
    Write-Color "Looking up hostname: $Hostname..." "Cyan"
    
    try {
        # Use the search endpoint to find by name
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames?name=$Hostname" -Method GET -Headers $Headers
        
        if ($response.hostnames.Count -eq 0) {
            Write-Color "No hostname found with name: $Hostname" "Red"
            exit 1
        }
        elseif ($response.hostnames.Count -gt 1) {
            # Multiple matches found, let user select
            Write-Color "Multiple hostnames found matching '$Hostname':" "Yellow"
            for ($i = 0; $i -lt $response.hostnames.Count; $i++) {
                Write-Color "  $($i+1). $($response.hostnames[$i].name) (ID: $($response.hostnames[$i].id), Status: $($response.hostnames[$i].status))" "White"
            }
            
            Write-Color "Enter the number of the hostname to delete (1-$($response.hostnames.Count)): " "Yellow" -NoNewline
            $selection = Read-Host
            $index = [int]$selection - 1
            
            if ($index -ge 0 -and $index -lt $response.hostnames.Count) {
                $HostnameId = $response.hostnames[$index].id
                $Hostname = $response.hostnames[$index].name
                Write-Color "Selected: $Hostname (ID: $HostnameId)" "Cyan"
            }
            else {
                Write-Color "Invalid selection" "Red"
                exit 1
            }
        }
        else {
            # Single match found
            $HostnameId = $response.hostnames[0].id
            $Hostname = $response.hostnames[0].name
            Write-Color "Found hostname: $Hostname (ID: $HostnameId)" "Green"
        }
    }
    catch {
        Write-Color "Error looking up hostname: $_" "Red"
        exit 1
    }
}
elseif ($HostnameId -and -not $Hostname) {
    # If only ID is provided, get the hostname details
    try {
        $hostnameDetails = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$HostnameId" -Method GET -Headers $Headers
        $Hostname = $hostnameDetails.name
        Write-Color "Found hostname: $Hostname (ID: $HostnameId)" "Green"
    }
    catch {
        Write-Color "Error: Hostname with ID $HostnameId not found" "Red"
        exit 1
    }
}
elseif (-not $Hostname -and -not $HostnameId) {
    # If neither is provided, list hostnames and let user select
    Write-Color "No hostname specified. Listing recent hostnames..." "Cyan"
    
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames?limit=10" -Method GET -Headers $Headers
        
        if ($response.hostnames.Count -eq 0) {
            Write-Color "No hostnames found in the system" "Yellow"
            exit 0
        }
        
        Write-Color "Recent hostnames:" "Green"
        for ($i = 0; $i -lt $response.hostnames.Count; $i++) {
            Write-Color "  $($i+1). $($response.hostnames[$i].name) (ID: $($response.hostnames[$i].id), Status: $($response.hostnames[$i].status))" "White"
        }
        
        Write-Color "Enter the number of the hostname to delete (1-$($response.hostnames.Count)) or 0 to exit: " "Yellow" -NoNewline
        $selection = Read-Host
        
        if ($selection -eq "0") {
            Write-Color "Operation cancelled" "Yellow"
            exit 0
        }
        
        $index = [int]$selection - 1
        
        if ($index -ge 0 -and $index -lt $response.hostnames.Count) {
            $HostnameId = $response.hostnames[$index].id
            $Hostname = $response.hostnames[$index].name
            Write-Color "Selected: $Hostname (ID: $HostnameId)" "Cyan"
        }
        else {
            Write-Color "Invalid selection" "Red"
            exit 1
        }
    }
    catch {
        Write-Color "Error listing hostnames: $_" "Red"
        exit 1
    }
}

# 3. Get the hostname status to determine if it needs to be released first
try {
    $hostnameDetails = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$HostnameId" -Method GET -Headers $Headers
    $status = $hostnameDetails.status
    
    Write-Color "Hostname details:" "Cyan"
    Write-Color "  - Name: $($hostnameDetails.name)" "White"
    Write-Color "  - ID: $($hostnameDetails.id)" "White"
    Write-Color "  - Status: $($hostnameDetails.status)" "White"
    Write-Color "  - Reserved by: $($hostnameDetails.reserved_by)" "White"
    
    # Check if hostname is in a committed state and -Force was specified
    if ($status -eq "committed" -and $Force) {
        Write-Color "Hostname is committed and -Force was specified. Will attempt to release first." "Yellow"
        $needsRelease = $true
    }
    elseif ($status -eq "committed" -and -not $Force) {
        Write-Color "ERROR: Cannot delete a committed hostname without using -Force" "Red"
        Write-Color "Use -Force to automatically release the hostname before deletion" "Yellow"
        exit 1
    }
    else {
        $needsRelease = $false
    }
}
catch {
    Write-Color "Error getting hostname details: $_" "Red"
    exit 1
}

# 4. Ask for confirmation unless -SkipConfirmation is specified
if (-not $SkipConfirmation) {
    Write-Color "Are you sure you want to delete hostname '$Hostname' (ID: $HostnameId)? [y/N] " "Yellow" -NoNewline
    $confirmation = Read-Host
    
    if ($confirmation.ToLower() -ne "y") {
        Write-Color "Operation cancelled" "Yellow"
        exit 0
    }
}

# 5. Handle hostname in committed state if needed
if ($needsRelease) {
    Write-Color "Releasing hostname before deletion..." "Cyan"
    
    $releaseBody = @{
        hostname_id = $HostnameId
        released_by = "Delete-Script"
    } | ConvertTo-Json
    
    try {
        $releasedHostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/release" -Method POST -Body $releaseBody -Headers $Headers
        Write-Color "Hostname released successfully" "Green"
    }
    catch {
        Write-Color "Failed to release hostname: $_" "Red"
        Write-Color "Cannot proceed with deletion" "Red"
        exit 1
    }
}

# 6. Delete the hostname
Write-Color "Deleting hostname '$Hostname' (ID: $HostnameId)..." "Cyan"

try {
    # Send DELETE request to remove the hostname
    $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$HostnameId" -Method DELETE -Headers $Headers
    Write-Color "Hostname deleted successfully!" "Green"
}
catch {
    # Handle specific error cases
    if ($_.Exception.Response.StatusCode -eq 409) {
        Write-Color "Cannot delete hostname: It's currently in use or has dependencies" "Red"
        
        # Try to get error details
        try {
            $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Color "Error: $($errorDetails.error)" "Red"
            Write-Color "Message: $($errorDetails.message)" "Red"
        }
        catch {
            # If we can't parse error details, show the raw error
            Write-Color "Error: $_" "Red"
        }
        exit 1
    }
    else {
        Write-Color "Failed to delete hostname: $_" "Red"
        exit 1
    }
}

# 7. Verify deletion
Write-Color "Verifying deletion..." "Cyan"

try {
    $check = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$HostnameId" -Method GET -Headers $Headers -ErrorAction SilentlyContinue
    Write-Color "WARNING: Hostname still exists! Deletion may have failed." "Yellow"
    exit 1
}
catch {
    # If we get a 404, the hostname was successfully deleted
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Color "Verification successful - hostname no longer exists" "Green"
    }
    else {
        Write-Color "Error during verification: $_" "Red"
        exit 1
    }
}

Write-Color "Hostname deletion completed successfully" "Green"
exit 0