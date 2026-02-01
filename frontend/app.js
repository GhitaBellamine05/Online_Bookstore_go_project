let currentUser = null;
let currentToken = localStorage.getItem('auth_token');
let currentSection = 'welcome';

// ==================== INITIALIZATION ====================
document.addEventListener('DOMContentLoaded', () => {
    if (currentToken) {
        verifyToken();
    } else {
        showSection('welcome');
    }
});

// ==================== AUTHENTICATION ====================
function verifyToken() {
    fetch('/users/me', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => {
        if (res.status === 429) {
            throw new Error('Rate limit exceeded. Please wait before trying again.');
        }
        if (res.ok) return res.json();
        throw new Error('Invalid token');
    })
    .then(user => {
        currentUser = user;
        updateUI();
        showSection('dashboard');
    })
    .catch(() => {
        localStorage.removeItem('auth_token');
        currentToken = null;
        currentUser = null;
        updateUI();
        showSection('welcome');
    });
}

function login() {
    const username = document.getElementById('auth-username').value.trim();
    const password = document.getElementById('auth-password').value;
    if (!username || !password) {
        showError('auth-error', 'Please fill in all fields');
        return;
    }
    fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    })
    .then(res => {
        if (res.status === 429) {
            throw new Error('Too many login attempts. Please wait a moment and try again.');
        }
        if (!res.ok) throw new Error('Invalid credentials');
        return res.json();
    })
    .then(data => {
        currentToken = data.token;
        localStorage.setItem('auth_token', currentToken);
        verifyToken();
        closeAuthModal();
        showNotification('success', 'Welcome back!');
    })
    .catch(err => {
        showError('auth-error', err.message || 'Login failed');
    });
}

function register() {
    const name = document.getElementById('reg-name').value.trim();
    const username = document.getElementById('reg-username').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const password = document.getElementById('reg-password').value;

    console.log('Sending to /auth/register:', { name, username, email, password });
    if (!name || !username || !email || !password) {
        const missing = [];
        if (!name) missing.push('Full Name');
        if (!username) missing.push('Username');
        if (!email) missing.push('Email');
        if (!password) missing.push('Password');
        showError('register-error', `Missing: ${missing.join(', ')}`);
        return;
    }
    if (password.length < 8) {
        showError('register-error', 'Password must be at least 8 characters');
        return;
    }
    fetch('/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            name: name,       
            username: username,
            email: email,
            password: password
        })
    })
    .then(response => {
    if (response.status === 429) {
        throw new Error('Too many registration attempts. Please wait and try again.');
    }
    console.log(' Response status:', response.status);
    return response.json().then(data => ({ status: response.status, data }));
    })
    .then(result => {
        if (result.status !== 201) {
            throw new Error(result.data.error || 'Registration failed');
        }
        showNotification('success', 'Account created! Please login.');
        switchToLogin();
    })
    .catch(error => {
        console.error(' Registration error:', error);
        showError('register-error', error.message);
    });
}
function logout() {
    localStorage.removeItem('auth_token');
    currentToken = null;
    currentUser = null;
    updateUI();
    showSection('welcome');
    showNotification('info', 'You have been logged out');
}

// ==================== UI MANAGEMENT ====================
function updateUI() {
    const guestSection = document.getElementById('guest-section');
    const userSection = document.getElementById('user-section');
    const usernameDisplay = document.getElementById('username-display');
    const userRoleBadge = document.getElementById('user-role-badge');
    
    if (currentUser) {
        guestSection.style.display = 'none';
        userSection.style.display = 'flex';
        usernameDisplay.textContent = currentUser.username;
        userRoleBadge.textContent = currentUser.role.toUpperCase();
        userRoleBadge.className = `user-role role-${currentUser.role}`;
    } else {
        guestSection.style.display = 'flex';
        userSection.style.display = 'none';
    }
}

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(sec => {
        sec.classList.remove('active');
    });
    
    document.getElementById(sectionId).classList.add('active');
    currentSection = sectionId;
    
    switch(sectionId) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'books':
            loadBooks();
            break;
        case 'authors':
            loadAuthors();
            break;
        case 'customers':
            loadCustomers();
            break;
        case 'orders':
            loadOrders();
            break;
        case 'all-orders':
            loadAllOrders();
            break;
        case 'reports':
            loadReports();
            break;
    }
}

