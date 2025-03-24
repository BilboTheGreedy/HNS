# Demo-TemplateHostname.ps1
# Demonstrates creating templates and generating hostnames from them
# With improved validation and user input for parameter values

param(
    # Optional parameters
    [Parameter(Mandatory = $false)]
    [string]$HnsUrl = "http://localhost:8080",
    
    [Parameter(Mandatory = $false)]
    [string]$Username = "admin",
    
    [Parameter(Mandatory = $false)]
    [string]$Password = "admin123",
    
    [Parameter(Mandatory = $false)]
    [switch]$CreateNewTemplate,
    
    [Parameter(Mandatory = $false)]
    [string]$TemplateName = "Demo-Template-" + (Get-Random -Minimum 1000 -Maximum 9999)
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

# Helper function to pause for a key press
function Pause-Script {
    param([string]$Message = "Press any key to continue...")
    Write-Color $Message "Yellow" -NoNewline
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    Write-Host ""
}

# Helper function to get user selection from a list
function Get-UserSelection {
    param(
        [string]$Prompt,
        [array]$Options,
        [string]$DefaultOption = ""
    )
    
    Write-Color $Prompt "Yellow"
    
    for ($i = 0; $i -lt $Options.Count; $i++) {
        if ($Options[$i] -eq $DefaultOption) {
            Write-Color "  $($i+1). $($Options[$i]) (default)" "Green"
        } else {
            Write-Color "  $($i+1). $($Options[$i])" "White"
        }
    }
    
    $defaultIndex = $Options.IndexOf($DefaultOption) + 1
    if ($defaultIndex -gt 0) {
        Write-Color "Enter selection (1-$($Options.Count)) [default=$defaultIndex]: " "Yellow" -NoNewline
    } else {
        Write-Color "Enter selection (1-$($Options.Count)): " "Yellow" -NoNewline
    }
    
    $selection = Read-Host
    
    # Use default if empty input and default is provided
    if ([string]::IsNullOrEmpty($selection) -and $defaultIndex -gt 0) {
        $selection = $defaultIndex
    }
    
    $index = [int]$selection - 1
    
    if ($index -ge 0 -and $index -lt $Options.Count) {
        return $Options[$index]
    } else {
        Write-Color "Invalid selection. Please try again." "Red"
        return Get-UserSelection -Prompt $Prompt -Options $Options -DefaultOption $DefaultOption
    }
}

# Helper function to parse allowed values from validation_value
function Get-AllowedValues {
    param([string]$ValidationValue)
    
    if ([string]::IsNullOrEmpty($ValidationValue)) {
        return @()
    }
    
    return $ValidationValue.Split(',') | ForEach-Object { $_.Trim() }
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
    exit 1
}

# 2. List existing templates or create a new one
if ($CreateNewTemplate) {
    Write-Color "`nCreating a new template: $TemplateName" "Cyan"
    
    # Define template groups - using 'dc', 'env', 'app' as expected by server
    $groups = @(
        @{
            name = "dc"      # Datacenter
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "NY,LA,SF,DC,CH,AM"
        },
        @{
            name = "env"     # Environment
            length = 3
            is_required = $true
            validation_type = "list"
            validation_value = "DEV,TST,STG,PRD"
        },
        @{
            name = "app"     # Application/Role
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "WS,DB,AP,LB"
        },
        @{
            name = "seq"    # Sequence
            length = 3
            is_required = $true
            validation_type = "sequence"
            validation_value = ""
        }
    )
    
    # Create the template request body
    $templateBody = @{
        name = $TemplateName
        description = "Demo template created by PowerShell script"
        max_length = 15
        sequence_start = 1
        sequence_length = 3
        sequence_padding = $true
        sequence_increment = 1
        created_by = "PowerShell-Demo"
        groups = $groups
    } | ConvertTo-Json -Depth 5
    
    try {
        $template = Invoke-RestMethod -Uri "$HnsUrl/api/templates" -Method POST -Body $templateBody -Headers $Headers
        Write-Color "Template created successfully with ID: $($template.id)" "Green"
        $templateId = $template.id
    }
    catch {
        Write-Color "Failed to create template: $_" "Red"
        exit 1
    }
}
else {
    # List existing templates
    Write-Color "`nListing existing templates..." "Cyan"
    
    try {
        $response = Invoke-RestMethod -Uri "$HnsUrl/api/templates?limit=10" -Method GET -Headers $Headers
        
        if ($response.templates.Count -eq 0) {
            Write-Color "No templates found. Creating a new template instead." "Yellow"
            $CreateNewTemplate = $true
            # Re-run this part of the script recursively with CreateNewTemplate set
            & $PSCommandPath -HnsUrl $HnsUrl -Username $Username -Password $Password -CreateNewTemplate
            exit
        }
        
        Write-Color "Available templates:" "Green"
        for ($i = 0; $i -lt $response.templates.Count; $i++) {
            Write-Color "  $($i+1). $($response.templates[$i].name) (ID: $($response.templates[$i].id))" "White"
        }
        
        Write-Color "`nEnter the number of the template to use (1-$($response.templates.Count)) or 0 to create a new one: " "Yellow" -NoNewline
        $selection = Read-Host
        
        if ($selection -eq "0") {
            # Create a new template
            $CreateNewTemplate = $true
            # Re-run this part of the script recursively with CreateNewTemplate set
            & $PSCommandPath -HnsUrl $HnsUrl -Username $Username -Password $Password -CreateNewTemplate
            exit
        }
        
        $index = [int]$selection - 1
        
        if ($index -ge 0 -and $index -lt $response.templates.Count) {
            $templateId = $response.templates[$index].id
            $TemplateName = $response.templates[$index].name
            Write-Color "Selected template: $TemplateName (ID: $templateId)" "Cyan"
        }
        else {
            Write-Color "Invalid selection. Using the first template." "Yellow"
            $templateId = $response.templates[0].id
            $TemplateName = $response.templates[0].name
            Write-Color "Using template: $TemplateName (ID: $templateId)" "Cyan"
        }
    }
    catch {
        Write-Color "Failed to get templates: $_" "Red"
        exit 1
    }
}

# 3. Get template details to understand its groups
Write-Color "`nGetting template details..." "Cyan"

try {
    $templateDetails = Invoke-RestMethod -Uri "$HnsUrl/api/templates/$templateId" -Method GET -Headers $Headers
    
    Write-Color "Template details:" "Green"
    Write-Color "  Name: $($templateDetails.name)" "White"
    Write-Color "  Description: $($templateDetails.description)" "White"
    Write-Color "  Max Length: $($templateDetails.max_length)" "White"
    Write-Color "  Sequence Start: $($templateDetails.sequence_start)" "White"
    
    Write-Color "`nTemplate groups:" "Green"
    
    # Store group info for later use
    $templateGroups = @{}
    
    foreach ($group in $templateDetails.groups) {
        if ($group.validation_type -eq "sequence") {
            Write-Color "  - $($group.name): Type=$($group.validation_type), Length=$($group.length), Position=$($group.position)" "White"
        }
        else {
            $allowedValues = Get-AllowedValues -ValidationValue $group.validation_value
            $valuesDisplay = $allowedValues -join ', '
            Write-Color "  - $($group.name): Type=$($group.validation_type), Allowed Values=[$valuesDisplay]" "White"
            
            # Store group info
            $templateGroups[$group.name] = @{
                validation_type = $group.validation_type
                validation_value = $group.validation_value
                allowed_values = $allowedValues
                is_required = $group.is_required
            }
        }
    }
    
    if ($templateGroups.Count -eq 0) {
        Write-Color "`nWARNING: No groups with validation found in template." "Red"
        exit 1
    }
}
catch {
    Write-Color "Failed to get template details: $_" "Red"
    exit 1
}

Pause-Script

# 4. Generate hostnames from the template without reserving them
Write-Color "`nGenerating hostnames from the template without reserving..." "Cyan"

# Create interactive parameter sets with the user selecting values from each group's allowed values
$parameterSets = @()

# Get 3 parameter sets from user
for ($setIndex = 0; $setIndex -lt 3; $setIndex++) {
    Write-Color "`nParameter Set #$($setIndex+1):" "Magenta"
    
    $params = @{}
    
    # For each non-sequence group, prompt for a value
    foreach ($groupName in $templateGroups.Keys) {
        $group = $templateGroups[$groupName]
        
        if ($group.validation_type -eq "list" -and $group.allowed_values.Count -gt 0) {
            # Have the user select a value from the allowed list
            $selectedValue = Get-UserSelection -Prompt "Select a value for '$groupName':" -Options $group.allowed_values
            $params[$groupName] = $selectedValue
        }
        elseif ($group.validation_type -eq "fixed") {
            # Fixed value, just use it
            $params[$groupName] = $group.validation_value
            Write-Color "Using fixed value for '$groupName': $($group.validation_value)" "Yellow"
        }
        elseif ($group.is_required) {
            # Required but no validation or empty validation list
            Write-Color "Enter a value for '$groupName': " "Yellow" -NoNewline
            $params[$groupName] = Read-Host
        }
    }
    
    $parameterSets += $params
    
    Write-Color "Parameter set #$($setIndex+1) configured:" "Green"
    foreach ($key in $params.Keys) {
        Write-Color "  $key = $($params[$key])" "White"
    }
}

$generatedHostnames = @()

foreach ($params in $parameterSets) {
    # Display the parameters we're using
    Write-Color "`nGenerating hostname with parameters:" "Yellow"
    foreach ($key in $params.Keys) {
        Write-Color "  $key = $($params[$key])" "White"
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
                Write-Color "  DNS check: Already exists with IP $($generatedResult.dns_check.ip_address)" "Yellow"
            } else {
                Write-Color "  DNS check: Available" "Green"
            }
        }
        
        $generatedHostnames += $generatedResult.hostname
    }
    catch {
        Write-Color "Failed to generate hostname: $_" "Red"
        
        # Parse error message for validation issues
        $errorDetails = $null
        try {
            $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json -ErrorAction SilentlyContinue
        } catch {}
        
        if ($errorDetails -and $errorDetails.error) {
            Write-Color "Error details: $($errorDetails.error)" "Red"
            
            # If error message contains information about invalid values
            if ($errorDetails.error -like "*invalid*" -or $errorDetails.error -like "*not in allowed list*") {
                Write-Color "This appears to be a validation error. Please check the allowed values for each group." "Yellow"
                Write-Color "Allowed values for template groups:" "Yellow"
                
                foreach ($groupKey in $templateGroups.Keys) {
                    $group = $templateGroups[$groupKey]
                    if ($group.validation_type -eq "list") {
                        $valuesDisplay = $group.allowed_values -join ', '
                        Write-Color "  Group $($groupKey): [$valuesDisplay]" "White"
                    }
                }
            }
        }
    }
}

