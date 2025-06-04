# HNS-Standalone PowerShell Module

A professional standalone PowerShell module for Hostname Naming Service (HNS) - Generate and manage standardized hostnames based on configurable templates without requiring server infrastructure.

## Features

- **Template-Based Generation**: Create flexible naming templates with validation rules
- **Lifecycle Management**: Track hostname states (Available → Reserved → Committed → Released)
- **Sequence Management**: Automatic sequence number allocation with gap detection
- **DNS Verification**: Built-in DNS checking capabilities
- **Audit Trail**: Comprehensive logging of all operations
- **Data Persistence**: JSON-based storage with backup capabilities
- **Epiroc VM Standard**: Pre-built support for Epiroc's 15-character naming convention

## Requirements

- PowerShell 5.1 or higher
- Windows, Linux, or macOS

## Installation

### From PowerShell Gallery (Recommended)

```powershell
Install-Module -Name HNS-Standalone -Scope CurrentUser
```

### Manual Installation

1. Download the latest release
2. Extract to your PowerShell modules directory
3. Import the module:

```powershell
Import-Module HNS-Standalone
```

## Quick Start

### 1. Initialize the Environment

```powershell
# Initialize with default settings
Initialize-HNSEnvironment

# Initialize with Epiroc template and custom path
Initialize-HNSEnvironment -Path "C:\HNS" -LoadEpirocTemplate
```

### 2. Create a Template

```powershell
$groups = @(
    @{Name='Department'; Length=3; Position=1; ValidationType='List'; ValidationValue='IT,HR,FIN,OPS'}
    @{Name='Type'; Length=1; Position=2; ValidationType='Fixed'; ValidationValue='S'}
    @{Name='Location'; Length=2; Position=3; ValidationType='List'; ValidationValue='NY,LA,CH,TX'}
    @{Name='Sequence'; Length=3; Position=4; ValidationType='Sequence'; ValidationValue=''}
)

New-HNSTemplate -Name "Corporate Servers" `
    -Description "Standard naming for corporate servers" `
    -MaxLength 9 `
    -Groups $groups `
    -SequencePosition 4
```

### 3. Generate Hostnames

```powershell
# Generate and reserve a hostname
$hostname = New-HNSHostname -TemplateName "Corporate Servers" `
    -Parameters @{Department='IT'; Location='NY'} `
    -Reserve

# Output: ITSNY001

# Generate multiple hostnames
1..5 | ForEach-Object {
    New-HNSHostname -TemplateName "Corporate Servers" `
        -Parameters @{Department='HR'; Location='LA'}
}
```

### 4. Manage Hostname Lifecycle

```powershell
# Reserve an available hostname
Reserve-HNSHostname -Name "HRSLA001"

# Commit when deployed
Commit-HNSHostname -Name "HRSLA001"

# Release when decommissioned
Release-HNSHostname -Name "HRSLA001"
```

## Epiroc VM Standard Example

```powershell
# Initialize with Epiroc template
Initialize-HNSEnvironment -LoadEpirocTemplate

# Generate Epiroc-compliant hostname
$hostname = New-HNSHostname -TemplateName "Epiroc VM Standard" -Parameters @{
    unit_code = 'SFD'
    provider = 'M'      # Microsoft Azure
    region = 'WEEU'     # West Europe
    environment = 'P'   # Production
    function = 'AS'     # Application Server
} -Reserve

# Output: SFDSMWEEUPAS001
```

## Module Structure

```
HNS-Standalone/
├── Source/
│   ├── Classes/          # PowerShell classes
│   ├── Enums/           # Enumerations
│   ├── Private/         # Internal functions
│   └── Public/          # Exported functions
├── Tests/               # Pester tests
├── Build/              # Build scripts
├── docs/               # Documentation
└── HNS-Standalone.psd1 # Module manifest
```

## Building from Source

```powershell
# Clone the repository
git clone https://github.com/yourorg/HNS-Standalone.git
cd HNS-Standalone

# Run build script
.\Build\build.ps1 -Task Package

# Output will be in the Output directory
```

## Key Commands

### Templates
- `New-HNSTemplate` - Create a new template
- `Get-HNSTemplate` - List templates
- `Set-HNSTemplate` - Update template
- `Remove-HNSTemplate` - Delete template
- `Import-HNSTemplate` - Import from file
- `Export-HNSTemplate` - Export to file

### Hostnames
- `New-HNSHostname` - Generate hostname
- `Get-HNSHostname` - List hostnames
- `Reserve-HNSHostname` - Reserve hostname
- `Commit-HNSHostname` - Mark as deployed
- `Release-HNSHostname` - Mark as decommissioned
- `Test-HNSDns` - Verify DNS record

### Configuration
- `Initialize-HNSEnvironment` - Setup module
- `Get-HNSConfiguration` - View settings
- `Set-HNSConfiguration` - Update settings
- `Backup-HNSData` - Create backup
- `Restore-HNSData` - Restore from backup

## Data Storage

By default, data is stored in:
- Windows: `%APPDATA%\HNS-Standalone`
- Linux/macOS: `~/.config/HNS-Standalone`

The module uses JSON files for:
- `config.json` - Module configuration
- `templates.json` - Template definitions
- `hostnames.json` - Generated hostnames
- `audit.log` - Audit trail

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests: `.\Build\build.ps1 -Task Test`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- Documentation: [Wiki](https://github.com/yourorg/HNS-Standalone/wiki)
- Issues: [GitHub Issues](https://github.com/yourorg/HNS-Standalone/issues)
- Discussions: [GitHub Discussions](https://github.com/yourorg/HNS-Standalone/discussions)