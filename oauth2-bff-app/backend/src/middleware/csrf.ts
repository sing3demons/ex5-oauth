import csrf from 'csurf';

// CSRF protection middleware (cookie-based)
export const csrfProtection = csrf({ cookie: true });
