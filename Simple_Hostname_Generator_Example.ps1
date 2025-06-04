# Simple Epiroc Hostname Generator Example
# ========================================
# No database, no reservations - just available hostnames!

Write-Host "Epiroc VM Standard - Available Hostname Generator" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# Import the simplified module
Import-Module -Name ".\HNS-Standalone\HNS-Standalone.psm1" -Force

# Example 1: Basic usage - Get next available Azure production app server
Write-Host "`n1. Basic Example - Azure Production Application Server" -ForegroundColor Yellow
$hostname1 = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS'

Write-Host "   Available hostname: $($hostname1.Hostname)" -ForegroundColor Green
Write-Host "   Structure: $($hostname1.UnitCode) + S + $($hostname1.Provider) + $($hostname1.Region) + $($hostname1.Environment) + $($hostname1.Function) + $($hostname1.Sequence.ToString('000'))" -ForegroundColor DarkGray
Write-Host "   Sequences tested: $($hostname1.SequencesTested)" -ForegroundColor DarkGray

# Example 2: Start from specific sequence number
Write-Host "`n2. Start from Sequence 10 - Test Database Server" -ForegroundColor Yellow
$hostname2 = Get-AvailableHostname -UnitCode 'USS' -Provider 'M' -Region 'SCUS' -Environment 'T' -Function 'DB' -StartSequence 10

Write-Host "   Available hostname: $($hostname2.Hostname)" -ForegroundColor Green
Write-Host "   Started from sequence: 10, found at: $($hostname2.Sequence)" -ForegroundColor DarkGray

# Example 3: Only DNS and ICMP checks (skip AD)
Write-Host "`n3. DNS + ICMP Only - Web Server" -ForegroundColor Yellow
$hostname3 = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'D' -Function 'WS' -CheckActiveDirectory $false

Write-Host "   Available hostname: $($hostname3.Hostname)" -ForegroundColor Green
Write-Host "   Checks performed: DNS=$($hostname3.ChecksPerformed.DNS), ICMP=$($hostname3.ChecksPerformed.ICMP), AD=$($hostname3.ChecksPerformed.ActiveDirectory)" -ForegroundColor DarkGray

# Example 4: Epiroc Private Cloud with all checks
Write-Host "`n4. Epiroc Private Cloud - Backup Server" -ForegroundColor Yellow
$hostname4 = Get-AvailableHostname -UnitCode 'SFD' -Provider 'E' -Region 'EUSE' -Environment 'P' -Function 'BU'

Write-Host "   Available hostname: $($hostname4.Hostname)" -ForegroundColor Green
Write-Host "   Provider: Epiroc Private Cloud (E)" -ForegroundColor DarkGray

# Example 5: Bulk generation for different environments
Write-Host "`n5. Bulk Generation - File Servers for All Environments" -ForegroundColor Yellow
$environments = @(
    @{Code='D'; Name='Development'},
    @{Code='T'; Name='Test'},
    @{Code='P'; Name='Production'}
)

foreach ($env in $environments) {
    $hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment $env.Code -Function 'FS'
    Write-Host "   $($env.Name): $($hostname.Hostname)" -ForegroundColor Green
}

# Example 6: Error handling - what happens when no hostnames available
Write-Host "`n6. Error Handling Example" -ForegroundColor Yellow
try {
    # Try to find hostname but limit to very small range
    $hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS' -StartSequence 999 -MaxSequence 999 -CheckDNS $false -CheckICMP $false -CheckActiveDirectory $false
    Write-Host "   Found: $($hostname.Hostname)" -ForegroundColor Green
} catch {
    Write-Host "   ✓ Properly handled: $($_.Exception.Message)" -ForegroundColor Yellow
}

# Example 7: ServiceNow integration (commented out - requires credentials)
Write-Host "`n7. ServiceNow Integration (Example)" -ForegroundColor Yellow
Write-Host "   # Uncomment and configure to enable ServiceNow CMDB checking:" -ForegroundColor DarkGray
Write-Host "   # `$cred = Get-Credential" -ForegroundColor DarkGray
Write-Host "   # `$hostname = Get-AvailableHostname -UnitCode 'SFD' -Provider 'M' -Region 'WEEU' -Environment 'P' -Function 'AS' ``" -ForegroundColor DarkGray
Write-Host "   #             -CheckServiceNow -ServiceNowInstance 'https://company.servicenow.com' -ServiceNowCredential `$cred" -ForegroundColor DarkGray

Write-Host "`n" + "="*60 -ForegroundColor Cyan
Write-Host "Summary of New Approach:" -ForegroundColor Green
Write-Host "• ✓ No database required" -ForegroundColor Green
Write-Host "• ✓ No reservations or lifecycle management" -ForegroundColor Green  
Write-Host "• ✓ Only returns available hostnames" -ForegroundColor Green
Write-Host "• ✓ Checks DNS, ICMP, AD, ServiceNow" -ForegroundColor Green
Write-Host "• ✓ Auto-increments until available" -ForegroundColor Green
Write-Host "• ✓ Full Epiroc standard compliance" -ForegroundColor Green
Write-Host "• ✓ Lightweight and fast" -ForegroundColor Green
Write-Host "="*60 -ForegroundColor Cyan

Write-Host "`nReady to use! Each call returns an immediately usable hostname." -ForegroundColor White