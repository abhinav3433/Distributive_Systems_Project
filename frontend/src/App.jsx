import React, { useState, useEffect } from 'react';
import { api, getToken, setToken, getUser, setUser, clearAuth } from './services/api';
import { ProductImage, getProductMeta } from './services/productImages';

export function ShazeLogo({ size = 36, showWordmark = true, variant = 'light' }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', userSelect: 'none' }}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 64 64"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ flexShrink: 0, borderRadius: '10px', boxShadow: '0 2px 8px rgba(0,0,0,0.2)' }}
      >
        <rect width="64" height="64" rx="14" fill="#17382A" />
        <path
          d="M44 20C44 15.58 40.42 12 36 12H24C19.58 12 16 15.58 16 20C16 24.42 19.58 28 24 28H40C44.42 28 48 31.58 48 36C48 40.42 44.42 44 40 44H26"
          stroke="#D63A2B"
          strokeWidth="6.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M20 44C20 48.42 23.58 52 28 52H40C44.42 52 48 48.42 48 44"
          stroke="#FBF5EA"
          strokeWidth="6.5"
          strokeLinecap="round"
        />
        <circle cx="44" cy="20" r="3.5" fill="#F2A93B" />
        <circle cx="20" cy="44" r="3.5" fill="#F2A93B" />
      </svg>
      {showWordmark && (
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <span
            style={{
              fontFamily: "'Bricolage Grotesque', sans-serif",
              fontSize: `${Math.round(size * 0.58)}px`,
              fontWeight: 800,
              letterSpacing: '-0.035em',
              lineHeight: 1,
              color: variant === 'light' ? '#FBF5EA' : '#17382A',
            }}
          >
            Shaze<span style={{ color: '#F2A93B' }}>.</span>
          </span>
          <span
            style={{
              fontFamily: "'Plus Jakarta Sans', sans-serif",
              fontSize: '0.64rem',
              fontWeight: 600,
              letterSpacing: '0.12em',
              textTransform: 'uppercase',
              color: variant === 'light' ? '#DCE7DA' : '#728177',
              marginTop: '0.2rem',
            }}
          >
            Curated Living
          </span>
        </div>
      )}
    </div>
  );
}

