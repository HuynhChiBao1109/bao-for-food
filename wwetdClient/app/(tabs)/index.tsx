import { useEffect, useState } from 'react';
import { Modal, Pressable, StyleSheet, View } from 'react-native';
import { Link } from 'expo-router';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { useCurrentLocation } from '@/contexts/location-context';

export default function HomeScreen() {
  const { coordinates, loading, permissionStatus, requestCurrentLocation } = useCurrentLocation();
  const [showLocationPopup, setShowLocationPopup] = useState(false);

  useEffect(() => {
    if (!coordinates && !permissionStatus) {
      const timer = setTimeout(() => setShowLocationPopup(true), 450);
      return () => clearTimeout(timer);
    }
  }, [coordinates, permissionStatus]);

  const askLocation = async () => {
    await requestCurrentLocation();
    setShowLocationPopup(false);
  };

  return (
    <ThemedView style={styles.container}>
      <Link href="/rate-food" asChild>
        <Pressable style={styles.card}>
          <ThemedText type="title" style={styles.cardTitle}>
            🍽️ Đánh giá món ăn
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Chia sẻ cảm nhận về món bạn vừa ăn</ThemedText>
        </Pressable>
      </Link>

      <Link href="/today-eat" asChild>
        <Pressable style={styles.card}>
          <ThemedText type="title" style={styles.cardTitle}>
            👀 Hôm nay ăn gì
          </ThemedText>
          <ThemedText style={styles.cardDesc}>Anh ơi bữa nay ăn gì</ThemedText>
        </Pressable>
      </Link>

      <Modal transparent visible={showLocationPopup} animationType="fade">
        <View style={styles.overlay}>
          <View style={styles.popup}>
            <ThemedText type="title" style={styles.popupTitle}>
              Cho WWETD biết bạn đang ở đâu?
            </ThemedText>
            <ThemedText style={styles.popupDesc}>
              Mình sẽ dùng vị trí hiện tại để gợi ý quán ăn gần bạn hơn. Nếu bỏ qua, app vẫn chọn
              quán dựa trên IP.
            </ThemedText>

            <Pressable style={styles.allowBtn} onPress={askLocation} disabled={loading}>
              <ThemedText style={styles.allowText}>
                {loading ? 'Đang lấy vị trí...' : 'Cho phép vị trí'}
              </ThemedText>
            </Pressable>

            <Pressable style={styles.skipBtn} onPress={() => setShowLocationPopup(false)}>
              <ThemedText style={styles.skipText}>Để sau</ThemedText>
            </Pressable>
          </View>
        </View>
      </Modal>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  card: {
    width: '100%',
    maxWidth: 350,
    height: 150,
    backgroundColor: '#a7c068',
    borderRadius: 20,
    padding: 20,
    justifyContent: 'center',
    marginBottom: 20,

    shadowColor: '#000',
    shadowOpacity: 0.15,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 4 },
  },
  cardTitle: {
    color: '#fff',
    marginBottom: 8,
  },
  cardDesc: {
    color: '#f4f4f4',
    fontSize: 14,
  },
  overlay: {
    alignItems: 'center',
    backgroundColor: 'rgba(19,31,12,0.48)',
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },
  popup: {
    backgroundColor: '#fbfff3',
    borderRadius: 22,
    padding: 20,
    width: '100%',
  },
  popupTitle: {
    color: '#21320f',
    marginBottom: 8,
  },
  popupDesc: {
    color: '#667653',
    fontSize: 15,
    lineHeight: 22,
    marginBottom: 18,
  },
  allowBtn: {
    alignItems: 'center',
    backgroundColor: '#e67e45',
    borderRadius: 16,
    paddingVertical: 14,
  },
  allowText: {
    color: '#fffdf5',
    fontWeight: '800',
  },
  skipBtn: {
    alignItems: 'center',
    paddingVertical: 13,
  },
  skipText: {
    color: '#496a24',
    fontWeight: '800',
  },
});