Pause-Script

# 5. Reserve one of the generated hostnames
Write-Color "`nReserving a hostname..." "Cyan"

# Ask which parameter set to use
Write-Color "Which parameter set would you like to use for reservation?" "Yellow"
for ($i = 0; $i -lt $parameterSets.Count; $i++) {
    Write-Color "  $($i+1). Parameter Set #$($i+1):" "White"
    foreach ($key in $parameterSets[$i].Keys) {
        Write-Color "     $key = $($parameterSets[$i][$key])" "Gray"
    }
}

Write-Color "Enter selection (1-$($parameterSets.Count)): " "Yellow" -NoNewline
$selection = Read-Host
$index = [int]$selection - 1

if ($index -ge 0 -and $index -lt $parameterSets.Count) {
    $reserveParams = $parameterSets[$index]
} else {
    Write-Color "Invalid selection. Using the first parameter set." "Yellow"
    $reserveParams = $parameterSets[0]
}

$reserveBody = @{
    template_id = $templateId
    params = $reserveParams
    requested_by = "PowerShell-Demo"
    dns_check = $true
} | ConvertTo-Json

try {
    $hostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/reserve" -Method POST -Body $reserveBody -Headers $Headers
    
    Write-Color "Reserved hostname successfully:" "Green"
    Write-Color "  - Name: $($hostname.name)" "White"
    Write-Color "  - ID: $($hostname.id)" "White"
    Write-Color "  - Sequence: $($hostname.sequence_num)" "White"
    Write-Color "  - Status: $($hostname.status)" "White"
    
    $reservedHostnameId = $hostname.id
    $reservedHostnameName = $hostname.name
}
catch {
    Write-Color "Failed to reserve hostname: $_" "Red"
    
    # Parse error message for validation issues
    $errorDetails = $null
    try {
        $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json -ErrorAction SilentlyContinue
    } catch {}
    
    if ($errorDetails -and $errorDetails.error) {
        Write-Color "Error details: $($errorDetails.error)" "Red"
    }
}

