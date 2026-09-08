import React, { useState, useEffect } from 'react';
import { api, getToken, setToken, getUser, setUser, clearAuth } from './services/api';

export default function App() {
  // Navigation & User State
  const [currentPage, setCurrentPage] = useState('products');
  const [selectedProductId, setSelectedProductId] = useState(null);
  const [selectedOrderId, setSelectedOrderId] = useState(null);
  const [currentUser, setCurrentUserState] = useState(getUser());

  // Global Cart State (item count)
  const [cartCount, setCartCount] = useState(0);

  // Notifications
  const [errorMsg, setErrorMsg] = useState('');
  const [successMsg, setSuccessMsg] = useState('');

  const clearMessages = () => {
    setErrorMsg('');
    setSuccessMsg('');
  };

  const navigateTo = (page, data = null) => {
    clearMessages();
    if (page === 'product-details' && data) {
      setSelectedProductId(data);
    }
    if (page === 'order-details' && data) {
      setSelectedOrderId(data);
    }
    setCurrentPage(page);
  };

  // Sync Cart Count
  const refreshCartCount = async () => {
    try {
      const cart = await api.getCart();
      const count = (cart.items || []).reduce((acc, it) => acc + it.quantity, 0);
      setCartCount(count);
    } catch {
      // Cart might be empty or service unreachable
    }
  };

  useEffect(() => {
    refreshCartCount();
  }, [currentUser]);

  const handleLogout = () => {
    clearAuth();
    setCurrentUserState(null);
    setCartCount(0);
    setSuccessMsg('Logged out successfully');
    navigateTo('login');
  };

  return (
    <div className="app-container">
      {/* Top Navigation Bar */}
      <header className="navbar">
        <div className="nav-brand" onClick={() => navigateTo('products')}>
          ⚡ <span>Distributed Store</span>
        </div>
        <nav className="nav-links">
          <button
            className={`nav-link ${currentPage === 'products' ? 'active' : ''}`}
            onClick={() => navigateTo('products')}
          >
            Products
          </button>
          <button
            className={`nav-link ${currentPage === 'cart' ? 'active' : ''}`}
            onClick={() => navigateTo('cart')}
          >
            Cart <span className="nav-badge">{cartCount}</span>
          </button>
          <button
            className={`nav-link ${currentPage === 'orders' ? 'active' : ''}`}
            onClick={() => navigateTo('orders')}
          >
            Orders
          </button>

          {currentUser ? (
            <>
              <span style={{ fontSize: '0.9rem', color: '#9ca3af', marginLeft: '0.5rem' }}>
                👤 {currentUser.name}
              </span>
              <button className="btn btn-secondary btn-sm" onClick={handleLogout}>
                Logout
              </button>
            </>
          ) : (
            <>
              <button
                className={`nav-link ${currentPage === 'login' ? 'active' : ''}`}
                onClick={() => navigateTo('login')}
              >
                Login
              </button>
              <button
                className={`btn btn-primary btn-sm`}
                onClick={() => navigateTo('register')}
              >
                Register
              </button>
            </>
          )}
        </nav>
      </header>

      {/* Main Body */}
      <main className="main-content">
        {errorMsg && <div className="alert alert-error">{errorMsg}</div>}
        {successMsg && <div className="alert alert-success">{successMsg}</div>}

        {/* 1. Register Page */}
        {currentPage === 'register' && (
          <RegisterPage
            onSuccess={(msg) => {
              setSuccessMsg(msg);
              navigateTo('login');
            }}
            onError={setErrorMsg}
            onNavigateLogin={() => navigateTo('login')}
          />
        )}

        {/* 2. Login Page */}
        {currentPage === 'login' && (
          <LoginPage
            onSuccess={(user) => {
              setCurrentUserState(user);
              setSuccessMsg(`Welcome back, ${user.name}!`);
              navigateTo('products');
            }}
            onError={setErrorMsg}
            onNavigateRegister={() => navigateTo('register')}
          />
        )}

        {/* 3. Products Catalog */}
        {currentPage === 'products' && (
          <ProductsPage
            onSelectProduct={(id) => navigateTo('product-details', id)}
            onAddToCart={async (product) => {
              try {
                await api.addToCart(product.id, 1, product.name, product.price);
                setSuccessMsg(`Added ${product.name} to cart!`);
                refreshCartCount();
              } catch (err) {
                setErrorMsg(err.message);
              }
            }}
            onError={setErrorMsg}
          />
        )}

        {/* 4. Product Details */}
        {currentPage === 'product-details' && (
          <ProductDetailsPage
            productId={selectedProductId}
            onBack={() => navigateTo('products')}
            onAddToCart={async (product, qty) => {
              try {
                await api.addToCart(product.id, qty, product.name, product.price);
                setSuccessMsg(`Added ${qty}x ${product.name} to cart!`);
                refreshCartCount();
              } catch (err) {
                setErrorMsg(err.message);
              }
            }}
            onError={setErrorMsg}
          />
        )}

        {/* 5. Cart Page */}
        {currentPage === 'cart' && (
          <CartPage
            onCheckout={() => navigateTo('checkout')}
            onCartUpdated={refreshCartCount}
            onError={setErrorMsg}
            onSuccess={setSuccessMsg}
            onBrowse={() => navigateTo('products')}
          />
        )}

        {/* 6. Checkout Page */}
        {currentPage === 'checkout' && (
          <CheckoutPage
            onOrderCreated={(order) => {
              setSuccessMsg(`Order #${order.id} placed successfully!`);
              refreshCartCount();
              navigateTo('order-details', order.id);
            }}
            onCancel={() => navigateTo('cart')}
            onError={setErrorMsg}
          />
        )}

        {/* 7. Orders List Page */}
        {currentPage === 'orders' && (
          <OrdersPage
            onViewOrder={(id) => navigateTo('order-details', id)}
            onError={setErrorMsg}
            onBrowse={() => navigateTo('products')}
          />
        )}

        {/* 8. Order Details Page */}
        {currentPage === 'order-details' && (
          <OrderDetailsPage
            orderId={selectedOrderId}
            onBack={() => navigateTo('orders')}
            onError={setErrorMsg}
          />
        )}
      </main>
    </div>
  );
}

