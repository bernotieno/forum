document.addEventListener('DOMContentLoaded', function() {
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

    // Handle voting
    document.querySelectorAll('.vote-button').forEach(button => {
        button.addEventListener('click', async function() {
            if (!isAuthenticated) {
                showToast('Please log in to vote');
                return;
            }
            const postId = this.dataset.postId;
            const voteType = this.dataset.vote;
            try {
                const response = await fetch(`/vote`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
                    },
                    body: JSON.stringify({
                        postId: postId,
                        voteType: voteType
                    })
                });
                if (response.ok) {
                    const data = await response.json();
                    updateVoteCount(postId, data.likes);
                } else {
                    showToast('Failed to vote');
                }
            } catch (error) {
                showToast('An error occurred');
            }
        });
    });
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
});

function showToast(message) {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toastMessage');
    toastMessage.textContent = message;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 3000);
}

async function submitComment(button) {
    const postId = button.getAttribute('data-post-id');
    const content = document.getElementById('commentText').value.trim();
    if (!content) {
        showToast('Comment cannot be empty');
        return;
    }

    try {
        const response = await fetch(`/comment/${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
            },
            body: JSON.stringify({ content })
        });

        if (response.ok) {
            location.reload();
        } else {
            const data = await response.json();
            showToast(data.error || 'Failed to post comment');
        }
    } catch (error) {
        showToast('An error occurred while posting the comment');
    }
}

function showReplyForm(button) {
    const commentId = button.getAttribute('data-comment-id');
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    replyForm.style.display = replyForm.style.display === 'none' ? 'block' : 'none';
}

async function submitReply(button) {
    const parentId = button.getAttribute('data-comment-id');
    const postId = button.getAttribute('data-post-id');
    const content = document.getElementById(`replyText-${parentId}`).value.trim();
    
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
                content,
                parentId: parseInt(parentId, 10)
            })
        });

        if (response.ok) {
            location.reload();
        } else {
            const data = await response.json();
            showToast(data.error || 'Failed to post reply');
        }
    } catch (error) {
        showToast('An error occurred while posting the reply');
    }
} 