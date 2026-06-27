import AsyncStorage from '@react-native-async-storage/async-storage';
import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { API_BASE_URL } from '@/constants/api';

const AUTH_TOKEN_KEY = 'wwetd.auth.token';
const AUTH_USER_KEY = 'wwetd.auth.user';
const LOGIN_SKIP_UNTIL_KEY = 'wwetd.auth.skip_until';
const SKIP_DURATION_MS = 24 * 60 * 60 * 1000;

type AuthUser = {
  id: string;
  phone: string;
  created_at: string;
};

type AuthResponse = {
  data: {
    token: string;
    user: AuthUser;
  };
};

type OTPResponse = {
  data: {
    phone: string;
    ttl: number;
    debug_otp?: string;
  };
};

type AuthContextValue = {
  user: AuthUser | null;
  token: string | null;
  booting: boolean;
  shouldAskLogin: () => Promise<boolean>;
  skipLoginPrompt: () => Promise<void>;
  register: (phone: string, password: string) => Promise<void>;
  login: (phone: string, password: string) => Promise<void>;
  requestOTP: (phone: string) => Promise<OTPResponse['data']>;
  verifyOTP: (phone: string, otp: string) => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [booting, setBooting] = useState(true);

  useEffect(() => {
    async function restore() {
      try {
        const [storedToken, storedUser] = await Promise.all([
          AsyncStorage.getItem(AUTH_TOKEN_KEY),
          AsyncStorage.getItem(AUTH_USER_KEY),
        ]);

        if (storedToken && storedUser) {
          setToken(storedToken);
          setUser(JSON.parse(storedUser) as AuthUser);
        }
      } finally {
        setBooting(false);
      }
    }

    restore();
  }, []);

  const persistAuth = useCallback(async (payload: AuthResponse) => {
    setToken(payload.data.token);
    setUser(payload.data.user);
    await Promise.all([
      AsyncStorage.setItem(AUTH_TOKEN_KEY, payload.data.token),
      AsyncStorage.setItem(AUTH_USER_KEY, JSON.stringify(payload.data.user)),
      AsyncStorage.removeItem(LOGIN_SKIP_UNTIL_KEY),
    ]);
  }, []);

  const shouldAskLogin = useCallback(async () => {
    if (token) return false;

    const skipUntil = await AsyncStorage.getItem(LOGIN_SKIP_UNTIL_KEY);
    if (!skipUntil) return true;

    return Date.now() >= Number(skipUntil);
  }, [token]);

  const skipLoginPrompt = useCallback(async () => {
    await AsyncStorage.setItem(LOGIN_SKIP_UNTIL_KEY, String(Date.now() + SKIP_DURATION_MS));
  }, []);

  const post = useCallback(async <T,>(path: string, body: unknown): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });

    const payload = await response.json();
    if (!response.ok) {
      throw new Error(payload?.error?.message ?? 'Có lỗi xảy ra');
    }

    return payload as T;
  }, []);

  const register = useCallback(
    async (phone: string, password: string) => {
      const payload = await post<AuthResponse>('/api/v1/auth/register', { phone, password });
      await persistAuth(payload);
    },
    [persistAuth, post],
  );

  const login = useCallback(
    async (phone: string, password: string) => {
      const payload = await post<AuthResponse>('/api/v1/auth/login', { phone, password });
      await persistAuth(payload);
    },
    [persistAuth, post],
  );

  const requestOTP = useCallback(
    async (phone: string) => {
      const payload = await post<OTPResponse>('/api/v1/auth/otp/request', { phone });
      return payload.data;
    },
    [post],
  );

  const verifyOTP = useCallback(
    async (phone: string, otp: string) => {
      const payload = await post<AuthResponse>('/api/v1/auth/otp/verify', { phone, otp });
      await persistAuth(payload);
    },
    [persistAuth, post],
  );

  const value = useMemo(
    () => ({
      user,
      token,
      booting,
      shouldAskLogin,
      skipLoginPrompt,
      register,
      login,
      requestOTP,
      verifyOTP,
    }),
    [booting, login, register, requestOTP, shouldAskLogin, skipLoginPrompt, token, user, verifyOTP],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return value;
}
