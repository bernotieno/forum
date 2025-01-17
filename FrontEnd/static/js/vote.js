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

        console.log("==CSRF Token:", csrfToken);
        
        // Prepare headers
        const headers = {
            'Content-Type': 'application/x-www-form-urlencoded'
        };

        // Only add CSRF token if it exists
        if (csrfToken) {
            headers['X-CSRF-Token'] = csrfToken;
        }

        try {
            const response = await fetch('/likePost', {
                method: 'POST',
                headers: headers,
                body: new URLSearchParams({
                    'post_id': postId,
                    'vote': voteType
                })
            });

            if (response.ok) {
                const data = await response.json();
                // Update the likes counter
                const likesContainer = document.getElementById(`likes-container-${postId}`);
                if (likesContainer) {
                    likesContainer.textContent = data.likes || '0';
                }
                showToast(`Post ${voteType}d successfully!`);
                
                // Toggle active state of buttons
                toggleButtonStates(postId, voteType);
            } else {
                const errorData = await response.json();
                showToast(errorData.message || `Failed to ${voteType} post`);
            }
        } catch (error) {
            console.error('Error:', error);
            showToast(`An error occurred while ${voteType}ing the post`);
        }
    };
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

// Function to toggle button states
function toggleButtonStates(postId, activeVoteType) {
    const likeButton = document.querySelector(`[id="Like"][data-postId="${postId}"]`);
    const dislikeButton = document.querySelector(`[id="DisLike"][data-postId="${postId}"]`);

    if (likeButton && dislikeButton) {
        likeButton.classList.remove('active');
        dislikeButton.classList.remove('active');

        if (activeVoteType === 'like') {
            likeButton.classList.add('active');
        } else {
            dislikeButton.classList.add('active');
        }
    }
}

// Function to toggle button states
function toggleButtonStates(postId, activeVoteType) {
    const likeButton = document.querySelector(`[id="Like"][data-postId="${postId}"]`);
    const dislikeButton = document.querySelector(`[id="DisLike"][data-postId="${postId}"]`);

    if (likeButton && dislikeButton) {
        // Remove active class from both buttons
        likeButton.classList.remove('active');
        dislikeButton.classList.remove('active');

        // Add active class to the clicked button
        if (activeVoteType === 'like') {
            likeButton.classList.add('active');
        } else {
            dislikeButton.classList.add('active');
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

// Add CSS for styling
// const style = document.createElement('style');
// style.textContent = `
//     .vote-button {
//         cursor: pointer;
//         padding: 5px 10px;
//         margin: 0 5px;
//         border: 1px solid #ddd;
//         background: none;
//         transition: all 0.3s ease;
//     }

//     .vote-button.active {
//         background-color: #e0e0e0;
//         border-color: #999;
//     }

//     .counter {
//         display: inline-block;
//         margin: 0 10px;
//         font-weight: bold;
//     }

//     #toast {
//         visibility: hidden;
//         min-width: 250px;
//         margin-left: -125px;
//         background-color: #333;
//         color: #fff;
//         text-align: center;
//         border-radius: 2px;
//         padding: 16px;
//         position: fixed;
//         z-index: 1;
//         left: 50%;
//         bottom: 30px;
//         font-size: 14px;
//     }

//     #toast.show {
//         visibility: visible;
//         animation: fadein 0.5s, fadeout 0.5s 2.5s;
//     }

//     @keyframes fadein {
//         from {bottom: 0; opacity: 0;}
//         to {bottom: 30px; opacity: 1;}
//     }

//     @keyframes fadeout {
//         from {bottom: 30px; opacity: 1;}
//         to {bottom: 0; opacity: 0;}
//     }
// `;
// document.head.appendChild(style);