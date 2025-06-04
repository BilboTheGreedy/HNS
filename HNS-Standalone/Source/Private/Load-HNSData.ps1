function Load-HNSData {
    <#
    .SYNOPSIS
        Loads HNS data from disk
    .DESCRIPTION
        Loads templates, hostnames, and configuration from JSON files
    .PARAMETER Type
        Type of data to load (Templates, Hostnames, Configuration, All)
    .PARAMETER Force
        Force reload even if data is already loaded
    #>
    [CmdletBinding()]
    param(
        [Parameter()]
        [ValidateSet('Templates', 'Hostnames', 'Configuration', 'All')]
        [string]$Type = 'All',
        
        [Parameter()]
        [switch]$Force
    )
    
    begin {
        if (-not $script:Configuration -and $Type -ne 'Configuration') {
            throw "HNS environment not initialized. Run Initialize-HNSEnvironment first."
        }
    }
    
    process {
        try {
            # Load configuration
            if ($Type -in 'Configuration', 'All') {
                $configPath = $script:Configuration.GetConfigPath()
                if (Test-Path $configPath) {
                    $configData = Get-Content -Path $configPath -Raw | ConvertFrom-Json
                    $script:Configuration = [HNSConfiguration]::new($configData)
                    Write-Verbose "Loaded configuration"
                }
            }
            
            # Load templates
            if ($Type -in 'Templates', 'All') {
                if ($Force -or -not $script:Templates) {
                    $templatesPath = $script:Configuration.GetTemplatesPath()
                    if (Test-Path $templatesPath) {
                        $templateData = Get-Content -Path $templatesPath -Raw | ConvertFrom-Json
                        $script:Templates = @()
                        
                        foreach ($item in $templateData) {
                            # Convert PSCustomObject to hashtable recursively
                            $hashtable = ConvertTo-HashtableRecursive $item
                            $template = [Template]::new($hashtable)
                            $script:Templates += $template
                        }
                        
                        Write-Verbose "Loaded $($script:Templates.Count) templates"
                    } else {
                        $script:Templates = @()
                        Write-Verbose "No templates file found, initialized empty collection"
                    }
                }
            }
            
            # Load hostnames
            if ($Type -in 'Hostnames', 'All') {
                if ($Force -or -not $script:Hostnames) {
                    $hostnamesPath = $script:Configuration.GetHostnamesPath()
                    if (Test-Path $hostnamesPath) {
                        $hostnameData = Get-Content -Path $hostnamesPath -Raw | ConvertFrom-Json
                        $script:Hostnames = @()
                        
                        foreach ($item in $hostnameData) {
                            # Convert PSCustomObject to hashtable recursively
                            $hashtable = ConvertTo-HashtableRecursive $item
                            $hostname = [Hostname]::new($hashtable)
                            $script:Hostnames += $hostname
                        }
                        
                        Write-Verbose "Loaded $($script:Hostnames.Count) hostnames"
                    } else {
                        $script:Hostnames = @()
                        Write-Verbose "No hostnames file found, initialized empty collection"
                    }
                }
            }
            
            Write-Verbose "HNS data loaded successfully"
        }
        catch {
            Write-Error "Failed to load HNS data: $_"
            throw
        }
    }
}