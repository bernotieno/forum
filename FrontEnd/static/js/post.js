document.addEventListener('DOMContentLoaded', function() {
    const postForm = document.getElementById('postForm');
    const textTab = document.getElementById('text-tab');
    const mediaTab = document.getElementById('media-tab');
    const textContent = document.getElementById('text-content');
    const mediaContent = document.getElementById('media-content');
    const categorySelect = document.getElementById('category-select');
    const selectedCategories = document.getElementById('selected-categories');
    const selectedCats = new Set();

    // Tab switching
    textTab.addEventListener('click', () => {
        textTab.classList.add('active');
        mediaTab.classList.remove('active');
        textContent.classList.add('active');
        mediaContent.classList.remove('active');
    });

    mediaTab.addEventListener('click', () => {
        mediaTab.classList.add('active');
        textTab.classList.remove('active');
        mediaContent.classList.add('active');
        textContent.classList.remove('active');
    });

    // Category handling
    categorySelect.addEventListener('change', function() {
        const selectedValue = this.value;
        if (!selectedValue) return;
        
        if (!selectedCats.has(selectedValue)) {
            selectedCats.add(selectedValue);
            const categoryTag = document.createElement('div');
            categoryTag.className = 'category-tag';
            categoryTag.innerHTML = `
                ${this.options[this.selectedIndex].text}
                <span class="remove-category" data-value="${selectedValue}">×</span>
            `;
            selectedCategories.appendChild(categoryTag);
        }
        this.value = '';
    });

    selectedCategories.addEventListener('click', (e) => {
        if (e.target.classList.contains('remove-category')) {
            const value = e.target.dataset.value;
            selectedCats.delete(value);
            e.target.parentElement.remove();
        }
    });

    // Form submission
    postForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const title = document.getElementById('post-title').value;
        const content = document.getElementById('post-body').innerText;
        const categories = Array.from(selectedCats);

        if (!title || !content || categories.length === 0) {
            showToast('Please fill in all required fields');
            return;
        }

        const formData = {
            title: title,
            content: content,
            categories: categories
        };

        try {
            const response = await fetch('/createPost', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value
                },
                body: JSON.stringify(formData)
            });

            const data = await response.json();
            
            if (response.ok) {
                window.location.href = '/';
            } else {
                showToast(data.error || 'Failed to create post');
            }
        } catch (error) {
            showToast('An error occurred. Please try again.');
        }
    });

    // Toast notification
    function showToast(message) {
        const toast = document.getElementById('toast');
        const toastMessage = document.getElementById('toastMessage');
        toastMessage.textContent = message;
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), 3000);
    }
}); 