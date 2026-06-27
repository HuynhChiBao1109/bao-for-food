import { View, StyleSheet, Animated, Image, ScrollView, Pressable, Modal } from 'react-native';
import { useEffect, useRef, useState } from 'react';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';

const PRIMARY = '#a7c068';

const ALL_ICONS = ['🍜', '🍕', '🥗', '🍔', '🍣', '🥪', '🍰', '🍛', '🍗', '🍩'];

const MOCK_RESTAURANT = {
  name: 'Quán Bún Bò Huế O Hạnh',
  images: [
    'https://images.unsplash.com/photo-1551218808-94e220e084d2',
    'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe',
    'https://images.unsplash.com/photo-1604908177522-429a2d9abbd6',
  ],
  address: '12 Nguyễn Trãi, Quận 1, TP.HCM',
  open: '07:00',
  close: '22:00',
  menu: ['Bún bò huế', 'Chả cua', 'Giò heo', 'Bún đặc biệt'],
  rating: 4.6,
  reviews: ['Ngon, đúng vị Huế', 'Nước lèo đậm đà', 'Sẽ quay lại lần sau'],
};

export default function TodayEatScreen() {
  const [startIndex, setStartIndex] = useState(0);
  const [dots, setDots] = useState('');
  const [loading, setLoading] = useState(true);
  const [showReason, setShowReason] = useState(false);

  const scaleAnim = useRef(new Animated.Value(1)).current;
  const fadeAnim = useRef(new Animated.Value(0)).current;

  /** LOADING CAROUSEL */
  useEffect(() => {
    if (!loading) return;

    const interval = setInterval(() => {
      setStartIndex((prev) => (prev + 1) % ALL_ICONS.length);

      Animated.sequence([
        Animated.timing(scaleAnim, {
          toValue: 1.3,
          duration: 250,
          useNativeDriver: true,
        }),
        Animated.timing(scaleAnim, {
          toValue: 1,
          duration: 250,
          useNativeDriver: true,
        }),
      ]).start();
    }, 900);

    return () => clearInterval(interval);
  }, [loading]);

  /** DOTS */
  useEffect(() => {
    if (!loading) return;

    const dotInterval = setInterval(() => {
      setDots((prev) => (prev.length < 3 ? prev + '.' : ''));
    }, 400);

    return () => clearInterval(dotInterval);
  }, [loading]);

  /** STOP AFTER 10s */
  useEffect(() => {
    const timer = setTimeout(() => {
      setLoading(false);
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 500,
        useNativeDriver: true,
      }).start();
    }, 5000);

    return () => clearTimeout(timer);
  }, []);

  const visibleIcons = Array.from({ length: 5 }).map(
    (_, i) => ALL_ICONS[(startIndex + i) % ALL_ICONS.length],
  );

  return (
    <ThemedView style={styles.container}>
      {loading ? (
        <View style={styles.loadingBox}>
          <View style={styles.iconRow}>
            {visibleIcons.map((icon, index) => {
              const isCenter = index === 2;
              return (
                <Animated.Text
                  key={index}
                  style={[
                    styles.icon,
                    {
                      opacity: isCenter ? 1 : 0.4,
                      transform: [{ scale: isCenter ? scaleAnim : 1 }],
                    },
                  ]}
                >
                  {icon}
                </Animated.Text>
              );
            })}
          </View>

          <ThemedText style={styles.text}>Đang chọn quán phù hợp{dots}</ThemedText>
        </View>
      ) : (
        <Animated.View style={{ flex: 1, opacity: fadeAnim }}>
          <ScrollView showsVerticalScrollIndicator={false}>
            {/* IMAGES */}
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {MOCK_RESTAURANT.images.map((img, i) => (
                <Image key={i} source={{ uri: img }} style={styles.image} />
              ))}
            </ScrollView>

            <View style={styles.content}>
              <ThemedText type="title" style={styles.title}>
                🍜 {MOCK_RESTAURANT.name}
              </ThemedText>

              <ThemedText style={styles.info}>📍 {MOCK_RESTAURANT.address}</ThemedText>

              <ThemedText style={styles.info}>
                ⏰ {MOCK_RESTAURANT.open} – {MOCK_RESTAURANT.close}
              </ThemedText>

              <ThemedText style={styles.info}>⭐ {MOCK_RESTAURANT.rating} / 5</ThemedText>

              <ThemedText style={styles.section}>🍽️ Menu</ThemedText>
              {MOCK_RESTAURANT.menu.map((m) => (
                <ThemedText key={m} style={styles.item}>
                  • {m}
                </ThemedText>
              ))}

              <ThemedText style={styles.section}>💬 Đánh giá</ThemedText>
              {MOCK_RESTAURANT.reviews.map((r, i) => (
                <View key={i} style={styles.reviewBox}>
                  <ThemedText style={styles.review}>“{r}”</ThemedText>
                </View>
              ))}
            </View>
          </ScrollView>

          {/* FOOTER */}
          <View style={styles.footer}>
            <Pressable style={styles.okBtn}>
              <ThemedText>OK</ThemedText>
            </Pressable>

            <Pressable style={styles.nextBtn} onPress={() => setShowReason(true)}>
              <ThemedText style={{ color: '#fff' }}>Next</ThemedText>
            </Pressable>
          </View>
        </Animated.View>
      )}

      {/* POPUP */}
      <Modal transparent visible={showReason} animationType="fade">
        <View style={styles.overlay}>
          <View style={styles.popup}>
            <ThemedText type="title" style={{ marginBottom: 12 }}>
              Vì sao bạn muốn đổi quán?
            </ThemedText>

            {['Không thích món này hôm nay', 'Quán không phù hợp', 'Khác'].map((r) => (
              <Pressable
                key={r}
                style={styles.reasonBtn}
                onPress={() => {
                  setShowReason(false);
                  setLoading(true);
                  setStartIndex(0);
                }}
              >
                <ThemedText>{r}</ThemedText>
              </Pressable>
            ))}
          </View>
        </View>
      </Modal>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: PRIMARY },

  loadingBox: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },

  iconRow: { flexDirection: 'row', marginBottom: 24 },
  icon: { fontSize: 36, marginHorizontal: 10 },

  text: { color: '#fff', fontSize: 16 },

  image: {
    width: 300,
    height: 200,
    margin: 10,
    borderRadius: 14,
  },

  content: { padding: 16 },

  title: { color: '#fff', marginBottom: 6 },

  info: { color: '#f1f1f1', marginBottom: 4 },

  section: {
    marginTop: 16,
    color: '#fff',
    fontWeight: '600',
  },

  item: { color: '#f4f4f4', marginLeft: 6 },

  reviewBox: {
    backgroundColor: 'rgba(255,255,255,0.15)',
    borderRadius: 10,
    padding: 10,
    marginTop: 6,
  },

  review: { color: '#fff', fontStyle: 'italic' },

  footer: {
    flexDirection: 'row',
    padding: 12,
  },

  okBtn: {
    flex: 1,
    backgroundColor: '#fff',
    padding: 12,
    borderRadius: 12,
    alignItems: 'center',
    marginRight: 8,
  },

  nextBtn: {
    flex: 1,
    backgroundColor: '#8fab4f',
    padding: 12,
    borderRadius: 12,
    alignItems: 'center',
  },

  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.4)',
    justifyContent: 'center',
    alignItems: 'center',
  },

  popup: {
    width: '80%',
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
  },

  reasonBtn: { paddingVertical: 12 },
});
