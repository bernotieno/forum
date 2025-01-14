const logoutButton = document.getElementById('logoutButton');

if (logoutButton) {
    logoutButton.addEventListener('click', function () {
        // Retrieve the CSRF token (if stored in a meta tag or elsewhere)
        const csrfMetaTag = document.querySelector('meta[name="csrf-token"]');
        const csrfToken = csrfMetaTag ? csrfMetaTag.getAttribute('content') : null;

        if (!csrfToken) {
            console.error('CSRF token not found.');
            return;
        }

        // Send a POST request to /logout with credentials (cookies)
        fetch('/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken, 
            },
            credentials: 'include', 
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
} else {
    console.log('Logout button not found. No event listener attached.');
}

document.addEventListener("DOMContentLoaded", () => {
    // Select all sidebar links
    const communityLinks = document.querySelectorAll(".sidebar .sidebar-link");
    // Select all posts in the main content area
    const posts = document.querySelectorAll(".main-content .post");

    // Add click event listener to each sidebar link
    communityLinks.forEach(link => {
        link.addEventListener("click", (event) => {
            event.preventDefault(); 

            // Get the clicked category name
            const selectedCategory = link.textContent.trim().toLowerCase();
    
            // Ensure 'data-category' exists and is normalized
            posts.forEach(post => {
                
                const postCategory = post.getAttribute("data-category")?.toLowerCase() || "";
                
                if (selectedCategory === "all" || postCategory === selectedCategory) {
                    post.style.display = "block"; 
                } else {
                    post.style.display = "none"; 
                }
            });
        });
    });

    // Show all posts by default on page load
    posts.forEach(post => (post.style.display = ""));
});

    