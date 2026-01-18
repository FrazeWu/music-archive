import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { User } from '@/types';

const API_URL = import.meta.env.VITE_API_URL || '/api';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'));
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'));
  const isAuthenticated = ref(!!token.value);

  const loginWithPassword = async (email: string, password: string) => {
    try {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: email, password })
        });
        
        if (!response.ok) {
             const err = await response.json();
             throw new Error(err.error || 'Login failed');
        }

        const data = await response.json();
        token.value = data.token;
        user.value = data.user;
        isAuthenticated.value = true;
        
        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
    } catch (error) {
        throw error;
    }
  };

  const register = async (username: string, email: string, password: string) => {
    try {
        const response = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });

        if (!response.ok) {
             const err = await response.json();
             throw new Error(err.error || 'Registration failed');
        }

        const data = await response.json();
        token.value = data.token;
        user.value = data.user;
        isAuthenticated.value = true;

        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
    } catch (error) {
        throw error;
    }
  };

  const logout = () => {
    token.value = null;
    user.value = null;
    isAuthenticated.value = false;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  };

  return {
    token,
    user,
    isAuthenticated,
    loginWithPassword,
    register,
    logout
  };
});
