document.addEventListener('DOMContentLoaded', function () {
    // Handle voting
    document.querySelectorAll('.vote-button').forEach(button => {
        button.addEventListener('click', async function () {
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

    // Function to update vote count
    function updateVoteCount(postId, newCount) {
        const voteCount = document.querySelector(`.vote-count[data-post-id="${postId}"]`);
        if (voteCount) {
            voteCount.textContent = newCount;
        }
    }

    // Function to show toast messages
    function showToast(message) {
        const toast = document.getElementById('toast');
        const toastMessage = document.getElementById('toastMessage');
        toastMessage.textContent = message;
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), 3000);
    }

    // Function to format time ago
    function formatTimeAgo(timestamp) {
        const now = new Date();
        const postDate = new Date(timestamp);
        const duration = now - postDate; // Duration in milliseconds

        const seconds = Math.floor(duration / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (seconds < 60) {
            return "just now";
        } else if (minutes < 60) {
            return `${minutes} min${minutes === 1 ? "" : "s"} ago`;
        } else if (hours < 24) {
            return `${hours} hour${hours === 1 ? "" : "s"} ago`;
        } else if (days < 30) {
            return `${days} day${days === 1 ? "" : "s"} ago`;
        } else {
            return postDate.toLocaleDateString("en-US", {
                year: "numeric",
                month: "short",
                day: "numeric",
            });
        }
    }

    // Function to update all timestamps on the page
    function updateTimestamps() {
        document.querySelectorAll('.timestamp').forEach((element) => {
            const timestamp = element.getAttribute('data-timestamp');
            if (timestamp) {
                element.textContent = formatTimeAgo(timestamp);
            }
        });
    }

    // Update timestamps every minute
    setInterval(updateTimestamps, 60000);

    // Initial update of timestamps
    updateTimestamps();
});