// ==========================================
// 1. REGISTER PAGE
// ==========================================
function RegisterPage({ onSuccess, onError, onNavigateLogin }) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      await api.register(name, email, password);
      onSuccess('Registration successful! Please login.');
    } catch (err) {
      onError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card form-container">
      <h2 style={{ marginBottom: '1.5rem', textAlign: 'center' }}>Create Account</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Full Name</label>
          <input
            className="form-input"
            type="text"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="John Doe"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Email Address</label>
          <input
            className="form-input"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="john@example.com"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Password</label>
          <input
            className="form-input"
            type="password"
            required
            minLength={6}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 6 characters"
          />
        </div>
        <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
          {loading ? 'Registering...' : 'Register'}
        </button>
        <p style={{ marginTop: '1rem', textAlign: 'center', fontSize: '0.9rem', color: '#9ca3af' }}>
          Already have an account?{' '}
          <span style={{ color: '#818cf8', cursor: 'pointer' }} onClick={onNavigateLogin}>
            Login here
          </span>
        </p>
      </form>
    </div>
  );
}

// ==========================================
// 2. LOGIN PAGE
// ==========================================
function LoginPage({ onSuccess, onError, onNavigateRegister }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const resp = await api.login(email, password);
      if (resp.token && resp.user) {
        setToken(resp.token);
        setUser(resp.user);
        onSuccess(resp.user);
      } else {
        throw new Error('Invalid login response');
      }
    } catch (err) {
      onError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card form-container">
      <h2 style={{ marginBottom: '1.5rem', textAlign: 'center' }}>Account Login</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Email Address</label>
          <input
            className="form-input"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="john@example.com"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Password</label>
          <input
            className="form-input"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
          />
        </div>
        <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
        <p style={{ marginTop: '1rem', textAlign: 'center', fontSize: '0.9rem', color: '#9ca3af' }}>
          Don't have an account?{' '}
          <span style={{ color: '#818cf8', cursor: 'pointer' }} onClick={onNavigateRegister}>
            Register here
          </span>
        </p>
      </form>
    </div>
  );
}

