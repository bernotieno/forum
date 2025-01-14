// Store likes and dislikes for each post
const postVotes = new Map();

function initializePost(postId) {
    if (!postVotes.has(postId)) {
        postVotes.set(postId, {
            likes: 0,
            dislikes: 0
        });
    }
}

function AddLike(postId) {
    initializePost(postId);
    const votes = postVotes.get(postId);
    votes.likes++;
    UpdateLikes(postId);
    console.log(`Post ${postId} likes:`, votes.likes);
}

function MinusLike(postId) {
    initializePost(postId);
    const votes = postVotes.get(postId);
    votes.dislikes++;
    UpdateDisLikes(postId);
    console.log(`Post ${postId} dislikes:`, votes.dislikes);
}

function UpdateLikes(postId) {
    document.getElementById(`likes-container-${postId}`).textContent = postVotes.get(postId).likes;
}

function UpdateDisLikes(postId) {
    document.getElementById(`dislikes-container-${postId}`).textContent = postVotes.get(postId).dislikes;
}