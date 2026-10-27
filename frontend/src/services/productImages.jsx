import React from 'react';

// Product Image Resolution & Fallback Engine
// Maps actual product data to authentic high-resolution local assets with CDN and SVG fallbacks

const PRODUCT_MAP = {
  1: {
    local: '/images/products/probook-laptop.jpg',
    cdn: 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=800&auto=format&fit=crop&q=80',
    alt: 'ProBook Laptop 15" High-Performance Computing',
    rating: 4.8,
    reviews: 142,
    badge: 'Best Seller',
    originalPrice: 1499.99,
  },
  2: {
    local: '/images/products/ultraphone-14-pro.jpg',
    cdn: 'https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=800&auto=format&fit=crop&q=80',
    alt: 'UltraPhone 14 Pro Flagship Smartphone OLED',
    rating: 4.9,
    reviews: 328,
    badge: 'Popular',
    originalPrice: 1099.00,
  },
  3: {
    local: '/images/products/noisecancelling-headphones.jpg',
    cdn: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&auto=format&fit=crop&q=80',
    alt: 'NoiseCancelling Wireless Over-Ear Headphones',
    rating: 4.7,
    reviews: 95,
    badge: 'Top Rated',
    originalPrice: 299.00,
  },
  4: {
    local: '/images/products/mechanical-keyboard.jpg',
    cdn: 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&auto=format&fit=crop&q=80',
    alt: 'Mechanical RGB Backlit Gaming Keyboard',
    rating: 4.6,
    reviews: 210,
    badge: null,
    originalPrice: 119.99,
  },
  5: {
    local: '/images/products/ergonomic-mouse.jpg',
    cdn: 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&auto=format&fit=crop&q=80',
    alt: 'Wireless Precision Ergonomic Multi-Device Mouse',
    rating: 4.8,
    reviews: 184,
    badge: 'Amazon Choice',
    originalPrice: 69.99,
  },
  6: {
    local: '/images/products/4k-monitor.jpg',
    cdn: 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&auto=format&fit=crop&q=80',
    alt: '4K UltraHD 27-inch 144Hz IPS Gaming Monitor',
    rating: 4.9,
    reviews: 76,
    badge: 'Hot Deal',
    originalPrice: 479.99,
  },
  7: {
    local: '/images/products/smartwatch-pro.jpg',
    cdn: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800&auto=format&fit=crop&q=80',
    alt: 'SmartWatch Pro V2 Fitness & ECG Health Tracker',
    rating: 4.7,
    reviews: 119,
    badge: null,
    originalPrice: 249.00,
  },
  8: {
    local: '/images/products/octatab-tablet.jpg',
    cdn: 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?w=800&auto=format&fit=crop&q=80',
    alt: 'OctaTab 11 Tablet with Active Stylus Support',
    rating: 4.5,
    reviews: 64,
    badge: null,
    originalPrice: 549.00,
  },
  9: {
    local: '/images/products/mirrorless-camera.jpg',
    cdn: 'https://images.unsplash.com/photo-1516035069371-29a1b244cc32?w=800&auto=format&fit=crop&q=80',
    alt: 'Mirrorless 4K Digital Camera 24MP Sensor',
    rating: 4.9,
    reviews: 51,
    badge: 'Pro Gear',
    originalPrice: 999.00,
  },
  10: {
    local: '/images/products/gaming-console.jpg',
    cdn: 'https://images.unsplash.com/photo-1606813907291-d86efa9b94db?w=800&auto=format&fit=crop&q=80',
    alt: 'NextGen 1TB Ultra-Fast SSD Gaming Console',
    rating: 4.9,
    reviews: 412,
    badge: 'Trending',
    originalPrice: 549.99,
  },
  11: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1541807084-5c52b6b3adef?w=800&auto=format&fit=crop&q=80',
    alt: 'SlimBook Air 13" Ultra-Lightweight Laptop',
    rating: 4.8,
    reviews: 89,
    badge: 'Ultralight',
    originalPrice: 1249.00,
  },
  12: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1593640408182-31c70c8268f5?w=800&auto=format&fit=crop&q=80',
    alt: 'DeskStation Mini PC 32GB RAM Compact Desktop',
    rating: 4.7,
    reviews: 47,
    badge: null,
    originalPrice: 849.00,
  },
  13: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&auto=format&fit=crop&q=80',
    alt: 'True Wireless Pro Earbuds with Spatial Audio',
    rating: 4.8,
    reviews: 204,
    badge: 'Popular',
    originalPrice: 199.99,
  },
  14: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1545454675-3531b543be5d?w=800&auto=format&fit=crop&q=80',
    alt: 'Hi-Fi Studio Monitor Powered Desktop Speakers',
    rating: 4.9,
    reviews: 62,
    badge: 'Studio Grade',
    originalPrice: 349.00,
  },
  15: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?w=800&auto=format&fit=crop&q=80',
    alt: 'Rugged Waterproof Bluetooth Speaker',
    rating: 4.6,
    reviews: 175,
    badge: 'Waterproof',
    originalPrice: 99.99,
  },
  16: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1583863788434-e58a36330cf0?w=800&auto=format&fit=crop&q=80',
    alt: '65W GaN USB-C 3-Port Wall Fast Charger',
    rating: 4.9,
    reviews: 310,
    badge: 'Fast Charge',
    originalPrice: 49.99,
  },
  17: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c?w=800&auto=format&fit=crop&q=80',
    alt: 'Braided Thunderbolt 4 Cable 40Gbps 100W',
    rating: 4.8,
    reviews: 145,
    badge: null,
    originalPrice: 39.99,
  },
  18: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1629429408209-1f912961dbd8?w=800&auto=format&fit=crop&q=80',
    alt: 'Ergonomic Aluminum Adjustable Laptop Stand',
    rating: 4.7,
    reviews: 112,
    badge: null,
    originalPrice: 59.99,
  },
  19: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1575311373937-040b8e1fd5b6?w=800&auto=format&fit=crop&q=80',
    alt: 'Active Fitness Tracker Band with SpO2 Sensor',
    rating: 4.6,
    reviews: 98,
    badge: 'Fitness Essential',
    originalPrice: 79.99,
  },
  20: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1508685096489-7aacd43bd3b1?w=800&auto=format&fit=crop&q=80',
    alt: 'GPS Adventure Rugged Outdoor Smartwatch',
    rating: 4.9,
    reviews: 83,
    badge: 'Adventure',
    originalPrice: 399.00,
  },
  21: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1533158307587-828f0a76ef46?w=800&auto=format&fit=crop&q=80',
    alt: 'Professional 65" Carbon Fiber Tripod with Fluid Head',
    rating: 4.8,
    reviews: 67,
    badge: 'Pro Photo',
    originalPrice: 179.99,
  },
  22: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1617005082133-548c4dd27f35?w=800&auto=format&fit=crop&q=80',
    alt: '50mm f/1.8 Prime Portrait Lens Bokeh',
    rating: 4.9,
    reviews: 140,
    badge: 'Sharp Optics',
    originalPrice: 269.00,
  },
  23: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1527799820374-dcf8d9d4a388?w=800&auto=format&fit=crop&q=80',
    alt: 'Studio 18" Bi-Color Dimmable LED Ring Light',
    rating: 4.7,
    reviews: 88,
    badge: null,
    originalPrice: 89.99,
  },
  24: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1600080972464-8e5f35f63d08?w=800&auto=format&fit=crop&q=80',
    alt: 'Wireless Pro Game Controller Hall-Effect Triggers',
    rating: 4.8,
    reviews: 195,
    badge: 'Pro Gamer',
    originalPrice: 84.99,
  },
  25: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1599669454699-248893623440?w=800&auto=format&fit=crop&q=80',
    alt: '7.1 Surround Gaming Headset with Boom Mic',
    rating: 4.7,
    reviews: 133,
    badge: 'Surround 7.1',
    originalPrice: 119.99,
  },
  26: {
    local: null,
    cdn: 'https://images.unsplash.com/photo-1616440347437-b1c73416efc2?w=800&auto=format&fit=crop&q=80',
    alt: 'RGB Extended Micro-Textured Gaming Desk Mat',
    rating: 4.8,
    reviews: 240,
    badge: 'Best Value',
    originalPrice: 34.99,
  },
};

