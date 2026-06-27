import { API_BASE_URL } from '@/constants/api';
import { Fonts } from '@/constants/theme';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Animated,
  Image,
  Linking,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';

const BACKGROUND = '#6f8f46';
const SURFACE = '#fbfff3';
const SURFACE_SOFT = '#eef6df';
const INK = '#21320f';
const MUTED = '#667653';
const ACCENT = '#e67e45';
const PRIMARY = '#496a24';

const ALL_ICONS = ['🍜', '🍕', '🥗', '🍔', '🍣', '🥪', '🍰', '🍛', '🍗', '🍩'];

const FALLBACK_IMAGES = [
  'https://images.unsplash.com/photo-1551218808-94e220e084d2',
  'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe',
  'https://images.unsplash.com/photo-1604908177522-429a2d9abbd6',
];

type Media = {
  url?: string;
  image_url?: string;
};

type PhotoGroup = {
  category_label?: string;
  media?: Media[];
};

type Review = {
  author_name?: string;
  rating?: number;
  text?: string;
};

type OpeningHour = {
  day?: string;
  hours?: string;
};

type PisoRestaurant = {
  data_id?: string;
  title?: string;
  description?: string;
  type?: string;
  rating?: number;
  reviews?: number | Review[];
  pricing?: {
    min?: number;
    max?: number;
  };
  contacts?: {
    phone?: string;
    website?: string;
  };
  location?: {
    address?: {
      full?: string;
      street?: string;
      ward?: string;
      city?: string;
    };
    latitude?: number;
    longitude?: number;
  };
  open_state?: {
    is_open_now?: boolean;
    text?: string;
  };
  opening_hours?: OpeningHour[];
  features?: Record<string, string[]>;
  link_google_maps?: string;
  photos?: PhotoGroup[];
};

type TodayEatResponse = {
  data: {
    detail_source?: string;
    restaurant: PisoRestaurant;
  };
};

function restaurantImages(restaurant?: PisoRestaurant) {
  const urls =
    restaurant?.photos
      ?.flatMap((group) => group.media ?? [])
      .map((photo) => photo.url ?? photo.image_url)
      .filter((url): url is string => Boolean(url)) ?? [];

  return urls.length > 0 ? urls.slice(0, 8) : FALLBACK_IMAGES;
}

function reviewCount(restaurant?: PisoRestaurant) {
  if (typeof restaurant?.reviews === 'number') return restaurant.reviews;
  return restaurant?.reviews?.length ?? 0;
}

function reviewList(restaurant?: PisoRestaurant) {
  if (!Array.isArray(restaurant?.reviews)) return [];
  return restaurant.reviews.filter((review) => review.text).slice(0, 3);
}

function formatPrice(min?: number, max?: number) {
  if (!min && !max) return null;

  const formatter = new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: 'VND',
    maximumFractionDigits: 0,
  });

  if (min && max) return `${formatter.format(min)} - ${formatter.format(max)}`;
  return formatter.format(min ?? max ?? 0);
}

function restaurantType(type?: string) {
  if (!type) return 'Nhà hàng';
  return type
    .split('_')
    .map((word) => word.charAt(0) + word.slice(1).toLowerCase())
    .join(' ');
}

function featureLabels(restaurant?: PisoRestaurant) {
  const features = restaurant?.features;
  if (!features) return [];

  const labels: Record<string, string> = {
    has_delivery: 'Giao hàng',
    has_takeout: 'Mang đi',
    serves_dine_in: 'Ăn tại chỗ',
    has_parking_lot_free: 'Đậu xe miễn phí',
    accepts_reservations: 'Nhận đặt bàn',
    has_wi_fi: 'Wi-Fi',
    welcomes_children: 'Phù hợp trẻ em',
    has_seating_outdoors: 'Có chỗ ngồi ngoài trời',
  };

  return Object.values(features)
    .flat()
    .map((feature) => labels[feature])
    .filter((label): label is string => Boolean(label))
    .slice(0, 6);
}

