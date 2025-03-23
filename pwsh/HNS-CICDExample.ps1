# HNS-CICDExample.ps1
# Example of HNS API integration with CI/CD pipelines

# Import the module
Import-Module -Name '.\HNS-API.psm1' -Force

# CI/CD pipeline parameters (these would typically come from the CI/CD system)
param(
    [Parameter(Mandatory = $false)]
    [string]$Environment = "DEV",  # DEV, TST, STG, PRD
    
    [Parameter(Mandatory = $false)]
    [string]$Location = "NY",      # Data center location
    
    [Parameter(Mandatory = $false)]
    [string]$AppType = "APP",      # Application type: WEB, APP, DB, etc.
    
    [Parameter(Mandatory = $false)]
    [string]$HnsApiUrl = "http://localhost:8080",
    
    [Parameter(Mandatory = $false)]
    [string]$HnsApiKey = "",
    
    [Parameter(Mandatory = $false)]
    [string]$TemplateId = "",
    
    [Parameter(Mandatory = $false)]
    [switch]$Force = $false,
    
    [Parameter(Mandatory = $false)]
    [switch]$SkipDnsCheck = $false
)

# Logging function
function Write-Log {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message,
        
        [Parameter(Mandatory = $false)]
        [ValidateSet("INFO", "WARNING", "ERROR", "SUCCESS")]
        [string]$Level = "INFO"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $colorMap = @{
        "INFO" = "White"
        "WARNING" = "Yellow"
        "ERROR" = "Red"
        "SUCCESS" = "Green"
    }
    
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $colorMap[$Level]
    
    # In a real CI/CD pipeline, you might also log to a file or external logging system
}

# Error handling
$ErrorActionPreference = "Stop"
trap {
    Write-Log "An error occurred: $_" -Level "ERROR"
    Write-Log "Stack Trace: $($_.ScriptStackTrace)" -Level "ERROR"
    exit 1
}

# Main function
function Main {
    Write-Log "Starting HNS integration for CI/CD pipeline" -Level "INFO"
    Write-Log "Environment: $Environment, Location: $Location, App Type: $AppType" -Level "INFO"
    
    # Initialize connection to HNS
    if (-not $HnsApiKey) {
        Write-Log "No API key provided. Attempting to use JWT authentication." -Level "WARNING"
        try {
            Initialize-HnsConnection -BaseUrl $HnsApiUrl
            # In a real pipeline, these would come from secure variables or a credential store
            $username = "admin"
            $password = ConvertTo-SecureString "admin123" -AsPlainText -Force
            Connect-HnsService -Username $username -Password $password
        }
        catch {
            Write-Log "Authentication failed: $_" -Level "ERROR"
            exit 1
        }
    }
    else {
        Write-Log "Using API key authentication" -Level "INFO"
        Initialize-HnsConnection -BaseUrl $HnsApiUrl -ApiKey $HnsApiKey
    }
    
    # Test connection
    if (-not (Test-HnsConnection)) {
        Write-Log "Failed to connect to HNS service" -Level "ERROR"
        exit 1
    }
    
    # Find or specify template
    if (-not $TemplateId) {
        Write-Log "No template ID specified. Searching for a suitable template..." -Level "INFO"
        $templates = Get-HnsTemplates -Limit 100
        
        # Find the first active template - in a real scenario, you'd use specific criteria
        $template = $templates | Where-Object { $_.is_active -eq $true } | Select-Object -First 1
        
        if (-not $template) {
            Write-Log "No active templates found. Please specify a template ID." -Level "ERROR"
            exit 1
        }
        
        $TemplateId = $template.id
        Write-Log "Using template: $($template.name) (ID: $TemplateId)" -Level "INFO"
    }
    else {
        # Verify the template exists and is active
        try {
            $template = Get-HnsTemplate -Id $TemplateId
            Write-Log "Using specified template: $($template.name)" -Level "INFO"
        }
        catch {
            Write-Log "Template with ID $TemplateId not found or inaccessible" -Level "ERROR"
            exit 1
        }
    }
    
    # Map environment names if needed (CI/CD system might use different naming conventions)
    $envMap = @{
        "Development" = "DEV"
        "Testing" = "TST"
        "Staging" = "STG"
        "Production" = "PRD"
    }
    
    if ($envMap.ContainsKey($Environment)) {
        $mappedEnv = $envMap[$Environment]
        Write-Log "Mapped environment '$Environment' to '$mappedEnv'" -Level "INFO"
        $Environment = $mappedEnv
    }
    
    # Set up hostname parameters based on template groups
    $params = @{}
    foreach ($group in $template.groups) {
        if ($group.validation_type -eq "sequence") {
            continue  # Skip sequence groups
        }
        
        # Match parameters to group names
        switch ($group.name.ToLower()) {
            "location" { $params[$group.name] = $Location }
            "loc" { $params[$group.name] = $Location }
            
            "environment" { $params[$group.name] = $Environment }
            "env" { $params[$group.name] = $Environment }
            
            "type" { $params[$group.name] = $AppType }
            "app" { $params[$group.name] = $AppType }
            "role" { $params[$group.name] = $AppType }
            
            default {
                # For other groups, try to use the first allowed value
                if ($group.validation_type -eq "list" -and $group.validation_value) {
                    $values = $group.validation_value.Split(",")
                    if ($values.Length -gt 0) {
                        $params[$group.name] = $values[0].Trim()
                        Write-Log "Using default value '$($values[0].Trim())' for parameter '$($group.name)'" -Level "WARNING"
                    }
                }
            }
        }
    }
    
    Write-Log "Using hostname parameters:" -Level "INFO"
    $params.GetEnumerator() | ForEach-Object {
        Write-Log "  $($_.Key): $($_.Value)" -Level "INFO"
    }
    
    # Generate a hostname without reserving it first
    try {
        $generatedHostname = New-HnsHostnameGeneration -TemplateId $TemplateId -Params $params
        $hostname = $generatedHostname.hostname
        Write-Log "Generated hostname: $hostname" -Level "SUCCESS"
    }
    catch {
        Write-Log "Failed to generate hostname: $_" -Level "ERROR"
        exit 1
    }
    
    # Check DNS if requested
    if (-not $SkipDnsCheck) {
        Write-Log "Checking if hostname exists in DNS..." -Level "INFO"
        try {
            $dnsResult = Test-HnsDnsHostname -Hostname $hostname
            
            if ($dnsResult.exists) {
                $message = "Hostname $hostname already exists in DNS with IP: $($dnsResult.ip_address)"
                if ($Force) {
                    Write-Log "$message. Continuing due to -Force flag." -Level "WARNING"
                }
                else {
                    Write-Log "$message. Use -Force to override." -Level "ERROR"
                    exit 1
                }
            }
            else {
                Write-Log "Hostname $hostname is not in use in DNS" -Level "SUCCESS"
            }
        }
        catch {
            Write-Log "DNS check failed: $_. Continuing with reservation." -Level "WARNING"
        }
    }
    
    # Reserve the hostname
    try {
        $reservedHostname = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params
        Write-Log "Hostname reserved: $($reservedHostname.name) (ID: $($reservedHostname.id))" -Level "SUCCESS"
        
        # Save hostname ID to pipeline variable or file for later stages
        $hostnameId = $reservedHostname.id
        $reservedHostname.id | Out-File -FilePath ".\hostname_id.txt"
        $reservedHostname.name | Out-File -FilePath ".\hostname.txt"
        
        Write-Log "Hostname ID saved to hostname_id.txt" -Level "INFO"
        Write-Log "Hostname saved to hostname.txt" -Level "INFO"
    }
    catch {
        Write-Log "Failed to reserve hostname: $_" -Level "ERROR"
        exit 1
    }
    
    # In a real CI/CD pipeline, you might want to wait for later stages
    # before committing the hostname
    Write-Log "To commit this hostname in a later stage, use:" -Level "INFO"
    Write-Log "Set-HnsHostnameCommit -HostnameId $hostnameId" -Level "INFO"
    
    return @{
        HostnameId = $hostnameId
        Hostname = $reservedHostname.name
        Status = $reservedHostname.status
    }
}

