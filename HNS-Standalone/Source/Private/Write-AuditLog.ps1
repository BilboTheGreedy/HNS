function Write-AuditLog {
    <#
    .SYNOPSIS
        Writes an entry to the audit log
    .DESCRIPTION
        Records actions taken in the HNS system for audit purposes
    .PARAMETER Action
        The action being performed
    .PARAMETER Details
        Additional details about the action
    .PARAMETER User
        The user performing the action (defaults to current user)
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Action,
        
        [Parameter()]
        [string]$Details = "",
        
        [Parameter()]
        [string]$User = $env:USERNAME
    )
    
    # Check if audit logging is enabled
    if (-not $script:Configuration -or -not $script:Configuration.EnableAuditLog) {
        return
    }
    
    try {
        $auditPath = $script:Configuration.GetAuditLogPath()
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        
        # Escape details for CSV
        $escapedDetails = $Details -replace '"', '""'
        if ($escapedDetails -match '[,\r\n"]') {
            $escapedDetails = "`"$escapedDetails`""
        }
        
        $logEntry = "$timestamp,$Action,$User,$escapedDetails"
        
        # Use mutex to prevent concurrent writes
        $mutex = New-Object System.Threading.Mutex($false, "HNS_AuditLog_Mutex")
        try {
            $mutex.WaitOne() | Out-Null
            Add-Content -Path $auditPath -Value $logEntry -Encoding UTF8
        }
        finally {
            $mutex.ReleaseMutex()
            $mutex.Dispose()
        }
        
        Write-Verbose "Audit log entry: $Action - $Details"
    }
    catch {
        Write-Warning "Failed to write audit log: $_"
    }
}