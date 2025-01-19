// Move showToast function outside DOMContentLoaded and make it globally available
window.showToast = function(message) {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toastMessage');
    toastMessage.textContent = message;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 3000);
};

document.addEventListener('DOMContentLoaded', function() {
    // Check if we came from homepage
    if (document.referrer.includes('/')) {
        sessionStorage.setItem('scrollPosition', window.scrollY);
    }

    // Make goBack function globally available
    window.goBack = function() {
        window.location.href = '/';
    };

    // Update timestamps
    function updateTimestamps() {
        document.querySelectorAll('.timestamp').forEach((element) => {
            const timestamp = element.getAttribute('data-timestamp');
            if (timestamp) {
                element.textContent = formatTimeAgo(timestamp);
            }
        });
    }

    function formatTimeAgo(timestamp) {
        const now = new Date();
        const postDate = new Date(timestamp);
        const duration = now - postDate;
        const seconds = Math.floor(duration / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (seconds < 60) return "just now";
        if (minutes < 60) return `${minutes} min${minutes === 1 ? "" : "s"} ago`;
        if (hours < 24) return `${hours} hour${hours === 1 ? "" : "s"} ago`;
        if (days < 30) return `${days} day${days === 1 ? "" : "s"} ago`;
        return postDate.toLocaleDateString("en-US", {
            year: "numeric",
            month: "short",
            day: "numeric"
        });
    }

    // Initial update of timestamps
    updateTimestamps();
    // Update timestamps every minute
    setInterval(updateTimestamps, 60000);

    // Add this at the beginning of DOMContentLoaded
    const isAuthenticated = document.querySelector('.comment-input-container') !== null;

    // Handle options menu
    document.querySelectorAll('.options-btn').forEach(button => {
        button.addEventListener('click', function(e) {
            e.stopPropagation(); // Prevent event from bubbling up
            const menu = this.nextElementSibling;
            
            // Close all other open menus first
            document.querySelectorAll('.options-menu.show').forEach(m => {
                if (m !== menu) {
                    m.classList.remove('show');
                }
            });
            
            // Toggle current menu
            menu.classList.toggle('show');
        });
    });

    // Close menu when clicking anywhere else on the page
    document.addEventListener('click', function(e) {
        if (!e.target.closest('.post-options')) {
            document.querySelectorAll('.options-menu.show').forEach(menu => {
                menu.classList.remove('show');
            });
        }
    });

    // comment button
    const commentInput = document.getElementById('commentText');
    const commentButton = document.querySelector('.comment-button');

    commentInput.addEventListener('input', function() {
        if (commentInput.value.trim() === '') {
            commentButton.classList.remove('active');
            commentButton.disabled = true;
        } else {
            commentButton.classList.add('active');
            commentButton.disabled = false;
        }
    });

    document.addEventListener('click', (e) => {
        if (!e.target.closest('.comment-options')) {
            document.querySelectorAll('.options-menu').forEach(menu => {
                menu.classList.remove('show');
            });
        }
    });

    // Hide reply buttons when max depth is reached
    document.querySelectorAll('.comment').forEach(comment => {
        // Calculate comment depth by counting parent comments
        let depth = 0;
        let parent = comment;
        while (parent.parentElement.closest('.comment')) {
            depth++;
            parent = parent.parentElement.closest('.comment');
        }
        
        // Hide reply button and form if max depth reached
        if (depth >= 4) {
            // Hide reply button
            const replyButton = comment.querySelector('.reply-button');
            if (replyButton) {
                replyButton.remove(); 
            }
            
            // Hide reply form container
            const replyForm = comment.querySelector('.reply-input-container');
            if (replyForm) {
                replyForm.remove(); 
            }
        }
    });
});

// Move submitComment outside DOMContentLoaded to make it globally available
window.submitComment = async function(button) {
    // Disable the button immediately to prevent double clicks
    button.disabled = true;
    
    const postId = button.getAttribute('data-post-id');
    const commentInput = document.getElementById('commentText');
    const content = commentInput.value.trim();
    
    if (!content) {
        showToast('Comment cannot be empty');
        button.disabled = false;  // Re-enable the button if validation fails
        return;
    }

    const csrfTokenElement = document.querySelector('input[name="csrf_token"]');
    if (!csrfTokenElement) {
        console.error('CSRF token not found');
        showToast('An error occurred. Please try again.');
        button.disabled = false;  // Re-enable the button if validation fails
        return;
    }

    try {
        const response = await fetch(`/comment/${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfTokenElement.value
            },
            body: JSON.stringify({ content })
        });

        if (response.ok) {
            // Update comment count
            const commentCountElement = document.querySelector('.comments-count .counter');
            const currentCount = parseInt(commentCountElement.textContent);
            commentCountElement.textContent = currentCount + 1;
            
            // Clear input and reset button state
            commentInput.value = '';
            button.classList.remove('active');
            
            // Reload page after everything is done
            window.location.reload();
        } else {
            const data = await response.json();
            showToast(data.error || 'Failed to post comment');
            button.disabled = false;  // Re-enable the button on error
        }
    } catch (error) {
        console.error('Error occurred while posting the comment:', error);
        showToast('An error occurred while posting the comment');
        button.disabled = false;  // Re-enable the button on error
    }
};

// Move these functions outside DOMContentLoaded and make them globally available
window.showReplyForm = function(button) {
    const commentId = button.getAttribute('data-comment-id');
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    
    // Hide all other reply forms first
    document.querySelectorAll('.reply-input-container').forEach(container => {
        if (container.id !== `reply-form-${commentId}`) {
            container.style.display = 'none';
        }
    });
    
    // Toggle visibility - if it's not 'block', make it 'block', otherwise 'none'
    replyForm.style.display = replyForm.style.display === 'block' ? 'none' : 'block';
};

window.cancelReply = function(button) {
    const commentId = button.getAttribute('data-comment-id');
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    const replyInput = document.getElementById(`replyText-${commentId}`);
    replyInput.value = '';
    replyForm.style.display = 'none';
};

window.submitReply = async function(button) {
    const commentId = button.getAttribute('data-comment-id');
    const postId = button.getAttribute('data-post-id');
    const content = document.getElementById(`replyText-${commentId}`).value.trim();
    
    if (!content) {
        showToast('Reply cannot be empty');
        return;
    }

    try {
        const response = await fetch(`/comment/${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
            },
            body: JSON.stringify({ 
                content: content,
                parentId: parseInt(commentId, 10)
            })
        });

        if (response.ok) {
            location.reload();
        } else {
            const data = await response.json();
            showToast(data.error || 'Failed to post reply');
        }
    } catch (error) {
        console.error('Error:', error);
        showToast('An error occurred while posting the reply');
    }
};

