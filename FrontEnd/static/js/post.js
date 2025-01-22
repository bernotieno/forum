document.addEventListener('DOMContentLoaded', function() {
    const textTab = document.getElementById('text-tab');
    const mediaTab = document.getElementById('media-tab');
    const textContent = document.getElementById('text-content');
    const mediaContent = document.getElementById('media-content');
    const dropzone = document.getElementById('dropzone');
    const fileInput = document.getElementById('file-input');
    const postForm = document.getElementById('postForm');
    const categorySelect = document.getElementById('category-select');
    const selectedCategories = document.getElementById('selected-categories');
    const selectedCats = new Set();
    let uploadedFiles = new Set();

    // Determine if the form is for creating or editing
    const isEditForm = postForm.getAttribute('action') === '/updatePost';

     // Pre-fill selected categories if editing
     if (isEditForm) {
         const preSelectedCategories = document.querySelectorAll('.category-tag');
         console.log("preSelectedCategories",preSelectedCategories);
        preSelectedCategories.forEach(categoryTag => {
            const value = categoryTag.querySelector('.remove-category').dataset.value;
            selectedCats.add(value);
        });
    }
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

    // Handle file upload via drag and drop
    dropzone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropzone.classList.add('dragover');
    });

    dropzone.addEventListener('dragleave', () => {
        dropzone.classList.remove('dragover');
    });

    dropzone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropzone.classList.remove('dragover');
        handleFiles(e.dataTransfer.files);
    });

    // Handle file upload via click
    dropzone.addEventListener('click', () => {
        fileInput.click();
    });

    function validateAndOptimizeImage(file) {
        return new Promise((resolve, reject) => {
            const maxSize = 5 * 1024 * 1024; // 5MB
            if (file.size > maxSize) {
                reject('File size should be less than 5MB');
                return;
            }

            const img = new Image();
            img.onload = function() {
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                
                // Maintain aspect ratio while resizing if needed
                let width = img.width;
                let height = img.height;
                const maxDim = 2000;
                
                if (width > maxDim || height > maxDim) {
                    if (width > height) {
                        height = (height / width) * maxDim;
                        width = maxDim;
                    } else {
                        width = (width / height) * maxDim;
                        height = maxDim;
                    }
                }

                canvas.width = width;
                canvas.height = height;
                ctx.drawImage(img, 0, 0, width, height);

                canvas.toBlob((blob) => {
                    resolve(new File([blob], file.name, {
                        type: 'image/jpeg',
                        lastModified: Date.now()
                    }));
                }, 'image/jpeg', 0.9); // 90% quality
            };

            img.onerror = () => reject('Invalid image file');
            img.src = URL.createObjectURL(file);
        });
    }

    fileInput.addEventListener('change', async (e) => {
        try {
            const optimizedFile = await validateAndOptimizeImage(e.target.files[0]);
            handleFiles([optimizedFile]);
        } catch (error) {
            showToast(error);
        }
    });

    function handleFiles(files) {
        Array.from(files).forEach(file => {
            if (!file.type.match('image.*') && !file.type.match('video.*')) {
                showToast('Only image  files are allowed');
                return;
            }
            
            if (file.size > 5 * 1024 * 1024) { // 5MB limit
                showToast('File size should be less than 5MB');
                return;
            }

            uploadedFiles.add(file);
            displayPreview(file);
        });
    }

    function displayPreview(file) {
        const preview = document.createElement('div');
        preview.className = 'media-preview';
        
        if (file.type.startsWith('image/')) {
            const img = document.createElement('img');
            img.src = URL.createObjectURL(file);
            preview.appendChild(img);
        } else {
            const video = document.createElement('video');
            video.src = URL.createObjectURL(file);
            video.controls = true;
            preview.appendChild(video);
        }

        const removeBtn = document.createElement('button');
        removeBtn.innerHTML = '×';
        removeBtn.className = 'remove-media';
        removeBtn.onclick = () => {
            uploadedFiles.delete(file);
            preview.remove();
        };

        preview.appendChild(removeBtn);
        mediaContent.appendChild(preview);
    }

    // Update form submission to include media files
    postForm.addEventListener('submit', async (e) => {
        e.preventDefault();
    
        const title = document.getElementById('post-title').value;
        const content = document.getElementById('post-body').innerText;
        const categories = Array.from(selectedCats);
        const fileInput = document.getElementById('file-input');
    
        // Check if required fields are filled
        if (!title || categories.length === 0) {
            console.log('Title and categories are required', categories);
            showToast('Title and categories are required');
            return;
        }
    
        // Check if at least one of content or file is provided
        if (!content && (!fileInput || fileInput.files.length === 0)) {
            showToast('Please provide either text content or an image');
            return;
        }
    
        // Create a FormData object
        const formData = new FormData();
        formData.append('title', title);
        formData.append('content', content);
        formData.append('category', categories.join(","));
    
        // Append the file if selected
        if (fileInput.files.length > 0) {
            formData.append('post-file', fileInput.files[0]);
        }
    
        try {
            const csrfToken = document.querySelector('input[name="csrf_token"]').value;
            formData.append('csrf_token', csrfToken);
    
            // Determine the endpoint and include postId in the URL if editing
            let endpoint = '/createPost';
            let method = 'POST';
            if (isEditForm) {
                const postId = document.querySelector('input[name="id"]').value;
                endpoint = `/updatePost?id=${postId}`;
                method = 'PUT';
            }
    
            const response = await fetch(endpoint, {
                method: method,
                body: formData
            });
    
            const data = await response.json();
    
            if (response.ok) {
                window.location.href = '/';
            } else {
                showToast(data.error || 'Failed to submit post');
            }
        } catch (error) {
            console.error('Error:', error);
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