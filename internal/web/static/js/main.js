/* Main JavaScript for HNS */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
      return new bootstrap.Tooltip(tooltipTriggerEl);
    });
  
    // Initialize popovers
    const popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    popoverTriggerList.map(function (popoverTriggerEl) {
      return new bootstrap.Popover(popoverTriggerEl);
    });
  
    // Auto-dismiss alerts after 5 seconds
    setTimeout(function() {
      const alerts = document.querySelectorAll('.alert:not(.alert-permanent)');
      alerts.forEach(function(alert) {
        const bsAlert = new bootstrap.Alert(alert);
        bsAlert.close();
      });
    }, 5000);
  
    // Confirm delete actions
    document.querySelectorAll('.confirm-delete').forEach(function(button) {
      button.addEventListener('click', function(e) {
        if (!confirm('Are you sure you want to delete this item? This action cannot be undone.')) {
          e.preventDefault();
        }
      });
    });
  
    // Responsive tables - add data attributes for mobile view
    document.querySelectorAll('.table-responsive-cards').forEach(function(table) {
      const headerCells = table.querySelectorAll('thead th');
      const headerLabels = Array.from(headerCells).map(cell => cell.textContent.trim());
      
      table.querySelectorAll('tbody tr').forEach(function(row) {
        const dataCells = row.querySelectorAll('td');
        dataCells.forEach(function(cell, i) {
          if (headerLabels[i]) {
            cell.setAttribute('data-label', headerLabels[i]);
          }
        });
      });
    });
  
    // Disable form resubmission
    if (window.history.replaceState) {
      window.history.replaceState(null, null, window.location.href);
    }
  });
  
  // Function to copy text to clipboard
  function copyToClipboard(text, buttonElement) {
    navigator.clipboard.writeText(text).then(function() {
      // Visual feedback
      const originalText = buttonElement.innerHTML;
      buttonElement.innerHTML = '<i class="fas fa-check"></i> Copied!';
      setTimeout(function() {
        buttonElement.innerHTML = originalText;
      }, 2000);
    }).catch(function(err) {
      console.error('Could not copy text: ', err);
    });
  }
  
  // Confirmation modal handler
  function showConfirmationModal(title, message, confirmCallback) {
    const modalHtml = `
      <div class="modal fade" id="confirmationModal" tabindex="-1" aria-hidden="true">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">${title}</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>
            <div class="modal-body">
              ${message}
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
              <button type="button" class="btn btn-danger" id="confirmButton">Confirm</button>
            </div>
          </div>
        </div>
      </div>
    `;
    
    // Add modal to the DOM
    const modalContainer = document.createElement('div');
    modalContainer.innerHTML = modalHtml;
    document.body.appendChild(modalContainer);
    
    // Initialize the modal
    const modalElement = document.getElementById('confirmationModal');
    const modal = new bootstrap.Modal(modalElement);
    
    // Add confirm button event listener
    document.getElementById('confirmButton').addEventListener('click', function() {
      modal.hide();
      if (confirmCallback && typeof confirmCallback === 'function') {
        confirmCallback();
      }
    });
    
    // Show the modal
    modal.show();
    
    // Clean up when modal is hidden
    modalElement.addEventListener('hidden.bs.modal', function() {
      document.body.removeChild(modalContainer);
    });
  }
  
  // Form validation helpers
  function validateForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return false;
    
    let isValid = true;
    
    // Check required fields
    form.querySelectorAll('[required]').forEach(function(field) {
      if (!field.value.trim()) {
        isValid = false;
        field.classList.add('is-invalid');
      } else {
        field.classList.remove('is-invalid');
      }
    });
    
    return isValid;
  }