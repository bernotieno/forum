homepage.html:
{{define "title"}}Home - ThreadHub{{end}}
{{define "content"}}
<div class="posts-container">
    {{range .Posts}}
    <div class="post" data-category="{{.Category}}">
        <div class="post-header">
            <div class="post-info">
                <div class="post-meta">
                    <div class="post-author-info">
                        <div class="author-initial">{{slice .Author 0 1}}</div>
                        <span class="post-author">{{.Author}}</span>
                    </div>
                    <span class="timestamp" data-timestamp="{{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}"></span>
                    <span class="post-category">{{.Category}}</span>                 
                {{if .IsAuthor}}
                    <div class="post-options">
                        <button class="options-btn">
                            <i class="fa-solid fa-ellipsis"></i>
                        </button>
                        <div class="options-menu">
                            <button class="option-item" onclick="editPost('{{.ID}}')">
                                <i class="fa-solid fa-edit"></i> Edit
                            </button>
                            <button class="option-item" onclick="deletePost('{{.ID}}')">
                                <i class="fa-solid fa-trash"></i> Delete
                            </button>
                        </div>
                    </div>
                    {{end}}
                    
                </div>
                <h3 class="post-title">
                    <a href="/viewPost?id={{.ID}}">{{.Title}}</a>
                </h3>
            </div>
        </div>
        <div class="post-content">{{.Content}}</div>
        {{if .ImageUrl.Valid}}
        <div class="post-image">
            <img src="{{.ImageUrl.String}}" alt="Post image" loading="lazy">
        </div>
        {{end}}
        <div class="post-footer">
            <div class="vote-buttons">
                <button class="vote-button" onclick="AddLike('{{.ID}}')">
                    <i class="fa-regular fa-thumbs-up"></i>
                </button>
                <div class="counter" id="likes-container-{{.ID}}">0</div>
                <button class="vote-button" onclick="MinusLike('{{.ID}}')">
                    <i class="fa-regular fa-thumbs-down"></i>
                </button>
                <div class="counter" id="dislikes-container-{{.ID}}">0</div>
            </div>
            <div class="comments-count">
                <a href="/viewPost?id={{.ID}}#commentText">
                    <i class="fa-regular fa-comment"></i>
                    <span class="counter">{{len .Comments}}</span>
                </a>
            </div>
        </div>
    </div>
    {{end}}
</div>
{{end}}
{{define "scripts"}}
<script src="https://cdnjs.cloudflare.com/ajax/libs/feather-icons/4.29.0/feather.min.js"></script>
<script src="../static/js/homepage.js"></script>
<script src="../static/js/theme.js"></script>
<script src="../static/js/vote.js"></script>
{{end}} 

viewPost.html:
{{define "title"}}View Post - ThreadHub{{end}}
{{define "content"}}
<div class="posts-container">
    <div class="post">
        <div class="post-header">
            <div class="post-info">
                <div class="post-meta">
                    <div class="post-author-info">
                        <div class="author-initial">{{slice .Post.Author 0 1}}</div>
                        <span class="post-author">{{.Post.Author}}</span>
                    </div>
                    <span class="timestamp" data-timestamp="{{.Post.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}"></span>
                    <span class="post-category">{{.Post.Category}}</span>
                    {{if .IsAuthenticated}}
                        {{if .IsAuthor}}
                        <div class="post-options">
                            <button class="options-btn">
                                <i class="fa-solid fa-ellipsis"></i>
                            </button>
                            <div class="options-menu">
                                <button class="option-item" onclick="editPost('{{.Post.ID}}')">
                                    <i class="fa-solid fa-edit"></i> Edit
                                </button>
                                <button class="option-item" onclick="deletePost('{{.Post.ID}}')">
                                    <i class="fa-solid fa-trash"></i> Delete
                                </button>
                            </div>
                        </div>
                        {{end}}
                    {{end}}

                </div>
                <h3 class="post-title">{{.Post.Title}}</h3>
            </div>
        </div>
        <div class="post-content">{{.Post.Content}}</div>
        {{if .Post.ImageUrl.Valid}}
        <div class="post-image">
            <img src="{{.Post.ImageUrl.String}}" alt="Post image" loading="lazy">
        </div>
        {{end}}
        <div class="post-footer">
            <div class="footer-icons">
                <div class="vote-buttons">
                    <button class="vote-button" onclick="AddLike('{{.Post.ID}}')">
                        <i class="fa-regular fa-thumbs-up"></i>
                    </button>
                    <div class="counter" id="likes-container-{{.Post.ID}}">{{.Post.Likes}}</div>
                    <button class="vote-button" onclick="MinusLike('{{.Post.ID}}')">
                        <i class="fa-regular fa-thumbs-down"></i>
                    </button>
                    <div class="counter" id="dislikes-container-{{.Post.ID}}">0</div>
                </div>
                <div class="comments-count">
                    <i class="fa-regular fa-comment"></i>
                    <span class="counter">{{len .Comments}}</span>
                </div>
            </div>

            <!-- Comments section -->
            <div class="comments-section">
                {{if .IsAuthenticated}}
                <div class="comment-input-container">
                    <div class="textarea-container">
                        <textarea class="main-comment-input" placeholder="Write a comment..." id="commentText"></textarea>
                        <button class="button button-primary comment-button" data-post-id="{{.Post.ID}}" onclick="submitComment(this)">Comment</button>
                    </div>
                    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
                </div>
                {{else}}
                <p class="login-prompt">Please <a href="/login_Page">login</a> to comment</p>
                {{end}}

                <div class="comments-container">
                    {{template "comments" dict "Comments" .Comments "IsAuthenticated" .IsAuthenticated "Post" .Post}}
                </div>
            </div>
        </div>
    </div>
</div>

<div id="toast" class="toast">
    <div id="toastMessage" class="toast-message"></div>
</div>
{{end}}

{{define "comments"}}
    {{range $comment := .Comments}}
    <div class="comment" data-comment-id="{{$comment.ID}}">
        <div class="comment-header">
            <div class="post-author-info">
                <div class="author-initial">{{slice $comment.Author 0 1}}</div>
                <span class="comment-author">{{$comment.Author}}</span>
            </div>
            <span class="timestamp" data-timestamp="{{$comment.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}"></span>
        </div>
        <div class="comment-content">{{$comment.Content}}</div>
        <div class="comment-footer">
            <div class="vote-buttons">
                <button class="vote-button">
                    <i class="fa-regular fa-thumbs-up"></i>
                </button>
                <div class="counter">0</div>
                <button class="vote-button">
                    <i class="fa-regular fa-thumbs-down"></i>
                </button>
                <div class="counter">0</div>
            </div>
            {{if $.IsAuthenticated}}
            <div class="comment-actions">
                <button class="reply-button">Reply</button>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}
{{end}}

{{define "scripts"}}
<script src="../static/js/viewPost.js"></script>
<script src="../static/js/theme.js"></script>
{{end}}

viewpost.js:
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

    const csrfTokenElement = document.querySelector('input[name="csrf_token"]');
    if (!csrfTokenElement) {
        console.error('CSRF token not found');
        showToast('An error occurred. Please try again.');
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
        console.error('Error occurred while posting the comment:', error);
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

homepage.js:
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