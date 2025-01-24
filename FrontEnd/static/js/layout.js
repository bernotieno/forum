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
    // Select all posts
    const posts = document.querySelectorAll(".post");

    // Function to filter posts on the homepage
    const filterPosts = (selectedCategory) => {
        posts.forEach(post => {
            // Get the categories for the post and split into an array
            const postCategories = post.getAttribute("data-category")?.toLowerCase() || "";
            const categoriesArray = postCategories.split(",").map(cat => cat.trim());
            // Check if the selected category matches any in the array
            if (selectedCategory === "all" || selectedCategory === "home" || categoriesArray.includes(selectedCategory)) {
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
            toggleSidebar();

            // Get the clicked category name
            const selectedCategory = link.textContent.trim().toLowerCase();
            const currentPath = window.location.pathname;
            console.log(currentPath);

            if (currentPath === "/") {
                // If on the homepage, directly filter posts
                filterPosts(selectedCategory);
                history.pushState(null, "", "/");
            } else if (currentPath === "/viewPost" || currentPath === "/create-post") {
                // If on the viewPost page, redirect to the homepage
                sessionStorage.setItem("filterCategory", selectedCategory);
                window.location.href = "/";
            }
        });
    });

    // On page load, apply the filter if redirected
    const filterCategory = sessionStorage.getItem("filterCategory");
    if (filterCategory) {
        filterPosts(filterCategory);
        sessionStorage.removeItem("filterCategory");
    }
});




// Toggle the main dropdown on click
function toggleDropdown() {
    const dropdown = document.querySelector('.dropdown-content');
    dropdown.classList.toggle('hidden');
}

// Toggle the sub-dropdown on click
// Toggle the sub-dropdown and rotate the icon
function toggleSubDropdown(event) {
    event.preventDefault();
    const subDropdown = document.getElementById('myActivitiesDropdown');
    const activitiesIcon = document.getElementById('activitiesIcon');

    subDropdown.classList.toggle('hidden');
    activitiesIcon.classList.toggle('rotate');
}

// Close dropdowns when clicking outside
document.addEventListener('click', function(event) {
    const dropdown = document.querySelector('.dropdown-content');
    const profileImage = document.querySelector('.profile-image');
    
    // Only proceed if both elements exist
    if (profileImage && dropdown) {
        // Check if the click is outside the profile image and dropdown
        if (!profileImage.contains(event.target) && !dropdown.contains(event.target)) {
            dropdown.classList.add('hidden');
            const activitiesDropdown = document.getElementById('myActivitiesDropdown');
            if (activitiesDropdown) {
                activitiesDropdown.classList.add('hidden');
            }
        }
    }
});

function filterContent(type) {
    // Get the logged-in user's ID
    const userId = document.getElementById('userSection').getAttribute('data-user-id');

    const posts = document.querySelectorAll('.post');

    let itemsToFilter;


    switch (type) {
        case 'posts':
            itemsToFilter = posts;
            break;
        default:
            console.log('Invalid filter type');
            return;
    }

    // Filter and display the items
    itemsToFilter.forEach(item => {
        const itemUserId = item.getAttribute('data-post-user-id');
        if (itemUserId === userId) {
            item.style.display = "block";
        } else {
            item.style.display = "none";
        }
    });
}

// Display the filtered content in the DOM
function displayContent(data, type) {
    const contentContainer = document.getElementById('contentContainer');
    contentContainer.innerHTML = '';

    if (data.length === 0) {
        contentContainer.innerHTML = `<p>No ${type} found.</p>`;
        return;
    }

    data.forEach(item => {
        const itemElement = document.createElement('div');
        itemElement.classList.add('content-item');

        switch (type) {
            case 'posts':
                itemElement.innerHTML = `<p>${item.content}</p>`;
                break;
            case 'likes':
                itemElement.innerHTML = `<p>Liked Post ID: ${item.postId}</p>`;
                break;
            case 'comments':
                itemElement.innerHTML = `<p>${item.content}</p>`;
                break;
        }

        contentContainer.appendChild(itemElement);
    });
}

function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const hamburgerBtn = document.querySelector('.hamburger-menu');
    const overlay = document.querySelector('.sidebar-overlay');
    const isExpanded = hamburgerBtn.getAttribute('aria-expanded') === 'true';
    
    // Toggle ARIA attributes
    hamburgerBtn.setAttribute('aria-expanded', !isExpanded);
    
    // Toggle classes
    sidebar.classList.toggle('active');
    overlay?.classList.toggle('active');
    
    // Handle body scroll
    document.body.style.overflow = !isExpanded ? 'hidden' : '';
    
    // Handle escape key
    if (!isExpanded) {
        document.addEventListener('keydown', handleEscKey);
    } else {
        document.removeEventListener('keydown', handleEscKey);
    }
}

function handleEscKey(e) {
    if (e.key === 'Escape') {
        toggleSidebar();
    }
}

// Add overlay div to the DOM
document.body.insertAdjacentHTML('beforeend', '<div class="sidebar-overlay"></div>');

// Event Listeners
document.querySelector('.hamburger-menu').addEventListener('click', toggleSidebar);
document.querySelector('.sidebar-overlay').addEventListener('click', toggleSidebar);

// Handle click outside
document.addEventListener('click', (e) => {
    const sidebar = document.querySelector('.sidebar');
    const hamburgerBtn = document.querySelector('.hamburger-menu');
    
    if (sidebar.classList.contains('active') && 
        !sidebar.contains(e.target) && 
        !hamburgerBtn.contains(e.target)) {
        toggleSidebar();
    }
});