// ==========================================
// 3. PRODUCTS PAGE
// ==========================================
function ProductsPage({ onSelectProduct, onAddToCart, onError }) {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getProducts()
      .then((data) => setProducts(data || []))
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
        <p style={{ marginTop: '1rem' }}>Loading product catalog...</p>
      </div>
    );
  }

  if (products.length === 0) {
    return <div className="empty-state">No products found.</div>;
  }

  return (
    <div>
      <h2 style={{ marginBottom: '1.5rem' }}>Product Catalog</h2>
      <div className="product-grid">
        {products.map((p) => (
          <div key={p.id} className="card product-card">
            <div>
              <div className="product-category">{p.category}</div>
              <h3
                className="product-title"
                style={{ cursor: 'pointer' }}
                onClick={() => onSelectProduct(p.id)}
              >
                {p.name}
              </h3>
              <p className="product-desc">{p.description}</p>
            </div>
            <div>
              <div className="product-price">${Number(p.price).toFixed(2)}</div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  className="btn btn-secondary btn-sm"
                  style={{ flex: 1 }}
                  onClick={() => onSelectProduct(p.id)}
                >
                  Details
                </button>
                <button
                  className="btn btn-primary btn-sm"
                  style={{ flex: 1 }}
                  onClick={() => onAddToCart(p)}
                >
                  Add to Cart
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ==========================================
// 4. PRODUCT DETAILS PAGE
// ==========================================
function ProductDetailsPage({ productId, onBack, onAddToCart, onError }) {
  const [product, setProduct] = useState(null);
  const [quantity, setQuantity] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getProduct(productId)
      .then(setProduct)
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  }, [productId]);

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="empty-state">
        <p>Product not found.</p>
        <button className="btn btn-secondary" style={{ marginTop: '1rem' }} onClick={onBack}>
          Back to Catalog
        </button>
      </div>
    );
  }

  return (
    <div className="card" style={{ maxWidth: '600px', margin: '0 auto' }}>
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: '1.5rem' }} onClick={onBack}>
        ← Back to Products
      </button>
      <div className="product-category">{product.category}</div>
      <h2 style={{ marginBottom: '1rem' }}>{product.name}</h2>
      <p style={{ color: '#9ca3af', marginBottom: '1.5rem', fontSize: '1.05rem' }}>
        {product.description}
      </p>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '2rem' }}>
        <div style={{ fontSize: '1.8rem', fontWeight: 700 }}>
          ${Number(product.price).toFixed(2)}
        </div>
        <div style={{ color: '#9ca3af', fontSize: '0.9rem' }}>
          In Stock: <strong style={{ color: '#fff' }}>{product.stock}</strong> units
        </div>
      </div>

      <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <label style={{ fontSize: '0.9rem', color: '#9ca3af' }}>Quantity:</label>
          <input
            className="form-input"
            type="number"
            min={1}
            max={product.stock || 99}
            value={quantity}
            onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
            style={{ width: '70px', textAlign: 'center' }}
          />
        </div>
        <button
          className="btn btn-primary"
          style={{ flex: 1 }}
          onClick={() => onAddToCart(product, quantity)}
        >
          Add to Cart
        </button>
      </div>
    </div>
  );
}

