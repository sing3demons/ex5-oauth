import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Header from '../Header';
import { User } from '../../types';

describe('Header Component', () => {
  const mockUser: User = {
    sub: '123',
    email: 'test@example.com',
    name: 'Test User'
  };

  const mockLogout = vi.fn();

  it('renders user profile information', () => {
    render(<Header user={mockUser} onLogout={mockLogout} />);
    
    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });

  it('displays user initials when no picture is provided', () => {
    render(<Header user={mockUser} onLogout={mockLogout} />);
    
    expect(screen.getByText('T')).toBeInTheDocument();
  });

  it('shows dropdown menu when profile button is clicked', () => {
    render(<Header user={mockUser} onLogout={mockLogout} />);
    
    const profileButton = screen.getByRole('button', { name: /user menu/i });
    fireEvent.click(profileButton);
    
    expect(screen.getAllByText('Test User').length).toBeGreaterThan(1);
    expect(screen.getByText('Logout')).toBeInTheDocument();
  });

  it('calls onLogout when logout button is clicked', () => {
    render(<Header user={mockUser} onLogout={mockLogout} />);
    
    const profileButton = screen.getByRole('button', { name: /user menu/i });
    fireEvent.click(profileButton);
    
    const logoutButton = screen.getByText('Logout');
    fireEvent.click(logoutButton);
    
    expect(mockLogout).toHaveBeenCalledTimes(1);
  });
});
