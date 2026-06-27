import { Platform } from 'react-native';

const fallbackBaseURL = Platform.select({
  android: 'http://10.0.2.2:8081',
  default: 'http://localhost:8081',
});

export const API_BASE_URL = process.env.EXPO_PUBLIC_API_BASE_URL ?? fallbackBaseURL;
