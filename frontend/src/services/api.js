const API_BASE_URL = import.meta.env.VITE_API_GATEWAY_URL || 'http://localhost:8080';

// Auth State Helpers
export const getToken = () => localStorage.getItem('token') || '';
export const setToken = (token) => localStorage.setItem('token', token);
export const getUser = () => {
  const user = localStorage.getItem('user');
  try {
    return user ? JSON.parse(user) : null;
  } catch {
    return null;
  }
};
export const setUser = (user) => localStorage.setItem('user', JSON.stringify(user));
export const clearAuth = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
};

// Generic fetch wrapper communicating strictly through the API Gateway
async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  const token = getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const currentUser = getUser();
  if (currentUser && currentUser.id) {
    headers['X-User-ID'] = currentUser.id.toString();
  }

  const config = {
    ...options,
    headers,
  };

  try {
    const response = await fetch(url, config);
    const contentType = response.headers.get('content-type');
    let data = null;
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    if (!response.ok) {
      const errorMessage = (data && data.error) || (data && data.details) || response.statusText || 'Request failed';
      throw new Error(errorMessage);
    }

    return data;
  } catch (err) {
    throw err;
  }
}

export const api = {
  // 1. Auth Service via Gateway
  register: (name, email, password) =>
    request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    }),

  login: (email, password) =>
    request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  getMe: () => request('/api/auth/me'),

  // 2. Product Service via Gateway
  getProducts: () => request('/api/products'),
  getProduct: (id) => request(`/api/products/${id}`),

  // 3. Cart Service via Gateway
  getCart: () => request('/api/cart'),
  addToCart: (productId, quantity, productName, price) =>
    request('/api/cart/items', {
      method: 'POST',
      body: JSON.stringify({
        product_id: productId,
        quantity,
        product_name: productName,
        price,
      }),
    }),
  updateCartItem: (productId, quantity) =>
    request(`/api/cart/items/${productId}`, {
      method: 'PUT',
      body: JSON.stringify({ quantity }),
    }),
  removeCartItem: (productId) =>
    request(`/api/cart/items/${productId}`, {
      method: 'DELETE',
    }),
  clearCart: () =>
    request('/api/cart', {
      method: 'DELETE',
    }),

  // 4. Order Service via Gateway
  createOrder: (items, shippingAddress) =>
    request('/api/orders', {
      method: 'POST',
      body: JSON.stringify({ items, shipping_address: shippingAddress }),
    }),
  getOrders: () => request('/api/orders'),
  getOrder: (id) => request(`/api/orders/${id}`),

  // 5. Payment Service via Gateway
  getPayment: (orderId) => request(`/api/payments/${orderId}`),

  // Gateway Health Check
  getHealth: () => request('/health'),
};