// ==========================================
// 5. CART PAGE
// ==========================================
function CartPage({ onCheckout, onCartUpdated, onError, onSuccess, onBrowse }) {
  const [cart, setCart] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchCart = () => {
    setLoading(true);
    api
      .getCart()
      .then(setCart)
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchCart();
  }, []);

  const handleUpdateQty = async (productId, delta, currentQty) => {
    const newQty = currentQty + delta;
    if (newQty <= 0) {
      handleRemove(productId);
      return;
    }
    try {
      await api.updateCartItem(productId, newQty);
      fetchCart();
      onCartUpdated();
    } catch (err) {
      onError(err.message);
    }
  };

  const handleRemove = async (productId) => {
    try {
      await api.removeCartItem(productId);
      fetchCart();
      onCartUpdated();
      onSuccess('Item removed from cart');
    } catch (err) {
      onError(err.message);
    }
  };

  const handleClear = async () => {
    try {
      await api.clearCart();
      fetchCart();
      onCartUpdated();
      onSuccess('Cart cleared');
    } catch (err) {
      onError(err.message);
    }
  };

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
      </div>
    );
  }

  const items = (cart && cart.items) || [];

  if (items.length === 0) {
    return (
      <div className="empty-state">
        <h2>Your Cart is Empty</h2>
        <p style={{ marginTop: '0.5rem', marginBottom: '1.5rem' }}>
          Looks like you haven't added any items yet.
        </p>
        <button className="btn btn-primary" onClick={onBrowse}>
          Browse Catalog
        </button>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2>Shopping Cart</h2>
        <button className="btn btn-danger btn-sm" onClick={handleClear}>
          Clear Cart
        </button>
      </div>

      <div className="card table-container">
        <table>
          <thead>
            <tr>
              <th>Product</th>
              <th>Price</th>
              <th style={{ textAlign: 'center' }}>Quantity</th>
              <th>Subtotal</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {items.map((it) => (
              <tr key={it.product_id}>
                <td style={{ fontWeight: 600 }}>{it.product_name || `Product #${it.product_id}`}</td>
                <td>${Number(it.price).toFixed(2)}</td>
                <td style={{ textAlign: 'center' }}>
                  <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem' }}>
                    <button
                      className="btn btn-secondary btn-sm"
                      style={{ padding: '0.2rem 0.5rem' }}
                      onClick={() => handleUpdateQty(it.product_id, -1, it.quantity)}
                    >
                      -
                    </button>
                    <span style={{ minWidth: '24px', display: 'inline-block' }}>{it.quantity}</span>
                    <button
                      className="btn btn-secondary btn-sm"
                      style={{ padding: '0.2rem 0.5rem' }}
                      onClick={() => handleUpdateQty(it.product_id, 1, it.quantity)}
                    >
                      +
                    </button>
                  </div>
                </td>
                <td style={{ fontWeight: 600 }}>${(it.price * it.quantity).toFixed(2)}</td>
                <td>
                  <button
                    className="btn btn-danger btn-sm"
                    onClick={() => handleRemove(it.product_id)}
                  >
                    Remove
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            gap: '2rem',
            marginTop: '1.5rem',
            paddingTop: '1rem',
            borderTop: '1px solid var(--panel-border)',
          }}
        >
          <div style={{ fontSize: '1.3rem' }}>
            Total:{' '}
            <strong style={{ color: '#fff', fontSize: '1.6rem' }}>
              ${Number(cart.total_price || 0).toFixed(2)}
            </strong>
          </div>
          <button className="btn btn-primary" onClick={onCheckout}>
            Proceed to Checkout →
          </button>
        </div>
      </div>
    </div>
  );
}

// ==========================================
// 6. CHECKOUT PAGE
// ==========================================
function CheckoutPage({ onOrderCreated, onCancel, onError }) {
  const [cart, setCart] = useState(null);
  const [shippingAddress, setShippingAddress] = useState('100 Distributed Systems Way, Suite 400');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    api
      .getCart()
      .then((c) => {
        if (!c || !c.items || c.items.length === 0) {
          onCancel();
        } else {
          setCart(c);
        }
      })
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const handleCreateOrder = async (e) => {
    e.preventDefault();
    if (!shippingAddress.trim()) {
      onError('Please provide a shipping address');
      return;
    }

    setSubmitting(true);
    try {
      const orderItems = cart.items.map((it) => ({
        product_id: it.product_id,
        quantity: it.quantity,
        price: it.price,
      }));

      // 1. Create order through Gateway -> Order Service
      const order = await api.createOrder(orderItems, shippingAddress);

      // 2. Clear Cart
      await api.clearCart();

      onOrderCreated(order);
    } catch (err) {
      onError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading || !cart) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '650px', margin: '0 auto' }}>
      <h2 style={{ marginBottom: '1.5rem' }}>Checkout & Order Placement</h2>
      <div className="card">
        <h3 style={{ marginBottom: '1rem', borderBottom: '1px solid var(--panel-border)', paddingBottom: '0.5rem' }}>
          Order Summary
        </h3>
        <div style={{ marginBottom: '1.5rem' }}>
          {cart.items.map((it) => (
            <div
              key={it.product_id}
              style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}
            >
              <span>
                {it.product_name || `Product #${it.product_id}`} × {it.quantity}
              </span>
              <strong style={{ color: '#fff' }}>${(it.price * it.quantity).toFixed(2)}</strong>
            </div>
          ))}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              marginTop: '1rem',
              paddingTop: '1rem',
              borderTop: '1px dashed var(--panel-border)',
              fontSize: '1.2rem',
            }}
          >
            <span>Total Payable</span>
            <strong style={{ color: '#818cf8' }}>${Number(cart.total_price).toFixed(2)}</strong>
          </div>
        </div>

        <form onSubmit={handleCreateOrder}>
          <div className="form-group">
            <label className="form-label">Shipping Address</label>
            <textarea
              className="form-input"
              rows={3}
              required
              value={shippingAddress}
              onChange={(e) => setShippingAddress(e.target.value)}
              placeholder="Street address, City, Country..."
            />
          </div>

          <div style={{ display: 'flex', gap: '1rem', marginTop: '1.5rem' }}>
            <button
              type="button"
              className="btn btn-secondary"
              style={{ flex: 1 }}
              onClick={onCancel}
              disabled={submitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              style={{ flex: 2 }}
              disabled={submitting}
            >
              {submitting ? 'Placing Order...' : 'Place Order & Pay →'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ==========================================
// 7. ORDERS PAGE
// ==========================================
function OrdersPage({ onViewOrder, onError, onBrowse }) {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getOrders()
      .then((data) => setOrders(data || []))
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
        <p style={{ marginTop: '1rem' }}>Loading your orders...</p>
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="empty-state">
        <h2>No Orders Yet</h2>
        <p style={{ marginTop: '0.5rem', marginBottom: '1.5rem' }}>
          You haven't placed any orders yet.
        </p>
        <button className="btn btn-primary" onClick={onBrowse}>
          Start Shopping
        </button>
      </div>
    );
  }

  return (
    <div>
      <h2 style={{ marginBottom: '1.5rem' }}>Your Orders</h2>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {orders.map((o) => (
          <div
            key={o.id}
            className="card"
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              cursor: 'pointer',
            }}
            onClick={() => onViewOrder(o.id)}
          >
            <div>
              <div style={{ fontSize: '1.15rem', fontWeight: 600, marginBottom: '0.3rem' }}>
                Order #{o.id}
              </div>
              <div style={{ fontSize: '0.85rem', color: '#9ca3af' }}>
                Placed on: {new Date(o.created_at).toLocaleString()} • {o.items?.length || 0} items
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>
                ${Number(o.total_amount).toFixed(2)}
              </div>
              <span className={`badge ${o.status === 'PENDING' ? 'badge-pending' : 'badge-success'}`}>
                {o.status}
              </span>
              <button className="btn btn-secondary btn-sm">View Details →</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ==========================================
// 8. ORDER DETAILS PAGE (Includes RabbitMQ Payment Status)
// ==========================================
function OrderDetailsPage({ orderId, onBack, onError }) {
  const [order, setOrder] = useState(null);
  const [payment, setPayment] = useState(null);
  const [loading, setLoading] = useState(true);
  const [checkingPayment, setCheckingPayment] = useState(false);

  const fetchOrderAndPayment = async () => {
    setLoading(true);
    try {
      const ord = await api.getOrder(orderId);
      setOrder(ord);

      // Check payment status from Payment Service via Gateway
      try {
        const pay = await api.getPayment(orderId);
        setPayment(pay);
      } catch {
        // Payment might still be processing through RabbitMQ
        setPayment(null);
      }
    } catch (err) {
      onError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const recheckPayment = async () => {
    setCheckingPayment(true);
    try {
      const pay = await api.getPayment(orderId);
      setPayment(pay);
    } catch (err) {
      onError('Payment is still processing in RabbitMQ queue...');
    } finally {
      setCheckingPayment(false);
    }
  };

  useEffect(() => {
    fetchOrderAndPayment();
    // Auto polling for 5 seconds to catch async RabbitMQ processing
    const timer = setTimeout(() => {
      api
        .getPayment(orderId)
        .then(setPayment)
        .catch(() => {});
    }, 1500);
    return () => clearTimeout(timer);
  }, [orderId]);

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
      </div>
    );
  }

  if (!order) {
    return (
      <div className="empty-state">
        <p>Order not found.</p>
        <button className="btn btn-secondary" onClick={onBack}>
          Back to Orders
        </button>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '700px', margin: '0 auto' }}>
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: '1.5rem' }} onClick={onBack}>
        ← Back to Orders
      </button>

      <div className="card" style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h2>Order #{order.id}</h2>
          <span className={`badge ${order.status === 'PENDING' ? 'badge-pending' : 'badge-success'}`}>
            Order Status: {order.status}
          </span>
        </div>

        <p style={{ fontSize: '0.85rem', color: '#9ca3af', marginBottom: '1rem' }}>
          Created: {new Date(order.created_at).toLocaleString()}
        </p>

        {order.shipping_address && (
          <div style={{ marginBottom: '1rem', fontSize: '0.9rem' }}>
            <span style={{ color: '#9ca3af' }}>Shipping to:</span> <strong>{order.shipping_address}</strong>
          </div>
        )}

        {/* Payment Status Card */}
        <div
          style={{
            background: '#0d1322',
            border: '1px solid var(--panel-border)',
            borderRadius: '8px',
            padding: '1rem',
            marginTop: '1.5rem',
            marginBottom: '1rem',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div style={{ fontWeight: 600, marginBottom: '0.2rem' }}>RabbitMQ Payment Pipeline</div>
              <div style={{ fontSize: '0.85rem', color: '#9ca3af' }}>
                Processed asynchronously by Payment Service via <code>OrderCreated</code> event
              </div>
            </div>
            {payment ? (
              <span className="badge badge-success">PAYMENT {payment.status}</span>
            ) : (
              <span className="badge badge-pending">PROCESSING...</span>
            )}
          </div>

          {payment && (
            <div style={{ marginTop: '0.8rem', fontSize: '0.85rem', color: '#9ca3af' }}>
              <div>
                Transaction ID: <code style={{ color: '#818cf8' }}>{payment.transaction_id}</code>
              </div>
              <div>Method: {payment.payment_method}</div>
              <div>Settled: {new Date(payment.created_at).toLocaleTimeString()}</div>
            </div>
          )}

          {!payment && (
            <div style={{ marginTop: '0.8rem' }}>
              <button
                className="btn btn-secondary btn-sm"
                onClick={recheckPayment}
                disabled={checkingPayment}
              >
                {checkingPayment ? 'Checking...' : 'Refresh Payment Status'}
              </button>
            </div>
          )}
        </div>

        {/* Order Items Table */}
        <h3 style={{ marginTop: '1.5rem', marginBottom: '0.8rem' }}>Ordered Items</h3>
        <table style={{ width: '100%' }}>
          <thead>
            <tr>
              <th>Item</th>
              <th>Quantity</th>
              <th>Unit Price</th>
              <th style={{ textAlign: 'right' }}>Total</th>
            </tr>
          </thead>
          <tbody>
            {(order.items || []).map((it) => (
              <tr key={it.id}>
                <td>{it.product_name || `Product #${it.product_id}`}</td>
                <td>{it.quantity}</td>
                <td>${Number(it.price).toFixed(2)}</td>
                <td style={{ textAlign: 'right', fontWeight: 600 }}>
                  ${(it.price * it.quantity).toFixed(2)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div style={{ textAlign: 'right', marginTop: '1.5rem', fontSize: '1.3rem' }}>
          Grand Total:{' '}
          <strong style={{ color: '#fff', fontSize: '1.6rem' }}>
            ${Number(order.total_amount).toFixed(2)}
          </strong>
        </div>
      </div>
    </div>
  );
}
