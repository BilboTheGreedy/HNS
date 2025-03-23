# HNS PowerShell Module Documentation

## Overview

The HNS (Hostname Naming System) PowerShell module provides a comprehensive set of cmdlets for interacting with the HNS API service. The module enables you to manage hostname templates, generate and reserve hostnames, check DNS, and manage API keys - all from PowerShell.

## Installation

1. Download the module file (`HNS-API.psm1`)
2. Import the module:

```powershell
Import-Module -Path '.\HNS-API.psm1'
```

## Getting Started

### Initialize Connection

Before using any other cmdlets, you need to initialize the connection to the HNS server:

```powershell
# With API key authentication
Initialize-HnsConnection -BaseUrl "http://hns-server:8080" -ApiKey "your-api-key"

# Or with JWT authentication
Initialize-HnsConnection -BaseUrl "http://hns-server:8080"
Connect-HnsService -Username "admin" -Password (ConvertTo-SecureString "password" -AsPlainText -Force)
```

### Test Connection

```powershell
Test-HnsConnection
```

## Managing Templates

Templates define the structure and rules for generating hostnames.

### List Templates

```powershell
# Get all templates with pagination
Get-HnsTemplates -Limit 10 -Offset 0
```

### Get Template Details

```powershell
# Get details for a specific template
Get-HnsTemplate -Id 1
```

### Create Template

```powershell
# Define template groups
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

# Create the template
New-HnsTemplate -Name "Server Template" -Description "For server hostnames" `
    -MaxLength 15 -SequenceStart 1 -SequenceLength 3 -SequencePadding $true `
    -SequenceIncrement 1 -Groups $groups
```

## Managing Hostnames

### Generate Hostname

Generate a hostname without reserving it:

```powershell
# Generate a hostname with specific parameters
New-HnsHostnameGeneration -TemplateId 1 -Params @{
    location = "NY"
    environment = "DEV"
}

# Generate with a specific sequence number
New-HnsHostnameGeneration -TemplateId 1 -SequenceNum 42 -Params @{
    location = "LA"
    environment = "PRD"
}

# Generate and check DNS
New-HnsHostnameGeneration -TemplateId 1 -Params @{
    location = "SF"
    environment = "TST"
} -CheckDns
```

### Reserve Hostname

Reserve a hostname for use:

```powershell
New-HnsHostnameReservation -TemplateId 1 -Params @{
    location = "DC"
    environment = "STG"
}
```

### List Hostnames

```powershell
# List all reserved hostnames
Get-HnsHostnames -Status "reserved"

# List hostnames for a specific template
Get-HnsHostnames -TemplateId 1

# List hostnames with name filtering
Get-HnsHostnames -Name "web"

# List hostnames reserved by a specific user
Get-HnsHostnames -ReservedBy "admin"

# Combined filtering with pagination
Get-HnsHostnames -Status "committed" -TemplateId 2 -Limit 5 -Offset 10
```

### Get Hostname Details

```powershell
Get-HnsHostname -Id 123
```

### Commit Hostname

Change a hostname from reserved to committed status:

```powershell
Set-HnsHostnameCommit -HostnameId 123
```

### Release Hostname

Change a hostname from committed to released status:

```powershell
Set-HnsHostnameRelease -HostnameId 123
```

## DNS Operations

### Check Hostname in DNS

```powershell
Test-HnsDnsHostname -Hostname "server1.example.com"
```

### Scan DNS for Hostname Range

```powershell
# Scan a range of hostnames to see if they exist in DNS
Start-HnsDnsScan -TemplateId 1 -StartSeq 1 -EndSeq 100 -Params @{
    location = "NY"
    environment = "DEV"
} -MaxConcurrent 10
```

## Sequence Management

### Get Next Sequence Number

```powershell
# Get the next available sequence number for a template
Get-HnsNextSequenceNumber -TemplateId 1
```

## API Key Management

### List API Keys

```powershell
Get-HnsApiKeys
```

### Create API Key

```powershell
# Create API key with specific permissions
New-HnsApiKey -Name "Automation Key" -Scope "read,reserve,commit"
```

### Delete API Key

```powershell
Remove-HnsApiKey -Id 5
```

## Example Workflow

Here's a complete example workflow for creating a template and reserving hostnames:

```powershell
# Initialize connection
Initialize-HnsConnection -BaseUrl "http://localhost:8080"
Connect-HnsService -Username "admin" -Password (ConvertTo-SecureString "admin123" -AsPlainText -Force)

# Create a template
$groups = @(
    @{ name = "loc"; length = 2; is_required = $true; validation_type = "list"; validation_value = "NY,LA,SF,DC" },
    @{ name = "app"; length = 3; is_required = $true; validation_type = "list"; validation_value = "WEB,APP,DB" },
    @{ name = "env"; length = 3; is_required = $true; validation_type = "list"; validation_value = "DEV,TST,PRD" },
    @{ name = "seq"; length = 3; is_required = $true; validation_type = "sequence"; validation_value = "" }
)

$template = New-HnsTemplate -Name "WebServer" -Description "Web server naming convention" `
    -MaxLength 15 -SequenceStart 1 -SequenceLength 3 -SequencePadding $true -SequenceIncrement 1 -Groups $groups

# Reserve a hostname
$hostname = New-HnsHostnameReservation -TemplateId $template.id -Params @{
    loc = "NY"
    app = "WEB"
    env = "DEV"
}