export default function App() {
  // Navigation & User State
  const [currentPage, setCurrentPage] = useState('products');
  const [selectedProductId, setSelectedProductId] = useState(null);
  const [selectedOrderId, setSelectedOrderId] = useState(null);
  const [currentUser, setCurrentUserState] = useState(getUser());

  // Global Catalog Filters & Search (synced between navbar, subnav, and catalog)
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('All');
  const [sortOption, setSortOption] = useState('featured');

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
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Sync Cart Count
  const refreshCartCount = async () => {
    try {
      const cart = await api.getCart();
      const count = (cart.items || []).reduce((acc, it) => acc + it.quantity, 0);
      setCartCount(count);
    } catch {
      // Cart might be empty or unauthenticated
    }
  };

  useEffect(() => {
    refreshCartCount();
  }, [currentUser]);

  const handleLogout = () => {
    clearAuth();
    setCurrentUserState(null);
    setCartCount(0);
    setSuccessMsg('You have been logged out.');
    navigateTo('login');
  };

  const handleSearchSubmit = (e) => {
    if (e) e.preventDefault();
    if (currentPage !== 'products') {
      navigateTo('products');
    }
  };

  const handleCategorySelect = (cat) => {
    setSelectedCategory(cat);
    if (currentPage !== 'products') {
      navigateTo('products');
    }
  };

  const categoriesList = ['All', 'Electronics', 'Audio', 'Accessories', 'Wearables', 'Photography', 'Gaming'];

  return (
    <div className="app-container">
      {/* ==========================================
          HEADER & TOP NAVIGATION BAR
          ========================================== */}
      <header className="navbar-header">
        <div className="navbar">
          {/* Brand Logo & Name */}
          <div
            className="nav-brand"
            onClick={() => {
              setSearchQuery('');
              setSelectedCategory('All');
              navigateTo('products');
            }}
          >
            <ShazeLogo size={36} variant="light" />
          </div>

          {/* Integrated Global Search Bar */}
          <form className="nav-search-bar" onSubmit={handleSearchSubmit}>
            <input
              type="text"
              className="nav-search-input"
              placeholder="Search curated electronics, studio audio, photography, gaming..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button
                type="button"
                className="nav-search-clear"
                onClick={() => setSearchQuery('')}
                title="Clear search"
              >
                ✕
              </button>
            )}
            <button type="submit" className="nav-search-btn" title="Search">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
            </button>
          </form>

          {/* Nav Links & Actions */}
          <nav className="nav-links">
            <button
              className={`nav-link ${currentPage === 'products' ? 'active' : ''}`}
              onClick={() => navigateTo('products')}
            >
              <span>Catalog</span>
            </button>

            <button
              className={`nav-link ${currentPage === 'orders' ? 'active' : ''}`}
              onClick={() => navigateTo('orders')}
            >
              <span>Orders</span>
            </button>

            <button
              className={`nav-link ${currentPage === 'cart' ? 'active' : ''}`}
              onClick={() => navigateTo('cart')}
            >
              <span>🛒 Cart</span>
              {cartCount > 0 && <span className="nav-badge">{cartCount}</span>}
            </button>

            {currentUser ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.45rem', marginLeft: '0.25rem' }}>
                <div className="nav-user-chip" title={currentUser.email}>
                  <span className="nav-user-avatar">{currentUser.name ? currentUser.name[0].toUpperCase() : 'U'}</span>
                  <span style={{ fontWeight: 600 }}>{currentUser.name}</span>
                </div>
                <button className="btn btn-secondary btn-sm" onClick={handleLogout} style={{ padding: '0.35rem 0.65rem' }}>
                  Logout
                </button>
              </div>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', marginLeft: '0.25rem' }}>
                <button
                  className={`nav-link ${currentPage === 'login' ? 'active' : ''}`}
                  onClick={() => navigateTo('login')}
                >
                  Sign In
                </button>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => navigateTo('register')}
                >
                  Register
                </button>
              </div>
            )}
          </nav>
        </div>

        {/* Category Sub-Nav Bar (Amazon/Myntra inspired quick browse) */}
        <div className="sub-nav">
          <div className="sub-nav-inner">
            {categoriesList.map((cat) => (
              <button
                key={cat}
                className={`sub-nav-item ${selectedCategory === cat ? 'active' : ''}`}
                onClick={() => handleCategorySelect(cat)}
              >
                {cat === 'All' ? '⚡ All Categories' : cat}
              </button>
            ))}
          </div>
        </div>
      </header>

      {/* ==========================================
          MAIN BODY CONTENT
          ========================================== */}
      <main className="main-content">
        {/* Flash Notifications */}
        {errorMsg && (
          <div className="alert alert-error">
            <span>{errorMsg}</span>
            <button
              onClick={clearMessages}
              style={{ background: 'transparent', border: 'none', color: 'inherit', cursor: 'pointer', fontWeight: 700 }}
            >
              ✕
            </button>
          </div>
        )}
        {successMsg && (
          <div className="alert alert-success">
            <span>{successMsg}</span>
            <button
              onClick={clearMessages}
              style={{ background: 'transparent', border: 'none', color: 'inherit', cursor: 'pointer', fontWeight: 700 }}
            >
              ✕
            </button>
          </div>
        )}

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

        {/* 3. Products Catalog Page */}
        {currentPage === 'products' && (
          <ProductsPage
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            selectedCategory={selectedCategory}
            setSelectedCategory={setSelectedCategory}
            sortOption={sortOption}
            setSortOption={setSortOption}
            onSelectProduct={(id) => navigateTo('product-details', id)}
            onAddToCart={async (product) => {
              try {
                await api.addToCart(product.id, 1, product.name, product.price);
                setSuccessMsg(`Added "${product.name}" to cart.`);
                refreshCartCount();
              } catch (err) {
                setErrorMsg(err.message);
              }
            }}
            onError={setErrorMsg}
          />
        )}

        {/* 4. Product Details Page */}
        {currentPage === 'product-details' && (
          <ProductDetailsPage
            productId={selectedProductId}
            onBack={() => navigateTo('products')}
            onAddToCart={async (product, qty) => {
              try {
                await api.addToCart(product.id, qty, product.name, product.price);
                setSuccessMsg(`Added ${qty}× "${product.name}" to cart.`);
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
              setSuccessMsg(`Order #${order.id} confirmed! Dispatched to RabbitMQ for settlement.`);
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

        {/* 8. Order Details Page (RabbitMQ Timeline) */}
        {currentPage === 'order-details' && (
          <OrderDetailsPage
            orderId={selectedOrderId}
            onBack={() => navigateTo('orders')}
            onError={setErrorMsg}
          />
        )}
      </main>

      {/* Shaze Editorial Footer */}
      <ShazeFooter
        onCategorySelect={handleCategorySelect}
        onNavigate={navigateTo}
      />
    </div>
  );
}

// ==========================================
// SHAZE FOOTER COMPONENT
// ==========================================
function ShazeFooter({ onCategorySelect, onNavigate }) {
  return (
    <footer className="shaze-footer">
      <div className="shaze-footer-inner">
        <div className="shaze-footer-grid">
          <div>
            <div style={{ marginBottom: '1.15rem' }}>
              <ShazeLogo size={36} variant="light" />
            </div>
            <p className="shaze-footer-brand-desc">
              Thoughtful design meets verified engineering. High-performance computing, studio noise-cancelling audio, precision photography gear, and mechanical gaming peripherals.
            </p>
          </div>

          <div>
            <h4 className="shaze-footer-heading">Collections</h4>
            <ul className="shaze-footer-links">
              {['Electronics', 'Audio', 'Accessories', 'Wearables', 'Photography', 'Gaming'].map((cat) => (
                <li key={cat}>
                  <button
                    className="shaze-footer-link"
                    onClick={() => {
                      onCategorySelect(cat);
                      onNavigate('products');
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
                  >
                    {cat}
                  </button>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="shaze-footer-heading">Navigation</h4>
            <ul className="shaze-footer-links">
              <li>
                <button
                  className="shaze-footer-link"
                  onClick={() => {
                    onCategorySelect('All');
                    onNavigate('products');
                  }}
                >
                  All 26 Curated Products
                </button>
              </li>
              <li>
                <button className="shaze-footer-link" onClick={() => onNavigate('cart')}>
                  Shopping Cart
                </button>
              </li>
              <li>
                <button className="shaze-footer-link" onClick={() => onNavigate('orders')}>
                  Tracked Orders
                </button>
              </li>
              <li>
                <button className="shaze-footer-link" onClick={() => onNavigate('login')}>
                  Customer Sign In
                </button>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="shaze-footer-heading">System Trust</h4>
            <div style={{ fontSize: '0.82rem', color: 'rgba(251, 245, 234, 0.78)', lineHeight: 1.65 }}>
              <div>• Distributed Microservice Fabric</div>
              <div>• Internal gRPC RPC Layer</div>
              <div>• RabbitMQ Asynchronous Settlement</div>
              <div>• PostgreSQL &amp; Redis Cache</div>
              <div style={{ marginTop: '0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.45rem', color: '#F2A93B', fontWeight: 600, fontSize: '0.78rem' }}>
                <span style={{ width: 7, height: 7, borderRadius: '50%', background: '#F2A93B', display: 'inline-block' }}></span>
                All Services Healthy
              </div>
            </div>
          </div>
        </div>

        <div className="shaze-footer-bottom">
          <div>© {new Date().getFullYear()} Shaze E-Commerce Platform. All rights reserved.</div>
          <div style={{ display: 'flex', gap: '1.25rem' }}>
            <span>Curated Living</span>
            <span>Express Delivery</span>
            <span>Secure Checkout</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

// ==========================================
// HERO BANNER COMPONENT
// ==========================================
function HeroBanner({ onShopNow, onExploreCategories }) {
  return (
    <div className="hero-container">
      <div className="hero-banner">
        <div className="hero-content">
          <div className="hero-badge">
            <span className="hero-badge-dot" />
            <span>Shaze Curated Selection • 2026 Edition</span>
          </div>
          <h1 className="hero-title">
            Curated Hardware.
            <span>Designed for Intentional Living.</span>
          </h1>
          <p className="hero-subtitle">
            Explore 26 handpicked items across high-performance computing, studio noise-cancelling audio, precision photography, and mechanical gaming peripherals. Distributed reliability with real-time async settlement.
          </p>
          <div className="hero-actions">
            <button className="btn btn-primary btn-lg" onClick={onShopNow}>
              Explore Catalogue ↓
            </button>
            <button className="btn btn-sage btn-lg" onClick={onExploreCategories}>
              Browse Collections
            </button>
          </div>
        </div>
      </div>

      {/* Value Perks Strip */}
      <div className="hero-perks">
        <div className="perk-card">
          <div className="perk-icon">📦</div>
          <div>
            <div className="perk-text-title">Express Delivery</div>
            <div className="perk-text-desc">Complimentary shipping on all orders</div>
          </div>
        </div>
        <div className="perk-card">
          <div className="perk-icon">🌿</div>
          <div>
            <div className="perk-text-title">Curated Quality</div>
            <div className="perk-text-desc">Individually evaluated hardware standards</div>
          </div>
        </div>
        <div className="perk-card">
          <div className="perk-icon">🔄</div>
          <div>
            <div className="perk-text-title">30-Day Guarantee</div>
            <div className="perk-text-desc">Effortless returns &amp; full satisfaction</div>
          </div>
        </div>
        <div className="perk-card">
          <div className="perk-icon">⚡</div>
          <div>
            <div className="perk-text-title">Async Settlement</div>
            <div className="perk-text-desc">RabbitMQ event-driven choreography</div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ==========================================
// 3. PRODUCTS CATALOG PAGE
// ==========================================
function ProductsPage({
  searchQuery,
  setSearchQuery,
  selectedCategory,
  setSelectedCategory,
  sortOption,
  setSortOption,
  onSelectProduct,
  onAddToCart,
  onError,
}) {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [addingId, setAddingId] = useState(null);

  useEffect(() => {
    api
      .getProducts()
      .then((data) => setProducts(data || []))
      .catch((err) => onError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const CANONICAL_CATEGORIES = ['All', 'Electronics', 'Audio', 'Accessories', 'Wearables', 'Photography', 'Gaming'];
  const dynamicCategories = Array.from(new Set(products.map((p) => (p.category || '').trim()).filter(Boolean)));
  const categories = [
    'All',
    ...CANONICAL_CATEGORIES.slice(1),
    ...dynamicCategories.filter((c) => !CANONICAL_CATEGORIES.some((cat) => cat.toLowerCase() === c.toLowerCase())),
  ];

  // Filter & Search with case/whitespace normalization
  let filtered = products.filter((p) => {
    const pCat = (p.category || '').trim().toLowerCase();
    const sCat = (selectedCategory || 'All').trim().toLowerCase();
    const matchesCategory = sCat === 'all' || pCat === sCat;
    const q = searchQuery.toLowerCase().trim();
    const matchesSearch =
      !q ||
      (p.name && p.name.toLowerCase().includes(q)) ||
      (p.description && p.description.toLowerCase().includes(q)) ||
      (p.category && p.category.toLowerCase().includes(q));
    return matchesCategory && matchesSearch;
  });

  // Sorting with reliable numerical parsing
  if (sortOption === 'price-asc') {
    filtered.sort((a, b) => Number(a.price) - Number(b.price));
  } else if (sortOption === 'price-desc') {
    filtered.sort((a, b) => Number(b.price) - Number(a.price));
  } else if (sortOption === 'name-asc') {
    filtered.sort((a, b) => (a.name || '').localeCompare(b.name || ''));
  } else if (sortOption === 'rating-desc') {
    filtered.sort((a, b) => {
      const rA = getProductMeta(a).rating || 0;
      const rB = getProductMeta(b).rating || 0;
      return rB - rA;
    });
  }

  const handleAddWithFeedback = async (p) => {
    setAddingId(p.id);
    try {
      await onAddToCart(p);
    } finally {
      setAddingId(null);
    }
  };

  const scrollToGrid = () => {
    const grid = document.getElementById('product-catalog-grid');
    if (grid) {
      grid.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  };

  if (loading) {
    return (
      <div className="empty-state">
        <div className="spinner"></div>
        <p style={{ marginTop: '1rem', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
          Loading product catalog from microservices...
        </p>
      </div>
    );
  }

  return (
    <div>
      {/* Hero Banner displayed when browsing initial catalog */}
      {selectedCategory === 'All' && !searchQuery && (
        <HeroBanner
          onShopNow={scrollToGrid}
          onExploreCategories={() => {
            setSelectedCategory('Electronics');
            scrollToGrid();
          }}
        />
      )}

      {/* Catalog Toolbar */}
      <div className="catalog-toolbar" id="product-catalog-grid">
        <div className="catalog-title-group">
          <h1 className="catalog-title">Curated Collections</h1>
          <p className="catalog-count-pill">
            Showing {filtered.length} of {products.length} distinct items
            {selectedCategory !== 'All' && <span> in <strong>{selectedCategory}</strong></span>}
            {searchQuery && <span> matching "<strong>{searchQuery}</strong>"</span>}
          </p>
        </div>

        <div className="catalog-controls">
          {searchQuery && (
            <button
              className="btn btn-secondary btn-sm"
              onClick={() => setSearchQuery('')}
            >
              Clear Search ✕
            </button>
          )}

          <label style={{ fontSize: '0.82rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Sort By:</label>
          <select
            className="sort-select"
            value={sortOption}
            onChange={(e) => setSortOption(e.target.value)}
          >
            <option value="featured">Featured Picks</option>
            <option value="price-asc">Price: Low to High</option>
            <option value="price-desc">Price: High to Low</option>
            <option value="rating-desc">Highest Rated</option>
            <option value="name-asc">Name: A to Z</option>
          </select>
        </div>
      </div>

      {/* Category Filter Tabs */}
      <div className="category-tabs-container">
        <div className="category-tabs">
          {categories.map((cat) => {
            const count = cat === 'All' ? products.length : products.filter((p) => p.category === cat).length;
            return (
              <button
                key={cat}
                className={`category-tab ${selectedCategory === cat ? 'active' : ''}`}
                onClick={() => setSelectedCategory(cat)}
              >
                <span>{cat === 'All' ? '✦ All Collections' : cat}</span>
                <span className="category-tab-count">{count}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Product Grid */}
      {filtered.length === 0 ? (
        <div className="empty-state card">
          <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>🔍</div>
          <h2 style={{ fontSize: '1.15rem', fontWeight: 600 }}>No matching products found</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginTop: '0.25rem' }}>
            Try searching with different keywords or switch the category filter.
          </p>
          <button
            className="btn btn-primary btn-sm"
            style={{ marginTop: '1.25rem' }}
            onClick={() => {
              setSearchQuery('');
              setSelectedCategory('All');
            }}
          >
            Reset Filters
          </button>
        </div>
      ) : (
        <div className="product-grid">
          {filtered.map((p) => {
            const meta = getProductMeta(p);
            const discountPct = meta.originalPrice
              ? Math.round(((meta.originalPrice - p.price) / meta.originalPrice) * 100)
              : null;

            return (
              <div key={p.id} className="product-card">
                {/* Product Image Frame */}
                <div
                  className="product-card-img-wrapper"
                  onClick={() => onSelectProduct(p.id)}
                  title={`View details for ${p.name}`}
                >
                  {meta.badge && <span className="product-badge-tag">{meta.badge}</span>}
                  <span className={`product-stock-pill ${p.stock > 0 ? 'stock-pill-in' : 'stock-pill-out'}`}>
                    {p.stock > 0 ? `${p.stock} in stock` : 'Out of Stock'}
                  </span>
                  <ProductImage product={p} className="product-card-img" />
                </div>

                {/* Product Card Body */}
                <div className="product-card-body">
                  <div className="product-card-meta">
                    <span className="product-category-tag">{p.category}</span>
                    <div className="rating-stars" title={`${meta.rating} out of 5 stars`}>
                      <span>★</span>
                      <span>{meta.rating}</span>
                      <span className="rating-count">({meta.reviews})</span>
                    </div>
                  </div>

                  <h3
                    className="product-title"
                    onClick={() => onSelectProduct(p.id)}
                  >
                    {p.name}
                  </h3>

                  <p className="product-desc">{p.description}</p>

                  {/* Price Section */}
                  <div className="product-pricing-box">
                    <div className="product-price-row">
                      <span className="product-price">${Number(p.price).toFixed(2)}</span>
                      {meta.originalPrice && (
                        <span className="product-original-price">
                          ${meta.originalPrice.toFixed(2)}
                        </span>
                      )}
                      {discountPct > 0 && (
                        <span className="product-discount-badge">Save {discountPct}%</span>
                      )}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="product-card-actions">
                    <button
                      className="btn btn-secondary btn-sm"
                      style={{ flex: 1 }}
                      onClick={() => onSelectProduct(p.id)}
                    >
                      Details
                    </button>
                    <button
                      className="btn btn-primary btn-sm"
                      style={{ flex: 1.3 }}
                      onClick={() => handleAddWithFeedback(p)}
                      disabled={p.stock <= 0 || addingId === p.id}
                    >
                      {addingId === p.id ? 'Adding...' : '🛒 Add to Cart'}
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
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
  const [adding, setAdding] = useState(false);

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
      <div className="empty-state card">
        <p>Product record not found.</p>
        <button className="btn btn-secondary btn-sm" style={{ marginTop: '1rem' }} onClick={onBack}>
          ← Return to Catalog
        </button>
      </div>
    );
  }

  const meta = getProductMeta(product);
  const discountPct = meta.originalPrice
    ? Math.round(((meta.originalPrice - product.price) / meta.originalPrice) * 100)
    : null;

  const handleAdd = async () => {
    setAdding(true);
    try {
      await onAddToCart(product, quantity);
    } finally {
      setAdding(false);
    }
  };

  return (
    <div className="product-details-container">
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: '1.25rem' }} onClick={onBack}>
        ← Back to Catalog
      </button>

      <div className="product-details-card">
        {/* Left: Product Image & Badges */}
        <div className="product-details-media">
          {meta.badge && (
            <span className="product-badge-tag" style={{ position: 'absolute', top: 16, left: 16 }}>
              {meta.badge}
            </span>
          )}
          <span
            className={`product-stock-pill ${product.stock > 0 ? 'stock-pill-in' : 'stock-pill-out'}`}
            style={{ position: 'absolute', top: 16, right: 16 }}
          >
            {product.stock > 0 ? `${product.stock} Units In Stock` : 'Out of Stock'}
          </span>

          <ProductImage product={product} className="product-details-img" />

          {/* Guarantee Badges */}
          <div style={{ display: 'flex', gap: '1rem', marginTop: '1.5rem', width: '100%', justifyContent: 'center' }}>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textAlign: 'center' }}>
              ✓ Genuine Brand
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textAlign: 'center' }}>
              ✓ Free Shipping
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textAlign: 'center' }}>
              ✓ 2-Year Warranty
            </div>
          </div>
        </div>

        {/* Right: Product Information & Checkout Control */}
        <div className="product-details-info">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
            <span className="product-category-tag">{product.category}</span>
            <div className="rating-stars">
              <span>★</span>
              <span>{meta.rating}</span>
              <span className="rating-count">({meta.reviews} verified reviews)</span>
            </div>
          </div>

          <h1 style={{ fontSize: '1.65rem', fontWeight: 800, marginBottom: '0.75rem', color: 'var(--shaze-ink)', lineHeight: 1.25 }}>
            {product.name}
          </h1>

          <p style={{ color: 'var(--text-secondary)', fontSize: '0.925rem', lineHeight: 1.6, marginBottom: '1.5rem' }}>
            {product.description}
          </p>

          {/* Pricing Row */}
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <span style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--shaze-ink)', fontFamily: 'var(--font-mono)' }}>
              ${Number(product.price).toFixed(2)}
            </span>
            {meta.originalPrice && (
              <span style={{ fontSize: '1.1rem', color: 'var(--text-muted)', textDecoration: 'line-through', fontFamily: 'var(--font-mono)' }}>
                ${meta.originalPrice.toFixed(2)}
              </span>
            )}
            {discountPct > 0 && (
              <span className="product-discount-badge" style={{ fontSize: '0.8rem', padding: '0.2rem 0.5rem' }}>
                Save {discountPct}%
              </span>
            )}
          </div>

          {/* Technical Specifications Table */}
          <table className="details-specs-table">
            <tbody>
              <tr>
                <td className="label">Category</td>
                <td className="value">{product.category}</td>
              </tr>
              <tr>
                <td className="label">Availability</td>
                <td className="value" style={{ color: product.stock > 0 ? 'var(--success-text)' : 'var(--danger-text)' }}>
                  {product.stock > 0 ? `In Stock (${product.stock} units)` : 'Out of Stock'}
                </td>
              </tr>
              <tr>
                <td className="label">Delivery</td>
                <td className="value">Standard Express (2-3 business days)</td>
              </tr>
              <tr>
                <td className="label">Hardware SKU</td>
                <td className="value"><code>EDP-PROD-00{product.id}</code></td>
              </tr>
            </tbody>
          </table>

          {/* Stepper & Add to Cart Controls */}
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginTop: '1.25rem' }}>
            <div className="quantity-stepper">
              <button
                type="button"
                className="quantity-stepper-btn"
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
              >
                −
              </button>
              <input
                type="text"
                readOnly
                className="quantity-stepper-input"
                value={quantity}
              />
              <button
                type="button"
                className="quantity-stepper-btn"
                onClick={() => setQuantity(Math.min(product.stock || 99, quantity + 1))}
              >
                +
              </button>
            </div>

            <button
              className="btn btn-primary"
              style={{ flex: 1, padding: '0.75rem 1.25rem', fontSize: '0.95rem' }}
              onClick={handleAdd}
              disabled={product.stock <= 0 || adding}
            >
              {adding ? 'Adding to Cart...' : '🛒 Add to Shopping Cart'}
            </button>
          </div>
        </div>
      </div>
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
      onSuccess('Account created successfully. Please sign in with your credentials.');
    } catch (err) {
      onError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card form-container">
      <div style={{ marginBottom: '1.25rem', textAlign: 'center' }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '0.65rem' }}>
          <ShazeLogo size={42} showWordmark={false} variant="dark" />
        </div>
        <h2 style={{ fontSize: '1.45rem', fontWeight: 800, color: 'var(--shaze-ink)' }}>Create an Account</h2>
        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
          Join Shaze to explore curated collections, place orders, and track deliveries.
        </p>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Full Name</label>
          <input
            className="form-input"
            type="text"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Jane Doe"
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
            placeholder="jane@example.com"
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
            placeholder="••••••••"
          />
        </div>

        <button className="btn btn-primary btn-block" type="submit" disabled={loading} style={{ marginTop: '0.5rem', width: '100%' }}>
          {loading ? 'Creating Account...' : 'Create Account'}
        </button>
      </form>

      <div style={{ marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border)', textAlign: 'center', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
        Already have an account?{' '}
        <span
          style={{ color: 'var(--shaze-tomato)', cursor: 'pointer', fontWeight: 600 }}
          onClick={onNavigateLogin}
        >
          Sign In
        </span>
      </div>
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
      const data = await api.login(email, password);
      setToken(data.token);
      setUser(data.user);
      onSuccess(data.user);
    } catch (err) {
      onError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card form-container">
      <div style={{ marginBottom: '1.25rem', textAlign: 'center' }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '0.65rem' }}>
          <ShazeLogo size={42} showWordmark={false} variant="dark" />
        </div>
        <h2 style={{ fontSize: '1.45rem', fontWeight: 800, color: 'var(--shaze-ink)' }}>Sign In to Shaze</h2>
        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
          Enter your email and password to access your account.
        </p>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Email Address</label>
          <input
            className="form-input"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="user@example.com"
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

        <button className="btn btn-primary btn-block" type="submit" disabled={loading} style={{ marginTop: '0.5rem', width: '100%' }}>
          {loading ? 'Authenticating...' : 'Sign In'}
        </button>
      </form>

      <div style={{ marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border)', textAlign: 'center', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
        Don't have an account?{' '}
        <span
          style={{ color: 'var(--shaze-tomato)', cursor: 'pointer', fontWeight: 600 }}
          onClick={onNavigateRegister}
        >
          Create an Account
        </span>
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
      onSuccess('Item removed from cart.');
    } catch (err) {
      onError(err.message);
    }
  };

  const handleClear = async () => {
    try {
      await api.clearCart();
      fetchCart();
      onCartUpdated();
      onSuccess('Cart has been cleared.');
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
      <div className="empty-state card" style={{ maxWidth: '480px', margin: '2rem auto' }}>
        <div style={{ fontSize: '2.5rem', marginBottom: '0.75rem' }}>🛒</div>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 700 }}>Your Cart is Empty</h2>
        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.4rem', marginBottom: '1.25rem' }}>
          Looks like you haven't added any products to your shopping cart yet.
        </p>
        <button className="btn btn-primary" onClick={onBrowse}>
          Browse Product Catalog
        </button>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 800 }}>Shopping Cart</h1>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            {items.length} unique {items.length === 1 ? 'item' : 'items'} in your cart
          </p>
        </div>
        <button className="btn btn-danger btn-sm" onClick={handleClear}>
          Clear Entire Cart
        </button>
      </div>

      <div className="cart-grid">
        {/* Left Column: Items List */}
        <div>
          {items.map((it) => (
            <div key={it.product_id} className="cart-item-card">
              <div className="cart-item-thumb">
                <ProductImage product={{ id: it.product_id, name: it.product_name }} />
              </div>
              <div className="cart-item-info">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                    {it.product_name || `Product #${it.product_id}`}
                  </h3>
                  <button
                    className="btn btn-danger btn-sm"
                    style={{ padding: '0.2rem 0.5rem' }}
                    onClick={() => handleRemove(it.product_id)}
                  >
                    Delete
                  </button>
                </div>

                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', margin: '0.35rem 0' }}>
                  Unit Price: ${Number(it.price).toFixed(2)}
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.75rem' }}>
                  <div className="quantity-stepper" style={{ width: '95px' }}>
                    <button
                      className="quantity-stepper-btn"
                      style={{ width: '28px', height: '30px' }}
                      onClick={() => handleUpdateQty(it.product_id, -1, it.quantity)}
                    >
                      −
                    </button>
                    <span style={{ flex: 1, textAlign: 'center', fontWeight: 600, fontSize: '0.85rem' }}>
                      {it.quantity}
                    </span>
                    <button
                      className="quantity-stepper-btn"
                      style={{ width: '28px', height: '30px' }}
                      onClick={() => handleUpdateQty(it.product_id, 1, it.quantity)}
                    >
                      +
                    </button>
                  </div>

                  <span style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--shaze-ink)', fontFamily: 'var(--font-mono)' }}>
                    ${(it.price * it.quantity).toFixed(2)}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Right Column: Order Summary */}
        <div className="order-summary-box">
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '1.25rem' }}>Order Summary</h2>
          <div className="summary-row">
            <span>Subtotal ({items.reduce((acc, it) => acc + it.quantity, 0)} items)</span>
            <span>${Number(cart.total_price || 0).toFixed(2)}</span>
          </div>
          <div className="summary-row">
            <span>Standard Express Shipping</span>
            <span style={{ color: 'var(--success-text)', fontWeight: 600 }}>FREE</span>
          </div>
          <div className="summary-row">
            <span>Estimated Sales Tax</span>
            <span>$0.00</span>
          </div>
          <div className="summary-row total">
            <span>Estimated Total</span>
            <span>${Number(cart.total_price || 0).toFixed(2)}</span>
          </div>

          <button
            className="btn btn-primary btn-block"
            style={{ marginTop: '1.5rem', padding: '0.75rem', fontSize: '0.95rem' }}
            onClick={onCheckout}
          >
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
  const [formData, setFormData] = useState({
    recipient_name: 'John Doe',
    phone_number: '+1 (555) 234-5678',
    address_line1: '100 Distributed Systems Way',
    address_line2: 'Suite 400',
    city: 'San Francisco',
    state: 'CA',
    postal_code: '94105',
    country: 'United States',
  });
  const [errors, setErrors] = useState({});
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

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validateForm = () => {
    const errs = {};
    if (!formData.recipient_name.trim()) {
      errs.recipient_name = 'Recipient name is required';
    } else if (formData.recipient_name.length > 255) {
      errs.recipient_name = 'Max 255 characters';
    }

    if (!formData.phone_number.trim()) {
      errs.phone_number = 'Phone number is required';
    } else if (formData.phone_number.length > 50) {
      errs.phone_number = 'Max 50 characters';
    }

    if (!formData.address_line1.trim()) {
      errs.address_line1 = 'Address line 1 is required';
    } else if (formData.address_line1.length > 255) {
      errs.address_line1 = 'Max 255 characters';
    }

    if (formData.address_line2 && formData.address_line2.length > 255) {
      errs.address_line2 = 'Max 255 characters';
    }

    if (!formData.city.trim()) {
      errs.city = 'City is required';
    } else if (formData.city.length > 100) {
      errs.city = 'Max 100 characters';
    }

    if (!formData.state.trim()) {
      errs.state = 'State / Province is required';
    } else if (formData.state.length > 100) {
      errs.state = 'Max 100 characters';
    }

    if (!formData.postal_code.trim()) {
      errs.postal_code = 'Postal code is required';
    } else if (formData.postal_code.length > 50) {
      errs.postal_code = 'Max 50 characters';
    }

    if (!formData.country.trim()) {
      errs.country = 'Country is required';
    } else if (formData.country.length > 100) {
      errs.country = 'Max 100 characters';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleCreateOrder = async (e) => {
    e.preventDefault();
    if (!validateForm()) {
      onError('Please fill in all required shipping address fields.');
      return;
    }

    setSubmitting(true);
    try {
      const orderItems = cart.items.map((it) => ({
        product_id: it.product_id,
        quantity: it.quantity,
        price: it.price,
      }));

      // 1. Create order atomically via Gateway -> Order Service with immutable snapshot
      const order = await api.createOrder(orderItems, {
        recipient_name: formData.recipient_name.trim(),
        phone_number: formData.phone_number.trim(),
        address_line1: formData.address_line1.trim(),
        address_line2: formData.address_line2.trim(),
        city: formData.city.trim(),
        state: formData.state.trim(),
        postal_code: formData.postal_code.trim(),
        country: formData.country.trim(),
      });

      // 2. Clear shopping cart
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
    <div style={{ maxWidth: '960px', margin: '0 auto' }}>
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: '1.25rem' }} onClick={onCancel}>
        ← Return to Cart
      </button>

      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1.4fr) minmax(320px, 1fr)', gap: '1.75rem', alignItems: 'start' }}>
        {/* Shipping Address Form */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '1.25rem' }}>
            <span style={{ fontSize: '1.25rem' }}>📦</span>
            <div>
              <h2 style={{ fontSize: '1.2rem', fontWeight: 700, margin: 0 }}>Shipping Destination</h2>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', margin: 0 }}>
                Enter the recipient address for this delivery
              </p>
            </div>
          </div>

          <form onSubmit={handleCreateOrder}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div className="form-group">
                <label className="form-label">
                  Recipient Name <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  type="text"
                  name="recipient_name"
                  className="form-input"
                  required
                  maxLength={255}
                  value={formData.recipient_name}
                  onChange={handleChange}
                  placeholder="e.g. John Doe"
                  style={errors.recipient_name ? { borderColor: '#ef4444' } : {}}
                />
                {errors.recipient_name && (
                  <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.recipient_name}</span>
                )}
              </div>

              <div className="form-group">
                <label className="form-label">
                  Contact Phone <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  type="tel"
                  name="phone_number"
                  className="form-input"
                  required
                  maxLength={50}
                  value={formData.phone_number}
                  onChange={handleChange}
                  placeholder="e.g. +1 (555) 000-0000"
                  style={errors.phone_number ? { borderColor: '#ef4444' } : {}}
                />
                {errors.phone_number && (
                  <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.phone_number}</span>
                )}
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">
                Address Line 1 <span style={{ color: '#ef4444' }}>*</span>
              </label>
              <input
                type="text"
                name="address_line1"
                className="form-input"
                required
                maxLength={255}
                value={formData.address_line1}
                onChange={handleChange}
                placeholder="Street address or P.O. Box"
                style={errors.address_line1 ? { borderColor: '#ef4444' } : {}}
              />
              {errors.address_line1 && (
                <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.address_line1}</span>
              )}
            </div>

            <div className="form-group">
              <label className="form-label">Address Line 2 (Optional)</label>
              <input
                type="text"
                name="address_line2"
                className="form-input"
                maxLength={255}
                value={formData.address_line2}
                onChange={handleChange}
                placeholder="Apt, Suite, Unit, Building, Floor, etc."
                style={errors.address_line2 ? { borderColor: '#ef4444' } : {}}
              />
              {errors.address_line2 && (
                <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.address_line2}</span>
              )}
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr 1fr', gap: '0.85rem' }}>
              <div className="form-group">
                <label className="form-label">
                  City <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  type="text"
                  name="city"
                  className="form-input"
                  required
                  maxLength={100}
                  value={formData.city}
                  onChange={handleChange}
                  placeholder="City"
                  style={errors.city ? { borderColor: '#ef4444' } : {}}
                />
                {errors.city && (
                  <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.city}</span>
                )}
              </div>

              <div className="form-group">
                <label className="form-label">
                  State / Province <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  type="text"
                  name="state"
                  className="form-input"
                  required
                  maxLength={100}
                  value={formData.state}
                  onChange={handleChange}
                  placeholder="State"
                  style={errors.state ? { borderColor: '#ef4444' } : {}}
                />
                {errors.state && (
                  <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.state}</span>
                )}
              </div>

              <div className="form-group">
                <label className="form-label">
                  Postal Code <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  type="text"
                  name="postal_code"
                  className="form-input"
                  required
                  maxLength={50}
                  value={formData.postal_code}
                  onChange={handleChange}
                  placeholder="ZIP / Postal"
                  style={errors.postal_code ? { borderColor: '#ef4444' } : {}}
                />
                {errors.postal_code && (
                  <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.postal_code}</span>
                )}
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">
                Country <span style={{ color: '#ef4444' }}>*</span>
              </label>
              <input
                type="text"
                name="country"
                className="form-input"
                required
                maxLength={100}
                value={formData.country}
                onChange={handleChange}
                placeholder="Country (e.g. United States)"
                style={errors.country ? { borderColor: '#ef4444' } : {}}
              />
              {errors.country && (
                <span style={{ fontSize: '0.75rem', color: '#ef4444' }}>{errors.country}</span>
              )}
            </div>

            <div style={{ background: 'var(--bg-surface-raised)', border: '1px solid var(--border)', borderRadius: 'var(--radius-md)', padding: '0.85rem', marginBottom: '1.25rem', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              🔒 Simulated payment pipeline: Order creation atomically captures your immutable shipping address snapshot and dispatches an <code>OrderCreated</code> AMQP event to RabbitMQ for settlement.
            </div>

            <div style={{ display: 'flex', gap: '0.75rem' }}>
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
                {submitting ? 'Placing Order...' : 'Confirm & Place Order'}
              </button>
            </div>
          </form>
        </div>

        {/* Order Breakdown Panel */}
        <div className="card">
          <h2 style={{ fontSize: '1.15rem', fontWeight: 700, marginBottom: '1rem' }}>Order Items ({cart.items.length})</h2>
          <div style={{ marginBottom: '1.25rem' }}>
            {cart.items.map((it) => (
              <div
                key={it.product_id}
                style={{ display: 'flex', justifyContent: 'space-between', padding: '0.55rem 0', borderBottom: '1px solid var(--border-subtle)', fontSize: '0.85rem' }}
              >
                <span>
                  {it.product_name || `Product #${it.product_id}`} <span style={{ color: 'var(--text-muted)' }}>× {it.quantity}</span>
                </span>
                <span style={{ fontWeight: 600 }}>${(it.price * it.quantity).toFixed(2)}</span>
              </div>
            ))}
          </div>

          <div className="summary-row total" style={{ marginTop: '0.5rem', paddingTop: '0.85rem' }}>
            <span>Total to Pay</span>
            <span>${Number(cart.total_price).toFixed(2)}</span>
          </div>
        </div>
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
        <p style={{ marginTop: '1rem', fontSize: '0.85rem' }}>Loading order history...</p>
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="empty-state card" style={{ maxWidth: '440px', margin: '2rem auto' }}>
        <h2 style={{ fontSize: '1.15rem', fontWeight: 600 }}>No Order History</h2>
        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.4rem', marginBottom: '1.25rem' }}>
          You have not placed any orders under this user session yet.
        </p>
        <button className="btn btn-primary btn-sm" onClick={onBrowse}>
          Browse Catalog
        </button>
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: '1.5rem' }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 800 }}>Your Orders</h1>
        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
          {orders.length} confirmed {orders.length === 1 ? 'order' : 'orders'} placed
        </p>
      </div>

      <div className="card table-container" style={{ padding: 0 }}>
        <table>
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Placed Timestamp</th>
              <th>Items</th>
              <th>Total Amount</th>
              <th>Status</th>
              <th style={{ textAlign: 'right' }}>Details</th>
            </tr>
          </thead>
          <tbody>
            {orders.map((o) => (
              <tr key={o.id}>
                <td style={{ fontWeight: 600 }}>Order #{o.id}</td>
                <td style={{ color: 'var(--text-secondary)' }}>
                  {new Date(o.created_at).toLocaleString()}
                </td>
                <td>{o.items?.length || 1} {o.items?.length === 1 ? 'item' : 'items'}</td>
                <td style={{ fontWeight: 700, color: 'var(--shaze-ink)', fontFamily: 'var(--font-mono)' }}>${Number(o.total_amount).toFixed(2)}</td>
                <td>
                  <span className={`badge ${o.status === 'PENDING' ? 'badge-pending' : 'badge-success'}`}>
                    {o.status}
                  </span>
                </td>
                <td style={{ textAlign: 'right' }}>
                  <button className="btn btn-secondary btn-sm" onClick={() => onViewOrder(o.id)}>
                    View Order →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
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

      try {
        const pay = await api.getPayment(orderId);
        setPayment(pay);
      } catch {
        // Payment might still be settling in RabbitMQ queue
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
      onError('Payment event is still processing via RabbitMQ...');
    } finally {
      setCheckingPayment(false);
    }
  };

  useEffect(() => {
    fetchOrderAndPayment();
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
      <div className="empty-state card">
        <p>Order not found.</p>
        <button className="btn btn-secondary btn-sm" style={{ marginTop: '1rem' }} onClick={onBack}>
          Back to Orders
        </button>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '780px', margin: '0 auto' }}>
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: '1.25rem' }} onClick={onBack}>
        ← Back to Orders
      </button>

      <div className="card">
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border)' }}>
          <div>
            <h1 style={{ fontSize: '1.35rem', fontWeight: 800 }}>Order #{order.id}</h1>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              Placed: {new Date(order.created_at).toLocaleString()}
            </p>
          </div>
          <span className={`badge ${order.status === 'PENDING' ? 'badge-pending' : 'badge-success'}`}>
            Order Status: {order.status}
          </span>
        </div>

        {/* Shipping Address Details Card */}
        <div style={{ background: 'var(--bg-surface-raised)', border: '1px solid var(--border)', borderRadius: 'var(--radius-md)', padding: '1rem', marginBottom: '1.25rem' }}>
          <div style={{ fontWeight: 700, fontSize: '0.9rem', marginBottom: '0.6rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--shaze-ink)' }}>
            <span>📍 Shipping Details</span>
          </div>
          {order.address ? (
            <div style={{ fontSize: '0.85rem', lineHeight: '1.5', color: 'var(--text-secondary)' }}>
              <div style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.95rem' }}>{order.address.recipient_name}</div>
              <div>{order.address.address_line1}{order.address.address_line2 ? `, ${order.address.address_line2}` : ''}</div>
              <div>{order.address.city}, {order.address.state} {order.address.postal_code}</div>
              <div>{order.address.country}</div>
              <div style={{ marginTop: '0.35rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                📞 Contact: {order.address.phone_number}
              </div>
            </div>
          ) : order.shipping_address ? (
            <div style={{ fontSize: '0.85rem', color: 'var(--text-primary)' }}>
              {order.shipping_address}
            </div>
          ) : (
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
              No shipping address recorded for this order.
            </div>
          )}
        </div>

        {/* RabbitMQ Async Payment Status Pipeline */}
        <div className="pipeline-card">
          <div className="pipeline-header">
            <div>
              <div style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--shaze-cream)' }}>RabbitMQ Payment Event Pipeline</div>
              <div style={{ fontSize: '0.75rem', color: 'rgba(251, 245, 234, 0.75)' }}>
                Event <code>OrderCreated</code> consumed asynchronously by Payment Service
              </div>
            </div>
            {payment ? (
              <span className="badge badge-success">PAYMENT {payment.status}</span>
            ) : (
              <span className="badge badge-pending">PROCESSING...</span>
            )}
          </div>

          {payment ? (
            <div className="pipeline-meta">
              <div>Transaction ID: <code style={{ color: 'var(--shaze-saffron)' }}>{payment.transaction_id}</code></div>
              <div>Settlement Method: {payment.payment_method}</div>
              <div>Settled Timestamp: {new Date(payment.created_at).toLocaleTimeString()}</div>
              <div>Amount Paid: ${Number(payment.amount).toFixed(2)}</div>
            </div>
          ) : (
            <div style={{ marginTop: '0.5rem' }}>
              <button
                className="btn btn-secondary btn-sm"
                onClick={recheckPayment}
                disabled={checkingPayment}
              >
                {checkingPayment ? 'Polling...' : 'Re-check Payment Status'}
              </button>
            </div>
          )}
        </div>

        {/* Ordered Line Items */}
        <h2 style={{ fontSize: '1.05rem', fontWeight: 700, marginTop: '1.5rem', marginBottom: '0.75rem' }}>
          Line Items
        </h2>
        <div className="table-container" style={{ border: '1px solid var(--border)', borderRadius: 'var(--radius-md)' }}>
          <table>
            <thead>
              <tr>
                <th>Product Description</th>
                <th style={{ textAlign: 'center' }}>Qty</th>
                <th>Unit Price</th>
                <th style={{ textAlign: 'right' }}>Total</th>
              </tr>
            </thead>
            <tbody>
              {(order.items || []).map((it) => (
                <tr key={it.id}>
                  <td style={{ fontWeight: 600 }}>{it.product_name || `Product #${it.product_id}`}</td>
                  <td style={{ textAlign: 'center' }}>{it.quantity}</td>
                  <td>${Number(it.price).toFixed(2)}</td>
                  <td style={{ textAlign: 'right', fontWeight: 700, color: 'var(--shaze-ink)', fontFamily: 'var(--font-mono)' }}>
                    ${(it.price * it.quantity).toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'baseline', gap: '0.75rem', marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid var(--border)' }}>
          <span style={{ fontSize: '0.9rem', color: 'var(--text-muted)' }}>Order Total:</span>
          <span style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--shaze-ink)', fontFamily: 'var(--font-mono)' }}>
            ${Number(order.total_amount).toFixed(2)}
          </span>
        </div>
      </div>
    </div>
  );
}