Pause-Script

# 6. Generate hostname for a specific sequence number
Write-Color "`nGenerating hostname with a specific sequence number..." "Cyan"

# Ask which parameter set to use
Write-Color "Which parameter set would you like to use for specific sequence?" "Yellow"
for ($i = 0; $i -lt $parameterSets.Count; $i++) {
    Write-Color "  $($i+1). Parameter Set #$($i+1):" "White"
    foreach ($key in $parameterSets[$i].Keys) {
        Write-Color "     $key = $($parameterSets[$i][$key])" "Gray"
    }
}

Write-Color "Enter selection (1-$($parameterSets.Count)): " "Yellow" -NoNewline
$selection = Read-Host
$index = [int]$selection - 1

if ($index -ge 0 -and $index -lt $parameterSets.Count) {
    $specificParams = $parameterSets[$index]
} else {
    Write-Color "Invalid selection. Using the second parameter set." "Yellow"
    $specificParams = $parameterSets[1]
}

Write-Color "Enter a specific sequence number to use: " "Yellow" -NoNewline
$specificSeqInput = Read-Host
$specificSeq = 42 # Default value

if (-not [string]::IsNullOrEmpty($specificSeqInput) -and [int]::TryParse($specificSeqInput, [ref]$specificSeq)) {
    # Use the parsed value
} else {
    Write-Color "Invalid sequence number. Using default (42)." "Yellow"
}

