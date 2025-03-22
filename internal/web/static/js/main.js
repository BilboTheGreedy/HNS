// Main JavaScript functions for HNS application

document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    if (tooltipTriggerList.length) {
        [...tooltipTriggerList].map(el => new bootstrap.Tooltip(el));
    }

    // Initialize popovers
    const popoverTriggerList = document.querySelectorAll('[data-bs-toggle="popover"]');
    if (popoverTriggerList.length) {
        [...popoverTriggerList].map(el => new bootstrap.Popover(el));
    }

    // Initialize template-related forms
    initTemplateForms();
    
    // Initialize hostname forms
    initHostnameForms();
    
    // Auto-dismiss alerts after 5 seconds
    setTimeout(function() {
        const alerts = document.querySelectorAll('.alert');
        alerts.forEach(function(alert) {
            // Create a new bootstrap alert and close it
            const bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        });
    }, 5000);
});

// Copy text to clipboard
function copyToClipboard(text, successMsg = "Copied to clipboard!") {
    // Use the clipboard API if available
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text)
            .then(() => showCopySuccess(successMsg))
            .catch(() => showCopyError());
    } else {
        // Fallback method
        const textArea = document.createElement("textarea");
        textArea.value = text;
        
        // Make the textarea invisible
        textArea.style.position = "fixed";
        textArea.style.left = "-999999px";
        textArea.style.top = "-999999px";
        document.body.appendChild(textArea);
        
        textArea.focus();
        textArea.select();
        
        try {
            document.execCommand('copy') ? 
                showCopySuccess(successMsg) : showCopyError();
        } catch (err) {
            showCopyError();
        }
        
        document.body.removeChild(textArea);
    }
}

// Show copy success message
function showCopySuccess(message = "Copied to clipboard!") {
    const alertDiv = document.createElement('div');
    alertDiv.className = 'alert alert-success position-fixed top-0 start-50 translate-middle-x mt-4 shadow';
    alertDiv.style.zIndex = '9999';
    alertDiv.innerHTML = `<i class="fas fa-check-circle me-2"></i> ${message}`;
    document.body.appendChild(alertDiv);
    
    // Remove the alert after 2 seconds
    setTimeout(function() {
        alertDiv.remove();
    }, 2000);
}

// Show copy error message
function showCopyError() {
    const alertDiv = document.createElement('div');
    alertDiv.className = 'alert alert-danger position-fixed top-0 start-50 translate-middle-x mt-4 shadow';
    alertDiv.style.zIndex = '9999';
    alertDiv.innerHTML = '<i class="fas fa-times-circle me-2"></i> Failed to copy to clipboard';
    document.body.appendChild(alertDiv);
    
    // Remove the alert after 2 seconds
    setTimeout(function() {
        alertDiv.remove();
    }, 2000);
}

// Initialize template-related forms
function initTemplateForms() {
    // Add group button for template forms
    const addGroupBtn = document.getElementById('add-group-btn');
    if (addGroupBtn) {
        let groupCounter = document.querySelectorAll('.template-group').length;
        
        addGroupBtn.addEventListener('click', function() {
            groupCounter++;
            
            const newGroup = document.createElement('div');
            newGroup.className = 'template-group card mb-3';
            newGroup.innerHTML = `
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                        <h5 class="mb-0">Group ${groupCounter}</h5>
                        <button type="button" class="btn btn-sm btn-outline-danger remove-group-btn">
                            <i class="fas fa-trash"></i> Remove
                        </button>
                    </div>
                    <div class="row">
                        <div class="col-md-6 mb-3">
                            <label for="group_name_${groupCounter}" class="form-label form-required">Group Name</label>
                            <input type="text" class="form-control" id="group_name_${groupCounter}" 
                                name="groups[${groupCounter-1}][name]" required>
                        </div>
                        <div class="col-md-3 mb-3">
                            <label for="group_length_${groupCounter}" class="form-label form-required">Length</label>
                            <input type="number" class="form-control" id="group_length_${groupCounter}" 
                                name="groups[${groupCounter-1}][length]" min="1" value="2" required>
                        </div>
                        <div class="col-md-3 mb-3">
                            <label for="group_required_${groupCounter}" class="form-label">Required</label>
                            <div class="form-check mt-2">
                                <input class="form-check-input" type="checkbox" id="group_required_${groupCounter}" 
                                    name="groups[${groupCounter-1}][is_required]" checked>
                                <label class="form-check-label" for="group_required_${groupCounter}">
                                    Required
                                </label>
                            </div>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-md-6 mb-3">
                            <label for="group_validation_type_${groupCounter}" class="form-label form-required">Validation Type</label>
                            <select class="form-select validation-type-select" id="group_validation_type_${groupCounter}" 
                                name="groups[${groupCounter-1}][validation_type]" required>
                                <option value="fixed">Fixed Value</option>
                                <option value="list" selected>List of Values</option>
                                <option value="regex">Regular Expression</option>
                                <option value="sequence">Sequence Number</option>
                            </select>
                        </div>
                        <div class="col-md-6 mb-3 validation-value-container">
                            <label for="group_validation_value_${groupCounter}" class="form-label validation-value-label form-required">Value</label>
                            <input type="text" class="form-control" id="group_validation_value_${groupCounter}" 
                                name="groups[${groupCounter-1}][validation_value]" placeholder="e.g. US,EU,AP" required>
                            <small class="validation-info text-muted">
                                For list values, enter comma-separated options (e.g., US,EU,AP)
                            </small>
                        </div>
                    </div>
                </div>
            `;
            
            document.getElementById('template-groups-container').appendChild(newGroup);
            
            // Add event listener to new remove button
            newGroup.querySelector('.remove-group-btn').addEventListener('click', function() {
                newGroup.remove();
            });
            
            // Add event listener to new validation type select
            newGroup.querySelector('.validation-type-select').addEventListener('change', handleValidationTypeChange);
        });
    }
    
    // Handle existing validation type selects
    document.querySelectorAll('.validation-type-select').forEach(select => {
        select.addEventListener('change', handleValidationTypeChange);
    });
    
    // Handle existing remove group buttons
    document.querySelectorAll('.remove-group-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            this.closest('.template-group').remove();
        });
    });
}

