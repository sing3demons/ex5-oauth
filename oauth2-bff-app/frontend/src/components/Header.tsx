import { useState, useRef, useEffect } from 'react';
import { User } from '../types';

interface HeaderProps {
  user: User;
  onLogout: () => void;
}

export default function Header({ user, onLogout }: HeaderProps) {
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };

    if (showDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showDropdown]);

  const getInitials = () => {
    if (user.name) {
      return user.name.charAt(0).toUpperCase();
    }
    if (user.email) {
      return user.email.charAt(0).toUpperCase();
    }
    return '?';
  };

  return (
    <header className="flex flex-col sm:flex-row justify-between items-center p-3 md:p-4 lg:p-6 bg-white/95 rounded-xl mb-6 shadow-md gap-3 sm:gap-0">
      <div className="flex items-center gap-2">
        <span className="text-2xl">📝</span>
        <span className="text-lg md:text-xl font-bold text-gray-800">Todo App</span>
      </div>

      <div className="relative" ref={dropdownRef}>
        <button
          onClick={() => setShowDropdown(!showDropdown)}
          className="flex items-center gap-3 p-2 md:p-2 bg-transparent border-2 border-transparent rounded-lg cursor-pointer transition-all hover:bg-indigo-50 hover:border-indigo-600 touch-manipulation min-h-[44px]"
          aria-label="User menu"
        >
          {user.picture ? (
            <img
              src={user.picture}
              alt={user.name || 'User'}
              className="w-10 h-10 rounded-full object-cover"
            />
          ) : (
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-lg font-bold text-white">
              {getInitials()}
            </div>
          )}
          <div className="hidden sm:flex flex-col items-start">
            <div className="text-sm font-bold text-gray-800 leading-tight">
              {user.name || 'User'}
            </div>
            <div className="text-xs text-gray-600 leading-tight">
              {user.email}
            </div>
          </div>
          <span className="text-[10px] text-gray-600 transition-transform">▼</span>
        </button>

        {showDropdown && (
          <div className="absolute top-[calc(100%+8px)] right-0 bg-white rounded-lg shadow-xl min-w-[240px] z-[1000] overflow-hidden">
            <div className="p-4 bg-indigo-50">
              <div className="text-base font-bold text-gray-800 mb-1">
                {user.name || 'User'}
              </div>
              <div className="text-sm text-gray-600">
                {user.email}
              </div>
            </div>
            <div className="h-px bg-gray-200" />
            <button 
              onClick={onLogout} 
              className="w-full flex items-center gap-3 p-3 md:p-3 bg-transparent border-none text-sm font-medium text-red-600 cursor-pointer transition-colors hover:bg-red-50 touch-manipulation min-h-[44px]"
            >
              <span className="text-base">🚪</span>
              Logout
            </button>
          </div>
        )}
      </div>
    </header>
  );
}


