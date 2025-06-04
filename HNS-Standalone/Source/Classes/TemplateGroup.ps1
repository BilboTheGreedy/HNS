class TemplateGroup {
    [int] $ID
    [int] $TemplateID
    [string] $Name
    [int] $Length
    [int] $Position
    [bool] $IsRequired
    [ValidationType] $ValidationType
    [string] $ValidationValue
    
    # Default constructor
    TemplateGroup() {
        $this.IsRequired = $true
    }
    
    # Constructor with hashtable
    TemplateGroup([hashtable]$Properties) {
        $this.IsRequired = $true
        
        foreach ($key in $Properties.Keys) {
            if ($null -ne $this.PSObject.Properties[$key]) {
                $this.$key = $Properties[$key]
            }
        }
    }
    
    # Constructor with all parameters
    TemplateGroup(
        [string]$Name,
        [int]$Length,
        [int]$Position,
        [ValidationType]$ValidationType,
        [string]$ValidationValue
    ) {
        $this.Name = $Name
        $this.Length = $Length
        $this.Position = $Position
        $this.ValidationType = $ValidationType
        $this.ValidationValue = $ValidationValue
        $this.IsRequired = $true
    }
    
    # Validate a value against this group
    [bool] Validate([string]$Value) {
        # Check length
        if ($Value.Length -ne $this.Length) {
            return $false
        }
        
        switch ($this.ValidationType) {
            ([ValidationType]::Fixed) {
                return $Value -eq $this.ValidationValue
            }
            ([ValidationType]::List) {
                $validValues = $this.ValidationValue -split ','
                return $Value -in $validValues
            }
            ([ValidationType]::Regex) {
                return $Value -match $this.ValidationValue
            }
            ([ValidationType]::Sequence) {
                return $Value -match '^\d+$'
            }
            default {
                return $true
            }
        }
        return $true
    }
    
    # ToString override for display
    [string] ToString() {
        return "$($this.Name) (Pos: $($this.Position), Len: $($this.Length), Type: $($this.ValidationType))"
    }
    
    # Clone method
    [TemplateGroup] Clone() {
        return [TemplateGroup]::new(@{
            ID = $this.ID
            TemplateID = $this.TemplateID
            Name = $this.Name
            Length = $this.Length
            Position = $this.Position
            IsRequired = $this.IsRequired
            ValidationType = $this.ValidationType
            ValidationValue = $this.ValidationValue
        })
    }
}