// Function to delete a comment
async function deleteComment(commentId) {
    if (!confirm('Are you sure you want to delete this comment?')) return;

    try {
        const response = await fetch(`/deleteComment?id=${commentId}`, {
            method: 'DELETE',
            headers: {
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
            }
        });

        if (response.ok) {
            // Find and remove the comment element
            const commentElement = document.querySelector(`[data-comment-id="${commentId}"]`);
            if (commentElement) {
                // Remove any replies to this comment as well
                const replies = commentElement.querySelectorAll('.comment');
                let totalRemoved = 1 + replies.length; // Count main comment + replies
                
                commentElement.remove();
                
                // Update the comment count in the UI
                const commentCountElement = document.querySelector('.comments-count .counter');
                if (commentCountElement) {
                    const currentCount = parseInt(commentCountElement.textContent);
                    commentCountElement.textContent = Math.max(0, currentCount - totalRemoved);
                }
                
                showToast('Comment deleted successfully');
            }
        } else {
            showToast('Failed to delete comment');
        }
    } catch (error) {
        console.error('Error:', error);
        showToast('An error occurred while deleting the comment');
    }
}

// Function to edit a comment
async function editComment(commentId) {
    const contentDiv = document.getElementById(`comment-content-${commentId}`);
    const currentContent = contentDiv.textContent;

    contentDiv.innerHTML = `
        <textarea class="edit-input" id="edit-${commentId}">${currentContent}</textarea>
        <div class="edit-buttons">
            <button class="button button-primary" id="save-edit-${commentId}">Save</button>
            <button class="button button-secondary" id="cancel-edit-${commentId}">Cancel</button>
        </div>
    `;

    // Add event listeners for the Save and Cancel buttons
    document.getElementById(`save-edit-${commentId}`).addEventListener('click', () => saveEdit(commentId));
    document.getElementById(`cancel-edit-${commentId}`).addEventListener('click', () => cancelEdit(commentId, currentContent));
}

// Function to save the edited comment
async function saveEdit(commentId) {
    const editedContent = document.getElementById(`edit-${commentId}`).value;

    try {
        const response = await fetch(`/updateComment?id=${commentId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
            },
            body: JSON.stringify({ content: editedContent })
        });

        if (response.ok) {
            // Update the comment content without page reload
            const contentDiv = document.getElementById(`comment-content-${commentId}`);
            contentDiv.textContent = editedContent;
            showToast('Comment updated successfully');
        } else {
            const data = await response.json();
            showToast(data.error || 'Failed to save the edited comment');
        }
    } catch (error) {
        console.error('Error:', error);
        showToast('An error occurred while saving the edited comment');
    }
}

// Function to cancel the edit and restore the original content
function cancelEdit(commentId, originalContent) {
    const contentDiv = document.getElementById(`comment-content-${commentId}`);
    contentDiv.textContent = originalContent;
}

// Add event listeners to all delete and edit buttons
document.addEventListener('DOMContentLoaded', () => {
    // Loop through all comments and attach event listeners
    document.querySelectorAll('.options-menu .option-item').forEach(button => {
        if (button.id.startsWith('delete-comment-')) {
            const commentId = button.id.replace('delete-comment-', '');
            button.addEventListener('click', () => deleteComment(commentId));
        } else if (button.id.startsWith('edit-comment-')) {
            const commentId = button.id.replace('edit-comment-', '');
            button.addEventListener('click', () => editComment(commentId));
        }
    });
});