# HNS-BulkOperations.ps1
# Example script for performing bulk operations with the HNS API

# Import the module
Import-Module -Name '.\HNS-API.psm1' -Force

# Configuration
$HnsUrl = "http://localhost:8080"  # Change to your HNS server URL
$ApiKey = "YOUR_API_KEY_HERE"      # Insert your API key with appropriate permissions

# Initialize the connection with API key
Initialize-HnsConnection -BaseUrl $HnsUrl -ApiKey $ApiKey

# Alternative method using username/password authentication
# $Username = "admin"
# $Password = ConvertTo-SecureString "admin123" -AsPlainText -Force
# Initialize-HnsConnection -BaseUrl $HnsUrl
# Connect-HnsService -Username $Username -Password $Password

# Test the connection
if (-not (Test-HnsConnection)) {
    Write-Error "Failed to connect to HNS server. Please check the connection details."
    exit 1
}

# Function to get or create a template
function Get-OrCreateTemplate {
    param(
        [string]$TemplateName = "BulkOperationsTemplate"
    )
    
    # Check if template already exists
    $templates = Get-HnsTemplates -Limit 100
    $template = $templates | Where-Object { $_.name -eq $TemplateName }
    
    if ($template) {
        Write-Host "Using existing template: $($template.name) (ID: $($template.id))" -ForegroundColor Green
        return $template
    }
    
    # Create a new template
    Write-Host "Creating new template: $TemplateName" -ForegroundColor Yellow
    
    $groups = @(
        @{
            name = "location"
            length = 2
            is_required = $true
            validation_type = "list"
            validation_value = "NY,LA,SF,DC,CH,AM,TO,LN"
        },
        @{
            name = "role"
            length = 3
            is_required = $true
            validation_type = "list"
            validation_value = "WEB,APP,DB,LB,NAS,BKP"
        },
        @{
            name = "env"
            length = 3
            is_required = $true
            validation_type = "list"
            validation_value = "DEV,TST,STG,PRD,DR"
        },
        @{
            name = "sequence"
            length = 3
            is_required = $true
            validation_type = "sequence"
            validation_value = ""
        }
    )
    
    $newTemplate = New-HnsTemplate -Name $TemplateName -Description "Template for bulk operations" `
        -MaxLength 15 -SequenceStart 1 -SequenceLength 3 -SequencePadding $true -SequenceIncrement 1 -Groups $groups
        
    Write-Host "Created new template with ID: $($newTemplate.id)" -ForegroundColor Green
    return $newTemplate
}

# Function to bulk reserve hostnames
function Reserve-BulkHostnames {
    param(
        [int]$TemplateId,
        [int]$Count = 5,
        [string]$Location = "NY",
        [string]$Role = "WEB",
        [string]$Environment = "DEV"
    )
    
    Write-Host "Bulk reserving $Count hostnames with Location=$Location, Role=$Role, Environment=$Environment" -ForegroundColor Cyan
    
    $reservedHostnames = @()
    
    for ($i = 0; $i -lt $Count; $i++) {
        $params = @{
            location = $Location
            role = $Role
            env = $Environment
        }
        
        try {
            $hostname = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params
            $reservedHostnames += $hostname
            Write-Host "Reserved: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Green
        }
        catch {
            Write-Error "Failed to reserve hostname #$($i+1): $_"
        }
    }
    
    Write-Host "Successfully reserved $($reservedHostnames.Count) of $Count hostnames" -ForegroundColor Cyan
    return $reservedHostnames
}

# Function to commit multiple hostnames
function Commit-BulkHostnames {
    param(
        [array]$HostnameIds
    )
    
    Write-Host "Committing $($HostnameIds.Count) hostnames..." -ForegroundColor Cyan
    
    $committedHostnames = @()
    
    foreach ($id in $HostnameIds) {
        try {
            $hostname = Set-HnsHostnameCommit -HostnameId $id
            $committedHostnames += $hostname
            Write-Host "Committed: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Green
        }
        catch {
            Write-Error "Failed to commit hostname ID $id: $_"
        }
    }
    
    Write-Host "Successfully committed $($committedHostnames.Count) of $($HostnameIds.Count) hostnames" -ForegroundColor Cyan
    return $committedHostnames
}

# Function to scan for existing hostnames in DNS
function Find-ExistingHostnames {
    param(
        [int]$TemplateId,
        [hashtable]$Params,
        [int]$StartSeq = 1,
        [int]$EndSeq = 100
    )
    
    Write-Host "Scanning for existing hostnames in DNS (range $StartSeq-$EndSeq)..." -ForegroundColor Cyan
    
    $scanResult = Start-HnsDnsScan -TemplateId $TemplateId -StartSeq $StartSeq -EndSeq $EndSeq -Params $Params -MaxConcurrent 10
    
    Write-Host "Scan complete: $($scanResult.existing_hostnames) of $($scanResult.total_hostnames) hostnames exist in DNS" -ForegroundColor Green
    
    if ($scanResult.existing_hostnames -gt 0) {
        $existingHosts = $scanResult.results | Where-Object { $_.exists -eq $true }
        $existingHosts | Format-Table -Property hostname, ip_address -AutoSize
    }
    
    return $scanResult
}

# Function to export hostname data to CSV
function Export-HostnamesToCsv {
    param(
        [array]$Hostnames,
        [string]$FilePath = ".\hostnames_export.csv"
    )
    
    Write-Host "Exporting $($Hostnames.Count) hostnames to $FilePath..." -ForegroundColor Cyan
    
    # Create simplified objects for CSV export
    $exportData = $Hostnames | Select-Object id, name, status, sequence_num, reserved_by, 
        @{Name="reserved_at"; Expression={$_.reserved_at}},
        @{Name="committed_by"; Expression={$_.committed_by}},
        @{Name="committed_at"; Expression={$_.committed_at}}
    
    $exportData | Export-Csv -Path $FilePath -NoTypeInformation
    
    Write-Host "Export complete. File saved to $FilePath" -ForegroundColor Green
}

# Function to import hostname data from CSV and reserve them
function Import-HostnamesFromCsv {
    param(
        [string]$FilePath,
        [int]$TemplateId
    )
    
    if (-not (Test-Path $FilePath)) {
        Write-Error "File not found: $FilePath"
        return
    }
    
    Write-Host "Importing hostnames from $FilePath..." -ForegroundColor Cyan
    
    $importData = Import-Csv -Path $FilePath
    $reservedHostnames = @()
    
    foreach ($item in $importData) {
        # Parse parameters from hostname or use provided fields
        # This is a simplified example - you may need to adjust based on your actual CSV format
        $params = @{}
        
        if ($item.location) { $params.location = $item.location }
        if ($item.role) { $params.role = $item.role }
        if ($item.env) { $params.env = $item.env }
        
        try {
            $hostname = New-HnsHostnameReservation -TemplateId $TemplateId -Params $params
            $reservedHostnames += $hostname
            Write-Host "Reserved: $($hostname.name) (ID: $($hostname.id))" -ForegroundColor Green
        }
        catch {
            Write-Error "Failed to reserve hostname for row $($importData.IndexOf($item)+1): $_"
        }
    }
    
    Write-Host "Successfully reserved $($reservedHostnames.Count) of $($importData.Count) hostnames" -ForegroundColor Cyan
    return $reservedHostnames
}

# Function to generate a report of hostname usage by template
function Get-HostnameUsageReport {
    param()
    
    Write-Host "Generating hostname usage report..." -ForegroundColor Cyan
    
    $templates = Get-HnsTemplates -Limit 100
    $report = @()
    
    foreach ($template in $templates) {
        $reservedCount = 0
        $committedCount = 0
        $releasedCount = 0
        
        # Get counts for each status
        try {
            $reservedHostnames = Get-HnsHostnames -Status "reserved" -TemplateId $template.id -Limit 1
            $reservedCount = $reservedHostnames.total
        } catch {}
        
        try {
            $committedHostnames = Get-HnsHostnames -Status "committed" -TemplateId $template.id -Limit 1
            $committedCount = $committedHostnames.total
        } catch {}
        
        try {
            $releasedHostnames = Get-HnsHostnames -Status "released" -TemplateId $template.id -Limit 1
            $releasedCount = $releasedHostnames.total
        } catch {}
        
        $totalCount = $reservedCount + $committedCount + $releasedCount
        
        # Add to report
        $report += [PSCustomObject]@{
            TemplateID = $template.id
            TemplateName = $template.name
            TotalHostnames = $totalCount
            Reserved = $reservedCount
            Committed = $committedCount
            Released = $releasedCount
        }
    }
    
    Write-Host "Hostname usage by template:" -ForegroundColor Green
    $report | Format-Table -AutoSize
    
    return $report
}

# Example usage - uncomment sections as needed

# Get or create a template
$template = Get-OrCreateTemplate -TemplateName "BulkDemo"

# Reserve multiple hostnames for different environments
$devHostnames = Reserve-BulkHostnames -TemplateId $template.id -Count 3 -Location "NY" -Role "WEB" -Environment "DEV"
$tstHostnames = Reserve-BulkHostnames -TemplateId $template.id -Count 3 -Location "SF" -Role "APP" -Environment "TST"
$prdHostnames = Reserve-BulkHostnames -TemplateId $template.id -Count 3 -Location "DC" -Role "DB" -Environment "PRD"

# Commit all development hostnames
$devHostnameIds = $devHostnames | Select-Object -ExpandProperty id
$committedDevHostnames = Commit-BulkHostnames -HostnameIds $devHostnameIds

# Export hostnames to CSV
$allHostnames = $devHostnames + $tstHostnames + $prdHostnames
Export-HostnamesToCsv -Hostnames $allHostnames -FilePath ".\bulk_demo_hostnames.csv"

# Scan DNS for existing hostnames
$scanParams = @{
    location = "NY"
    role = "WEB"
    env = "PRD"
}
$scanResults = Find-ExistingHostnames -TemplateId $template.id -Params $scanParams -StartSeq 1 -EndSeq 20

# Generate usage report
$usageReport = Get-HostnameUsageReport

Write-Host "Bulk operations completed successfully!" -ForegroundColor Cyan