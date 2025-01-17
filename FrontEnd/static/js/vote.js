document.addEventListener('DOMContentLoaded', function() {
    // Add event listeners to all like and dislike buttons
    const likeButtons = document.querySelectorAll('[id="Like"]');
    const dislikeButtons = document.querySelectorAll('[id="DisLike"]');

    likeButtons.forEach(button => {
        button.addEventListener('click', handleVote('like'));
    });

    dislikeButtons.forEach(button => {
        button.addEventListener('click', handleVote('dislike'));
    });
}
);

// Function to get CSRF token - add flexibility in how we find it
function getCSRFToken() {
    // Try different ways to find the CSRF token
    const csrfElement = 
        document.querySelector('meta[name="csrf-token"]') 
    
    if (csrfElement) {
        return csrfElement.content 
    }
    
    // If no token found, return null or an empty string
    return '';
}

// Generic vote handler function
function handleVote(voteType) {
    return async function(event) {
        event.preventDefault();
        
        const button = event.currentTarget;
        const postId = button.getAttribute('data-postId');

        // Get CSRF token
        const csrfToken = getCSRFToken();
        
       
        try {
            const response = await fetch('/likePost', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'X-CSRF-Token': csrfToken
                },
                body: new URLSearchParams({
                    'post_id': postId,
                    'vote': voteType
                })
            });
        
        
        if (response.ok) {
            const data = await response.json();
            const likesContainer = document.getElementById(`likes-container-${postId}`);
            const dislikesContainer = document.getElementById(`dislikes-container-${postId}`);
            if (likesContainer) {
                likesContainer.textContent = data.likes || '0';
            }
            if (dislikesContainer) {
                dislikesContainer.textContent = data.dislikes || '0';
            }
            showToast(`Post ${voteType}d successfully!`);
            toggleButtonStates(postId, voteType);
        }
        } catch (error) {
            console.error('Error:', error);
            showToast(`An error occurred while ${voteType}ing the post`);
        }
    };
}



// Function to toggle button states
function toggleButtonStates(postId, activeVoteType) {
    const likeButton = document.querySelector(`[id="Like"][data-postId="${postId}"]`);
    const dislikeButton = document.querySelector(`[id="DisLike"][data-postId="${postId}"]`);

    if (likeButton && dislikeButton) {
        // Remove active class from both buttons
        likeButton.classList.remove('active');
        dislikeButton.classList.remove('dactive');

        // Add active class to the clicked button
        if (activeVoteType === 'like') {
            likeButton.classList.add('active');
        } else  {
            dislikeButton.classList.add('dactive');
        }
    }
}

// Toast notification function
function showToast(message) {
    let toast = document.getElementById('toast');
    if (!toast) {
        toast = document.createElement('div');
        toast.id = 'toast';
        document.body.appendChild(toast);
    }

    toast.textContent = message;
    toast.classList.add('show');

    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

// Fetch the user's vote status for all posts
async function fetchUserVotes() {
    try {
        const response = await fetch('/getUserVotes', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (response.ok) {
            const data = await response.json();
            return data; 
        } else {
            console.error('Failed to fetch user votes');
            return {};
        }
    } catch (error) {
        console.error('Error fetching user votes:', error);
        return {};
    }
}



// Initialize button states for all posts
async function initializeButtonStates() {
    const userVotes = await fetchUserVotes();
    Object.keys(userVotes).forEach(postId => {
        const activeVoteType = userVotes[postId];
        toggleButtonStates(postId, activeVoteType);
    });
}

// Call the initialization function when the page loads
document.addEventListener('DOMContentLoaded', initializeButtonStates);


// Check if the user is logged in
async function checkLoginStatus() {
    try {
        const response = await fetch('/checkLoginStatus', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (response.ok) {
            const data = await response.json();
            return data.loggedIn; // Example: { loggedIn: true }
        } else {
            console.error('Failed to check login status');
            return false;
        }
    } catch (error) {
        console.error('Error checking login status:', error);
        return false;
    }
}

// Disable vote buttons if the user is not logged in
function disableVoteButtons() {
    const voteButtons = document.querySelectorAll('.vote-button');
    voteButtons.forEach(button => {
        button.disabled = true;
        button.title = "You must be logged in to vote."; // Add a tooltip
    });
}

// Enable vote buttons if the user is logged in
function enableVoteButtons() {
    const voteButtons = document.querySelectorAll('.vote-button');
    voteButtons.forEach(button => {
        button.disabled = false;
        button.title = ""; // Remove the tooltip
    });
}

// Initialize button states on page load
document.addEventListener('DOMContentLoaded', async () => {
    const loggedIn = await checkLoginStatus();
    if (!loggedIn) {
        disableVoteButtons();
    } else {
        enableVoteButtons();
    }
});