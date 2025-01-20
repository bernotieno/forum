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
    // Select the posts container
    const postsContainer = document.querySelector(".posts-container");

    // Function to filter posts on the homepage
    const filterPosts = (selectedCategory) => {
        const posts = postsContainer.querySelectorAll(".post");
        posts.forEach(post => {
            const postCategory = post.getAttribute("data-category")?.toLowerCase() || "";
            if (selectedCategory === "all" || selectedCategory === "home" || postCategory.includes(selectedCategory)) {
                post.style.display = "block";
            } else {
                post.style.display = "none";
            }
        });
    };

    // Add click event listener to each sidebar link
    communityLinks.forEach(link => {
        link.addEventListener("click", (event) => {
            event.preventDefault();

            // Get the clicked category name
            const selectedCategory = link.textContent.trim().toLowerCase();
            const currentPath = window.location.pathname;

            // Check the current path
            if (currentPath === "/") {
                // If on the homepage, directly filter posts
                filterPosts(selectedCategory);
                // Optionally, update the browser history without query parameters
                history.pushState(null, "", "/");
            } else if (currentPath === "/viewPost") {
                // If on the viewPost page, redirect to the homepage
                sessionStorage.setItem("filterCategory", selectedCategory);
                window.location.href = "/";
            }
        });
    });

    // On homepage, check for saved category in sessionStorage
    if (window.location.pathname === "/") {
        const savedCategory = sessionStorage.getItem("filterCategory");
        if (savedCategory) {
            filterPosts(savedCategory);
            sessionStorage.removeItem("filterCategory");
        } else {
            filterPosts("all");
        }
    }
});


function toggleDropdown() {
    const dropdown = document.querySelector('.dropdown-content');
    if (dropdown.style.display === 'block') {
        dropdown.style.display = 'none';
    } else {
        dropdown.style.display = 'block';
    }
} 