$specificBody = @{
    template_id = $templateId
    sequence_num = $specificSeq
    params = $specificParams
} | ConvertTo-Json

try {
    $specificResult = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/generate" -Method POST -Body $specificBody -Headers $Headers
    
    # Fixed line with proper variable expansion
    Write-Color "Generated hostname with sequence $($specificSeq): $($specificResult.hostname)" "Green"
}
catch {
    Write-Color "Failed to generate hostname with specific sequence: $_" "Red"
    
    # Parse error message for validation issues
    $errorDetails = $null
    try {
        $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json -ErrorAction SilentlyContinue
    } catch {}
    
    if ($errorDetails -and $errorDetails.error) {
        Write-Color "Error details: $($errorDetails.error)" "Red"
    }
}

Pause-Script

# 7. Get the next available sequence number
Write-Color "`nGetting next available sequence number..." "Cyan"

try {
    $nextSeq = Invoke-RestMethod -Uri "$HnsUrl/api/sequences/next/$templateId" -Method GET -Headers $Headers
    
    Write-Color "Next available sequence number: $($nextSeq.sequence_num)" "Green"
}
catch {
    Write-Color "Failed to get next sequence number: $_" "Red"
}

Pause-Script

# 8. Clean up - ask if user wants to release and/or delete the reserved hostname
if ($reservedHostnameId) {
    Write-Color "`nDo you want to release the reserved hostname ($reservedHostnameName)? [y/N] " "Yellow" -NoNewline
    $releaseResponse = Read-Host
    
    if ($releaseResponse.ToLower() -eq "y") {
        # First commit the hostname (required before releasing)
        Write-Color "Committing hostname before release..." "Cyan"
        $commitBody = @{
            hostname_id = $reservedHostnameId
            committed_by = "PowerShell-Demo"
        } | ConvertTo-Json
        
        try {
            $committedHostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/commit" -Method POST -Body $commitBody -Headers $Headers
            Write-Color "Hostname committed successfully: $($committedHostname.name)" "Green"
            
            # Now release it
            $releaseBody = @{
                hostname_id = $reservedHostnameId
                released_by = "PowerShell-Demo"
            } | ConvertTo-Json
            
            $releasedHostname = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/release" -Method POST -Body $releaseBody -Headers $Headers
            Write-Color "Hostname released successfully: $($releasedHostname.name)" "Green"
            
            # Ask if user wants to delete the hostname
            Write-Color "`nDo you want to delete the hostname? [y/N] " "Yellow" -NoNewline
            $deleteResponse = Read-Host
            
            if ($deleteResponse.ToLower() -eq "y") {
                # Check if Delete-Hostname.ps1 exists in the current directory
                $deleteScriptPath = Join-Path -Path $PSScriptRoot -ChildPath "Delete-Hostname.ps1"
                $deleteScriptExists = Test-Path -Path $deleteScriptPath
                
                if ($deleteScriptExists) {
                    Write-Color "Running deletion script..." "Cyan"
                    & $deleteScriptPath -HostnameId $reservedHostnameId -SkipConfirmation -HnsUrl $HnsUrl -Username $Username -Password $Password
                }
                else {
                    # Delete directly using API
                    try {
                        $response = Invoke-RestMethod -Uri "$HnsUrl/api/hostnames/$reservedHostnameId" -Method DELETE -Headers $Headers
                        Write-Color "Hostname deleted successfully!" "Green"
                    }
                    catch {
                        Write-Color "Failed to delete hostname: $_" "Red"
                    }
                }
            }
        }
        catch {
            Write-Color "Failed to commit or release hostname: $_" "Red"
        }
    }
}

# 9. Summary
Write-Color "`n=== Template and Hostname Demo Summary ===" "Magenta"
Write-Color "Template: $TemplateName (ID: $templateId)" "White"
Write-Color "Generated Hostnames:" "White"

foreach ($name in $generatedHostnames) {
    Write-Color "  - $name" "White"
}

if ($CreateNewTemplate) {
    Write-Color "`nNote: A new template was created during this demo." "Yellow"
    Write-Color "You can reuse this template in future demos by selecting it from the list." "Yellow"
}

Write-Color "`nDemo completed successfully!" "Green"