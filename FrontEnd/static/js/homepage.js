// Wait for DOM content and feather icons to load
document.addEventListener('DOMContentLoaded', function() {
    // Initialize Feather icons
    feather.replace();
    // Theme toggle functionality 
    window.toggleTheme = function() {
        const body = document.body;
        const currentTheme = body.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        body.setAttribute('data-theme', newTheme);
        
        const themeIcon = document.querySelector('.theme-toggle i');
        if (themeIcon) {
            themeIcon.setAttribute('data-feather', newTheme === 'dark' ? 'sun' : 'moon');
            feather.replace();
        }
        
        localStorage.setItem('theme', newTheme);
    }

    const showToast = (message) => {
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.classList.add('show');
        setTimeout(() => {
          toast.classList.remove('show');
        }, 3000);
    };
    
    let isLoggedIn = false;
    // Check initial login status
    async function initializeApp() {
        try {
            isLoggedIn = await checkLoginStatus();
            updateUserSection();
            renderPosts();
        } catch (error) {
            console.error('Error checking login status:', error);
            showToast('Error checking login status');
        }
    }
    
    initializeApp();
    

    // Set initial theme
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.body.setAttribute('data-theme', savedTheme);
    const initialThemeIcon = document.querySelector('.theme-toggle i');
    if (initialThemeIcon) {
        initialThemeIcon.setAttribute('data-feather', savedTheme === 'dark' ? 'sun' : 'moon');
    }

    // Sample post data
    // Helper function to format time difference
    function formatTimeAgo(timestamp) {
        const seconds = Math.floor((Date.now() - timestamp) / 1000);
        
        if (seconds < 60) {
            return `${seconds} seconds ago`;
        }
        
        const minutes = Math.floor(seconds / 60);
        if (minutes < 60) {
            return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`;
        }
        
        const hours = Math.floor(minutes / 60);
        if (hours < 24) {
            return `${hours} hour${hours !== 1 ? 's' : ''} ago`;
        }
        
        const days = Math.floor(hours / 24);
        return `${days} day${days !== 1 ? 's' : ''} ago`;
    }

    const posts = [
        {
            id: 1,
            title: "Check out this amazing project I built!",
            author: "dev_enthusiast",
            category: "programming", 
            likes: 1234,
            dislikes: 234,
            userVote: null,
            comments: [
                {
                    id: 1,
                    author: "coder123",
                    content: "This is amazing! How long did it take you to build?",
                    likes: 45,
                    dislikes: 5,
                    userVote: null,
                    timestamp: Date.now() - (3 * 60 * 60 * 1000), // 3 hours ago
                    replies: [
                        {
                            id: 2,
                            author: "dev_enthusiast",
                            content: "Thanks! It took about 2 months of work",
                            likes: 12,
                            dislikes: 2,
                            userVote: null,
                            timestamp: Date.now() - (2 * 60 * 60 * 1000) // 2 hours ago
                        }
                    ]
                },
                {
                    id: 3, 
                    author: "webdev_pro",
                    content: "The architecture looks solid. Have you considered adding...",
                    likes: 23,
                    dislikes: 3,
                    userVote: null,
                    timestamp: Date.now() - (60 * 60 * 1000), // 1 hour ago
                    replies: []
                }
            ],
            timestamp: Date.now() - (4 * 60 * 60 * 1000), // 4 hours ago
            content: "Just finished building a full-stack application using the latest technologies..."
        },
        {
            id: 2,
            title: "Announcing a new web framework",
            author: "tech_wizard",
            category: "technology",
            likes: 567,
            dislikes: 67,
            userVote: null,
            comments: [
                {
                    id: 4,
                    author: "framework_fan",
                    content: "Looks promising! What's the learning curve like?",
                    likes: 34,
                    dislikes: 4,
                    userVote: null,
                    timestamp: Date.now() - (30 * 60 * 1000), // 30 minutes ago
                    replies: []
                }
            ],
            timestamp: Date.now() - (45 * 60 * 1000), // 45 minutes ago
            content: "Introducing a revolutionary new framework for building web applications..."
        }
    ];

    // Render posts
    function renderPosts() {
        const container = document.getElementById('posts-container');
        container.innerHTML = ''; // Clear existing posts
        console.log("isLoggedIn", isLoggedIn)
        posts.forEach(post => {
            const postElement = document.createElement('div');
            postElement.className = 'post';
            postElement.setAttribute('data-post-id', post.id);
            postElement.innerHTML = `
                <div class="post-header">
                    <div class="post-votes">
                        <button class="vote-button ${post.userVote === 'up' ? 'active' : ''} ${!isLoggedIn ? 'disabled' : ''}" data-vote="up">
                            <i data-feather="thumbs-up"></i>
                            <span class="vote-count-up">${post.likes}</span>
                        </button>
                        
                        <button class="vote-button ${post.userVote === 'down' ? 'active' : ''} ${!isLoggedIn ? 'disabled' : ''}" data-vote="down">
                            <i data-feather="thumbs-down"></i>
                            <span class="vote-count-down">${post.dislikes}</span>
                        </button>
                    </div>
                    <div>
                        <small>r/${post.category} • Posted by u/${post.author} ${formatTimeAgo(post.timestamp)}</small>
                        <h2 class="post-title" data-post-id="${post.id}">${post.title}</h2>
                    </div>
                </div>
                <div class="post-content">
                    ${post.content}
                </div>
                <div class="post-footer">
                    <span class="comment-button" data-post-id="${post.id}">
                        <i data-feather="message-square"></i>
                        ${post.comments.length} Comments
                    </span>
                    <span class="reply-icon" data-post-id="${post.id}">
                        <i data-feather="message-circle"></i>
                        Reply
                    </span>
                </div>
            `;
            container.appendChild(postElement);

            // Add click handlers
            postElement.querySelector('.post-title').addEventListener('click', () => showComments(post.id));
            postElement.querySelector('.comment-button').addEventListener('click', () => showComments(post.id));
            postElement.querySelector('.reply-icon').addEventListener('click', () => showComments(post.id));
            
            
        });
        feather.replace();
    }

    async function showComments(postId) {
        const container = document.getElementById('posts-container');
        container.innerHTML = ''; // Clear container
        
        const post = posts.find(p => p.id === postId);
        if (!post) return;

        console.log("Show comments", isLoggedIn)

        // Render single post with comments
        const postElement = document.createElement('div');
        postElement.className = 'post';
        postElement.setAttribute('data-post-id', post.id);
        postElement.innerHTML = `
            <div class="post-header">
                <div class="post-votes">
                    ${isLoggedIn ? `
                        <button class="vote-button ${post.userVote === 'up' ? 'active' : ''}" data-vote="up">
                            <i data-feather="thumbs-up"></i>
                            <span class="vote-count-up">${post.likes}</span>
                        </button>
                       
                        <button class="vote-button ${post.userVote === 'down' ? 'active' : ''}" data-vote="down">
                            <i data-feather="thumbs-down"></i>
                            <span class="vote-count-down">${post.dislikes}</span>
                        </button>
                    ` : `
                        <div class="vote-button disabled">
                            <i data-feather="thumbs-up"></i>
                            <span class="vote-count-up">${post.likes}</span>
                        </div>
                       
                        <div class="vote-button disabled">
                            <i data-feather="thumbs-down"></i>
                            <span class="vote-count-down">${post.dislikes}</span>
                        </div>
                    `}
                </div>
                <div>
                    <small>r/${post.category} • Posted by u/${post.author} ${formatTimeAgo(post.timestamp)}</small>
                    <h2>${post.title}</h2>
                </div>
            </div>
            <div class="post-content">
                ${post.content}
            </div>
            <div class="comments-section">
                ${isLoggedIn ? `
                    <div class="comment-input">
                        <div class="comment-input-wrapper">
                            <textarea placeholder="Add a comment..." class="main-comment-input"></textarea>
                            <div class="comment-actions" style="display: none;">
                                <button class="button button-primary" onclick="window.submitReply(${post.id}, '.main-comment-input')">Comment</button>
                            </div>
                        </div>
        
                    </div>
                ` : `
                    <div class="login-prompt">
                        <p>Please <a href="/login_Page">login</a> to comment</p>
                    </div>
                `}
                <div class="comments-container">
                    ${renderComments(post.comments, isLoggedIn)}
                </div>
            </div>
        `;
        
        container.appendChild(postElement);
        feather.replace();
    }

    function renderComments(comments, level = 0) {
        return comments.map(comment => {
            const hasReplies = comment.replies && comment.replies.length > 0;
            const repliesHtml = hasReplies ? renderComments(comment.replies, isLoggedIn, level + 1) : '';
            
            return `
                <div class="comment" style="margin-left: ${level * 4}px" data-comment-id="${comment.id}">
                    <div class="comment-votes">
                        ${isLoggedIn ? `
                            <button class="comment-vote-button ${comment.userVote === 'up' ? 'active' : ''}" 
                                    onclick="window.handleCommentVote(${comment.id}, 'up')" 
                                    data-vote="up">
                                <i data-feather="thumbs-up"></i>
                                <span class="comment-vote-count-up">${comment.likes}</span>
                            </button>
                            
                            <button class="comment-vote-button ${comment.userVote === 'down' ? 'active' : ''}" 
                                    onclick="window.handleCommentVote(${comment.id}, 'down')"
                                    data-vote="down">
                                <i data-feather="thumbs-down"></i>
                                <span class="comment-vote-count-down">${comment.dislikes}</span>
                            </button>
                        ` : `
                            <div class="vote-button disabled">
                                <i data-feather="thumbs-up"></i>
                                <span class="vote-count-up">${comment.likes}</span>
                            </div>
                            
                            <div class="vote-button disabled">
                                <i data-feather="thumbs-down"></i>
                                <span class="vote-count-down">${comment.dislikes}</span>
                            </div>
                        `}
                    </div>
                    <div class="comment-content">
                        <small>u/${comment.author} • ${formatTimeAgo(comment.timestamp)}</small>
                        <p class="comment-text">${comment.content}</p>
                        <div class="comment-actions">
                            ${isLoggedIn ? `
                                <button class="reply-button" onclick="window.showReplyInput(${comment.id})">Reply</button>
                            ` : `
                                <button class="reply-button" onclick="window.location.href='/login_Page'">Login to Reply</button>
                            `}
                            ${hasReplies ? `
                                <button class="toggle-replies-button" onclick="window.toggleReplies(${comment.id})">
                                    <i class="toggle-icon" data-feather="chevron-down"></i>
                                    <span>${comment.replies.length} replies</span>
                                </button>
                            ` : ''}
                        </div>
                        <div class="reply-input-container" style="display: none;">
                            <textarea class="reply-input" placeholder="Write a reply..."></textarea>
                            <div class="reply-buttons">
                                <button class="button button-primary" onclick="window.submitReply(${comment.id},'.reply-input')">Submit</button>
                                <button class="button" onclick="window.cancelReply(${comment.id})">Cancel</button>
                            </div>
                        </div>
                        <div class="comment-replies" style="display: none;">
                            ${repliesHtml}
                        </div>
                    </div>
                </div>
            `;
            
        }).join('');
    }

    window.handleCommentVote = async function(commentId, voteType) {
        if (!isLoggedIn) {
            showToast('Please login to vote');
            return;
        }

        try {
            const response = await fetch('http://localhost:8080/vote', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    postId: commentId,
                    voteType: voteType
                })
            });

            if (!response.ok) {
                throw new Error('Failed to vote');
            }

            const data = await response.json();
            
            // Update UI
            const comment = document.querySelector(`[data-comment-id="${commentId}"]`);
            const upCount = comment.querySelector('.comment-vote-count-up');
            const downCount = comment.querySelector('.comment-vote-count-down');
            const upButton = comment.querySelector('[data-vote="up"]');
            const downButton = comment.querySelector('[data-vote="down"]');

            upCount.textContent = data.likes;
            downCount.textContent = data.dislikes;

            // Update active states
            upButton.classList.toggle('active', data.userVote === 'up');
            downButton.classList.toggle('active', data.userVote === 'down');

        } catch (error) {
            console.error('Error voting:', error);
            showToast('Failed to vote. Please try again.');
        }
    }

    window.showReplyInput = function (commentId) {
        if (!isLoggedIn) {
            console.log("Reply Input user not displaying: User not logged in");
            return;
        }
        
        const comment = document.querySelector(`[data-comment-id="${commentId}"]`);
        if (!comment) {
            console.error(`Comment with ID ${commentId} not found`);
            return;
        }

        const replyContainer = comment.querySelector('.reply-input-container');
        if (!replyContainer) {
            console.error(`Reply container not found for comment ${commentId}`);
            return;
        }

        replyContainer.style.display = 'block';
    }

    window.cancelReply = function(commentId) {
        const comment = document.querySelector(`[data-comment-id="${commentId}"]`);
        if (!comment) return;

        const replyContainer = comment.querySelector('.reply-input-container');
        if (!replyContainer) return;

        const replyInput = replyContainer.querySelector('.reply-input');
        if (replyInput) {
            replyInput.value = '';
        }
        replyContainer.style.display = 'none';
    }

    window.toggleReplies = function(commentId) {
        const comment = document.querySelector(`[data-comment-id="${commentId}"]`);
        if (!comment) return;

        const repliesContainer = comment.querySelector('.comment-replies');
        const toggleButton = comment.querySelector('.toggle-replies-button');
        const toggleIcon = toggleButton?.querySelector('.toggle-icon');
        
        if (repliesContainer && toggleIcon) {
            if (repliesContainer.style.display === 'none') {
                repliesContainer.style.display = 'block';
                toggleIcon.dataset.feather = 'chevron-up';
            } else {
                repliesContainer.style.display = 'none';
                toggleIcon.dataset.feather = 'chevron-down';
            }
            feather.replace();
        }
    }
    // Show comment actions when textarea is focused
    document.addEventListener('click', function(e) {
        if (e.target && e.target.classList.contains('main-comment-input')) {
            const actionsDiv = e.target.parentElement.querySelector('.comment-actions');
            if (actionsDiv) {
                actionsDiv.style.display = 'flex';
            }
        }
    });

    // Hide comment actions when clicking outside
    document.addEventListener('click', function(e) {
        if (!e.target.classList.contains('main-comment-input') && 
            !e.target.closest('.comment-actions')) {
            const textarea = document.querySelector('.main-comment-input');
            if (textarea && textarea.value.trim() === '') {
                const actionsDiv = textarea.parentElement.querySelector('.comment-actions');
                if (actionsDiv) {
                    actionsDiv.style.display = 'none';
                }
            }
        }
    });

    window.submitReply = async function(commentId,selector) {
        let replyInput;
        
        if (selector === '.main-comment-input') {
            // For main comment input, find it using post ID
            const post = document.querySelector(`[data-post-id="${commentId}"]`);
            if (!post) return;
            replyInput = post.querySelector(selector);
        } else {
            // For replies to comments, use the existing logic
            const comment = document.querySelector(`[data-comment-id="${commentId}"]`);
            if (!comment) return;
            replyInput = comment.querySelector(selector);
        }

        if (!replyInput) return;

        const replyContent = replyInput.value.trim();
        
        console.log("selector:", selector)
        if (replyContent) {
            try {
                const response = await fetch('http://localhost:8080/reply', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        commentId: commentId,
                        content: replyContent
                    })
                });

                if (!response.ok) {
                    throw new Error('Failed to submit reply');
                }

                const result = await response.json();

                // Find the post containing this comment
                const post = posts.find(p => 
                    p.comments.some(c => c.id === commentId) ||
                    p.comments.some(c => c.replies && c.replies.some(r => r.id === commentId))
                );

                if (post) {
                    // Re-render the comments after successful reply
                    showComments(post.id);
                    
                    // Clear and hide the reply input
                    replyInput.value = '';
                    const replyContainer = document.querySelector('.reply-input-container');
                    if (replyContainer) {
                        replyContainer.style.display = 'none';
                    }
                }

            } catch (error) {
                console.error('Error submitting reply:', error);
                // Could add user-facing error handling here
            }
        }
    }

    // Render trending communities
    function renderTrending() {
        const container = document.getElementById('trending-communities');
        const category = [
            { name: 'programming', members: '2.5M' },
            { name: 'technology', members: '1.8M' },
            { name: 'movies', members: '900K' }
        ];

        category.forEach(category => {
            const div = document.createElement('div');
            div.className = 'sidebar-link';
            div.innerHTML = `
                <i data-feather="users"></i>
                r/${category.name}
                <small style="margin-left: auto">${category.members} members</small>
            `;
            container.appendChild(div);
        });
        feather.replace();
    }

    // Initialize the page
    renderPosts();
    renderTrending();

    // Add voting functionality
    document.addEventListener('click', async function(e) {
        if (e.target.closest('.vote-button')) {
            const button = e.target.closest('.vote-button');
            if (!button || button.classList.contains('disabled')) return;

            const voteType = button.getAttribute('data-vote');
            const postElement = button.closest('.post');
            const postId = parseInt(postElement.getAttribute('data-post-id'));

            try {
                const response = await fetch('http://localhost:8080/vote', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        postId: postId,
                        voteType: voteType
                    })
                });

                if (!response.ok) {
                    throw new Error('Failed to vote');
                }

                const data = await response.json();
                
                // Update vote counts
                const upCount = postElement.querySelector('.vote-count-up');
                const downCount = postElement.querySelector('.vote-count-down');
                console.log("Up Count:", upCount)
                console.log("Down Count:", downCount)
                if (upCount) upCount.textContent = data.likes;
                if (downCount) downCount.textContent = data.dislikes;

                // Update button states
                const upvoteBtn = postElement.querySelector('.vote-button[data-vote="up"]');
                const downvoteBtn = postElement.querySelector('.vote-button[data-vote="down"]');
                if (upvoteBtn) upvoteBtn.classList.toggle('active', data.userVote === 'up');
                if (downvoteBtn) downvoteBtn.classList.toggle('active', data.userVote === 'down');

            } catch (error) {
                console.error('Error voting:', error);
                showToast('Failed to vote. Please try again.');
            }
        }
    });

    // Create post functionality
    const createPostButton = document.getElementById('createPostButton');
    const postEditorContainer = document.getElementById('postEditorContainer');
    if (createPostButton) {
        createPostButton.addEventListener('click', function() {
            postEditorContainer.classList.toggle('show');
            feather.replace();
        });
    }

    // Check login status
    async function checkLoginStatus() {
        const response = await fetch('http://localhost:8080/check_login', {
            method: 'GET',
            credentials: 'include' // Ensures cookies are sent with the request
        });
       
        if (response.ok) {
            console.log('User is logged in');
            return true;
        } else {
            console.log('User is not logged in');
            return false;
        }
    }

   
    window.removeSession = async () => {
        const response = await fetch('http://localhost:8080/logout', {
            method: 'POST',
            credentials: 'include'
        });

        if (response.ok) {
            console.log('Successfully logged out');
            window.location.href = '/login_Page';
        } else {
            console.log('Logout failed');
        }
    }

    // Update UI based on login status 
    function updateUserSection() {
        if (!isLoggedIn) {
            userSection.innerHTML = `
                <div class="profile-section">
                    <button class="sign-btn" onclick="window.location.href='/login_Page'">
                        <i data-feather="user"></i>
                        <span>Sign In</span>
                    </button>
                    <button class="theme-toggle1" onclick="toggleTheme()">
                        <i data-feather="moon"></i>
                    </button>
                </div>
            `;
        } else {
            userSection.innerHTML = `
                <div class="profile-section">
                   <button class="theme-toggle" onclick="toggleTheme()">
                        <i data-feather="moon"></i>
                    </button>
                    <button class="button-user button-primary-user" id="createPostButton">
                        <i data-feather="edit-3"></i>
                        <span>Create Post</span>
                    </button>
                    <button class="button-user button-outline" id="logoutButton" onclick="removeSession()">
                        <i data-feather="log-out"></i>
                        <span>Logout</span>
                    </button>
                    <div class="profile-image">
                        <img src="${localStorage.getItem('profileImage') || '/default-avatar.png'}" 
                             alt="Profile" 
                             class="avatar">
                    </div>
                   
                </div>
            `;
            
            // Reattach create post button listener
            const createPostButton = document.getElementById('createPostButton');
            if (createPostButton) {
                createPostButton.addEventListener('click', function() {
                    const postsContainer = document.getElementById('posts-container');
                    // Clear existing content
                    postsContainer.innerHTML = '';
                    
                    const postEditorHTML = `
                        <div class="post-editor">
                            <div class="community-selector">
                                <select class="input-field" id="category-select">
                                    <option value="">Select Category</option>
                                    <option value="programming">r/programming</option>
                                    <option value="webdev">r/webdev</option>
                                    <option value="javascript">r/javascript</option>
                                </select>
                                <div id="selected-categories" class="selected-categories"></div>
                            </div>

                            <div class="post-tabs">
                                <div class="tab active" id="text-tab">
                                    <i data-feather="type"></i>
                                    Text
                                </div>
                                <div class="tab" id="media-tab">
                                    <i data-feather="image"></i>
                                    Images 
                                </div>
                               
                            </div>

                            <div id="text-content">
                                <input type="text" id="post-title" class="input-field" placeholder="Title" maxlength="300">
                                <div id="post-body" class="rich-editor" contenteditable="true">
                                    Share your thoughts...
                                </div>
                            </div>

                            <div id="media-content" style="display: none;">
                                <input type="text" class="input-field" placeholder="Title" maxlength="300">
                                <div class="media-upload-area" style="min-height: 300px; border: 2px dashed var(--border-color); border-radius: 8px; display: flex; align-items: center; justify-content: center; margin: 20px 0; cursor: pointer;">
                                    <div style="text-align: center;">
                                        <i data-feather="upload" style="width: 48px; height: 48px; margin-bottom: 10px;"></i>
                                        <p>Drag and drop images/videos here</p>
                                        <p style="font-size: 12px; color: var(--text-secondary);">or click to upload</p>
                                    </div>
                                </div>
                            </div>

                            <div class="actions">
                                <button class="button button-primary" id="submit-post">Post</button>
                            </div>
                        </div>
                    `;
                    postsContainer.insertAdjacentHTML('afterbegin', postEditorHTML);
                    feather.replace();

                    // Set up tab switching
                    const textTab = document.getElementById('text-tab');
                    const mediaTab = document.getElementById('media-tab');
                    const textContent = document.getElementById('text-content');
                    const mediaContent = document.getElementById('media-content');

                    textTab.addEventListener('click', () => {
                        textTab.classList.add('active');
                        mediaTab.classList.remove('active');
                        textContent.style.display = 'block';
                        mediaContent.style.display = 'none';
                    });

                    mediaTab.addEventListener('click', () => {
                        mediaTab.classList.add('active');
                        textTab.classList.remove('active');
                        mediaContent.style.display = 'block';
                        textContent.style.display = 'none';
                    });

                    // Set up category selection handling
                    const categorySelect = document.getElementById('category-select');
                    const selectedCategories = document.getElementById('selected-categories');
                    const selectedCats = new Set();

                    categorySelect.addEventListener('change', function() {
                        const selectedValue = this.value;
                        const selectedText = this.options[this.selectedIndex].text;
                        
                        if (selectedValue && !selectedCats.has(selectedValue)) {
                            selectedCats.add(selectedValue);
                            
                            const categoryTag = document.createElement('div');
                            categoryTag.className = 'category-tag';
                            categoryTag.innerHTML = `
                                ${selectedText}
                                    <span class="remove-category" data-value="${selectedValue}">×</span>
                                `;
                            
                            selectedCategories.appendChild(categoryTag);
                            
                            // Reset select to placeholder
                            this.value = '';
                        }
                    });

                    selectedCategories.addEventListener('click', function(e) {
                        if (e.target.classList.contains('remove-category')) {
                            const value = e.target.dataset.value;
                            selectedCats.delete(value);
                            e.target.parentElement.remove();
                        }
                    });

                    // Handle post submission
                    const submitButton = document.getElementById('submit-post');
                    submitButton.addEventListener('click', async () => {
                        const title = document.getElementById('post-title').value;
                        const content = document.getElementById('post-body').innerText;
                        const categories = Array.from(selectedCats);

                        if (!title || !content || categories.length === 0) {
                            showToast('Please fill in all fields and select at least one category');
                            return;
                        }

                        try {
                            const response = await fetch('http://localhost:8080/posts', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json'
                                },
                                credentials: 'include',
                                body: JSON.stringify({
                                    title,
                                    content,
                                    categories
                                })
                            });

                            if (!response.ok) {
                                throw new Error('Failed to create post');
                            }

                            showToast('Post created successfully!');
                            window.location.reload(); // Refresh to show new post
                        } catch (error) {
                            console.error('Error creating post:', error);
                            showToast('Failed to create post. Please try again.');
                        }
                    });
                });
            }
        }
        feather.replace();
    }
    

    window.toggleMobileMenu = function() {
        const sidebar = document.querySelector('.sidebar');
        sidebar.classList.toggle('active');
        
        // Update menu icon
        const menuIcon = document.querySelector('.menu-toggle i');
        if (sidebar.classList.contains('active')) {
        menuIcon.setAttribute('data-feather', 'x');
    } else {
        menuIcon.setAttribute('data-feather', 'menu'); 
    }
    feather.replace();
}
});