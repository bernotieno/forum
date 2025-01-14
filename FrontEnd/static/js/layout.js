logoutButtton = document.getElementById('logoutButton');

console.log('Logout button:', logoutButtton);
// Add an event listener to the logout button
logoutButtton.addEventListener('click', function () {
    // Retrieve the CSRF token (if stored in a meta tag or elsewhere)
    const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
 
     console.log('Logout button clicked');
    // Send a POST request to /logout with credentials (cookies)
    fetch('/logout', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfToken, // Include CSRF token in headers
        },
        credentials: 'include', // Include cookies in the request
    })
    .then(response => {
        console.log('Logout response:', response);
        if (response.ok) {
            // Redirect to the login page or home page after successful logout
            window.location.href = '/';
        } else {
            console.error('Logout failed');
        }
    })
    .catch(error => {
        console.error('Error during logout:', error);
    });
 });
    