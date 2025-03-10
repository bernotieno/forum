document.addEventListener('DOMContentLoaded', function() {
    // Tab functionality
    const tabButtons = document.querySelectorAll('.tab-button');
    const tabContents = document.querySelectorAll('.tab-content');
    
    function openTab(tabName) {
        // Hide all tab content
        tabContents.forEach(content => {
            content.classList.remove('active');
        });
        
        // Remove active class from all buttons
        tabButtons.forEach(button => {
            button.classList.remove('active');
        });
        
        // Show the selected tab content
        const selectedTab = document.getElementById(tabName);
        const selectedButton = document.querySelector(`.tab-button[data-tab="${tabName}"]`);
        
        if (selectedTab) {
            selectedTab.classList.add('active');
        }
        
        if (selectedButton) {
            selectedButton.classList.add('active');
        }
    }
    
    // Add click event to each tab button
    tabButtons.forEach(button => {
        button.addEventListener('click', function() {
            const tabName = this.getAttribute('data-tab');
            openTab(tabName);
        });
    });
    
    // Set default tab on page load
    // Check if there's already an active tab button
    const activeButton = document.querySelector('.tab-button.active');
    if (activeButton) {
        // If there is, open that tab
        const defaultTab = activeButton.getAttribute('data-tab');
        openTab(defaultTab);
    } else {
        // If no active tab is set, default to the first tab
        const firstTab = tabButtons[0]?.getAttribute('data-tab');
        if (firstTab) {
            openTab(firstTab);
        }
    }
}); 