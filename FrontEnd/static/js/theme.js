// theme-manager.js - Include this file in all your pages

class ThemeManager {
    constructor() {
      this.theme = localStorage.getItem('theme') || 'light';
      this.toggleButton = document.getElementById('theme-toggle');
      
      // Initialize theme immediately before page loads completely
      this.initializeTheme();
      
      // If toggle button exists on this page, set it up
      if (this.toggleButton) {
        this.setupToggleButton();
      }
      
      // Listen for theme changes from other tabs/windows
      this.setupStorageListener();
    }
  
    initializeTheme() {
      // Apply theme immediately to prevent flash of wrong theme
      document.documentElement.setAttribute('data-theme', this.theme);
      
      // Add class to body for additional theme-specific styles
      document.body.classList.remove('light-theme', 'dark-theme');
      document.body.classList.add(`${this.theme}-theme`);
    }
  
    setupToggleButton() {
      const iconElement = this.toggleButton.querySelector('i');
      const textElement = this.toggleButton.querySelector('span');
  
      // Update button appearance based on current theme
      this.updateToggleButton(iconElement, textElement);
  
      // Add click handler
      this.toggleButton.addEventListener('click', () => {
        this.toggleTheme();
        this.updateToggleButton(iconElement, textElement);
      });
    }
  
    updateToggleButton(iconElement) {
      if (!iconElement) return;
    
      if (this.theme === 'dark') {
        iconElement.classList.remove('fa-moon');
        iconElement.classList.add('fa-sun');
      } else {
        iconElement.classList.remove('fa-sun');
        iconElement.classList.add('fa-moon');
      }
    }
  
    toggleTheme() {
      this.theme = this.theme === 'light' ? 'dark' : 'light';
      localStorage.setItem('theme', this.theme);
      this.initializeTheme();
    }
  
    setupStorageListener() {
      // Listen for theme changes from other tabs/windows
      window.addEventListener('storage', (event) => {
        if (event.key === 'theme') {
          this.theme = event.newValue;
          this.initializeTheme();
          
          // Update toggle button if it exists on this page
          if (this.toggleButton) {
            this.updateToggleButton(
              this.toggleButton.querySelector('i'),
              this.toggleButton.querySelector('span')
            );
          }
        }
      });
    }
  }
  
  // Prevent flash of wrong theme by adding this script in the <head> of each page
  const preloadTheme = `
    (function() {
      const theme = localStorage.getItem('theme') || 'light';
      document.documentElement.setAttribute('data-theme', theme);
      document.documentElement.classList.add('theme-loaded');
    })();
  `;
  
  // Add preload script to head
  const script = document.createElement('script');
  script.textContent = preloadTheme;
  document.head.appendChild(script);
  
  // Add necessary styles
  const themeStyles = document.createElement('style');
  themeStyles.textContent = `
    /* Prevent flash of wrong theme */
    html:not(.theme-loaded) {
      visibility: hidden;
    }
  
    /* Theme toggle button styles */
    #theme-toggle {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 16px;
      border-radius: 20px;
      cursor: pointer;
      transition: all 0.3s ease;
      background: var(--bg-primary);
      border: 1px solid var(--border-color);
      color: var(--text-primary);
    }
  
    #theme-toggle:hover {
      background: var(--hover-bg);
      transform: translateY(-1px);
    }
  
    #theme-toggle i {
      font-size: 16px;
    }
  `;
  document.head.appendChild(themeStyles);
  
  // Initialize theme manager when DOM is ready
  document.addEventListener('DOMContentLoaded', () => {
    new ThemeManager();
  });