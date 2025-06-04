BeforeAll {
    $ModulePath = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
    Import-Module "$ModulePath\HNS-Standalone.psd1" -Force
    
    # Initialize test environment
    $TestPath = Join-Path $TestDrive 'HNS-Test'
    Initialize-HNSEnvironment -Path $TestPath -Force
}

Describe 'Template Management Tests' {
    Context 'New-HNSTemplate' {
        It 'Should create a basic template' {
            $groups = @(
                @{Name='Prefix'; Length=3; Position=1; ValidationType='Fixed'; ValidationValue='TST'}
                @{Name='Sequence'; Length=3; Position=2; ValidationType='Sequence'; ValidationValue=''}
            )
            
            $template = New-HNSTemplate -Name 'Test Template' -MaxLength 6 -Groups $groups -SequencePosition 2
            
            $template | Should -Not -BeNullOrEmpty
            $template.Name | Should -Be 'Test Template'
            $template.MaxLength | Should -Be 6
            $template.Groups.Count | Should -Be 2
        }
        
        It 'Should reject duplicate template names' {
            { New-HNSTemplate -Name 'Test Template' -MaxLength 6 -Groups @() -SequencePosition 1 } | 
                Should -Throw "*already exists*"
        }
        
        It 'Should validate total group length equals MaxLength' {
            $groups = @(
                @{Name='Part1'; Length=5; Position=1; ValidationType='Fixed'; ValidationValue='ABCDE'}
            )
            
            { New-HNSTemplate -Name 'Invalid Length' -MaxLength 10 -Groups $groups -SequencePosition 1 } | 
                Should -Throw "*must equal MaxLength*"
        }
    }
    
    Context 'Get-HNSTemplate' {
        It 'Should retrieve all templates' {
            $templates = Get-HNSTemplate
            $templates.Count | Should -BeGreaterThan 0
        }
        
        It 'Should filter by name with wildcards' {
            $templates = Get-HNSTemplate -Name 'Test*'
            $templates | ForEach-Object { $_.Name | Should -BeLike 'Test*' }
        }
        
        It 'Should include example when requested' {
            $template = Get-HNSTemplate -Name 'Test Template' -IncludeExample
            $template.ExampleHostname | Should -Be 'TST001'
        }
    }
}