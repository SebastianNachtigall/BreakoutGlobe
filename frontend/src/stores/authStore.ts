import { create } from 'zustand';

export interface UserProfile {
  id: string;
  email: string;
  displayName: string;
  accountType: 'guest' | 'full';
  role: 'user' | 'admin' | 'superadmin';
  avatarUrl?: string;
  aboutMe?: string;
  createdAt: string;
}

export interface AuthState {
  token: string | null;
  user: UserProfile | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  // Actions
  setToken: (token: string | null) => void;
  setUser: (user: UserProfile | null) => void;
  login: (email: string, password: string) => Promise<void>;
  signup: (signupData: SignupData) => Promise<void>;
  logout: () => void;
  loadAuthFromStorage: () => void;
  checkAuth: () => Promise<boolean>;
}

export interface SignupData {
  email: string;
  password: string;
  displayName: string;
  aboutMe?: string;
}

const AUTH_TOKEN_KEY = 'authToken';
const AUTH_USER_KEY = 'authUser';

export const authStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isAuthenticated: false,
  isLoading: false,

  setToken: (token) => {
    if (token) {
      localStorage.setItem(AUTH_TOKEN_KEY, token);
    } else {
      localStorage.removeItem(AUTH_TOKEN_KEY);
    }
    set({ token, isAuthenticated: !!token });
  },

  setUser: (user) => {
    if (user) {
      localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
    } else {
      localStorage.removeItem(AUTH_USER_KEY);
    }
    set({ user });
  },

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      
      const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Login failed');
      }

      const data = await response.json();
      
      get().setToken(data.token);
      get().setUser(data.user);
      
      console.log('✅ Login successful:', data.user.displayName);
    } catch (error) {
      console.error('❌ Login failed:', error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  signup: async (signupData) => {
    set({ isLoading: true });
    try {
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      
      const response = await fetch(`${API_BASE_URL}/api/auth/signup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(signupData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Signup failed');
      }

      const data = await response.json();
      
      get().setToken(data.token);
      get().setUser(data.user);
      
      console.log('✅ Signup successful:', data.user.displayName);
    } catch (error) {
      console.error('❌ Signup failed:', error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    get().setToken(null);
    get().setUser(null);
    console.log('✅ Logged out');
  },

  loadAuthFromStorage: () => {
    try {
      const token = localStorage.getItem(AUTH_TOKEN_KEY);
      const userStr = localStorage.getItem(AUTH_USER_KEY);
      
      if (token && userStr) {
        const user = JSON.parse(userStr);
        set({ 
          token, 
          user, 
          isAuthenticated: true 
        });
        console.log('✅ Auth loaded from storage:', user.displayName);
      }
    } catch (error) {
      console.error('❌ Failed to load auth from storage:', error);
      // Clear invalid data
      localStorage.removeItem(AUTH_TOKEN_KEY);
      localStorage.removeItem(AUTH_USER_KEY);
    }
  },

  checkAuth: async () => {
    const { token } = get();
    
    if (!token) {
      return false;
    }

    try {
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      
      const response = await fetch(`${API_BASE_URL}/api/auth/me`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        // Token is invalid or expired
        get().logout();
        return false;
      }

      const user = await response.json();
      get().setUser(user);
      return true;
    } catch (error) {
      console.error('❌ Auth check failed:', error);
      get().logout();
      return false;
    }
  },
}));
