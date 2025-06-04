function Backup-HNSData {
    <#
    .SYNOPSIS
        Creates a backup of HNS data files
    .DESCRIPTION
        Backs up configuration, templates, and hostnames data to the backup directory
    .PARAMETER Type
        Type of data to backup (Configuration, Templates, Hostnames, All)
    .PARAMETER BackupPath
        Custom backup path (optional)
    #>
    [CmdletBinding()]
    param(
        [Parameter()]
        [ValidateSet('Configuration', 'Templates', 'Hostnames', 'All')]
        [string]$Type = 'All',
        
        [Parameter()]
        [string]$BackupPath
    )
    
    if (-not (Test-HNSInitialized)) {
        Write-Warning "HNS not initialized, skipping backup"
        return
    }
    
    try {
        # Determine backup directory
        if (-not $BackupPath) {
            $BackupPath = $script:Configuration.GetBackupPath()
        }
        
        # Ensure backup directory exists
        if (-not (Test-Path $BackupPath)) {
            New-Item -Path $BackupPath -ItemType Directory -Force | Out-Null
        }
        
        $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
        $backupResults = @()
        
        # Backup functions
        $backupActions = @{
            'Configuration' = {
                $sourceFile = $script:Configuration.GetConfigPath()
                if (Test-Path $sourceFile) {
                    $backupFile = Join-Path $BackupPath "config-$timestamp.json"
                    Copy-Item -Path $sourceFile -Destination $backupFile -Force
                    return $backupFile
                }
            }
            'Templates' = {
                $sourceFile = $script:Configuration.GetTemplatesPath()
                if (Test-Path $sourceFile) {
                    $backupFile = Join-Path $BackupPath "templates-$timestamp.json"
                    Copy-Item -Path $sourceFile -Destination $backupFile -Force
                    return $backupFile
                }
            }
            'Hostnames' = {
                $sourceFile = $script:Configuration.GetHostnamesPath()
                if (Test-Path $sourceFile) {
                    $backupFile = Join-Path $BackupPath "hostnames-$timestamp.json"
                    Copy-Item -Path $sourceFile -Destination $backupFile -Force
                    return $backupFile
                }
            }
        }
        
        # Execute backups
        if ($Type -eq 'All') {
            foreach ($backupType in $backupActions.Keys) {
                $result = & $backupActions[$backupType]
                if ($result) {
                    $backupResults += @{ Type = $backupType; File = $result }
                    Write-Verbose "Backed up $backupType to: $result"
                }
            }
        } else {
            $result = & $backupActions[$Type]
            if ($result) {
                $backupResults += @{ Type = $Type; File = $result }
                Write-Verbose "Backed up $Type to: $result"
            }
        }
        
        Write-Verbose "Backup completed. $($backupResults.Count) files backed up."
        return $backupResults
        
    } catch {
        Write-Error "Backup failed: $($_.Exception.Message)"
        throw
    }
}