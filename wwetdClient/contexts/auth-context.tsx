import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';

import { AuthFormModal } from '@/components/auth-form-modal';
import { NameFormModal } from '@/components/name-form-modal';
import { API_BASE_URL } from '@/constants/api';

const AUTH_TOKEN_KEY = 'wwetd.auth.token';
const AUTH_REFRESH_TOKEN_KEY = 'wwetd.auth.refresh_token';
const AUTH_USER_KEY = 'wwetd.auth.user';
const LOGIN_SKIP_UNTIL_KEY = 'wwetd.auth.skip_until';
const SKIP_DURATION_MS = 24 * 60 * 60 * 1000;

export type AuthUser = {
  id: string;
  name: string;
  avatar: string;
  phone: string;
  created_at: string;
  updated_at: string;
};

export type AvatarUpload = {
  uri: string;
  fileName?: string | null;
  mimeType?: string | null;
  file?: Blob | null;
};

type AuthResponse = {
  data: {
    access_token?: string;
    expires_in?: number;
    refresh_expires_in?: number;
    refresh_token?: string;
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
  refreshToken: string | null;
  booting: boolean;
  authFetch: (input: string, init?: RequestInit) => Promise<Response>;
  clearAuth: () => Promise<void>;
  requireLogin: () => void;
  shouldAskLogin: () => Promise<boolean>;
  skipLoginPrompt: () => Promise<void>;
  register: (phone: string, password: string) => Promise<void>;
  login: (phone: string, password: string) => Promise<void>;
  requestOTP: (phone: string) => Promise<OTPResponse['data']>;
  verifyOTP: (phone: string, otp: string) => Promise<void>;
  updateName: (name: string) => Promise<void>;
  refreshUser: () => Promise<void>;
  uploadAvatar: (avatar: AvatarUpload) => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [booting, setBooting] = useState(true);
  const [showLoginForm, setShowLoginForm] = useState(false);
  const refreshPromiseRef = useRef<Promise<string | null> | null>(null);

  useEffect(() => {
    async function restore() {
      try {
        const [storedToken, storedRefreshToken, storedUser] = await Promise.all([
          AsyncStorage.getItem(AUTH_TOKEN_KEY),
          AsyncStorage.getItem(AUTH_REFRESH_TOKEN_KEY),
          AsyncStorage.getItem(AUTH_USER_KEY),
        ]);

        if (storedToken && storedUser) {
          setToken(storedToken);
          setRefreshToken(storedRefreshToken);
          setUser(JSON.parse(storedUser) as AuthUser);
        }
      } finally {
        setBooting(false);
      }
    }

    restore();
  }, []);

  const persistAuth = useCallback(async (payload: AuthResponse) => {
    const nextToken = payload.data.access_token ?? payload.data.token;
    const nextRefreshToken = payload.data.refresh_token ?? null;

    setToken(nextToken);
    setRefreshToken(nextRefreshToken);
    setUser(payload.data.user);
    await Promise.all([
      AsyncStorage.setItem(AUTH_TOKEN_KEY, nextToken),
      AsyncStorage.setItem(AUTH_USER_KEY, JSON.stringify(payload.data.user)),
      AsyncStorage.removeItem(LOGIN_SKIP_UNTIL_KEY),
      nextRefreshToken
        ? AsyncStorage.setItem(AUTH_REFRESH_TOKEN_KEY, nextRefreshToken)
        : AsyncStorage.removeItem(AUTH_REFRESH_TOKEN_KEY),
    ]);
    setShowLoginForm(false);
  }, []);

  const clearAuth = useCallback(async () => {
    setToken(null);
    setRefreshToken(null);
    setUser(null);
    await Promise.all([
      AsyncStorage.removeItem(AUTH_TOKEN_KEY),
      AsyncStorage.removeItem(AUTH_REFRESH_TOKEN_KEY),
      AsyncStorage.removeItem(AUTH_USER_KEY),
    ]);
  }, []);

  const requireLogin = useCallback(() => {
    setShowLoginForm(true);
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

  const refreshAuth = useCallback(async () => {
    if (refreshPromiseRef.current) {
      return refreshPromiseRef.current;
    }

    refreshPromiseRef.current = (async () => {
      const storedRefreshToken =
        refreshToken ?? (await AsyncStorage.getItem(AUTH_REFRESH_TOKEN_KEY));
      if (!storedRefreshToken) {
        await clearAuth();
        return null;
      }

      const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: storedRefreshToken }),
      });

      if (!response.ok) {
        await clearAuth();
        return null;
      }

      const payload = (await response.json()) as AuthResponse;
      await persistAuth(payload);
      return payload.data.access_token ?? payload.data.token;
    })();

    try {
      return await refreshPromiseRef.current;
    } finally {
      refreshPromiseRef.current = null;
    }
  }, [clearAuth, persistAuth, refreshToken]);

  const authFetch = useCallback(
    async (input: string, init: RequestInit = {}) => {
      const requestURL = input.startsWith('http') ? input : `${API_BASE_URL}${input}`;
      const currentToken = token ?? (await AsyncStorage.getItem(AUTH_TOKEN_KEY));
      const headers = new Headers(init.headers);
      if (currentToken && !headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${currentToken}`);
      }

      let response = await fetch(requestURL, { ...init, headers });
      if (response.status !== 401) {
        return response;
      }

      const nextToken = await refreshAuth();
      if (!nextToken) {
        setShowLoginForm(true);
        return response;
      }

      const retryHeaders = new Headers(init.headers);
      retryHeaders.set('Authorization', `Bearer ${nextToken}`);
      response = await fetch(requestURL, { ...init, headers: retryHeaders });
      if (response.status === 401) {
        await clearAuth();
        setShowLoginForm(true);
      }
      return response;
    },
    [clearAuth, refreshAuth, token],
  );

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

  const updateName = useCallback(
    async (name: string) => {
      const response = await authFetch('/api/v1/auth/profile', {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name }),
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error?.message ?? 'Không lưu được tên');
      }

      const nextUser = payload.data as AuthUser;
      setUser(nextUser);
      await AsyncStorage.setItem(AUTH_USER_KEY, JSON.stringify(nextUser));
    },
    [authFetch],
  );

  const persistUser = useCallback(async (nextUser: AuthUser) => {
    setUser(nextUser);
    await AsyncStorage.setItem(AUTH_USER_KEY, JSON.stringify(nextUser));
  }, []);

  const refreshUser = useCallback(async () => {
    const response = await authFetch('/api/v1/auth/me');
    const payload = await response.json();
    if (!response.ok) {
      throw new Error(payload?.error?.message ?? 'Không tải được thông tin cá nhân');
    }
    await persistUser(payload.data as AuthUser);
  }, [authFetch, persistUser]);

  const uploadAvatar = useCallback(
    async (avatar: AvatarUpload) => {
      const form = new FormData();
      const fileName = avatar.fileName ?? `avatar-${Date.now()}.jpg`;
      const mimeType = avatar.mimeType ?? 'image/jpeg';

      if (avatar.file) {
        form.append('avatar', avatar.file, fileName);
      } else {
        form.append(
          'avatar',
          {
            uri: avatar.uri,
            name: fileName,
            type: mimeType,
          } as unknown as Blob,
        );
      }

      const response = await authFetch('/api/v1/auth/avatar', {
        method: 'POST',
        body: form,
      });
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload?.error?.message ?? 'Không tải được ảnh đại diện');
      }
      await persistUser(payload.data as AuthUser);
    },
    [authFetch, persistUser],
  );

  const value = useMemo(
    () => ({
      user,
      token,
      refreshToken,
      booting,
      authFetch,
      clearAuth,
      requireLogin,
      shouldAskLogin,
      skipLoginPrompt,
      register,
      login,
      requestOTP,
      verifyOTP,
      updateName,
      refreshUser,
      uploadAvatar,
    }),
    [
      authFetch,
      booting,
      clearAuth,
      login,
      refreshToken,
      refreshUser,
      register,
      requestOTP,
      requireLogin,
      shouldAskLogin,
      skipLoginPrompt,
      token,
      user,
      updateName,
      uploadAvatar,
      verifyOTP,
    ],
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
      <AuthFormModal
        visible={showLoginForm}
        onClose={() => setShowLoginForm(false)}
        login={login}
        register={register}
        requestOTP={requestOTP}
        verifyOTP={verifyOTP}
      />
      <NameFormModal
        visible={Boolean(user && !user.name?.trim())}
        updateName={updateName}
      />
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return value;
}