# Example of implementing a post-deployment step to commit the hostname
function Commit-DeployedHostname {
    param(
        [Parameter(Mandatory = $true)]
        [int]$HostnameId
    )
    
    Write-Log "Committing hostname ID $HostnameId after successful deployment" -Level "INFO"
    
    try {
        $committedHostname = Set-HnsHostnameCommit -HostnameId $HostnameId
        Write-Log "Hostname $($committedHostname.name) successfully committed" -Level "SUCCESS"
        return $committedHostname
    }
    catch {
        Write-Log "Failed to commit hostname: $_" -Level "ERROR"
        # In a real pipeline, you might not want to fail the deployment if committing fails
        # but instead create a task for manual follow-up
        return $null
    }
}

# Example of implementing a rollback step to release the hostname if deployment fails
function Release-HostnameOnFailure {
    param(
        [Parameter(Mandatory = $true)]
        [int]$HostnameId
    )
    
    Write-Log "Deployment failed - releasing hostname ID $HostnameId" -Level "WARNING"
    
    try {
        $releasedHostname = Set-HnsHostnameRelease -HostnameId $HostnameId
        Write-Log "Hostname $($releasedHostname.name) released due to deployment failure" -Level "SUCCESS"
        return $releasedHostname
    }
    catch {
        Write-Log "Failed to release hostname: $_" -Level "ERROR"
        return $null
    }
}

# Example of how this might be used in a multi-stage pipeline
function Example-MultiStagePipeline {
    # Stage 1: Reserve hostname
    $result = Main
    $hostnameId = $result.HostnameId
    $hostname = $result.Hostname
    
    # Pass hostname to the deployment process
    # ...
    
    # Stage 2: Deploy application
    $deploymentSuccess = $true  # This would be determined by your deployment process
    
    # Stage 3: Post-deployment - commit or release hostname based on deployment result
    if ($deploymentSuccess) {
        Commit-DeployedHostname -HostnameId $hostnameId
        Write-Log "Deployment of $hostname completed successfully" -Level "SUCCESS"
    }
    else {
        Release-HostnameOnFailure -HostnameId $hostnameId
        Write-Log "Deployment of $hostname failed - hostname released" -Level "ERROR"
    }
}

# Uncomment to run the entire multi-stage example
# Example-MultiStagePipeline

# Run just the hostname reservation part by default
$result = Main
Write-Log "Pipeline completed successfully" -Level "SUCCESS"
return $result