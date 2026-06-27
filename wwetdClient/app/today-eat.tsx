import { View, StyleSheet, Animated, Image, ScrollView, Pressable, Modal } from 'react-native';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { API_BASE_URL } from '@/constants/api';

const PRIMARY = '#a7c068';

const ALL_ICONS = ['🍜', '🍕', '🥗', '🍔', '🍣', '🥪', '🍰', '🍛', '🍗', '🍩'];

const FALLBACK_IMAGES = [
  'https://images.unsplash.com/photo-1551218808-94e220e084d2',
  'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe',
  'https://images.unsplash.com/photo-1604908177522-429a2d9abbd6',
];

type PisoRestaurant = {
  title?: string;
  type?: string;
  rating?: number;
  reviews?: number;
  contacts?: {
    phone?: string;
    website?: string;
  };
  location?: {
    address?: {
      full?: string;
    };
    latitude?: number;
    longitude?: number;
  };
  open_state?: {
    is_open_now?: boolean;
    text?: string;
  };
  link_google_maps?: string;
  link_place_detail?: string;
  images?: string[];
  photos?: (string | { url?: string; image_url?: string })[];
};

type TodayEatResponse = {
  data: {
    restaurant: PisoRestaurant;
  };
};

function restaurantImages(restaurant?: PisoRestaurant) {
  const photoURLs =
    restaurant?.photos
      ?.map((photo) => {
        if (typeof photo === 'string') return photo;
        return photo.url ?? photo.image_url;
      })
      .filter((url): url is string => Boolean(url)) ?? [];

  const images = [...(restaurant?.images ?? []), ...photoURLs];
  return images.length > 0 ? images : FALLBACK_IMAGES;
}

export default function TodayEatScreen() {
  const [startIndex, setStartIndex] = useState(0);
  const [dots, setDots] = useState('');
  const [loading, setLoading] = useState(true);
  const [showReason, setShowReason] = useState(false);
  const [restaurant, setRestaurant] = useState<PisoRestaurant | null>(null);
  const [error, setError] = useState<string | null>(null);

  const scaleAnim = useRef(new Animated.Value(1)).current;
  const fadeAnim = useRef(new Animated.Value(0)).current;

  const images = useMemo(() => restaurantImages(restaurant ?? undefined), [restaurant]);

  const revealResult = useCallback(() => {
    fadeAnim.setValue(0);
    setLoading(false);
    Animated.timing(fadeAnim, {
      toValue: 1,
      duration: 500,
      useNativeDriver: true,
    }).start();
  }, [fadeAnim]);

  const pickRestaurant = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const url = `${API_BASE_URL}/api/v1/restaurants/today?query=${encodeURIComponent(
        'quán ăn',
      )}&limit=20`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Không chọn được quán lúc này');
      }

      const payload = (await response.json()) as TodayEatResponse;
      setRestaurant(payload.data.restaurant);
    } catch (err) {
      setRestaurant(null);
      setError(err instanceof Error ? err.message : 'Không chọn được quán lúc này');
    } finally {
      revealResult();
    }
  }, [revealResult]);

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
  }, [loading, scaleAnim]);

  /** DOTS */
  useEffect(() => {
    if (!loading) return;

    const dotInterval = setInterval(() => {
      setDots((prev) => (prev.length < 3 ? prev + '.' : ''));
    }, 400);

    return () => clearInterval(dotInterval);
  }, [loading]);

  useEffect(() => {
    pickRestaurant();
  }, [pickRestaurant]);

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
          {error ? (
            <View style={styles.errorBox}>
              <ThemedText type="title" style={styles.title}>
                Chưa chọn được quán
              </ThemedText>
              <ThemedText style={styles.info}>{error}</ThemedText>
            </View>
          ) : (
            <ScrollView showsVerticalScrollIndicator={false}>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                {images.map((img, i) => (
                  <Image key={`${img}-${i}`} source={{ uri: img }} style={styles.image} />
                ))}
              </ScrollView>

              <View style={styles.content}>
                <ThemedText type="title" style={styles.title}>
                  🍜 {restaurant?.title ?? 'Quán ăn hôm nay'}
                </ThemedText>

                <ThemedText style={styles.info}>
                  📍 {restaurant?.location?.address?.full ?? 'Chưa có địa chỉ'}
                </ThemedText>

                {restaurant?.open_state?.text ? (
                  <ThemedText style={styles.info}>⏰ {restaurant.open_state.text}</ThemedText>
                ) : null}

                {restaurant?.rating ? (
                  <ThemedText style={styles.info}>
                    ⭐ {restaurant.rating} / 5
                    {restaurant.reviews ? ` (${restaurant.reviews} đánh giá)` : ''}
                  </ThemedText>
                ) : null}

                {restaurant?.type ? (
                  <>
                    <ThemedText style={styles.section}>🍽️ Loại quán</ThemedText>
                    <ThemedText style={styles.item}>• {restaurant.type}</ThemedText>
                  </>
                ) : null}

                {restaurant?.contacts?.phone || restaurant?.contacts?.website ? (
                  <>
                    <ThemedText style={styles.section}>📞 Liên hệ</ThemedText>
                    {restaurant.contacts.phone ? (
                      <ThemedText style={styles.item}>• {restaurant.contacts.phone}</ThemedText>
                    ) : null}
                    {restaurant.contacts.website ? (
                      <ThemedText style={styles.item}>• {restaurant.contacts.website}</ThemedText>
                    ) : null}
                  </>
                ) : null}
              </View>
            </ScrollView>
          )}

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
                  setStartIndex(0);
                  pickRestaurant();
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

  errorBox: {
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },

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