// Category Vector Fallback SVGs (Data URIs for guaranteed rendering with zero network dependency)
const SVG_FALLBACKS = {
  Electronics: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='80' y='60' width='240' height='150' rx='8' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Crect x='100' y='80' width='200' height='110' rx='4' fill='%230f172a'/%3E%3Cpath d='M60 220h280a8 8 0 0 1 8 8v4H52v-4a8 8 0 0 1 8-8z' fill='%23334155'/%3E%3Ctext x='200' y='145' font-family='sans-serif' font-size='14' fill='%2394a3b8' text-anchor='middle'%3EElectronics Device%3C/text%3E%3C/svg%3E",
  Audio: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Ccircle cx='200' cy='150' r='60' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Cpath d='M140 150a60 60 0 0 1 120 0' stroke='%2360a5fa' stroke-width='6' stroke-linecap='round'/%3E%3Crect x='130' y='140' width='20' height='40' rx='6' fill='%2338bdf8'/%3E%3Crect x='250' y='140' width='20' height='40' rx='6' fill='%2338bdf8'/%3E%3Ctext x='200' y='230' font-family='sans-serif' font-size='14' fill='%2394a3b8' text-anchor='middle'%3EAudio Gear%3C/text%3E%3C/svg%3E",
  Accessories: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='110' y='90' width='180' height='120' rx='10' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Ccircle cx='200' cy='150' r='25' fill='%23334155'/%3E%3Ctext x='200' y='240' font-family='sans-serif' font-size='14' fill='%2394a3b8' text-anchor='middle'%3EAccessory Device%3C/text%3E%3C/svg%3E",
  Wearables: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='170' y='40' width='60' height='220' rx='10' fill='%231e2436' stroke='%23334155' stroke-width='2'/%3E%3Crect x='150' y='100' width='100' height='100' rx='20' stroke='%233b82f6' stroke-width='4' fill='%230f172a'/%3E%3Ctext x='200' y='155' font-family='sans-serif' font-size='14' fill='%2338bdf8' text-anchor='middle'%3E10:09%3C/text%3E%3C/svg%3E",
  Photography: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='100' y='100' width='200' height='130' rx='14' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Ccircle cx='200' cy='165' r='40' stroke='%2360a5fa' stroke-width='5' fill='%230f172a'/%3E%3Ccircle cx='200' cy='165' r='20' fill='%2338bdf8'/%3E%3Crect x='130' y='80' width='40' height='20' rx='4' fill='%23334155'/%3E%3C/svg%3E",
  Gaming: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='90' y='100' width='220' height='120' rx='30' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Ccircle cx='140' cy='150' r='12' fill='%2338bdf8'/%3E%3Ccircle cx='260' cy='150' r='8' fill='%23f43f5e'/%3E%3Ccircle cx='240' cy='135' r='8' fill='%2310b981'/%3E%3Ccircle cx='280' cy='135' r='8' fill='%23f59e0b'/%3E%3Ccircle cx='260' cy='120' r='8' fill='%233b82f6'/%3E%3C/svg%3E",
  Default: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300' fill='none'%3E%3Crect width='400' height='300' fill='%23131722'/%3E%3Crect x='120' y='80' width='160' height='140' rx='12' stroke='%233b82f6' stroke-width='4' fill='%231e2436'/%3E%3Ctext x='200' y='155' font-family='sans-serif' font-size='16' fill='%2394a3b8' text-anchor='middle'%3EProduct Item%3C/text%3E%3C/svg%3E",
};