// Handle validation type change in template forms
function handleValidationTypeChange() {
    const validationType = this.value;
    const container = this.closest('.template-group').querySelector('.validation-value-container');
    const label = container.querySelector('.validation-value-label');
    const input = container.querySelector('input');
    const info = container.querySelector('.validation-info');
    
    if (validationType === 'sequence') {
        // Hide validation value for sequence type
        container.style.display = 'none';
        input.removeAttribute('required');
    } else {
        container.style.display = 'block';
        input.setAttribute('required', 'required');
        
        // Update label and info text based on validation type
        if (validationType === 'fixed') {
            label.textContent = 'Fixed Value';
            info.textContent = 'Enter the exact value for this group.';
            input.placeholder = 'e.g. PROD';
        } else if (validationType === 'list') {
            label.textContent = 'Allowed Values';
            info.textContent = 'Enter comma-separated list of allowed values (e.g., US,UK,DE,FR).';
            input.placeholder = 'e.g. US,EU,AP';
        } else if (validationType === 'regex') {
            label.textContent = 'Regex Pattern';
            info.textContent = 'Enter a valid regular expression pattern.';
            input.placeholder = 'e.g. [A-Z]{2}';
        }
    }
}

// Initialize hostname-related forms
function initHostnameForms() {
    // Template select change event for hostname generator and reservation
    const templateSelect = document.getElementById('template_id');
    if (templateSelect) {
        templateSelect.addEventListener('change', function() {
            const templateId = this.value;
            
            if (!templateId) {
                document.getElementById('template-params-container').innerHTML = '';
                return;
            }
            
            // Show loading indicator
            document.getElementById('template-params-container').innerHTML = '<div class="text-center my-4"><div class="spinner-border text-primary" role="status"><span class="visually-hidden">Loading...</span></div><p class="mt-2">Loading template parameters...</p></div>';
            
            // Fetch template details via AJAX
            fetch(`/api/templates/${templateId}`)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(template => {
                    let paramsHtml = '';
                    
                    // Generate form fields for each group
                    template.groups.forEach(group => {
                        // Skip sequence groups
                        if (group.validation_type === 'sequence') {
                            return;
                        }
                        
                        let inputType = 'text';
                        let inputHtml = '';
                        
                        if (group.validation_type === 'list') {
                            // Create select dropdown for list type
                            const options = group.validation_value.split(',').map(value => {
                                value = value.trim();
                                return `<option value="${value}">${value}</option>`;
                            }).join('');
                            
                            inputHtml = `
                                <select class="form-select template-param" id="param_${group.name}" 
                                    name="param_${group.name}" ${group.is_required ? 'required' : ''}>
                                    <option value="">Select ${group.name}</option>
                                    ${options}
                                </select>
                            `;
                        } else if (group.validation_type === 'fixed') {
                            // Fixed value should be readonly
                            inputHtml = `
                                <input type="text" class="form-control template-param" id="param_${group.name}" 
                                    name="param_${group.name}" value="${group.validation_value}" readonly>
                            `;
                        } else {
                            // Create text input for other types
                            inputHtml = `
                                <input type="${inputType}" class="form-control template-param" id="param_${group.name}" 
                                    name="param_${group.name}" maxlength="${group.length}" ${group.is_required ? 'required' : ''}>
                            `;
                        }
                        
                        paramsHtml += `
                            <div class="mb-3">
                                <label for="param_${group.name}" class="form-label ${group.is_required ? 'form-required' : ''}">
                                    ${group.name.charAt(0).toUpperCase() + group.name.slice(1)} (${group.length} chars)
                                </label>
                                ${inputHtml}
                            </div>
                        `;
                    });
                    
                    document.getElementById('template-params-container').innerHTML = paramsHtml;
                })
                .catch(error => {
                    console.error('Error fetching template:', error);
                    document.getElementById('template-params-container').innerHTML = `
                        <div class="alert alert-danger">
                            <i class="fas fa-exclamation-circle me-2"></i>
                            Error loading template parameters. Please try again.
                        </div>
                    `;
                });
        });
    }
}