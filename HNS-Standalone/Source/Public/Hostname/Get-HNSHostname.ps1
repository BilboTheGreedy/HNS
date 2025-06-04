function Get-HNSHostname {
    <#
    .SYNOPSIS
        Retrieves hostname records from the HNS system
    .DESCRIPTION
        Gets hostname records based on various filters and criteria
    .PARAMETER Name
        Specific hostname to retrieve
    .PARAMETER ID
        Specific hostname ID to retrieve
    .PARAMETER TemplateID
        Filter by template ID
    .PARAMETER TemplateName
        Filter by template name
    .PARAMETER Status
        Filter by hostname status
    .PARAMETER ReservedBy
        Filter by who reserved the hostname
    .PARAMETER All
        Return all hostnames
    .EXAMPLE
        Get-HNSHostname -Name "server001"
    .EXAMPLE
        Get-HNSHostname -TemplateName "Epiroc VM Standard" -Status Reserved
    #>
    [CmdletBinding(DefaultParameterSetName = 'All')]
    param(
        [Parameter(ParameterSetName = 'ByName')]
        [string]$Name,
        
        [Parameter(ParameterSetName = 'ByID')]
        [int]$ID,
        
        [Parameter(ParameterSetName = 'Filter')]
        [int]$TemplateID,
        
        [Parameter(ParameterSetName = 'Filter')]
        [string]$TemplateName,
        
        [Parameter(ParameterSetName = 'Filter')]
        [HostnameStatus]$Status,
        
        [Parameter(ParameterSetName = 'Filter')]
        [string]$ReservedBy,
        
        [Parameter(ParameterSetName = 'All')]
        [switch]$All
    )
    
    begin {
        Test-HNSInitialized -Throw
        Write-Verbose "Retrieving hostname records"
    }
    
    process {
        try {
            # Start with all hostnames
            $results = $script:Hostnames
            
            # Apply filters based on parameter set
            switch ($PSCmdlet.ParameterSetName) {
                'ByName' {
                    $results = $results | Where-Object { $_.Name -eq $Name }
                }
                'ByID' {
                    $results = $results | Where-Object { $_.ID -eq $ID }
                }
                'Filter' {
                    if ($TemplateID) {
                        $results = $results | Where-Object { $_.TemplateID -eq $TemplateID }
                    }
                    if ($TemplateName) {
                        $results = $results | Where-Object { $_.TemplateName -eq $TemplateName }
                    }
                    if ($PSBoundParameters.ContainsKey('Status')) {
                        $results = $results | Where-Object { $_.Status -eq $Status }
                    }
                    if ($ReservedBy) {
                        $results = $results | Where-Object { $_.ReservedBy -eq $ReservedBy }
                    }
                }
                'All' {
                    # Return all results (no additional filtering)
                }
            }
            
            Write-Verbose "Found $($results.Count) matching hostname(s)"
            return $results
            
        } catch {
            Write-Error "Failed to retrieve hostnames: $($_.Exception.Message)"
            throw
        }
    }
}