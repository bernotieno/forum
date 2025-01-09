document.addEventListener('DOMContentLoaded', function() {
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

    function updateVoteCount(postId, newCount) {
        const voteCount = document.querySelector(`.vote-count[data-post-id="${postId}"]`);
        if (voteCount) {
            voteCount.textContent = newCount;
        }
    }

    function showToast(message) {
        const toast = document.getElementById('toast');
        const toastMessage = document.getElementById('toastMessage');
        toastMessage.textContent = message;
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), 3000);
    }
}); 