function openAuthModal() {
    document.getElementById('auth-modal').style.display = 'block';
}

function closeAuthModal() {
    document.getElementById('auth-modal').style.display = 'none';
    document.getElementById('auth-form').style.display = 'block';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('auth-error').textContent = '';
    document.getElementById('register-error').textContent = '';
}

function switchToRegister() {
    document.getElementById('auth-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
    document.getElementById('auth-modal-title').innerHTML = '<i class="fas fa-user-plus"></i> Create Account';
}

function switchToLogin() {
    document.getElementById('auth-form').style.display = 'block';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('auth-modal-title').innerHTML = '<i class="fas fa-user-circle"></i> Welcome to BookWise';
}

// ==================== MODAL MANAGEMENT ====================
function openModal(formType, data = null) {
    const modal = document.getElementById('modal');
    const modalTitle = document.getElementById('modal-title');
    const modalForm = document.getElementById('modal-form');
    
    let title = '';
    let formHTML = '';
    
    switch(formType) {
        case 'book-form':
            title = data ? 'Edit Book' : 'Add New Book';
            formHTML = `
                <input type="hidden" id="form-id">
                <div class="form-group">
                    <label><i class="fas fa-heading"></i> Title</label>
                    <input type="text" id="form-title" placeholder="Book title" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-user-edit"></i> Author</label>
                    <select id="form-author-id" required></select>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-dollar-sign"></i> Price ($)</label>
                    <input type="number" id="form-price" step="0.01" placeholder="Price" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-cubes"></i> Stock</label>
                    <input type="number" id="form-stock" placeholder="Available stock" required>
                </div>
            `;
            break;
            
        case 'author-form':
            title = data ? 'Edit Author' : 'Add New Author';
            formHTML = `
                <input type="hidden" id="form-id">
                <div class="form-group">
                    <label><i class="fas fa-user"></i> First Name</label>
                    <input type="text" id="form-first-name" placeholder="First name" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-user"></i> Last Name</label>
                    <input type="text" id="form-last-name" placeholder="Last name" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-book"></i> Biography</label>
                    <textarea id="form-bio" placeholder="Author biography" rows="3"></textarea>
                </div>
            `;
            break;
            
        case 'order-form':
            title = 'Place New Order';
            formHTML = `
                <div class="form-group">
                    <label><i class="fas fa-book"></i> Select Book</label>
                    <select id="form-book-id" required></select>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-hashtag"></i> Quantity</label>
                    <input type="number" id="form-quantity" min="1" value="1" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-dollar-sign"></i> Total Price</label>
                    <input type="text" id="form-total-price" readonly>
                </div>
            `;
            break;
            
        case 'customer-form':
            title = 'Add Customer';
            formHTML = `
                <div class="form-group">
                    <label><i class="fas fa-user-tag"></i> Full Name</label>
                    <input type="text" id="form-name" placeholder="Full name" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-user"></i> Username</label>
                    <input type="text" id="form-username" placeholder="Username" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-envelope"></i> Email</label>
                    <input type="email" id="form-email" placeholder="Email" required>
                </div>
                <div class="form-group">
                    <label><i class="fas fa-lock"></i> Password</label>
                    <input type="password" id="form-password" placeholder="Password" required>
                </div>
            `;
            break;
    }
    
    modalTitle.innerHTML = `<i class="fas fa-edit"></i> ${title}`;
    modalForm.innerHTML = formHTML;
    
    if (data) {
        for (let key in data) {
            const el = document.getElementById(`form-${key}`);
            if (el) el.value = data[key];
        }
    }
    
    if (formType === 'book-form' || formType === 'order-form') {
        loadAuthorsForDropdown();
    }
    if (formType === 'order-form') {
        loadBooksForDropdown();
        setupOrderPriceCalculation();
    }
    
    modal.style.display = 'block';
}

function closeModal() {
    document.getElementById('modal').style.display = 'none';
    document.getElementById('form-error').textContent = '';
}

function submitModalForm() {
    const formType = document.querySelector('#modal-title').textContent.includes('Book') ? 'book' :
                     document.querySelector('#modal-title').textContent.includes('Author') ? 'author' :
                     document.querySelector('#modal-title').textContent.includes('Order') ? 'order' :
                     document.querySelector('#modal-title').textContent.includes('Customer') ? 'customer' : '';
    
    switch(formType) {
        case 'book':
            submitBookForm();
            break;
        case 'author':
            submitAuthorForm();
            break;
        case 'order':
            submitOrderForm();
            break;
        case 'customer':
            submitCustomerForm();
            break;
    }
}

// ==================== DASHBOARD ====================
function loadDashboard() {
    if (!currentUser) return;
    
    if (currentUser.role === 'admin') {
        loadAdminDashboard();
    } else {
        loadCustomerDashboard();
    }
}

function loadAdminDashboard() {
    document.getElementById('admin-dashboard').style.display = 'block';
    document.getElementById('customer-dashboard').style.display = 'none';
    
    Promise.all([
        fetch('/books', { headers: { 'Authorization': `Bearer ${currentToken}` } }),
        fetch('/authors', { headers: { 'Authorization': `Bearer ${currentToken}` } }),
        fetch('/users?role=customer', { headers: { 'Authorization': `Bearer ${currentToken}` } }),
        fetch('/orders', { headers: { 'Authorization': `Bearer ${currentToken}` } })
    ])
    .then(responses => Promise.all(responses.map(r => r.json())))
    .then(([books, authors, customers, orders]) => {
        const booksCount = normalizeData(books).length;
        const authorsCount = normalizeData(authors).length;
        const customersCount = normalizeData(customers).length;
        const ordersCount = normalizeData(orders).length;
        
        document.getElementById('stat-books').textContent = booksCount;
        document.getElementById('stat-authors').textContent = authorsCount;
        document.getElementById('stat-customers').textContent = customersCount;
        document.getElementById('stat-orders').textContent = ordersCount;
        
        const ordersArray = normalizeData(orders);
        let revenue = 0;
        ordersArray.forEach(o => revenue += o.total_price || 0);
        document.getElementById('stat-revenue').textContent = revenue.toFixed(2);
    })
    .catch(err => {
        console.error('Error loading dashboard:', err);
    });
}

function loadCustomerDashboard() {
    document.getElementById('admin-dashboard').style.display = 'none';
    document.getElementById('customer-dashboard').style.display = 'block';
    document.getElementById('customer-name').textContent = currentUser.name || currentUser.username;
    
    Promise.all([
        fetch('/books', { headers: { 'Authorization': `Bearer ${currentToken}` } }),
        fetch('/authors', { headers: { 'Authorization': `Bearer ${currentToken}` } }),
        fetch('/orders', { headers: { 'Authorization': `Bearer ${currentToken}` } })
    ])
    .then(responses => Promise.all(responses.map(r => r.json())))
    .then(([books, authors, orders]) => {
        const booksCount = normalizeData(books).length;
        const authorsCount = normalizeData(authors).length;
        
        document.getElementById('customer-books-count').textContent = booksCount;
        document.getElementById('customer-authors-count').textContent = authorsCount;
        
        const ordersArray = normalizeData(orders);
        const userOrders = ordersArray.filter(o => o.user_id === currentUser.id);
        
        document.getElementById('customer-orders-count').textContent = userOrders.length;
        
        const totalSpent = userOrders.reduce((sum, o) => sum + (o.total_price || 0), 0);
        document.getElementById('customer-total-spent').textContent = totalSpent.toFixed(2);
    })
    .catch(err => {
        console.error('Error loading dashboard:', err);
    });
}

// ==================== BOOKS MANAGEMENT ====================
function loadBooks() {
    const isAdmin = currentUser && currentUser.role === 'admin';
    
    document.getElementById('books-title').textContent = isAdmin ? 'Manage Books' : 'Browse Books';
    document.getElementById('add-book-btn').style.display = isAdmin ? 'block' : 'none';
    
    fetch('/books', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(books => {
        displayBooks(books, isAdmin);
    })
    .catch(err => {
        console.error('Error loading books:', err);
        document.getElementById('books-list').innerHTML = '<p class="error">Failed to load books</p>';
    });
}

function displayBooks(books, isAdmin) {
    const container = document.getElementById('books-list');
    
    const booksArray = normalizeData(books);
    
    if (!booksArray || booksArray.length === 0) {
        container.innerHTML = '<p class="empty-state">No books available</p>';
        return;
    }
    
    const html = booksArray.map(book => {
        const authorName = book.author 
            ? `${book.author.first_name || ''} ${book.author.last_name || ''}`.trim() 
            : (book.author_id ? `Author ID: ${book.author_id}` : 'Unknown Author');
        
        let actions = '';
        if (isAdmin) {
            actions = `
            <div class="card-actions">
                <button class="btn-primary btn-sm edit-book" data-id="${book.id}">
                    <i class="fas fa-edit"></i> Edit
                </button>
                <button class="btn-danger btn-sm delete-book" data-id="${book.id}">
                    <i class="fas fa-trash"></i> Delete
                </button>
            </div>
            `;
        }
        
        return `
            <div class="card book-card">
                <div class="book-info">
                    <h3>${escapeHtml(book.title || 'Untitled')}</h3>
                    <p class="book-author"><i class="fas fa-user-edit"></i> ${escapeHtml(authorName)}</p>
                    <p class="book-price"><i class="fas fa-dollar-sign"></i> $${(book.price || 0).toFixed(2)}</p>
                    <p class="book-stock"><i class="fas fa-cubes"></i> Stock: ${book.stock || 0}</p>
                </div>
                ${actions}
            </div>
        `;
    }).join('');
    
    container.innerHTML = html;
    
    if (isAdmin) {
        document.querySelectorAll('.edit-book').forEach(btn => {
            btn.addEventListener('click', function() {
                const bookId = parseInt(this.dataset.id);
                fetch(`/books/${bookId}`, {
                    headers: { 'Authorization': `Bearer ${currentToken}` }
                })
                .then(res => res.json())
                .then(book => {
                    openModal('book-form', {
                        id: book.id,
                        title: book.title,
                        author_id: book.author_id || (book.author ? book.author.id : ''),
                        price: book.price,
                        stock: book.stock
                    });
                });
            });
        });
        
        document.querySelectorAll('.delete-book').forEach(btn => {
            btn.addEventListener('click', function() {
                deleteBook(parseInt(this.dataset.id));
            });
        });
    }
}

function submitBookForm() {
    const id = document.getElementById('form-id').value;
    const title = document.getElementById('form-title').value;
    const authorId = document.getElementById('form-author-id').value;
    const price = parseFloat(document.getElementById('form-price').value);
    const stock = parseInt(document.getElementById('form-stock').value);
    
    if (!title || !authorId || isNaN(price) || isNaN(stock)) {
        showError('form-error', 'Please fill in all fields correctly');
        return;
    }
    
    const url = id ? `/books/${id}` : '/books';
    const method = id ? 'PUT' : 'POST';
    
    fetch(url, {
        method: method,
        headers: {
            'Authorization': `Bearer ${currentToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            title: title,
            author: { id: parseInt(authorId) },
            genres: ["General"], 
            price: price,
            stock: stock
        })
    })
    .then(async (res) => {
        if (!res.ok) {
            const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
            throw new Error(errorData.error || 'Operation failed');
        }
        return res.json();
    })
    .then(() => {
        closeModal();
        loadBooks();
        showNotification('success', `Book ${id ? 'updated' : 'created'} successfully`);
    })
    .catch(err => {
        console.error('Book error:', err);
        showError('form-error', err.message);
    });
}

function deleteBook(id) {
    if (!confirm('Are you sure you want to delete this book?')) return;
    
    fetch(`/books/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => {
        if (!res.ok) throw new Error('Delete failed');
        loadBooks();
        showNotification('success', 'Book deleted successfully');
    })
    .catch(err => {
        showNotification('error', err.message || 'Failed to delete book');
    });
}

// ==================== AUTHORS MANAGEMENT ====================
function loadAuthors() {
    const isAdmin = currentUser && currentUser.role === 'admin';
    
    document.getElementById('authors-title').textContent = isAdmin ? 'Manage Authors' : 'View Authors';
    document.getElementById('add-author-btn').style.display = isAdmin ? 'block' : 'none';
    
    fetch('/authors', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(authors => {
        displayAuthors(authors, isAdmin);
    })
    .catch(err => {
        console.error('Error loading authors:', err);
        document.getElementById('authors-list').innerHTML = '<p class="error">Failed to load authors</p>';
    });
}

function displayAuthors(authors, isAdmin) {
    const container = document.getElementById('authors-list');
    
    const authorsArray = normalizeData(authors);
    
    if (!authorsArray || authorsArray.length === 0) {
        container.innerHTML = '<p class="empty-state">No authors available</p>';
        return;
    }
    
    const html = authorsArray.map(author => {
        const fullName = `${author.first_name || ''} ${author.last_name || ''}`.trim() || 'Unnamed Author';
        
        let actions = '';
        if (isAdmin) {
            actions = `
            <div class="card-actions">
                <button class="btn-primary btn-sm edit-author" data-id="${author.id}">
                    <i class="fas fa-edit"></i> Edit
                </button>
                <button class="btn-danger btn-sm delete-author" data-id="${author.id}">
                    <i class="fas fa-trash"></i> Delete
                </button>
            </div>
            `;
        }
        
        return `
            <div class="card author-card">
                <div class="author-info">
                    <h3>${escapeHtml(fullName)}</h3>
                    ${author.bio ? `<p class="author-bio">${escapeHtml(author.bio)}</p>` : ''}
                </div>
                ${actions}
            </div>
        `;
    }).join('');
    
    container.innerHTML = html;
    
    if (isAdmin) {
        document.querySelectorAll('.edit-author').forEach(btn => {
    btn.addEventListener('click', function() {
        const authorId = parseInt(this.dataset.id);
        console.log(' Editing author ID:', authorId);
        
        fetch(`/authors/${authorId}`, {
            headers: { 'Authorization': `Bearer ${currentToken}` }
        })
        .then(res => {
            console.log(' Response status:', res.status);
            return res.json();
        })
        .then(author => {
            console.log(' Author data received:', author);
            const firstName = author.first_name || author.firstName || '';
            const lastName = author.last_name || author.lastName || '';
            const bio = author.bio || author.Bio || '';
            
            openModal('author-form', {
                id: author.id || author.ID || '',
                first_name: firstName,
                last_name: lastName,
                bio: bio
            });
        })
        .catch(err => {
            console.error(' Error fetching author:', err);
            showError('form-error', 'Failed to load author data');
        });
    });
});
        document.querySelectorAll('.delete-author').forEach(btn => {
            btn.addEventListener('click', function() {
                deleteAuthor(parseInt(this.dataset.id));
            });
        });
    }
}

function editAuthor(authorId) {
    if (!authorId || authorId <= 0) {
        showError('form-error', 'Invalid author ID');
        return;
    }
    fetch(`/authors/${authorId}`, {
        headers: {
            'Authorization': `Bearer ${currentToken}`
        }
    })
    .then(async (res) => {
        if (!res.ok) {
            const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
            throw new Error(errorData.error || 'Failed to load author');
        }
        return res.json();
    })
    .then((author) => {
        openModal('author-form');
        document.getElementById('form-id').value = author.id;
        document.getElementById('form-first-name').value = author.first_name;
        document.getElementById('form-last-name').value = author.last_name;
        document.getElementById('form-bio').value = author.bio || '';
    })
    .catch(err => {
        console.error('Load author error:', err);
        showError('form-error', err.message);
    });
}

function submitAuthorForm() {
    const id = document.getElementById('form-id').value;
    const firstName = document.getElementById('form-first-name').value;
    const lastName = document.getElementById('form-last-name').value;
    const bio = document.getElementById('form-bio').value;
    
    if (!firstName || !lastName) {
        showError('form-error', 'First and last name are required');
        return;
    }
    const isCreating = !id || id === 'undefined' || id === 'null' || id.trim() === '';
    
    if (isCreating) {
        fetch('/authors', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${currentToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                bio: bio || null
            })
        })
        .then(async (res) => {
            if (!res.ok) {
                const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
                throw new Error(errorData.error || 'Create failed');
            }
            return res.json();
        })
        .then(() => {
            closeModal();
            loadAuthors();
            showNotification('success', 'Author created successfully');
        })
        .catch(err => {
            console.error(' Author create error:', err);
            showError('form-error', err.message);
        });
    } else {
        fetch(`/authors/${id}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${currentToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                bio: bio || null
            })
        })
        .then(async (res) => {
            if (!res.ok) {
                const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
                throw new Error(errorData.error || 'Update failed');
            }
            return res.json();
        })
        .then(() => {
            closeModal();
            loadAuthors();
            showNotification('success', 'Author updated successfully');
        })
        .catch(err => {
            console.error(' Author update error:', err);
            showError('form-error', err.message);
        });
    }
}
function deleteAuthor(authorId) {
    if (!authorId || authorId <= 0) {
        showError('form-error', 'Invalid author ID');
        return;
    }
    
    if (!confirm('Are you sure you want to delete this author?')) {
        return;
    }
    
    fetch(`/authors/${authorId}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${currentToken}`,
            'Content-Type': 'application/json'
        }
    })
    .then(async (res) => {
        if (!res.ok) {
            const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
            throw new Error(errorData.error || 'Delete failed');
        }
        return true;
    })
    .then(() => {
        loadAuthors();
        showNotification('success', 'Author deleted successfully');
    })
    .catch(err => {
        console.error('Author delete error:', err);
        showError('form-error', err.message);
    });
}
// ==================== ORDERS MANAGEMENT ====================
function loadOrders() {
    fetch('/orders', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(orders => {
        const ordersArray = normalizeData(orders);
        const userOrders = ordersArray.filter(o => o.user_id === currentUser.id);
        displayOrders(userOrders, 'orders-list');
    })
    .catch(err => {
        console.error('Error loading orders:', err);
        document.getElementById('orders-list').innerHTML = '<p class="error">Failed to load orders</p>';
    });
}

function loadAllOrders() {
    fetch('/orders', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(orders => {
        const ordersArray = normalizeData(orders);
        displayOrders(ordersArray, 'all-orders-list');
    })
    .catch(err => {
        console.error('Error loading orders:', err);
        document.getElementById('all-orders-list').innerHTML = '<p class="error">Failed to load orders</p>';
    });
}

function displayOrders(orders, containerId) {
    const container = document.getElementById(containerId);
    
    if (!orders || orders.length === 0) {
        container.innerHTML = '<p class="empty-state">No orders found</p>';
        return;
    }
    
    const html = orders.map(order => {
        let bookTitle = 'Unknown Book';
        let quantity = 1;
        
        if (order.items && order.items.length > 0) {
            const item = order.items[0];
            if (item.book) {
                bookTitle = item.book.title || 'Unknown Book';
            } else if (item.book_id) {
                bookTitle = `Book ID: ${item.book_id}`;
            }
            quantity = item.quantity || 1;
        } else if (order.book_id) {
            bookTitle = `Book ID: ${order.book_id}`;
            quantity = order.quantity || 1;
        }
        
        return `
            <div class="card order-card">
                <div class="order-header">
                    <span class="order-id">Order #${order.id}</span>
                    <span class="order-status status-${order.status || 'pending'}">
                        ${order.status || 'Pending'}
                    </span>
                </div>
                <div class="order-details">
                    <p><i class="fas fa-book"></i> ${escapeHtml(bookTitle)}</p>
                    <p><i class="fas fa-hashtag"></i> Quantity: ${quantity}</p>
                    <p><i class="fas fa-dollar-sign"></i> Total: $${(order.total_price || 0).toFixed(2)}</p>
                    <p><i class="fas fa-clock"></i> ${formatDate(order.created_at)}</p>
                </div>
            </div>
        `;
    }).join('');
    
    container.innerHTML = html;
}

function submitOrderForm() {
    const bookSelect = document.getElementById('form-book-id');
    const quantityInput = document.getElementById('form-quantity');
    
    if (!bookSelect || !quantityInput) {
        showError('form-error', 'Form elements missing');
        return;
    }
    const bookId = bookSelect.value;
    const quantity = parseInt(quantityInput.value);
    const selectedOption = bookSelect.options[bookSelect.selectedIndex];
    const bookPrice = parseFloat(selectedOption?.dataset.price || 0);
    const totalPrice = bookPrice * quantity;
    
    if (!bookId || isNaN(quantity) || quantity < 1) {
        showError('form-error', 'Please select a book and enter valid quantity');
        return;
    }
    
    if (bookPrice <= 0) {
        showError('form-error', 'Invalid book price. Please refresh books list.');
        return;
    }
    const orderData = {
    items: [{
        book_id: parseInt(bookId),  // Just the ID
        quantity: quantity
    }],
    total_price: totalPrice
};
    
    console.log(' Sending order:', orderData);
    
    fetch('/orders', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${currentToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(orderData)
    })
    .then(async (res) => {
        if (!res.ok) {
            const errorData = await res.json().catch(() => ({error: 'Unknown error'}));
            throw new Error(errorData.error || 'Order failed');
        }
        return res.json();
    })
    .then(() => {
        closeModal();
        if (currentSection === 'orders') loadOrders();
        if (currentSection === 'dashboard') loadDashboard();
        showNotification('success', 'Order placed successfully!');
    })
    .catch(err => {
        console.error('❌ Order error:', err);
        showError('form-error', err.message);
    });
}
// ==================== CUSTOMERS MANAGEMENT (Admin) ====================
function loadCustomers() {
    fetch('/users?role=customer', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(customers => {
        const customersArray = normalizeData(customers);
        displayCustomers(customersArray);
    })
    .catch(err => {
        console.error('Error loading customers:', err);
        document.getElementById('customers-list').innerHTML = '<p class="error">Failed to load customers</p>';
    });
}

function displayCustomers(customers) {
    const container = document.getElementById('customers-list');
    
    if (!customers || customers.length === 0) {
        container.innerHTML = '<p class="empty-state">No customers found</p>';
        return;
    }
    
    const html = customers.map(customer => `
        <div class="card customer-card">
            <div class="customer-info">
                <h3>${escapeHtml(customer.name || customer.username)}</h3>
                <p><i class="fas fa-user"></i> ${escapeHtml(customer.username)}</p>
                <p><i class="fas fa-envelope"></i> ${escapeHtml(customer.email || '')}</p>
            </div>
        </div>
    `).join('');
    
    container.innerHTML = html;
}

function submitCustomerForm() {
    const name = document.getElementById('form-name').value;
    const username = document.getElementById('form-username').value;
    const email = document.getElementById('form-email').value;
    const password = document.getElementById('form-password').value;
    
    if (!name || !username || !email || !password) {
        showError('form-error', 'Please fill in all fields');
        return;
    }
    
    fetch('/auth/register', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${currentToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            name,
            username,
            email,
            password,
            role: 'customer'
        })
    })
    .then(res => {
        if (!res.ok) throw new Error('Failed to create customer');
        return res.json();
    })
    .then(() => {
        closeModal();
        loadCustomers();
        showNotification('success', 'Customer created successfully');
    })
    .catch(err => {
        showError('form-error', err.message || 'Failed to create customer');
    });
}

// ==================== REPORTS (Admin) ====================
function loadReports() {
    fetch('/reports', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(reports => {
        displayReports(reports);
    })
    .catch(err => {
        console.error('Error loading reports:', err);
        document.getElementById('reports-list').innerHTML = '<p class="error">Failed to load reports</p>';
    });
}

function displayReports(reports) {
    const container = document.getElementById('reports-list');
    
    if (!reports) {
        container.innerHTML = '<p class="empty-state">No reports available</p>';
        return;
    }
    
    let html = `
        <div class="report-card">
            <h3><i class="fas fa-chart-bar"></i> Sales Summary</h3>
            <div class="report-stats">
                <div class="stat-item">
                    <span class="stat-label">Total Orders</span>
                    <span class="stat-value">${reports.total_orders || 0}</span>
                </div>
                <div class="stat-item">
                    <span class="stat-label">Total Revenue</span>
                    <span class="stat-value">$${(reports.total_revenue || 0).toFixed(2)}</span>
                </div>
                <div class="stat-item">
                    <span class="stat-label">Average Order Value</span>
                    <span class="stat-value">$${reports.total_orders ? (reports.total_revenue / reports.total_orders).toFixed(2) : '0.00'}</span>
                </div>
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

function loadAuthorsForDropdown() {
    fetch('/authors', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(authors => {
        const select = document.getElementById('form-author-id');
        if (!select) return;
        
        const authorsArray = normalizeData(authors);
        select.innerHTML = authorsArray.map(a => `
            <option value="${a.id}">${escapeHtml(a.first_name || '')} ${escapeHtml(a.last_name || '')}</option>
        `).join('');
    })
    .catch(err => console.error('Error loading authors:', err));
}

function loadBooksForDropdown() {
    fetch('/books', {
        headers: { 'Authorization': `Bearer ${currentToken}` }
    })
    .then(res => res.json())
    .then(books => {
        const select = document.getElementById('form-book-id');
        if (!select) return;
        
        const booksArray = normalizeData(books);
        select.innerHTML = booksArray.map(b => `
            <option value="${b.id}" data-price="${b.price || 0}">
                ${escapeHtml(b.title || 'Untitled')} - $${(b.price || 0).toFixed(2)}
            </option>
        `).join('');
    })
    .catch(err => console.error('Error loading books:', err));
}

function setupOrderPriceCalculation() {
    const bookSelect = document.getElementById('form-book-id');
    const quantityInput = document.getElementById('form-quantity');
    const totalPriceInput = document.getElementById('form-total-price');
    
    if (!bookSelect || !quantityInput || !totalPriceInput) return;
    
    const updatePrice = () => {
        const selectedOption = bookSelect.options[bookSelect.selectedIndex];
        const price = parseFloat(selectedOption?.dataset.price || 0);
        const quantity = parseInt(quantityInput.value || 1);
        totalPriceInput.value = `$${(price * quantity).toFixed(2)}`;
    };
    
    bookSelect.addEventListener('change', updatePrice);
    quantityInput.addEventListener('input', updatePrice);
    updatePrice();
}

function filterBooks() {
    const searchTerm = document.getElementById('book-search').value.toLowerCase();
    const cards = document.querySelectorAll('.book-card');
    
    cards.forEach(card => {
        const title = card.querySelector('h3').textContent.toLowerCase();
        const author = card.querySelector('.book-author').textContent.toLowerCase();
        
        if (title.includes(searchTerm) || author.includes(searchTerm)) {
            card.style.display = 'block';
        } else {
            card.style.display = 'none';
        }
    });
}

function showError(elementId, message) {
    const el = document.getElementById(elementId);
    if (el) {
        el.textContent = message;
        el.style.display = 'block';
    }
}

function showNotification(type, message) {
    const notifications = document.getElementById('notifications');
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <i class="fas ${type === 'success' ? 'fa-check-circle' : type === 'error' ? 'fa-exclamation-circle' : 'fa-info-circle'}"></i>
        ${escapeHtml(message)}
    `;
    
    notifications.appendChild(notification);
    
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = String(text);
    return div.innerHTML;
}

function formatDate(dateString) {
    if (!dateString) return 'N/A';
    try {
        return new Date(dateString).toLocaleDateString();
    } catch (e) {
        return 'N/A';
    }
}

function normalizeData(data) {
    if (!data) return [];
    if (Array.isArray(data)) return data;
    if (typeof data === 'object') return Object.values(data);
    return [];
}

window.onclick = function(event) {
    const authModal = document.getElementById('auth-modal');
    const modal = document.getElementById('modal');
    
    if (event.target === authModal) closeAuthModal();
    if (event.target === modal) closeModal();
};