export default function TodayEatScreen() {
  const [startIndex, setStartIndex] = useState(0);
  const [dots, setDots] = useState('');
  const [loading, setLoading] = useState(true);
  const [showReason, setShowReason] = useState(false);
  const [restaurant, setRestaurant] = useState<PisoRestaurant | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activePhotoIndex, setActivePhotoIndex] = useState(0);

  const scaleAnim = useRef(new Animated.Value(1)).current;
  const fadeAnim = useRef(new Animated.Value(0)).current;

  const images = useMemo(() => restaurantImages(restaurant ?? undefined), [restaurant]);
  const price = useMemo(
    () => formatPrice(restaurant?.pricing?.min, restaurant?.pricing?.max),
    [restaurant],
  );
  const features = useMemo(() => featureLabels(restaurant ?? undefined), [restaurant]);
  const visibleReviews = useMemo(() => reviewList(restaurant ?? undefined), [restaurant]);

  const revealResult = useCallback(() => {
    fadeAnim.setValue(0);
    setLoading(false);
    Animated.timing(fadeAnim, {
      toValue: 1,
      duration: 450,
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
      setActivePhotoIndex(0);
    } catch (err) {
      setRestaurant(null);
      setError(err instanceof Error ? err.message : 'Không chọn được quán lúc này');
    } finally {
      revealResult();
    }
  }, [revealResult]);

  useEffect(() => {
    if (!loading) return;

    const interval = setInterval(() => {
      setStartIndex((prev) => (prev + 1) % ALL_ICONS.length);

      Animated.sequence([
        Animated.timing(scaleAnim, {
          toValue: 1.25,
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
                      opacity: isCenter ? 1 : 0.45,
                      transform: [{ scale: isCenter ? scaleAnim : 1 }],
                    },
                  ]}
                >
                  {icon}
                </Animated.Text>
              );
            })}
          </View>

          <ThemedText style={styles.loadingText}>Đang chọn quán phù hợp{dots}</ThemedText>
        </View>
      ) : (
        <Animated.View style={{ flex: 1, opacity: fadeAnim }}>
          {error ? (
            <View style={styles.errorBox}>
              <ThemedText type="title" style={styles.errorTitle}>
                Chưa chọn được quán
              </ThemedText>
              <ThemedText style={styles.errorText}>{error}</ThemedText>
            </View>
          ) : (
            <ScrollView
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
            >
              <View style={styles.hero}>
                <Image source={{ uri: images[0] }} style={styles.heroImage} />
                <View style={styles.heroShade} />
                <View style={styles.heroContent}>
                  <ThemedText style={styles.kicker}>Hôm nay ăn ở đây</ThemedText>
                  <ThemedText type="title" style={styles.title}>
                    {restaurant?.title ?? 'Quán ăn hôm nay'}
                  </ThemedText>
                  <View style={styles.badgeRow}>
                    {restaurant?.open_state?.text ? (
                      <View
                        style={[
                          styles.badge,
                          restaurant.open_state.is_open_now ? styles.openBadge : styles.closedBadge,
                        ]}
                      >
                        <ThemedText style={styles.badgeText}>
                          {restaurant.open_state.text}
                        </ThemedText>
                      </View>
                    ) : null}
                  </View>
                </View>
              </View>

              <View style={styles.sheet}>
                <View style={styles.statsRow}>
                  <View style={styles.statBox}>
                    <ThemedText style={styles.statValue}>
                      {restaurant?.rating ? restaurant.rating.toFixed(1) : '-'}
                    </ThemedText>
                    <ThemedText style={styles.statLabel}>Điểm</ThemedText>
                  </View>
                  <View style={styles.statBox}>
                    <ThemedText style={styles.statValue}>
                      {reviewCount(restaurant ?? undefined)}
                    </ThemedText>
                    <ThemedText style={styles.statLabel}>Đánh giá</ThemedText>
                  </View>
                  <View style={styles.statBox}>
                    <ThemedText style={styles.statValue}>🍽️</ThemedText>
                    <ThemedText style={styles.statLabel}>
                      {restaurantType(restaurant?.type)}
                    </ThemedText>
                  </View>
                </View>

                <View style={styles.sectionBlock}>
                  <ThemedText style={styles.sectionTitle}>Địa chỉ</ThemedText>
                  <ThemedText style={styles.bodyText}>
                    {restaurant?.location?.address?.full ?? 'Chưa có địa chỉ'}
                  </ThemedText>
                </View>

                {price ? (
                  <View style={styles.sectionBlock}>
                    <ThemedText style={styles.sectionTitle}>Khoảng giá</ThemedText>
                    <ThemedText style={styles.bodyText}>{price}</ThemedText>
                  </View>
                ) : null}

                {features.length > 0 ? (
                  <View style={styles.sectionBlock}>
                    <ThemedText style={styles.sectionTitle}>Tiện ích</ThemedText>
                    <View style={styles.chipRow}>
                      {features.map((feature) => (
                        <View key={feature} style={styles.chip}>
                          <ThemedText style={styles.chipText}>{feature}</ThemedText>
                        </View>
                      ))}
                    </View>
                  </View>
                ) : null}

                {images.length > 1 ? (
                  <View style={styles.sectionBlock}>
                    <View style={styles.sectionHeader}>
                      <ThemedText style={styles.sectionTitle}>Hình ảnh</ThemedText>
                      <ThemedText style={styles.photoCounter}>
                        {activePhotoIndex + 1}/{images.length}
                      </ThemedText>
                    </View>
                    <ScrollView
                      horizontal
                      pagingEnabled
                      decelerationRate="fast"
                      showsHorizontalScrollIndicator={false}
                      snapToAlignment="center"
                      onMomentumScrollEnd={(event) => {
                        const index = Math.round(event.nativeEvent.contentOffset.x / 304);
                        setActivePhotoIndex(Math.min(Math.max(index, 0), images.length - 1));
                      }}
                    >
                      {images.map((img, i) => (
                        <Image
                          key={`${img}-${i}`}
                          source={{ uri: img }}
                          style={styles.galleryImage}
                        />
                      ))}
                    </ScrollView>
                    <View style={styles.dotRow}>
                      {images.map((img, index) => (
                        <View
                          key={`${img}-dot-${index}`}
                          style={[
                            styles.photoDot,
                            index === activePhotoIndex ? styles.photoDotActive : null,
                          ]}
                        />
                      ))}
                    </View>
                  </View>
                ) : null}

                {/* {restaurant?.opening_hours?.length ? (
                  <View style={styles.sectionBlock}>
                    <ThemedText style={styles.sectionTitle}>Giờ mở cửa</ThemedText>
                    {restaurant.opening_hours.slice(0, 7).map((hour) => (
                      <View key={`${hour.day}-${hour.hours}`} style={styles.hourRow}>
                        <ThemedText style={styles.hourDay}>{hour.day}</ThemedText>
                        <ThemedText style={styles.hourText}>{hour.hours}</ThemedText>
                      </View>
                    ))}
                  </View>
                ) : null} */}

                {visibleReviews.length > 0 ? (
                  <View style={styles.sectionBlock}>
                    <ThemedText style={styles.sectionTitle}>Người khác nói gì</ThemedText>
                    {visibleReviews.map((review, index) => (
                      <View key={`${review.author_name}-${index}`} style={styles.reviewBox}>
                        <ThemedText style={styles.reviewAuthor}>
                          {review.author_name ?? 'Ẩn danh'} · ⭐ {review.rating ?? '-'}
                        </ThemedText>
                        <ThemedText style={styles.reviewText}>{review.text}</ThemedText>
                      </View>
                    ))}
                  </View>
                ) : null}
              </View>
            </ScrollView>
          )}

          <View style={styles.footer}>
            <Pressable
              style={[styles.actionBtn, styles.mapBtn]}
              onPress={() => {
                if (restaurant?.link_google_maps) {
                  Linking.openURL(restaurant.link_google_maps);
                }
              }}
            >
              <ThemedText style={styles.mapBtnText}>Mở bản đồ</ThemedText>
            </Pressable>

            <Pressable
              style={[styles.actionBtn, styles.nextBtn]}
              onPress={() => setShowReason(true)}
            >
              <ThemedText style={styles.nextBtnText}>Đổi quán</ThemedText>
            </Pressable>
          </View>
        </Animated.View>
      )}

      <Modal transparent visible={showReason} animationType="fade">
        <View style={styles.overlay}>
          <View style={styles.popup}>
            <ThemedText type="title" style={styles.popupTitle}>
              Vì sao bạn muốn đổi quán?
            </ThemedText>

            {['Không thích món này hôm nay', 'Quán không phù hợp', 'Khác'].map((reason) => (
              <Pressable
                key={reason}
                style={styles.reasonBtn}
                onPress={() => {
                  setShowReason(false);
                  setStartIndex(0);
                  pickRestaurant();
                }}
              >
                <ThemedText style={styles.reasonText}>{reason}</ThemedText>
              </Pressable>
            ))}
          </View>
        </View>
      </Modal>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: BACKGROUND },

  loadingBox: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },

  iconRow: { flexDirection: 'row', marginBottom: 24 },
  icon: { fontSize: 38, marginHorizontal: 9 },
  loadingText: {
    color: '#f8ffe9',
    fontFamily: Fonts.rounded,
    fontSize: 17,
    fontWeight: '700',
  },

  scrollContent: {
    paddingBottom: 104,
  },

  hero: {
    height: 330,
    position: 'relative',
  },

  heroImage: {
    height: '100%',
    width: '100%',
  },

  heroShade: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(20,32,10,0.36)',
  },

  heroContent: {
    bottom: 26,
    left: 18,
    position: 'absolute',
    right: 18,
  },

  kicker: {
    color: '#f4ffd8',
    fontFamily: Fonts.rounded,
    fontSize: 14,
    fontWeight: '800',
    marginBottom: 8,
    textTransform: 'uppercase',
  },

  title: {
    color: '#fffdf5',
    fontFamily: Fonts.rounded,
    fontSize: 34,
    fontWeight: '900',
    lineHeight: 40,
  },

  badgeRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 12,
  },

  badge: {
    borderRadius: 999,
    paddingHorizontal: 12,
    paddingVertical: 7,
  },

  openBadge: {
    backgroundColor: 'rgba(218,247,181,0.92)',
  },

  closedBadge: {
    backgroundColor: 'rgba(255,223,206,0.94)',
  },

  badgeText: {
    color: INK,
    fontFamily: Fonts.rounded,
    fontSize: 12,
    fontWeight: '800',
  },

  sheet: {
    backgroundColor: SURFACE,
    borderTopLeftRadius: 26,
    borderTopRightRadius: 26,
    marginTop: -22,
    padding: 18,
  },

  statsRow: {
    flexDirection: 'row',
    gap: 10,
    marginBottom: 20,
  },

  statBox: {
    backgroundColor: SURFACE_SOFT,
    borderRadius: 14,
    flex: 1,
    minHeight: 78,
    justifyContent: 'center',
    padding: 10,
  },

  statValue: {
    color: PRIMARY,
    fontFamily: Fonts.rounded,
    fontSize: 21,
    fontWeight: '900',
    textAlign: 'center',
  },

  statLabel: {
    color: MUTED,
    fontFamily: Fonts.sans,
    fontSize: 12,
    fontWeight: '700',
    marginTop: 4,
    textAlign: 'center',
  },

  sectionBlock: {
    marginBottom: 20,
  },

  sectionTitle: {
    color: INK,
    fontFamily: Fonts.rounded,
    fontSize: 18,
    fontWeight: '900',
    marginBottom: 8,
  },

  sectionHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },

  photoCounter: {
    color: MUTED,
    fontFamily: Fonts.rounded,
    fontSize: 13,
    fontWeight: '800',
  },

  bodyText: {
    color: MUTED,
    fontFamily: Fonts.sans,
    fontSize: 15,
    lineHeight: 22,
  },

  chipRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },

  chip: {
    backgroundColor: '#fff7e9',
    borderRadius: 999,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },

  chipText: {
    color: '#7b4a1c',
    fontFamily: Fonts.rounded,
    fontSize: 12,
    fontWeight: '800',
  },

  galleryImage: {
    borderRadius: 14,
    height: 210,
    marginRight: 10,
    width: 294,
  },

  dotRow: {
    flexDirection: 'row',
    justifyContent: 'center',
    marginTop: 10,
  },

  photoDot: {
    backgroundColor: '#cddbb8',
    borderRadius: 999,
    height: 7,
    marginHorizontal: 3,
    width: 7,
  },

  photoDotActive: {
    backgroundColor: PRIMARY,
    width: 18,
  },

  hourRow: {
    borderBottomColor: '#dce8c8',
    borderBottomWidth: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 9,
  },

  hourDay: {
    color: INK,
    fontFamily: Fonts.sans,
    fontSize: 14,
    fontWeight: '800',
  },

  hourText: {
    color: MUTED,
    fontFamily: Fonts.sans,
    fontSize: 14,
  },

  reviewBox: {
    backgroundColor: SURFACE_SOFT,
    borderRadius: 14,
    marginBottom: 10,
    padding: 12,
  },

  reviewAuthor: {
    color: PRIMARY,
    fontFamily: Fonts.rounded,
    fontSize: 13,
    fontWeight: '900',
    marginBottom: 4,
  },

  reviewText: {
    color: MUTED,
    fontFamily: Fonts.sans,
    fontSize: 14,
    lineHeight: 20,
  },

  errorBox: {
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },

  errorTitle: {
    color: '#fffdf5',
    fontFamily: Fonts.rounded,
    fontWeight: '900',
    marginBottom: 8,
  },

  errorText: {
    color: '#f5ffd9',
    fontFamily: Fonts.sans,
    fontSize: 15,
  },

  footer: {
    backgroundColor: 'rgba(251,255,243,0.96)',
    bottom: 0,
    flexDirection: 'row',
    gap: 10,
    left: 0,
    padding: 14,
    position: 'absolute',
    right: 0,
  },

  actionBtn: {
    alignItems: 'center',
    borderRadius: 16,
    flex: 1,
    paddingVertical: 14,
  },

  mapBtn: {
    backgroundColor: SURFACE_SOFT,
  },

  mapBtnText: {
    color: PRIMARY,
    fontFamily: Fonts.rounded,
    fontWeight: '900',
  },

  nextBtn: {
    backgroundColor: ACCENT,
  },

  nextBtnText: {
    color: '#fffdf5',
    fontFamily: Fonts.rounded,
    fontWeight: '900',
  },

  overlay: {
    alignItems: 'center',
    backgroundColor: 'rgba(17,28,9,0.5)',
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },

  popup: {
    backgroundColor: SURFACE,
    borderRadius: 20,
    padding: 20,
    width: '100%',
  },

  popupTitle: {
    color: INK,
    fontFamily: Fonts.rounded,
    fontWeight: '900',
    marginBottom: 12,
  },

  reasonBtn: {
    backgroundColor: SURFACE_SOFT,
    borderRadius: 14,
    marginBottom: 8,
    paddingHorizontal: 12,
    paddingVertical: 13,
  },

  reasonText: {
    color: INK,
    fontFamily: Fonts.sans,
    fontWeight: '700',
  },
});
