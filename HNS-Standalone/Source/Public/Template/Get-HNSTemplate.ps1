function Get-HNSTemplate {
    <#
    .SYNOPSIS
        Gets hostname templates
    .DESCRIPTION
        Retrieves one or more hostname templates based on specified criteria
    .PARAMETER ID
        The ID of a specific template
    .PARAMETER Name
        The name of a template (supports wildcards)
    .PARAMETER Active
        Filter to only active templates
    .PARAMETER CreatedBy
        Filter by creator username
    .PARAMETER IncludeExample
        Include an example hostname for each template
    .EXAMPLE
        Get-HNSTemplate
        
        Gets all templates
    .EXAMPLE
        Get-HNSTemplate -Name "Epiroc*" -Active
        
        Gets all active templates with names starting with "Epiroc"
    .EXAMPLE
        Get-HNSTemplate -ID 1 -IncludeExample
        
        Gets template with ID 1 and includes an example hostname
    #>
    [CmdletBinding(DefaultParameterSetName = 'All')]
    param(
        [Parameter(ParameterSetName = 'ByID', ValueFromPipelineByPropertyName)]
        [int]$ID,
        
        [Parameter(ParameterSetName = 'ByName', ValueFromPipeline)]
        [SupportsWildcards()]
        [string]$Name,
        
        [Parameter()]
        [switch]$Active,
        
        [Parameter()]
        [string]$CreatedBy,
        
        [Parameter()]
        [switch]$IncludeExample
    )
    
    begin {
        Test-HNSInitialized -Throw
        Write-Verbose "Getting templates with criteria: $($PSBoundParameters | Out-String)"
    }
    
    process {
        try {
            # Ensure we only work with Template objects
            $templates = $script:Templates | Where-Object { $_ -is [Template] }
            
            # Filter by ID
            if ($PSBoundParameters.ContainsKey('ID')) {
                $templates = $templates | Where-Object { $_.ID -eq $ID }
            }
            
            # Filter by Name (with wildcard support)
            if ($PSBoundParameters.ContainsKey('Name')) {
                $templates = $templates | Where-Object { $_.Name -like $Name }
            }
            
            # Filter by Active status
            if ($Active) {
                $templates = $templates | Where-Object { $_.IsActive }
            }
            
            # Filter by CreatedBy
            if ($PSBoundParameters.ContainsKey('CreatedBy')) {
                $templates = $templates | Where-Object { $_.CreatedBy -like $CreatedBy }
            }
            
            # Add example if requested
            if ($IncludeExample -and $templates) {
                foreach ($template in $templates) {
                    $example = $template.GenerateExample()
                    $template | Add-Member -NotePropertyName 'ExampleHostname' -NotePropertyValue $example -Force
                }
            }
            
            Write-Verbose "Found $($templates.Count) template(s)"
            return $templates
        }
        catch {
            Write-Error "Failed to get templates: $_"
            throw
        }
    }
}