/**
 * Resolves product image details with local, CDN, and vector fallbacks
 */
export function getProductMeta(product) {
  if (!product) {
    return {
      src: SVG_FALLBACKS.Default,
      cdn: SVG_FALLBACKS.Default,
      alt: 'Product',
      rating: 4.8,
      reviews: 120,
      badge: null,
      originalPrice: null,
      fallback: SVG_FALLBACKS.Default,
    };
  }

  // 1. Direct ID match from canonical catalog (IDs 1..26)
  if (PRODUCT_MAP[product.id]) {
    const meta = PRODUCT_MAP[product.id];
    return {
      src: meta.local || meta.cdn,
      cdn: meta.cdn,
      alt: meta.alt,
      rating: meta.rating,
      reviews: meta.reviews,
      badge: meta.badge,
      originalPrice: meta.originalPrice,
      fallback: SVG_FALLBACKS[product.category] || SVG_FALLBACKS.Default,
    };
  }

  // 2. Direct Name Match
  for (const id in PRODUCT_MAP) {
    if (PRODUCT_MAP[id].alt && product.name && PRODUCT_MAP[id].alt.toLowerCase().includes(product.name.toLowerCase())) {
      const meta = PRODUCT_MAP[id];
      return {
        src: meta.local || meta.cdn,
        cdn: meta.cdn,
        alt: product.name,
        rating: meta.rating,
        reviews: meta.reviews,
        badge: meta.badge,
        originalPrice: meta.originalPrice,
        fallback: SVG_FALLBACKS[product.category] || SVG_FALLBACKS.Default,
      };
    }
  }

  // 3. Keyword heuristic for custom/newly added products
  const name = (product.name || '').toLowerCase();
  const desc = (product.description || '').toLowerCase();

  let matched = null;
  if (name.includes('earbud') || name.includes('in-ear') || name.includes('airpod')) {
    matched = PRODUCT_MAP[13];
  } else if (name.includes('speaker') || name.includes('soundbar')) {
    matched = PRODUCT_MAP[15];
  } else if (name.includes('charger') || name.includes('adapter') || name.includes('gan')) {
    matched = PRODUCT_MAP[16];
  } else if (name.includes('cable') || name.includes('wire') || name.includes('cord')) {
    matched = PRODUCT_MAP[17];
  } else if (name.includes('stand') || name.includes('mount') || name.includes('holder')) {
    matched = PRODUCT_MAP[18];
  } else if (name.includes('fitness') || name.includes('band') || name.includes('tracker')) {
    matched = PRODUCT_MAP[19];
  } else if (name.includes('tripod') || name.includes('monopod')) {
    matched = PRODUCT_MAP[21];
  } else if (name.includes('lens') || name.includes('optics')) {
    matched = PRODUCT_MAP[22];
  } else if (name.includes('light') || name.includes('lamp') || name.includes('ring')) {
    matched = PRODUCT_MAP[23];
  } else if (name.includes('controller') || name.includes('gamepad') || name.includes('joystick')) {
    matched = PRODUCT_MAP[24];
  } else if (name.includes('headset') && (name.includes('game') || name.includes('gaming'))) {
    matched = PRODUCT_MAP[25];
  } else if (name.includes('pad') || name.includes('mat')) {
    matched = PRODUCT_MAP[26];
  } else if (name.includes('laptop') || name.includes('notebook') || name.includes('macbook')) {
    matched = PRODUCT_MAP[1];
  } else if (name.includes('phone') || name.includes('mobile') || name.includes('smartphone')) {
    matched = PRODUCT_MAP[2];
  } else if (name.includes('headphone') || name.includes('headset') || name.includes('audio')) {
    matched = PRODUCT_MAP[3];
  } else if (name.includes('keyboard')) {
    matched = PRODUCT_MAP[4];
  } else if (name.includes('mouse')) {
    matched = PRODUCT_MAP[5];
  } else if (name.includes('monitor') || name.includes('screen') || name.includes('display')) {
    matched = PRODUCT_MAP[6];
  } else if (name.includes('watch')) {
    matched = PRODUCT_MAP[7];
  } else if (name.includes('tablet') || name.includes('ipad')) {
    matched = PRODUCT_MAP[8];
  } else if (name.includes('camera')) {
    matched = PRODUCT_MAP[9];
  } else if (name.includes('console') || name.includes('gaming')) {
    matched = PRODUCT_MAP[10];
  }

  if (matched) {
    return {
      src: matched.local || matched.cdn,
      cdn: matched.cdn,
      alt: product.name,
      rating: 4.7,
      reviews: 88,
      badge: null,
      originalPrice: product.price ? Number((product.price * 1.15).toFixed(2)) : null,
      fallback: SVG_FALLBACKS[product.category] || SVG_FALLBACKS.Default,
    };
  }

  // 4. Fallback to category vector
  const fallbackSvg = SVG_FALLBACKS[product.category] || SVG_FALLBACKS.Default;
  return {
    src: product.image_url || fallbackSvg,
    cdn: fallbackSvg,
    alt: product.name || 'Product',
    rating: 4.8,
    reviews: 65,
    badge: null,
    originalPrice: product.price ? Number((product.price * 1.2).toFixed(2)) : null,
    fallback: fallbackSvg,
  };
}

/**
 * Image component with robust multi-tier fallback and loading state
 */
export function ProductImage({ product, className = '', style = {}, isHero = false }) {
  const meta = getProductMeta(product);

  const handleImageError = (e) => {
    // If local asset failed, attempt CDN; if CDN failed, fallback to SVG
    if (e.target.src.includes('/images/products/') && meta.cdn && !meta.cdn.startsWith('data:')) {
      e.target.src = meta.cdn;
    } else {
      e.target.src = meta.fallback;
    }
  };

  return (
    <img
      src={meta.src}
      alt={meta.alt}
      className={className}
      style={{
        objectFit: 'contain',
        width: '100%',
        height: '100%',
        display: 'block',
        ...style,
      }}
      onError={handleImageError}
      loading="lazy"
    />
  );
}
