import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';

import 'react-native-reanimated';

import { AuthProvider } from '@/contexts/auth-context';
import { LocationProvider } from '@/contexts/location-context';
import { useColorScheme } from '@/hooks/use-color-scheme';

export const unstable_settings = {
  anchor: '(tabs)',
};

export default function RootLayout() {
  const colorScheme = useColorScheme();
  const [booting, setBooting] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => setBooting(false), 2000);
    return () => clearTimeout(timer);
  }, []);

  if (booting) {
    return (
      <View style={styles.bootScreen}>
        <Text style={styles.bootIcon}>🍜</Text>
        <Text style={styles.bootTitle}>Hôm nay ăn gì?</Text>
        <ActivityIndicator color="#f8ffe9" size="small" />
      </View>
    );
  }

  return (
    <AuthProvider>
      <LocationProvider>
        <Stack
          screenOptions={{
            headerShown: false,
          }}
        >
          <Stack.Screen name="(tabs)" />
          <Stack.Screen name="profile" />
          <Stack.Screen
            name="modal"
            options={{
              presentation: 'modal',
              title: 'Modal',
              headerShown: true,
            }}
          />
        </Stack>

        <StatusBar style={colorScheme === 'dark' ? 'light' : 'dark'} />
      </LocationProvider>
    </AuthProvider>
  );
}

const styles = StyleSheet.create({
  bootScreen: {
    alignItems: 'center',
    backgroundColor: '#6f8f46',
    flex: 1,
    justifyContent: 'center',
  },

  bootIcon: {
    fontSize: 58,
    marginBottom: 12,
  },

  bootTitle: {
    color: '#fffdf5',
    fontSize: 24,
    fontWeight: '900',
    marginBottom: 18,
  },
});
