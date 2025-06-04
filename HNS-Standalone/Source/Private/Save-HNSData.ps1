function Save-HNSData {
    <#
    .SYNOPSIS
        Saves HNS data to disk
    .DESCRIPTION
        Persists templates, hostnames, and configuration to JSON files
    .PARAMETER Type
        Type of data to save (Templates, Hostnames, Configuration, All)
    .PARAMETER CreateBackup
        Create a backup before saving
    #>
    [CmdletBinding()]
    param(
        [Parameter()]
        [ValidateSet('Templates', 'Hostnames', 'Configuration', 'All')]
        [string]$Type = 'All',
        
        [Parameter()]
        [switch]$CreateBackup
    )
    
    begin {
        if (-not $script:Configuration) {
            throw "HNS environment not initialized. Run Initialize-HNSEnvironment first."
        }
    }
    
    process {
        try {
            # Create backup if requested
            if ($CreateBackup -and $script:Configuration.AutoBackup) {
                Backup-HNSData -Quiet
            }
            
            # Use mutex to prevent concurrent writes
            $mutex = New-Object System.Threading.Mutex($false, "HNS_DataSave_Mutex")
            try {
                $mutex.WaitOne() | Out-Null
                
                # Save templates
                if ($Type -in 'Templates', 'All') {
                    $templatesPath = $script:Configuration.GetTemplatesPath()
                    $script:Templates | ConvertTo-Json -Depth 10 | Set-Content -Path $templatesPath -Encoding UTF8
                    Write-Verbose "Saved $($script:Templates.Count) templates"
                }
                
                # Save hostnames
                if ($Type -in 'Hostnames', 'All') {
                    $hostnamesPath = $script:Configuration.GetHostnamesPath()
                    $script:Hostnames | ConvertTo-Json -Depth 10 | Set-Content -Path $hostnamesPath -Encoding UTF8
                    Write-Verbose "Saved $($script:Hostnames.Count) hostnames"
                }
                
                # Save configuration
                if ($Type -in 'Configuration', 'All') {
                    $configPath = $script:Configuration.GetConfigPath()
                    $script:Configuration | ConvertTo-Json -Depth 10 | Set-Content -Path $configPath -Encoding UTF8
                    Write-Verbose "Saved configuration"
                }
            }
            finally {
                $mutex.ReleaseMutex()
                $mutex.Dispose()
            }
            
            Write-Verbose "HNS data saved successfully"
        }
        catch {
            Write-Error "Failed to save HNS data: $_"
            throw
        }
    }
}