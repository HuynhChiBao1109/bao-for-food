import Constants from 'expo-constants';
import { Platform } from 'react-native';

const BACKEND_PORT = 8090;

function expoHost() {
  const hostUri = Constants.expoConfig?.hostUri;
  if (!hostUri) return null;

  return hostUri.split(':')[0] ?? null;
}

function fallbackBaseURL() {
  if (Platform.OS === 'web') {
    return `http://localhost:${BACKEND_PORT}`;
  }

  const host = expoHost();
  if (host) {
    return `http://${host}:${BACKEND_PORT}`;
  }

  if (Platform.OS === 'android') {
    return `http://10.0.2.2:${BACKEND_PORT}`;
  }

  return `http://192.168.1.25:${BACKEND_PORT}`;
}

export const API_BASE_URL = process.env.EXPO_PUBLIC_API_BASE_URL ?? fallbackBaseURL();
