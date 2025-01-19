document.addEventListener('DOMContentLoaded', function() {
    // Add event listeners to all like and dislike buttons
    const postLikeButtons = document.querySelectorAll('[id="Like"]:not(.comment-vote)');
    const postDislikeButtons = document.querySelectorAll('[id="DisLike"]:not(.comment-vote)');
    const commentVoteButtons = document.querySelectorAll('.comment-vote');

    // Check authentication status once at load
    const isAuthenticated = document.querySelector('.comment-input-container') !== null;

    // Post vote event listeners
    postLikeButtons.forEach(button => {
        button.addEventListener('click', handleVote('like'));
    });

    postDislikeButtons.forEach(button => {
        button.addEventListener('click', handleVote('dislike'));
    });

    // Comment vote event listeners
    commentVoteButtons.forEach(button => {
        button.addEventListener('click', function(event) {
            event.stopPropagation(); // Prevent event bubbling
            handleCommentVote(event, isAuthenticated);
        });
    });

    // Initialize vote states on page load
    initializeVoteStates();
});

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

// Comment vote handler function
async function handleCommentVote(event, isAuthenticated) {
    event.preventDefault();
    
    if (!isAuthenticated) {
        showToast('Please log in to vote');
        return;
    }
    
    const button = event.currentTarget;
    const commentId = button.getAttribute('data-comment-id');
    const voteType = button.getAttribute('data-vote') === 'up' ? 'like' : 'dislike';
    
    const formData = new URLSearchParams();
    formData.append('comment_id', commentId);
    formData.append('vote', voteType);
    
    try {
        const response = await fetch('/commentVote', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'X-CSRF-Token': getCSRFToken()
            },
            body: formData
        });
        
        if (response.ok) {
            const data = await response.json();
            document.getElementById(`comment-likes-${commentId}`).textContent = data.likes;
            document.getElementById(`comment-dislikes-${commentId}`).textContent = data.dislikes;
            toggleCommentButtonStates(commentId, voteType);
            showToast('Vote recorded successfully');
        } else {
            showToast('Failed to vote');
        }
    } catch (error) {
        console.error('Error:', error);
        showToast('An error occurred');
    }
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
        } else {
            dislikeButton.classList.add('dactive');
        }
    }
}

// Function to toggle comment button states
function toggleCommentButtonStates(commentId, activeVoteType) {
    const upButton = document.querySelector(`[data-vote="up"][data-comment-id="${commentId}"]`);
    const downButton = document.querySelector(`[data-vote="down"][data-comment-id="${commentId}"]`);

    if (upButton && downButton) {
        // Remove all active classes first
        upButton.classList.remove('active');
        downButton.classList.remove('dactive');

        // Add appropriate active class based on vote type
        if (activeVoteType === 'like') {
            upButton.classList.add('active');
        } else if (activeVoteType === 'dislike') {
            downButton.classList.add('dactive');
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

// Fetch user's comment votes
async function fetchUserCommentVotes() {
    try {
        const response = await fetch('/getUserCommentVotes', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (response.ok) {
            const data = await response.json();
            return data;
        } else {
            console.error('Failed to fetch user comment votes');
            return {};
        }
    } catch (error) {
        console.error('Error fetching user comment votes:', error);
        return {};
    }
}

// Initialize button states for all posts
async function initializeVoteStates() {
    try {
        const userVotes = await fetchUserVotes();
        Object.entries(userVotes).forEach(([postId, voteType]) => {
            toggleButtonStates(postId, voteType);
        });

        const userCommentVotes = await fetchUserCommentVotes();
        Object.entries(userCommentVotes).forEach(([commentId, voteType]) => {
            toggleCommentButtonStates(commentId, voteType);
        });
    } catch (error) {
        console.error('Error initializing vote states:', error);
    }
}

// Disable vote buttons if the user is not logged in
function disableVoteButtons() {
    const voteButtons = document.querySelectorAll('.vote-button, .comment-vote');
    voteButtons.forEach(button => {
        button.disabled = true;
        button.title = "You must be logged in to vote.";
    });
}

// Enable vote buttons if the user is logged in
function enableVoteButtons() {
    const voteButtons = document.querySelectorAll('.vote-button, .comment-vote');
    voteButtons.forEach(button => {
        button.disabled = false;
        button.title = "";
    });
}

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