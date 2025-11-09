import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import LoadingSkeleton from '../LoadingSkeleton';

describe('LoadingSkeleton', () => {
  it('renders skeleton structure', () => {
    const { container } = render(<LoadingSkeleton />);
    
    // Check for main container
    expect(container.querySelector('.min-h-screen')).toBeInTheDocument();
    
    // Check for gradient background
    expect(container.querySelector('.bg-gradient-to-br')).toBeInTheDocument();
  });

  it('renders three stat cards', () => {
    const { container } = render(<LoadingSkeleton />);
    
    // Stats grid should have 3 cards
    const statsGrid = container.querySelector('.grid.grid-cols-1.sm\\:grid-cols-3');
    expect(statsGrid).toBeInTheDocument();
    expect(statsGrid?.children).toHaveLength(3);
  });

  it('renders three column skeletons', () => {
    const { container } = render(<LoadingSkeleton />);
    
    // Board grid should have 3 columns
    const boardGrid = container.querySelectorAll('.grid.grid-cols-1.md\\:grid-cols-2.lg\\:grid-cols-3 > div');
    expect(boardGrid).toHaveLength(3);
  });

  it('has animate-pulse class for animation', () => {
    const { container } = render(<LoadingSkeleton />);
    
    const animatedElements = container.querySelectorAll('.animate-pulse');
    expect(animatedElements.length).toBeGreaterThan(0);
  });

  it('renders card skeletons in each column', () => {
    const { container } = render(<LoadingSkeleton />);
    
    // Each column should have 3 card skeletons
    const columns = container.querySelectorAll('.grid.grid-cols-1.md\\:grid-cols-2.lg\\:grid-cols-3 > div');
    columns.forEach((column) => {
      const cards = column.querySelectorAll('.space-y-3 > div');
      expect(cards).toHaveLength(3);
    });
  });
});
