import { useAuth } from '@/contexts/auth-context';
import { API_BASE_URL } from '@/constants/api';
import { useCallback, useEffect, useState } from 'react';
import { Image, ScrollView, StyleSheet, View } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';

type RestaurantItem = {
  data_id: string;
  recorded_at: string;
  detail: {
    title?: string;
    rating?: number;
    location?: { address?: { full?: string } };
    photos?: { media?: { url?: string }[] }[];
  };
};

type ListResponse = { data: RestaurantItem[] };

function firstPhoto(item: RestaurantItem) {
  return item.detail.photos?.flatMap((group) => group.media ?? []).find((photo) => photo.url)?.url;
}

export default function ViewedRestaurantsScreen() {
  const { token } = useAuth();
  const [items, setItems] = useState<RestaurantItem[]>([]);

  const load = useCallback(async () => {
    if (!token) return;
    const response = await fetch(`${API_BASE_URL}/api/v1/restaurants/viewed`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const payload = (await response.json()) as ListResponse;
    setItems(payload.data ?? []);
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <ThemedView style={styles.container}>
      <ThemedText type="title" style={styles.title}>
        Quán ăn bạn đã xem hôm nay
      </ThemedText>
      <ScrollView contentContainerStyle={styles.list}>
        {items.map((item) => (
          <View key={`${item.data_id}-${item.recorded_at}`} style={styles.card}>
            {firstPhoto(item) ? <Image source={{ uri: firstPhoto(item) }} style={styles.image} /> : null}
            <View style={styles.cardBody}>
              <ThemedText style={styles.name}>{item.detail.title ?? 'Quán ăn'}</ThemedText>
              <ThemedText style={styles.meta}>
                {new Date(item.recorded_at).toLocaleTimeString('vi-VN', {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
                {item.detail.rating ? ` · ⭐ ${item.detail.rating}` : ''}
              </ThemedText>
              <ThemedText style={styles.address} numberOfLines={2}>
                {item.detail.location?.address?.full ?? 'Chưa có địa chỉ'}
              </ThemedText>
            </View>
          </View>
        ))}
        {items.length === 0 ? (
          <ThemedText style={styles.empty}>Bạn chưa mở bản đồ quán nào hôm nay.</ThemedText>
        ) : null}
      </ScrollView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#6f8f46', paddingTop: 56 },
  title: { color: '#fffdf5', marginHorizontal: 18, marginBottom: 16 },
  list: { padding: 16, paddingBottom: 32 },
  card: {
    backgroundColor: '#fbfff3',
    borderRadius: 18,
    flexDirection: 'row',
    marginBottom: 12,
    overflow: 'hidden',
  },
  image: { height: 104, width: 104 },
  cardBody: { flex: 1, padding: 12 },
  name: { color: '#21320f', fontSize: 16, fontWeight: '900' },
  meta: { color: '#496a24', fontSize: 13, fontWeight: '700', marginTop: 4 },
  address: { color: '#667653', fontSize: 13, marginTop: 6 },
  empty: { color: '#f8ffe9', fontSize: 15, textAlign: 'center', marginTop: 40 },
});