Write-Output "Reserved hostname: $($hostname.name) with ID: $($hostname.id)"

# Check DNS
$dnsCheck = Test-HnsDnsHostname -Hostname $hostname.name

# Commit the hostname once deployment is complete
Set-HnsHostnameCommit -HostnameId $hostname.id

# If later decommissioned, release the hostname
Set-HnsHostnameRelease -HostnameId $hostname.id
```

## CI/CD Integration

The HNS PowerShell module is ideal for integration with CI/CD pipelines. Here's a basic example:

```powershell
# In your deployment pipeline script
param(
    [string]$Environment,   # DEV, TST, STG, PRD
    [string]$Location,      # Data center code
    [string]$ServerType     # Application type
)

# Initialize with API key stored in CI/CD system
Initialize-HnsConnection -BaseUrl $env:HNS_API_URL -ApiKey $env:HNS_API_KEY

# Reserve a hostname for the new server
$hostname = New-HnsHostnameReservation -TemplateId 1 -Params @{
    location = $Location
    type = $ServerType
    environment = $Environment
}

# Return the hostname for use in further deployment steps
Write-Output "HOSTNAME=$($hostname.name)"

# Later, after successful deployment, commit the hostname
Set-HnsHostnameCommit -HostnameId $hostname.id
```

## Bulk Operations

For bulk operations, you can combine the cmdlets with PowerShell loops:

```powershell
# Reserve multiple hostnames
$reservedHostnames = @()
foreach ($location in @("NY", "LA", "SF")) {
    foreach ($env in @("DEV", "TST")) {
        $hostname = New-HnsHostnameReservation -TemplateId 1 -Params @{
            location = $location
            environment = $env
        }
        $reservedHostnames += $hostname
    }
}

# Commit all hostnames
foreach ($hostname in $reservedHostnames) {
    Set-HnsHostnameCommit -HostnameId $hostname.id
}
```

## Error Handling

The module throws errors for unsuccessful API calls. Use standard PowerShell error handling:

```powershell
try {
    $hostname = New-HnsHostnameReservation -TemplateId 1 -Params @{
        location = "NY"
        environment = "DEV"
    }
    # Success
    Write-Output "Reserved hostname: $($hostname.name)"
}
catch {
    # Error handling
    Write-Error "Failed to reserve hostname: $_"
}
```

## Best Practices

1. **Use API Keys** - For scripts and automation, create API keys with appropriate scopes instead of using credentials.

2. **Error Handling** - Always include try/catch blocks for proper error handling.

3. **Check DNS Before Reserving** - Use `Test-HnsDnsHostname` before reserving to avoid conflicts.

4. **Release Unused Hostnames** - Always release hostnames that are no longer needed.

5. **Template Documentation** - Document your templates, including their group structure and validation rules.

6. **Filters and Pagination** - Use filters and pagination to limit the data returned by list operations.

7. **Bulk Operations** - For bulk operations, consider batching requests to avoid overwhelming the server.

## Troubleshooting

### Authentication Issues

If you encounter authentication issues:

1. Check that your API key is valid and has the required scopes
2. Verify the API key hasn't expired
3. If using JWT auth, ensure your credentials are correct

### Connection Issues

For connection problems:

1. Verify the base URL is correct and the HNS service is running
2. Check network connectivity and firewall rules
3. Verify TLS/SSL requirements if applicable

### DNS Scanning Timeouts

If DNS scans are timing out:

1. Reduce the range of sequence numbers to scan
2. Lower the `MaxConcurrent` parameter
3. Check DNS server responsiveness

## Command Reference

### Connection Management

| Command | Description |
|---------|-------------|
| `Initialize-HnsConnection` | Initializes the connection to the HNS API |
| `Connect-HnsService` | Authenticates with username/password |
| `Test-HnsConnection` | Tests if the HNS API is accessible |

### Template Management

| Command | Description |
|---------|-------------|
| `Get-HnsTemplates` | Lists templates with pagination |
| `Get-HnsTemplate` | Gets details for a specific template |
| `New-HnsTemplate` | Creates a new template |

### Hostname Management

| Command | Description |
|---------|-------------|
| `Get-HnsHostnames` | Lists hostnames with filtering and pagination |
| `Get-HnsHostname` | Gets details for a specific hostname |
| `New-HnsHostnameGeneration` | Generates a hostname without reserving it |
| `New-HnsHostnameReservation` | Reserves a hostname |
| `Set-HnsHostnameCommit` | Commits a reserved hostname |
| `Set-HnsHostnameRelease` | Releases a committed hostname |

### DNS Operations

| Command | Description |
|---------|-------------|
| `Test-HnsDnsHostname` | Checks if a hostname exists in DNS |
| `Start-HnsDnsScan` | Scans a range of hostnames in DNS |

### API Key Management

| Command | Description |
|---------|-------------|
| `Get-HnsApiKeys` | Lists API keys |
| `New-HnsApiKey` | Creates a new API key |
| `Remove-HnsApiKey` | Deletes an API key |

### Sequence Management

| Command | Description |
|---------|-------------|
| `Get-HnsNextSequenceNumber` | Gets the next available sequence number |

## Conclusion

The HNS PowerShell module provides a powerful interface for managing hostnames through the HNS API service. By automating hostname generation, reservation, and lifecycle management, you can ensure consistent naming conventions and avoid naming conflicts in your infrastructure.